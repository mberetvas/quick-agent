package tui

import (
	"github.com/charmbracelet/bubbletea"

	"github.com/yourname/clipboard-tui/internal/tui/models"
	"github.com/yourname/clipboard-tui/internal/tui/styles"
)

// ViewType represents different screens/views in the application
type ViewType int

const (
	ViewInitial ViewType = iota
	ViewOptions
	ViewResult
	ViewError
	ViewSetup
	ViewLanguagePicker
	ViewCustomPrompt
)

type Model struct {
	// Current view
	currentView ViewType

	// View stack for back navigation
	viewStack []ViewType

	// Shared state
	clipboardText string
	theme         styles.Theme

	// View-specific models
	InitialModel *models.InitialModel

	// Dimensions
	width  int
	height int
}

func NewModel(clipboardText string) *Model {
	theme := styles.DefaultTheme()
	return &Model{
		currentView:   ViewInitial,
		viewStack:     []ViewType{ViewInitial},
		clipboardText: clipboardText,
		theme:         theme,
		InitialModel:  models.NewInitialModel(clipboardText, theme),
	}
}

func (m *Model) PushView(v ViewType) {
	m.currentView = v
	m.viewStack = append(m.viewStack, v)
}

func (m *Model) PopView() {
	if len(m.viewStack) > 1 {
		m.viewStack = m.viewStack[:len(m.viewStack)-1]
		m.currentView = m.viewStack[len(m.viewStack)-1]
	}
}

func (m *Model) Init() tea.Cmd {
	return m.InitialModel.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Propagate dimensions to submodels if needed
	}

	// Route actual logic depending on current view
	switch m.currentView {
	case ViewInitial:
		newInit, cmd := m.InitialModel.Update(msg)
		m.InitialModel = &newInit
		return m, cmd
	}

	return m, nil
}

func (m *Model) View() string {
	switch m.currentView {
	case ViewInitial:
		return m.InitialModel.View()
	default:
		return "View not implemented yet."
	}
}
