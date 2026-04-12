// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
)

// validateAndSave persists the global CPU topology config
func (m *CPUTopologyFormModel) validateAndSave() (tea.Model, tea.Cmd) {
	m.errors = make(map[string]string)

	// Build selected CPU list from all selected cores' thread IDs
	var selectedCPUs []int
	for _, pos := range m.positions {
		if pos.kind != cpuTopoToggle {
			continue
		}
		key := coreKey(pos.dieID, pos.coreID)
		if m.coreSelected[key] && pos.coreInfo != nil {
			selectedCPUs = append(selectedCPUs, pos.coreInfo.Threads...)
		}
	}
	sort.Ints(selectedCPUs)

	if len(selectedCPUs) == 0 {
		m.errors["save"] = "At least one core must be allocated for VMs"
		m.syncViewport()
		return m, nil
	}

	topo := models.CPUTopology{
		Enabled:      true,
		SelectedCPUs: selectedCPUs,
	}

	if err := m.vmManager.SaveCPUTopology(topo); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		m.syncViewport()
		return m, nil
	}

	return m, func() tea.Msg {
		return CPUTopologyUpdatedMsg{}
	}
}
