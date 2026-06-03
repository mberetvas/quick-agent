# 06 - Result view: streaming + copy + LLM integration

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../../02-Implementation.md) — Phase 2, tasks 2.7, 2.8, 2.9

## What to build

Implement the result view that displays streaming LLM output, connects the LLM clients (Ollama from Phase 1 slice 06, OpenRouter from slice 04) to the TUI, and adds a copy-to-clipboard button that copies the final result.

This completes the core workflow: user selects an action in the options view (slice 05), the corresponding prompt is sent to the configured LLM backend, tokens stream into the result view, and the user can copy the result back to their clipboard.

## Acceptance criteria

- [ ] `internal/tui/models/result.go` implements the ResultModel
- [ ] Receives streaming tokens from LLM client and renders them incrementally
- [ ] Connects to LLM client via the interface (supports both Ollama and OpenRouter)
- [ ] Uses prompt templates from `internal/llm/prompts.go` based on selected action
- [ ] Displays a "Copy" button that copies the complete result to clipboard
- [ ] Copy button uses platform clipboard (not the poller's clipboard)
- [ ] Shows streaming indicator (e.g., spinning cursor or "...") while tokens arrive
- [ ] `clipboard-tui tui --text "hello"` → select action → see streaming result → copy works
- [ ] `config.TUI.StreamingDelayMS` controls the artificial delay for smooth rendering
- [ ] Uses theme from `internal/tui/styles/theme.go`
- [ ] `esc`/`q` returns to options view, `Ctrl+C` exits cleanly
- [ ] Unit tests cover streaming display, copy functionality, and view transitions

## Blocked by

- 05 - Options menu in TUI
- Phase 1 slice 06 - Ollama smoke: stream a prompt to stdout
- Phase 1 slice 02 - TUI displays piped clipboard text (for theme and view stack)
