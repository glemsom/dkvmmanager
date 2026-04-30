package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	theme "github.com/glemsom/dkvmmanager/internal/tui/theme"
)

// Colors defines the Tokyo Night Storm color palette
var Colors = theme.DefaultTheme

// StatusColors defines colors for VM status indicators
var StatusColors = struct {
	Running lipgloss.Color
	Stopped lipgloss.Color
	Error   lipgloss.Color
}{
	Running: theme.DefaultTheme.Success,
	Stopped: theme.DefaultTheme.Muted,
	Error:   theme.DefaultTheme.Error,
}

// LayeredPanelStyle returns a styled panel with normal border effect
func LayeredPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.Border).
		BorderBackground(Colors.Background).
		Background(Colors.Background).
		Padding(1, 2)
}

// ActiveLayeredPanelStyle returns a styled panel with active normal border
func ActiveLayeredPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.Primary).   // Outer border = accent
		BorderBackground(Colors.Border).    // Inner = border color
		Background(Colors.FocusedBackground).
		Padding(1, 2)
}

// PanelWithTitleStyle returns a styled panel with a title
func PanelWithTitleStyle(title string) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.Primary).
		Background(Colors.Background).
		Padding(1, 2).
		Bold(true).
		SetString(title)
}

// NormalTextStyle returns the default text style
func NormalTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Foreground)
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
		Foreground(Colors.ForegroundDim).
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
		Background(Colors.HoverBackground).
		Bold(true).
		PaddingLeft(1)
}

// ListItemNormalStyle returns the style for normal list items
func ListItemNormalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.ForegroundDim).
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
		Foreground(Colors.ForegroundDim)
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

// ScreenBackgroundStyle returns a style for the full-screen background
func ScreenBackgroundStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(Colors.Background)
}

// PadToScreen pads content to fill the full terminal height with background.
// Each line is padded to the given width, and empty lines are appended
// until the content reaches the given height.
func PadToScreen(content string, width, height int) string {
	lines := strings.Split(content, "\n")

	// Pad each line to full width (background fills behind spaces)
	for i, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth < width {
			lines[i] = line + strings.Repeat(" ", width-lineWidth)
		}
	}

	// Pad to full height with blank lines
	if len(lines) < height {
		blankLine := strings.Repeat(" ", width)
		for len(lines) < height {
			lines = append(lines, blankLine)
		}
	}

	return ScreenBackgroundStyle().Render(strings.Join(lines, "\n"))
}
