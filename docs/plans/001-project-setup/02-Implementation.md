# 02 - Implementation Details

## Project Structure

```
clipboard-tui/
├── cmd/
│   └── clipboard-tui/
│       ├── main.go          # Cobra root command, entry point
│       ├── daemon.go        # Daemon subcommand implementation
│       ├── tui.go           # TUI subcommand implementation
│       └── config.go        # Config subcommand implementation
│
├── internal/
│   ├── config/
│   │   ├── config.go        # Config struct definitions, loading, validation
│   │   ├── migration.go     # Schema version migration logic
│   │   └── keyring.go       # System keyring integration for API keys
│   │
│   ├── clipboard/
│   │   ├── poller.go        # Adaptive clipboard polling with change detection
│   │   └── sanitize.go      # Text validation, UTF-8 check, truncation
│   │
│   ├── hotkey/
│   │   └── listener.go      # robotgo wrapper, hotkey registration, debouncing
│   │
│   ├── llm/
│   │   ├── client.go        # LLMClient interface definition
│   │   ├── prompts.go       # Prompt templates for each action
│   │   ├── ollama/
│   │   │   └── client.go    # Ollama-specific implementation
│   │   └── openrouter/
│   │       └── client.go    # OpenRouter-specific implementation
│   │
│   ├── terminal/
│   │   └── spawner.go       # Platform-specific terminal detection and spawning
│   │
│   └── tui/
│       ├── app.go           # Root bubbletea model, view stack management
│       ├── keys.go          # Keybinding constants
│       ├── models/
│       │   ├── initial.go       # Initial view: clipboard text + options list
│       │   ├── options.go       # Options menu view
│       │   ├── result.go        # Streaming result display + Copy button
│       │   └── error.go         # Error display with Retry (r) and Back (esc/q) controls
│       └── styles/
│           └── theme.go      # Lipgloss style definitions, theme switching
│
├── pkg/                     # Empty for v1 (future public APIs)
│
├── scripts/
│   ├── build.sh             # Build binaries for all platforms
│   ├── build.ps1            # Windows build script
│   ├── dev.sh               # Development hot-reload (entr)
│   ├── test-e2e.sh          # End-to-end test runner
│   ├── test-e2e.ps1         # Windows e2e tests
│   ├── install.sh           # Unix installer (systemd, launchd)
│   └── install.ps1          # Windows installer (Scheduled Task)
│
├── docs/
│   ├── README.md
│   ├── INSTALL.md
│   ├── CONFIG.md
│   ├── USAGE.md
│   ├── BACKENDS.md
│   ├── TROUBLESHOOTING.md
│   ├── DEVELOPMENT.md
│   └── ARCHITECTURE.md
│
├── .github/
│   └── workflows/
│       ├── ci.yml           # Run tests on PR/push
│       ├── release.yml      # Build and release on tag
│       └── codeql.yml       # Security scanning (optional)
│
├── go.mod
├── go.sum
├── Makefile
├── tools.go                 # Development tool dependencies
└── LICENSE
```

---

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1-2)
**Goal**: Foundation - config, LLM clients, basic TUI, clipboard polling

| Task | File | Estimated Time | Dependencies |
|------|------|---------------|--------------|
| 1.1 Initialize Go module | go.mod | 0.5h | None |
| 1.2 Create directory structure | - | 0.5h | None |
| 1.3 Implement config system | internal/config/config.go | 4h | None |
| 1.4 Implement keyring storage | internal/config/keyring.go | 2h | 1.3 |
| 1.5 Define LLMClient interface | internal/llm/client.go | 1h | None |
| 1.6 Implement Ollama client | internal/llm/ollama/client.go | 4h | 1.5 |
| 1.7 Implement prompt templates | internal/llm/prompts.go | 2h | 1.5 |
| 1.8 Basic TUI root model | internal/tui/app.go | 4h | None |
| 1.9 Initial view | internal/tui/models/initial.go | 4h | 1.8 |
| 1.10 Clipboard poller | internal/clipboard/poller.go | 3h | 1.3 |
| 1.11 Text sanitization | internal/clipboard/sanitize.go | 2h | 1.10 |

**Phase 1 Deliverable**: Can run `clipboard-tui tui --text "hello"` and see text displayed in TUI

---

### Phase 2: Core Workflow (Week 2-3)
**Goal**: Connect daemon, hotkey, TUI, LLM - full end-to-end flow

