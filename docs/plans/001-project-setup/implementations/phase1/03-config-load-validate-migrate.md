# 03 - Config load, validate & migrate from disk

**Type**: HITL

## Parent

[docs/plans/001-project-setup/02-Implementation.md](../02-Implementation.md) — Phase 1, task 1.3

## What to build

Implement the configuration system end-to-end: struct definitions, JSON load from the platform-specific path, schema validation, default values for missing fields, and version-based migration scaffolding. Expose this through a `clipboard-tui config show` subcommand that prints the resolved config (including defaults) so the slice is verifiable from the CLI.

The exact schema and on-disk path are decisions that will outlive v1, so this slice is HITL: the schema should be reviewed before publishing. Migration logic only needs the framework (read `version` field, dispatch to migrators, write back) — no real migrations exist yet, but adding one later must not require restructuring.

See `03-Configuration.md` for schema details.

## Acceptance criteria

- [ ] `Config` struct defined with all fields from `03-Configuration.md`
- [ ] Loader reads JSON from the platform-specific config path
- [ ] Missing file produces a config populated entirely from defaults (no error)
- [ ] Validation rejects malformed values (e.g., negative timeouts) with a clear error
- [ ] `version` field present; migration dispatcher in place even if no migrators registered
- [ ] `clipboard-tui config show` prints the resolved config as JSON
- [ ] `--config <path>` global flag overrides the default path
- [ ] Unit tests cover: defaults, validation failures, migration dispatch

## Blocked by

- 01 - Project skeleton & CLI scaffolding
