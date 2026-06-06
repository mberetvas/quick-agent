package models

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mberetvas/quick-agent/internal/clipboard"
	"github.com/mberetvas/quick-agent/internal/config"
	apperrors "github.com/mberetvas/quick-agent/internal/errors"
	"github.com/mberetvas/quick-agent/internal/llm"
	"github.com/mberetvas/quick-agent/internal/tui/styles"
)

type tokenMsg struct {
	token string
}

type streamStartedMsg struct {
	tokens <-chan string
	errs   <-chan error
}

type streamDoneMsg struct{}

type streamErrMsg struct {
	err error
}

type copiedMsg struct{}

type flushDisplayMsg struct{}

// ResultModel displays streaming LLM output and supports copy-to-clipboard.
type ResultModel struct {
	action        ActionID
	clipboardText string
	language      string
	result        string
	pending       string
	streaming     bool
	streamErr     error
	copied        bool

	llmClient      llm.LLMClient
	prompts        *llm.PromptRegistry
	theme          styles.Theme
	keys           KeyMap
	cb             clipboard.Clipboard
	streamingDelay time.Duration

	tokens <-chan string
	errs   <-chan error
}

// NewResultModel creates a result view for the given action and dependencies.
func NewResultModel(
	action ActionID,
	clipboardText string,
	client llm.LLMClient,
	prompts *llm.PromptRegistry,
	theme styles.Theme,
	keys KeyMap,
	cb clipboard.Clipboard,
	streamingDelayMS int,
	language string,
) *ResultModel {
	if prompts == nil {
		prompts = llm.NewPromptRegistry(config.DefaultPromptsConfig())
	}
	if cb == nil {
		cb = clipboard.SystemClipboard{}
	}
	return &ResultModel{
		action:         action,
		clipboardText:  clipboardText,
		language:       language,
		llmClient:      client,
		prompts:        prompts,
		theme:          theme,
		keys:           keys,
		cb:             cb,
		streamingDelay: time.Duration(streamingDelayMS) * time.Millisecond,
	}
}

func (m ResultModel) Init() tea.Cmd {
	return m.startGenerateCmd()
}

func (m ResultModel) startGenerateCmd() tea.Cmd {
	return func() tea.Msg {
		if m.llmClient == nil {
			return streamErrMsg{err: fmt.Errorf("LLM client not configured")}
		}
		tmpl := m.prompts.Get(string(m.action))
		var prompt string
		if m.action == ActionTranslate && m.language != "" {
			prompt = tmpl.RenderWithOptions(m.clipboardText, map[string]string{"Language": m.language})
		} else {
			prompt = tmpl.Render(m.clipboardText)
		}
		tokens, errs, err := m.llmClient.Generate(context.Background(), prompt)
		if err != nil {
			return streamErrMsg{err: err}
		}
		return streamStartedMsg{tokens: tokens, errs: errs}
	}
}

// Retry resets streaming state and re-runs the generation command.
func (m *ResultModel) Retry() tea.Cmd {
	m.result = ""
	m.pending = ""
	m.streaming = false
	m.streamErr = nil
	m.tokens = nil
	m.errs = nil
	m.copied = false
	return m.startGenerateCmd()
}

func (m ResultModel) readNextTokenCmd() tea.Cmd {
	tokens := m.tokens
	errs := m.errs
	return func() tea.Msg {
		if tokens == nil {
			return streamDoneMsg{}
		}
		select {
		case tok, ok := <-tokens:
			if !ok {
				return streamDoneMsg{}
			}
			return tokenMsg{token: tok}
		case err, ok := <-errs:
			if ok && err != nil {
				return streamErrMsg{err: err}
			}
			select {
			case tok, ok := <-tokens:
				if !ok {
					return streamDoneMsg{}
				}
				return tokenMsg{token: tok}
			default:
				return streamDoneMsg{}
			}
		}
	}
}

func (m ResultModel) tickCmd() tea.Cmd {
	if m.streamingDelay <= 0 {
		return nil
	}
	return tea.Tick(m.streamingDelay, func(time.Time) tea.Msg {
		return flushDisplayMsg{}
	})
}

