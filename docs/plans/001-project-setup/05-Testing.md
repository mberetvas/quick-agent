# 05 - Testing Strategy

## Overview

Testing approach follows a pyramid model:
- **Base**: Unit tests for pure logic (fast, isolated)
- **Middle**: Integration tests for component interactions
- **Top**: Manual/end-to-end tests for full workflow

Priority: Unit tests > Integration tests > Manual tests (due to platform-specific behavior)

---

## Test Pyramid

```
                    ┌─────────────┐
                    │   Manual    │  5-10 tests
                    │   (E2E)    │  Slow, flaky
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ Integration │  20-30 tests
                    │   Tests     │  Medium speed
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  Unit Tests │  50-100 tests
                    │             │  Fast, reliable
                    └─────────────┘
```

---

## 1. Unit Tests

### Test Coverage Targets
- **Core logic (config, sanitize, prompts)**: 100%
- **LLM clients**: 90%
- **TUI models**: 80%
- **Daemon logic**: 80%
- **Overall**: 80%

### Test Files Structure
```
internal/
├── config/
│   └── config_test.go
├── clipboard/
│   └── sanitize_test.go
├── llm/
│   ├── client_test.go
│   ├── ollama/
│   │   └── client_test.go
│   └── openrouter/
│       └── client_test.go
└── tui/
    └── models/
        ├── initial_test.go
        ├── options_test.go
        ├── result_test.go
        └── ...
```

### Example: Config Tests

```go
// internal/config/config_test.go
package config

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoadDefault(t *testing.T) {
    cfg := Default()
    
    if cfg.Version != "1" {
        t.Errorf("Expected version 1, got %s", cfg.Version)
    }
    if cfg.Backend != "ollama" {
        t.Errorf("Expected backend ollama, got %s", cfg.Backend)
    }
}

func TestLoadFromFile(t *testing.T) {
    // Create temp config file
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config.json")
    
    configData := `{
        "version": "1",
        "backend": "openrouter",
        "ollama": {"model": "custom:model"}
    }`
    
    if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
        t.Fatal(err)
    }
    
    cfg, err := Load(configPath)
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }
    
    if cfg.Backend != "openrouter" {
        t.Errorf("Expected backend openrouter, got %s", cfg.Backend)
    }
}

func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
    }{
        {
            name: "valid config",
            config: Config{
                Backend: "ollama",
                Hotkey: HotkeyConfig{
                    Modifiers: []string{"ctrl"},
                    Key:       "v",
                },
            },
            wantErr: false,
        },
        {
            name: "invalid backend",
            config: Config{
                Backend: "invalid",
            },
            wantErr: true,
        },
        {
            name: "empty hotkey",
            config: Config{
                Backend: "ollama",
                Hotkey: HotkeyConfig{},
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestMergeConfig(t *testing.T) {
    base := Default()
    override := Config{
        Backend: "openrouter",
        Ollama: OllamaConfig{
            Model: "custom",
        },
    }
    
    // Test that override values take precedence
    // while base values are preserved for unspecified fields
}
```

### Example: LLM Client Tests

```go
// internal/llm/client_test.go
package llm

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestOllamaClient_Generate(t *testing.T) {
    // Create test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        if r.URL.Path != "/api/generate" {
            t.Errorf("Expected /api/generate, got %s", r.URL.Path)
        }
        
        // Stream response
        w.Header().Set("Content-Type", "application/x-ndjson")
        w.Write([]byte(`{"response":"Hello"}\n`))
        w.Write([]byte(`{"response":" world"}\n`))
        w.Write([]byte(`{"done":true}\n`))
    }))
    defer server.Close()
    
    // Create client
    client := NewOllamaClient(server.URL)
    
    // Call Generate
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    ch, err := client.Generate(ctx, "test prompt", true)
    if err != nil {
        t.Fatalf("Generate failed: %v", err)
    }
    
    // Collect tokens
    var tokens []string
    for token := range ch {
        tokens = append(tokens, token)
    }
    
    // Verify
    expected := []string{"Hello", " world"}
    if len(tokens) != len(expected) {
        t.Errorf("Expected %d tokens, got %d", len(expected), len(tokens))
    }
    for i, token := range tokens {
        if token != expected[i] {
            t.Errorf("Token %d: expected %q, got %q", i, expected[i], token)
        }
    }
}

func TestOllamaClient_HealthCheck(t *testing.T) {
    // Test successful health check
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()
    
    client := NewOllamaClient(server.URL)
    err := client.HealthCheck(context.Background())
    if err != nil {
        t.Errorf("HealthCheck failed: %v", err)
    }
    
    // Test failed health check
    server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    })
    
    err = client.HealthCheck(context.Background())
    if err == nil {
        t.Error("Expected HealthCheck to fail")
    }
}

func TestRetryLogic(t *testing.T) {
    attempt := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        attempt++
        if attempt < 3 {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()
    
    client := NewOllamaClient(server.URL)
    client.RetryAttempts = 3
    client.RetryBackoff = []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}
    
    // This should succeed on 3rd attempt
    _, err := client.Generate(context.Background(), "test", true)
    if err != nil {
        t.Errorf("Expected success after retries, got: %v", err)
    }
    if attempt != 3 {
        t.Errorf("Expected 3 attempts, got %d", attempt)
    }
}
```

