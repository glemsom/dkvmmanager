// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"log"

	tea "charm.land/bubbletea/v2"
)

// showVMSelection shows a list of VMs to edit
func (m *MainModel) showVMSelection() (tea.Model, tea.Cmd) {
	return m.showVMSelectionWithMode("edit", "No VMs available to edit")
}

// showVMSelectionForDeletion shows a list of VMs to delete
func (m *MainModel) showVMSelectionForDeletion() (tea.Model, tea.Cmd) {
	return m.showVMSelectionWithMode("delete", "No VMs available to delete")
}

// showVMSelectionWithMode shows a list of VMs with the specified selection mode.
// It creates a VMSelectModel and activates it in the view registry.
func (m *MainModel) showVMSelectionWithMode(mode string, emptyMessage string) (tea.Model, tea.Cmd) {
	vms, err := m.vmManager.ListVMs()
	if err != nil || len(vms) == 0 {
		m.statusBar.SetMessage(emptyMessage)
		return m, nil
	}

	// Create VMSelectModel
	selModel := NewVMSelectModel(m.vmManager, vms, mode, m.debugMode)
	if m.windowWidth > 0 && m.windowHeight > 0 {
		selModel.SetSize(m.windowWidth-4, m.contentHeight()-2)
	}

	// Activate through registry
	if m.viewRegistry != nil && m.viewRegistry.GetDef(ViewVMSelect) != nil {
		m.viewRegistry.SetActiveModel(m.viewRegistry.GetDef(ViewVMSelect), selModel)
	}
	m.currentView = ViewVMSelect

	// Update breadcrumbs
	m.breadcrumbs.Clear()
	m.breadcrumbs.AddItem("Configuration", "config", 1)
	if mode == "delete" {
		m.breadcrumbs.AddItem("Delete VM", "vm_delete", 1)
	} else {
		m.breadcrumbs.AddItem("Edit VM", "vm_edit", 1)
	}

	if m.debugMode {
		log.Printf("[DEBUG] showVMSelectionWithMode(%s): switching to ViewVMSelect with %d VMs", mode, len(vms))
	}

	return m, nil
}

// renderVMSelectView renders the VM selection view for editing
func (m *MainModel) renderVMSelectView() string {
	if m.viewRegistry != nil && m.viewRegistry.ActiveName() == ViewVMSelect {
		return m.viewRegistry.ActiveModel().View().Content
	}
	return "Loading..."
}
