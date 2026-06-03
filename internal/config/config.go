// Package config defines the configuration structs and methods for loading, validation,
// and schema migration for the clipboard-tui application.
package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// Config represents the root configuration structure
type Config struct {
	Version    string           `json:"version"`
	Backend    string           `json:"backend"` // "ollama" or "openrouter"
	Ollama     OllamaConfig     `json:"ollama"`
	OpenRouter OpenRouterConfig `json:"openrouter"`
	Hotkey     HotkeyConfig     `json:"hotkey"`
	Clipboard  ClipboardConfig  `json:"clipboard"`
	TUI        TUIConfig        `json:"tui"`
	LLM        LLMConfig        `json:"llm"`
	Logging    LoggingConfig    `json:"logging"`
	Daemon     DaemonConfig     `json:"daemon"`
}

// OllamaConfig holds configuration details for the Ollama backend
type OllamaConfig struct {
	URL         string  `json:"url"`
	Model       string  `json:"model"`
	Timeout     int     `json:"timeout"` // seconds
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// OpenRouterConfig holds configuration details for the OpenRouter backend
type OpenRouterConfig struct {
	Model       string  `json:"model"`
	Timeout     int     `json:"timeout"` // seconds
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// HotkeyConfig defines the global hotkey configuration
type HotkeyConfig struct {
	Modifiers  []string `json:"modifiers"` // e.g., ["ctrl", "alt"]
	Key        string   `json:"key"`
	DebounceMS int      `json:"debounce_ms"`
}

// ClipboardConfig dictates clipboard-specific limits and behaviors
type ClipboardConfig struct {
	MaxSize        int `json:"max_size"`         // chars
	TruncateSize   int `json:"truncate_size"`    // chars
	PollIntervalMS int `json:"poll_interval_ms"` // ms
}

// TUIConfig provides styles and behaviors for the TUI frontend
type TUIConfig struct {
	Theme            string              `json:"theme"` // "dark", "light", "auto"
	StreamingDelayMS int                 `json:"streaming_delay_ms"`
	Keybindings      map[string][]string `json:"keybindings"`
}

// LLMConfig controls execution and retry behaviors for LLM calls
type LLMConfig struct {
	MaxConcurrent int   `json:"max_concurrent"`
	RetryAttempts int   `json:"retry_attempts"`
	RetryBackoff  []int `json:"retry_backoff"` // milliseconds
}

// LoggingConfig directs structured and console logs
type LoggingConfig struct {
	Level      string `json:"level"` // "debug", "info", "warn", "error"
	File       string `json:"file"`
	MaxSizeMB  int    `json:"max_size_mb"`
	MaxBackups int    `json:"max_backups"`
	Console    bool   `json:"console"`
}

// DaemonConfig directs background daemon execution
type DaemonConfig struct {
	PIDFile   string `json:"pid_file"`
	AutoStart bool   `json:"auto_start"`
}

// DefaultOllamaConfig returns a default Ollama configuration
func DefaultOllamaConfig() OllamaConfig {
	return OllamaConfig{
		URL:         "http://localhost:11434",
		Model:       "llama3:8b",
		Timeout:     30,
		MaxTokens:   2048,
		Temperature: 0.7,
	}
}

// DefaultOpenRouterConfig returns a default OpenRouter configuration
func DefaultOpenRouterConfig() OpenRouterConfig {
	return OpenRouterConfig{
		Model:       "mistralai/mistral-7b-instruct",
		Timeout:     30,
		MaxTokens:   2048,
		Temperature: 0.7,
	}
}

// DefaultHotkeyConfig returns a platform-specific default hotkey configuration
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

// DefaultClipboardConfig returns a default clipboard configuration
func DefaultClipboardConfig() ClipboardConfig {
	return ClipboardConfig{
		MaxSize:        100000,
		TruncateSize:   10000,
		PollIntervalMS: 500,
	}
}

// DefaultTUIConfig returns a default configuration for keybindings and rendering
func DefaultTUIConfig() TUIConfig {
	return TUIConfig{
		Theme:            "auto",
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

// DefaultLLMConfig returns a default retry/concurrency policy config
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		MaxConcurrent: 1,
		RetryAttempts: 3,
		RetryBackoff:  []int{1000, 2000, 4000},
	}
}

// DefaultLoggingConfig returns a platform-specific default logging configuration
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

// DefaultDaemonConfig returns a platform-specific default daemon configuration
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

// Default returns a config populated entirely from defaults
func Default() *Config {
	return &Config{
		Version:    "1",
		Backend:    "ollama",
		Ollama:     DefaultOllamaConfig(),
		OpenRouter: DefaultOpenRouterConfig(),
		Hotkey:     DefaultHotkeyConfig(),
		Clipboard:  DefaultClipboardConfig(),
		TUI:        DefaultTUIConfig(),
		LLM:        DefaultLLMConfig(),
		Logging:    DefaultLoggingConfig(),
		Daemon:     DefaultDaemonConfig(),
	}
}

// GetConfigDir returns the default platform-specific configuration directory
func GetConfigDir() string {
	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		if appdata != "" {
			return filepath.Join(appdata, "clipboard-tui")
		}
		// Fallback
		return filepath.Join(os.Getenv("USERPROFILE"), ".config", "clipboard-tui")
	}
	return filepath.Join(os.Getenv("HOME"), ".config", "clipboard-tui")
}

// GetConfigPath returns the default platform-specific path to the config file
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.json")
}
