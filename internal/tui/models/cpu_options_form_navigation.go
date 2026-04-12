// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import "strings"

// rebuildPositions reconstructs the flat focus list from all fields
func (m *CPUOptionsFormModel) rebuildPositions() {
	m.positions = nil

	// Hypervisor Stealth section
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HideKVM"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptText, fieldName: "VendorID"})

	// Hyper-V Enlightenments section
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVFrequency"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVRelaxed"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVReset"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVRuntime"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptText, fieldName: "HVSpinlocks"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVStimer"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVSyncIC"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVTime"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVVapic"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVVPIndex"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVNoNonarchCoresharing"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVTLBFlush"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVTLBFlushExt"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVIPI"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HVAVIC"})

	// Advanced CPU Features section
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "TopoExt"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "L3Cache"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "X2APIC"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "Migratable"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "InvTSC"})
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptToggle, fieldName: "RTCUTC"})

	// Save button
	m.positions = append(m.positions, cpuOptFocusPos{kind: cpuOptSave, fieldName: "save"})
}

// currentPos returns the focus position at the current focusIndex
func (m *CPUOptionsFormModel) currentPos() cpuOptFocusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return cpuOptFocusPos{kind: cpuOptToggle, fieldName: "HideKVM"}
	}
	return m.positions[m.focusIndex]
}

// moveFocus moves focus by delta in the flat positions list
func (m *CPUOptionsFormModel) moveFocus(delta int) {
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// pageUp moves focus up by a page (approximately half the viewport height)
func (m *CPUOptionsFormModel) pageUp() {
	pageSize := 10
	m.focusIndex -= pageSize
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	m.syncViewport()
}

// pageDown moves focus down by a page (approximately half the viewport height)
func (m *CPUOptionsFormModel) pageDown() {
	pageSize := 10
	m.focusIndex += pageSize
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
	m.syncViewport()
}

// syncViewport regenerates the rendered lines and syncs the viewport
func (m *CPUOptionsFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	totalContent := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(totalContent)
	if m.focusedLineIndex() >= 0 {
		m.vp.SetYOffset(clampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height))
	}
}

// focusedLineIndex maps focusIndex to a rendered line index
func (m *CPUOptionsFormModel) focusedLineIndex() int {
	line := 0

	for i, p := range m.positions {
		if i == m.focusIndex {
			return line
		}

		// Section headers
		if i == 0 {
			line += 2 // section header + blank
		} else if i == 2 {
			line += 2 // section header + blank
		} else if i == 17 {
			line += 2 // section header + blank
		}

		switch p.kind {
		case cpuOptToggle:
			line++
		case cpuOptText:
			line++
		case cpuOptSave:
			line += 2 // blank + button
		}
	}

	return line
}
