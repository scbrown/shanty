package session

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// The session switcher: type-to-filter, most-recently-used first.
//
// WHY THIS IS NOT `choose-tree -Zs`
//
// tmux's stock prefix+s is a tree you scroll. Two things make it the wrong shape
// on a fleet of ~20 agent sessions:
//
//  1. It does not filter as you type. `f` enters a FORMAT filter and C-s does a
//     name search — both are a mode you must first know to enter. The ask is that
//     the first keystroke after the menu opens narrows the list.
//
//  2. It opens with the cursor on the CURRENT session — the one entry that is
//     never the answer — and its only recency sort is `-O time`, which tmux
//     documents as ACTIVITY. Activity is the wrong clock here: an agent session
//     printing output ranks above the session the operator was actually in a
//     moment ago, so `-O time` puts strangers on top. Measured on a live fleet:
//     ordering by activity and ordering by attachment agreed on nothing below the
//     first row.
//
// So the order is built here instead, from `session_last_attached` (which tmux
// updates on switch-client, not on output) with the client's own
// `client_last_session` pinned to the top. That makes prefix+s,Enter a toggle
// back to where you just were, and it stays a toggle when the list re-sorts —
// which a cursor-position hack on choose-tree would not.
//
// fzf does the fuzzy matching. It is a dependency, deliberately: writing a fuzzy
// matcher into shanty to avoid a tool the host already has is the wrong trade.
// When fzf is absent the generated binding falls back to `choose-tree -Zs`, so
// prefix+s always does SOMETHING (see config.RenderKeybindings).

// Candidate is one switchable session, as offered by the picker.
type Candidate struct {
	Name string
	// LastAttached is tmux's session_last_attached as a unix time. Zero means no
	// client has ever attached to or switched into this session — a real state on
	// a fleet whose sessions are created by a launcher, not by a human.
	LastAttached int64
	// IsLast marks the session this client switched away from. Exactly one
	// candidate can carry it, and it sorts first regardless of timestamps.
	IsLast bool
}

// orderCandidates ranks sessions for the switcher and drops the current one.
//
// Order: the last session switched from, then everything ever attached to,
// most-recent first, then the never-attached by name. The never-attached tail is
// sorted by NAME rather than left in tmux's order so that the bottom of the list
// is stable between invocations; an unstable tail makes muscle memory impossible.
//
// current is dropped because switching to the session you are in is a no-op that
// costs a row at the top of the list, where the cursor lands.
func orderCandidates(sessions []Candidate, current string) []Candidate {
	out := make([]Candidate, 0, len(sessions))
	for _, s := range sessions {
		if s.Name == current || s.Name == "" {
			continue
		}
		out = append(out, s)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.IsLast != b.IsLast {
			return a.IsLast
		}
		// Never-attached sinks below everything with a real attach time.
		if (a.LastAttached == 0) != (b.LastAttached == 0) {
			return b.LastAttached == 0
		}
		if a.LastAttached != b.LastAttached {
			return a.LastAttached > b.LastAttached
		}
		return a.Name < b.Name
	})
	return out
}

// tmuxArgs prefixes the socket flag only when we are NOT already inside tmux.
//
// Inside a display-popup, $TMUX names the server AND the client that opened the
// popup, so a bare `tmux` reaches exactly the right one — including for
// switch-client, which must act on the client that pressed the key. Adding -L
// there would be worse than redundant: the popup inherits the tmux SERVER's
// environment, which need not carry SHANTY_TMUX_SOCKET, so a resolved socket
// name can disagree with the server we are actually running under.
func tmuxArgs(args ...string) []string {
	if os.Getenv("TMUX") != "" {
		return args
	}
	return append([]string{"-L", socketName}, args...)
}

