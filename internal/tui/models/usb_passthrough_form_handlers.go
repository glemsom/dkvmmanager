// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// HandleEnter is called when the user presses Enter on a position.
func (m *USBPassthroughFormModel) HandleEnter(pos form.FocusPos) (form.FormResult, tea.Cmd) {
	switch pos.Kind {
	case form.FocusToggle:
		m.toggleDevice(pos.Key)
		m.positions = m.BuildPositions()
		return form.ResultNone, nil

	case form.FocusButton:
		if pos.Key == "save" {
			return m.validateAndSaveCmd()
		}
	}

	// Unknown position, move focus forward
	m.focusIndex++
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
	return form.ResultNone, nil
}

// HandleChar is a no-op (no text fields in USB passthrough form).
func (m *USBPassthroughFormModel) HandleChar(pos form.FocusPos, ch string) {}

// HandleBackspace is a no-op (no text fields in USB passthrough form).
func (m *USBPassthroughFormModel) HandleBackspace(pos form.FocusPos) {}

// HandleDelete is a no-op (no text fields in USB passthrough form).
func (m *USBPassthroughFormModel) HandleDelete(pos form.FocusPos) {}
