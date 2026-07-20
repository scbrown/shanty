# Keybindings

Shanty uses **ctrl-a** as the tmux prefix key (byobu convention), replacing tmux's default ctrl-b.

## Function keys (no prefix needed)

These work without pressing the prefix key first:

| Key | Action |
|-----|--------|
| **F2** | New window |
| **F3** | Previous window |
| **F4** | Next window |
| **F5** | Reload shanty config |
| **F6** | Detach from session |
| **F7** | Enter scrollback / copy mode |
| **F8** | Rename current window |

## Prefix-based keys

Press **ctrl-a** first, then the key:

| Key | Action |
|-----|--------|
| **\|** | Split pane vertically |
| **-** | Split pane horizontally |
| **Left** | Select pane to the left |
| **Right** | Select pane to the right |
| **Up** | Select pane above |
| **Down** | Select pane below |

## Standard tmux keys

All standard tmux keybindings work with the ctrl-a prefix. Some common ones:

| Key | Action |
|-----|--------|
| **ctrl-a d** | Detach |
| **ctrl-a c** | New window |
| **ctrl-a n** | Next window |
| **ctrl-a p** | Previous window |
| **ctrl-a w** | Window list |
| **ctrl-a x** | Kill pane |
| **ctrl-a z** | Zoom pane (toggle fullscreen) |
| **ctrl-a [** | Enter copy mode |
| **ctrl-a ]** | Paste buffer |

## Sending literal ctrl-a

Since ctrl-a is used as the prefix, press it twice to send a literal ctrl-a to the terminal:

```
ctrl-a ctrl-a    # sends ctrl-a to the running program
```
