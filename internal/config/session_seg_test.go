package config

import (
	"strings"
	"testing"
)

// Per-agent segments must be passed #{session_name} so the shared bar renders
// each session its own; fleet/host segments must NOT be.
func TestPerAgentSegmentsCarrySessionName(t *testing.T) {
	out := RenderStatusBar(DefaultTheme(), DefaultStatusBar())

	perAgent := []string{"anchor", "events", "inbox", "harness"}
	for _, name := range perAgent {
		want := "#(shanty seg " + name + " #{session_name})"
		if !strings.Contains(out, want) {
			t.Errorf("per-agent segment %q must carry the session: want %q in:\n%s",
				name, want, out)
		}
	}
	// crew is fleet-wide; it must render WITHOUT a session argument.
	if !strings.Contains(out, "#(shanty seg crew)") {
		t.Errorf("crew should render without a session arg")
	}
	if strings.Contains(out, "#(shanty seg crew #{session_name})") {
		t.Errorf("crew is fleet-wide and must not carry a session")
	}
}
