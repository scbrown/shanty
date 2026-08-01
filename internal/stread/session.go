package stread

import "strings"

// SessionPrefixes are the session-name prefixes that carry an agent identity.
//
// "shanty-" is shanty's own convention. "st-" is shantytown's: a fleet launched by
// `st` names each agent's session st-<agent>, and shanty is explicitly the bar and
// picker over those panes (`st attach` points SHANTY_TMUX_SOCKET at the fleet
// socket).
//
// This list lives here, next to the rest of shanty's knowledge of shantytown, so
// there is ONE of it. Three separate copies is how the fleet ended up with a blank
// per-agent status bar AND an empty session picker from the same root cause: each
// site knew only shanty's own prefix, and a session it did not recognize produced
// silence rather than a complaint.
var SessionPrefixes = []string{"shanty-", "st-"}

// AgentFromSession derives the agent a session belongs to, or "" when the name
// carries no known prefix.
//
// "" means "not ours to claim" — a foreign launcher's pane. Callers must not guess
// past it: attributing a session to the wrong agent puts another agent's plate on
// this bar.
func AgentFromSession(session string) string {
	for _, p := range SessionPrefixes {
		if strings.HasPrefix(session, p) {
			return strings.TrimPrefix(session, p)
		}
	}
	return ""
}

// IsAgentSession reports whether the session name carries an agent identity.
func IsAgentSession(session string) bool {
	return AgentFromSession(session) != ""
}

// StripSessionPrefix removes a recognized prefix for DISPLAY, and reports whether
// one was found. It differs from AgentFromSession in what it does with a bare
// prefix: `shanty-` alone strips to "" here (there is nothing left to show) where
// AgentFromSession refuses it (there is no agent to attribute work to). Display and
// attribution are different questions and a bare prefix is exactly where they part.
func StripSessionPrefix(session string) (string, bool) {
	for _, p := range SessionPrefixes {
		if strings.HasPrefix(session, p) {
			return strings.TrimPrefix(session, p), true
		}
	}
	return session, false
}
