package llm

import (
	"strings"
	"testing"

	"github.com/mberetvas/quick-agent/internal/config"
)

func TestNewClientFromConfig(t *testing.T) {
	tests := []struct {
		name      string
		backend   string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "ollama backend",
			backend: "ollama",
		},
		{
			name:    "openrouter backend",
			backend: "openrouter",
		},
		{
			name:      "invalid backend",
			backend:   "anthropic",
			wantErr:   true,
			errSubstr: "unsupported backend",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default()
			cfg.Backend = tt.backend

			client, err := NewClientFromConfig(cfg)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewClientFromConfig: %v", err)
			}
			if client == nil {
				t.Fatal("expected non-nil client")
			}
		})
	}
}
