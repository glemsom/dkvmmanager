// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// validateAndSaveCmd validates the form and saves the configuration.
func (m *VCPUPinningFormModel) validateAndSaveCmd() (form.FormResult, tea.Cmd) {
	m.errors = make(map[string]string)

	// If enabling, recompute mappings from current topology
	if m.pinning.Enabled {
		computed, err := vm.ComputePinningFromTopology(m.topology, m.hostTopo)
		if err != nil {
			m.errors["save"] = fmt.Sprintf("failed to compute pinning: %v", err)
			return form.ResultNone, nil
		}
		m.pinning.Mappings = computed.Mappings
	}

	// Save UseHostTopology to CPU topology config
	m.topology.UseHostTopology = m.useHostTopology
	if err := m.repo.SaveCPUTopology(m.topology); err != nil {
		m.errors["save"] = fmt.Sprintf("failed to save topology: %v", err)
		return form.ResultNone, nil
	}

	// Save the pinning configuration
	if err := m.repo.SaveVCPUPinningGlobal(m.pinning); err != nil {
		m.errors["save"] = fmt.Sprintf("failed to save: %v", err)
		return form.ResultNone, nil
	}

	return form.ResultSave, func() tea.Msg { return VCPUPinningUpdatedMsg{} }
}

// handleApplyKernelCmd applies the current vCPU pinning config to grub.cfg as CPU isolation params.
func (m *VCPUPinningFormModel) handleApplyKernelCmd() tea.Cmd {
	m.errors = make(map[string]string)
	m.kernelMsg = ""
	m.kernelMsgErr = false

	// If enabling, recompute mappings from current topology
	if m.pinning.Enabled {
		computed, err := vm.ComputePinningFromTopology(m.topology, m.hostTopo)
		if err != nil {
			m.kernelMsg = fmt.Sprintf("failed to compute pinning: %v", err)
			m.kernelMsgErr = true
			return nil
		}
		m.pinning.Mappings = computed.Mappings
	}

	// Save UseHostTopology to CPU topology config first
	m.topology.UseHostTopology = m.useHostTopology
	if err := m.repo.SaveCPUTopology(m.topology); err != nil {
		m.kernelMsg = fmt.Sprintf("Failed to save topology: %v", err)
		m.kernelMsgErr = true
		return nil
	}

	// Save config to disk first
	if err := m.repo.SaveVCPUPinningGlobal(m.pinning); err != nil {
		m.kernelMsg = fmt.Sprintf("Failed to save config: %v", err)
		m.kernelMsgErr = true
		return nil
	}

	// Apply to kernel grub.cfg asynchronously
	return func() tea.Msg {
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
