package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color palette for the TUI.
type Theme struct {
	// Primary colors
	Primary         lipgloss.Color // Main accent color
	PrimaryDim      lipgloss.Color // Dimmed for backgrounds
	Secondary       lipgloss.Color // Secondary accent color
	SecondaryDim    lipgloss.Color // Dimmed for backgrounds
	Success         lipgloss.Color // Success states
	SuccessDim      lipgloss.Color // Subtle background tints
	Warning         lipgloss.Color // Warning states
	WarningDim      lipgloss.Color // Subtle background tints
	Error           lipgloss.Color // Error states
	ErrorDim        lipgloss.Color // Subtle background tints
	Foreground      lipgloss.Color // Normal text color
	ForegroundDim   lipgloss.Color // Dimmed text for less emphasis
	Muted           lipgloss.Color // Disabled/muted text
	Background      lipgloss.Color // Panel backgrounds
	Border          lipgloss.Color // Border color
	FocusedBackground lipgloss.Color // Focused pane background
	UnfocusedBackground lipgloss.Color // Unfocused pane background
	FocusBorder     lipgloss.Color // Border color when focused
	HoverBackground lipgloss.Color // Background on hover/focus
}

// NewDarkTheme creates a theme suitable for dark terminals in pure textmode.
// All colors use the 16-color ANSI palette (0-15) for maximum compatibility.
func NewDarkTheme() Theme {
	return Theme{
		Primary:         lipgloss.Color("4"),  // Blue
		PrimaryDim:      lipgloss.Color("12"), // Bright blue
		Secondary:       lipgloss.Color("5"),  // Magenta
		SecondaryDim:    lipgloss.Color("13"), // Bright magenta
		Success:         lipgloss.Color("2"),  // Green
		SuccessDim:      lipgloss.Color("10"), // Bright green
		Warning:         lipgloss.Color("3"),  // Yellow
		WarningDim:      lipgloss.Color("11"), // Bright yellow
		Error:           lipgloss.Color("1"),  // Red
		ErrorDim:        lipgloss.Color("9"),  // Bright red
		Foreground:      lipgloss.Color("7"),  // Light gray
		ForegroundDim:   lipgloss.Color("8"),  // Dark gray
		Muted:           lipgloss.Color("8"),  // Dark gray
		Background:      lipgloss.Color("0"),  // Black
		Border:          lipgloss.Color("8"),  // Dark gray
		FocusedBackground: lipgloss.Color("0"), // Black
		UnfocusedBackground: lipgloss.Color("0"), // Black
		FocusBorder:     lipgloss.Color("4"),  // Blue (same as Primary)
		HoverBackground: lipgloss.Color("0"),  // Black
	}
}

// NewLightTheme creates a theme suitable for light terminals in pure textmode.
// All colors use the 16-color ANSI palette (0-15) for maximum compatibility.
func NewLightTheme() Theme {
	return Theme{
		Primary:         lipgloss.Color("4"),  // Blue
		PrimaryDim:      lipgloss.Color("12"), // Bright blue
		Secondary:       lipgloss.Color("5"),  // Magenta
		SecondaryDim:    lipgloss.Color("13"), // Bright magenta
		Success:         lipgloss.Color("2"),  // Green
		SuccessDim:      lipgloss.Color("10"), // Bright green
		Warning:         lipgloss.Color("3"),  // Yellow
		WarningDim:      lipgloss.Color("11"), // Bright yellow
		Error:           lipgloss.Color("1"),  // Red
		ErrorDim:        lipgloss.Color("9"),  // Bright red
		Foreground:      lipgloss.Color("0"),  // Black
		ForegroundDim:   lipgloss.Color("8"),  // Dark gray
		Muted:           lipgloss.Color("8"),  // Dark gray
		Background:      lipgloss.Color("15"), // White
		Border:          lipgloss.Color("8"),  // Dark gray
		FocusedBackground: lipgloss.Color("15"), // White
		UnfocusedBackground: lipgloss.Color("7"), // Light gray
		FocusBorder:     lipgloss.Color("4"),  // Blue (same as Primary)
		HoverBackground: lipgloss.Color("15"), // White
	}
}

// DefaultTheme returns the default theme (dark).
var DefaultTheme = NewDarkTheme()