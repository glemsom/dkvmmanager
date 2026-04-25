package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// StatusBar represents a status bar component with mode, message, and statistics
type StatusBar struct {
	mode    string // "Ready", "Editing", "Loading"
	message string // Status message
	vmCount int    // Number of VMs
	running int    // Running VMs
	help    string // Quick help text
	focused bool   // Whether the status bar is focused
}

// NewStatusBar creates a new StatusBar instance with default values
func NewStatusBar() *StatusBar {
	return &StatusBar{
		mode:    "Ready",
		message: "",
		vmCount: 0,
		running: 0,
		help:    "",
	}
}

// SetMode updates the mode indicator
func (s *StatusBar) SetMode(mode string) {
	s.mode = mode
}

// SetFocused sets whether the status bar is focused
func (s *StatusBar) SetFocused(focused bool) {
	s.focused = focused
}

// SetMessage updates the status message
func (s *StatusBar) SetMessage(message string) {
	s.message = message
}

// SetStats updates the VM statistics
func (s *StatusBar) SetStats(vmCount, running int) {
	s.vmCount = vmCount
	s.running = running
}

// SetHelp updates the help text
func (s *StatusBar) SetHelp(help string) {
	s.help = help
}

// Render renders the status bar with the given width
func (s *StatusBar) Render(width int) string {
	// Left section: mode indicator with icon and optional spinner
	left := s.renderModeIndicator()

	// Center section: message
	center := s.message

	// Right section: stats and help
	right := s.renderRightSection()

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	centerWidth := lipgloss.Width(center)
	rightWidth := lipgloss.Width(right)

	// Calculate gaps for centering
	gap := width - leftWidth - centerWidth - rightWidth
	if gap < 0 {
		gap = 0
	}
	leftGap := gap / 2
	rightGap := gap - leftGap

	// Style the center and right parts based on focus state
	var centerStyle lipgloss.Style
	var rightStyle lipgloss.Style
	if s.focused {
		centerStyle = lipgloss.NewStyle().Foreground(styles.Colors.Foreground)
		rightStyle = lipgloss.NewStyle().Foreground(styles.Colors.ForegroundDim)
	} else {
		centerStyle = lipgloss.NewStyle().Foreground(styles.Colors.ForegroundDim)
		rightStyle = lipgloss.NewStyle().Foreground(styles.Colors.ForegroundDim)
	}
	centerPart := centerStyle.Render(center)
	rightPart := rightStyle.Render(right)

	// Build the final rendered string
	content := left + strings.Repeat(" ", leftGap) + centerPart + strings.Repeat(" ", rightGap) + rightPart

	// Apply background styling to the entire bar
	style := lipgloss.NewStyle().
		Background(styles.Colors.Background).
		Width(width)

	return style.Render(content)
}

// renderModeIndicator renders the mode indicator with Unicode icon and optional spinner
func (s *StatusBar) renderModeIndicator() string {
	var color lipgloss.Color

	switch s.mode {
	case "Ready":
		color = styles.StatusColors.Running // Green
	case "Editing":
		color = styles.Colors.Warning // Yellow
	case "Loading":
		color = styles.Colors.Primary // Cyan
	default:
		color = styles.Colors.Muted // Gray
	}

	icon := styles.ModeIcons[s.mode]
	if icon == "" {
		icon = "◌"
	}

	iconRendered := lipgloss.NewStyle().
		Foreground(color).
		Bold(true).
		Render(icon)

	return fmt.Sprintf("%s %s", iconRendered, s.mode)
}

// renderRightSection renders the right section with stats and help
func (s *StatusBar) renderRightSection() string {
	parts := []string{}

	// Add VM status: Running (green ▶) or Stopped (red ■)
	if s.vmCount > 0 {
		var status string
		if s.running > 0 {
			runningIcon := lipgloss.NewStyle().
				Foreground(styles.StatusColors.Running).
				Render("▶")
			status = fmt.Sprintf("%s Running", runningIcon)
		} else {
			stoppedIcon := lipgloss.NewStyle().
				Foreground(styles.StatusColors.Stopped).
				Render("■")
			status = fmt.Sprintf("%s Stopped", stoppedIcon)
		}
		parts = append(parts, status)
	}

	// Add help text if available
	if s.help != "" {
		parts = append(parts, s.help)
	}

	// Join parts with separator
	if len(parts) > 0 {
		return strings.Join(parts, " | ")
	}

	return ""
}

// GetMode returns the current mode
func (s *StatusBar) GetMode() string {
	return s.mode
}

// GetMessage returns the current message
func (s *StatusBar) GetMessage() string {
	return s.message
}

// GetStats returns the current VM count and running count
func (s *StatusBar) GetStats() (vmCount, running int) {
	return s.vmCount, s.running
}

// GetHelp returns the current help text
func (s *StatusBar) GetHelp() string {
	return s.help
}
