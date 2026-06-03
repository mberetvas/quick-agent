# 03 - Configuration & Settings

## Configuration File Schema

```json
{
  "version": "1",
  "backend": "ollama",
  
  "ollama": {
    "url": "http://localhost:11434",
    "model": "llama3:8b",
    "timeout": 30,
    "max_tokens": 2048,
    "temperature": 0.7
  },
  
  "openrouter": {
    "model": "mistralai/mistral-7b-instruct",
    "timeout": 30,
    "max_tokens": 2048,
    "temperature": 0.7
  },
  
  "hotkey": {
    "modifiers": ["ctrl", "alt"],
    "key": "v",
    "debounce_ms": 300
  },
  
  "clipboard": {
    "max_size": 100000,
    "truncate_size": 10000,
    "poll_interval_ms": 500
  },
  
  "tui": {
    "theme": "auto",
    "streaming_delay_ms": 50,
    "keybindings": {
      "navigate_down": ["j", "down"],
      "navigate_up": ["k", "up"],
      "select": ["enter"],
      "back": ["esc", "q"],
      "copy": ["c", "ctrl+c"],
      "quit": ["ctrl+q"]
    }
  },
  
  "llm": {
    "max_concurrent": 1,
    "retry_attempts": 3,
    "retry_backoff": [1000, 2000, 4000]
  },
  
  "logging": {
    "level": "info",
    "file": "~/.config/clipboard-tui/clipboard-tui.log",
    "max_size_mb": 10,
    "max_backups": 5,
    "console": false
  },
  
  "daemon": {
    "pid_file": "~/.config/clipboard-tui/daemon.pid",
    "auto_start": true
  },

  "terminal": {
    "emulator": "auto",
    "fallback_dir": ""
  }
}
```

---

## Configuration Structure

### Root Config

```go
// internal/config/config.go
package config

import (
    "encoding/json"
    "os"
    "path/filepath"
    "runtime"
)

type Config struct {
    Version string `json:"version"`
    Backend string `json:"backend"` // "ollama" or "openrouter"
    Ollama  OllamaConfig `json:"ollama"`
    OpenRouter OpenRouterConfig `json:"openrouter"`
    Hotkey HotkeyConfig `json:"hotkey"`
    Clipboard ClipboardConfig `json:"clipboard"`
    TUI TUIConfig `json:"tui"`
    LLM LLMConfig `json:"llm"`
    Logging LoggingConfig `json:"logging"`
    Daemon DaemonConfig `json:"daemon"`
    Terminal TerminalConfig `json:"terminal"`
}

func Default() *Config {
    return &Config{
        Version: "1",
        Backend: "ollama",
        Ollama:  DefaultOllamaConfig(),
        OpenRouter: DefaultOpenRouterConfig(),
        Hotkey: DefaultHotkeyConfig(),
        Clipboard: DefaultClipboardConfig(),
        TUI: DefaultTUIConfig(),
        LLM: DefaultLLMConfig(),
        Logging: DefaultLoggingConfig(),
        Daemon: DefaultDaemonConfig(),
        Terminal: DefaultTerminalConfig(),
    }
}

func Load(path string) (*Config, error) {
    // Load from file, validate, migrate if needed
}

func Save(cfg *Config, path string) error {
    // Save to file with proper permissions
}

func GetConfigPath() string {
    // Return platform-specific config path
    if runtime.GOOS == "windows" {
        return filepath.Join(os.Getenv("APPDATA"), "clipboard-tui", "config.json")
    }
    return filepath.Join(os.Getenv("HOME"), ".config", "clipboard-tui", "config.json")
}
```

### Backend Configs

