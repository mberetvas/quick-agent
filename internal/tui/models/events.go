package models

import apperrors "github.com/mberetvas/quick-agent/internal/errors"

// ShowOptionsEvent requests navigation to the options menu.
type ShowOptionsEvent struct{}

// BackEvent requests popping to the previous view on the stack.
type BackEvent struct{}

// ShowViewEvent requests navigation to a named follow-up view (stub targets for now).
type ShowViewEvent struct {
	View string
}

// ActionSelectedEvent requests the result view for an LLM action.
// Language is populated by app.go for the Translate action.
type ActionSelectedEvent struct {
	Action   ActionID
	Language string
}

// ShowErrorEvent requests navigation to the error view.
type ShowErrorEvent struct {
	Err apperrors.UserError
}

// RetryEvent requests a retry of the last LLM generation.
type RetryEvent struct{}

// View name constants used with ShowViewEvent.
const (
	ViewNameResult = "result"
)
