package config

import (
	"os"
	"path/filepath"
)

// Dir returns shanty's config directory, creating it if needed.
//
// One function so every writer agrees on the location: the generated tmux.conf,
// the theme file, and the crew-mark registry all live together, and an operator
// looking for "shanty's config" should find all of it in one place.
func Dir() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	dir = filepath.Join(dir, "shanty")
	return dir, os.MkdirAll(dir, 0o755)
}
