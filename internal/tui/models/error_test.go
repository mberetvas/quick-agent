package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	apperrors "github.com/mberetvas/quick-agent/internal/errors"
	"github.com/mberetvas/quick-agent/internal/tui/styles"
)

func TestErrorModel_keys(t *testing.T) {
	keys := testKeyMap()
	theme := styles.DefaultTheme()
	err := apperrors.ErrLLMTimeout()
	m := *NewErrorModel(err, theme, keys)

	t.Run("r fires RetryEvent", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
		if cmd == nil {
			t.Fatal("expected RetryEvent command")
		}
		if _, ok := cmd().(RetryEvent); !ok {
			t.Fatalf("expected RetryEvent, got %T", cmd())
		}
	})

	t.Run("esc fires BackEvent", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
		if cmd == nil {
			t.Fatal("expected BackEvent command")
		}
		if _, ok := cmd().(BackEvent); !ok {
			t.Fatalf("expected BackEvent, got %T", cmd())
		}
	})
}

func TestErrorModel_View(t *testing.T) {
	theme := styles.DefaultTheme()
	err := apperrors.ErrLLMConnection("ollama", "http://localhost:11434", nil)
	m := NewErrorModel(err, theme, testKeyMap())

	view := m.View()
	if !strings.Contains(view, err.Title) {
		t.Errorf("view should contain title %q, got:\n%s", err.Title, view)
	}
	if !strings.Contains(view, err.Message) {
		t.Errorf("view should contain message, got:\n%s", view)
	}
	if !strings.Contains(view, "retry") {
		t.Errorf("view should contain retry hint, got:\n%s", view)
	}
}
