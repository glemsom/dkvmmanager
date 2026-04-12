// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"log"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// showVMSelection shows a list of VMs to edit
func (m *MainModel) showVMSelection() (tea.Model, tea.Cmd) {
	return m.showVMSelectionWithMode("edit", "No VMs available to edit")
}

// showVMSelectionForDeletion shows a list of VMs to delete
func (m *MainModel) showVMSelectionForDeletion() (tea.Model, tea.Cmd) {
	return m.showVMSelectionWithMode("delete", "No VMs available to delete")
}

// showVMSelectionWithMode shows a list of VMs with the specified selection mode
func (m *MainModel) showVMSelectionWithMode(mode string, emptyMessage string) (tea.Model, tea.Cmd) {
	vms, err := m.vmManager.ListVMs()
	if err != nil || len(vms) == 0 {
		m.statusMessage = emptyMessage
		return m, nil
	}

	// Sort VMs by ID to ensure deterministic ordering (matches menu list)
	sort.Slice(vms, func(i, j int) bool {
		return vms[i].ID < vms[j].ID
	})

	// Store VMs for selection and switch to selection view
	m.vmListForSelection = vms
	m.currentView = ViewVMSelect
	m.selectedIndex = 0
	m.selectionMode = mode

	// Create VM selection list
	vmListAdapter := buildVMListAdapter(vms)
	vmDelegate := VMListItemDelegate{}
	vmSelectList := list.New(vmListAdapter, vmDelegate, m.windowWidth-4, m.contentHeight()-2)
	vmSelectList.SetShowTitle(false)
	vmSelectList.SetShowStatusBar(false)
	vmSelectList.SetFilteringEnabled(false)
	vmSelectList.SetShowHelp(false)
	m.vmSelectList = vmSelectList

	// Update breadcrumbs
	m.breadcrumbs.Clear()
	m.breadcrumbs.AddItem("Configuration", "config", 1)
	if mode == "delete" {
		m.breadcrumbs.AddItem("Delete VM", "vm_delete", 1)
	} else {
		m.breadcrumbs.AddItem("Edit VM", "vm_edit", 1)
	}

	if debugMode {
		log.Printf("[DEBUG] showVMSelectionWithMode(%s): switching to ViewVMSelect with %d VMs", mode, len(vms))
		for i, vm := range vms {
			log.Printf("[DEBUG] VM[%d]: %s (ID: %s)", i, vm.Name, vm.ID)
		}
	}

	return m, nil
}

// ensureVMSelectList lazily initializes the VM selection list if it has items
// but the list model hasn't been set up yet (e.g., when the view is set directly
// without going through showVMSelectionWithMode).
func (m *MainModel) ensureVMSelectList() {
	if len(m.vmListForSelection) > 0 && len(m.vmSelectList.Items()) == 0 {
		vmListAdapter := buildVMListAdapter(m.vmListForSelection)
		vmDelegate := VMListItemDelegate{}
		m.vmSelectList = list.New(vmListAdapter, vmDelegate, m.windowWidth-4, m.contentHeight()-2)
		m.vmSelectList.SetShowTitle(false)
		m.vmSelectList.SetShowStatusBar(false)
		m.vmSelectList.SetFilteringEnabled(false)
		m.vmSelectList.SetShowHelp(false)
	}
}

// renderVMSelectView renders the VM selection view for editing
func (m *MainModel) renderVMSelectView() string {
	m.ensureVMSelectList()
	m.vmSelectList.SetSize(m.windowWidth-4, m.contentHeight()-2)
	output := m.vmSelectList.View()

	output += "\n\n↑/↓ Navigate  Space/Enter Select  ESC Cancel\n"

	return output
}
