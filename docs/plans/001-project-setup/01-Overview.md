# 01 - Architecture Overview

## Core Concept

A background daemon monitors the system clipboard. When the user presses a configurable hotkey, the daemon spawns a TUI in a new terminal window, pre-loaded with the clipboard content. The user selects an action (refine, translate, summarize, explain), the TUI sends the request to an LLM backend (Ollama or OpenRouter), streams the response, and allows the user to copy the result back to their clipboard.

```
┌─────────────────────────────────────────────────────────────────┐
│                         USER WORKFLOW                                │
├─────────────────────────────────────────────────────────────────┤
│  1. User copies text to clipboard                                      │
│  2. User presses hotkey (Ctrl+Alt+V / Cmd+Option+V)                  │
│  3. Daemon spawns TUI with clipboard text                              │
│  4. User selects action from menu                                     │
│  5. TUI sends request to LLM backend                                  │
│  6. Response streams token-by-token                                  │
│  7. User copies result to clipboard                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      DAEMON PROCESS                                  │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  Clipboard Poller                                             │  │  │
│  │    - Polls every 200-500ms (adaptive)                         │  │  │
│  │    - Detects text changes                                    │  │  │
│  │    - Stores latest snapshot                                  │  │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                 │  │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  Hotkey Listener (robotgo)                                  │  │  │
│  │    - Registers global hotkey                                 │  │  │
│  │    - Debounces rapid presses (300ms default)                 │  │  │
│  │    - Triggers TUI spawn on press                             │  │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                 │  │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  Terminal Spawner                                            │  │  │
│  │    - Detects available terminal                              │  │  │
│  │    - Spawns new window with TUI process                       │  │
│  │    - Pipes clipboard text via stdin                          │  │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                 │  │
│  PID File: ~/.config/clipboard-tui/daemon.pid (single instance) │  │
│  Log File: ~/.config/clipboard-tui/clipboard-tui.log            │  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ stdin pipe (clipboard text)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         TUI PROCESS                                   │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  View Stack (bubbletea)                                       │  │  │
│  │    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌───────┐│  │  │
│  │    │ Initial │───▶│ Options │───▶│  Result │───▶│ Error ││  │  │
│  │    │  View   │    │  View   │    │  View   │    │ View  ││  │  │
│  │    └─────────┘    └─────────┘    └─────────┘    └───────┘│  │  │
│  └─────────────────────────────────────────────────────────────┘  │
│                                                                 │  │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  LLM Client Layer                                            │  │  │
│  │    ┌─────────────┐      ┌─────────────────┐                 │  │  │
│  │    │ Ollama      │      │ OpenRouter      │                 │  │  │
│  │    │ Client      │      │ Client          │                 │  │  │
│  │    │ - HTTP JSON │      │ - HTTP SSE       │                 │  │  │
│  │    │ - /api/generate│    │ - /chat/completions│              │  │  │
│  │    └─────────────┘      └─────────────────┘                 │  │  │
│  └─────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Daemon (`internal/daemon/`)

**Responsibility**: Background service that monitors clipboard and responds to hotkey.

**Subcomponents**:
- **Clipboard Poller** (`internal/clipboard/poller.go`)
  - Adaptive polling: 100ms after change, 500ms idle, max 1s
  - Compares clipboard content to detect changes
  - Stores latest text snapshot in memory

- **Hotkey Listener** (`internal/hotkey/listener.go`)
  - Uses robotgo for cross-platform hotkey detection
  - Configurable hotkey with platform-aware defaults
  - Debouncing: ignores presses within 300ms of each other
  - On press: captures current clipboard, spawns TUI

- **Terminal Spawner** (`internal/terminal/spawner.go`)
  - Detects available terminal emulator per platform
  - Spawns new terminal window
  - Pipes clipboard text to TUI via stdin
  - Platform-specific fallback strategies

### 2. TUI (`internal/tui/`)

**Responsibility**: Interactive terminal UI built with bubbletea.

**Views**:
- **Initial View**: Displays clipboard text + action list (Refine, Translate, Summarize, Explain)
- **Options View**: Action selection menu
- **Result View**: Streaming response display with Copy button
- **Error View**: User-friendly display-only error message with keybindings to retry or go back

**Keybindings**:
- `j` / `↓` / `k` / `↑`: Navigate
- `Enter`: Select
- `Esc` / `q`: Back / Quit
- `c`: Copy result to clipboard
- `Ctrl+C`: Copy result (standard)
- `r`: Retry generation inside Error View

**Features**:
- Token-by-token streaming with 50ms delay (typewriter effect)
- Full response buffering for Copy functionality
- View stack for back navigation
- Theming: dark, light (configured via `tui.theme` config field)

### 3. LLM Clients (`internal/llm/`)

**Interface**:
```go
type LLMClient interface {
    Generate(ctx context.Context, prompt string, stream bool) (<-chan string, error)
}
```

**Implementations**:
- **Ollama Client** (`internal/llm/ollama/client.go`)
  - HTTP endpoint: `/api/generate`
  - Response format: NDJSON with `response` field
  - Connection check on first use

- **OpenRouter Client** (`internal/llm/openrouter/client.go`)
  - HTTP endpoint: `/api/v1/chat/completions`
  - Response format: Server-Sent Events (SSE)
  - API key from system keyring

**Shared Features**:
- Retry logic: 3 attempts with exponential backoff (1s, 2s, 4s)
- Timeout: configurable per-backend (default 30s)
- Concurrency: semaphore-limited to 1 request (configurable)
- Context cancellation support

### 4. Configuration (`internal/config/`)

**Responsibility**: Load, validate, and manage application configuration.

**Storage**:
- Config file: JSON at platform-specific location
- API keys: System keyring (via github.com/99designs/keyring)
- Auto-migration: version field for schema changes

---

## Key Design Decisions

| Decision Point | Choice | Rationale |
|----------------|--------|-----------|
| **TUI Library** | bubbletea | Modern, component-based, active community, built on tview |
| **Clipboard Monitoring** | Polling | Works cross-platform without platform-specific code |
| **Polling Frequency** | Adaptive (100-500ms) | Balances responsiveness and CPU usage |
| **Hotkey Library** | robotgo | Supports Windows, macOS, Linux with single API |
| **IPC Mechanism** | stdin pipe | Simple, works for large text, no filesystem cleanup |
| **API Key Storage** | System keyring | Secure, native per-OS, no plaintext in config |
| **Backend Abstraction** | Interface | Enables swapping backends, easy testing |
| **Config Format** | JSON | Human-readable, standard, easy to edit |
| **CLI Framework** | Cobra | Structured subcommands, flags, auto-help |
| **Build System** | Single binary | Simple deployment, no runtime dependencies |
| **Logging** | logrus | Structured logging, multiple outputs, levels |

---

## Data Flow

### Hotkey Press Flow
```
1. User presses hotkey
2. robotgo detects press
3. Daemon checks debounce timer (300ms)
4. Daemon reads current clipboard content
5. Daemon sanitizes text (validate UTF-8, truncate if >100k)
6. Daemon detects available terminal
7. Daemon spawns terminal with TUI process
8. Daemon pipes clipboard text to TUI stdin
9. TUI reads stdin, displays Initial View
```

### Action Selection Flow
```
1. User selects action (e.g., "Refine")
2. TUI constructs prompt for selected action
3. TUI calls LLMClient.Generate()
4. LLMClient acquires semaphore (concurrency limit)
5. LLMClient sends request to backend
6. Backend streams response tokens
7. LLMClient sends tokens to channel
8. TUI receives tokens, displays with 50ms delay
9. TUI buffers full response
10. User presses 'c' to copy
11. TUI writes buffer to clipboard via atotto/clipboard
```

---

## Error Handling Strategy

### Error Types
- **LLM Timeout**: Retry with backoff, then show error
- **Invalid API Key**: Show error with "Re-enter API key" action
- **Ollama Not Running**: Show error with "Start Ollama" action
- **Clipboard Empty**: Show warning, no action needed
- **Terminal Not Found**: Fallback to file output, show path
- **Hotkey Conflict**: Show error with "Change hotkey" action

### User Error Structure
```go
type UserError struct {
    Title      string // "Connection Failed"
    Message    string // "Could not connect to Ollama at http://localhost:11434"
    Severity   string // "error" | "warning" | "info"
    Err        error  // Internal error details for logging
}
```

---

## Performance Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Daemon CPU (idle) | < 0.1% | System monitor |
| Daemon Memory | < 10MB | System monitor |
| TUI Startup | < 500ms | Time to first render |
| Clipboard Poll | 200-500ms | Adaptive interval |
| LLM Response | Streamed | Token-by-token |
| Concurrent Requests | 1 | Semaphore limit |

---

## Security Considerations

1. **API Keys**: Never stored in config file, always in system keyring
2. **Clipboard Content**: Never logged in full, truncated in debug output
3. **Config Permissions**: Config dir 0700, config file 0600
4. **User Consent**: Explicit consent for LLM processing on first run
5. **Network**: Always HTTPS for OpenRouter, optional HTTPS for Ollama
6. **Temp Files**: Randomized names, immediate deletion after use

---

## Accessibility

1. **Keyboard Navigation**: Full support, no mouse required
2. **Color + Symbols**: Never rely on color alone (e.g., use `>` for selected)
3. **Themes**: Dark, light, high-contrast options
4. **Minimal Terminal**: 80x24, warn if smaller
5. **Screen Readers**: Document compatible terminals (future enhancement)
