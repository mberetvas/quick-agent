# 03 - Daemon: hotkey → clipboard → TUI spawning

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../../02-Implementation.md) — Phase 2, tasks 2.3, 2.4

## What to build

Implement the daemon that runs in the background, monitors for hotkey presses, reads the current clipboard text from the poller, and spawns a TUI instance with that text piped in.

This connects the infrastructure from slices 01 (hotkey), 02 (terminal), and Phase 1 slice 05 (clipboard poller) into a complete background workflow.

## Acceptance criteria

- [ ] `cmd/clipboard-tui/daemon.go` contains the `daemonCmd` Cobra command
- [ ] `internal/daemon/main.go` implements the daemon main loop with `Run(ctx, cfg *config.Config) error`
- [ ] PID file written to `config.Daemon.PIDFile` path
- [ ] Checks for existing instance via PID file on startup
- [ ] Starts clipboard poller from Phase 1 slice 05
- [ ] Starts hotkey listener from slice 01
- [ ] On hotkey press: gets latest text from poller, spawns TUI via terminal spawner from slice 02
- [ ] `clipboard-tui daemon` starts the daemon process
- [ ] `clipboard-tui daemon --stop` sends signal to stop running daemon
- [ ] Clean shutdown on SIGTERM/SIGINT (Ctrl+C)
- [ ] Logging writes to configured log file

## Blocked by

- 01 - Hotkey listener with debug verification
- 02 - Terminal spawner with debug verification
- Phase 1 slice 05 - Clipboard poller emits sanitized changes
