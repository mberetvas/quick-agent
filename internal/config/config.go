// Package config defines the configuration structs and methods for loading, validation,
// and schema migration for the quick-agent application.
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
	Terminal   TerminalConfig   `json:"terminal"`
	Prompts    PromptsConfig    `json:"prompts"`
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

// TerminalConfig selects the terminal emulator used to spawn the TUI.
type TerminalConfig struct {
	Emulator    string `json:"emulator"`     // "auto" or profile id: wt, cmd, terminal, gnome-terminal, ...
	FallbackDir string `json:"fallback_dir"` // empty = default under GetConfigDir()/output
}

// PromptsConfig holds override strings for the built-in prompt templates.
// Any field left empty falls back to the built-in default.
type PromptsConfig struct {
	TranslateTargetLanguage string `json:"translate_target_language"` // default "English"
	Refine                  string `json:"refine"`
	Translate               string `json:"translate"`
	Summarize               string `json:"summarize"`
	Explain                 string `json:"explain"`
}

// DefaultPromptsConfig returns sensible defaults for all prompt templates.
func DefaultPromptsConfig() PromptsConfig {
	return PromptsConfig{
		TranslateTargetLanguage: "English",
		Refine:                  "Please refine, format, and correct the following text, improving grammar, spelling, and readability without adding empty introductory conversational filler. Output only the improved text:\n\n{{.Text}}",
		Translate:               "Translate to {{.Language}}:\n\n{{.Text}}",
		Summarize:               "Please generate a concise, bullet-pointed summary of the major key points from the following text:\n\n{{.Text}}",
		Explain:                 "Please explain the technical concepts or code snippet shown below simply and clearly, using markdown formatting:\n\n{{.Text}}",
	}
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

// DefaultLoggingConfig returns the default logging configuration.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:      "info",
		File:       filepath.Join(GetConfigDir(), "quick-agent.log"),
		MaxSizeMB:  10,
		MaxBackups: 5,
		Console:    false,
	}
}

// DefaultTerminalConfig returns default terminal emulator settings.
func DefaultTerminalConfig() TerminalConfig {
	return TerminalConfig{
		Emulator:    "auto",
		FallbackDir: filepath.Join(GetConfigDir(), "output"),
	}
}

// DefaultDaemonConfig returns the default daemon configuration.
func DefaultDaemonConfig() DaemonConfig {
	return DaemonConfig{
		PIDFile:   filepath.Join(GetConfigDir(), "daemon.pid"),
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
		Terminal:   DefaultTerminalConfig(),
		Prompts:    DefaultPromptsConfig(),
	}
}

func userHomeDir() string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return home
	}
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return os.Getenv("USERPROFILE")
}

// GetConfigDir returns the default configuration directory under the user home.
func GetConfigDir() string {
	return filepath.Join(userHomeDir(), ".quick-agent")
}

// GetConfigPath returns the default path to the config file.
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.json")
}
