// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// View implements tea.Model
func (m *StartStopScriptFormModel) View() string {
	// If file browser is active, show it instead of the form
	if m.fileBrowser != nil && m.fileBrowser.active {
		return m.fileBrowser.View()
	}

	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// SetSize updates the form dimensions
func (m *StartStopScriptFormModel) SetSize(w, h int) {
	m.contentW = w
	m.contentH = h
	if !m.ready {
		m.vp = viewport.New(w, h)
		m.ready = true
	} else {
		m.vp.Width = w
		m.vp.Height = h
	}
	m.syncViewport()
}

// --- Rendering ---

var (
	startStopScriptLabelStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	startStopScriptFocusStyle         = styles.FormFocusStyle()
	startStopScriptInputStyle         = styles.FormInputStyle()
	startStopScriptErrorStyle         = styles.ErrorTextStyle()
	startStopScriptMutedStyle         = styles.FormMutedStyle()
	startStopScriptSaveStyle          = styles.FormSaveStyle()
	startStopScriptSectionStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Bold(true)
	startStopScriptToggleOn           = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	startStopScriptToggleOff          = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	// New: Side-by-side layout styles
	startStopScriptBrowseStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Bold(true)
	startStopScriptBrowseFocusStyle  = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Bold(true)
)

// renderAllLines produces the full list of output lines for the form
func (m *StartStopScriptFormModel) renderAllLines() []string {
	var lines []string

	// Header
	lines = append(lines, startStopScriptSectionStyle.Render("Custom Start/Stop Script"))

	// Toggle
	lines = m.renderToggle(lines)

	lines = append(lines, "")

	// Always show PCI devices for reference
	lines = m.renderBuiltinDevices(lines)
	lines = append(lines, "")

	// Script paths only in custom mode
	if !m.config.UseBuiltin {
		lines = m.renderCustomPaths(lines)
	}

	// Buttons
	lines = append(lines, "")
	lines = m.renderButtons(lines)

	return lines
}

// renderToggle renders the mode toggle
func (m *StartStopScriptFormModel) renderToggle(lines []string) []string {
	focused := (m.focusIndex == 0)

	prefix := "  "
	if focused {
		prefix = startStopScriptFocusStyle.Render("> ")
	}

	var toggleStr string
	if m.config.UseBuiltin {
		// Builtin is active: highlight green
		toggleStr = prefix + startStopScriptToggleOn.Render("[Builtin]") + " " + startStopScriptMutedStyle.Render("|") + " " + startStopScriptLabelStyle.Render("[Custom]")
	} else {
		// Custom is active: highlight red
		toggleStr = prefix + startStopScriptLabelStyle.Render("[Builtin]") + " " + startStopScriptMutedStyle.Render("|") + " " + startStopScriptToggleOff.Render("[Custom]")
	}

	lines = append(lines, toggleStr)

	return lines
}

// renderBuiltinDevices renders the PCI device list when using builtin script
func (m *StartStopScriptFormModel) renderBuiltinDevices(lines []string) []string {
	lines = append(lines, startStopScriptLabelStyle.Render("PCI Devices:"))

	if len(m.pciConfig.Devices) == 0 {
		lines = append(lines, startStopScriptMutedStyle.Render("  (none configured)"))
	} else {
		for _, dev := range m.pciConfig.Devices {
			lines = append(lines, startStopScriptLabelStyle.Render("  ")+dev.Address+" "+dev.Name)
		}
	}

	lines = append(lines, startStopScriptMutedStyle.Render("  (auto-initializes vfio-pci on VM start)"))

	return lines
}

// renderCustomPaths renders the script path fields when using custom scripts
func (m *StartStopScriptFormModel) renderCustomPaths(lines []string) []string {
	// Find start and stop path positions by searching through all positions
	var startPathIdx, stopPathIdx int = -1, -1
	for i, pos := range m.positions {
		if pos.fieldName == "start_path" {
			startPathIdx = i
		} else if pos.fieldName == "stop_path" {
			stopPathIdx = i
		}
	}

	// Start script row
	if startPathIdx >= 0 {
		focusedPath := (m.focusIndex == startPathIdx)
		focusedBrowse := (m.focusIndex == startPathIdx+1) // browse is next position

		val := m.config.StartScript
		if val == "" {
			val = "/media/dkvmdata/start.sh"
		}

		// Build row: "Start Script:  <path>    [Browse →]"
		var row string

		// Label
		if focusedPath || focusedBrowse {
			row = startStopScriptFocusStyle.Render("Start Script: ")
		} else {
			row = startStopScriptLabelStyle.Render("Start Script: ")
		}

		// Path (padded to fill space before browse button)
		pathWidth := 30 // configurable width
		paddedPath := padRight(val, pathWidth)
		if focusedPath {
			row += startStopScriptInputStyle.Render(paddedPath)
		} else {
			row += startStopScriptInputStyle.Render(paddedPath)
		}

		// Gap and Browse button
		row += "  "
		if focusedBrowse {
			row += startStopScriptBrowseFocusStyle.Render("[Browse →]")
		} else {
			row += startStopScriptBrowseStyle.Render("[Browse]")
		}

		lines = append(lines, row)
	}

	// Stop script row
	if stopPathIdx >= 0 {
		focusedPath := (m.focusIndex == stopPathIdx)
		focusedBrowse := (m.focusIndex == stopPathIdx+1) // browse is next position

		val := m.config.StopScript
		if val == "" {
			val = "/media/dkvmdata/stop.sh"
		}

		// Build row: "Stop Script:   <path>    [Browse →]"
		var row string

		// Label
		if focusedPath || focusedBrowse {
			row = startStopScriptFocusStyle.Render("Stop Script: ")
		} else {
			row = startStopScriptLabelStyle.Render("Stop Script: ")
		}

		// Path (padded to fill space before browse button)
		pathWidth := 30
		paddedPath := padRight(val, pathWidth)
		if focusedPath {
			row += startStopScriptInputStyle.Render(paddedPath)
		} else {
			row += startStopScriptInputStyle.Render(paddedPath)
		}

		// Gap and Browse button
		row += "  "
		if focusedBrowse {
			row += startStopScriptBrowseFocusStyle.Render("[Browse →]")
		} else {
			row += startStopScriptBrowseStyle.Render("[Browse]")
		}

		lines = append(lines, row)
	}

	// Help text showing the CLI signature
	lines = append(lines, "")
	lines = append(lines, startStopScriptMutedStyle.Render("Custom script receives:"))
	lines = append(lines, startStopScriptMutedStyle.Render("  <start|stop> [device1] [device2] ..."))

	return lines
}

// padRight pads a string with spaces to reach target width
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// renderButtons renders the Save and Cancel buttons
func (m *StartStopScriptFormModel) renderButtons(lines []string) []string {
	// Find save and cancel positions
	var saveIdx, cancelIdx int
	for i, pos := range m.positions {
		if pos.fieldName == "save" {
			saveIdx = i
		} else if pos.fieldName == "cancel" {
			cancelIdx = i
		}
	}

	// Save button - match pattern from other views: "[Space/Enter] Save    [ESC] Cancel"
	focused := (m.focusIndex == saveIdx || m.focusIndex == cancelIdx)
	if focused && m.focusIndex == saveIdx {
		lines = append(lines, startStopScriptSaveStyle.Render("[Space/Enter] Save")+"    "+startStopScriptMutedStyle.Render("[ESC] Cancel"))
	} else if focused && m.focusIndex == cancelIdx {
		lines = append(lines, startStopScriptMutedStyle.Render("[Space/Enter] Save")+"    "+startStopScriptLabelStyle.Render("[ESC] Cancel"))
	} else {
		lines = append(lines, startStopScriptMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel"))
	}

	return lines
}

// syncViewport updates the viewport content
func (m *StartStopScriptFormModel) syncViewport() {
	lines := m.renderAllLines()
	m.renderedLines = lines

	content := ""
	for i, line := range lines {
		if i > 0 {
			content += "\n"
		}
		content += line
	}

	m.vp.SetContent(content)
}
