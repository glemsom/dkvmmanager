// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"
)

// moveFocus moves focus by delta in the flat positions list
func (m *SSHPasswordFormModel) moveFocus(delta int) {
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// syncViewport regenerates the rendered lines and syncs the viewport
func (m *SSHPasswordFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	totalContent := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(totalContent)
	if m.focusedLineIndex() >= 0 {
		m.vp.SetYOffset(clampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height))
	}
}

// focusedLineIndex maps focusIndex to a rendered line index
func (m *SSHPasswordFormModel) focusedLineIndex() int {
	line := 0 // title
	line++    // blank after title

	for i := range m.positions {
		if i == m.focusIndex {
			return line
		}
		pos := m.positions[i]
		switch pos.kind {
		case sshPwText:
			line++ // label+input
			if _, hasErr := m.errors[pos.fieldName]; hasErr {
				line++ // error line
			}
		case sshPwApply:
			line++ // blank separator
			line++ // button
		}
	}
	return line
}
