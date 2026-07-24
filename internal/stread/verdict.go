package stread

import (
	"strconv"
	"strings"
)

// Verdict is st's work cell, decoded.
//
// st packs three facts into one column and documents the notation in its own source:
//
//	busy              the base verdict
//	busy+1sh          ...plus N background shells STILL RUNNING
//	saturated·948k    ...plus the context size, only when saturated
//
// Decoding it matters because the notation is lossy to a human at a glance and the
// most important case is the least obvious: st's comment on the `+Nsh` suffix says
// `idle+1sh` is "idle AND carrying live work". An agent whose turn ended with a
// build, a test run, or a `gh run watch` still live is NOT finished — and "idle" is
// exactly the word that misleads a coordinator into dispatching over the top of it.
//
// Nothing here infers anything. Every field is read out of st's own cell; this type
// only makes the facts st already published legible.
type Verdict struct {
	Word     string // busy / idle / waiting / wedged / queued / saturated / ?
	Shells   int    // background shells still running (0 when the cell says nothing)
	ContextK int    // context size in thousands of tokens (0 when absent)
	Raw      string // the cell exactly as st printed it
}

// ParseVerdict decodes a work cell.
func ParseVerdict(cell string) Verdict {
	v := Verdict{Raw: cell, Word: StateWord(cell)}
	rest := cell
	// The context suffix rides on a middot and only appears with `saturated`.
	if i := strings.LastIndex(rest, "·"); i >= 0 {
		tail := strings.TrimSuffix(strings.TrimSpace(rest[i+len("·"):]), "k")
		if n, err := strconv.Atoi(tail); err == nil {
			v.ContextK = n
		}
		rest = rest[:i]
	}
	// The shell suffix is `+Nsh` and can ride on ANY verdict, idle included.
	if i := strings.LastIndex(rest, "+"); i >= 0 {
		tail := strings.TrimSuffix(strings.TrimSpace(rest[i+1:]), "sh")
		if n, err := strconv.Atoi(tail); err == nil {
			v.Shells = n
		}
	}
	return v
}

// Why explains, in words, what st's cell says about this agent — "" when the
// verdict speaks for itself and carries no extra evidence.
//
// This is the answer to "why does the bar think this agent is busy/idle", and it is
// worth surfacing because a coordinator deciding whether to dispatch is choosing
// between the bar and some other idle signal. A verdict with its evidence attached
// can be trusted or argued with; a bare word can only be believed or ignored.
func (v Verdict) Why() string {
	var parts []string
	if v.Shells == 1 {
		parts = append(parts, "1 shell live")
	} else if v.Shells > 1 {
		parts = append(parts, strconv.Itoa(v.Shells)+" shells live")
	}
	if v.ContextK > 0 {
		parts = append(parts, strconv.Itoa(v.ContextK)+"k ctx")
	}
	switch v.Word {
	case "waiting":
		parts = append(parts, "blocked on a question")
	case "queued":
		parts = append(parts, "unsubmitted text in the box")
	case "wedged":
		parts = append(parts, "pane dead or stuck")
	case "?":
		parts = append(parts, "st could not tell")
	}
	return strings.Join(parts, " · ")
}

// WorkStillRunning reports whether st says background work outlives the turn.
//
// It is deliberately separate from Busy: an agent can be BOTH idle and carrying live
// work, and that combination is the one that gets dispatched over. Anything reading
// this for a dispatch decision wants it regardless of the base verdict.
func (v Verdict) WorkStillRunning() bool { return v.Shells > 0 }

// MisleadinglyIdle reports the shape a coordinator must not read as free: st calls
// the agent idle, and st also says work is still running in it.
func (v Verdict) MisleadinglyIdle() bool {
	return v.Word == "idle" && v.WorkStillRunning()
}
