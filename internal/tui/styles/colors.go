package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors defines the Tokyo Night Storm color palette
var Colors = struct {
	Primary             lipgloss.Color // Main accent (Blue)
	PrimaryDim          lipgloss.Color // Dimmed for backgrounds
	Secondary           lipgloss.Color // Secondary actions (Purple)
	SecondaryDim        lipgloss.Color // Dimmed for backgrounds
	Success             lipgloss.Color // Success states (Teal)
	SuccessDim          lipgloss.Color // Subtle background tints
	Warning             lipgloss.Color // Warning states (Orange)
	WarningDim          lipgloss.Color // Subtle background tints
	Error               lipgloss.Color // Error states (Red-pink)
	ErrorDim            lipgloss.Color // Subtle background tints
	Muted               lipgloss.Color // Disabled/muted text (Blue-gray)
	Background          lipgloss.Color // Panel backgrounds (Deep navy)
	Border              lipgloss.Color // Border color (Blue-gray)
	FocusedBackground   lipgloss.Color // Focused pane background (Lighter navy)
	UnfocusedBackground lipgloss.Color // Unfocused pane background (Darker navy)
}{
	Primary:             "7aa2f7",  // Blue (Storm)
	PrimaryDim:          "292e42",  // Blue-tinted dark
	Secondary:           "bb9af7",  // Purple
	SecondaryDim:        "2a2e3f",  // Purple-tinted dark
	Success:             "73daca",  // Teal green
	SuccessDim:          "1a1f2e",  // Subtle green tint
	Warning:            "e0af68",  // Orange/yellow
	WarningDim:         "2f1f1a",  // Subtle yellow tint
	Error:              "f7768e",  // Red pink
	ErrorDim:           "2f1a1f",  // Subtle red tint
	Muted:              "565f89",  // Muted blue-gray
	Background:          "#1a1b26", // Deep navy
	Border:              "#3b4261", // Muted blue
	FocusedBackground:   "#1f2335", // Lighter navy
	UnfocusedBackground: "#16161e", // Darker navy
}

// StatusColors defines colors for VM status indicators
var StatusColors = struct {
	Running lipgloss.Color
	Stopped lipgloss.Color
	Error   lipgloss.Color
}{
	Running: "73daca", // Teal green
	Stopped: "565f89", // Muted blue-gray
	Error:   "f7768e", // Red pink
}

// LayeredPanelStyle returns a styled panel with gradient border effect
func LayeredPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		BorderBackground(Colors.Background).
		Background(Colors.Background).
		Padding(1, 2)
}

// ActiveLayeredPanelStyle returns a styled panel with active gradient border
func ActiveLayeredPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).   // Outer border = accent
		BorderBackground(Colors.Border).    // Inner = border color
		Background(Colors.FocusedBackground).
		Padding(1, 2)
}

// PanelWithTitleStyle returns a styled panel with a title
func PanelWithTitleStyle(title string) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).
		Background(Colors.Background).
		Padding(1, 2).
		Bold(true).
		SetString(title)
}

// NormalTextStyle returns the default text style
func NormalTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")) // Light gray
}

// SelectedTextStyle returns the style for selected text
func SelectedTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true)
}

// FocusedTextStyle returns the style for focused text
func FocusedTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Underline(true)
}

// DisabledTextStyle returns the style for disabled text
func DisabledTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted).
		Strikethrough(false)
}

// PrimaryButtonStyle returns the style for primary buttons with glow effect
func PrimaryButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Background).
		Background(Colors.Primary).
		Bold(true).
		Padding(0, 2).
		MarginRight(1)
}

// SecondaryButtonStyle returns the style for secondary buttons with glow effect
func SecondaryButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Background).
		Background(Colors.Secondary).
		Bold(true).
		Padding(0, 2).
		MarginRight(1)
}

// DisabledButtonStyle returns the style for disabled buttons
func DisabledButtonStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted).
		Background(Colors.Background).
		Bold(false).
		Padding(0, 2).
		MarginRight(1)
}

// StatusIndicator returns a styled status indicator (bullet)
func StatusIndicator(status string) string {
	var color lipgloss.Color
	var symbol string

	switch status {
	case "running":
		color = StatusColors.Running
		symbol = "●"
	case "stopped":
		color = StatusColors.Stopped
		symbol = "○"
	case "error":
		color = StatusColors.Error
		symbol = "●"
	default:
		color = Colors.Muted
		symbol = "○"
	}

	return lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(symbol)
}

// ModeIcons maps status modes to Unicode icons
var ModeIcons = map[string]string{
	"Ready":   "◌",
	"Editing": "⚙",
	"Loading": "◌",
	"Running": "▶",
	"Stopped": "■",
	"Error":   "⚠",
}

// RunningStatusStyle returns the style for running status
func RunningStatusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(StatusColors.Running).
		Background(Colors.SuccessDim).
		Bold(true)
}

// StoppedStatusStyle returns the style for stopped status
func StoppedStatusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(StatusColors.Stopped).
		Background(Colors.Background).
		Bold(true)
}

// ErrorStatusStyle returns the style for error status
func ErrorStatusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(StatusColors.Error).
		Background(Colors.ErrorDim).
		Bold(true)
}

// TitleStyle returns the style for titles
func TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		MarginBottom(1)
}

// SubtitleStyle returns the style for subtitles
func SubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Secondary).
		Bold(true)
}

// ErrorTextStyle returns the style for error messages
func ErrorTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Error).
		Bold(true)
}

// WarningTextStyle returns the style for warning messages
func WarningTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Warning).
		Bold(true)
}

// SuccessTextStyle returns the style for success messages
func SuccessTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Success).
		Bold(true)
}

// MutedTextStyle returns the style for muted text
func MutedTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted)
}

// HeaderStyle returns the style for headers
func HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		MarginBottom(1)
}

// FooterStyle returns the style for footers
func FooterStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted).
		MarginTop(1)
}

// BorderStyle returns the style for borders
func BorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border)
}

// ActiveBorderStyle returns the style for active/focused borders
func ActiveBorderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary)
}

// ListItemSelectedStyle returns the style for selected list items
func ListItemSelectedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		PaddingLeft(1)
}

// ListItemNormalStyle returns the style for normal list items
func ListItemNormalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		PaddingLeft(1)
}

// ListItemDisabledStyle returns the style for disabled list items
func ListItemDisabledStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted).
		PaddingLeft(1)
}

// HelpKeyStyle returns the style for help keys
func HelpKeyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true)
}

// HelpDescStyle returns the style for help descriptions
func HelpDescStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
}

// HelpSeparatorStyle returns the style for help separators
func HelpSeparatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted)
}

// FormFocusStyle returns the style for focused form elements
func FormFocusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true)
}

// FormSaveStyle returns the style for form save buttons
func FormSaveStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Success).
		Bold(true)
}

// FormLabelStyle returns the style for form labels
func FormLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted)
}

// FormInputStyle returns the style for form inputs
func FormInputStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary)
}

// FormMutedStyle returns the style for muted form text
func FormMutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted)
}
