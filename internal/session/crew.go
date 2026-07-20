package session

import (
	"os/exec"
	"sort"
	"strings"
	"sync"
)

// CrewRow is one enriched line of the session picker: a session name joined with
// shantytown's own verdict about what that agent is doing. Every judgement here
// is READ from `st`, never re-derived — the picker and the tier must agree on who
// is busy, or the picker becomes a second, disagreeing opinion.
type CrewRow struct {
	Name     string // session/agent display name (shanty- prefix already stripped)
	Item     string // current item held (`st anchor <name> --short`), "" if none/unknown
	State    string // st's work verdict cell (busy / idle / waiting / saturated·948k / ...)
	Currency string // settings currency (current / STALE / unknown), "" if not reported
	rank     int    // attention sort key: lower surfaces first
}

// crewEntry is the slice of `st crew` the picker keeps per agent.
type crewEntry struct {
	state    string
	currency string
}

// stCrewAvailable reports whether we can ask st for crew state at all.
func stCrewAvailable() bool {
	_, err := exec.LookPath("st")
	return err == nil
}

// runST shells out to st and returns trimmed stdout, or "" on any error — the
// same fail-quiet contract shanty's status-bar segments already use. A missing
// binary, a non-zero exit, or an empty plate all collapse to "": the picker
// shows less, never a crash.
func runST(args ...string) string {
	out, err := exec.Command("st", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// stateRank orders st's work verdicts by how much they need a human's eyes. The
// ones an operator must look at FIRST — blocked on a question, wedged, a stalled
// send, a context wall — sort before the ones that are fine. These are st's
// words (triage.py), matched, never recomputed.
var stateRank = map[string]int{
	"waiting":   0, // BLOCKED on a question in the pane — needs a person, will not time out
	"wedged":    1, // dead / stuck
	"queued":    2, // unsubmitted text in the box — a stalled dispatch, or a human mid-sentence
	"saturated": 3, // over the context limit — looks free, is a wall
	"?":         4, // st could not tell
	"busy":      5, // mid-flight — working, leave it
	"idle":      6, // free
}

// stateWord extracts the leading verdict word from a work cell, so "saturated·948k"
// and "busy+1sh" map to "saturated"/"busy". Returns "" if the cell matches no
// known verdict (then the row sorts as a plain, un-judged session).
func stateWord(cell string) string {
	if cell == "" {
		return ""
	}
	if strings.HasPrefix(cell, "?") {
		return "?"
	}
	// Only one verdict can prefix a given cell (the words are disjoint), so map
	// iteration order does not affect the result.
	for word := range stateRank {
		if word != "?" && strings.HasPrefix(cell, word) {
			return word
		}
	}
	return ""
}

// rankOf maps a work cell to its attention rank; unknown/plain sessions sort
// after every judged one but before nothing.
func rankOf(cell string) int {
	if r, ok := stateRank[stateWord(cell)]; ok {
		return r
	}
	return 90
}

// parseCrew turns `st crew`'s table into name -> {state cell, currency}. It reads
// position-tolerantly on purpose — this picker does not own st's column order and
// must not break when it changes: the NAME is always the first field, the STATE
// is the field whose leading word is a known verdict, and CURRENCY is whichever
// field is one of st's three settings values. A line with no recognizable verdict
// is dropped (that session still lists, just without enrichment).
func parseCrew(out string) map[string]crewEntry {
	res := map[string]crewEntry{}
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := fields[0]
		var stateCell, currency string
		for _, f := range fields[1:] {
			if stateCell == "" && stateWord(f) != "" {
				stateCell = f
			}
			switch f {
			case "current", "STALE", "unknown":
				currency = f
			}
		}
		if stateCell == "" {
			continue
		}
		res[name] = crewEntry{state: stateCell, currency: currency}
	}
	return res
}

// buildRows joins the session list with parsed crew state and a per-agent item
// lookup, then sorts attention-first (then by name). Pure: `anchor` is injected
// so the join and the ordering are testable without a live st.
func buildRows(sessions []string, crew map[string]crewEntry, anchor func(name string) string) []CrewRow {
	rows := make([]CrewRow, len(sessions))
	for i, name := range sessions {
		row := CrewRow{Name: name, Item: anchor(name)}
		if e, ok := crew[name]; ok {
			row.State, row.Currency, row.rank = e.state, e.currency, rankOf(e.state)
		} else {
			row.rank = 99 // a live session st does not know as crew — sort last
		}
		rows[i] = row
	}
	sort.SliceStable(rows, func(a, b int) bool {
		if rows[a].rank != rows[b].rank {
			return rows[a].rank < rows[b].rank
		}
		return rows[a].Name < rows[b].Name
	})
	return rows
}

// EnrichedRows produces the crew-oriented picker rows for the given sessions, or
// (nil, false) when st cannot enrich them — a missing st binary OR an st that
// reports no judgeable crew. The caller falls back to the plain name list in that
// case: on a machine with no shantytown, or with only personal sessions, a plain
// list is the honest answer, not an empty table.
func EnrichedRows(sessions []string) ([]CrewRow, bool) {
	if len(sessions) == 0 || !stCrewAvailable() {
		return nil, false
	}
	crew := parseCrew(runST("crew"))
	if len(crew) == 0 {
		return nil, false
	}
	// Fetch each agent's held item concurrently: st is a separate process per
	// call, and a serial loop over a full crew is visibly slow for an on-demand
	// picker. anchor items only appear when st's configured tracker has a plate;
	// where it does not, every Item is "" and the picker still shows state.
	items := make([]string, len(sessions))
	var wg sync.WaitGroup
	for i, name := range sessions {
		i, name := i, name
		wg.Add(1)
		go func() {
			defer wg.Done()
			items[i] = runST("anchor", name, "--short")
		}()
	}
	wg.Wait()
	idx := make(map[string]int, len(sessions))
	for i, n := range sessions {
		idx[n] = i
	}
	return buildRows(sessions, crew, func(name string) string { return items[idx[name]] }), true
}
