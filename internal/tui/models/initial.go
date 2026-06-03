package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yourname/clipboard-tui/internal/tui/styles"
)

type InitialModel struct {
	clipboardText string
	theme         styles.Theme
	options       []string
	cursor        int
}

func NewInitialModel(clipboardText string, theme styles.Theme) *InitialModel {
	return &InitialModel{
		clipboardText: clipboardText,
		theme:         theme,
		options: []string{
			"1. Refactor/Improve",
			"2. Explain Code",
			"3. Translate",
			"4. Custom Prompt...",
		},
		cursor: 0,
	}
}

func (m InitialModel) Init() tea.Cmd {
	return nil
}

func (m InitialModel) Update(msg tea.Msg) (InitialModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.options) - 1
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}

		case "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m InitialModel) View() string {
	// Create headers and styles
	var sb strings.Builder

	header := m.theme.Header.Render("📋 Clipboard AI Assistant")
	sb.WriteString(header + "\n\n")

	// Render Clipboard Content Box
	contentStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.theme.Border).
		Padding(1, 2).
		Width(60)

	// Clean/Truncate clipboard text for rendering readability if needed
	dispText := m.clipboardText
	if len(dispText) == 0 {
		dispText = "(Clipboard is empty)"
	} else if len(dispText) > 200 {
		dispText = dispText[:197] + "..."
	}

	contentView := contentStyle.Render("Clipboard text:\n" + m.theme.NormalText.Render(dispText))
	sb.WriteString(contentView + "\n\n")

	// Render Options
	sb.WriteString("Select an action:\n")
	for i, option := range m.options {
		if i == m.cursor {
			selectedText := fmt.Sprintf("> %s", option)
			sb.WriteString(m.theme.Selected.Render(selectedText) + "\n")
		} else {
			unselectedText := fmt.Sprintf("  %s", option)
			sb.WriteString(m.theme.Item.Render(unselectedText) + "\n")
		}
	}

	sb.WriteString("\n" + m.theme.Footer.Render("Press j/k or up/down to navigate, q/esc to quit"))

	return sb.String()
}
