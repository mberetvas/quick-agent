# 01 - Hotkey listener with debug verification

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../../02-Implementation.md) — Phase 2, task 2.1

## What to build

Implement the hotkey listener using robotgo that listens for the configured hotkey combination and emits events on a channel. Expose this through a debug subcommand so the slice is independently verifiable.

The listener should support platform-specific modifier keys (Ctrl/Alt on Windows/Linux, Cmd/Option on macOS), debounce rapid presses using the configured `debounce_ms`, and emit a single event when the full combination is detected.

## Acceptance criteria

- [x] `internal/hotkey/listener.go` implements listener with `Start(ctx, chan<- struct{})` method
- [x] Hotkey configuration read from `config.Hotkey` (modifiers + key + debounce)
- [x] Debounce prevents duplicate events within the configured interval
- [x] Works on developer's primary platform; documents limitations for others
- [x] `clipboard-tui debug hotkey` prints "Hotkey pressed!" to stdout when triggered
- [x] Clean shutdown on context cancellation (Ctrl+C)
- [x] Unit tests cover hotkey combination detection with mock detector

## Blocked by

- Phase 1 (Project skeleton & CLI scaffolding)
