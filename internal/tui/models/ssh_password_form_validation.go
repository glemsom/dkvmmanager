// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// sshPasswordErrorMsg is sent when password change fails.
type sshPasswordErrorMsg struct {
	err string
}

// passwordStrength returns a score 0-5 based on password complexity
func passwordStrength(pw string) int {
	score := 0
	if len(pw) >= 8 {
		score++
	}
	if len(pw) >= 10 {
		score++
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false
	for _, ch := range pw {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	if hasLower {
		score++
	}
	if hasUpper {
		score++
	}
	if hasDigit || hasSpecial {
		score++
	}
	if score > 5 {
		score = 5
	}
	return score
}

// validate checks all fields and returns whether validation passed
func (m *SSHPasswordFormModel) validate() bool {
	m.errors = make(map[string]string)

	if m.newPassword == "" {
		m.errors["newPassword"] = "Password is required"
	}
	if m.confirmPassword == "" {
		m.errors["confirmPassword"] = "Please confirm the password"
	}
	if m.newPassword != "" && len(m.newPassword) < 6 {
		m.errors["newPassword"] = "Password must be at least 6 characters"
	}
	if m.newPassword != "" && m.confirmPassword != "" && m.newPassword != m.confirmPassword {
		m.errors["confirmPassword"] = "Passwords do not match"
	}

	return len(m.errors) == 0
}

// apply executes the password change and returns a tea.Cmd
func (m *SSHPasswordFormModel) apply() tea.Cmd {
	pw := m.newPassword
	return func() tea.Msg {
		if dryRunMode {
			log.Printf("[DEBUG] Dry-run mode: would execute: echo \"%s:***\" | chpasswd", os.Getenv("USER"))
			return SSHPasswordUpdatedMsg{}
		}

		username := os.Getenv("USER")
		if username == "" {
			username = "root"
		}

		cmd := exec.Command("chpasswd")
		cmd.Stdin = strings.NewReader(username + ":" + pw)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return sshPasswordErrorMsg{err: fmt.Sprintf("chpasswd failed: %s", strings.TrimSpace(string(output)))}
		}

		// Persist changes via lbu
		lbuCmd := exec.Command("lbu", "commit")
		lbuOutput, lbuErr := lbuCmd.CombinedOutput()
		if lbuErr != nil {
			return sshPasswordErrorMsg{err: fmt.Sprintf("Password changed but lbu commit failed: %s", strings.TrimSpace(string(lbuOutput)))}
		}

		return SSHPasswordUpdatedMsg{}
	}
}

// strengthLabel returns a label and color for the given strength score
func strengthLabel(score int) (string, lipgloss.Color) {
	switch {
	case score <= 1:
		return "Weak", styles.Colors.Error
	case score <= 3:
		return "Fair", styles.Colors.Warning
	default:
		return "Strong", styles.Colors.Success
	}
}

// fieldLabel returns the human-readable label for a field name
func (m *SSHPasswordFormModel) fieldLabel(name string) string {
	switch name {
	case "newPassword":
		return "New Password"
	case "confirmPassword":
		return "Confirm Password"
	}
	return name
}
