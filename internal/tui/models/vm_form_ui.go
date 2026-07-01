// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
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
		mutedBgStyle := lipgloss.NewStyle().
			Foreground(styles.Colors.Muted).
			Background(styles.Colors.Background)
		saveText := mutedBgStyle.Render("[Space/Enter] Save  [ESC] Cancel")
		if focused {
			successBgStyle := lipgloss.NewStyle().
				Foreground(styles.Colors.Success).
				Background(styles.Colors.Background)
			saveText = successBgStyle.Bold(true).Render("[Space/Enter] Save") + "  " + mutedBgStyle.Render("[ESC] Cancel")
		}
		return "\n" + saveText

	}

	// Handle list items (hardDisks_N, cdroms_N) and add buttons
	if strings.HasPrefix(pos.Key, "hardDisks_") || strings.HasPrefix(pos.Key, "cdroms_") {
		if strings.HasSuffix(pos.Key, "_add") {
			// Add button — use explicit Background on all parts to prevent
			// ANSI Bold bleed between concatenated Render() calls.
			btnStyle := lipgloss.NewStyle().
				Foreground(styles.Colors.Primary).
				Background(styles.Colors.Background)
			mutedBgStyle := lipgloss.NewStyle().
				Foreground(styles.Colors.Muted).
				Background(styles.Colors.Background)

			btnText := mutedBgStyle.Render("[+ Add]")
			if focused {
				btnText = btnStyle.Bold(true).Render("[+ Add]")
			}
			fieldLabel := "Disk"
			if strings.HasPrefix(pos.Key, "cdroms_") {
				fieldLabel = "CD/DVD"
			}
			return "  " + btnText + " " + mutedBgStyle.Render(fieldLabel)
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
	// Shared styles with explicit Background to prevent ANSI bleed on 16-color terms
	textStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Primary).
		Background(styles.Colors.Background)
	mutedStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted).
		Background(styles.Colors.Background)
	// Cursor uses explicit foreground/background swap instead of Reverse(true)
	// (SGR 7 reverse-video resets unreliably on 16-color consoles).
	cursorStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Background).
		Background(styles.Colors.Primary)

	prefix := "  "
	if focused {
		prefix = textStyle.Bold(true).Render("> ")
	}

	labelPart := mutedStyle.Render(label + ": ")

	// Build value with cursor indicator
	var valPart string
	if focused {
		// Show cursor as inverted-color character or underscore at end
		if cursor < len(value) {
			before := value[:cursor]
			at := string(value[cursor])
			after := ""
			if cursor+1 < len(value) {
				after = value[cursor+1:]
			}
			valPart = textStyle.Render(before) +
				cursorStyle.Render(at) +
				textStyle.Render(after)
		} else {
			valPart = textStyle.Render(value) + cursorStyle.Render("_")
		}
	} else {
		if value == "" {
			valPart = mutedStyle.Render("(empty)")
		} else {
			valPart = textStyle.Render(value)
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
	// Shared styles with explicit Background to prevent ANSI bleed on 16-color terms
	textStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Primary).
		Background(styles.Colors.Background)
	mutedStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted).
		Background(styles.Colors.Background)
	// Cursor uses explicit foreground/background swap instead of Reverse(true)
	// (SGR 7 reverse-video resets unreliably on 16-color consoles).
	cursorStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Background).
		Background(styles.Colors.Primary)

	numPart := mutedStyle.Render(fmt.Sprintf("  [%d] ", index+1))

	var valPart string
	if focused {
		if cursor < len(value) {
			before := value[:cursor]
			at := string(value[cursor])
			after := ""
			if cursor+1 < len(value) {
				after = value[cursor+1:]
			}
			valPart = textStyle.Render(before) +
				cursorStyle.Render(at) +
				textStyle.Render(after)
		} else {
			valPart = textStyle.Render(value) + cursorStyle.Render("_")
		}
	} else {
		if value == "" {
			valPart = mutedStyle.Render("(enter path)")
		} else {
			valPart = textStyle.Render(value)
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
	// Styles with explicit Background for consistent 16-color rendering
	onStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Primary).
		Background(styles.Colors.Background)
	offStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted).
		Background(styles.Colors.Background)

	prefix := "  "
	if focused {
		prefix = onStyle.Bold(true).Render("> ")
	}

	labelPart := offStyle.Render(label + ": ")

	var valPart string
	switch fieldName {
	case "networkMode":
		if m.networkMode == "bridge" {
			if focused {
				valPart = onStyle.Bold(true).Render("[Bridge]")
			} else {
				valPart = onStyle.Render("[Bridge]")
			}
		} else {
			if focused {
				valPart = offStyle.Bold(true).Render("[NAT]")
			} else {
				valPart = offStyle.Render("[NAT]")
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
				valPart = onStyle.Bold(true).Render("[ON]")
			} else {
				valPart = onStyle.Render("[ON]")
			}
		} else {
			if focused {
				valPart = offStyle.Bold(true).Render("[OFF]")
			} else {
				valPart = offStyle.Render("[OFF]")
			}
		}
	}

	return prefix + labelPart + valPart
}