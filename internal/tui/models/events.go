package models

// ShowOptionsEvent requests navigation to the options menu.
type ShowOptionsEvent struct{}

// BackEvent requests popping to the previous view on the stack.
type BackEvent struct{}

// ShowViewEvent requests navigation to a named follow-up view (stub targets for now).
type ShowViewEvent struct {
	View string
}

// View name constants used with ShowViewEvent.
const (
	ViewNameResult         = "result"
	ViewNameLanguagePicker = "language_picker"
	ViewNameCustomPrompt   = "custom_prompt"
)