```go
type OllamaConfig struct {
    URL         string  `json:"url"`
    Model       string  `json:"model"`
    Timeout     int     `json:"timeout"`     // seconds
    MaxTokens   int     `json:"max_tokens"`
    Temperature float64 `json:"temperature"`
}

func DefaultOllamaConfig() OllamaConfig {
    return OllamaConfig{
        URL:         "http://localhost:11434",
        Model:       "llama3:8b",
        Timeout:     30,
        MaxTokens:   2048,
        Temperature: 0.7,
    }
}

type OpenRouterConfig struct {
    Model       string  `json:"model"`
    Timeout     int     `json:"timeout"`
    MaxTokens   int     `json:"max_tokens"`
    Temperature float64 `json:"temperature"`
}

func DefaultOpenRouterConfig() OpenRouterConfig {
    return OpenRouterConfig{
        Model:       "mistralai/mistral-7b-instruct",
        Timeout:     30,
        MaxTokens:   2048,
        Temperature: 0.7,
    }
}
```

### Other Configs

```go
type HotkeyConfig struct {
    Modifiers   []string `json:"modifiers"` // ["ctrl", "alt", "shift", "cmd"]
    Key         string   `json:"key"`
    DebounceMS  int      `json:"debounce_ms"`
}

func DefaultHotkeyConfig() HotkeyConfig {
    if runtime.GOOS == "darwin" {
        return HotkeyConfig{
            Modifiers:  []string{"cmd", "option"},
            Key:        "v",
            DebounceMS: 300,
        }
    }
    return HotkeyConfig{
        Modifiers:  []string{"ctrl", "alt"},
        Key:        "v",
        DebounceMS: 300,
    }
}

type ClipboardConfig struct {
    MaxSize     int `json:"max_size"`      // chars
    TruncateSize int `json:"truncate_size" // chars
    PollIntervalMS int `json:"poll_interval_ms"`
}

func DefaultClipboardConfig() ClipboardConfig {
    return ClipboardConfig{
        MaxSize:     100000,
        TruncateSize: 10000,
        PollIntervalMS: 500,
    }
}

type TUIConfig struct {
    Theme            string            `json:"theme"` // "dark", "light", "auto"
    StreamingDelayMS int               `json:"streaming_delay_ms"`
    Keybindings      map[string][]string `json:"keybindings"`
}

func DefaultTUIConfig() TUIConfig {
    return TUIConfig{
        Theme:       "auto",
        StreamingDelayMS: 50,
        Keybindings: map[string][]string{
            "navigate_down": {"j", "down"},
            "navigate_up":   {"k", "up"},
            "select":        {"enter"},
            "back":          {"esc", "q"},
            "copy":          {"c", "ctrl+c"},
            "quit":          {"ctrl+q"},
        },
    }
}

type LLMConfig struct {
    MaxConcurrent int   `json:"max_concurrent"`
    RetryAttempts int   `json:"retry_attempts"`
    RetryBackoff  []int `json:"retry_backoff"` // milliseconds
}

func DefaultLLMConfig() LLMConfig {
    return LLMConfig{
        MaxConcurrent: 1,
        RetryAttempts: 3,
        RetryBackoff:  []int{1000, 2000, 4000},
    }
}

type LoggingConfig struct {
    Level      string `json:"level"` // debug, info, warn, error
    File       string `json:"file"`
    MaxSizeMB  int    `json:"max_size_mb"`
    MaxBackups int    `json:"max_backups"`
    Console    bool   `json:"console"`
}

func DefaultLoggingConfig() LoggingConfig {
    home := os.Getenv("HOME")
    if home == "" && runtime.GOOS == "windows" {
        home = os.Getenv("USERPROFILE")
    }
    configDir := filepath.Join(home, ".config", "clipboard-tui")
    return LoggingConfig{
        Level:      "info",
        File:       filepath.Join(configDir, "clipboard-tui.log"),
        MaxSizeMB:  10,
        MaxBackups: 5,
        Console:    false,
    }
}

type DaemonConfig struct {
    PIDFile    string `json:"pid_file"`
    AutoStart  bool   `json:"auto_start"`
}