func (m ResultModel) Update(msg tea.Msg) (ResultModel, tea.Cmd) {
	switch msg := msg.(type) {
	case streamStartedMsg:
		m.tokens = msg.tokens
		m.errs = msg.errs
		m.streaming = true
		m.streamErr = nil
		cmds := []tea.Cmd{m.readNextTokenCmd()}
		if tick := m.tickCmd(); tick != nil {
			cmds = append(cmds, tick)
		}
		return m, tea.Batch(cmds...)

	case tokenMsg:
		if m.streamingDelay <= 0 {
			m.result += msg.token
		} else {
			m.pending += msg.token
		}
		return m, m.readNextTokenCmd()

	case flushDisplayMsg:
		if m.pending != "" {
			m.result += m.pending
			m.pending = ""
		}
		var cmd tea.Cmd
		if m.streaming {
			cmd = m.tickCmd()
		}
		return m, cmd

	case streamDoneMsg:
		if m.pending != "" {
			m.result += m.pending
			m.pending = ""
		}
		m.streaming = false
		m.tokens = nil
		m.errs = nil
		return m, nil

	case streamErrMsg:
		m.streaming = false
		m.tokens = nil
		m.errs = nil
		userErr := classifyStreamError(msg.err)
		return m, func() tea.Msg { return ShowErrorEvent{Err: userErr} }

	case copiedMsg:
		m.copied = false
		return m, nil

	case tea.KeyMsg:
		key := msg.String()
		switch {
		case m.keys.Match("copy", key):
			text := m.result
			if m.pending != "" {
				text += m.pending
			}
			if text == "" {
				return m, nil
			}
			if err := m.cb.Set(text); err != nil {
				userErr := apperrors.UserError{
					Title:    "Copy Failed",
					Message:  err.Error(),
					Severity: apperrors.SeverityError,
					Err:      err,
				}
				return m, func() tea.Msg { return ShowErrorEvent{Err: userErr} }
			}
			m.copied = true
			return m, tea.Tick(800*time.Millisecond, func(time.Time) tea.Msg { return copiedMsg{} })
		case m.keys.Match("back", key):
			return m, func() tea.Msg { return BackEvent{} }
		}
	}

	return m, nil
}

func (m ResultModel) View() string {
	var sb strings.Builder

	title := actionTitle(m.action)
	sb.WriteString(m.theme.Header.Render(title) + "\n\n")

	body := m.result
	if m.streaming && m.pending != "" {
		body += m.pending
	}
	if body == "" && m.streaming {
		body = m.theme.NormalText.Render("Waiting for response...")
	} else if body == "" {
		body = m.theme.NormalText.Render("(no output)")
	} else {
		body = m.theme.NormalText.Render(body)
	}
	sb.WriteString(body + "\n")

	if m.streaming {
		warnStyle := lipgloss.NewStyle().Foreground(m.theme.Warning)
		sb.WriteString("\n" + warnStyle.Render("Streaming..."))
	}
	if m.copied {
		okStyle := lipgloss.NewStyle().Foreground(m.theme.Success)
		sb.WriteString("\n" + okStyle.Render("Copied to clipboard!"))
	}

	footer := "c: copy · esc/q: back"
	if m.streaming {
		footer = "Streaming... · " + footer
	}
	sb.WriteString("\n\n" + m.theme.Footer.Render(footer))

	return sb.String()
}

func actionTitle(id ActionID) string {
	switch id {
	case ActionRefine:
		return "Refine"
	case ActionSummarize:
		return "Summarize"
	case ActionExplain:
		return "Explain"
	case ActionTranslate:
		return "Translate"
	default:
		return string(id)
	}
}

// classifyStreamError maps a raw error to a structured UserError.
func classifyStreamError(err error) apperrors.UserError {
	if err == nil {
		return apperrors.UserError{Title: "Unknown Error", Message: "An unknown error occurred.", Severity: apperrors.SeverityError}
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "timeout") || strings.Contains(msg, "context deadline"):
		return apperrors.ErrLLMTimeout()
	case strings.Contains(msg, "connection refused") || strings.Contains(msg, "no such host") || strings.Contains(msg, "dial "):
		return apperrors.ErrLLMConnection("backend", "", err)
	case strings.Contains(msg, "401") || strings.Contains(msg, "unauthorized") || strings.Contains(msg, "invalid api key"):
		return apperrors.ErrInvalidAPIKey()
	default:
		return apperrors.UserError{
			Title:    "Generation Failed",
			Message:  msg,
			Severity: apperrors.SeverityError,
			Err:      err,
		}
	}
}
