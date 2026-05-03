// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// BuildPositions reconstructs the flat focus list from scanned USB devices.
// Each device becomes a FocusToggle position, followed by a FocusButton for Save.
// Returns the positions slice and stores it in m.positions.
func (m *USBPassthroughFormModel) BuildPositions() []form.FocusPos {
	var positions []form.FocusPos

	for _, dev := range m.devices {
		positions = append(positions, form.FocusPos{
			Kind:  form.FocusToggle,
			Label: dev.Name,
			Key:   usbDeviceKey(dev.Vendor, dev.Product),
		})
	}

	// Save button
	positions = append(positions, form.FocusPos{
		Kind:  form.FocusButton,
		Label: "Save",
		Key:   "save",
	})

	m.positions = positions
	return positions
}

// currentPos returns the focus position at the current focusIndex.
func (m *USBPassthroughFormModel) currentPos() form.FocusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return form.FocusPos{Kind: form.FocusButton, Key: "save", Label: "Save"}
	}
	return m.positions[m.focusIndex]
}

// moveFocus moves focus by delta in the flat positions list.
// The USB form has no non-focusable positions (no headers), so simple clamping.
func (m *USBPassthroughFormModel) moveFocus(delta int) {
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}
