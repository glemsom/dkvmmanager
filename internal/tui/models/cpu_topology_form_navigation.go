// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"
)

// moveFocus moves focus by delta in the flat positions list
func (m *CPUTopologyFormModel) moveFocus(delta int) {
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// syncViewport regenerates the rendered lines and syncs the viewport
func (m *CPUTopologyFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	totalContent := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(totalContent)
	if m.focusedLineIndex() >= 0 {
		m.vp.SetYOffset(clampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height))
	}
}

// focusedLineIndex maps focusIndex to a rendered line index
func (m *CPUTopologyFormModel) focusedLineIndex() int {
	line := 0

	// Header + blank + host info + blank = 4 lines
	line += 4

	if m.scanErr != nil {
		line += 2
	}

	lastDieID := -1
	for i, p := range m.positions {
		// Count lines for this position BEFORE checking if it's focused
		if p.kind == cpuTopoToggle && p.dieID != lastDieID {
			if lastDieID != -1 {
				line++
			}
			line++ // die header
			lastDieID = p.dieID
		}

		switch p.kind {
		case cpuTopoToggle:
			line++
		case cpuTopoSave:
			line += 4 // blank + summary + blank + button
			if m.hostCoreCount() == 0 {
				line++ // zero-host warning
			}
		}

		// Now check if this is the focused position
		if i == m.focusIndex {
			return line
		}
	}

	return line
}

// hostCoreCount returns the number of cores not allocated to VM
func (m *CPUTopologyFormModel) hostCoreCount() int {
	allocated := 0
	for _, pos := range m.positions {
		if pos.kind == cpuTopoToggle && m.coreSelected[coreKey(pos.dieID, pos.coreID)] {
			allocated++
		}
	}
	return m.hostTopo.TotalCores - allocated
}