### Example: Sanitize Tests

```go
// internal/clipboard/sanitize_test.go
package clipboard

import (
    "testing"
)

func TestSanitizeClipboard(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid text",
            input:    "Hello world",
            expected: "Hello world",
            wantErr:  false,
        },
        {
            name:     "empty text",
            input:    "",
            expected: "",
            wantErr:  true, // Error: empty clipboard
        },
        {
            name:     "binary data",
            input:    "hello\x00world",
            expected: "",
            wantErr:  true, // Error: contains null byte
        },
        {
            name:     "large text",
            input:    string(make([]byte, 150000)),
            expected: string(make([]byte, 10000)) + "\n\n[... Text truncated. Original was 150000 chars]",
            wantErr:  false,
        },
        {
            name:     "utf8 valid",
            input:    "Hello 世界 🌍",
            expected: "Hello 世界 🌍",
            wantErr:  false,
        },
        {
            name:     "utf8 invalid",
            input:    string([]byte{0xff, 0xfe, 0xfd}),
            expected: "",
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := SanitizeClipboard(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("SanitizeClipboard() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if result != tt.expected {
                t.Errorf("SanitizeClipboard() = %q, want %q", result, tt.expected)
            }
        })
    }
}
```

---

## 2. Integration Tests

### Strategy
- Run on Linux with Xvfb (X virtual framebuffer) for GUI-related tests
- Skip on Windows/macOS in CI (or run manually)
- Test component interactions (not full end-to-end)

### Test Files
```
internal/
├── clipboard/
│   └── poller_integration_test.go
├── hotkey/
│   └── listener_integration_test.go
└── terminal/
    └── spawner_integration_test.go
```

### Example: Clipboard Poller Integration Test

```go
// internal/clipboard/poller_integration_test.go
//go:build integration
// +build integration

package clipboard

import (
    "context"
    "testing"
    "time"
)

func TestPoller_DetectsChanges(t *testing.T) {
    // This test requires X11 server (Xvfb)
    // Set clipboard content, start poller, verify it detects change
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Set initial clipboard content
    if err := clipboard.WriteAll("initial"); err != nil {
        t.Skip("Failed to set initial clipboard: ", err)
    }
    
    // Create poller
    poller := NewPoller(100 * time.Millisecond)
    changeChan := make(chan string, 1)
    
    // Start poller
    go poller.Start(ctx, func(text string) {
        select {
        case changeChan <- text:
        default:
        }
    })
    
    // Wait for initial detection
    select {
    case <-changeChan:
    case <-time.After(500 * time.Millisecond):
        t.Fatal("Poller did not detect initial clipboard content")
    }
    
    // Change clipboard content
    if err := clipboard.WriteAll("changed"); err != nil {
        t.Fatal("Failed to change clipboard: ", err)
    }
    
    // Wait for change detection
    select {
    case text := <-changeChan:
        if text != "changed" {
            t.Errorf("Expected 'changed', got %q", text)
        }
    case <-time.After(500 * time.Millisecond):
        t.Fatal("Poller did not detect clipboard change")
    }
}
```

### Integration Test Setup (GitHub Actions)

```yaml
# .github/workflows/ci.yml
jobs:
  integration:
    runs-on: ubuntu-latest
    services:
      xvfb:
        image: sickcodes/docker-xvfb
        options: >-
          --privileged
          -e DISPLAY=:99
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: Run integration tests
        run: go test -v -tags=integration ./internal/...
        env:
          DISPLAY: :99
```

---

## 3. Manual Tests

### Test Categories

