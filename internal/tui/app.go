package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"

	"github.com/yourname/clipboard-tui/internal/config"
	"github.com/yourname/clipboard-tui/internal/tui/models"
	"github.com/yourname/clipboard-tui/internal/tui/styles"
)

// ViewType represents different screens/views in the application.
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

// Model is the root bubbletea model managing the view stack.
type Model struct {
	currentView ViewType
	viewStack   []ViewType

	clipboardText string
	theme         styles.Theme
	keys          models.KeyMap

	InitialModel *models.InitialModel
	OptionsModel *models.OptionsModel

	width  int
	height int
}

// NewModel creates the root TUI model.
func NewModel(clipboardText string, tuiCfg config.TUIConfig) *Model {
	theme := styles.DefaultTheme()
	keys := models.NewKeyMap(tuiCfg)

	return &Model{
		currentView:   ViewInitial,
		viewStack:     []ViewType{ViewInitial},
		clipboardText: clipboardText,
		theme:         theme,
		keys:          keys,
		InitialModel:  models.NewInitialModel(clipboardText, theme, keys),
		OptionsModel:  models.NewOptionsModel(theme, keys),
	}
}

// PushView pushes a view onto the stack and makes it current.
func (m *Model) PushView(v ViewType) {
	m.currentView = v
	m.viewStack = append(m.viewStack, v)
}

// PopView removes the top view and returns to the previous one.
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
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case models.ShowOptionsEvent:
		m.PushView(ViewOptions)
		return m, nil

	case models.BackEvent:
		m.PopView()
		return m, nil

	case models.ShowViewEvent:
		if v, ok := viewFromName(msg.View); ok {
			m.PushView(v)
		}
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.keys.Match("back", keyMsg.String()) {
			switch m.currentView {
			case ViewResult, ViewLanguagePicker, ViewCustomPrompt:
				m.PopView()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.currentView {
	case ViewInitial:
		var newInit models.InitialModel
		newInit, cmd = m.InitialModel.Update(msg)
		m.InitialModel = &newInit
	case ViewOptions:
		var newOpts models.OptionsModel
		newOpts, cmd = m.OptionsModel.Update(msg)
		m.OptionsModel = &newOpts
	}

	return m, cmd
}

func viewFromName(name string) (ViewType, bool) {
	switch name {
	case models.ViewNameResult:
		return ViewResult, true
	case models.ViewNameLanguagePicker:
		return ViewLanguagePicker, true
	case models.ViewNameCustomPrompt:
		return ViewCustomPrompt, true
	default:
		return ViewInitial, false
	}
}

func (m *Model) View() string {
	switch m.currentView {
	case ViewInitial:
		return m.InitialModel.View()
	case ViewOptions:
		return m.OptionsModel.View()
	case ViewResult:
		return m.stubView("Result")
	case ViewLanguagePicker:
		return m.stubView("Language picker")
	case ViewCustomPrompt:
		return m.stubView("Custom prompt")
	default:
		return "View not implemented yet."
	}
}

func (m *Model) stubView(title string) string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.theme.Header.Render(title+" (coming soon)"),
		m.theme.NormalText.Render("This view is stubbed until a later slice."),
		m.theme.Footer.Render("esc/q: back"),
	)
}
