package session

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateConfig(t *testing.T) {
	// Use a temp dir so we don't pollute the real config
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	confPath, err := GenerateConfig()
	if err != nil {
		t.Fatalf("GenerateConfig() error: %v", err)
	}

	data, err := os.ReadFile(confPath)
	if err != nil {
		t.Fatalf("reading generated config: %v", err)
	}
	conf := string(data)

	// Verify header
	if !strings.Contains(conf, "# Shanty") {
		t.Error("expected header comment")
	}

	// Verify prefix key
	if !strings.Contains(conf, "unbind-key C-b") {
		t.Error("expected C-b unbind")
	}
	if !strings.Contains(conf, "set-option -g prefix C-a") {
		t.Error("expected ctrl-a prefix")
	}
	// ctrl-a a and ctrl-a ctrl-a should switch to last window (byobu/screen convention)
	if !strings.Contains(conf, "bind-key a last-window") {
		t.Error("expected bind-key a last-window")
	}
	if !strings.Contains(conf, "bind-key C-a last-window") {
		t.Error("expected bind-key C-a last-window")
	}

	// Verify terminal settings
	if !strings.Contains(conf, "default-terminal 'tmux-256color'") {
		t.Error("expected tmux-256color terminal")
	}
	if !strings.Contains(conf, "xterm-256color:Tc") {
		t.Error("expected xterm-256color:Tc override")
	}
	if !strings.Contains(conf, "mouse on") {
		t.Error("expected mouse on")
	}
	if !strings.Contains(conf, "base-index 1") {
		t.Error("expected base-index 1")
	}
	if !strings.Contains(conf, "history-limit 50000") {
		t.Error("expected history-limit 50000")
	}

	// Verify theme colors (Dracula defaults)
	if !strings.Contains(conf, "#50fa7b") {
		t.Error("expected active border color #50fa7b")
	}
	if !strings.Contains(conf, "#6272a4") {
		t.Error("expected inactive border color #6272a4")
	}

	// Verify status bar uses segment calls
	if !strings.Contains(conf, "#(shanty seg session)") {
		t.Error("expected session segment call")
	}
	if !strings.Contains(conf, "#(shanty seg cpu)") {
		t.Error("expected cpu segment call")
	}
	if !strings.Contains(conf, "#(shanty seg mem)") {
		t.Error("expected mem segment call")
	}
	if !strings.Contains(conf, "#(shanty seg host)") {
		t.Error("expected host segment call")
	}
	if !strings.Contains(conf, "#(shanty seg clock)") {
		t.Error("expected clock segment call")
	}
	if !strings.Contains(conf, "status-interval 5") {
		t.Error("expected status-interval 5")
	}

	// Verify keybindings
	if !strings.Contains(conf, "F2") {
		t.Error("expected F2 binding")
	}
	if !strings.Contains(conf, "source-file ~/.config/shanty/tmux.conf") {
		t.Error("expected F5 to source shanty config")
	}

	// Verify pane borders
	if !strings.Contains(conf, "pane-active-border-style") {
		t.Error("expected pane-active-border-style")
	}
	if !strings.Contains(conf, "pane-border-style") {
		t.Error("expected pane-border-style")
	}
}

func TestGenerateConfigWritesFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	confPath, err := GenerateConfig()
	if err != nil {
		t.Fatalf("GenerateConfig() error: %v", err)
	}

	if !strings.HasSuffix(confPath, "shanty/tmux.conf") {
		t.Errorf("expected config path to end with shanty/tmux.conf, got %q", confPath)
	}

	info, err := os.Stat(confPath)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("expected non-empty config file")
	}
}
