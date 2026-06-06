package llm

import (
	"context"
	"strings"

	"github.com/mberetvas/quick-agent/internal/config"
)

// LLMClient defines the unified interface for interacting with LLM backends.
type LLMClient interface {
	// Generate takes the formatted prompt and returns a read-only channel of streamed tokens and an errors channel.
	Generate(ctx context.Context, prompt string) (<-chan string, <-chan error, error)

	// HealthCheck returns nil if the backend is reachable and functioning, else an error.
	HealthCheck(ctx context.Context) error
}

// Template represents a registered prompt template.
type Template struct {
	Name        string
	Description string
	UserPrompt  string
}

// Render replaces {{.Text}} to construct the final prompt.
func (t Template) Render(text string) string {
	return strings.ReplaceAll(t.UserPrompt, "{{.Text}}", text)
}

// RenderWithOptions replaces {{.Text}} and any additional {{.Key}} placeholders from opts.
func (t Template) RenderWithOptions(text string, opts map[string]string) string {
	result := strings.ReplaceAll(t.UserPrompt, "{{.Text}}", text)
	for k, v := range opts {
		result = strings.ReplaceAll(result, "{{."+k+"}}", v)
	}
	return result
}

// PromptRegistry manages the registry of built-in templates.
type PromptRegistry struct {
	templates map[string]Template
}

// builtinPrompts holds the hard-coded defaults, used when the config field is empty.
var builtinPrompts = map[string]string{
	"refine":    "Please refine, format, and correct the following text, improving grammar, spelling, and readability without adding empty introductory conversational filler. Output only the improved text:\n\n{{.Text}}",
	"translate": "Translate to {{.Language}}:\n\n{{.Text}}",
	"summarize": "Please generate a concise, bullet-pointed summary of the major key points from the following text:\n\n{{.Text}}",
	"explain":   "Please explain the technical concepts or code snippet shown below simply and clearly, using markdown formatting:\n\n{{.Text}}",
}

// NewPromptRegistry returns a PromptRegistry populated from cfg, falling back to built-ins for empty fields.
func NewPromptRegistry(cfg config.PromptsConfig) *PromptRegistry {
	pick := func(cfgVal, builtin string) string {
		if cfgVal != "" {
			return cfgVal
		}
		return builtin
	}

	reg := &PromptRegistry{templates: make(map[string]Template)}

	reg.Register(Template{
		Name:        "refine",
		Description: "Improve grammar, spelling, and clarity of the clipboard text",
		UserPrompt:  pick(cfg.Refine, builtinPrompts["refine"]),
	})

	reg.Register(Template{
		Name:        "translate",
		Description: "Translate the clipboard text to the configured target language",
		UserPrompt:  pick(cfg.Translate, builtinPrompts["translate"]),
	})

	reg.Register(Template{
		Name:        "summarize",
		Description: "Generate a concise summary of the clipboard text",
		UserPrompt:  pick(cfg.Summarize, builtinPrompts["summarize"]),
	})

	reg.Register(Template{
		Name:        "explain",
		Description: "Explain code snippets or technical concepts",
		UserPrompt:  pick(cfg.Explain, builtinPrompts["explain"]),
	})

	return reg
}

// Register registers a new template.
func (r *PromptRegistry) Register(t Template) {
	r.templates[t.Name] = t
}

// Get retrieves a template by name. Returns a passthrough template as fallback if not found.
func (r *PromptRegistry) Get(name string) Template {
	if t, ok := r.templates[name]; ok {
		return t
	}
	return Template{Name: name, UserPrompt: "{{.Text}}"}
}
