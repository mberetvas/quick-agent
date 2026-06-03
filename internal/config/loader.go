package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// EnsureConfigDir creates config directory with proper permissions (0700).
func EnsureConfigDir() (string, error) {
	path := GetConfigDir()

	if err := os.MkdirAll(path, 0700); err != nil {
		return "", err
	}

	return path, nil
}

// Load loads configuration from the specified path. It applies any migrations
// and defaults for missing fields. If the file does not exist, it returns a default
// configuration (no error).
func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Default(), nil
	} else if err != nil {
		return nil, err
	}

	cfg := Default()
	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(fileData, &raw); err != nil {
		return nil, err
	}

	raw = migrate(raw)
	migratedJSON, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(migratedJSON, cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadWithEnv loads configuration from a platform-specific location or customized path,
// applying default values, loading from disk (if exists), upgrading schemas if older,
// and merging environment variable overrides.
// Environment variables always override file values, which override standard defaults.
func LoadWithEnv(customPath string) (*Config, error) {
	path := customPath
	if path == "" {
		path = os.Getenv("CLIPBOARD_TUI_CONFIG")
		if path == "" {
			path = GetConfigPath()
		}
	}

	cfg := Default()

	// If file exists, load & migrate, then unmarshal over default
	if _, err := os.Stat(path); err == nil {
		fileData, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var raw map[string]interface{}
		if err := json.Unmarshal(fileData, &raw); err != nil {
			return nil, err
		}

		raw = migrate(raw)
		migratedJSON, err := json.Marshal(raw)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(migratedJSON, cfg); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	// Apply Environment overrides: CLIPBOARD_TUI_...
	if backend := os.Getenv("CLIPBOARD_TUI_BACKEND"); backend != "" {
		cfg.Backend = backend
	}
	if url := os.Getenv("CLIPBOARD_TUI_OLLAMA_URL"); url != "" {
		cfg.Ollama.URL = url
	}
	if model := os.Getenv("CLIPBOARD_TUI_OLLAMA_MODEL"); model != "" {
		cfg.Ollama.Model = model
	}
	if key := os.Getenv("CLIPBOARD_TUI_HOTKEY_KEY"); key != "" {
		cfg.Hotkey.Key = key
	}
	if level := os.Getenv("CLIPBOARD_TUI_LOG_LEVEL"); level != "" {
		cfg.Logging.Level = level
	}

	// Validate the fully resolved configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes the given configuration state back to a file. It creates any
// parent directories with appropriate permissions.
func Save(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
