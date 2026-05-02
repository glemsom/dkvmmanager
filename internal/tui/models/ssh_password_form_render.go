// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderPasswordInput renders a masked password input field.
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
