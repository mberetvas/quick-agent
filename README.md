# quick-agent

[![CI](https://github.com/mberetvas/quick-agent/actions/workflows/ci.yml/badge.svg)](https://github.com/mberetvas/quick-agent/actions/workflows/ci.yml)

AI-powered clipboard supercharger. Copy text anywhere, press a global hotkey, pick an action in a terminal UI, and stream an LLM result back to your clipboard.

quick-agent runs as a lightweight background **daemon** that watches the clipboard and listens for a hotkey. When triggered, it opens a new terminal window with a **Bubble Tea** TUI pre-loaded with your clipboard text. Choose refine, translate, summarize, or explain; the response streams token-by-token; copy the result with a single keystroke.

---

## Features

- **Global hotkey** — summon the TUI from any application (requires CGO build; see [Building](#building-from-source))
- **Clipboard-aware daemon** — adaptive polling, sanitization, and size limits
- **Terminal UI** — keyboard-driven view stack: initial → options → streaming result → error
- **Two LLM backends**
  - **Ollama** — local models via `http://localhost:11434`
  - **OpenRouter** — cloud models via OpenRouter API
- **Secure API keys** — OpenRouter (and Ollama) keys stored in the OS keyring, not in plain config
- **Configurable prompts** — override refine/translate/summarize/explain templates in `config.json`
- **Cross-platform** — Linux, macOS, and Windows
- **Terminal auto-detection** — Windows Terminal, PowerShell, cmd, GNOME Terminal, Kitty, and more
- **Fallback output** — if terminal spawn fails, writes result to a file and opens it with the OS handler
- **Debug commands** — clipboard watch, hotkey test, LLM stream tracer, terminal spawn test

---

## Quickstart

### Install with Go

```bash
go install github.com/mberetvas/quick-agent/cmd/quick-agent@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`.

### Download a release binary

Pre-built binaries (with CGO for hotkey support) are attached to [GitHub Releases](https://github.com/mberetvas/quick-agent/releases):

| Platform | Asset |
|----------|-------|
| Linux amd64 | `quick-agent-linux-amd64` |
| macOS amd64 | `quick-agent-darwin-amd64` |
| macOS arm64 | `quick-agent-darwin-arm64` |
| Windows amd64 | `quick-agent-windows-amd64.exe` |

Verify with the included `.sha256` checksum files.

### First run

```bash
# Terminal 1 — start the daemon (foreground)
quick-agent daemon

# Terminal 2 — try the TUI directly with sample text
quick-agent tui --text "Hello from quick-agent"
```

To install the daemon as a login service:

```bash
# Linux (systemd user unit) / macOS (LaunchAgent)
./scripts/install.sh

# Windows (Scheduled Task at logon)
.\scripts\install.ps1
```

---

## Usage

```
quick-agent [flags] <command>

Commands:
  daemon   Start the background clipboard + hotkey daemon
  tui      Launch the terminal UI (stdin or --text)
  config   Show, validate, and manage configuration / API keys
  debug    Developer diagnostics (clipboard, hotkey, LLM, terminal)
  version  Print version
```

Global flags (all subcommands):

| Flag | Description |
|------|-------------|
| `-c`, `--config` | Path to config file (overrides default location) |
| `--log-level` | `debug`, `info`, `warn`, or `error` |

### Daemon

```bash
quick-agent daemon              # run in foreground
quick-agent daemon --stop       # signal a running daemon to stop
```

The daemon:

1. Writes a PID file (default `daemon.pid` under the config directory)
2. Polls the clipboard with adaptive intervals
3. Listens for the configured global hotkey
4. Spawns `quick-agent tui` in a new terminal with the latest clipboard text

### TUI

```bash
quick-agent tui --text "paste this into the UI"
echo "piped content" | quick-agent tui
```

**Keybindings** (defaults, configurable in `config.json`):

| Key | Action |
|-----|--------|
| `Enter` | Continue / select |
| `j` / `k` or arrows | Navigate |
| `Esc` / `q` | Back |
| `c` | Copy result to clipboard |
| `Ctrl+Q` | Quit |

### Config

```bash
quick-agent config show              # resolved config as JSON
quick-agent config validate          # check config semantics
quick-agent config set-key openrouter   # store API key in OS keyring (prompt)
quick-agent config get-key openrouter   # retrieve key (or "not set")
```

---

## Default hotkey

| OS | Default combination |
|----|---------------------|
| Windows / Linux | **Ctrl + Alt + V** |
| macOS | **Cmd + Option + V** |

Debounce defaults to 300 ms. Override in `config.json` under `hotkey`.

**macOS:** grant **Accessibility** permission to the terminal or `quick-agent` binary if the hotkey does not fire.

**Linux:** requires an X11 session for global hooks in v1.

---

## Configuration

### File location

| OS | Config directory | Config file |
|----|------------------|-------------|
| All platforms | `~/.quick-agent/` (`%USERPROFILE%\.quick-agent\` on Windows) | `config.json` |

**Note:** Config used to live under `~/.config/quick-agent/` (Linux/macOS) or `%APPDATA%\quick-agent\` (Windows). Those paths are no longer read automatically — move your files to `~/.quick-agent/` or pass `--config` / `CLIPBOARD_TUI_CONFIG`.

Other paths (defaults):

- PID file: `{config-dir}/daemon.pid`
- Log file: `{config-dir}/quick-agent.log`
- Fallback output: `{config-dir}/output/`

### Environment overrides

Environment variables override file values (file overrides built-in defaults):

| Variable | Overrides |
|----------|-----------|
| `CLIPBOARD_TUI_CONFIG` | Config file path |
| `CLIPBOARD_TUI_BACKEND` | `backend` (`ollama` or `openrouter`) |
| `CLIPBOARD_TUI_OLLAMA_URL` | `ollama.url` |
| `CLIPBOARD_TUI_OLLAMA_MODEL` | `ollama.model` |
| `CLIPBOARD_TUI_HOTKEY_KEY` | `hotkey.key` |
| `CLIPBOARD_TUI_LOG_LEVEL` | `logging.level` |
| `CLIPBOARD_TUI_TERMINAL` | `terminal.emulator` |

### Minimal example

```json
{
  "version": "1",
  "backend": "ollama",
  "ollama": {
    "url": "http://localhost:11434",
    "model": "llama3:8b"
  },
  "hotkey": {
    "modifiers": ["ctrl", "alt"],
    "key": "v",
    "debounce_ms": 300
  }
}
```

---

## LLM backends

### Ollama (local)

1. Install and run [Ollama](https://ollama.com/)
2. Pull a model: `ollama pull llama3:8b`
3. Set `"backend": "ollama"` in config (this is the default)

Test without the daemon:

```bash
quick-agent debug llm "Summarize this paragraph" --backend ollama
```

### OpenRouter (cloud)

1. Create an API key at [openrouter.ai](https://openrouter.ai/)
2. Store it securely:

   ```bash
   quick-agent config set-key openrouter
   ```

3. Set `"backend": "openrouter"` and choose a `openrouter.model` (e.g. `mistralai/mistral-7b-instruct`)

```bash
quick-agent debug llm "Hello" --backend openrouter
```

### API keys and keyring

Secrets are **not** stored in `config.json`. `quick-agent config set-key <backend>` writes to the OS keyring (Keychain on macOS, Credential Manager on Windows, Secret Service on Linux). Service name: `quick-agent`.

---

## Building from source

**Requirements:** Go 1.25+, C toolchain for hotkey support (`CGO_ENABLED=1`).

```bash
git clone https://github.com/mberetvas/quick-agent.git
cd quick-agent

# Build binary (uses just; install from https://github.com/casey/just)
just build

# Or directly:
go build -o quick-agent ./cmd/quick-agent
```

### Platform build dependencies (CGO / hotkey)

| Platform | Packages |
|----------|----------|
| Linux | `gcc`, `libx11-dev`, `libxtst-dev` |
| macOS | Xcode Command Line Tools |
| Windows | MinGW-w64 or Visual Studio Build Tools |

```bash
# Linux example
sudo apt-get install -y gcc libx11-dev libxtst-dev
CGO_ENABLED=1 go build -o quick-agent ./cmd/quick-agent
```

### Development commands

```bash
just test           # unit tests
just check          # fmt + vet + test
just cover-check    # fail if internal/ coverage < 80%
just build-dist     # multi-arch binaries for current OS → dist/
```

---

## Debug utilities

```bash
quick-agent debug watch-clipboard    # print clipboard changes
quick-agent debug hotkey             # print when hotkey fires
quick-agent debug llm "text"         # stream LLM output to stdout
quick-agent debug spawn-terminal --command "echo hello"
```

---

## Architecture (short)

```
Clipboard ──► Daemon (poll + hotkey) ──► spawn terminal ──► TUI ──► LLM backend
                                                      └──► fallback .txt file
```

Packages:

| Path | Role |
|------|------|
| `cmd/quick-agent` | Cobra CLI entrypoint |
| `internal/daemon` | PID file, logging, main loop |
| `internal/clipboard` | Poll, sanitize, size limits |
| `internal/hotkey` | Global hotkey (gohook, CGO) |
| `internal/terminal` | Detect emulator, spawn window |
| `internal/tui` | Bubble Tea application |
| `internal/llm` | Ollama + OpenRouter clients, prompts |
| `internal/config` | Load, migrate, validate, keyring |

---

## Troubleshooting

| Symptom | Check |
|---------|-------|
| Hotkey never fires | Built with `CGO_ENABLED=1`? macOS Accessibility granted? |
| `another instance is already running` | Remove stale `daemon.pid` or run `quick-agent daemon --stop` |
| Ollama connection failed | `ollama serve` running? URL in config correct? |
| OpenRouter 401 | `quick-agent config set-key openrouter` with valid key |
| No terminal opens | Set `terminal.emulator` explicitly (`wt`, `gnome-terminal`, …) |

---

## License

This project is licensed under the terms in [LICENSE](LICENSE).

---

## Contributing

Run `just check` before opening a PR. Coverage gate: `just cover-check` (≥ 80% statement coverage on `internal/...`). Use `just cover-check-all` to include `cmd/` in the report.
