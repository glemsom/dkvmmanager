// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// --- Rendering ---

var (
	sshPwLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	sshPwFocusStyle = styles.FormFocusStyle()
	sshPwInputStyle = styles.FormInputStyle()
	sshPwErrorStyle = styles.ErrorTextStyle()
	sshPwMutedStyle = styles.FormMutedStyle()
	sshPwSaveStyle  = styles.FormSaveStyle()
)

// renderAllLines produces the full list of output lines for the form
func (m *SSHPasswordFormModel) renderAllLines() []string {
	var lines []string

	lines = append(lines, sshPwFocusStyle.Render("Set SSH Password"))
	lines = append(lines, "")

	for i, pos := range m.positions {
		focused := (i == m.focusIndex)
		switch pos.kind {
		case sshPwText:
			label := m.fieldLabel(pos.fieldName)
			val := m.getTextValue(pos.fieldName)
			cursor := m.effectiveCursor(pos.fieldName, val)
			lines = append(lines, m.renderPasswordInput(label, val, cursor, focused))
			if errMsg, ok := m.errors[pos.fieldName]; ok {
				lines = append(lines, "  "+sshPwErrorStyle.Render(errMsg))
			}
		case sshPwApply:
			lines = append(lines, "")
			// Strength indicator
			strength := passwordStrength(m.newPassword)
			strengthText, strengthColor := strengthLabel(strength)
			barWidth := 10
			filled := int(float64(strength) / 5.0 * float64(barWidth))
			bar := lipgloss.NewStyle().Foreground(strengthColor).Render(strings.Repeat("█", filled)) +
				sshPwMutedStyle.Render(strings.Repeat("░", barWidth-filled))
			lines = append(lines, sshPwLabelStyle.Render("Strength: ")+bar+" "+lipgloss.NewStyle().Foreground(strengthColor).Render(strengthText))
			lines = append(lines, "")

			// Apply button
			applyText := sshPwMutedStyle.Render("[Apply]")
			if focused {
				applyText = sshPwSaveStyle.Render("[Apply]")
			}
			lines = append(lines, applyText)
		}
	}

	// Status message
	if m.statusMessage != "" {
		lines = append(lines, "")
		lines = append(lines, sshPwErrorStyle.Render(m.statusMessage))
	}

	// Footer
	lines = append(lines, "")
	lines = append(lines, sshPwMutedStyle.Render("Tab/Shift+Tab Navigate  Space/Enter Select  ESC Cancel"))

	return lines
}

// renderPasswordInput renders a masked password input field
func (m *SSHPasswordFormModel) renderPasswordInput(label, value string, cursor int, focused bool) string {
	prefix := "  "
	if focused {
		prefix = sshPwFocusStyle.Render("> ")
	}

	labelPart := sshPwLabelStyle.Render(label + ": ")

	// Mask the password with asterisk characters
	masked := strings.Repeat("*", len(value))

	var valPart string
	if focused {
		if cursor < len(masked) {
			before := masked[:cursor]
			at := string(masked[cursor])
			after := ""
			if cursor+1 < len(masked) {
				after = masked[cursor+1:]
			}
			valPart = sshPwInputStyle.Render(before) +
				lipgloss.NewStyle().Reverse(true).Render(at) +
				sshPwInputStyle.Render(after)
		} else {
			valPart = sshPwInputStyle.Render(masked) + sshPwFocusStyle.Render("_")
		}
	} else {
		if value == "" {
			valPart = sshPwMutedStyle.Render("(empty)")
		} else {
			valPart = sshPwInputStyle.Render(masked)
		}
	}

	return prefix + labelPart + valPart
}
