# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `crewid` segment — who a pane belongs to: the agent's mark, name, role, and
  shantytown's own busy/idle verdict, plus a flag when st reports the agent is
  running settings older than the file on disk
- `task` segment — the held item's id *and* its title, so a pane says what the
  agent is working on rather than only which ticket number
- `stats` segment — activity, files touched and token traffic from `st stats`
- Per-crew emoji marks, assigned on first sight and never reassigned, stored in
  `~/.config/shanty/agents.toml`; `shanty marks` lists them and `shanty apply`
  assigns the whole roster in one deterministic pass
- `SHANTY_ST_BIN`, `SHANTY_ST_CWD`, `SHANTY_SEG_NOCACHE` for pointing the
  shantytown segments at the right binary, the right tracker, and around the cache

### Fixed

- **The whole status bar could render as a blank line.** tmux runs `#(...)` status
  commands from the tmux *server's* environment; a server started outside a login
  shell has no `~/.local/bin` on `PATH`, so every `#(shanty seg …)` failed to exec
  and the entire status-right came out as spaces. The generated config now names
  shanty by absolute path, and the segments locate `st` past `PATH` as well.
- **Per-agent segments were blank for every shantytown fleet pane.** Identity was
  derived by stripping only shanty's own `shanty-` prefix, but a fleet launched by
  `st` names its sessions `st-<agent>`. Both prefixes are now recognized.
- **The left pill said `shanty` on every pane.** It queried shanty's own socket for
  the session name, which cannot work when shanty is pointed at a foreign socket
  (`SHANTY_TMUX_SOCKET`); it now uses the session name tmux passes in, and shows
  the agent's mark and name.
- **st's summary prose was parsed as crew.** `st crew` follows its table with lines
  like `9 busy: bond, felix`, whose second field is verdict-shaped — the session
  picker read that as an agent named `9`. A row must now carry a role as well.
- Segments no longer collapse failure to an empty string. Empty is reserved for
  "no `st` installed"; anything else renders a visible `⚠`, including the
  empty-plate-for-a-busy-agent contradiction that a naive `anchor --short` read
  renders as a convincing blank. That state reads `busy, no item` and deliberately
  names no cause: an unreachable tracker and genuinely untracked work both produce
  it, and asserting a diagnosis we have not established is the same species of
  error as rendering blank.
- The segment cache was process-local and therefore never hit — each `shanty seg`
  invocation is its own process. It is now shared on disk.
- **`st` could not reach its own companions.** `st` shells out to its tracker's CLI,
  installed alongside it; under a tmux server with a narrow `PATH` that failed with
  `bd list failed: No such file or directory: 'bd'` even though `st` itself ran.
  `st` is now launched with its own directory prepended to `PATH`.
- **`shanty ls` reported "No active sessions" on a live fleet.** Sessions were
  filtered to shanty's own `shanty-` prefix, so a dozen `st-<agent>` panes were
  invisible — the session picker's version of the blank status bar, from the same
  single-prefix assumption. The prefix list now lives in one place.
- `shanty apply` copies `SHANTY_ST_BIN`/`SHANTY_ST_CWD` into the target tmux
  server's environment. Setting them in a shell and applying used to leave the bar
  unable to reach the tracker, with a blank segment as the only symptom.

### Changed

- Default status bar: `crewid task events inbox crew stats harness cpu mem host
  clock`, with the right-hand budget raised to 200 cells so the identity, title
  and stats fields are not silently truncated. `anchor` remains available but is
  superseded by `task`.

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
