// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// --- Rendering ---

var (
	startStopScriptLabelStyle        = lipgloss.NewStyle().Foreground(styles.Colors.ForegroundDim)
	startStopScriptFocusStyle        = styles.FormFocusStyle()
	startStopScriptInputStyle        = styles.FormInputStyle()
	startStopScriptErrorStyle        = styles.ErrorTextStyle()
	startStopScriptMutedStyle        = styles.FormMutedStyle()
	startStopScriptSaveStyle         = styles.FormSaveStyle()
	startStopScriptSectionStyle      = styles.DetailSectionStyle()
	startStopScriptToggleOn          = lipgloss.NewStyle().Foreground(styles.Colors.Success).Bold(true)
	startStopScriptToggleOff         = lipgloss.NewStyle().Foreground(styles.Colors.Error).Bold(true)
	startStopScriptBrowseStyle       = lipgloss.NewStyle().Foreground(styles.Colors.ForegroundDim).Bold(true)
	startStopScriptBrowseFocusStyle  = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Bold(true)
)

// --- FormModel Render Helpers ---

// renderTogglePosition renders the mode toggle for RenderPosition.
func (m *StartStopScriptFormModel) renderTogglePosition(focused bool) string {
	prefix := "  "
	if focused {
		prefix = startStopScriptFocusStyle.Render("> ")
	}

	var toggleStr string
	if m.config.UseBuiltin {
		toggleStr = prefix + startStopScriptToggleOn.Render("[Builtin]") + " " + startStopScriptMutedStyle.Render("|") + " " + startStopScriptLabelStyle.Render("[Custom]")
	} else {
		toggleStr = prefix + startStopScriptLabelStyle.Render("[Builtin]") + " " + startStopScriptMutedStyle.Render("|") + " " + startStopScriptToggleOff.Render("[Custom]")
	}

	return toggleStr
}

// renderTextPosition renders a text input field for RenderPosition.
func (m *StartStopScriptFormModel) renderTextPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
	val := m.getScriptPath(pos.Key)

	var row string
	if focused {
		row = startStopScriptFocusStyle.Render(pos.Label + ": ")
	} else {
		row = startStopScriptLabelStyle.Render(pos.Label + ": ")
	}

	pathWidth := 30
	paddedPath := padRight(val, pathWidth)

	// Insert cursor indicator for focused field
	if focused && cursorOffset >= 0 && cursorOffset <= len(val) {
		display := val
		if cursorOffset < len(display) {
			display = display[:cursorOffset] + "▌" + display[cursorOffset:]
		} else {
			display = display + "▌"
		}
		row += startStopScriptInputStyle.Render(padRight(display, pathWidth))
	} else {
		row += startStopScriptInputStyle.Render(paddedPath)
	}

	return row
}

// renderButtonPosition renders a button for RenderPosition.
func (m *StartStopScriptFormModel) renderButtonPosition(pos form.FocusPos, focused bool) string {
	if pos.Key == "start_browse" || pos.Key == "stop_browse" {
		if focused {
			return "  " + startStopScriptBrowseFocusStyle.Render("[Browse →]")
		}
		return "  " + startStopScriptBrowseStyle.Render("[Browse →]")
	}

	// Save/Cancel buttons: render as footer-style help
	if pos.Key == "save" {
		if focused {
			return startStopScriptSaveStyle.Render("[Space/Enter] Save") + "    " + startStopScriptMutedStyle.Render("[ESC] Cancel")
		}
		return startStopScriptMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
	}

	if pos.Key == "cancel" {
		if focused {
			return startStopScriptMutedStyle.Render("[Space/Enter] Save") + "    " + startStopScriptLabelStyle.Render("[ESC] Cancel")
		}
		return startStopScriptMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
	}

	return ""
}

// --- Backward-compatible rendering ---

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

// renderToggle renders the mode toggle (backward compat)
func (m *StartStopScriptFormModel) renderToggle(lines []string) []string {
	focused := (m.focusIndex == 0)

	prefix := "  "
	if focused {
		prefix = startStopScriptFocusStyle.Render("> ")
	}

	var toggleStr string
	if m.config.UseBuiltin {
		toggleStr = prefix + startStopScriptToggleOn.Render("[Builtin]") + " " + startStopScriptMutedStyle.Render("|") + " " + startStopScriptLabelStyle.Render("[Custom]")
	} else {
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
	// Find start and stop path/browse positions by searching through all positions
	var startPathIdx, stopPathIdx, startBrowseIdx, stopBrowseIdx int = -1, -1, -1, -1
	for i, pos := range m.positions {
		switch pos.Key {
		case "start_path":
			startPathIdx = i
		case "start_browse":
			startBrowseIdx = i
		case "stop_path":
			stopPathIdx = i
		case "stop_browse":
			stopBrowseIdx = i
		}
	}

	// Start script row
	if startPathIdx >= 0 {
		focusedPath := (m.focusIndex == startPathIdx)
		focusedBrowse := (m.focusIndex == startBrowseIdx)

		val := m.config.StartScript
		if val == "" {
			val = "/media/dkvmdata/start.sh"
		}

		var row string

		if focusedPath || focusedBrowse {
			row = startStopScriptFocusStyle.Render("Start Script: ")
		} else {
			row = startStopScriptLabelStyle.Render("Start Script: ")
		}

		pathWidth := 30
		paddedPath := padRight(val, pathWidth)
		row += startStopScriptInputStyle.Render(paddedPath)

		row += "  "
		if focusedBrowse {
			row += startStopScriptBrowseFocusStyle.Render("[Browse →]")
		} else {
			row += startStopScriptBrowseStyle.Render("[Browse →]")
		}

		lines = append(lines, row)
	}

	// Stop script row
	if stopPathIdx >= 0 {
		focusedPath := (m.focusIndex == stopPathIdx)
		focusedBrowse := (m.focusIndex == stopBrowseIdx)

		val := m.config.StopScript
		if val == "" {
			val = "/media/dkvmdata/stop.sh"
		}

		var row string

		if focusedPath || focusedBrowse {
			row = startStopScriptFocusStyle.Render("Stop Script: ")
		} else {
			row = startStopScriptLabelStyle.Render("Stop Script: ")
		}

		pathWidth := 30
		paddedPath := padRight(val, pathWidth)
		row += startStopScriptInputStyle.Render(paddedPath)

		row += "  "
		if focusedBrowse {
			row += startStopScriptBrowseFocusStyle.Render("[Browse →]")
		} else {
			row += startStopScriptBrowseStyle.Render("[Browse →]")
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
		if pos.Key == "save" {
			saveIdx = i
		} else if pos.Key == "cancel" {
			cancelIdx = i
		}
	}

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