func DefaultDaemonConfig() DaemonConfig {
    home := os.Getenv("HOME")
    if home == "" && runtime.GOOS == "windows" {
        home = os.Getenv("USERPROFILE")
    }
    configDir := filepath.Join(home, ".config", "clipboard-tui")
    return DaemonConfig{
        PIDFile:   filepath.Join(configDir, "daemon.pid"),
        AutoStart: true,
    }
}

type TerminalConfig struct {
    Emulator    string `json:"emulator"`     // "auto" or profile id
    FallbackDir string `json:"fallback_dir"` // empty uses GetConfigDir()/output
}

func DefaultTerminalConfig() TerminalConfig {
    return TerminalConfig{
        Emulator:    "auto",
        FallbackDir: filepath.Join(GetConfigDir(), "output"),
    }
}
```

Environment: `CLIPBOARD_TUI_TERMINAL` overrides `terminal.emulator` when set (see [02-terminal-spawner.md](implementations/phase2/02-terminal-spawner.md)).

---

## API Key Storage (System Keyring)

```go
// internal/config/keyring.go
package config

import (
    "errors"
    "github.com/99designs/keyring"
)

const (
    serviceName = "clipboard-tui"
    apiKeyUser  = "openrouter_api_key"
)

// keyringService wraps the keyring library
var kr keyring.Keyring

func init() {
    kr = keyring.New(serviceName, "config")
}

// SaveAPIKey stores the API key in system keyring
func SaveAPIKey(key string) error {
    return kr.Set(apiKeyUser, key)
}

// GetAPIKey retrieves the API key from system keyring
func GetAPIKey() (string, error) {
    key, err := kr.Get(apiKeyUser)
    if err != nil {
        if errors.Is(err, keyring.ErrNotFound) {
            return "", errors.New("API key not found in system keyring")
        }
        return "", err
    }
    return key, nil
}

// DeleteAPIKey removes the API key from system keyring
func DeleteAPIKey() error {
    return kr.Remove(apiKeyUser)
}
```

---

## Environment Variable Overrides

All config values can be overridden via environment variables:

| Config Path | Environment Variable | Example |
|-------------|---------------------|---------|
| backend | `CLIPBOARD_TUI_BACKEND` | `ollama` |
| ollama.url | `CLIPBOARD_TUI_OLLAMA_URL` | `http://localhost:11434` |
| ollama.model | `CLIPBOARD_TUI_OLLAMA_MODEL` | `llama3:8b` |
| hotkey.key | `CLIPBOARD_TUI_HOTKEY_KEY` | `v` |
| log_level | `CLIPBOARD_TUI_LOG_LEVEL` | `debug` |
| config file path | `CLIPBOARD_TUI_CONFIG` | `/path/to/config.json` |

**Implementation**:
```go
func LoadWithEnv() (*Config, error) {
    cfg := Default()
    
    // Override from environment
    if backend := os.Getenv("CLIPBOARD_TUI_BACKEND"); backend != "" {
        cfg.Backend = backend
    }
    if url := os.Getenv("CLIPBOARD_TUI_OLLAMA_URL"); url != "" {
        cfg.Ollama.URL = url
    }
    if level := os.Getenv("CLIPBOARD_TUI_LOG_LEVEL"); level != "" {
        cfg.Logging.Level = level
    }
    
    // Load from file (if exists)
    path := os.Getenv("CLIPBOARD_TUI_CONFIG")
    if path == "" {
        path = GetConfigPath()
    }
    
    if _, err := os.Stat(path); err == nil {
        fileCfg, err := Load(path)
        if err != nil {
            return nil, err
        }
        // Merge: env > file > defaults
        mergeConfig(cfg, fileCfg)
    }
    
    // Validate
    if err := cfg.Validate(); err != nil {
        return nil, err
    }
    
    return cfg, nil
}
```

---

## Validation

