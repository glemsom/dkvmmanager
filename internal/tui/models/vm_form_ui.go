// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// syncViewport regenerates the rendered lines and syncs the viewport
func (m *VMFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	totalContent := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(totalContent)
	// Ensure focused line is visible
	if m.focusedLineIndex() >= 0 {
		m.vp.SetYOffset(clampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height))
	}
}

// focusedLineIndex maps focusIndex to a rendered line index
// Accounts for header (2 lines), then per-position line counts, plus error lines
func (m *VMFormModel) focusedLineIndex() int {
	line := 2 // header + blank line

	for i, p := range m.positions {
		if i == m.focusIndex {
			return line
		}
		// Count lines this position produces
		switch p.kind {
		case focusText:
			if p.fieldName == "hardDisks_label" || p.fieldName == "cdroms_label" {
				line++
			} else {
				line++ // label+input on one line (renderTextInput returns 1 line)
				key := posKey(p)
				if _, hasErr := m.errors[key]; hasErr {
					line++ // error line
				}
			}
		case focusListItem:
			line++
		case focusAddBtn:
			line++
		case focusToggle:
			line++
		case focusSaveBtn:
			line += 2 // blank separator + button
		}
	}
	return line
}

// clampOffset adjusts viewport offset so targetLine is visible.
// Uses a margin to keep focused items away from edges, ensuring multi-line
// elements (like Save button) and content below them remain visible.
func clampOffset(offset, targetLine, viewHeight int) int {
	if viewHeight <= 0 {
		return offset
	}
	if targetLine < offset {
		return targetLine
	}
	if targetLine >= offset+viewHeight {
		// Scrolled past view: position target at ~1/3 from top to leave room
		// for multi-line elements and content below (e.g., Save button + footer)
		newOffset := targetLine - viewHeight/3
		if newOffset < 0 {
			newOffset = 0
		}
		return newOffset
	}
	return offset
}

// renderAllLines produces the full list of output lines for the form
func (m *VMFormModel) renderAllLines() []string {
	var lines []string

	lineIdx := 0

	for i, pos := range m.positions {
		focused := (i == m.focusIndex)
		lines, lineIdx = m.renderPosition(lines, lineIdx, pos, focused)
	}

	// Save error at the bottom
	if errMsg, ok := m.errors["save"]; ok {
		lines = append(lines, "")
		lines = append(lines, styles.ErrorTextStyle().Render("Error: "+errMsg))
	}

	// Footer
	lines = append(lines, "")
	lines = append(lines, styles.MutedTextStyle().Render("Tab Navigate  Space/Enter Browse  ESC Cancel"))

	return lines
}

// renderPosition appends lines for one focus position
func (m *VMFormModel) renderPosition(lines []string, lineIdx int, pos focusPos, focused bool) ([]string, int) {
	key := posKey(pos)

	switch pos.kind {
	case focusText:
		if pos.fieldName == "hardDisks_label" {
			lines = append(lines, styles.FormLabelStyle().Render("Hard Disks:"))
			return lines, lineIdx + 1
		}
		if pos.fieldName == "cdroms_label" {
			lines = append(lines, styles.FormLabelStyle().Render("CD/DVD Drives (ISOs):"))
			return lines, lineIdx + 1
		}

		// Regular text field with label
		label := m.fieldLabel(pos.fieldName)
		val := m.getValue(pos)
		cursor := m.effectiveCursor(key, val)
		rendered := m.renderTextInput(label, val, cursor, focused)
		lines = append(lines, rendered)

		// Inline error
		if errMsg, ok := m.errors[key]; ok {
			lines = append(lines, "  "+styles.ErrorTextStyle().Render(errMsg))
		}
		return lines, lineIdx + 2 // label+input + potential error

	case focusListItem:
		val := m.getValue(pos)
		cursor := m.effectiveCursor(key, val)
		rendered := m.renderListItem(pos.listIndex, val, cursor, focused)
		lines = append(lines, rendered)
		return lines, lineIdx + 1

	case focusAddBtn:
		btnText := styles.MutedTextStyle().Render("[+ Add]")
		if focused {
			btnText = styles.SelectedTextStyle().Render("[+ Add]")
		}
		fieldLabel := "Disk"
		if pos.fieldName == "cdroms" {
			fieldLabel = "CD/DVD"
		}
		lines = append(lines, "  "+btnText+" "+styles.MutedTextStyle().Render(fieldLabel))
		return lines, lineIdx + 1

	case focusToggle:
		label := m.fieldLabel(pos.fieldName)
		rendered := m.renderToggle(label, pos.fieldName, focused)
		lines = append(lines, rendered)
		// Show inline error if present (e.g., TPM binary missing)
		if errMsg, ok := m.errors[pos.fieldName]; ok {
			lines = append(lines, "  "+styles.ErrorTextStyle().Render(errMsg))
			return lines, lineIdx + 2
		}
		return lines, lineIdx + 1

	case focusSaveBtn:
		lines = append(lines, "")
		saveText := styles.MutedTextStyle().Render("[Space/Enter] Save  [ESC] Cancel")
		if focused {
			saveText = styles.SuccessTextStyle().Render("[Space/Enter] Save") + "  " + styles.MutedTextStyle().Render("[ESC] Cancel")
		}
		lines = append(lines, saveText)
		return lines, lineIdx + 2
	}

	return lines, lineIdx
}