| Task | File | Estimated Time | Dependencies |
|------|------|---------------|--------------|
| 2.1 Hotkey listener | internal/hotkey/listener.go | 4h | Phase 1 |
| 2.2 Terminal spawner | internal/terminal/spawner.go | 3h | Phase 1 |
| 2.3 Daemon entrypoint | cmd/clipboard-tui/daemon.go | 4h | 2.1, 2.2, 1.10 |
| 2.4 Daemon main logic | internal/daemon/main.go | 4h | 2.3 |
| 2.5 OpenRouter client | internal/llm/openrouter/client.go | 4h | 1.5 |
| 2.6 Options view | internal/tui/models/options.go | 3h | 1.9 |
| 2.7 Result view with streaming | internal/tui/models/result.go | 6h | 1.9, 1.6 |
| 2.8 Copy to clipboard | internal/tui/models/result.go | 2h | 2.7 |
| 2.9 Connect LLM to TUI | internal/tui/app.go | 4h | 2.7, 1.6 |
| 2.10 Retry logic | internal/llm/client.go | 2h | 1.6 |

**Phase 2 Deliverable**: Full workflow works - copy text, press hotkey, select action, see streaming result, copy to clipboard

---

### Phase 3: Polish & Error Handling (Week 3-4)
**Goal**: User experience - error handling, retry mechanism, light theme, custom prompts, installer scripts

| Task | File | Estimated Time | Dependencies |
|------|------|---------------|--------------|
| 3.1 Prompts config schema | internal/config/config.go | 2h | Phase 1 |
| 3.2 Update PromptRegistry | internal/llm/llm.go | 2h | 3.1 |
| 3.3 Simplify actions & events | internal/tui/models/options.go, events.go | 2h | Phase 2 |
| 3.4 Wire configuration prompts | internal/tui/app.go | 2h | 3.2, 3.3 |
| 3.5 UserError structure | internal/errors/errors.go | 2h | None |
| 3.6 ErrorModel view component | internal/tui/models/error.go | 3h | 3.5 |
| 3.7 Wire error view in App | internal/tui/app.go | 3h | 3.6 |
| 3.8 ResultModel retry & error | internal/tui/models/result.go | 3h | 3.7 |
| 3.9 Light Theme & Switching | internal/tui/styles/theme.go | 2h | Phase 1 |
| 3.10 Install Script (sh) | scripts/install.sh | 4h | None |
| 3.11 Install Script (ps1) | scripts/install.ps1 | 3h | None |

**Phase 3 Deliverable**: Robust error view with keyboard retry, custom prompts read from config, a light theme toggle, and OS-specific setup/daemon auto-start options. Tasks 3.9/3.10/3.11 from the original plan (sanitization, logging, and daemon PID locking) are already implemented.

---

### Phase 4: Testing & Release (Week 4)
**Goal**: Quality assurance and distribution

