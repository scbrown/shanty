package session

import (
	"strings"
	"testing"
)

func names(cands []Candidate) []string {
	out := make([]string, 0, len(cands))
	for _, c := range cands {
		out = append(out, c.Name)
	}
	return out
}

func eq(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// The headline requirement: the session you just switched AWAY from is the first
// row, which is where the cursor lands — so prefix+s,Enter is a toggle. It must
// win over a more recent attach timestamp, because the client's own notion of
// "last" is the one the operator means.
func TestLastSessionSortsFirstEvenWhenNotTheMostRecentlyAttached(t *testing.T) {
	got := orderCandidates([]Candidate{
		{Name: "alpha", LastAttached: 300},
		{Name: "bravo", LastAttached: 100, IsLast: true},
		{Name: "charlie", LastAttached: 200},
	}, "delta")
	eq(t, names(got), "bravo", "alpha", "charlie")
}

// Switching to the session you are already in is a no-op, and it would occupy
// the row the cursor starts on — the one row that has to be useful.
func TestCurrentSessionIsNotOffered(t *testing.T) {
	got := orderCandidates([]Candidate{
		{Name: "alpha", LastAttached: 300},
		{Name: "bravo", LastAttached: 200},
	}, "alpha")
	eq(t, names(got), "bravo")
}

// Recency here means "when was I last IN it", i.e. session_last_attached. This is
// the property that distinguishes the picker from `choose-tree -O time`, which
// tmux documents as sorting by ACTIVITY — output, not attendance.
func TestOrdersByMostRecentlyAttached(t *testing.T) {
	got := orderCandidates([]Candidate{
		{Name: "old", LastAttached: 100},
		{Name: "newest", LastAttached: 900},
		{Name: "middle", LastAttached: 500},
	}, "")
	eq(t, names(got), "newest", "middle", "old")
}

// A launcher-created fleet has sessions no client has ever attached to. They must
// still be reachable — but below everything with a real attach time, and in a
// STABLE order, so the bottom of the list does not reshuffle between invocations.
func TestNeverAttachedSinkToTheBottomInNameOrder(t *testing.T) {
	got := orderCandidates([]Candidate{
		{Name: "zulu"},
		{Name: "visited", LastAttached: 100},
		{Name: "alpha"},
	}, "")
	eq(t, names(got), "visited", "alpha", "zulu")
}

// An all-fresh fleet is the first-run case: nothing has a timestamp and there is
// no last session. It must still produce a usable, deterministic list.
func TestAllNeverAttachedIsStillOrdered(t *testing.T) {
	got := orderCandidates([]Candidate{
		{Name: "charlie"}, {Name: "alpha"}, {Name: "bravo"},
	}, "")
	eq(t, names(got), "alpha", "bravo", "charlie")
}

// clientContext returns "" for the current session when it cannot ask tmux. An
// empty current must not be treated as a session name and silently filter a real
// session out of the list.
func TestEmptyCurrentDropsNothingReal(t *testing.T) {
	got := orderCandidates([]Candidate{
		{Name: "alpha", LastAttached: 100},
		{Name: "bravo", LastAttached: 200},
	}, "")
	eq(t, names(got), "bravo", "alpha")
}

// The annotation column is decoration; fzf is told to match on field 1 only. The
// tab separator is what makes that split possible, so every line must carry one
// even when there is nothing to annotate — otherwise the name parse on the way
// back out would depend on which row was chosen.
func TestPickerLinesAlwaysCarryTheFieldSeparator(t *testing.T) {
	lines := strings.Split(strings.TrimRight(pickerLines([]Candidate{
		{Name: "bravo", LastAttached: 100, IsLast: true},
		{Name: "alpha", LastAttached: 50},
		{Name: "fresh"},
	}), "\n"), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), lines)
	}
	for _, l := range lines {
		if !strings.Contains(l, "\t") {
			t.Errorf("line %q has no tab separator", l)
		}
		if name, _, _ := strings.Cut(l, "\t"); name == "" {
			t.Errorf("line %q parses to an empty session name", l)
		}
	}
	if !strings.HasPrefix(lines[0], "bravo\t") || !strings.Contains(lines[0], "last") {
		t.Errorf("expected the last session annotated, got %q", lines[0])
	}
	if !strings.Contains(lines[2], "never") {
		t.Errorf("expected the never-attached session annotated, got %q", lines[2])
	}
}
