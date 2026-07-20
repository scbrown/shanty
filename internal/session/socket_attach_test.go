package session

import (
	"os/exec"
	"strings"
	"testing"
)

// resolveSocketName honours SHANTY_TMUX_SOCKET so a caller can point shanty at a
// fleet's existing socket, and defaults to shanty's own server otherwise. This
// is what makes `st attach` reach the fleet: before it, socketName
// a const "shanty", so shanty only ever saw its own sessions.
func TestResolveSocketName(t *testing.T) {
	t.Setenv("SHANTY_TMUX_SOCKET", "")
	if got := resolveSocketName(); got != "shanty" {
		t.Errorf("default socket = %q, want %q", got, "shanty")
	}
	t.Setenv("SHANTY_TMUX_SOCKET", "gt-fleet")
	if got := resolveSocketName(); got != "gt-fleet" {
		t.Errorf("override socket = %q, want %q", got, "gt-fleet")
	}
}

// Attach must find a session by its LITERAL name — a fleet pane is named by
// whoever created it (legacy-worker-3), not by shanty's convention, so
// force-prefixing shanty- would miss every foreign session. `dev` with no
// literal match still resolves to shanty-dev.
func TestAttachPrefersLiteralSession(t *testing.T) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed")
	}
	// An isolated socket for the test, torn down after — never the real fleet.
	sock := "shanty-test-attach"
	run := func(args ...string) {
		_ = exec.Command("tmux", append([]string{"-L", sock}, args...)...).Run()
	}
	// Point the package socket at our throwaway server for this test.
	old := socketName
	socketName = sock
	defer func() { socketName = old }()
	defer run("kill-server")

	// A session named with NO shanty- prefix, as a foreign launcher would.
	if err := exec.Command("tmux", "-L", sock, "new-session", "-d", "-s", "legacy-worker-3").Run(); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	m := NewManager()

	// The literal name resolves — the whole point.
	if !m.sessionExists("legacy-worker-3") {
		t.Fatal("literal session not found on the test socket")
	}
	// And a name that does NOT exist literally, and whose prefixed form also does
	// not exist, is a clean not-found error naming the socket — never a raw attach.
	err := m.Attach("nope", false)
	if err == nil {
		t.Fatal("Attach to a missing session must error")
	}
	if !strings.Contains(err.Error(), sock) {
		t.Errorf("error must name the socket: %v", err)
	}
}