```go
func (cfg *Config) Validate() error {
    // Validate backend
    if cfg.Backend != "ollama" && cfg.Backend != "openrouter" {
        return errors.New("invalid backend: must be 'ollama' or 'openrouter'")
    }
    
    // Validate hotkey
    if len(cfg.Hotkey.Modifiers) == 0 {
        return errors.New("hotkey must have at least one modifier")
    }
    if cfg.Hotkey.Key == "" {
        return errors.New("hotkey must have a key")
    }
    
    // Validate polling interval
    if cfg.Clipboard.PollIntervalMS < 100 || cfg.Clipboard.PollIntervalMS > 5000 {
        return errors.New("poll interval must be between 100ms and 5000ms")
    }
    
    // Validate max_size > truncate_size
    if cfg.Clipboard.TruncateSize >= cfg.Clipboard.MaxSize {
        return errors.New("truncate_size must be less than max_size")
    }
    
    // Validate log level
    validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
    if !validLevels[cfg.Logging.Level] {
        return errors.New("invalid log level: must be debug, info, warn, or error")
    }
    
    return nil
}
```

---

## Migration System

```go
// internal/config/migration.go
package config

import (
    "encoding/json"
)

// Current schema version
const currentVersion = "1"

// Migration function type
type Migration func(map[string]interface{}) map[string]interface{}

// Migrations map from version to migration function
var migrations = map[string]Migration{
    "0_to_1": migrate0to1,
}

// migrate applies all necessary migrations
func migrate(data map[string]interface{}) map[string]interface{} {
    version := data["version"].(string)
    
    for v, migration := range migrations {
        if shouldMigrate(version, v) {
            data = migration(data)
        }
    }
    
    data["version"] = currentVersion
    return data
}

func shouldMigrate(from, migration string) bool {
    // Parse version numbers and compare
    // e.g., "0" < "0_to_1" target
    return true // simplified
}

// Example migration: add new fields with defaults
func migrate0to1(data map[string]interface{}) map[string]interface{} {
    if _, ok := data["tui"]; !ok {
        data["tui"] = map[string]interface{}{
            "theme": "auto",
            "streaming_delay_ms": 50,
        }
    }
    if _, ok := data["llm"]; !ok {
        data["llm"] = map[string]interface{}{
            "max_concurrent": 1,
            "retry_attempts": 3,
        }
    }
    return data
}

// LoadWithMigration loads config and applies migrations
func LoadWithMigration(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var raw map[string]interface{}
    if err := json.Unmarshal(data, &raw); err != nil {
        return nil, err
    }
    
    // Apply migrations
    raw = migrate(raw)
    
    // Marshal back to JSON and unmarshal into Config struct
    migratedJSON, err := json.Marshal(raw)
    if err != nil {
        return nil, err
    }
    
    var cfg Config
    if err := json.Unmarshal(migratedJSON, &cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

---

## Config Directory Setup

```go
// internal/config/setup.go
package config

import (
    "os"
    "path/filepath"
)

// EnsureConfigDir creates config directory with proper permissions
func EnsureConfigDir() (string, error) {
    path := GetConfigDir()
    
    if err := os.MkdirAll(path, 0700); err != nil {
        return "", err
    }
    
    return path, nil
}

// GetConfigDir returns the config directory path
func GetConfigDir() string {
    if runtime.GOOS == "windows" {
        return filepath.Join(os.Getenv("APPDATA"), "clipboard-tui")
    }
    return filepath.Join(os.Getenv("HOME"), ".config", "clipboard-tui")
}

// CheckPermissions verifies config directory has secure permissions
func CheckPermissions() error {
    path := GetConfigDir()
    info, err := os.Stat(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // Doesn't exist yet, will be created with proper perms
        }
        return err
    }
    
    // Check directory permissions (should be 0700)
    if info.Mode().Perm()&077 != 0 {
        return errors.New("config directory has insecure permissions")
    }
    
    return nil
}
```
