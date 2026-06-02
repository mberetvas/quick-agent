# 05 - Clipboard poller emits sanitized changes

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../02-Implementation.md) — Phase 1, tasks 1.10, 1.11

## What to build

Implement the adaptive clipboard poller and the text sanitization pipeline as a working end-to-end path. Add a hidden/debug subcommand (e.g. `clipboard-tui debug watch-clipboard`) that runs the poller and prints each detected change to stdout — this makes the slice independently demoable without needing the daemon, hotkey, or TUI.

Adaptive polling: 100 ms after a recent change, backing off to the configured idle interval (default 500 ms, max 1 s). Sanitization: validate UTF-8, drop or replace invalid sequences, truncate to the configured max length (default 100k chars), and skip empty/whitespace-only updates.

Polling and idle intervals should come from the config (slice 03), proving the config layer end-to-end.

## Acceptance criteria

- [ ] Poller detects clipboard text changes and emits them on a channel
- [ ] Adaptive interval: faster after a change, slower when idle, capped at the configured max
- [ ] Non-UTF-8 input is rejected or cleaned, not panicked on
- [ ] Text longer than the configured max is truncated with a marker
- [ ] Empty / whitespace-only changes are not emitted
- [ ] `clipboard-tui debug watch-clipboard` prints changes live to stdout
- [ ] Polling intervals are read from the loaded config
- [ ] Unit tests cover sanitization edge cases (invalid UTF-8, oversize, empty)

## Blocked by

- 01 - Project skeleton & CLI scaffolding
- 03 - Config load, validate & migrate from disk
