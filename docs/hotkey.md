# Hotkey listener

The daemon and `clipboard-tui debug hotkey` use `internal/hotkey` with [github.com/robotn/gohook](https://github.com/robotn/gohook) for global key-combination detection. This requires **CGO** and a C toolchain on the build machine.

## Build

```bash
# Linux / macOS
export CGO_ENABLED=1
go build -o quick-agent ./cmd/quick-agent

# Windows (PowerShell) — MinGW-w64 or MSVC required
$env:CGO_ENABLED = "1"
go build -o quick-agent.exe ./cmd/quick-agent
```

### Platform dependencies

| Platform | Requirements |
|----------|----------------|
| **Windows** | MinGW-w64 or Visual Studio Build Tools |
| **macOS** | Xcode Command Line Tools (`xcode-select --install`) |
| **Linux** | GCC, X11 dev libs: `sudo apt-get install -y gcc libx11-dev libxtst-dev` |

Without CGO, the package builds for tests (mock detector only); production hotkey registration is unavailable.

## Verify locally

1. Build with `CGO_ENABLED=1` as above.
2. Run: `clipboard-tui debug hotkey`
3. Press your configured combination (default: **Ctrl+Alt+V** on Windows/Linux, **Cmd+Option+V** on macOS).
4. Expect `Hotkey pressed!` on stdout (debounced per `debounce_ms` in config).
5. Press **Ctrl+C** — the process should exit cleanly without hanging.

## Permissions

- **macOS**: System Settings → Privacy & Security → **Accessibility** — allow the terminal or `clipboard-tui` binary if the hook does not fire.
- **Linux**: Requires an X11 session; Wayland is not supported in v1.
- **Windows**: No extra permission step for typical setups.

## Config modifiers and keys

Config `hotkey.modifiers` may use: `ctrl`, `alt`, `shift`, `cmd`, and `option` (normalized to `alt` for gohook’s keycode map).

The main `hotkey.key` and every modifier must exist in gohook’s keycode map (letters `a`–`z`, digits, `f1`–`f12`, arrows, `space`, `enter`, `tab`, `esc`, etc.). Unknown tokens cause `Start` to return a registration error.

## CI

The default CI matrix runs `go build ./...` **without** CGO on all OSes, so `gohook_detector.go` is excluded and unit tests use the mock detector.

A separate **`build-cgo`** job on Ubuntu installs GCC and X11 libraries and compiles the CLI with `CGO_ENABLED=1` to ensure the production detector builds. End-to-end hook testing is manual (needs a display and input focus).
