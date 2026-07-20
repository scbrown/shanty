# Status Bar Segments

The shanty status bar is built from pluggable segments. Each segment is a small Go module that renders a short string for the tmux status bar.

## How segments work

tmux calls `shanty seg <name>` at the configured interval (default: 5 seconds). The segment renders its output — optionally with tmux color codes — and tmux displays it in the status bar.

```bash
# Render a single segment (for testing)
shanty seg cpu

# List all available segments
shanty seg list
```

## Default layout

- **Left:** session name
- **Right:** cpu, mem, host, clock

## System segments

### session

Displays the current tmux session name with Dracula purple styling.

### clock

Current time in `HH:MM` format.

### host

System hostname.

### cpu

CPU usage as a percentage, color-coded by threshold. Samples `/proc/stat` over a 200ms window.

### mem

Memory usage as a percentage, color-coded. Reads from `/proc/meminfo`.

### load

1-minute load average from `/proc/loadavg`.

### disk

Root partition (`/`) usage as a percentage, color-coded.

## shantytown segments

Five further segments surface state from [shantytown](https://github.com/scbrown/shantytown),
a multi-agent workspace manager. They shell out to shantytown's `st` CLI, so they
require `st` on your `PATH` and render an empty string without it — they are inert
if you don't use shantytown.

| Segment | Description | Example |
|---------|-------------|---------|
| `anchor` | Current plate item | `⚓ st-1` |
| `crew` | Busy / total workers | `⚙ 3/9` |
| `events` | Undelivered stop events for you | `⚠ 2` |
| `inbox` | Unread messages | `✉ 1` |
| `harness` | Agent runtime backing you | `⏱ claude` |

Each hides itself when there is nothing to report — an empty plate, no judgeable
crew, zero pending events, no unread mail — so the bar stays quiet until something
wants your attention. `crew` hides only on `0/0`; an idle-but-present crew such as
`0/9` still shows, because "no crew" and "everyone idle" are different facts.

### Agent identity

Every segment except `crew` is per-agent, and resolves which agent you are from
`$SHANTY_AGENT` — the same variable `st` itself uses. If it is unset these segments
render empty rather than guess, so a mis-set environment shows you nothing instead
of showing you somebody else's plate.

Results are cached for 30 seconds, so a 5-second status interval does not mean an
`st` invocation every 5 seconds. `inbox --count` and `anchor --events` are pure
reads — polling them from the status bar never marks a message read or an event
delivered.

## Caching

Segments that call external commands cache their results for 30 seconds to minimize overhead. System segments that read from `/proc` are not cached since those reads are fast.
