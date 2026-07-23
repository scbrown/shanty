# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-07-22

### Added

- `shanty apply` — apply the theme + status bar to the current socket's server
  without attaching; re-runnable after a server restart
- `shanty attach -r/--read-only` — observe a session without keystrokes reaching it
- `SHANTY_TMUX_SOCKET` — point shanty at an existing tmux server/socket instead
  of its own dedicated `-L shanty` socket
- Shantytown segments (anchor, crew, events, inbox, harness) included in the
  default status bar; each self-hides when the `st` CLI or agent identity is absent
- Per-agent segments in a shared bar: segments derive their agent from the
  session name (`shanty-weaver` → `weaver`); `$SHANTY_AGENT` still wins when set
- `shanty ls` is a crew-oriented session selector when shantytown is present:
  agent name, work state, held item, sorted attention-first; falls back to the
  plain session list without `st` or with `--plain`

### Changed

- `attach` prefers a session that exists under the exact name given, applying
  the `shanty-` prefix only when no literal match exists

## [0.1.0] - 2026-02-18

### Added

- Core tmux wrapper with Dracula theme
- ctrl-a prefix key (byobu convention)
- Byobu-compatible F-key bindings (F2-F8)
- Session management: launch, attach, list, named sessions
- Dedicated tmux socket (`-L shanty`) for session isolation
- Pluggable status bar segments via `shanty seg <name>`
- System segments: session, clock, host, cpu, mem, load, disk
- Color-coded resource segments (green/orange/red thresholds)
- 30-second segment cache to reduce overhead
- Custom theme support via TOML files
- Config generation at `~/.config/shanty/tmux.conf`
- `shanty seg list` to discover available segments
- MIT license

[0.2.0]: https://github.com/scbrown/shanty/releases/tag/v0.2.0
[0.1.0]: https://github.com/scbrown/shanty/releases/tag/v0.1.0
