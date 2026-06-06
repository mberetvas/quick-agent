package llm

import (
	"strings"
	"testing"

	"github.com/mberetvas/quick-agent/internal/config"
)

func TestPromptTemplates(t *testing.T) {
	reg := NewPromptRegistry(config.DefaultPromptsConfig())

	// Test template rendering
	tmpl := reg.Get("refine")
	if tmpl.Name != "refine" {
		t.Fatalf("expected refine template, got %s", tmpl.Name)
	}

	testText := "speling mistaks"
	rendered := tmpl.Render(testText)

	if !strings.Contains(rendered, testText) {
		t.Errorf("rendered prompt should include the source clipboard text, got: %s", rendered)
	}
}

func TestPromptTemplates_TranslateRenderWithOptions(t *testing.T) {
	reg := NewPromptRegistry(config.DefaultPromptsConfig())
	tmpl := reg.Get("translate")

	rendered := tmpl.RenderWithOptions("Bonjour", map[string]string{"Language": "English"})
	if !strings.Contains(rendered, "English") {
		t.Errorf("expected language in rendered translate prompt, got: %s", rendered)
	}
	if !strings.Contains(rendered, "Bonjour") {
		t.Errorf("expected text in rendered translate prompt, got: %s", rendered)
	}
}

func TestPromptTemplates_ConfigOverride(t *testing.T) {
	cfg := config.DefaultPromptsConfig()
	cfg.Refine = "Custom refine: {{.Text}}"
	reg := NewPromptRegistry(cfg)

	rendered := reg.Get("refine").Render("hello")
	if rendered != "Custom refine: hello" {
		t.Errorf("expected custom refine prompt, got: %s", rendered)
	}
}

func TestPromptTemplates_Fallback(t *testing.T) {
	reg := NewPromptRegistry(config.DefaultPromptsConfig())

	// Unknown name falls back to passthrough template.
	tmpl := reg.Get("nonexistent_special_template")
	if tmpl.Render("passthrough") != "passthrough" {
		t.Errorf("fallback should render original text unmodified")
	}
}
