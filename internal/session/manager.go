package session

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const sessionPrefix = "shanty-"

// socketName is the dedicated tmux socket for shanty sessions.
// Using a separate socket (-L) ensures shanty gets its own tmux server
// with its own config, independent of any other tmux server (e.g., agent
// sessions). Without this, tmux -f is silently ignored when a server
// is already running on the default socket.
const socketName = "shanty"

// fullName returns the tmux session name with the shanty- prefix.
func fullName(name string) string {
	if strings.HasPrefix(name, sessionPrefix) {
		return name
	}
	return sessionPrefix + name
}

// displayName strips the shanty- prefix for user-facing output.
func displayName(name string) string {
	return strings.TrimPrefix(name, sessionPrefix)
}

// Manager handles shanty tmux session lifecycle.
type Manager struct {
	tmuxBin string
}

// NewManager creates a session manager, locating the tmux binary.
func NewManager() *Manager {
	bin, err := exec.LookPath("tmux")
	if err != nil {
		bin = "tmux"
	}
	return &Manager{tmuxBin: bin}
}

// LaunchOrAttach starts a new session or attaches to an existing one.
// If the session exists, attaches (works as new client if already attached).
// If the session doesn't exist, creates it with generated config.
func (m *Manager) LaunchOrAttach(name string) error {
	full := fullName(name)
	if m.sessionExists(full) {
		return m.attach(full)
	}
	return m.create(full)
}

// Attach connects to an existing shanty tmux session by user-facing name.
func (m *Manager) Attach(name string) error {
	full := fullName(name)
	if !m.sessionExists(full) {
		return fmt.Errorf("session %q not found (looked for tmux session %q)", name, full)
	}
	return m.attach(full)
}

// List returns all shanty-managed tmux sessions (shanty- prefix stripped).
func (m *Manager) List() ([]string, error) {
	cmd := exec.Command(m.tmuxBin, "-L", socketName, "list-sessions", "-F", "#{session_name}")
	out, err := cmd.Output()
	if err != nil {
		// tmux returns error when no server is running
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var sessions []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, sessionPrefix) {
			sessions = append(sessions, displayName(line))
		}
	}
	return sessions, nil
}

// attach connects to a tmux session by its full (prefixed) name.
// It regenerates and sources the shanty config so that existing sessions
// pick up theme, keybindings, and status bar changes.
func (m *Manager) attach(fullSessionName string) error {
	// Regenerate config and source it into the shanty server so that
	// prefix, keybindings, theme, and status bar are always applied.
	if confPath, err := GenerateConfig(); err == nil {
		_ = exec.Command(m.tmuxBin, "-L", socketName, "source-file", confPath).Run()
	}

	cmd := exec.Command(m.tmuxBin, "-L", socketName, "attach-session", "-t", fullSessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Manager) sessionExists(fullSessionName string) bool {
	cmd := exec.Command(m.tmuxBin, "-L", socketName, "has-session", "-t", fullSessionName)
	return cmd.Run() == nil
}

func (m *Manager) create(fullSessionName string) error {
	confPath, err := GenerateConfig()
	if err != nil {
		return fmt.Errorf("generating tmux config: %w", err)
	}

	cmd := exec.Command(m.tmuxBin, "-L", socketName, "-f", confPath, "new-session", "-s", fullSessionName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
