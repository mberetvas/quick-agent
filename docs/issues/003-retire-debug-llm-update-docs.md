# 003 - Retire debug llm and document action commands

**Type**: AFK

## Parent

CLI action subcommands plan — replaces `debug llm` with user-facing action subcommands.

## What to build

Remove the `debug llm` subcommand now that `refine`, `translate`, `summarize`, and `explain` cover the same LLM smoke-test path with better UX. Keep the other `debug` utilities (`watch-clipboard`, `hotkey`, `spawn-terminal`) unchanged.

Update project documentation so users discover the new commands and are not directed to the removed subcommand. The README usage section should list all four action commands, include a short **Actions** section with representative examples (arg input, stdin pipe, `--copy`, `--language`, `--verbose`), and replace every `debug llm` example with the equivalent action command.

## Acceptance criteria

- [ ] `quick-agent debug llm` is no longer registered; running it prints the standard unknown-command help
- [ ] `debug watch-clipboard`, `debug hotkey`, and `debug spawn-terminal` still work unchanged
- [ ] README command list includes `refine`, `translate`, `summarize`, and `explain`
- [ ] README **Actions** section documents arg input, stdin pipe, `--copy`, `--language`, and `--verbose` with examples
- [ ] All former `debug llm` references in README are removed or replaced
- [ ] `just check` passes

## Blocked by

- [002 - Translate, summarize, explain subcommands + copy](./002-remaining-action-subcommands.md)
