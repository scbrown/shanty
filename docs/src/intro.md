# Introduction

Shanty is a terminal multiplexer wrapper — like byobu is to tmux, but rebuilt as a standalone Go binary with the Dracula color palette and pluggable status bar segments.

## What problem does it solve?

tmux is an excellent terminal multiplexer, but getting it to look good and feel productive requires significant configuration: colors, keybindings, status bar scripts, and more. byobu solved this years ago by wrapping tmux with sensible defaults, but byobu is a collection of shell scripts that can be difficult to extend or customize.

Shanty takes the best ideas from byobu and packages them into a single binary:

- **Instant productivity** — launch `shanty` and get a fully themed, well-configured tmux session
- **Familiar keybindings** — F2 for new window, F3/F4 for prev/next, ctrl-a prefix
- **Pluggable segments** — the status bar is built from small Go modules, easy to add new ones
- **Session isolation** — shanty uses its own tmux socket, so it won't interfere with other tmux servers

## Quick example

```bash
# Launch a shanty session
shanty

# Named sessions for different contexts
shanty -s work
shanty -s monitoring

# List all shanty sessions
shanty ls
```

## Requirements

- Go 1.24+ (for building from source)
- tmux (runtime dependency)
