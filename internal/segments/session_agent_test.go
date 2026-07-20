package segments

import (
	"testing"
)

// The shared fleet bar has one $SHANTY_AGENT env but many sessions, so per-agent
// segments must derive identity from the session they are drawn in.
func TestAgentNameFromSession(t *testing.T) {
	t.Setenv("SHANTY_AGENT", "") // unset -> fall back to the session

	SetSession("shanty-weaver")
	if got := agentName(); got != "weaver" {
		t.Errorf("agentName() from shanty-weaver = %q, want %q", got, "weaver")
	}

	SetSession("shanty-sattler")
	if got := agentName(); got != "sattler" {
		t.Errorf("agentName() from shanty-sattler = %q, want %q", got, "sattler")
	}

	// A session with no shanty- prefix derives NO agent — the segment renders the
	// same safe empty it already falls back to, rather than guessing.
	SetSession("some-foreign-session")
	if got := agentName(); got != "" {
		t.Errorf("agentName() from a non-shanty session = %q, want empty", got)
	}
	SetSession("")
}

// $SHANTY_AGENT always wins over the session — a pane that exports its own
// identity is authoritative.
func TestExplicitAgentWinsOverSession(t *testing.T) {
	t.Setenv("SHANTY_AGENT", "explicit")
	SetSession("shanty-weaver")
	if got := agentName(); got != "explicit" {
		t.Errorf("agentName() = %q, want the explicit env value", got)
	}
	SetSession("")
}
