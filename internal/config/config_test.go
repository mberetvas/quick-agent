package config

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	// Verify defaults are present
	if cfg.Version != "1" {
		t.Errorf("expected default version '1', got '%s'", cfg.Version)
	}
	if cfg.Backend != "ollama" {
		t.Errorf("expected default backend 'ollama', got '%s'", cfg.Backend)
	}
	if cfg.Ollama.URL != "http://localhost:11434" {
		t.Errorf("default ollama URL mismatch, got '%s'", cfg.Ollama.URL)
	}
	if cfg.Ollama.Model != "llama3:8b" {
		t.Errorf("default ollama model mismatch, got '%s'", cfg.Ollama.Model)
	}
	if cfg.Ollama.Timeout != 30 {
		t.Errorf("default ollama timeout is not 30, got %d", cfg.Ollama.Timeout)
	}
	if cfg.Ollama.MaxTokens != 2048 {
		t.Errorf("default ollama max_tokens is not 2048, got %d", cfg.Ollama.MaxTokens)
	}
	if cfg.Ollama.Temperature != 0.7 {
		t.Errorf("default ollama temperature is not 0.7, got %v", cfg.Ollama.Temperature)
	}

	if cfg.OpenRouter.Model != "mistralai/mistral-7b-instruct" {
		t.Errorf("default openrouter model mismatch, got '%s'", cfg.OpenRouter.Model)
	}

	if cfg.Clipboard.PollIntervalMS != 500 {
		t.Errorf("default poll interval mismatch, got %d", cfg.Clipboard.PollIntervalMS)
	}

	if cfg.TUI.Theme != "auto" {
		t.Errorf("default tui theme mismatch, got '%s'", cfg.TUI.Theme)
	}

	if cfg.Terminal.Emulator != "auto" {
		t.Errorf("default terminal emulator mismatch, got '%s'", cfg.Terminal.Emulator)
	}
	expectedFallback := filepath.Join(GetConfigDir(), "output")
	if cfg.Terminal.FallbackDir != expectedFallback {
		t.Errorf("default terminal fallback_dir = %q, want %q", cfg.Terminal.FallbackDir, expectedFallback)
	}

	// Default config must pass validation
	if err := cfg.Validate(); err != nil {
		t.Errorf("default config is invalid: %v", err)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid base",
			mutate:  func(c *Config) {},
			wantErr: false,
		},
		{
			name: "invalid backend",
			mutate: func(c *Config) {
				c.Backend = "invalid_llm"
			},
			wantErr: true,
		},
		{
			name: "negative ollama timeout",
			mutate: func(c *Config) {
				c.Ollama.Timeout = -1
			},
			wantErr: true,
		},
		{
			name: "negative openrouter timeout",
			mutate: func(c *Config) {
				c.OpenRouter.Timeout = -1
			},
			wantErr: true,
		},
		{
			name: "negative ollama max_tokens",
			mutate: func(c *Config) {
				c.Ollama.MaxTokens = -5
			},
			wantErr: true,
		},
		{
			name: "ollama temperature too high",
			mutate: func(c *Config) {
				c.Ollama.Temperature = 2.5
			},
			wantErr: true,
		},
		{
			name: "hotkey missing modifiers",
			mutate: func(c *Config) {
				c.Hotkey.Modifiers = []string{}
			},
			wantErr: true,
		},
		{
			name: "hotkey missing key",
			mutate: func(c *Config) {
				c.Hotkey.Key = ""
			},
			wantErr: true,
		},
		{
			name: "hotkey debounce_ms negative",
			mutate: func(c *Config) {
				c.Hotkey.DebounceMS = -10
			},
			wantErr: true,
		},
		{
			name: "polling interval too low",
			mutate: func(c *Config) {
				c.Clipboard.PollIntervalMS = 50
			},
			wantErr: true,
		},
		{
			name: "polling interval too high",
			mutate: func(c *Config) {
				c.Clipboard.PollIntervalMS = 10000
			},
			wantErr: true,
		},
		{
			name: "clipboard max_size negative",
			mutate: func(c *Config) {
				c.Clipboard.MaxSize = -100
			},
			wantErr: true,
		},
		{
			name: "truncate_size larger than max_size",
			mutate: func(c *Config) {
				c.Clipboard.MaxSize = 500
				c.Clipboard.TruncateSize = 600
			},
			wantErr: true,
		},
		{
			name: "truncate_size negative",
			mutate: func(c *Config) {
				c.Clipboard.TruncateSize = -100
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			mutate: func(c *Config) {
				c.Logging.Level = "verbose"
			},
			wantErr: true,
		},
		{
			name: "negative concurrent llm value",
			mutate: func(c *Config) {
				c.LLM.MaxConcurrent = 0
			},
			wantErr: true,
		},
		{
			name: "negative retry attempts",
			mutate: func(c *Config) {
				c.LLM.RetryAttempts = -1
			},
			wantErr: true,
		},
		{
			name: "negative logging size",
			mutate: func(c *Config) {
				c.Logging.MaxSizeMB = 0
			},
			wantErr: true,
		},
		{
			name: "negative logging backups",
			mutate: func(c *Config) {
				c.Logging.MaxBackups = -5
			},
			wantErr: true,
		},
		{
			name: "empty terminal emulator",
			mutate: func(c *Config) {
				c.Terminal.Emulator = ""
			},
			wantErr: true,
		},
		{
			name: "invalid terminal emulator",
			mutate: func(c *Config) {
				c.Terminal.Emulator = "nonexistent"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.mutate(cfg)
			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMigration(t *testing.T) {
	// A vintage v0 config file structure mapping (missing version, missing tui, missing llm)
	v0JSON := `{
		"backend": "openrouter",
		"openrouter": {
			"model": "mistralai/test",
			"timeout": 45,
			"max_tokens": 1024,
			"temperature": 0.5
		}
	}`

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(v0JSON), &raw); err != nil {
		t.Fatalf("failed to unmarshal v0 test schema: %v", err)
	}

	// Apply migration dispatcher
	raw = migrate(raw)

	// Validate version update
	version, ok := raw["version"].(string)
	if !ok || version != "1" {
		t.Errorf("expected migrated version to be '1', got '%v'", raw["version"])
	}

	// Marshal back to JSON and unmarshal into Config struct to verify final migrated state
	migratedJSON, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("failed to marshal migrated JSON: %v", err)
	}

	var cfg Config
	if err := json.Unmarshal(migratedJSON, &cfg); err != nil {
		t.Fatalf("failed to unmarshal into Config: %v", err)
	}

	if cfg.TUI.Theme != "auto" {
		t.Errorf("expected tui theme 'auto', got '%v'", cfg.TUI.Theme)
	}

	if cfg.LLM.MaxConcurrent != 1 {
		t.Errorf("expected max_concurrent 1, got '%v'", cfg.LLM.MaxConcurrent)
	}
}

func TestLoadMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist.json")

	cfg, err := Load(nonExistentPath)
	if err != nil {
		t.Fatalf("Load on non-existent config path should not return an error, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected a non-nil config filled with defaults")
	}

	// Ensure defaults are populated
	if cfg.Backend != "ollama" {
		t.Errorf("expected ollama backend by default, got '%s'", cfg.Backend)
	}
}

func TestLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	cfg := Default()
	cfg.Backend = "openrouter"
	cfg.OpenRouter.Model = "custom-openrouter-model"
	cfg.Ollama.Timeout = 15

	err := Save(cfg, configPath)
	if err != nil {
		t.Fatalf("failed to Save config: %v", err)
	}

	// Load it back
	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to Load config back: %v", err)
	}

	if loadedCfg.Backend != "openrouter" {
		t.Errorf("expected backend 'openrouter', got '%s'", loadedCfg.Backend)
	}
	if loadedCfg.OpenRouter.Model != "custom-openrouter-model" {
		t.Errorf("expected openrouter model 'custom-openrouter-model', got '%s'", loadedCfg.OpenRouter.Model)
	}
	if loadedCfg.Ollama.Timeout != 15 {
		t.Errorf("expected ollama timeout 15, got %d", loadedCfg.Ollama.Timeout)
	}
}

func TestLoadWithEnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Save a base config
	baseCfg := Default()
	baseCfg.Backend = "ollama"
	baseCfg.Ollama.URL = "http://localhost:11434"
	baseCfg.Hotkey.Key = "v"
	if err := Save(baseCfg, configPath); err != nil {
		t.Fatalf("failed to save base config: %v", err)
	}

	// Setup environment variables to override
	t.Setenv("CLIPBOARD_TUI_BACKEND", "openrouter")
	t.Setenv("CLIPBOARD_TUI_OLLAMA_URL", "http://overridden-ollama:11434")
	t.Setenv("CLIPBOARD_TUI_OLLAMA_MODEL", "overridden-model:latest")
	t.Setenv("CLIPBOARD_TUI_HOTKEY_KEY", "x")
	t.Setenv("CLIPBOARD_TUI_LOG_LEVEL", "debug")
	t.Setenv("CLIPBOARD_TUI_TERMINAL", "cmd")

	// Load using environmental support
	loaded, err := LoadWithEnv(configPath)
	if err != nil {
		t.Fatalf("failed to LoadWithEnv: %v", err)
	}

	if loaded.Backend != "openrouter" {
		t.Errorf("env backend override failed, got '%s'", loaded.Backend)
	}
	if loaded.Ollama.URL != "http://overridden-ollama:11434" {
		t.Errorf("env ollama URL override failed, got '%s'", loaded.Ollama.URL)
	}
	if loaded.Ollama.Model != "overridden-model:latest" {
		t.Errorf("env ollama model override failed, got '%s'", loaded.Ollama.Model)
	}
	if loaded.Hotkey.Key != "x" {
		t.Errorf("env hotkey key override failed, got '%s'", loaded.Hotkey.Key)
	}
	if loaded.Logging.Level != "debug" {
		t.Errorf("env log level override failed, got '%s'", loaded.Logging.Level)
	}
	if loaded.Terminal.Emulator != "cmd" {
		t.Errorf("env terminal override failed, got '%s'", loaded.Terminal.Emulator)
	}
}