// clientContext reports the session the client is in and the one it switched
// away from, asking tmux from wherever the picker is running.
func (m *Manager) clientContext() (current, last string) {
	out, err := exec.Command(m.tmuxBin,
		tmuxArgs("display-message", "-p", "#{client_session}\t#{client_last_session}")...).Output()
	if err != nil {
		return "", ""
	}
	parts := strings.SplitN(strings.TrimRight(string(out), "\n"), "\t", 2)
	if len(parts) != 2 {
		return strings.TrimSpace(parts[0]), ""
	}
	return parts[0], parts[1]
}

// Candidates returns the switch targets in the order the picker offers them.
func (m *Manager) Candidates() ([]Candidate, error) {
	current, last := m.clientContext()

	out, err := exec.Command(m.tmuxBin,
		tmuxArgs("list-sessions", "-F", "#{session_last_attached}\t#{session_name}")...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil // no server: no sessions, not an error
		}
		return nil, fmt.Errorf("listing sessions: %w", err)
	}

	var all []Candidate
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if line == "" {
			continue
		}
		ts, name, ok := strings.Cut(line, "\t")
		if !ok || name == "" {
			continue
		}
		// An empty timestamp is tmux's answer for "never attached", not a parse
		// failure — treat it as zero rather than dropping the session, or a fresh
		// fleet's sessions would be unreachable from the picker entirely.
		n, _ := strconv.ParseInt(ts, 10, 64)
		all = append(all, Candidate{Name: name, LastAttached: n, IsLast: name == last})
	}
	return orderCandidates(all, current), nil
}

// pickerLines renders candidates for fzf: the name, a tab, then an annotation
// that is shown but never matched against (fzf --nth=1).
func pickerLines(cands []Candidate) string {
	var b strings.Builder
	for _, c := range cands {
		b.WriteString(c.Name)
		switch {
		case c.IsLast:
			b.WriteString("\t← last")
		case c.LastAttached == 0:
			b.WriteString("\tnever visited")
		default:
			b.WriteString("\t")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// Pick runs the interactive switcher: fuzzy-filter the candidates, then switch
// the calling client to the chosen session.
func (m *Manager) Pick() error {
	fzf, err := exec.LookPath("fzf")
	if err != nil {
		// The generated binding checks for fzf before opening the popup, so this
		// is the hand-run path. Name the dependency instead of failing blank.
		return fmt.Errorf("fzf not found in PATH — shanty pick needs it " +
			"(prefix+s falls back to tmux's choose-tree without it)")
	}

	cands, err := m.Candidates()
	if err != nil {
		return err
	}
	if len(cands) == 0 {
		return fmt.Errorf("no other sessions to switch to")
	}

	cmd := exec.Command(fzf,
		// --no-sort keeps OUR order. fzf re-ranks matches by score by default,
		// which would scramble the recency ordering the moment you typed a
		// character — the requirement is that recency survives filtering.
		"--no-sort",
		// --reverse puts the first candidate at the TOP, under the cursor, so
		// prefix+s then Enter is a toggle to the last session.
		"--reverse",
		"--no-multi",
		"--delimiter", "\t",
		// Match on the name only; the annotation column is decoration and must
		// not make "last" a search term.
		"--nth", "1",
		"--prompt", "switch to > ",
		"--header", fmt.Sprintf("%d sessions · most recent first · Esc cancels", len(cands)),
	)
	cmd.Stdin = strings.NewReader(pickerLines(cands))
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		// fzf exits 130 on Esc/C-c and 1 on no match. Cancelling is not a failure:
		// returning an error here would paint a tmux popup red for pressing Esc.
		return nil
	}
	choice := strings.TrimSpace(string(out))
	if choice == "" {
		return nil
	}
	name, _, _ := strings.Cut(choice, "\t")
	if name == "" {
		return nil
	}

	if out, err := exec.Command(m.tmuxBin,
		tmuxArgs("switch-client", "-t", name)...).CombinedOutput(); err != nil {
		return fmt.Errorf("switching to %q: %v: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}
