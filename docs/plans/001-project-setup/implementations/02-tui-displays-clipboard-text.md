# 02 - TUI displays piped clipboard text

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../02-Implementation.md) — Phase 1, tasks 1.8, 1.9

## What to build

Implement the bubbletea root model and the Initial view so that the `tui` subcommand can be fed text and render it. Reading text from stdin is the primary path (matching how the daemon will pipe it later); accept a `--text` flag as a developer-friendly alternative. The Initial view shows the clipboard text and a placeholder action list — the actions don't need to do anything yet, but the list must be focusable and navigable.

A skeleton `internal/tui/styles/theme.go` should exist with at least one default theme so views don't render unstyled. View-stack scaffolding should be in place even if only the Initial view is pushed onto it for now.

This is the Phase 1 deliverable: `echo hello | clipboard-tui tui` (or `clipboard-tui tui --text hello`) shows "hello" in a styled TUI.

## Acceptance criteria

- [ ] `clipboard-tui tui --text "hello"` opens the TUI with "hello" displayed
- [ ] `echo hello | clipboard-tui tui` produces the same result
- [ ] Initial view shows the input text plus a navigable list of actions (stub entries are fine)
- [ ] `j`/`k`/arrow keys move focus in the list
- [ ] `q` and `esc` exit cleanly with no terminal artifacts
- [ ] `Ctrl+C` exits cleanly
- [ ] A default theme is applied via lipgloss (no raw unstyled output)
- [ ] Root model has a view-stack field ready for additional views in later slices

## Blocked by

- 01 - Project skeleton & CLI scaffolding
