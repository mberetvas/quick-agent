# 001 - Refine action CLI tracer

**Type**: AFK

## Parent

CLI action subcommands plan — grill session decisions (input via arg or stdin, quiet stdout, config-driven backend, sanitization, silent healthcheck).

## What to build

Introduce a shared **action runner** that executes an LLM prompt end-to-end outside the TUI, then prove it with a single user-facing subcommand: `quick-agent refine`.

The runner loads config, resolves input from a positional argument or stdin pipe (args win when both are present), sanitizes text with the same rules as the clipboard daemon, performs a silent healthcheck, renders the `refine` prompt template, streams tokens to stdout at full speed, and returns the accumulated result for testing.

Wire only the `refine` subcommand for now. Shared flags on this command: `--verbose` (print healthcheck, template name, rendered prompt) and `--backend` (override `ollama` / `openrouter`).

Core type shape (decision from plan):

```go
type Options struct {
    Action     string // refine | translate | summarize | explain
    Text       string
    Language   string // translate only; empty → cfg.Prompts.TranslateTargetLanguage
    Copy       bool
    Verbose    bool
    Backend    string
    ConfigPath string
    LogLevel   string
}

func Run(ctx context.Context, opts Options, out io.Writer) (string, error)
```

Implement prompt rendering for all four action names inside the runner (translate uses `RenderWithOptions` with `Language`) even though only `refine` is exposed via CLI — this avoids a refactor when the remaining subcommands land in the next slice.

Handle `SIGINT` / `SIGTERM` via a cancellable context so streaming stops cleanly.

## Acceptance criteria

- [ ] `quick-agent refine "helllo wooorld"` streams the LLM response to stdout against a configured backend
- [ ] `echo "text" | quick-agent refine` works when no positional arg is given
- [ ] Positional arg takes precedence over piped stdin when both are present
- [ ] Whitespace-only or empty input after sanitization returns a clear error
- [ ] Input longer than `clipboard.truncate_size` is truncated with the standard truncate marker
- [ ] Healthcheck runs before generation; status is printed only with `--verbose`
- [ ] `--backend` overrides the config backend for this invocation
- [ ] `Ctrl+C` cancels the stream without a panic
- [ ] Unit tests with a mock `LLMClient` cover refine streaming, verbose output, sanitization errors, and input precedence (no network required)
- [ ] `just check` and `just cover-check` pass

## Blocked by

None — can start immediately.
