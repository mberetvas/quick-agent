package models

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/yourname/clipboard-tui/internal/tui/styles"
)

func TestNewInitialModel(t *testing.T) {
	theme := styles.DefaultTheme()
	text := "test clipboard text"
	m := NewInitialModel(text, theme)

	if m.clipboardText != text {
		t.Errorf("Expected clipboardText to be %q, got %q", text, m.clipboardText)
	}

	if m.cursor != 0 {
		t.Errorf("Expected initial cursor to be 0, got %d", m.cursor)
	}

	if len(m.options) == 0 {
		t.Error("Expected options to be initialized, got empty list")
	}
}

func TestInitialModel_Update_Navigation(t *testing.T) {
	theme := styles.DefaultTheme()
	m := *NewInitialModel("hello", theme)

	// Test cursor wraps around on key up
	msgUp := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	m, _ = m.Update(msgUp)
	expectedCursor := len(m.options) - 1
	if m.cursor != expectedCursor {
		t.Errorf("Expected cursor to wrap to %d, got %d", expectedCursor, m.cursor)
	}

	// Test cursor moves down on key down
	msgDown := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	m, _ = m.Update(msgDown)
	if m.cursor != 0 {
		t.Errorf("Expected cursor to move back to 0, got %d", m.cursor)
	}

	m, _ = m.Update(msgDown)
	if m.cursor != 1 {
		t.Errorf("Expected cursor to move to 1, got %d", m.cursor)
	}
}

func TestInitialModel_Update_Quit(t *testing.T) {
	theme := styles.DefaultTheme()
	m := *NewInitialModel("hello", theme)

	msgEsc := tea.KeyMsg{Type: tea.KeyEscape}
	_, cmd := m.Update(msgEsc)

	if cmd == nil {
		t.Fatal("Expected tea.Quit command, got nil")
	}

	msgRun := cmd()
	if _, ok := msgRun.(tea.QuitMsg); !ok {
		t.Errorf("Expected tea.QuitMsg, got %T", msgRun)
	}
}

func TestInitialModel_View(t *testing.T) {
	theme := styles.DefaultTheme()
	text := "special testing content"
	m := NewInitialModel(text, theme)

	viewStr := m.View()

	if !strings.Contains(viewStr, text) {
		t.Errorf("Expected view to display clipboard content %q", text)
	}

	if !strings.Contains(viewStr, "Refactor/Improve") {
		t.Error("Expected view to show options list")
	}
}
