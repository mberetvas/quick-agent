# 002 - Translate, summarize, explain subcommands + copy

**Type**: AFK

## Parent

CLI action subcommands plan — completes the four TUI actions as top-level CLI commands.

## What to build

Extend the action runner CLI wiring from slice 001 with the three remaining subcommands: `translate`, `summarize`, and `explain`. Reuse a shared command factory so all four commands expose the same flags.

`translate` must mirror TUI behavior for target language: default to `translate_target_language` from config (falls back to `"English"`), with an optional `--language` flag to override per invocation. The runner must substitute `{{.Language}}` in the translate prompt (fixing the bug that existed in the old `debug llm` path).

Add `--copy` to all four action commands. When set, write the full accumulated result to the system clipboard after streaming completes; surface a clear error if clipboard write fails.

Expand unit tests to cover translate language resolution (config default and flag override), `--copy` writing via a mock clipboard, and at least one happy-path test per newly wired subcommand.

## Acceptance criteria

- [ ] `quick-agent translate "ik ben een dokter"` translates using the config target language
- [ ] `quick-agent translate "bonjour" --language Dutch` overrides the config language for that run
- [ ] `quick-agent summarize "..."` and `quick-agent explain "..."` stream results to stdout
- [ ] All four commands accept `--copy` and write the full result to the clipboard when the flag is set
- [ ] All four commands share `--verbose`, `--backend`, and stdin/arg input behavior from slice 001
- [ ] Translate prompt never contains a literal `{{.Language}}` placeholder in the rendered output
- [ ] Unit tests cover translate language default, `--language` override, and `--copy` without network
- [ ] `just check` and `just cover-check` pass

## Blocked by

- [001 - Refine action CLI tracer](./001-refine-action-cli.md)
