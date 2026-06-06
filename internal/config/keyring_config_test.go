package config

import (
	"testing"

	"github.com/99designs/keyring"
)

func TestConfigGetAPIKey(t *testing.T) {
	mockKeyring := keyring.NewArrayKeyring(nil)
	oldKr := kr
	kr = mockKeyring
	defer func() { kr = oldKr }()

	const secret = "sk-test-key"
	if err := SaveAPIKey("openrouter", secret); err != nil {
		t.Fatal(err)
	}

	cfg := Default()
	cfg.Backend = "openrouter"

	got, err := cfg.GetAPIKey()
	if err != nil {
		t.Fatalf("Config.GetAPIKey() error = %v", err)
	}
	if got != secret {
		t.Errorf("Config.GetAPIKey() = %q, want %q", got, secret)
	}

	orKey, err := cfg.OpenRouter.GetAPIKey()
	if err != nil {
		t.Fatalf("OpenRouterConfig.GetAPIKey() error = %v", err)
	}
	if orKey != secret {
		t.Errorf("OpenRouterConfig.GetAPIKey() = %q, want %q", orKey, secret)
	}
}
