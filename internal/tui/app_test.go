package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourname/clipboard-tui/internal/config"
	"github.com/yourname/clipboard-tui/internal/tui/models"
)

func TestViewStack_push_options_from_initial(t *testing.T) {
	m := NewModel("hello", config.DefaultTUIConfig())

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
	m := NewModel("hello", config.DefaultTUIConfig())
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

func TestViewStack_select_action_pushes_stub(t *testing.T) {
	m := NewModel("hello", config.DefaultTUIConfig())
	m.PushView(ViewOptions)

	updated, _ := m.Update(models.ShowViewEvent{View: models.ViewNameResult})
	root := updated.(*Model)

	if root.currentView != ViewResult {
		t.Errorf("expected ViewResult, got %v", root.currentView)
	}
	if len(root.viewStack) != 3 {
		t.Fatalf("expected stack depth 3, got %d: %v", len(root.viewStack), root.viewStack)
	}
}

func TestInitial_enter_navigates_to_options(t *testing.T) {
	m := NewModel("hello", config.DefaultTUIConfig())

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
	m := NewModel("hello", config.DefaultTUIConfig())
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

func TestStubView_esc_pops_to_options(t *testing.T) {
	m := NewModel("hello", config.DefaultTUIConfig())
	m.PushView(ViewOptions)
	m.PushView(ViewResult)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	final := updated.(*Model)

	if final.currentView != ViewOptions {
		t.Errorf("expected options after stub back, got %v", final.currentView)
	}
}

func TestOptions_view_lists_all_actions(t *testing.T) {
	m := NewModel("hello", config.DefaultTUIConfig())
	m.PushView(ViewOptions)

	view := m.View()
	for _, label := range []string{"Refine", "Translate", "Summarize", "Explain", "Custom Prompt"} {
		if !strings.Contains(view, label) {
			t.Errorf("options view missing %q", label)
		}
	}
}
