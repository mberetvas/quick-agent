package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourname/clipboard-tui/internal/config"
	"github.com/yourname/clipboard-tui/internal/tui/models"
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

func TestOptions_view_lists_all_actions(t *testing.T) {
	m := testModel("hello")
	m.PushView(ViewOptions)

	view := m.View()
	for _, label := range []string{"Refine", "Translate", "Summarize", "Explain", "Custom Prompt"} {
		if !strings.Contains(view, label) {
			t.Errorf("options view missing %q", label)
		}
	}
}
