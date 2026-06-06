package apperrors

import (
	"errors"
	"strings"
	"testing"
)

func TestUserError_Error(t *testing.T) {
	wrapped := errors.New("inner")
	e := UserError{Title: "T", Message: "msg", Severity: SeverityError, Err: wrapped}
	if !strings.Contains(e.Error(), "inner") {
		t.Errorf("expected wrapped error in message, got: %s", e.Error())
	}
	if !errors.Is(e, wrapped) {
		t.Error("Unwrap should expose inner error")
	}
}

func TestFactories(t *testing.T) {
	tests := []struct {
		name     string
		err      UserError
		wantTitle string
	}{
		{"timeout", ErrLLMTimeout(), "Request Timed Out"},
		{"connection", ErrLLMConnection("ollama", "http://localhost:11434", errors.New("refused")), "Connection Failed"},
		{"apikey", ErrInvalidAPIKey(), "Invalid API Key"},
		{"clipboard", ErrClipboardEmpty(), "Clipboard Empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", tt.err.Title, tt.wantTitle)
			}
			if tt.err.Severity == "" {
				t.Error("Severity should be set")
			}
		})
	}
}
