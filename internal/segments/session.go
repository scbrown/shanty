package segments

import (
	"fmt"
	"os/exec"
	"strings"
)

// Session renders the tmux session name with Dracula styling.
// Called by tmux via #(shanty seg session) at status-interval.
type Session struct{}

func (s Session) Name() string {
	return "session"
}

func (s Session) Render() string {
	name := tmuxSessionName()
	if name == "" {
		name = "shanty"
	}
	// Strip the internal shanty- prefix for display
	display := strings.TrimPrefix(name, "shanty-")
	return fmt.Sprintf("#[fg=#282a36,bg=#bd93f9,bold] %s #[default]", display)
}

// tmuxSessionName queries the shanty tmux server for the current session name.
func tmuxSessionName() string {
	out, err := exec.Command("tmux", "-L", "shanty", "display-message", "-p", "#{session_name}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
