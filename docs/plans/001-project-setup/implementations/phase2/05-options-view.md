# 05 - Options menu in TUI

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../../02-Implementation.md) — Phase 2, task 2.6

## What to build

Implement the options menu view that displays a list of available actions (refine, translate, summarize, explain, custom) and allows the user to navigate and select one. The view should integrate with the existing TUI root model from Phase 1 slice 02.

Actions don't need to execute yet — they can be stubbed — but the view must be fully navigable and push onto the view stack.

## Acceptance criteria

- [x] `internal/tui/models/options.go` implements the OptionsModel
- [x] Displays list of actions: Refine, Translate, Summarize, Explain, Custom Prompt
- [x] Actions are navigable with keys from `config.TUI.Keybindings` (default: j/k/arrow, enter)
- [x] Selected action is highlighted, others are not
- [x] `enter` on an action pushes the corresponding view onto the stack (stubbed for now)
- [x] `esc`/`q` pops back to the previous view (initial view)
- [x] `clipboard-tui tui --text "hello"` shows text in initial view with options accessible
- [x] View uses theme from `internal/tui/styles/theme.go`
- [x] View is properly integrated into root model's view stack
- [x] Unit tests cover navigation, selection, and view stack behavior

## Blocked by

- Phase 1 slice 02 - TUI displays piped clipboard text
