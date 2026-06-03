package llm

import (
	"strings"
	"testing"
)

func TestPromptTemplates(t *testing.T) {
	reg := NewPromptRegistry()

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

	// Test fallback handler 'custom'
	unknownTmpl := reg.Get("nonexistent_special_template")
	if unknownTmpl.Name != "custom" {
		t.Errorf("expected nonexistent template query to fallback to 'custom', got: %s", unknownTmpl.Name)
	}

	renderedCustom := unknownTmpl.Render("free prompt style template")
	if renderedCustom != "free prompt style template" {
		t.Errorf("custom template should render original text unmodified, got: %s", renderedCustom)
	}
}
