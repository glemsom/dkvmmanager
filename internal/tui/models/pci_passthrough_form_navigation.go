// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strconv"

	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// pciFocusData carries per-position metadata through form.FocusPos.Data.
type pciFocusData struct {
	Address  string // PCI device address (for toggles)
	GroupNum int    // IOMMU group number (for headers and toggles)
}

// BuildPositions reconstructs the flat focus list from IOMMU-grouped devices.
// Group headers are inserted as FocusHeader positions (render-only).
// Returns the positions slice and stores it in m.positions.
func (m *PCIPassthroughFormModel) BuildPositions() []form.FocusPos {
	var positions []form.FocusPos

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
		// Group header (render-only, no interaction)
		var label string
		if group < 0 {
			label = "Ungrouped Devices"
		} else {
			label = "IOMMU Group " + strconv.Itoa(group)
		}
		positions = append(positions, form.FocusPos{
			Kind:  form.FocusHeader,
			Label: label,
			Key:   "group_" + strconv.Itoa(group),
			Data:  pciFocusData{GroupNum: group},
		})

		// Devices within this group
		for _, dev := range m.iommuGroups[group] {
			positions = append(positions, form.FocusPos{
				Kind:  form.FocusToggle,
				Label: dev.Name,
				Key:   dev.Address,
				Data:  pciFocusData{Address: dev.Address, GroupNum: group},
			})
		}
	}

	// Save button
	positions = append(positions, form.FocusPos{
		Kind:  form.FocusButton,
		Label: "Save",
		Key:   "save",
	})

	// Apply to Kernel button
	positions = append(positions, form.FocusPos{
		Kind:  form.FocusButton,
		Label: "Apply to Kernel",
		Key:   "apply_kernel",
	})

	m.positions = positions
	return positions
}

// currentPos returns the focus position at the current focusIndex
func (m *PCIPassthroughFormModel) currentPos() form.FocusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return form.FocusPos{Kind: form.FocusButton, Key: "save", Label: "Save"}
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
// skipping non-focusable positions (headers).
func (m *PCIPassthroughFormModel) moveFocus(delta int) {
	if delta == 0 {
		return
	}
	m.focusIndex += delta

	// Skip non-focusable positions (headers)
	for m.focusIndex >= 0 && m.focusIndex < len(m.positions) &&
		m.positions[m.focusIndex].Kind == form.FocusHeader {
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
