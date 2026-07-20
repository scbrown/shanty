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
func DefaultStatusBar() StatusBarConfig {
	return StatusBarConfig{
		Left:  []string{"session"},
		Right: []string{"cpu", "mem", "host", "clock"},
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
	out += "set-option -g status-right-length 80\n"

	// Window status
	out += fmt.Sprintf("set-option -g window-status-current-style 'fg=%s,bg=%s,bold'\n",
		theme.BG, theme.Highlight)
	out += fmt.Sprintf("set-option -g window-status-style 'fg=%s,bg=%s'\n", theme.FG, theme.StatusBG)

	return out
}

// renderSegmentCalls builds tmux format strings that invoke shanty seg for each segment.
func renderSegmentCalls(names []string) string {
	var parts []string
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("#(shanty seg %s)", name))
	}
	return strings.Join(parts, " ")
}
