# 02 - Terminal spawner with debug verification

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../../02-Implementation.md) — Phase 2, task 2.2

## What to build

Implement platform-specific terminal detection and spawning. The spawner should find the user's preferred terminal (from config or environment), construct the appropriate command to open a new window/tab, and execute it with the provided arguments.

Expose through a debug subcommand that demonstrates the spawner works end-to-end.

## Acceptance criteria

- [ ] `internal/terminal/spawner.go` implements spawner with `Spawn(ctx, args []string) error` method
- [ ] Detects terminal from `config.Terminal` or falls back to platform defaults
- [ ] Supports Windows (powershell, wt, cmd), macOS (Terminal.app, iTerm), Linux (xterm, gnome-terminal, konsole, etc.)
- [ ] `clipboard-tui debug spawn-terminal --command "echo hello"` opens terminal running the command
- [ ] Command arguments are properly escaped for the target shell
- [ ] Returns clear error if no terminal is found
- [ ] Works on developer's primary platform; documents fallback behavior
- [ ] Unit tests cover terminal detection and command construction

## Blocked by

- Phase 1 (Project skeleton & CLI scaffolding)
