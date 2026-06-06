package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mberetvas/quick-agent/internal/config"
	apperrors "github.com/mberetvas/quick-agent/internal/errors"
	"github.com/mberetvas/quick-agent/internal/tui/models"
)

type testLLM struct {
	tokens []string
}

func (m *testLLM) Generate(ctx context.Context, prompt string) (<-chan string, <-chan error, error) {
	ch := make(chan string, len(m.tokens))
	for _, t := range m.tokens {
		ch <- t
	}
	close(ch)
	errCh := make(chan error)
	close(errCh)
	return ch, errCh, nil
}

func (m *testLLM) HealthCheck(ctx context.Context) error { return nil }

func testModel(text string) *Model {
	cfg := config.Default()
	cfg.TUI.StreamingDelayMS = 0
	return NewModel(text, cfg, &testLLM{tokens: []string{"ok"}})
}

func TestViewStack_push_options_from_initial(t *testing.T) {
	m := testModel("hello")

	updated, _ := m.Update(models.ShowOptionsEvent{})
	root := updated.(*Model)

	if root.currentView != ViewOptions {
		t.Errorf("expected ViewOptions, got %v", root.currentView)
	}
	if len(root.viewStack) != 2 {
		t.Fatalf("expected stack depth 2, got %d", len(root.viewStack))
	}
	if root.viewStack[0] != ViewInitial || root.viewStack[1] != ViewOptions {
		t.Errorf("unexpected stack: %v", root.viewStack)
	}
}

func TestViewStack_pop_back_to_initial(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewOptions)

	updated, _ := m.Update(models.BackEvent{})
	root := updated.(*Model)

	if root.currentView != ViewInitial {
		t.Errorf("expected ViewInitial, got %v", root.currentView)
	}
	if len(root.viewStack) != 1 {
		t.Fatalf("expected stack depth 1, got %d", len(root.viewStack))
	}
}

func TestViewStack_action_selected_pushes_result(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewOptions)

	updated, cmd := m.Update(models.ActionSelectedEvent{Action: models.ActionRefine})
	root := updated.(*Model)

	if root.currentView != ViewResult {
		t.Errorf("expected ViewResult, got %v", root.currentView)
	}
	if root.ResultModel == nil {
		t.Fatal("expected ResultModel")
	}
	if cmd == nil {
		t.Fatal("expected Init cmd for result model")
	}
}

func TestViewStack_action_selected_streams_result(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewOptions)

	updated, cmd := m.Update(models.ActionSelectedEvent{Action: models.ActionRefine})
	root := updated.(*Model)

	for i := 0; i < 20 && cmd != nil; i++ {
		msg := cmd()
		if msg == nil {
			break
		}
		var next tea.Cmd
		updated, next = root.Update(msg)
		root = updated.(*Model)
		cmd = next
	}

	view := root.View()
	if !strings.Contains(view, "ok") {
		t.Errorf("expected streamed result in view, got:\n%s", view)
	}
}

func TestInitial_enter_navigates_to_options(t *testing.T) {
	m := testModel("hello")

	_, cmd := m.InitialModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected ShowOptionsEvent command")
	}

	updated, _ := m.Update(cmd())
	root := updated.(*Model)

	if root.currentView != ViewOptions {
		t.Errorf("expected options view after enter flow, got %v", root.currentView)
	}
}

func TestOptions_esc_pops_to_initial(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewOptions)

	newOpts, cmd := m.OptionsModel.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m.OptionsModel = &newOpts
	if cmd == nil {
		t.Fatal("expected back command from options")
	}

	updated, _ := m.Update(cmd())
	final := updated.(*Model)

	if final.currentView != ViewInitial {
		t.Errorf("expected initial view after back, got %v", final.currentView)
	}
}

func TestResultView_esc_pops_to_options(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewOptions)
	updated, _ := m.Update(models.ActionSelectedEvent{Action: models.ActionRefine})
	m = updated.(*Model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	final := updated.(*Model)

	if final.currentView != ViewOptions {
		t.Errorf("expected options after result back, got %v", final.currentView)
	}
}

func TestNewModel_nilConfigUsesDefaults(t *testing.T) {
	m := NewModel("text", nil, &testLLM{})
	if m.cfg == nil {
		t.Fatal("expected default config")
	}
	if m.InitialModel == nil || m.OptionsModel == nil {
		t.Fatal("expected child models")
	}
}

