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

// NewDarkTheme creates a theme suitable for dark terminals.
func NewDarkTheme() Theme {
	return Theme{
		Primary:         lipgloss.Color("#7aa2f7"), // Blue (Storm)
		PrimaryDim:      lipgloss.Color("#292e42"), // Blue-tinted dark
		Secondary:       lipgloss.Color("#bb9af7"), // Purple
		SecondaryDim:    lipgloss.Color("#2a2e3f"), // Purple-tinted dark
		Success:         lipgloss.Color("#73daca"), // Teal green
		SuccessDim:      lipgloss.Color("#1a1f2e"), // Subtle green tint
		Warning:         lipgloss.Color("#e0af68"), // Orange/yellow
		WarningDim:      lipgloss.Color("#2f1f1a"), // Subtle yellow tint
		Error:           lipgloss.Color("#f7768e"), // Red pink
		ErrorDim:        lipgloss.Color("#2f1a1f"), // Subtle red tint
		Foreground:      lipgloss.Color("#a9b1d6"), // Foreground (light gray-blue)
		ForegroundDim:   lipgloss.Color("#828baa"), // Dimmed text (blue-gray, WCAG AA: 5.06:1 on bg)
		Muted:           lipgloss.Color("#828baa"), // Muted blue-gray (WCAG AA: 5.06:1 on bg)
		Background:      lipgloss.Color("#1a1b26"), // Deep navy
		Border:          lipgloss.Color("#626e98"), // Muted blue (WCAG AA 1.4.11: 3.42:1 on bg)
		FocusedBackground: lipgloss.Color("#1f2335"), // Lighter navy
		UnfocusedBackground: lipgloss.Color("#16161e"), // Darker navy
		FocusBorder:     lipgloss.Color("#7aa2f7"), // Same as Primary
		HoverBackground: lipgloss.Color("#1f2335"), // Same as FocusedBackground
	}
}

// NewLightTheme creates a theme suitable for light terminals (if needed).
// All color combinations are verified to meet WCAG AA 4.5:1 contrast ratio.
func NewLightTheme() Theme {
	return Theme{
		Primary:         lipgloss.Color("#2f6dde"), // Blue
		PrimaryDim:      lipgloss.Color("#f6f8fc"), // Very light blue tint (WCAG AA section headers)
		Secondary:       lipgloss.Color("#7c3aed"), // Purple (5.70:1 on white)
		SecondaryDim:    lipgloss.Color("#ede9fe"), // Light purple tint
		Success:         lipgloss.Color("#15803d"), // Green (5.02:1 on white)
		SuccessDim:      lipgloss.Color("#dcfce7"), // Light green tint (4.57:1 with success text)
		Warning:         lipgloss.Color("#a16207"), // Amber (4.92:1 on white)
		WarningDim:      lipgloss.Color("#fef9c3"), // Light amber tint (4.58:1 with warning text)
		Error:           lipgloss.Color("#c53030"), // Red
		ErrorDim:        lipgloss.Color("#fef2f2"), // Light red tint (5.00:1 with error text)
		Foreground:      lipgloss.Color("#212529"), // Dark gray for text
		ForegroundDim:   lipgloss.Color("#5f6878"), // Gray (5.62:1 on white, WCAG AA on all bg variants)
		Muted:           lipgloss.Color("#5f6878"), // Gray (5.62:1 on white, WCAG AA on all bg variants)
		Background:      lipgloss.Color("#ffffff"), // White
		Border:          lipgloss.Color("#6b7280"), // Gray border (4.83:1 on white, WCAG 1.4.11)
		FocusedBackground: lipgloss.Color("#f8f9fa"), // Light gray
		UnfocusedBackground: lipgloss.Color("#e9ecef"), // Slightly lighter gray
		FocusBorder:     lipgloss.Color("#2f6dde"), // Same as Primary
		HoverBackground: lipgloss.Color("#f8f9fa"), // Same as FocusedBackground (WCAG AA for text on hover)
	}
}

// DefaultTheme returns the default theme (dark).
var DefaultTheme = NewDarkTheme()