package config

import (
	"errors"
	"fmt"
	"os"
	"runtime"
)

// Validate checks whether the config is semantically correct.
// It returns a descriptive error if any values are malformed or out of bounds.
func (cfg *Config) Validate() error {
	// Validate backend
	if cfg.Backend != "ollama" && cfg.Backend != "openrouter" {
		return fmt.Errorf("invalid backend: must be 'ollama' or 'openrouter', got '%s'", cfg.Backend)
	}

	// Validate backend specific fields
	if cfg.Ollama.Timeout < 0 {
		return errors.New("ollama timeout cannot be negative")
	}
	if cfg.Ollama.MaxTokens < 0 {
		return errors.New("ollama max_tokens cannot be negative")
	}
	if cfg.Ollama.Temperature < 0.0 || cfg.Ollama.Temperature > 2.0 {
		return errors.New("ollama temperature must be between 0.0 and 2.0")
	}

	if cfg.OpenRouter.Timeout < 0 {
		return errors.New("openrouter timeout cannot be negative")
	}
	if cfg.OpenRouter.MaxTokens < 0 {
		return errors.New("openrouter max_tokens cannot be negative")
	}
	if cfg.OpenRouter.Temperature < 0.0 || cfg.OpenRouter.Temperature > 2.0 {
		return errors.New("openrouter temperature must be between 0.0 and 2.0")
	}

	// Validate hotkey
	if len(cfg.Hotkey.Modifiers) == 0 {
		return errors.New("hotkey must have at least one modifier")
	}
	for _, mod := range cfg.Hotkey.Modifiers {
		if mod != "ctrl" && mod != "alt" && mod != "shift" && mod != "cmd" && mod != "option" {
			return fmt.Errorf("invalid modifier in hotkey: %s", mod)
		}
	}
	if cfg.Hotkey.Key == "" {
		return errors.New("hotkey must have a key")
	}
	if cfg.Hotkey.DebounceMS < 0 {
		return errors.New("hotkey debounce_ms cannot be negative")
	}

	// Validate polling interval
	if cfg.Clipboard.PollIntervalMS < 100 || cfg.Clipboard.PollIntervalMS > 5000 {
		return fmt.Errorf("poll interval must be between 100ms and 5000ms, got %dms", cfg.Clipboard.PollIntervalMS)
	}

	// Validate max_size > truncate_size
	if cfg.Clipboard.MaxSize <= 0 {
		return errors.New("clipboard max_size must be greater than 0")
	}
	if cfg.Clipboard.TruncateSize < 0 {
		return errors.New("clipboard truncate_size cannot be negative")
	}
	if cfg.Clipboard.TruncateSize >= cfg.Clipboard.MaxSize {
		return fmt.Errorf("truncate_size (%d) must be less than max_size (%d)", cfg.Clipboard.TruncateSize, cfg.Clipboard.MaxSize)
	}

	// Validate LLM Config
	if cfg.LLM.MaxConcurrent < 1 {
		return errors.New("max_concurrent must be at least 1")
	}
	if cfg.LLM.RetryAttempts < 0 {
		return errors.New("retry_attempts cannot be negative")
	}

	// Validate log level
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		return fmt.Errorf("invalid log level: must be debug, info, warn, or error, got '%s'", cfg.Logging.Level)
	}

	if cfg.Logging.MaxSizeMB <= 0 {
		return errors.New("logging max_size_mb must be greater than 0")
	}
	if cfg.Logging.MaxBackups < 0 {
		return errors.New("logging max_backups cannot be negative")
	}

	if err := cfg.Terminal.validate(); err != nil {
		return err
	}

	return nil
}

func (t TerminalConfig) validate() error {
	if t.Emulator == "" {
		return errors.New("terminal.emulator must not be empty")
	}
	if t.Emulator == "auto" {
		return nil
	}
	for _, id := range validTerminalEmulators() {
		if t.Emulator == id {
			return nil
		}
	}
	return fmt.Errorf("invalid terminal.emulator %q for %s", t.Emulator, runtime.GOOS)
}

func validTerminalEmulators() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"wt", "powershell", "cmd"}
	case "darwin":
		return []string{"terminal", "iterm"}
	default:
		return []string{
			"x-terminal-emulator",
			"gnome-terminal",
			"konsole",
			"xfce4-terminal",
			"alacritty",
			"kitty",
		}
	}
}

// CheckPermissions verifies config directory has secure permissions (0700).
// Permissions are only strictly checked on non-Windows platforms.
func CheckPermissions() error {
	path := GetConfigDir()
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Doesn't exist yet, will be created with proper perms
		}
		return err
	}

	// Check directory permissions (should be 0700 equivalent, so no group/other access)
	if runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 {
		return errors.New("config directory has insecure permissions")
	}

	return nil
}
