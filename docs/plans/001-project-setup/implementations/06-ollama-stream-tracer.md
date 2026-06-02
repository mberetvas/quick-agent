# 06 - Ollama smoke: stream a prompt to stdout

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../02-Implementation.md) — Phase 1, tasks 1.5, 1.6, 1.7

## What to build

Define the `LLMClient` interface, implement the Ollama backend, and wire up the prompt template registry — then prove the whole path works through a debug subcommand. `clipboard-tui debug llm <text>` should: load config, resolve the configured backend (Ollama for this slice), select the `refine` prompt template, call `Generate` with streaming enabled, and print tokens to stdout as they arrive.

This is a tracer bullet: it intentionally bypasses the TUI so the LLM path is verifiable on its own and can be reused as a smoke test in later phases. The interface, prompt registry, and client must be designed so that adding OpenRouter (Phase 2) is a drop-in addition with no refactor.

## Acceptance criteria

- [ ] `LLMClient` interface defined with `Generate(ctx, prompt, stream) (<-chan string, error)` and `HealthCheck(ctx) error`
- [ ] Ollama client implements the interface against `/api/generate` with NDJSON streaming
- [ ] Prompt registry exposes templates for `refine`, `translate`, `summarize`, `explain`, `custom`
- [ ] `clipboard-tui debug llm <text>` streams tokens to stdout against a running Ollama
- [ ] Context cancellation (`Ctrl+C`) stops the stream cleanly
- [ ] `HealthCheck` failure surfaces a clear error before any prompt is sent
- [ ] Unit tests cover prompt rendering and Ollama response parsing (with a fake HTTP server)
- [ ] Backend selection is config-driven (no hardcoded URLs)

## Blocked by

- 03 - Config load, validate & migrate from disk
