// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// RenderPosition returns the markup for a single position.
func (m *VMFormModel) RenderPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
	switch pos.Key {
	case "hardDisks_label", "cdroms_label":
		// Headers
		return styles.FormLabelStyle().Render(pos.Key)

	case "vmName":
		val := m.vmName
		cursor := m.effectiveCursor("vmName", val)
		if cursorOffset >= 0 {
			cursor = cursorOffset
		}
		return m.renderTextInput("VM Name", val, cursor, focused)

	case "macAddress":
		val := m.macAddress
		cursor := m.effectiveCursor("macAddress", val)
		if cursorOffset >= 0 {
			cursor = cursorOffset
		}
		return m.renderTextInput("MAC Address", val, cursor, focused)

	case "vncEnabled", "tpmEnabled", "networkMode":
		label := m.fieldLabel(pos.Key)
		rendered := m.renderToggle(label, pos.Key, focused)
		if errMsg, ok := m.errors[pos.Key]; ok {
			return rendered + "\n  " + styles.ErrorTextStyle().Render(errMsg)
		}
		return rendered

	case "save":
		saveText := styles.MutedTextStyle().Render("[Space/Enter] Save  [ESC] Cancel")
		if focused {
			saveText = styles.SuccessTextStyle().Render("[Space/Enter] Save") + "  " + styles.MutedTextStyle().Render("[ESC] Cancel")
		}
		return "\n" + saveText

	}

	// Handle list items (hardDisks_N, cdroms_N) and add buttons
	if strings.HasPrefix(pos.Key, "hardDisks_") || strings.HasPrefix(pos.Key, "cdroms_") {
		if strings.HasSuffix(pos.Key, "_add") {
			// Add button
			btnText := styles.MutedTextStyle().Render("[+ Add]")
			if focused {
				btnText = styles.SelectedTextStyle().Render("[+ Add]")
			}
			fieldLabel := "Disk"
			if strings.HasPrefix(pos.Key, "cdroms_") {
				fieldLabel = "CD/DVD"
			}
			return "  " + btnText + " " + styles.MutedTextStyle().Render(fieldLabel)
		}

		// List item - extract index from key like "hardDisks_0" or "cdroms_0"
		var disks *[]string
		var idx int
		if strings.HasPrefix(pos.Key, "hardDisks_") {
			disks = &m.hardDisks
			if n, err := strconv.Atoi(pos.Key[10:]); err == nil {
				idx = n
			}
		} else if strings.HasPrefix(pos.Key, "cdroms_") {
			disks = &m.cdroms
			if n, err := strconv.Atoi(pos.Key[7:]); err == nil {
				idx = n
			}
		}

		if disks != nil && idx < len(*disks) {
			val := (*disks)[idx]
			cursor := m.effectiveCursor(pos.Key, val)
			if cursorOffset >= 0 {
				cursor = cursorOffset
			}
			return m.renderListItem(idx, val, cursor, focused)
		}
	}

	return ""
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

	// Add error after the input
	result := prefix + labelPart + valPart
	if errMsg, ok := m.errors[strings.ToLower(label)]; ok {
		result += "\n  " + styles.ErrorTextStyle().Render(errMsg)
	}

	return result
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

	return prefix + labelPart + valPart
}