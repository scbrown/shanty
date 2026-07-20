package segments

import (
	"fmt"
	"os"
	"strings"
)

// sessionPrefix is shanty's own session-naming prefix; the agent identity is the
// session name with it stripped (shanty-weaver -> weaver).
const sessionPrefix = "shanty-"

// sessionName is the tmux session the segment is being drawn in, set by the seg
// command from tmux's #{session_name}. It is the per-pane identity fallback the
// SHARED fleet bar needs: one tmux server has one $SHANTY_AGENT env, but many
// sessions, so a global status-right rendered $SHANTY_AGENT-based segments blank
// for everyone. Deriving the agent from the session each segment is drawn in
// lets ONE bar show each pane its OWN anchor/events/inbox/harness.
var sessionName string

// SetSession records the session the current render is for (the seg command
// passes it from #{session_name}).
func SetSession(s string) {
	sessionName = s
}

// agentName resolves this agent's shantytown identity.
//
// $SHANTY_AGENT wins — it is the same variable st itself uses, so a pane that
// exports its own identity is authoritative. Otherwise the identity is DERIVED
// from the session name by stripping shanty's own prefix (shanty-weaver ->
// weaver): that is how a shared bar over many sessions gives each its own
// segments. A session that carries no shanty- prefix (a foreign launcher's pane)
// derives no agent here and the segment simply renders nothing — the same safe
// empty every segment already falls back to. An empty result still means "we
// cannot tell who we are"; guessing would put another agent's plate on this bar.
func agentName() string {
	if a := os.Getenv("SHANTY_AGENT"); a != "" {
		return a
	}
	if strings.HasPrefix(sessionName, sessionPrefix) {
		return strings.TrimPrefix(sessionName, sessionPrefix)
	}
	return ""
}

// stReady reports whether we can ask st anything at all: the binary must be on
// PATH and we must know which agent we are.
func stReady() (string, bool) {
	if !stAvailable() {
		return "", false
	}
	agent := agentName()
	if agent == "" {
		return "", false
	}
	return agent, true
}

// Anchor renders the current plate item from shantytown.
type Anchor struct{}

func (a Anchor) Name() string {
	return "anchor"
}

func (a Anchor) Render() string {
	agent, ok := stReady()
	if !ok {
		return ""
	}
	if val, ok := cache.get("anchor"); ok {
		return val
	}
	// An anchor has no zero form, only empty: st prints nothing when the
	// plate is empty.
	item := runST("anchor", agent, "--short")
	if item == "" {
		cache.set("anchor", "")
		return ""
	}
	result := fmt.Sprintf("#[fg=#f8f8f2]⚓ %s#[default]", item)
	cache.set("anchor", result)
	return result
}

// Crew renders the busy/total worker count.
type Crew struct{}

func (c Crew) Name() string {
	return "crew"
}

func (c Crew) Render() string {
	if !stAvailable() {
		return ""
	}
	if val, ok := cache.get("crew"); ok {
		return val
	}
	count := runST("crew", "--count")
	// Suppress only "0/0" — st's own "nothing judgeable" answer. A crew of
	// "0/9" is NOT hidden: a fully idle crew is real information for a
	// coordinator, and hiding it would make "no crew configured" and "every
	// worker idle" look identical on the bar.
	if count == "" || count == "0/0" {
		cache.set("crew", "")
		return ""
	}
	result := fmt.Sprintf("#[fg=#50fa7b]⚙ %s#[default]", count)
	cache.set("crew", result)
	return result
}

// Events renders the count of undelivered stop events addressed to this agent.
type Events struct{}

func (e Events) Name() string {
	return "events"
}

func (e Events) Render() string {
	agent, ok := stReady()
	if !ok {
		return ""
	}
	if val, ok := cache.get("events"); ok {
		return val
	}
	count := runST("anchor", agent, "--events")
	if count == "" || count == "0" {
		cache.set("events", "")
		return ""
	}
	result := fmt.Sprintf("#[fg=#ff5555]⚠ %s#[default]", count)
	cache.set("events", result)
	return result
}

// Inbox renders this agent's unread message count.
type Inbox struct{}

func (i Inbox) Name() string {
	return "inbox"
}

func (i Inbox) Render() string {
	agent, ok := stReady()
	if !ok {
		return ""
	}
	if val, ok := cache.get("inbox"); ok {
		return val
	}
	// --count is a pure read: it never marks anything read, so polling it
	// from the status bar cannot destroy unread state.
	count := runST("inbox", "--count", agent)
	if count == "" || count == "0" {
		cache.set("inbox", "")
		return ""
	}
	result := fmt.Sprintf("#[fg=#bd93f9]✉ %s#[default]", count)
	cache.set("inbox", result)
	return result
}

// Harness renders the name of the agent runtime backing this agent.
type Harness struct{}

func (h Harness) Name() string {
	return "harness"
}

func (h Harness) Render() string {
	agent, ok := stReady()
	if !ok {
		return ""
	}
	if val, ok := cache.get("harness"); ok {
		return val
	}
	name := runST("anchor", agent, "--harness")
	if name == "" {
		cache.set("harness", "")
		return ""
	}
	// A harness is a NAME, not a duration — render it bare. No unit suffix.
	result := fmt.Sprintf("#[fg=#8be9fd]⏱ %s#[default]", name)
	cache.set("harness", result)
	return result
}
