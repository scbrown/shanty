package stread

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// DefaultHarness returns the deployment-wide harness used when a card does not
// name one. The status bar uses this to suppress the expected value while still
// surfacing deviations: repeating the default on every pane is noise, but a
// different harness is operationally useful.
//
// SHANTY_ROOT is the same deployment root st uses. When it is absent, follow
// st's user-level deployment pointer. A deployment with neither mechanism is an
// old/default configuration, whose harness is Claude.
func DefaultHarness() (string, error) {
	root := strings.TrimSpace(os.Getenv("SHANTY_ROOT"))
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		pointer := filepath.Join(home, ".config", "shantytown", "root")
		data, err := os.ReadFile(pointer)
		if errors.Is(err, os.ErrNotExist) {
			return "claude", nil
		}
		if err != nil {
			return "", err
		}
		root = strings.TrimSpace(string(data))
		if root == "" {
			return "", errors.New("empty shantytown root pointer")
		}
	}

	var cfg struct {
		Harness struct {
			Default string `toml:"default"`
		} `toml:"harness"`
	}
	if _, err := toml.DecodeFile(filepath.Join(root, "shantytown.toml"), &cfg); err != nil {
		return "", err
	}
	if name := strings.TrimSpace(cfg.Harness.Default); name != "" {
		return name, nil
	}
	return "claude", nil
}
