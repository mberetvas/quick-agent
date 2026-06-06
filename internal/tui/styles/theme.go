package styles

import "github.com/charmbracelet/lipgloss"

// Theme defines the interface or set of styles used in the TUI application.
type Theme struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Border    lipgloss.Color
	Success   lipgloss.Color
	Warning   lipgloss.Color
	Error     lipgloss.Color
	NormalText lipgloss.Style
	Header    lipgloss.Style
	Footer    lipgloss.Style
	Title     lipgloss.Style
	Item      lipgloss.Style
	Selected  lipgloss.Style
	InputBox  lipgloss.Style
}

// DefaultTheme returns a dark-mode oriented default theme.
func DefaultTheme() Theme {
	primary := lipgloss.Color("86")    // Cyan / Teal
	secondary := lipgloss.Color("205") // Pink / Magenta
	borderColor := lipgloss.Color("240") // Gray
	successColor := lipgloss.Color("46")  // Green
	warningColor := lipgloss.Color("214") // Yellow / Orange
	errColor := lipgloss.Color("196")     // Red

	return Theme{
		Primary:    primary,
		Secondary:  secondary,
		Border:     borderColor,
		Success:    successColor,
		Warning:    warningColor,
		Error:      errColor,
		NormalText: lipgloss.NewStyle().Foreground(lipgloss.Color("252")),

		Header: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(primary).
			Bold(true).
			Padding(0, 1),

		Item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			PaddingLeft(2),

		Selected: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			PaddingLeft(2),

		InputBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2),
	}
}

// LightTheme returns a light-mode theme with dark text and blue primary.
func LightTheme() Theme {
	primary := lipgloss.Color("27")     // Blue
	secondary := lipgloss.Color("129")  // Purple
	borderColor := lipgloss.Color("250") // Light gray
	successColor := lipgloss.Color("28")  // Dark green
	warningColor := lipgloss.Color("166") // Orange
	errColor := lipgloss.Color("160")     // Dark red

	return Theme{
		Primary:    primary,
		Secondary:  secondary,
		Border:     borderColor,
		Success:    successColor,
		Warning:    warningColor,
		Error:      errColor,
		NormalText: lipgloss.NewStyle().Foreground(lipgloss.Color("232")),

		Header: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")).
			Italic(true).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(primary).
			Bold(true).
			Padding(0, 1),

		Item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("232")).
			PaddingLeft(2),

		Selected: lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			PaddingLeft(2),

		InputBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(1, 2),
	}
}

// ThemeForConfig returns the Theme matching the config string.
// "light" returns LightTheme; anything else returns DefaultTheme.
func ThemeForConfig(s string) Theme {
	if s == "light" {
		return LightTheme()
	}
	return DefaultTheme()
}
