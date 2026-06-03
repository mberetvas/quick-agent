# 03 - Daemon: hotkey → clipboard → TUI spawning

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../../02-Implementation.md) — Phase 2, tasks 2.3, 2.4

## What to build

Implement the daemon that runs in the background, monitors for hotkey presses, reads the current clipboard text from the poller, and spawns a TUI instance with that text via the terminal spawner from slice 02.

This connects the infrastructure from slices 01 (hotkey), 02 (terminal), and Phase 1 slice 05 (clipboard poller) into a complete background workflow.

## Terminal spawner integration (slice 02)

On each hotkey press (after debounce at the hotkey layer):

```go
spawner := terminal.NewSpawner(cfg.Terminal)
exe, err := os.Executable()
if err != nil {
    return err
}
text := poller.LatestText() // or last value from Changes channel — see poller API
spawnErr := spawner.SpawnTUI(ctx, exe, text)
if errors.Is(spawnErr, terminal.ErrUsedFallback) {
    // Terminal missing or launch failed; clipboard saved and opened in default app
    log.Info("TUI spawn used fallback", "error", spawnErr)
} else if spawnErr != nil {
    log.Error("failed to spawn TUI", "error", spawnErr)
}
```

- Use `os.Executable()` for the TUI binary path (same binary as the daemon).
- Do **not** call `Spawn` or `debug spawn-terminal` from the daemon.
- Empty clipboard: skip spawn or log debug (align with poller empty-skip behavior).
- `ErrUsedFallback` is success-with-degradation, not a fatal daemon error.

See [02-terminal-spawner.md](./02-terminal-spawner.md) and [docs/terminal.md](../../../../terminal.md).

## Acceptance criteria

- [ ] `cmd/clipboard-tui/daemon.go` contains the `daemonCmd` Cobra command
- [ ] `internal/daemon/main.go` implements the daemon main loop with `Run(ctx, cfg *config.Config) error`
- [ ] PID file written to `config.Daemon.PIDFile` path
- [ ] Checks for existing instance via PID file on startup
- [ ] Starts clipboard poller from Phase 1 slice 05
- [ ] Starts hotkey listener from slice 01
- [ ] On hotkey press: gets latest text from poller, calls `terminal.NewSpawner(cfg.Terminal).SpawnTUI(ctx, os.Executable(), text)`
- [ ] Handles `terminal.ErrUsedFallback` as info-level log, other spawn errors as errors
- [ ] `clipboard-tui daemon` starts the daemon process
- [ ] `clipboard-tui daemon --stop` sends signal to stop running daemon
- [ ] Clean shutdown on SIGTERM/SIGINT (Ctrl+C)
- [ ] Logging writes to configured log file

## Blocked by

- 01 - Hotkey listener with debug verification
- 02 - Terminal spawner with debug verification
- Phase 1 slice 05 - Clipboard poller emits sanitized changes
