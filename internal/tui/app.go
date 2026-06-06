package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mberetvas/quick-agent/internal/clipboard"
	"github.com/mberetvas/quick-agent/internal/config"
	"github.com/mberetvas/quick-agent/internal/llm"
	"github.com/mberetvas/quick-agent/internal/tui/models"
	"github.com/mberetvas/quick-agent/internal/tui/styles"
)

// ViewType represents different screens/views in the application.
type ViewType int

const (
	ViewInitial ViewType = iota
	ViewOptions
	ViewResult
	ViewError
)

// Model is the root bubbletea model managing the view stack.
type Model struct {
	currentView ViewType
	viewStack   []ViewType

	cfg           *config.Config
	clipboardText string
	llmClient     llm.LLMClient
	prompts       *llm.PromptRegistry
	theme         styles.Theme
	keys          models.KeyMap
	cb            clipboard.Clipboard

	InitialModel *models.InitialModel
	OptionsModel *models.OptionsModel
	ResultModel  *models.ResultModel
	ErrorModel   *models.ErrorModel

	width  int
	height int
}

// NewModel creates the root TUI model.
func NewModel(clipboardText string, cfg *config.Config, client llm.LLMClient) *Model {
	if cfg == nil {
		cfg = config.Default()
	}
	theme := styles.ThemeForConfig(cfg.TUI.Theme)
	keys := models.NewKeyMap(cfg.TUI)
	prompts := llm.NewPromptRegistry(cfg.Prompts)

	return &Model{
		currentView:   ViewInitial,
		viewStack:     []ViewType{ViewInitial},
		cfg:           cfg,
		clipboardText: clipboardText,
		llmClient:     client,
		prompts:       prompts,
		theme:         theme,
		keys:          keys,
		cb:            clipboard.SystemClipboard{},
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

	case models.ActionSelectedEvent:
		language := msg.Language
		if msg.Action == models.ActionTranslate && language == "" {
			language = m.cfg.Prompts.TranslateTargetLanguage
		}
		m.ResultModel = models.NewResultModel(
			msg.Action,
			m.clipboardText,
			m.llmClient,
			m.prompts,
			m.theme,
			m.keys,
			m.cb,
			m.cfg.TUI.StreamingDelayMS,
			language,
		)
		m.PushView(ViewResult)
		return m, m.ResultModel.Init()

	case models.ShowViewEvent:
		if v, ok := viewFromName(msg.View); ok {
			m.PushView(v)
		}
		return m, nil

	case models.ShowErrorEvent:
		m.ErrorModel = models.NewErrorModel(msg.Err, m.theme, m.keys)
		m.PushView(ViewError)
		return m, nil

	case models.RetryEvent:
		m.PopView() // back to ViewResult
		if m.ResultModel != nil {
			return m, m.ResultModel.Retry()
		}
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.keys.Match("back", keyMsg.String()) {
			switch m.currentView {
			case ViewResult:
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
	case ViewResult:
		if m.ResultModel != nil {
			var newResult models.ResultModel
			newResult, cmd = m.ResultModel.Update(msg)
			m.ResultModel = &newResult
		}
	case ViewError:
		if m.ErrorModel != nil {
			var newErr models.ErrorModel
			newErr, cmd = m.ErrorModel.Update(msg)
			m.ErrorModel = &newErr
		}
	}

	return m, cmd
}

func viewFromName(name string) (ViewType, bool) {
	switch name {
	case models.ViewNameResult:
		return ViewResult, true
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
		if m.ResultModel != nil {
			return m.ResultModel.View()
		}
		return "Loading..."
	case ViewError:
		if m.ErrorModel != nil {
			return m.ErrorModel.View()
		}
		return "Error"
	default:
		return "View not implemented yet."
	}
}
