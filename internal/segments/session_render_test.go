package segments

import (
	"strings"
	"testing"
)

// The bug this pins: Session.Render asked tmux for its own name over a
// HARDCODED `-L shanty` socket. The fleet runs on its own socket (shantytown
// declares it: [tmux] socket in shantytown.toml), so the query failed for every
// pane and Render fell through to the literal "shanty" — every agent's bar drew
// the brand name instead of the agent's. The failure was silent and the
// fallback string looked deliberate, so it read as styling, not breakage.
//
// The fix is that the session the segment is being drawn in is PASSED (tmux
// expands #{session_name}), so no socket lookup is needed at all.
func TestSessionRenderUsesPassedSession(t *testing.T) {
	t.Cleanup(func() { SetSession("") })

	SetSession("shanty-sattler")
	got := (Session{}).Render()

	if !strings.Contains(got, " sattler ") {
		t.Errorf("expected the agent name in the segment, got: %q", got)
	}
	if strings.Contains(got, " shanty ") {
		t.Errorf("rendered the brand fallback instead of the passed session: %q", got)
	}
}

// A foreign session (no shanty- prefix) is displayed as-is rather than being
// forced through the prefix strip — the same "we render what we were told"
// rule agentName already follows.
func TestSessionRenderForeignSessionRendersLiterally(t *testing.T) {
	t.Cleanup(func() { SetSession("") })

	SetSession("some-foreign-session")
	got := (Session{}).Render()

	if !strings.Contains(got, " some-foreign-session ") {
		t.Errorf("expected the foreign session name rendered as-is, got: %q", got)
	}
}

// With nothing passed, Render must still fall back rather than panic — this is
// the hand-run `shanty seg session` path and the older-applied-bar path.
func TestSessionRenderFallsBackWhenNothingPassed(t *testing.T) {
	t.Cleanup(func() { SetSession("") })

	SetSession("")
	if got := (Session{}).Render(); got == "" {
		t.Error("expected a non-empty fallback render")
	}
}
