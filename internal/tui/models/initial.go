package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yourname/clipboard-tui/internal/tui/styles"
)

// InitialModel shows the clipboard text and a path into the options menu.
type InitialModel struct {
	clipboardText string
	theme         styles.Theme
	keys          KeyMap
}

// NewInitialModel creates the initial TUI view.
func NewInitialModel(clipboardText string, theme styles.Theme, keys KeyMap) *InitialModel {
	return &InitialModel{
		clipboardText: clipboardText,
		theme:         theme,
		keys:          keys,
	}
}

func (m InitialModel) Init() tea.Cmd {
	return nil
}

func (m InitialModel) Update(msg tea.Msg) (InitialModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch {
		case m.keys.Match("select", key):
			return m, func() tea.Msg { return ShowOptionsEvent{} }
		case m.keys.Match("back", key):
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m InitialModel) View() string {
	var sb strings.Builder

	header := m.theme.Header.Render("📋 Clipboard AI Assistant")
	sb.WriteString(header + "\n\n")

	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(1, 2).
		Width(60)

	dispText := m.clipboardText
	if len(dispText) == 0 {
		dispText = "(Clipboard is empty)"
	} else {
		runes := []rune(dispText)
		if len(runes) > 200 {
			dispText = string(runes[:197]) + "..."
		}
	}

	contentView := contentStyle.Render("Clipboard text:\n" + m.theme.NormalText.Render(dispText))
	sb.WriteString(contentView + "\n\n")

	sb.WriteString(m.theme.Footer.Render("enter: choose action · esc/q: quit"))

	return sb.String()
}
