// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// rebuildPositions reconstructs the flat focus list from scanned devices
func (m *PCIPassthroughFormModel) rebuildPositions() {
	m.positions = nil

	for _, dev := range m.devices {
		// Toggle for each device
		m.positions = append(m.positions, pciFocusPos{
			kind:       pciToggle,
			deviceAddr: dev.Address,
		})

		// ROM path field only for selected devices
		if m.selected[dev.Address] {
			m.positions = append(m.positions, pciFocusPos{
				kind:       pciROMPath,
				deviceAddr: dev.Address,
			})
		}
	}

	// Save button
	m.positions = append(m.positions, pciFocusPos{
		kind:       pciSave,
		deviceAddr: "",
	})
}

// currentPos returns the focus position at the current focusIndex
func (m *PCIPassthroughFormModel) currentPos() pciFocusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return pciFocusPos{kind: pciSave}
	}
	return m.positions[m.focusIndex]
}

// getDeviceByAddr finds a device by PCI address
func (m *PCIPassthroughFormModel) getDeviceByAddr(addr string) *models.PCIDevice {
	for i := range m.devices {
		if m.devices[i].Address == addr {
			return &m.devices[i]
		}
	}
	return nil
}

// cursorOffset returns the cursor offset for the given position key
func (m *PCIPassthroughFormModel) cursorOffset(key string) int {
	if off, ok := m.cursorOffsets[key]; ok {
		return off
	}
	return -1
}

// setCursorOffset sets cursor offset; -1 means end
func (m *PCIPassthroughFormModel) setCursorOffset(key string, off int) {
	m.cursorOffsets[key] = off
}

// effectiveCursor returns the actual cursor position
func (m *PCIPassthroughFormModel) effectiveCursor(key string, val string) int {
	off := m.cursorOffset(key)
	if off < 0 {
		return len(val)
	}
	if off > len(val) {
		return len(val)
	}
	return off
}

// moveFocus moves focus by delta in the flat positions list
func (m *PCIPassthroughFormModel) moveFocus(delta int) {
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// toggleDevice toggles selection of a PCI device
func (m *PCIPassthroughFormModel) toggleDevice(addr string) {
	if m.selected[addr] {
		delete(m.selected, addr)
	} else {
		m.selected[addr] = true
	}
}

// focusedLineIndex maps focusIndex to a rendered line index
func (m *PCIPassthroughFormModel) focusedLineIndex() int {
	line := 0

	// Header + blank = 2 lines
	line += 2

	// Scan error warning if any
	if m.scanErr != nil {
		line += 2
	}

	for i, p := range m.positions {
		if i == m.focusIndex {
			return line
		}

		switch p.kind {
		case pciToggle:
			line++
		case pciROMPath:
			line++
		case pciSave:
			line++ // blank before button
			line++ // button
		}
	}

	return line
}

// syncViewport regenerates the rendered lines and syncs the viewport
func (m *PCIPassthroughFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	totalContent := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(totalContent)
	if m.focusedLineIndex() >= 0 {
		m.vp.SetYOffset(clampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height))
	}
}
