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
| **a** / **ctrl-a** | Last window |
| **s** | Session switcher (see below) |

## Session switcher (ctrl-a s)

`ctrl-a s` opens a fuzzy session switcher in a popup instead of tmux's stock
`choose-tree`:

- **Type to filter.** The first keystroke narrows the list — no mode to enter
  first. Matching is fuzzy, so `dta` finds `delta-agent`.
- **Most recently used first.** The session you last switched *away from* is the
  top row, under the cursor, so **ctrl-a s Enter** toggles back to it. Below that,
  sessions are ordered by when you were last in them; sessions no client has ever
  attached to sort last, by name.
- **The current session is not listed** — switching to where you already are is a
  no-op, and it would occupy the row the cursor starts on.
- **Esc** cancels without switching.

Ordering uses tmux's `session_last_attached`, i.e. when you were last *in* a
session. This is deliberately not `choose-tree -O time`, which tmux documents as
sorting by *activity* — on a busy server that ranks whichever session most
recently printed output, which is rarely the one you want.

The picker is `shanty pick`. Run it by hand with `--list` to see the ordering
without switching:

```
shanty pick --list
```

### Requirements

The switcher uses [fzf](https://github.com/junegunn/fzf) for matching. If fzf is
not installed, `ctrl-a s` falls back to tmux's standard `choose-tree -Zs`, so the
key always does something. The check happens when you press the key, so
installing fzf takes effect immediately — no need to regenerate the config.

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