| Task | File | Estimated Time | Dependencies |
|------|------|---------------|--------------|
| 4.1 Unit tests for config | internal/config/config_test.go | 3h | 1.3 |
| 4.2 Unit tests for LLM clients | internal/llm/*_test.go | 4h | 1.6, 2.5 |
| 4.3 Unit tests for sanitize | internal/clipboard/sanitize_test.go | 2h | 1.11 |
| 4.4 Integration tests | scripts/test-e2e.sh | 4h | Phase 2 |
| 4.5 CI pipeline | .github/workflows/ci.yml | 2h | 4.1-4.4 |
| 4.6 Release workflow | .github/workflows/release.yml | 2h | None |
| 4.7 Build scripts | scripts/build.sh, build.ps1 | 2h | None |
| 4.8 Documentation | docs/* | 6h | All phases |
| 4.9 README | README.md | 2h | All phases |
| 4.10 Final testing | - | 4h | All |

**Phase 4 Deliverable**: Tested, documented, releasable v1

---

## Code Organization Principles

### 1. Package Structure
- **cmd/**: Entry points only, minimal logic
- **internal/**: All application logic, not importable by others
- **pkg/**: Future public APIs (empty for v1)

### 2. Layer Separation
```
User Interface Layer (TUI)
    ↓ (calls)
Application Layer (Models, Views)
    ↓ (uses)
Domain Layer (LLM, Clipboard, Hotkey)
    ↓ (uses)
Infrastructure Layer (Config, Logging, Terminal)
```

### 3. Dependency Injection
- Pass dependencies explicitly (e.g., `LLMClient` to models)
- Avoid global state
- Use interfaces for all external dependencies

### 4. Error Handling
- Return errors, don't panic
- Wrap errors with context at boundaries
- Convert to UserError for display

### 5. Testing
- Unit tests for pure logic
- Integration tests for component interactions
- Manual tests for platform-specific features

---

## File Templates

### Main Entry Point (cmd/clipboard-tui/main.go)
```go
package main

import (
    "os"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "clipboard-tui",
    Short: "AI-powered clipboard supercharger",
    // PersistentPreRun for global flags
    // RunE for root command logic (if any)
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    // Add subcommands
    rootCmd.AddCommand(daemonCmd, tuiCmd, configCmd)
    
    // Add global flags
    rootCmd.PersistentFlags().StringP("config", "c", "", "Config file path")
    rootCmd.PersistentFlags().String("log-level", "info", "Log level")
}
```

### LLM Client Interface (internal/llm/client.go)
```go
package llm

import (
    "context"
)

// Client interface for LLM backends
type Client interface {
    // Generate sends prompt to LLM and returns streaming tokens
    Generate(ctx context.Context, prompt string, stream bool) (<-chan string, error)
    
    // HealthCheck verifies backend is available
    HealthCheck(ctx context.Context) error
}

// PromptType defines the type of prompt to generate
type PromptType string

const (
    PromptTypeRefine    PromptType = "refine"
    PromptTypeTranslate PromptType = "translate"
    PromptTypeSummarize PromptType = "summarize"
    PromptTypeExplain   PromptType = "explain"
    PromptTypeCustom    PromptType = "custom"
)

// GetPrompt returns the prompt for a given type and text
func GetPrompt(promptType PromptType, text string, options ...string) string {
    // Implementation returns appropriate prompt template
}
```

### TUI Root Model (internal/tui/app.go)
```go
package tui

import (
    "github.com/charmbracelet/bubbles/viewport"
    "github.com/charmbracelet/bubbletea"
)

type model struct {
    // Current view
    currentView ViewType
    
    // View stack for back navigation
    viewStack []ViewType
    
    // Shared state
    clipboardText string
    llmClient     llm.Client
    config        *config.Config
    
    // View-specific models
    initialModel      *InitialModel
    optionsModel      *OptionsModel
    resultModel       *ResultModel
    errorModel        *ErrorModel
    
    // Dimensions
    width  int
    height int
}

func (m model) Init() tea.Cmd {
    // Initialize all sub-models
    // Return initial command (e.g., focus first element)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Route message to current view
    // Handle view transitions
}

func (m model) View() string {
    // Render current view
}
```

---

## View Implementation Pattern

Each view follows this pattern:

```go
// 1. Define message types for the view
type (
    viewSpecificMsg1 struct{}
    viewSpecificMsg2 struct{}
)

// 2. Define the view model
type ViewNameModel struct {
    // View-specific state
    focusIndex int
    // ...
}

// 3. Implement view-specific methods
func (m *ViewNameModel) Init() tea.Cmd {
    // Initial commands
}

func (m *ViewNameModel) Update(msg tea.Msg) (ViewNameModel, tea.Cmd) {
    // Handle messages
}

func (m *ViewNameModel) View() string {
    // Render view
}

// 4. Integrate into root model
func (m *model) handleViewNameMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Update view-specific model
    newModel, cmd := m.viewNameModel.Update(msg)
    m.viewNameModel = &newModel
    return m, cmd
}
```

---

## Daemon Implementation

### Main Loop (internal/daemon/main.go)
```go
package daemon

import (
    "context"
    "os"
    "os/signal"
    "syscall"
)

func Run(ctx context.Context, cfg *config.Config) error {
    // 1. Check for running instance (PID file)
    if err := checkRunning(); err != nil {
        return err
    }
    
    // 2. Write PID file
    if err := writePIDFile(); err != nil {
        return err
    }
    defer removePIDFile()
    
    // 3. Start clipboard poller
    clipboardChan := make(chan string)
    poller := clipboard.NewPoller(cfg.Clipboard.PollInterval)
    go poller.Start(ctx, clipboardChan)
    
    // 4. Start hotkey listener
    hotkeyChan := make(chan struct{})
    listener := hotkey.NewListener(cfg.Hotkey)
    go listener.Start(ctx, hotkeyChan)
    
    // 5. Main loop
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-hotkeyChan:
            // Hotkey pressed
            text := poller.LatestText()
            if err := spawnTUI(cfg, text); err != nil {
                log.Error("Failed to spawn TUI", "error", err)
            }
        case text := <-clipboardChan:
            // Clipboard changed
            log.Debug("Clipboard changed", "length", len(text))
        }
    }
}

func checkRunning() error {
    // Check PID file, verify process exists
}

func spawnTUI(cfg *config.Config, text string) error {
    // Find terminal, spawn process, pipe text to stdin
}
```
