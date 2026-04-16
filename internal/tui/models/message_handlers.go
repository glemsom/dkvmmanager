// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

// handleViewChangeMsg handles ViewChangeMsg from sub-models
func (m *MainModel) handleViewChangeMsg(vcm ViewChangeMsg) (tea.Model, tea.Cmd) {
	if debugMode {
		log.Printf("[DEBUG] View change: %s -> %s", m.currentView, vcm.View)
	}
	m.currentView = vcm.View
	if vcm.View == ViewMainMenu {
		// Refresh menu items when returning from VM creation
		m.rebuildMenuList()
	}
	return m, nil
}

// handleVMCreatedMsg handles VMCreatedMsg from sub-models
func (m *MainModel) handleVMCreatedMsg(vcm VMCreatedMsg) (tea.Model, tea.Cmd) {
	m.statusMessage = fmt.Sprintf("VM '%s' created successfully", vcm.VMName)
	m.currentView = ViewConfigMenu
	m.rebuildMenuList()
	m.breadcrumbs.Clear()
	return m, nil
}

// handleVMUpdatedMsg handles VMUpdatedMsg from sub-models
func (m *MainModel) handleVMUpdatedMsg(vcm VMUpdatedMsg) (tea.Model, tea.Cmd) {
	m.statusMessage = fmt.Sprintf("VM '%s' updated successfully", vcm.VMName)
	m.currentView = ViewConfigMenu
	m.rebuildMenuList()
	m.breadcrumbs.Clear()
	return m, nil
}

// handleVMDeletedMsg handles VMDeletedMsg from sub-models
func (m *MainModel) handleVMDeletedMsg(vdm VMDeletedMsg) (tea.Model, tea.Cmd) {
	m.statusMessage = fmt.Sprintf("VM '%s' deleted successfully", vdm.VMName)
	m.currentView = ViewConfigMenu
	m.rebuildMenuList()
	return m, nil
}

// handleVMStoppedMsg handles VMStoppedMsg from running model
func (m *MainModel) handleVMStoppedMsg(vsm VMStoppedMsg) (tea.Model, tea.Cmd) {
	m.statusBar.SetMessage(fmt.Sprintf("VM '%s' stopped: %s", vsm.VMName, vsm.Reason))
	return m, nil
}

// handleLBUCommitMsg handles LBUCommitMsg completion
func (m *MainModel) handleLBUCommitMsg(lcm LBUCommitMsg) (tea.Model, tea.Cmd) {
	if lcm.Success {
		m.statusBar.SetMessage("LBU commit: " + lcm.Output)
	} else {
		m.statusBar.SetMessage("LBU commit failed: " + lcm.Output)
	}
	return m, nil
}

