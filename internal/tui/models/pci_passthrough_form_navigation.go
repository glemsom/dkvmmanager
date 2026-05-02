// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// rebuildPositions reconstructs the flat focus list from IOMMU-grouped devices.
// Group headers are inserted as pciGroupHeader positions but are marked as
// non-focusable (they are rendered for context but skipped during navigation).
func (m *PCIPassthroughFormModel) rebuildPositions() {
	m.positions = nil

	// Collect and sort group keys for deterministic ordering
	var groupKeys []int
	for k := range m.iommuGroups {
		groupKeys = append(groupKeys, k)
	}
	// Sort: negative (ungrouped) goes last, then ascending numeric order
	for i := 0; i < len(groupKeys); i++ {
		for j := i + 1; j < len(groupKeys); j++ {
			ki, kj := groupKeys[i], groupKeys[j]
			// Treat -1 as infinity for sorting (put it last)
			if ki == -1 || (kj != -1 && ki > kj) {
				groupKeys[i], groupKeys[j] = groupKeys[j], groupKeys[i]
			}
		}
	}

	for _, group := range groupKeys {
		// Group header (non-focusable — only for visual separation)
		m.positions = append(m.positions, pciFocusPos{
			kind:     pciGroupHeader,
			groupNum: group,
		})

		// Devices within this group
		for _, dev := range m.iommuGroups[group] {
			m.positions = append(m.positions, pciFocusPos{
				kind:       pciToggle,
				deviceAddr: dev.Address,
			})
		}
	}

	// Save button
	m.positions = append(m.positions, pciFocusPos{
		kind:       pciSave,
		deviceAddr: "",
	})

	// Apply to Kernel button
	m.positions = append(m.positions, pciFocusPos{
		kind:       pciApplyKernel,
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

// moveFocus moves focus by delta in the flat positions list,
// skipping non-focusable positions (group headers).
func (m *PCIPassthroughFormModel) moveFocus(delta int) {
	if delta == 0 {
		return
	}
	m.focusIndex += delta

	// Skip non-focusable positions (group headers)
	for m.focusIndex >= 0 && m.focusIndex < len(m.positions) &&
		m.positions[m.focusIndex].kind == pciGroupHeader {
		m.focusIndex += delta
	}

	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// toggleDevice toggles selection of a PCI device.
// If the device belongs to an IOMMU group with multiple devices,
// all devices in the same group are toggled together (strict mode).
func (m *PCIPassthroughFormModel) toggleDevice(addr string) {
	dev := m.getDeviceByAddr(addr)
	if dev == nil {
		return
	}

	group := dev.IOMMUGroup
	if group < 0 {
		group = -1
	}

	groupDevices, ok := m.iommuGroups[group]
	if !ok || len(groupDevices) <= 1 {
		// Ungrouped device or single-device group: toggle only this device
		if m.selected[addr] {
			delete(m.selected, addr)
		} else {
			m.selected[addr] = true
		}
		return
	}

	// Multi-device IOMMU group: strict mode — toggle ALL devices in the group
	newState := !m.selected[addr]
	for _, d := range groupDevices {
		if newState {
			m.selected[d.Address] = true
		} else {
			delete(m.selected, d.Address)
		}
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
		case pciGroupHeader:
			line++
		case pciToggle:
			line++
		case pciSave:
			line++ // blank before button
			line++ // button
		case pciApplyKernel:
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
