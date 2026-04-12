// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// validateAndSave validates the form and saves the configuration
func (m *VCPUPinningFormModel) validateAndSave() (tea.Model, tea.Cmd) {
	m.errors = make(map[string]string)

	// If enabling, recompute mappings from current topology
	if m.pinning.Enabled {
		computed, err := vm.ComputePinningFromTopology(m.topology, m.hostTopo)
		if err != nil {
			m.errors["save"] = fmt.Sprintf("failed to compute pinning: %v", err)
			m.syncViewport()
			return m, nil
		}
		m.pinning.Mappings = computed.Mappings
	}

	// Save the pinning configuration
	if err := m.vmManager.SaveVCPUPinningGlobal(m.pinning); err != nil {
		m.errors["save"] = fmt.Sprintf("failed to save: %v", err)
		m.syncViewport()
		return m, nil
	}

	m.syncViewport()
	return m, func() tea.Msg { return VCPUPinningUpdatedMsg{} }
}