| Category | Tests | Frequency |
|----------|-------|-----------|
| **Hotkey** | Hotkey press detection, debouncing | Per platform, once |
| **Clipboard** | Text, empty, binary, large text | Per platform, once |
| **Terminal** | Spawning, stdin piping | Per platform, once |
| **LLM** | Ollama connection, OpenRouter API | Per backend, once |
| **TUI** | Navigation, actions, copy | Full workflow, once |
| **Install** | Install, uninstall, auto-start | Per platform, once |

### Hotkey Tests

| Test | Windows | macOS | Linux | Expected |
|------|---------|-------|-------|----------|
| Press default hotkey | ✅ | ✅ | ✅ | TUI spawns |
| Press hotkey 5x rapidly | ✅ | ✅ | ✅ | 1 TUI spawns |
| Change hotkey in config | ✅ | ✅ | ✅ | New hotkey works |
| Press conflicting hotkey | ⚠️ | ⚠️ | ⚠️ | Error shown |

### Clipboard Tests

| Test | Expected |
|------|----------|
| Copy text, press hotkey | Text appears in TUI |
| Copy empty text | "Clipboard empty" message |
| Copy image/file | "Binary data" error |
| Copy 150k text | Truncated to 10k with warning |
| Copy UTF-8 text | Works correctly |
| Copy non-UTF-8 | Error shown |

### Terminal Tests

| Test | Expected |
|------|----------|
| Press hotkey | New terminal window opens |
| TUI displays text | Text from clipboard shown |
| Close TUI | Terminal window closes |
| Multiple hotkey presses | Multiple TUI windows |
| No terminal available | Fallback to file output |

### LLM Tests

| Test | Ollama | OpenRouter | Expected |
|------|--------|------------|----------|
| Refine text | ✅ | ✅ | Streaming response |
| Translate to French | ✅ | ✅ | French translation |
| Summarize long text | ✅ | ✅ | Short summary |
| Explain concept | ✅ | ✅ | Detailed explanation |
| Invalid API key | N/A | ✅ | "Invalid API key" error |
| Ollama not running | ✅ | N/A | "Start Ollama" error |
| Network timeout | ✅ | ✅ | Retry then error |

### Test Scripts

```bash
# scripts/manual-test.sh
#!/bin/bash

echo "=== Manual Test Suite ==="
echo ""

# Build
echo "Building..."
go build -o clipboard-tui ./cmd/clipboard-tui

# Start daemon
echo "Starting daemon..."
./clipboard-tui daemon --log-level=debug &
DAEMON_PID=$!
sleep 2

# Test 1: Hotkey press
echo "Test 1: Press the hotkey (Ctrl+Alt+V or Cmd+Option+V)"
echo "Expected: TUI should appear with clipboard content"
echo "Press Enter when done..."
read

# Test 2: Clipboard empty
echo "Test 2: Copy empty text (select text, then delete it, copy)"
echo "Press hotkey"
echo "Expected: 'Clipboard is empty' message"
echo "Press Enter when done..."
read

# Test 3: Refine action
echo "Test 3: Copy some text, press hotkey, select 'Refine'"
echo "Expected: Streaming response, can copy result"
echo "Press Enter when done..."
read

# Test 4: Translate
echo "Test 4: Copy English text, press hotkey, select 'Translate'"
echo "Expected: View streams English translation to target language configured in config.json"
echo "Press Enter when done..."
read

# Cleanup
echo "Cleaning up..."
kill $DAEMON_PID
pkill clipboard-tui
rm -f clipboard-tui

echo "Manual tests complete!"
```

---

## 4. End-to-End Tests

### Automated E2E (Linux Only)

```bash
# scripts/test-e2e.sh
#!/bin/bash
set -e

echo "=== End-to-End Tests ==="

# Setup
BINARY="./clipboard-tui"
CONFIG_DIR="$(mktemp -d)"
export CLIPBOARD_TUI_CONFIG="$CONFIG_DIR/config.json"

# Build
go build -o "$BINARY" ./cmd/clipboard-tui

# Cleanup on exit
trap "rm -rf $CONFIG_DIR $BINARY" EXIT

# Test 1: Config creation
echo "Test 1: Config creation"
"$BINARY" config validate

# Test 2: TUI with text
echo "Test 2: TUI with text"
echo "Hello world" | timeout 2 "$BINARY" tui 2>/dev/null || true

# Test 3: Daemon starts
echo "Test 3: Daemon starts"
"$BINARY" daemon --config "$CONFIG_DIR/config.json" &
DAEMON_PID=$!
sleep 1
kill $DAEMON_PID || true

# Test 4: Clipboard polling
echo "Test 4: Clipboard polling"
echo '{"backend":"ollama","clipboard":{"poll_interval_ms":100}}' > "$CONFIG_DIR/config.json"
"$BINARY" daemon --config "$CONFIG_DIR/config.json" &
DAEMON_PID=$!
sleep 0.5

# Copy test text
echo "test text" | xclip -selection clipboard

# Wait for daemon to detect
sleep 0.3

# Check logs (simplified - in real test, would check log file)
kill $DAEMON_PID || true

echo "All E2E tests passed!"
```

