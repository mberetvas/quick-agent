# 01 - Project skeleton & CLI scaffolding

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../02-Implementation.md) — Phase 1, tasks 1.1, 1.2

## What to build

Initialize the Go module and create the full directory layout described in the implementation plan. Wire up a Cobra root command with three empty subcommands (`daemon`, `tui`, `config`) so that the binary builds and `clipboard-tui --help` lists them. Subcommands may print "not implemented" and exit 0; this slice is purely about establishing the project shell that every subsequent slice plugs into.

Global persistent flags (`--config`, `--log-level`) should be registered on the root command even if unused, so later slices have somewhere to bind to.

## Acceptance criteria

- [ ] `go.mod` exists with the agreed module path and Go version
- [ ] Directory tree matches the layout in `02-Implementation.md` (cmd, internal/{config,clipboard,hotkey,llm,terminal,tui}, pkg, scripts, docs, .github/workflows)
- [ ] `go build ./...` succeeds with no warnings
- [ ] `clipboard-tui --help` lists `daemon`, `tui`, and `config` subcommands
- [ ] Each subcommand is invokable and exits 0 (stub output is fine)
- [ ] Persistent flags `--config` and `--log-level` are registered on the root command
- [ ] A minimal CI workflow runs `go build ./...` and `go vet ./...` on push

## Blocked by

None - can start immediately
