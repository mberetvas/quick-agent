package models

import (
	"context"
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mberetvas/quick-agent/internal/config"
	"github.com/mberetvas/quick-agent/internal/llm"
	"github.com/mberetvas/quick-agent/internal/tui/styles"
	"strings"
	"testing"
)

type errStreamLLM struct {
	streamErr error
}

func (e *errStreamLLM) Generate(ctx context.Context, prompt string) (<-chan string, <-chan error, error) {
	tokens := make(chan string)
	close(tokens)
	errs := make(chan error, 1)
	errs <- e.streamErr
	close(errs)
	return tokens, errs, nil
}

func (e *errStreamLLM) HealthCheck(ctx context.Context) error { return nil }

func TestResultModel_View_states(t *testing.T) {
	theme := styles.DefaultTheme()
	keys := resultTestKeyMap()

	t.Run("streaming empty", func(t *testing.T) {
		m := NewResultModel(ActionRefine, "x", &mockLLM{}, llm.NewPromptRegistry(config.DefaultPromptsConfig()), theme, keys, &mockClipboard{}, 0, "")
		m.streaming = true
		view := m.View()
		if !strings.Contains(view, "Waiting for response") {
			t.Errorf("view = %q", view)
		}
	})

	t.Run("idle empty", func(t *testing.T) {
		m := NewResultModel(ActionSummarize, "x", &mockLLM{}, llm.NewPromptRegistry(config.DefaultPromptsConfig()), theme, keys, &mockClipboard{}, 0, "")
		view := m.View()
		if !strings.Contains(view, "Summarize") || !strings.Contains(view, "no output") {
			t.Errorf("view = %q", view)
		}
	})

	t.Run("with result and copied", func(t *testing.T) {
		m := NewResultModel(ActionExplain, "x", &mockLLM{}, llm.NewPromptRegistry(config.DefaultPromptsConfig()), theme, keys, &mockClipboard{}, 0, "")
		m.result = "output text"
		m.copied = true
		view := m.View()
		if !strings.Contains(view, "output text") || !strings.Contains(view, "Copied") {
			t.Errorf("view = %q", view)
		}
	})

	t.Run("translate title", func(t *testing.T) {
		m := NewResultModel(ActionTranslate, "x", &mockLLM{}, llm.NewPromptRegistry(config.DefaultPromptsConfig()), theme, keys, &mockClipboard{}, 0, "Dutch")
		view := m.View()
		if !strings.Contains(view, "Translate") {
			t.Errorf("view = %q", view)
		}
	})
}

func TestResultModel_streamError_emitsShowError(t *testing.T) {
	theme := styles.DefaultTheme()
	keys := resultTestKeyMap()
	model := *NewResultModel(ActionRefine, "x", &mockLLM{}, llm.NewPromptRegistry(config.DefaultPromptsConfig()), theme, keys, &mockClipboard{}, 0, "")

	_, cmd := model.Update(streamErrMsg{err: errors.New("connection refused")})
	if cmd == nil {
		t.Fatal("expected ShowErrorEvent cmd")
	}
	if _, ok := cmd().(ShowErrorEvent); !ok {
		t.Fatalf("got %T", cmd())
	}
}

func TestResultModel_retry_restartsGeneration(t *testing.T) {
	theme := styles.DefaultTheme()
	keys := resultTestKeyMap()
	model := *NewResultModel(ActionRefine, "x", &mockLLM{tokens: []string{"x"}}, llm.NewPromptRegistry(config.DefaultPromptsConfig()), theme, keys, &mockClipboard{}, 0, "")
	model.result = "old"
	cmd := model.Retry()
	if cmd == nil {
		t.Fatal("expected retry cmd")
	}
}

func TestResultModel_streamingDelay_flush(t *testing.T) {
	theme := styles.DefaultTheme()
	keys := resultTestKeyMap()
	model := *NewResultModel(ActionRefine, "x", &mockLLM{tokens: []string{"a"}}, llm.NewPromptRegistry(config.DefaultPromptsConfig()), theme, keys, &mockClipboard{}, 10, "")

	cmd := model.Init()
	for i := 0; i < 30 && cmd != nil; i++ {
		msg := cmd()
		if msg == nil {
			break
		}
		var next tea.Cmd
		model, next = model.Update(msg)
		cmd = next
	}
	if model.result == "" && model.pending == "" {
		t.Log("streaming delay path exercised")
	}
}

func TestClassifyStreamError_viaUpdate(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		contain string
	}{
		{"timeout", errors.New("context deadline exceeded"), "Timed Out"},
		{"connection", errors.New("dial tcp: connection refused"), "Connection Failed"},
		{"auth", errors.New("401 unauthorized"), "Invalid API Key"},
		{"generic", errors.New("something else"), "Generation Failed"},
	}
	theme := styles.DefaultTheme()
	keys := resultTestKeyMap()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := *NewResultModel(ActionRefine, "x", &errStreamLLM{streamErr: tt.err}, llm.NewPromptRegistry(config.DefaultPromptsConfig()), theme, keys, &mockClipboard{}, 0, "")
			_, cmd := model.Update(streamErrMsg{err: tt.err})
			if cmd == nil {
				t.Fatal("expected error navigation cmd")
			}
			ev, ok := cmd().(ShowErrorEvent)
			if !ok {
				t.Fatalf("got %T", cmd())
			}
			if !strings.Contains(ev.Err.Title, tt.contain) && !strings.Contains(ev.Err.Message, tt.contain) {
				t.Errorf("error = %+v, want substring %q", ev.Err, tt.contain)
			}
		})
	}
}
