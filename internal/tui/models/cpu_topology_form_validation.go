// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// validateAndSaveCmd persists the global CPU topology config and returns a tea.Cmd.
func (m *CPUTopologyFormModel) validateAndSaveCmd() (form.FormResult, tea.Cmd) {
	m.errors = make(map[string]string)

	// Build selected CPU list from all selected cores' thread IDs
	var selectedCPUs []int
	for _, pos := range m.positions {
		if pos.Kind != form.FocusToggle {
			continue
		}
		d := pos.Data.(cpuTopoFocusData)
		key := coreKey(d.dieID, d.coreID)
		if m.coreSelected[key] && d.coreInfo != nil {
			selectedCPUs = append(selectedCPUs, d.coreInfo.Threads...)
		}
	}
	sort.Ints(selectedCPUs)

	if len(selectedCPUs) == 0 {
		m.errors["save"] = "At least one core must be allocated for VMs"
		return form.ResultNone, nil
	}

	topo := models.CPUTopology{
		Enabled:      true,
		SelectedCPUs: selectedCPUs,
	}

	if err := m.vmManager.SaveCPUTopology(topo); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		return form.ResultNone, nil
	}

	return form.ResultNone, func() tea.Msg {
		return CPUTopologyUpdatedMsg{}
	}
}
