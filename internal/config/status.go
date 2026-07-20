package config

import (
	"fmt"
	"strings"
)

// StatusBarConfig holds the status bar layout.
type StatusBarConfig struct {
	Left  []string
	Right []string
}

// DefaultStatusBar returns the default status bar segment layout.
//
// The shantytown segments lead the right side because they are the ones that
// want you to act: what you hold, who is free, what is waiting on you. CPU and
// memory are ambient and can sit further out.
//
// They are included unconditionally, which is safe and deliberate. Every one
// self-hides unless BOTH the `st` CLI is on PATH and $SHANTY_AGENT names an
// agent, so a user who does not run shantytown sees exactly the old bar and
// pays one `exec` per segment per interval to learn nothing changed. The
// alternative — probing for `st` here and building a different default — would
// make the bar's contents depend on installation order, which is worse than a
// cheap no-op.
func DefaultStatusBar() StatusBarConfig {
	return StatusBarConfig{
		Left:  []string{"session"},
		Right: []string{"anchor", "events", "inbox", "crew", "harness", "cpu", "mem", "host", "clock"},
	}
}

// RenderStatusBar generates tmux status bar configuration.
// All segments are rendered by calling `shanty seg <name>` at status-interval.
func RenderStatusBar(theme Theme, cfg StatusBarConfig) string {
	var out string

	out += "# Status bar\n"
	out += "set-option -g status on\n"
	out += "set-option -g status-interval 5\n"
	out += fmt.Sprintf("set-option -g status-style 'bg=%s,fg=%s'\n", theme.BG, theme.FG)

	// Left status — segments rendered via shanty seg calls
	left := renderSegmentCalls(cfg.Left)
	out += fmt.Sprintf("set-option -g status-left '%s '\n", left)
	out += "set-option -g status-left-length 30\n"

	// Right status — dynamic segments via shanty seg
	right := renderSegmentCalls(cfg.Right)
	out += fmt.Sprintf("set-option -g status-right '%s '\n", right)
	out += "set-option -g status-right-length 140\n"

	// Window status
	out += fmt.Sprintf("set-option -g window-status-current-style 'fg=%s,bg=%s,bold'\n",
		theme.BG, theme.Highlight)
	out += fmt.Sprintf("set-option -g window-status-style 'fg=%s,bg=%s'\n", theme.FG, theme.StatusBG)

	return out
}

// perAgentSegments read a shantytown identity ($SHANTY_AGENT, else the session).
// They get #{session_name} passed so the SHARED bar renders each pane its own —
// the fleet runs one status-right over many sessions. The others (crew, cpu,
// clock, …) are fleet- or host-wide and need no session.
var perAgentSegments = map[string]bool{
	"anchor": true, "events": true, "inbox": true, "harness": true,
}

// renderSegmentCalls builds tmux format strings that invoke shanty seg for each
// segment. A per-agent segment is passed #{session_name}, which tmux expands to
// the session being drawn before running the command (verified: the arg arrives
// as the literal session name), so the segment can derive its own agent.
func renderSegmentCalls(names []string) string {
	var parts []string
	for _, name := range names {
		if perAgentSegments[name] {
			parts = append(parts, fmt.Sprintf("#(shanty seg %s #{session_name})", name))
		} else {
			parts = append(parts, fmt.Sprintf("#(shanty seg %s)", name))
		}
	}
	return strings.Join(parts, " ")
}
