package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mberetvas/quick-agent/internal/tui/styles"
)

func TestNewInitialModel(t *testing.T) {
	theme := styles.DefaultTheme()
	text := "test clipboard text"
	m := NewInitialModel(text, theme, testKeyMap())

	if m.clipboardText != text {
		t.Errorf("Expected clipboardText to be %q, got %q", text, m.clipboardText)
	}
}

func TestInitialModel_Update_ShowOptions(t *testing.T) {
	m := *NewInitialModel("hello", styles.DefaultTheme(), testKeyMap())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected ShowOptionsEvent command")
	}
	if _, ok := cmd().(ShowOptionsEvent); !ok {
		t.Fatalf("expected ShowOptionsEvent, got %T", cmd())
	}
}

func TestInitialModel_Update_Quit(t *testing.T) {
	m := *NewInitialModel("hello", styles.DefaultTheme(), testKeyMap())

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
	m := NewInitialModel(text, theme, testKeyMap())

	viewStr := m.View()

	if !strings.Contains(viewStr, text) {
		t.Errorf("Expected view to display clipboard content %q", text)
	}

	if !strings.Contains(viewStr, "choose action") {
		t.Error("Expected view to prompt for action selection")
	}
}

func TestInitialModel_View_RuneSafeTruncate(t *testing.T) {
	theme := styles.DefaultTheme()
	var builder strings.Builder
	for i := 0; i < 125; i++ {
		builder.WriteString("世界")
	}
	text := builder.String()
	m := NewInitialModel(text, theme, testKeyMap())

	viewStr := m.View()

	if !strings.Contains(viewStr, "世...") {
		t.Errorf("Expected view to contain safely-truncated multi-byte text ending with '世...', but got: %s", viewStr)
	}

	if !strings.Contains(viewStr, "世界") {
		t.Errorf("Expected view to contain the repeated CJK characters '世界', but got: %s", viewStr)
	}

	if strings.Contains(viewStr, "\uFFFD") {
		t.Error("View contains replacement character \uFFFD (bad split UTF-8)")
	}
}
