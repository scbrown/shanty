package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Theme holds the color palette for shanty's tmux configuration.
type Theme struct {
	Name           string `toml:"name"`
	BG             string `toml:"bg"`
	FG             string `toml:"fg"`
	StatusBG       string `toml:"status_bg"`
	Highlight      string `toml:"highlight"`
	ActiveBorder   string `toml:"active_border"`
	InactiveBorder string `toml:"inactive_border"`
	Alert          string `toml:"alert"`
	Warning        string `toml:"warning"`
}

// DefaultTheme returns the built-in Dracula theme.
func DefaultTheme() Theme {
	return Theme{
		Name:           "dracula",
		BG:             "#282a36",
		FG:             "#f8f8f2",
		StatusBG:       "#44475a",
		Highlight:      "#bd93f9",
		ActiveBorder:   "#50fa7b",
		InactiveBorder: "#6272a4",
		Alert:          "#ff5555",
		Warning:        "#ffb86c",
	}
}

// LoadTheme reads a theme from a TOML file. Falls back to DefaultTheme on error.
func LoadTheme(path string) (Theme, error) {
	var t Theme
	if _, err := toml.DecodeFile(path, &t); err != nil {
		return Theme{}, fmt.Errorf("loading theme %s: %w", path, err)
	}
	return t, nil
}

// FindThemeFile locates dracula.toml relative to the binary or in standard paths.
func FindThemeFile() string {
	candidates := []string{
		"themes/dracula.toml",
	}

	// Check relative to executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(dir, "themes", "dracula.toml"),
			filepath.Join(dir, "..", "share", "shanty", "themes", "dracula.toml"),
		)
	}

	// Check XDG config
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		candidates = append(candidates, filepath.Join(xdg, "shanty", "themes", "dracula.toml"))
	} else if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "shanty", "themes", "dracula.toml"))
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}