func TestModel_Init_delegatesToInitial(t *testing.T) {
	m := testModel("hello")
	_ = m.Init() // InitialModel.Init may return nil; ensure no panic
}

func TestModel_ctrlC_quits(t *testing.T) {
	m := testModel("hello")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
}

func TestModel_windowSize_updatesDimensions(t *testing.T) {
	m := testModel("hello")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	root := updated.(*Model)
	if root.width != 120 || root.height != 40 {
		t.Errorf("size = %dx%d", root.width, root.height)
	}
}

func TestModel_showError_pushesErrorView(t *testing.T) {
	m := testModel("hello")
	updated, _ := m.Update(models.ShowErrorEvent{Err: apperrors.UserError{
		Title: "Test", Message: "boom", Severity: apperrors.SeverityError,
	}})
	root := updated.(*Model)
	if root.currentView != ViewError {
		t.Errorf("view = %v, want ViewError", root.currentView)
	}
	if root.ErrorModel == nil {
		t.Fatal("expected ErrorModel")
	}
	if !strings.Contains(root.View(), "boom") {
		t.Errorf("view = %q", root.View())
	}
}

func TestModel_showViewEvent_result(t *testing.T) {
	m := testModel("hello")
	updated, _ := m.Update(models.ShowViewEvent{View: models.ViewNameResult})
	root := updated.(*Model)
	if root.currentView != ViewResult {
		t.Errorf("view = %v, want ViewResult", root.currentView)
	}
}

func TestModel_view_loadingResult(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewResult)
	if view := m.View(); view != "Loading..." {
		t.Errorf("view = %q, want Loading...", view)
	}
}

func TestViewFromName(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   ViewType
		wantOK bool
	}{
		{name: "result", input: models.ViewNameResult, want: ViewResult, wantOK: true},
		{name: "unknown", input: "options", want: ViewInitial, wantOK: false},
		{name: "empty", input: "", want: ViewInitial, wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := viewFromName(tt.input)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("viewFromName(%q) = (%v, %v), want (%v, %v)", tt.input, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestModel_view_errorWithoutModel(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewError)
	m.ErrorModel = nil
	if view := m.View(); view != "Error" {
		t.Errorf("view = %q, want Error", view)
	}
}

func TestModel_view_initial(t *testing.T) {
	m := testModel("hello clipboard")
	view := m.View()
	if !strings.Contains(view, "hello clipboard") {
		t.Errorf("initial view = %q", view)
	}
}

func TestModel_showViewEvent_unknownNoop(t *testing.T) {
	m := testModel("hello")
	before := m.currentView
	updated, cmd := m.Update(models.ShowViewEvent{View: "unknown"})
	root := updated.(*Model)
	if root.currentView != before {
		t.Errorf("view changed from %v to %v", before, root.currentView)
	}
	if cmd != nil {
		t.Fatal("expected nil cmd for unknown view")
	}
}

func TestModel_showViewEvent_options(t *testing.T) {
	m := testModel("hello")
	updated, _ := m.Update(models.ShowOptionsEvent{})
	root := updated.(*Model)
	if root.currentView != ViewOptions {
		t.Errorf("view = %v, want ViewOptions", root.currentView)
	}
}

func TestModel_unknownKey_noop_on_options(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewOptions)
	before := m.currentView
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyF1})
	root := updated.(*Model)
	if root.currentView != before {
		t.Errorf("view changed from %v to %v on unknown key", before, root.currentView)
	}
}

func TestModel_retryEvent(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewOptions)
	updated, _ := m.Update(models.ActionSelectedEvent{Action: models.ActionRefine})
	m = updated.(*Model)

	updated, cmd := m.Update(models.RetryEvent{})
	root := updated.(*Model)
	if root.currentView != ViewOptions {
		t.Errorf("view = %v, want ViewOptions after retry pop", root.currentView)
	}
	if cmd == nil {
		t.Fatal("expected retry cmd")
	}
}

func TestModel_view_default_unimplemented(t *testing.T) {
	m := testModel("hello")
	m.currentView = ViewType(99)
	if view := m.View(); view != "View not implemented yet." {
		t.Errorf("view = %q", view)
	}
}

func TestOptions_view_lists_all_actions(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewOptions)

	view := m.View()
	for _, label := range []string{"Refine", "Translate", "Summarize", "Explain"} {
		if !strings.Contains(view, label) {
			t.Errorf("options view missing %q", label)
		}
	}
	if strings.Contains(view, "Custom Prompt") {
		t.Error("options view should not list Custom Prompt")
	}
}
