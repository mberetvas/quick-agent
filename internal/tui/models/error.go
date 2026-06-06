package models

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	apperrors "github.com/mberetvas/quick-agent/internal/errors"
	"github.com/mberetvas/quick-agent/internal/tui/styles"
)

// ErrorModel renders a structured error with severity-appropriate styling.
type ErrorModel struct {
	err   apperrors.UserError
	theme styles.Theme
	keys  KeyMap
}

// NewErrorModel creates an error view for the given UserError.
func NewErrorModel(err apperrors.UserError, theme styles.Theme, keys KeyMap) *ErrorModel {
	return &ErrorModel{err: err, theme: theme, keys: keys}
}

func (m ErrorModel) Init() tea.Cmd {
	return nil
}

func (m ErrorModel) Update(msg tea.Msg) (ErrorModel, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	key := keyMsg.String()
	switch {
	case key == "r":
		return m, func() tea.Msg { return RetryEvent{} }
	case m.keys.Match("back", key):
		return m, func() tea.Msg { return BackEvent{} }
	}
	return m, nil
}

func (m ErrorModel) View() string {
	var sb strings.Builder

	titleColor := m.theme.Error
	if m.err.Severity == apperrors.SeverityWarning {
		titleColor = m.theme.Warning
	}
	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Padding(0, 1)
	sb.WriteString(titleStyle.Render(m.err.Title) + "\n\n")

	sb.WriteString(m.theme.NormalText.Render(m.err.Message) + "\n")

	sb.WriteString("\n" + m.theme.Footer.Render("r: retry · esc: back"))
	return sb.String()
}
