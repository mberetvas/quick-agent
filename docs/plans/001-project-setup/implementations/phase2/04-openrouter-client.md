# 04 - OpenRouter LLM backend with retry

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../../02-Implementation.md) — Phase 2, tasks 2.5, 2.10

## What to build

Implement the OpenRouter backend for the `LLMClient` interface defined in Phase 1 slice 06. The client should call the OpenRouter API, handle streaming NDJSON responses, and integrate the retry logic from the LLM config.

API keys should be retrieved from the keyring (Phase 1 slice 04) rather than the config file.

## Acceptance criteria

- [ ] `internal/llm/openrouter/client.go` implements the `LLMClient` interface
- [ ] Uses OpenRouter API endpoint `https://openrouter.ai/api/v1/chat/completions`
- [ ] Streams tokens via SSE (Server-Sent Events) with NDJSON parsing
- [ ] Reads `config.OpenRouter` for model, timeout, max_tokens, temperature
- [ ] Retrieves API key from keyring (service: clipboard-tui, user: openrouter_api_key)
- [ ] Implements retry logic from `config.LLM.RetryAttempts` and `config.LLM.RetryBackoff`
- [ ] Retries on transient errors (5xx, rate limiting, network timeouts)
- [ ] Does NOT retry on 4xx client errors or context cancellation
- [ ] `clipboard-tui debug llm --backend openrouter <text>` streams tokens to stdout
- [ ] `HealthCheck(ctx)` verifies API key is set and endpoint is reachable
- [ ] Context cancellation stops the stream cleanly
- [ ] Unit tests with fake HTTP server cover streaming, retry, and error cases

## Blocked by

- Phase 1 slice 04 - API keys round-trip through system keyring
- Phase 1 slice 06 - Ollama smoke: stream a prompt to stdout (for LLMClient interface)
