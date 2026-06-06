package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoad_missingFileReturnsDefault(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Backend != "ollama" {
		t.Errorf("Backend = %q, want ollama", cfg.Backend)
	}
}

func TestLoad_validFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	original := Default()
	original.Backend = "openrouter"
	original.OpenRouter.Model = "test/model"
	if err := Save(original, path); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Backend != "openrouter" {
		t.Errorf("Backend = %q, want openrouter", cfg.Backend)
	}
	if cfg.OpenRouter.Model != "test/model" {
		t.Errorf("OpenRouter.Model = %q", cfg.OpenRouter.Model)
	}
}

func TestLoad_appliesMigration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	v0 := `{"backend":"ollama","hotkey":{"modifiers":["ctrl"],"key":"v","debounce_ms":300},"clipboard":{"max_size":1000,"truncate_size":100,"poll_interval_ms":500},"logging":{"level":"info","max_size_mb":10,"max_backups":1}}`
	if err := os.WriteFile(path, []byte(v0), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Version != CurrentVersion {
		t.Errorf("Version = %q, want %q", cfg.Version, CurrentVersion)
	}
	if cfg.TUI.Theme != "auto" {
		t.Errorf("TUI.Theme = %q, want auto", cfg.TUI.Theme)
	}
	if cfg.LLM.MaxConcurrent != 1 {
		t.Errorf("LLM.MaxConcurrent = %d, want 1", cfg.LLM.MaxConcurrent)
	}
}

func TestLoad_invalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("{not json"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoad_validationError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	invalid := Default()
	invalid.Backend = "invalid"
	data, _ := json.Marshal(invalid)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadWithEnv_allOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := Save(Default(), path); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLIPBOARD_TUI_BACKEND", "openrouter")
	t.Setenv("CLIPBOARD_TUI_OLLAMA_URL", "http://env-ollama:11434")
	t.Setenv("CLIPBOARD_TUI_OLLAMA_MODEL", "env-model")
	t.Setenv("CLIPBOARD_TUI_HOTKEY_KEY", "x")
	t.Setenv("CLIPBOARD_TUI_LOG_LEVEL", "debug")
	t.Setenv("CLIPBOARD_TUI_TERMINAL", validTerminalForTest())

	cfg, err := LoadWithEnv(path)
	if err != nil {
		t.Fatalf("LoadWithEnv() error = %v", err)
	}
	if cfg.Backend != "openrouter" {
		t.Errorf("Backend = %q", cfg.Backend)
	}
	if cfg.Ollama.URL != "http://env-ollama:11434" {
		t.Errorf("Ollama.URL = %q", cfg.Ollama.URL)
	}
	if cfg.Ollama.Model != "env-model" {
		t.Errorf("Ollama.Model = %q", cfg.Ollama.Model)
	}
	if cfg.Hotkey.Key != "x" {
		t.Errorf("Hotkey.Key = %q", cfg.Hotkey.Key)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q", cfg.Logging.Level)
	}
	if cfg.Terminal.Emulator != validTerminalForTest() {
		t.Errorf("Terminal.Emulator = %q", cfg.Terminal.Emulator)
	}
}

func TestLoadWithEnv_missingEnvConfigUsesDefaults(t *testing.T) {
	t.Setenv("CLIPBOARD_TUI_CONFIG", filepath.Join(t.TempDir(), "does-not-exist.json"))

	cfg, err := LoadWithEnv("")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Backend != "ollama" {
		t.Errorf("Backend = %q, want ollama", cfg.Backend)
	}
}

func TestLoadWithEnv_configPathFromEnv(t *testing.T) {
	dir := t.TempDir()
	customPath := filepath.Join(dir, "custom.json")

	cfg := Default()
	cfg.Backend = "openrouter"
	if err := Save(cfg, customPath); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLIPBOARD_TUI_CONFIG", customPath)

	loaded, err := LoadWithEnv("")
	if err != nil {
		t.Fatalf("LoadWithEnv() error = %v", err)
	}
	if loaded.Backend != "openrouter" {
		t.Errorf("Backend = %q, want openrouter", loaded.Backend)
	}
}

func TestLoadWithEnv_envOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	fileCfg := Default()
	fileCfg.Backend = "ollama"
	fileCfg.Hotkey.Key = "v"
	if err := Save(fileCfg, path); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CLIPBOARD_TUI_BACKEND", "openrouter")
	t.Setenv("CLIPBOARD_TUI_HOTKEY_KEY", "z")

	loaded, err := LoadWithEnv(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Backend != "openrouter" {
		t.Errorf("env should override file backend, got %q", loaded.Backend)
	}
	if loaded.Hotkey.Key != "z" {
		t.Errorf("env should override file hotkey, got %q", loaded.Hotkey.Key)
	}
}

func TestSave_fileAndDirPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix permission bits not enforced on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.json")

	if err := Save(Default(), path); err != nil {
		t.Fatal(err)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode().Perm() != 0700 {
		t.Errorf("dir perm = %o, want 0700", dirInfo.Mode().Perm())
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Mode().Perm() != 0600 {
		t.Errorf("file perm = %o, want 0600", fileInfo.Mode().Perm())
	}
}

func TestSave_roundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	original := Default()
	original.Backend = "openrouter"
	original.Ollama.Timeout = 99
	if err := Save(original, path); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Backend != original.Backend {
		t.Errorf("Backend = %q, want %q", loaded.Backend, original.Backend)
	}
	if loaded.Ollama.Timeout != original.Ollama.Timeout {
		t.Errorf("Ollama.Timeout = %d, want %d", loaded.Ollama.Timeout, original.Ollama.Timeout)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", t.TempDir())
	} else {
		t.Setenv("HOME", t.TempDir())
	}

	path, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}
	if path != GetConfigDir() {
		t.Errorf("path = %q, want %q", path, GetConfigDir())
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Fatal("expected directory")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0700 {
		t.Errorf("perm = %o, want 0700", info.Mode().Perm())
	}
}

func TestGetConfigPath(t *testing.T) {
	want := filepath.Join(GetConfigDir(), "config.json")
	if got := GetConfigPath(); got != want {
		t.Errorf("GetConfigPath() = %q, want %q", got, want)
	}
}

func validTerminalForTest() string {
	switch runtime.GOOS {
	case "windows":
		return "cmd"
	case "darwin":
		return "terminal"
	default:
		return "gnome-terminal"
	}
}
