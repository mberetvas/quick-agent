# 04 - API keys round-trip through system keyring

**Type**: AFK

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../02-Implementation.md) — Phase 1, task 1.4

## What to build

Integrate `github.com/99designs/keyring` so that secrets (initially the OpenRouter API key) are stored in the OS-native keychain rather than the JSON config. Expose `clipboard-tui config set-key <backend>` (reads the secret from stdin or a secure prompt) and `clipboard-tui config get-key <backend>` (prints the stored value, or a "not set" message). The config loader from slice 03 should consult the keyring transparently when an API key is requested.

Plaintext secrets must never be written to the JSON config file. The config struct may hold a reference/handle but not the raw key.

## Acceptance criteria

- [ ] `clipboard-tui config set-key openrouter` stores a secret in the OS keyring
- [ ] `clipboard-tui config get-key openrouter` retrieves the stored secret
- [ ] Missing key produces a clear "not set" message, not a panic
- [ ] Config JSON file never contains the raw secret value
- [ ] Works on the developer's primary platform; documents fallback for other platforms
- [ ] Unit tests use an in-memory keyring backend
- [ ] `config show` does not print secret values (masked or omitted)

## Blocked by

- 03 - Config load, validate & migrate from disk
