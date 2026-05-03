// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// moveFocus moves focus by delta in the flat positions list (backward compat for tests).
func (m *CPUTopologyFormModel) moveFocus(delta int) {
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// hostCoreCount returns the number of cores not allocated to VM
func (m *CPUTopologyFormModel) hostCoreCount() int {
	allocated := 0
	for _, pos := range m.positions {
		if pos.Kind == form.FocusToggle {
			d := pos.Data.(cpuTopoFocusData)
			key := coreKey(d.dieID, d.coreID)
			if m.coreSelected[key] {
				allocated++
			}
		}
	}
	return m.hostTopo.TotalCores - allocated
}
