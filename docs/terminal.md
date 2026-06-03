# Terminal spawner

The daemon and `clipboard-tui debug spawn-terminal` use `internal/terminal` to open a **new terminal window** and run a command. Slice 02 also provides `SpawnTUI`, which launches `clipboard-tui tui` with clipboard text via a temp file redirect.

## Config

```json
"terminal": {
  "emulator": "auto",
  "fallback_dir": ""
}
```

| Field | Description |
|-------|-------------|
| `emulator` | `"auto"` or a profile id (see below) |
| `fallback_dir` | Directory for fallback `.txt` files when spawn fails; default `~/.config/clipboard-tui/output` |

Environment override: `CLIPBOARD_TUI_TERMINAL=wt` sets the emulator for one process (overrides config file).

## Supported emulators

### Windows

| ID | Binary | Notes |
|----|--------|-------|
| `wt` | Windows Terminal | Preferred when installed |
| `powershell` | PowerShell | `Start-Process` new window |
| `cmd` | Command Prompt | Last resort; always on PATH |

### macOS

| ID | App |
|----|-----|
| `terminal` | Terminal.app (`open -a Terminal`) |
| `iterm` | iTerm2 (`open -a iTerm`) |

### Linux

| ID | Binary |
|----|--------|
| `x-terminal-emulator` | Debian alternatives |
| `gnome-terminal` | GNOME |
| `konsole` | KDE |
| `xfce4-terminal` | Xfce |
| `alacritty` | Alacritty |
| `kitty` | Kitty |

With `"emulator": "auto"`, the first profile in the platform list whose launcher is on `PATH` is used.

## Verify locally

### Debug spawn

```bash
go build -o clipboard-tui ./cmd/clipboard-tui

# Windows (PowerShell)
.\clipboard-tui debug spawn-terminal --command "echo hello"

# Optional: force a profile
.\clipboard-tui debug spawn-terminal --command "echo hello" --emulator wt
```

Expect a **new terminal window** running the command. Stdout should print which profile was used.

### SpawnTUI (manual / integration)

After building, trigger `SpawnTUI` from a small test or future daemon:

- Clipboard text appears in the TUI (via temp file → stdin redirect).
- If no emulator is available, a file under `fallback_dir` opens in the default editor and the error satisfies `errors.Is(err, terminal.ErrUsedFallback)`.

### Fallback test

1. Set `"emulator": "nonexistent"` in config (or point `CLIPBOARD_TUI_TERMINAL` at an invalid id).
2. Call `SpawnTUI` with sample text.
3. Confirm `clipboard-<timestamp>.txt` under the fallback directory and that the OS opens it.

## Fallback behavior

When terminal launch fails, `SpawnTUI` only:

1. Writes clipboard text to `<fallback_dir>/clipboard-<unix>.txt`
2. Opens that file with the OS default application (`start` / `open` / `xdg-open`)
3. Returns `ErrUsedFallback` (not a hard failure for the daemon)

`debug spawn-terminal` does **not** use fallback — failures surface directly for debugging.

## Security

- `SpawnTUI` never embeds clipboard text in shell `echo` or argv; it uses a `0600` temp file and stdin redirect.
- Debug `--command` runs a shell one-liner; use only for local testing.

## Platform notes

- **Windows:** Install [Windows Terminal](https://aka.ms/terminal) for the best experience (`wt` profile). `cmd` remains the fallback.
- **macOS:** No extra permissions for spawning terminals; Accessibility applies to hotkeys, not this package.
- **Linux:** Requires a display server (X11 in v1). Headless CI cannot run spawn integration tests without `DISPLAY`.

## CI

Default `go test ./...` exercises profile resolution and argv construction with mocks — no GUI spawn.

Optional `integration` build tag tests may skip when `DISPLAY` is unset (same pattern as hotkey CGO builds).

## See also

- [Phase 2 slice 02 spec](plans/001-project-setup/implementations/phase2/02-terminal-spawner.md)
- [04-Platforms — Terminal Spawning](plans/001-project-setup/04-Platforms.md#terminal-spawning)
