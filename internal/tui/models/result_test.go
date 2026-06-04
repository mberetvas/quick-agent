package models

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourname/clipboard-tui/internal/config"
	"github.com/yourname/clipboard-tui/internal/llm"
	"github.com/yourname/clipboard-tui/internal/tui/styles"
)

type mockLLM struct {
	tokens []string
	err    error
}

func (m *mockLLM) Generate(ctx context.Context, prompt string) (<-chan string, <-chan error, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	ch := make(chan string, len(m.tokens))
	for _, t := range m.tokens {
		ch <- t
	}
	close(ch)
	errCh := make(chan error)
	close(errCh)
	return ch, errCh, nil
}

func (m *mockLLM) HealthCheck(ctx context.Context) error {
	return nil
}

type mockClipboard struct {
	setText string
}

func (m *mockClipboard) Get() (string, error)  { return "", nil }
func (m *mockClipboard) Set(text string) error { m.setText = text; return nil }

func TestRenderPromptForAction(t *testing.T) {
	reg := llm.NewPromptRegistry()
	prompt := reg.Get(string(ActionRefine)).Render("hello")
	if prompt == "" || !strings.Contains(prompt, "hello") {
		t.Fatalf("expected prompt containing hello, got %q", prompt)
	}
}

func TestResultModel_accumulatesTokens(t *testing.T) {
	theme := styles.DefaultTheme()
	keys := resultTestKeyMap()
	cb := &mockClipboard{}
	model := *NewResultModel(ActionRefine, "input text", &mockLLM{tokens: []string{"A", "B"}}, llm.NewPromptRegistry(), theme, keys, cb, 0)

	cmd := model.Init()
	if cmd == nil {
		t.Fatal("expected Init command")
	}

	// Drain async commands until stream completes.
	for i := 0; i < 20; i++ {
		msg := cmd()
		if msg == nil {
			break
		}
		var next tea.Cmd
		model, next = model.Update(msg)
		cmd = next
		if !model.streaming && model.result != "" {
			break
		}
	}

	if model.result != "AB" {
		t.Errorf("result = %q, want AB", model.result)
	}
}

func TestResultModel_copy(t *testing.T) {
	theme := styles.DefaultTheme()
	keys := resultTestKeyMap()
	cb := &mockClipboard{}
	model := *NewResultModel(ActionRefine, "x", &mockLLM{tokens: []string{"done"}}, llm.NewPromptRegistry(), theme, keys, cb, 0)

	cmd := model.Init()
	for i := 0; i < 20; i++ {
		if cmd == nil {
			break
		}
		msg := cmd()
		if msg == nil {
			break
		}
		var next tea.Cmd
		model, next = model.Update(msg)
		cmd = next
	}

	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}, Alt: false})
	if cb.setText != "done" {
		t.Errorf("clipboard set = %q, want done", cb.setText)
	}
}

func TestResultModel_back_emitsBackEvent(t *testing.T) {
	theme := styles.DefaultTheme()
	keys := resultTestKeyMap()
	model := *NewResultModel(ActionRefine, "x", &mockLLM{}, llm.NewPromptRegistry(), theme, keys, &mockClipboard{}, 0)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected back command")
	}
	if _, ok := cmd().(BackEvent); !ok {
		t.Fatalf("expected BackEvent, got %T", cmd())
	}
}

func resultTestKeyMap() KeyMap {
	return NewKeyMap(config.TUIConfig{
		Keybindings: map[string][]string{
			"navigate_down": {"j", "down"},
			"navigate_up":   {"k", "up"},
			"select":        {"enter"},
			"back":          {"esc", "q"},
			"copy":          {"c"},
		},
	})
}
