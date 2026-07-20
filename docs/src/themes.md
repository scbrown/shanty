# Themes

Shanty ships with the [Dracula](https://draculatheme.com/) color palette as its default theme.

## Default Dracula theme

| Element | Color | Hex |
|---------|-------|-----|
| Background | Dark grey | `#282a36` |
| Foreground | White | `#f8f8f2` |
| Status bar background | Grey | `#44475a` |
| Highlights (active tab, session name) | Purple | `#bd93f9` |
| Active pane border | Green | `#50fa7b` |
| Inactive pane border | Comment grey | `#6272a4` |
| Alerts | Red | `#ff5555` |
| Warnings | Orange | `#ffb86c` |

## Custom themes

Create a TOML file at `~/.config/shanty/themes/dracula.toml` to override the built-in theme:

```toml
name = "my-theme"

bg              = "#1e1e2e"
fg              = "#cdd6f4"
status_bg       = "#313244"
highlight       = "#cba6f7"
active_border   = "#a6e3a1"
inactive_border = "#585b70"
alert           = "#f38ba8"
warning         = "#fab387"
```

All fields are required. Colors must be hex format (`#rrggbb`).

## Theme fields

| Field | Used for |
|-------|----------|
| `bg` | General background color |
| `fg` | General foreground / text color |
| `status_bg` | Status bar background |
| `highlight` | Active window tab, session name background |
| `active_border` | Border of the currently focused pane |
| `inactive_border` | Borders of unfocused panes |
| `alert` | Error and alert text (e.g., high CPU) |
| `warning` | Warning text (e.g., moderate resource usage) |

## Resource color coding

Status segments that show percentages (CPU, memory, disk) use three colors from the theme:

- **Green** (active_border color): below 50%
- **Orange** (warning color): 50% to 79%
- **Red** (alert color): 80% and above