// handleSubViewMsg processes messages from sub-view command execution
func (m *MainModel) handleSubViewMsg(nextMsg tea.Msg) (tea.Model, tea.Cmd) {
	// Forward disk selection messages to the active sub-view's form.
	// Note: FileSelectedMsg is NOT handled here — it must be routed through
	// addDiskModel.Update() (in the view-specific blocks above) to convert
	// it to DiskAddedMsg. handleSubViewMsg would send it directly to
	// form.Update which only handles the CDROM case, bypassing addDiskModel.
	switch nextMsg.(type) {
	case DiskAddedMsg:
		switch m.currentView {
		case ViewVMCreate:
			if m.vmCreateModel != nil {
				inner, cmd := m.vmCreateModel.form.Update(nextMsg)
				if f, ok := inner.(*VMFormModel); ok {
					m.vmCreateModel.form = f
				}
				if cmd != nil {
					return m.handleSubViewMsg(cmd())
				}
			}
		case ViewVMEdit:
			if m.vmEditModel != nil {
				inner, cmd := m.vmEditModel.form.Update(nextMsg)
				if f, ok := inner.(*VMFormModel); ok {
					m.vmEditModel.form = f
				}
				if cmd != nil {
					return m.handleSubViewMsg(cmd())
				}
			}
		}
		return m, nil
	}

	if vcm, ok := nextMsg.(ViewChangeMsg); ok {
		m.currentView = vcm.View
		if vcm.View == ViewMainMenu {
			m.rebuildMenuList()
			m.breadcrumbs.Clear()
		}
		return m, nil
	}
	if vcm, ok := nextMsg.(VMCreatedMsg); ok {
		m.statusBar.SetMessage(fmt.Sprintf("VM '%s' created successfully", vcm.VMName))
		m.currentView = ViewConfigMenu
		m.rebuildMenuList()
		m.breadcrumbs.Clear()
		return m, nil
	}
	if vcm, ok := nextMsg.(VMUpdatedMsg); ok {
		m.statusBar.SetMessage(fmt.Sprintf("VM '%s' updated successfully", vcm.VMName))
		m.currentView = ViewConfigMenu
		m.rebuildMenuList()
		m.breadcrumbs.Clear()
		return m, nil
	}
	if vdm, ok := nextMsg.(VMDeletedMsg); ok {
		m.statusBar.SetMessage(fmt.Sprintf("VM '%s' deleted successfully", vdm.VMName))
		m.currentView = ViewConfigMenu
		m.rebuildMenuList()
		m.breadcrumbs.Clear()
		return m, nil
	}
	if _, ok := nextMsg.(CPUOptionsUpdatedMsg); ok {
		m.statusBar.SetMessage("CPU options saved successfully")
		m.currentView = ViewConfigMenu
		m.breadcrumbs.Clear()
		return m, nil
	}
	if _, ok := nextMsg.(PCIPassthroughUpdatedMsg); ok {
		m.statusBar.SetMessage("PCI passthrough saved successfully")
		m.currentView = ViewConfigMenu
		m.breadcrumbs.Clear()
		return m, nil
	}
	if _, ok := nextMsg.(USBPassthroughUpdatedMsg); ok {
		m.statusBar.SetMessage("USB passthrough saved successfully")
		m.currentView = ViewConfigMenu
		m.breadcrumbs.Clear()
		return m, nil
	}
	if _, ok := nextMsg.(CPUTopologyUpdatedMsg); ok {
		m.statusBar.SetMessage("CPU topology saved successfully")
		m.currentView = ViewConfigMenu
		m.breadcrumbs.Clear()
		return m, nil
	}
	if _, ok := nextMsg.(VCPUPinningUpdatedMsg); ok {
		m.statusBar.SetMessage("CPU topology and vCPU pinning saved successfully")
		m.currentView = ViewConfigMenu
		m.breadcrumbs.Clear()
		return m, nil
	}
	if _, ok := nextMsg.(SSHPasswordUpdatedMsg); ok {
		m.statusBar.SetMessage("Password changed successfully")
		m.sshPasswordModel = nil
		m.currentView = ViewConfigMenu
		m.breadcrumbs.Clear()
		return m, nil
	}
	if _, ok := nextMsg.(lvVGsLoadedMsg); ok {
		// lvVGsLoadedMsg is handled directly by LVCreateFormModel.Update()
		// when the view is active - no need to forward here
		return m, nil
	}
	if _, ok := nextMsg.(LVCreateUpdatedMsg); ok {
		if m.lvCreateFormModel != nil && m.lvCreateFormModel.preview != "" {
			m.statusBar.SetMessage(m.lvCreateFormModel.preview)
		} else {
			m.statusBar.SetMessage("Logical volume created successfully")
		}
		m.lvCreateFormModel = nil
		m.currentView = ViewConfigMenu
		m.breadcrumbs.Clear()
		return m, nil
	}
	if em, ok := nextMsg.(lvCreateErrorMsg); ok {
		m.statusBar.SetMessage("LV create failed: " + em.err)
		return m, nil
	}
	return m, nil
}
