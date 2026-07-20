package segments

import (
	"fmt"
	"os"
)

// agentName resolves this agent's shantytown identity from $SHANTY_AGENT, the
// same variable st itself uses. An empty result means "we cannot tell who we
// are", and every segment below treats that exactly like a missing st binary:
// render nothing. Guessing an identity would put another agent's plate, inbox
// and event count on this operator's status bar.
func agentName() string {
	return os.Getenv("SHANTY_AGENT")
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
