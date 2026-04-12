package styles

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HighlightStyle returns the style for highlighted text
func HighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true)
}

// AccentStyle returns the style for accent text
func AccentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary)
}

// InverseStyle returns the style for inverse text (background and foreground swapped)
func InverseStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(Colors.Primary).
		Foreground(Colors.Background)
}

// DimStyle returns the style for dimmed text
func DimStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted).
		Faint(true)
}

// ItalicStyle returns the style for italic text
func ItalicStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Italic(true)
}

// UnderlineStyle returns the style for underlined text
func UnderlineStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Underline(true)
}

// StrikethroughStyle returns the style for strikethrough text
func StrikethroughStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted).
		Strikethrough(true)
}

// BadgeStyle returns the style for badges/labels
func BadgeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Background).
		Background(Colors.Primary).
		Bold(true).
		Padding(0, 1)
}

// TagStyle returns the style for tags
func TagStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.Primary).
		Padding(0, 1)
}

// SeparatorStyle returns the style for separators
func SeparatorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted)
}

// ProgressBarStyle returns the style for progress bar background
func ProgressBarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(Colors.Muted)
}

// ProgressBarFillStyle returns the style for progress bar fill
func ProgressBarFillStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(Colors.Primary)
}

// InputStyle returns the style for input fields
func InputStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.Border).
		Padding(0, 1)
}

// InputFocusedStyle returns the style for focused input fields
func InputFocusedStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.Primary).
		Padding(0, 1)
}

// InputErrorStyle returns the style for input fields with errors
func InputErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.Error).
		Padding(0, 1)
}

// TableHeaderStyle returns the style for table headers
func TableHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		Padding(0, 1)
}

// TableRowStyle returns the style for table rows
func TableRowStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Padding(0, 1)
}

// TableSelectedRowStyle returns the style for selected table rows
func TableSelectedRowStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		Background(Colors.Background).
		Padding(0, 1)
}

// TooltipStyle returns the style for tooltips
func TooltipStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(Colors.Background).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Border).
		Padding(1, 2)
}

// ModalStyle returns the style for modal dialogs
func ModalStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).
		Background(Colors.Background).
		Padding(1, 2)
}

// NotificationStyle returns the style for notifications
func NotificationStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Primary).
		Background(Colors.Background).
		Padding(1, 2)
}

// ErrorNotificationStyle returns the style for error notifications
func ErrorNotificationStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Error).
		Background(Colors.Background).
		Padding(1, 2)
}

// SuccessNotificationStyle returns the style for success notifications
func SuccessNotificationStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Success).
		Background(Colors.Background).
		Padding(1, 2)
}

// WarningNotificationStyle returns the style for warning notifications
func WarningNotificationStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Colors.Warning).
		Background(Colors.Background).
		Padding(1, 2)
}

// DetailPanelStyle returns the style for the VM detail panel
func DetailPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Colors.Muted).
		Padding(0, 1)
}

// DetailHeaderStyle returns the style for VM name in detail view
func DetailHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true)
}

// DetailSectionStyle returns the style for section headers in detail view
func DetailSectionStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true).
		Padding(1, 0, 0, 0)
}

// DetailValueStyle returns the style for property values in detail view
func DetailValueStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))
}

// DetailLabelStyle returns the style for property labels in detail view
func DetailLabelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted)
}

// ScrollbarStyle returns the style for scrollbar
func ScrollbarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted)
}

// ScrollbarThumbStyle returns the style for scrollbar thumb
func ScrollbarThumbStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary)
}

// LinkStyle returns the style for links
func LinkStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Underline(true)
}

// CodeStyle returns the style for code snippets
func CodeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Secondary).
		Background(Colors.Background).
		Padding(0, 1)
}

// QuoteStyle returns the style for quoted text
func QuoteStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted).
		Italic(true).
		PaddingLeft(1).
		BorderLeft(true).
		BorderForeground(Colors.Muted)
}

// HeadingStyle returns the style for headings
func HeadingStyle(level int) lipgloss.Style {
	style := lipgloss.NewStyle().
		Bold(true)

	switch level {
	case 1:
		return style.
			Foreground(Colors.Primary).
			MarginBottom(1)
	case 2:
		return style.
			Foreground(Colors.Secondary).
			MarginBottom(1)
	case 3:
		return style.
			Foreground(lipgloss.Color("252")).
			MarginBottom(1)
	default:
		return style.
			Foreground(Colors.Primary).
			MarginBottom(1)
	}
}

// CaptionStyle returns the style for captions
func CaptionStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted).
		Italic(true)
}

// TimestampStyle returns the style for timestamps
func TimestampStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted)
}

// CounterStyle returns the style for counters
func CounterStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Primary).
		Bold(true)
}

// DividerStyle returns the style for dividers
func DividerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Colors.Muted)
}

// RenderDivider renders a horizontal divider
func RenderDivider(width int) string {
	if width <= 0 {
		return ""
	}
	return DividerStyle().Render(lipgloss.NewStyle().Width(width).Render("─"))
}

// RenderVerticalDivider renders a vertical divider
func RenderVerticalDivider(height int) string {
	if height <= 0 {
		return ""
	}
	divider := ""
	for i := 0; i < height; i++ {
		divider += "│\n"
	}
	return DividerStyle().Render(divider)
}

// RenderGradientText renders text with a per-character color gradient
// interpolating between startColor and endColor (ANSI 256 indices).
func RenderGradientText(text string, startColor, endColor int) string {
	if len(text) == 0 {
		return ""
	}
	var b strings.Builder
	for i, ch := range text {
		t := 0.0
		if len(text) > 1 {
			t = float64(i) / float64(len(text)-1)
		}
		colorIdx := int(math.Round(float64(startColor) + t*float64(endColor-startColor)))
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(fmt.Sprintf("%d", colorIdx)))
		b.WriteString(style.Render(string(ch)))
	}
	return b.String()
}

// RenderProgressBar renders a horizontal progress bar using █░ characters.
// fraction is 0.0-1.0, width is the bar character count (excluding label/percentage).
// label is displayed to the left. detail is optional size info shown after the bar.
func RenderProgressBar(fraction float64, width int, label, detail string) string {
	if width < 1 {
		width = 1
	}
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	filled := int(math.Round(fraction * float64(width)))
	if filled > width {
		filled = width
	}

	fillStyle := lipgloss.NewStyle().Foreground(Colors.Success)
	emptyStyle := lipgloss.NewStyle().Foreground(Colors.Muted)
	pctStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	bar := fillStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", width-filled))

	pct := pctStyle.Render(fmt.Sprintf("%.0f%%", fraction*100))

	if label != "" {
		labelStyled := lipgloss.NewStyle().Foreground(Colors.Muted).Render(label + ": ")
		if detail != "" {
			detailStyled := lipgloss.NewStyle().Foreground(Colors.Muted).Render(" (" + detail + ")")
			return labelStyled + "[" + bar + "] " + pct + detailStyled
		}
		return labelStyled + "[" + bar + "] " + pct
	}
	if detail != "" {
		detailStyled := lipgloss.NewStyle().Foreground(Colors.Muted).Render(" (" + detail + ")")
		return "[" + bar + "] " + pct + detailStyled
	}
	return "[" + bar + "] " + pct
}
