package segments

import (
	"fmt"
	"os"
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
	name := sessionName
	if name == "" {
		name = tmuxSessionName()
	}
	if name == "" {
		name = "shanty"
	}
	// Strip the internal shanty- prefix for display
	display := strings.TrimPrefix(name, "shanty-")
	return fmt.Sprintf("#[fg=#282a36,bg=#bd93f9,bold] %s #[default]", display)
}

// tmuxSessionName queries tmux for the current session name — the FALLBACK for
// when the seg command was invoked without #{session_name} (an older applied
// bar, or a hand-run `shanty seg session`).
//
// The socket is resolved the same way session.resolveSocketName does, and NOT
// hardcoded to "shanty". Hardcoding it is what broke this segment: the fleet
// runs on its own socket (shantytown declares it — [tmux] socket in
// shantytown.toml, e.g. "gt-ae5f35"), so `tmux -L shanty` failed for every
// pane, Render fell through to the literal "shanty", and every agent's bar
// showed the brand instead of its name. Nothing errored; the fallback string
// is a plausible-looking label, which is why it read as a styling choice
// rather than a broken lookup.
func tmuxSessionName() string {
	socket := os.Getenv("SHANTY_TMUX_SOCKET")
	if socket == "" {
		socket = "shanty"
	}
	out, err := exec.Command("tmux", "-L", socket, "display-message", "-p", "#{session_name}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
