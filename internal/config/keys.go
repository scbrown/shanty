package config

import "fmt"

// Keybinding represents a tmux key binding.
type Keybinding struct {
	Key     string
	Command string
	Comment string
}

// DefaultKeybindings returns byobu-compatible keybindings with ctrl-a prefix.
func DefaultKeybindings() []Keybinding {
	return []Keybinding{
		{Key: "F2", Command: "new-window", Comment: "New window"},
		{Key: "F3", Command: "previous-window", Comment: "Previous window"},
		{Key: "F4", Command: "next-window", Comment: "Next window"},
		{Key: "F5", Command: "source-file ~/.config/shanty/tmux.conf", Comment: "Reload config"},
		{Key: "F6", Command: "detach-client", Comment: "Detach"},
		{Key: "F7", Command: "copy-mode", Comment: "Scrollback mode"},
		{Key: "F8", Command: `command-prompt -I "#W" "rename-window '%%'"`, Comment: "Rename window"},
	}
}

// RenderKeybindings generates tmux keybinding configuration lines.
func RenderKeybindings(bindings []Keybinding) string {
	var out string
	for _, b := range bindings {
		out += fmt.Sprintf("# %s\nbind-key -n %s %s\n\n", b.Comment, b.Key, b.Command)
	}

	// Prefix-based bindings
	out += "# Split panes\n"
	out += `bind-key | split-window -h -c "#{pane_current_path}"` + "\n"
	out += `bind-key - split-window -v -c "#{pane_current_path}"` + "\n\n"

	// Pane navigation
	out += "# Pane navigation\n"
	out += "bind-key Left select-pane -L\n"
	out += "bind-key Right select-pane -R\n"
	out += "bind-key Up select-pane -U\n"
	out += "bind-key Down select-pane -D\n\n"

	// Last window (byobu/screen convention: ctrl-a a or ctrl-a ctrl-a)
	out += "# Last window\n"
	out += "bind-key a last-window\n"
	out += "bind-key C-a last-window\n"

	return out
}
