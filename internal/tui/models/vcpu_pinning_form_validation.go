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

// handleApplyKernel applies the current vCPU pinning config to grub.cfg as CPU isolation params
func (m *VCPUPinningFormModel) handleApplyKernel() (tea.Model, tea.Cmd) {
	m.errors = make(map[string]string)
	m.kernelMsg = ""
	m.kernelMsgErr = false

	// If enabling, recompute mappings from current topology
	if m.pinning.Enabled {
		computed, err := vm.ComputePinningFromTopology(m.topology, m.hostTopo)
		if err != nil {
			m.kernelMsg = fmt.Sprintf("failed to compute pinning: %v", err)
			m.kernelMsgErr = true
			m.syncViewport()
			return m, nil
		}
		m.pinning.Mappings = computed.Mappings
	}

	// Save config to disk first
	if err := m.vmManager.SaveVCPUPinningGlobal(m.pinning); err != nil {
		m.kernelMsg = fmt.Sprintf("Failed to save config: %v", err)
		m.kernelMsgErr = true
		m.syncViewport()
		return m, nil
	}

	// Apply to kernel grub.cfg asynchronously
	return m, func() tea.Msg {
		err := m.vmManager.ApplyCPUParamsToKernel()
		if err != nil {
			return VCPUCPUKernelAppliedMsg{
				Success: false,
				Error:   err.Error(),
			}
		}
		return VCPUCPUKernelAppliedMsg{
			Success: true,
		}
	}
}