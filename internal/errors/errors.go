// Package apperrors defines structured user-facing error types for the application.
package apperrors

import "fmt"

// Severity levels for UserError.
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// UserError is a structured error carrying a user-visible title, message, and severity.
type UserError struct {
	Title    string
	Message  string
	Severity string
	Err      error
}

func (e UserError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e UserError) Unwrap() error {
	return e.Err
}

// ErrLLMTimeout returns a UserError for a request that timed out.
func ErrLLMTimeout() UserError {
	return UserError{
		Title:    "Request Timed Out",
		Message:  "The LLM backend did not respond in time. Check that it is running and try again.",
		Severity: SeverityError,
	}
}

// ErrLLMConnection returns a UserError for a connection failure to the LLM backend.
func ErrLLMConnection(backend, url string, err error) UserError {
	return UserError{
		Title:    "Connection Failed",
		Message:  fmt.Sprintf("Could not connect to %s at %s. Make sure the service is running.", backend, url),
		Severity: SeverityError,
		Err:      err,
	}
}

// ErrInvalidAPIKey returns a UserError for an authentication failure.
func ErrInvalidAPIKey() UserError {
	return UserError{
		Title:    "Invalid API Key",
		Message:  "The API key was rejected. Check your key configuration and try again.",
		Severity: SeverityError,
	}
}

// ErrClipboardEmpty returns a UserError when no clipboard text is available.
func ErrClipboardEmpty() UserError {
	return UserError{
		Title:    "Clipboard Empty",
		Message:  "No text was found in the clipboard. Copy some text and try again.",
		Severity: SeverityWarning,
	}
}
