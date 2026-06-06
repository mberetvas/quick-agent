package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mberetvas/quick-agent/internal/tui/styles"
)

// ActionID identifies a prompt action.
type ActionID string

const (
	ActionRefine    ActionID = "refine"
	ActionTranslate ActionID = "translate"
	ActionSummarize ActionID = "summarize"
	ActionExplain   ActionID = "explain"
)

// Action is one selectable option in the options menu.
type Action struct {
	ID    ActionID
	Label string
}

// DefaultActions is the built-in action list shown in the options view.
var DefaultActions = []Action{
	{ID: ActionRefine, Label: "Refine"},
	{ID: ActionTranslate, Label: "Translate"},
	{ID: ActionSummarize, Label: "Summarize"},
	{ID: ActionExplain, Label: "Explain"},
}

// OptionsModel renders the action selection menu.
type OptionsModel struct {
	theme   styles.Theme
	keys    KeyMap
	actions []Action
	cursor  int
}

// NewOptionsModel creates an options menu model.
func NewOptionsModel(theme styles.Theme, keys KeyMap) *OptionsModel {
	return &OptionsModel{
		theme:   theme,
		keys:    keys,
		actions: DefaultActions,
		cursor:  0,
	}
}

func (m OptionsModel) Init() tea.Cmd {
	return nil
}

func (m OptionsModel) Update(msg tea.Msg) (OptionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch {
		case m.keys.Match("navigate_up", key):
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.actions) - 1
			}
		case m.keys.Match("navigate_down", key):
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case m.keys.Match("select", key):
			action := m.actions[m.cursor].ID
			return m, func() tea.Msg { return ActionSelectedEvent{Action: action} }
		case m.keys.Match("back", key):
			return m, func() tea.Msg { return BackEvent{} }
		}
	}
	return m, nil
}

func (m OptionsModel) View() string {
	var sb strings.Builder

	sb.WriteString(m.theme.Header.Render("Select an action") + "\n\n")

	for i, action := range m.actions {
		line := fmt.Sprintf("%d. %s", i+1, action.Label)
		if i == m.cursor {
			sb.WriteString(m.theme.Selected.Render("> "+line) + "\n")
		} else {
			sb.WriteString(m.theme.Item.Render("  "+line) + "\n")
		}
	}

	sb.WriteString("\n" + m.theme.Footer.Render("j/k or arrows: navigate · enter: select · esc/q: back"))
	return sb.String()
}
