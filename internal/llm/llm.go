package llm

import (
	"context"
	"strings"
)

// LLMClient defines the unified interface for interacting with LLM backends.
type LLMClient interface {
	// Generate takes the formatted prompt and returns a read-only channel of streamed tokens.
	Generate(ctx context.Context, prompt string) (<-chan string, error)

	// HealthCheck returns nil if the backend is reachable and functioning, else an error.
	HealthCheck(ctx context.Context) error
}

// Template represents a registered prompt template.
type Template struct {
	Name        string
	Description string
	UserPrompt  string
}

// Render replaces any placeholder patterns (like {{.Text}}) to construct the final prompt.
func (t Template) Render(text string) string {
	return strings.ReplaceAll(t.UserPrompt, "{{.Text}}", text)
}

// PromptRegistry manages the registry of built-in templates.
type PromptRegistry struct {
	templates map[string]Template
}

// NewPromptRegistry returns a pre-configured PromptRegistry with built-in templates.
func NewPromptRegistry() *PromptRegistry {
	reg := &PromptRegistry{
		templates: make(map[string]Template),
	}

	reg.Register(Template{
		Name:        "refine",
		Description: "Improve grammar, spelling, and clarity of the clipboard text",
		UserPrompt:  "Please refine, format, and correct the following text, improving grammar, spelling, and readability without adding empty introductory conversational filler. Output only the improved text:\n\n{{.Text}}",
	})

	reg.Register(Template{
		Name:        "translate",
		Description: "Translate the clipboard text to English",
		UserPrompt:  "Please translate the following text into clear, modern English. Do not add introductory or conversational filler, print only the translated output:\n\n{{.Text}}",
	})

	reg.Register(Template{
		Name:        "summarize",
		Description: "Generate a concise summary of the clipboard text",
		UserPrompt:  "Please generate a concise, bullet-pointed summary of the major key points from the following text:\n\n{{.Text}}",
	})

	reg.Register(Template{
		Name:        "explain",
		Description: "Explain code snippets or technical concepts",
		UserPrompt:  "Please explain the technical concepts or code snippet shown below simply and clearly, using markdown formatting:\n\n{{.Text}}",
	})

	reg.Register(Template{
		Name:        "custom",
		Description: "Generic prompt placeholder",
		UserPrompt:  "{{.Text}}",
	})

	return reg
}

// Register registers a new template.
func (r *PromptRegistry) Register(t Template) {
	r.templates[t.Name] = t
}

// Get retrieves a template by name. Returns the 'custom' template as fallback if not found.
func (r *PromptRegistry) Get(name string) Template {
	if t, ok := r.templates[name]; ok {
		return t
	}
	return r.templates["custom"]
}
