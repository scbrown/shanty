package config

import (
	"strings"
	"testing"
)

// Per-agent segments must be passed #{session_name} so the shared bar renders
// each session its own; fleet/host segments must NOT be.
func TestPerAgentSegmentsCarrySessionName(t *testing.T) {
	out := RenderStatusBar(DefaultTheme(), DefaultStatusBar())

	// session carries it too: the pill is this pane's identity, and on a shared bar
	// the session name is the only way it can know whose pane it is on.
	perAgent := []string{"session", "crewid", "task", "stats", "events", "inbox", "harness"}
	for _, name := range perAgent {
		want := " seg " + name + " #{session_name})"
		if !strings.Contains(out, want) {
			t.Errorf("per-agent segment %q must carry the session: want %q in:\n%s",
				name, want, out)
		}
	}
	// crew is fleet-wide; it must render WITHOUT a session argument.
	if !strings.Contains(out, " seg crew)") {
		t.Errorf("crew should render without a session arg, got:\n%s", out)
	}
	if strings.Contains(out, " seg crew #{session_name})") {
		t.Errorf("crew is fleet-wide and must not carry a session")
	}
}

// The generated bar must invoke shanty by ABSOLUTE path.
//
// Regression guard for a total-blackout failure: tmux runs `#(...)` status
// commands from the SERVER's environment. A server started outside a login shell
// had no ~/.local/bin on PATH, so every `#(shanty seg …)` failed to exec and tmux
// rendered the entire status-right as empty spaces — a bar that looked like it had
// nothing to say, on a fleet of twelve busy agents.
func TestSegmentCallsDoNotDependOnPath(t *testing.T) {
	out := RenderStatusBar(DefaultTheme(), DefaultStatusBar())
	if strings.Contains(out, "#(shanty seg ") {
		t.Errorf("segment calls must not rely on PATH resolving 'shanty':\n%s", out)
	}
	if !strings.Contains(out, "#(/") {
		t.Errorf("expected absolute-path segment calls:\n%s", out)
	}
}

// The identity, plate title and stats segments need real room. tmux truncates
// status-right silently, so a budget sized for the old count-only bar would hide
// exactly the fields this layout exists to show.
func TestStatusRightBudgetFitsTheDefaultLayout(t *testing.T) {
	out := RenderStatusBar(DefaultTheme(), DefaultStatusBar())
	if !strings.Contains(out, "status-right-length 200") {
		t.Errorf("expected a 200-cell right budget:\n%s", out)
	}
}
