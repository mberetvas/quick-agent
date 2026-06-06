package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mberetvas/quick-agent/internal/config"
	"github.com/mberetvas/quick-agent/internal/tui/styles"
)

func testKeyMap() KeyMap {
	return NewKeyMap(config.DefaultTUIConfig())
}

func TestNewOptionsModel(t *testing.T) {
	m := NewOptionsModel(styles.DefaultTheme(), testKeyMap())

	if m.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.cursor)
	}
	if len(m.actions) != len(DefaultActions) {
		t.Errorf("expected %d actions, got %d", len(DefaultActions), len(m.actions))
	}
}

func TestOptionsModel_Update_Navigation(t *testing.T) {
	m := *NewOptionsModel(styles.DefaultTheme(), testKeyMap())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	want := len(m.actions) - 1
	if m.cursor != want {
		t.Errorf("expected cursor %d after up wrap, got %d", want, m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after down wrap, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.cursor)
	}
}

func TestOptionsModel_Update_Back(t *testing.T) {
	m := *NewOptionsModel(styles.DefaultTheme(), testKeyMap())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected BackEvent command")
	}
	if _, ok := cmd().(BackEvent); !ok {
		t.Fatalf("expected BackEvent, got %T", cmd())
	}
}

func TestOptionsModel_Update_Select(t *testing.T) {
	tests := []struct {
		name       string
		cursor     int
		wantAction ActionID
	}{
		{name: "refine", cursor: 0, wantAction: ActionRefine},
		{name: "translate", cursor: 1, wantAction: ActionTranslate},
		{name: "summarize", cursor: 2, wantAction: ActionSummarize},
		{name: "explain", cursor: 3, wantAction: ActionExplain},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := *NewOptionsModel(styles.DefaultTheme(), testKeyMap())
			m.cursor = tt.cursor

			_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if cmd == nil {
				t.Fatal("expected ActionSelectedEvent command")
			}
			ev, ok := cmd().(ActionSelectedEvent)
			if !ok {
				t.Fatalf("expected ActionSelectedEvent, got %T", cmd())
			}
			if ev.Action != tt.wantAction {
				t.Errorf("expected action %q, got %q", tt.wantAction, ev.Action)
			}
		})
	}
}

func TestOptionsModel_View_highlight(t *testing.T) {
	m := NewOptionsModel(styles.DefaultTheme(), testKeyMap())
	m.cursor = 2

	view := m.View()
	if !strings.Contains(view, "> 3. Summarize") {
		t.Errorf("expected highlighted Summarize line, got:\n%s", view)
	}
	if strings.Contains(view, "> 1. Refine") {
		t.Error("Refine should not be highlighted")
	}
	for _, label := range []string{"Refine", "Translate", "Summarize", "Explain"} {
		if !strings.Contains(view, label) {
			t.Errorf("expected label %q in view", label)
		}
	}
	if strings.Contains(view, "Custom Prompt") {
		t.Error("Custom Prompt should not appear in options view")
	}
}
