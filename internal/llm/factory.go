package llm

import (
	"fmt"

	"github.com/mberetvas/quick-agent/internal/config"
	"github.com/mberetvas/quick-agent/internal/llm/ollama"
	"github.com/mberetvas/quick-agent/internal/llm/openrouter"
)

// NewClientFromConfig returns an LLM client for the configured backend.
func NewClientFromConfig(cfg *config.Config) (LLMClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	switch cfg.Backend {
	case "ollama":
		return ollama.NewClient(cfg.Ollama, cfg.LLM), nil
	case "openrouter":
		return openrouter.NewClient(cfg.OpenRouter, cfg.LLM), nil
	default:
		return nil, fmt.Errorf("unsupported backend: %s (use ollama or openrouter)", cfg.Backend)
	}
}
