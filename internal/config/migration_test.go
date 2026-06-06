package config

import (
	"encoding/json"
	"testing"
)

func TestMigrate0to1_fillsTUIAndLLMDefaults(t *testing.T) {
	data := map[string]interface{}{
		"backend": "ollama",
	}

	out := migrate0to1(data)

	tui, ok := out["tui"].(map[string]interface{})
	if !ok {
		t.Fatalf("tui = %T, want map", out["tui"])
	}
	if tui["theme"] != "auto" {
		t.Errorf("tui.theme = %v, want auto", tui["theme"])
	}
	if asInt(tui["streaming_delay_ms"]) != 50 {
		t.Errorf("tui.streaming_delay_ms = %v, want 50", tui["streaming_delay_ms"])
	}
	kb, ok := tui["keybindings"].(map[string]interface{})
	if !ok || len(kb) == 0 {
		t.Fatalf("expected keybindings map, got %v", tui["keybindings"])
	}

	llm, ok := out["llm"].(map[string]interface{})
	if !ok {
		t.Fatalf("llm = %T, want map", out["llm"])
	}
	if asInt(llm["max_concurrent"]) != 1 {
		t.Errorf("llm.max_concurrent = %v, want 1", llm["max_concurrent"])
	}
	if asInt(llm["retry_attempts"]) != 3 {
		t.Errorf("llm.retry_attempts = %v, want 3", llm["retry_attempts"])
	}
}

func TestMigrate0to1_preservesExistingSections(t *testing.T) {
	data := map[string]interface{}{
		"tui": map[string]interface{}{"theme": "dark"},
		"llm": map[string]interface{}{"max_concurrent": float64(5)},
	}

	out := migrate0to1(data)

	tui := out["tui"].(map[string]interface{})
	if tui["theme"] != "dark" {
		t.Errorf("existing tui.theme overwritten: %v", tui["theme"])
	}
	llm := out["llm"].(map[string]interface{})
	if llm["max_concurrent"] != float64(5) {
		t.Errorf("existing llm.max_concurrent overwritten: %v", llm["max_concurrent"])
	}
}

func TestMigrate_versionDispatch(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]interface{}
		wantTUI     bool
		wantLLM     bool
		wantVersion string
	}{
		{
			name:        "missing version migrates from v0",
			input:       map[string]interface{}{"backend": "ollama"},
			wantTUI:     true,
			wantLLM:     true,
			wantVersion: CurrentVersion,
		},
		{
			name:        "version zero migrates",
			input:       map[string]interface{}{"version": "0", "backend": "ollama"},
			wantTUI:     true,
			wantLLM:     true,
			wantVersion: CurrentVersion,
		},
		{
			name: "version one skips migration fill",
			input: map[string]interface{}{
				"version": "1",
				"backend": "ollama",
				"tui":     map[string]interface{}{"theme": "light"},
				"llm":     map[string]interface{}{"max_concurrent": float64(2)},
			},
			wantTUI:     false,
			wantLLM:     false,
			wantVersion: CurrentVersion,
		},
		{
			name:        "non-string version treated as zero",
			input:       map[string]interface{}{"version": 1, "backend": "ollama"},
			wantTUI:     true,
			wantLLM:     true,
			wantVersion: CurrentVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := copyMap(tt.input)
			hadTUI := raw["tui"] != nil
			hadLLM := raw["llm"] != nil

			out := migrate(raw)

			version, _ := out["version"].(string)
			if version != tt.wantVersion {
				t.Errorf("version = %q, want %q", version, tt.wantVersion)
			}

			if tt.wantTUI && !hadTUI {
				if _, ok := out["tui"]; !ok {
					t.Error("expected tui section after migration")
				}
			}
			if tt.wantLLM && !hadLLM {
				if _, ok := out["llm"]; !ok {
					t.Error("expected llm section after migration")
				}
			}
			if tt.name == "version one skips migration fill" {
				tui := out["tui"].(map[string]interface{})
				if tui["theme"] != "light" {
					t.Errorf("tui.theme = %v, want light", tui["theme"])
				}
			}
		})
	}
}

func TestMigrate_stampsCurrentVersion(t *testing.T) {
	raw := map[string]interface{}{"backend": "openrouter"}
	out := migrate(raw)

	if out["version"] != CurrentVersion {
		t.Errorf("version = %v, want %q", out["version"], CurrentVersion)
	}

	migratedJSON, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var cfg Config
	if err := json.Unmarshal(migratedJSON, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Version != CurrentVersion {
		t.Errorf("cfg.Version = %q, want %q", cfg.Version, CurrentVersion)
	}
}

func copyMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func asInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return -1
	}
}
