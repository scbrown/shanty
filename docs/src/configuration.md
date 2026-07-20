# Configuration

Shanty is designed to work with zero configuration. When you run `shanty`, it generates a tmux config at `~/.config/shanty/tmux.conf` and uses it automatically.

## Config location

Shanty follows the XDG Base Directory Specification:

- Config directory: `$XDG_CONFIG_HOME/shanty/` (default: `~/.config/shanty/`)
- Generated tmux config: `~/.config/shanty/tmux.conf`
- Custom theme: `~/.config/shanty/themes/dracula.toml`

## How config generation works

Each time shanty launches or attaches to a session, it:

1. Loads the theme (custom file or built-in Dracula)
2. Generates a tmux config with keybindings, theme colors, and status bar
3. Writes it to `~/.config/shanty/tmux.conf`
4. Tells tmux to use this config

This means changes to your theme file take effect on the next attach or launch — no restart needed.

## Session names

Sessions are prefixed with `shanty-` internally to avoid collisions with other tmux sessions:

```bash
shanty              # creates tmux session "shanty-main"
shanty -s work      # creates tmux session "shanty-work"
shanty ls           # shows "main", "work" (prefix stripped)
```

## tmux socket

Shanty uses a dedicated tmux socket (`-L shanty`). This means:

- Shanty sessions are isolated from your regular tmux sessions
- Shanty's config doesn't affect other tmux servers
- You can run shanty alongside your existing tmux setup
