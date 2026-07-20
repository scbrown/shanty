# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.0]: https://github.com/scbrown/shanty/releases/tag/v0.1.0