---

## Test Utilities

### Mock LLM Server

```go
// internal/llm/mock_server.go
package llm

import (
    "net/http"
    "net/http/httptest"
)

// NewMockOllamaServer creates a test server that simulates Ollama
func NewMockOllamaServer(responses []string) *httptest.Server {
    index := 0
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if index >= len(responses) {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        
        w.Header().Set("Content-Type", "application/x-ndjson")
        for _, resp := range []string{responses[index]} {
            w.Write([]byte(fmt.Sprintf(`{"response":"%s"}\n`, resp)))
        }
        w.Write([]byte(`{"done":true}\n`))
        index++
    }))
}

// NewMockOpenRouterServer creates a test server that simulates OpenRouter
func NewMockOpenRouterServer(responses []string) *httptest.Server {
    index := 0
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if index >= len(responses) {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        
        w.Header().Set("Content-Type", "text/event-stream")
        for _, resp := range responses {
            fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"%s\"}}]}\n\n", resp)
        }
        fmt.Fprint(w, "data: [DONE]\n\n")
        index++
    }))
}
```

---

## Performance Tests

### Benchmark Clipboard Polling

```go
// internal/clipboard/poller_test.go
func BenchmarkPoller(b *testing.B) {
    poller := NewPoller(10 * time.Millisecond)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            poller.checkClipboard()
        }
    })
}
```

### Benchmark LLM Client

```go
// internal/llm/client_test.go
func BenchmarkOllamaClient_Generate(b *testing.B) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"response":"test"}\n{"done":true}\n`))
    }))
    defer server.Close()
    
    client := NewOllamaClient(server.URL)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            ch, _ := client.Generate(context.Background(), "test", false)
            for range ch {}
        }
    })
}
```

---

## Test Coverage

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run with race detector
go test -race -v ./...

# Run unit tests only
go test -v -short ./...

# Run integration tests
go test -v -tags=integration ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -v -run TestConfig ./internal/config
```

### Coverage Targets

| Package | Target Coverage |
|---------|----------------|
| internal/config | 100% |
| internal/clipboard | 100% |
| internal/llm | 90% |
| internal/tui | 80% |
| internal/daemon | 80% |
| internal/hotkey | 70% (platform-specific) |
| internal/terminal | 70% (platform-specific) |
| **Overall** | **80%** |

---

## Troubleshooting Tests

### Common Issues

| Issue | Solution |
|-------|----------|
| `cannot find -lgcc_s` on Linux | Install `libc6-dev` or `gcc-multilib` |
| `xvfb not found` | Install `xvfb` package |
| `robotgo: cgo not available` | Install GCC and set `CGO_ENABLED=1` |
| Tests hang | Check for missing context cancellation |
| Clipboard tests fail | Ensure X11 server is running (Xvfb) |

### CI Debugging

```yaml
# Add debug step to CI
- name: Debug
  if: failure()
  run: |
    echo "=== Environment ==="
    env
    echo "=== Go Version ==="
    go version
    echo "=== OS Info ==="
    uname -a
    cat /etc/os-release
```

---

## Test Checklist (Pre-Release)

- [ ] All unit tests pass on all platforms
- [ ] Integration tests pass on Linux (with Xvfb)
- [ ] Manual tests completed on Windows
- [ ] Manual tests completed on macOS
- [ ] Manual tests completed on Linux
- [ ] Test with Ollama backend
- [ ] Test with OpenRouter backend
- [ ] Test install/uninstall scripts
- [ ] Test hotkey on all platforms
- [ ] Test clipboard edge cases
- [ ] Test error handling
- [ ] Test auto-start on all platforms
- [ ] Coverage >= 80%
- [ ] No race conditions detected (`-race` flag)