// renderTextInput renders a labeled text input with an optional cursor
func (m *VMFormModel) renderTextInput(label, value string, cursor int, focused bool) string {
	prefix := "  "
	if focused {
		prefix = styles.SelectedTextStyle().Render("> ")
	}

	labelPart := styles.FormLabelStyle().Render(label + ": ")

	// Build value with cursor indicator
	var valPart string
	if focused {
		// Show cursor as highlighted character or underscore at end
		if cursor < len(value) {
			before := value[:cursor]
			at := string(value[cursor])
			after := ""
			if cursor+1 < len(value) {
				after = value[cursor+1:]
			}
			valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(before) +
				lipgloss.NewStyle().Reverse(true).Render(at) +
				lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(after)
		} else {
			valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(value) + styles.SelectedTextStyle().Render("_")
		}
	} else {
		if value == "" {
			valPart = styles.MutedTextStyle().Render("(empty)")
		} else {
			valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(value)
		}
	}

	return prefix + labelPart + valPart
}

// renderListItem renders a single item in a list field
func (m *VMFormModel) renderListItem(index int, value string, cursor int, focused bool) string {
	numPart := styles.MutedTextStyle().Render(fmt.Sprintf("  [%d] ", index+1))

	var valPart string
	if focused {
		if cursor < len(value) {
			before := value[:cursor]
			at := string(value[cursor])
			after := ""
			if cursor+1 < len(value) {
				after = value[cursor+1:]
			}
			valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(before) +
				lipgloss.NewStyle().Reverse(true).Render(at) +
				lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(after)
		} else {
			valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(value) + styles.SelectedTextStyle().Render("_")
		}
	} else {
		if value == "" {
			valPart = styles.MutedTextStyle().Render("(enter path)")
		} else {
			valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(value)
		}
	}

	delPart := ""
	if focused {
		delPart = " " + styles.ErrorTextStyle().Render("[Del]")
	}

	return numPart + valPart + delPart
}

// fieldLabel returns the human-readable label for a field name
func (m *VMFormModel) fieldLabel(name string) string {
	switch name {
	case "vmName":
		return "VM Name"
	case "macAddress":
		return "MAC Address"
	case "vncEnabled":
		return "VNC"
	case "networkMode":
		return "Network"
	case "tpmEnabled":
		return "TPM"
	}
	return name
}

// renderToggle renders a toggle field (on/off)
func (m *VMFormModel) renderToggle(label, fieldName string, focused bool) string {
	prefix := "  "
	if focused {
		prefix = styles.SelectedTextStyle().Render("> ")
	}

	labelPart := styles.FormLabelStyle().Render(label + ": ")

	var valPart string
	switch fieldName {
	case "networkMode":
		if m.networkMode == "bridge" {
			if focused {
				valPart = styles.SelectedTextStyle().Render("[Bridge]")
			} else {
				valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render("[Bridge]")
			}
		} else {
			if focused {
				valPart = styles.SelectedTextStyle().Render("[NAT]")
			} else {
				valPart = styles.MutedTextStyle().Render("[NAT]")
			}
		}
	default:
		var on bool
		switch fieldName {
		case "vncEnabled":
			on = m.vncEnabled
		case "tpmEnabled":
			on = m.tpmEnabled
		}

		if on {
			if focused {
				valPart = styles.SelectedTextStyle().Render("[ON]")
			} else {
				valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render("[ON]")
			}
		} else {
			if focused {
				valPart = styles.SelectedTextStyle().Render("[OFF]")
			} else {
				valPart = styles.MutedTextStyle().Render("[OFF]")
			}
		}
	}

	toggleLine := prefix + labelPart + valPart

	return toggleLine
}
