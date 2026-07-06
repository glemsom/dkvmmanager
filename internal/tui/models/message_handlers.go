// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// UnifiedViewReturn handles the common pattern when returning from a sub-view:
// 1. Optionally set status bar message
// 2. Change view to ViewConfigMenu
// 3. Clear breadcrumbs
// 4. Deactivate registry if active
func (m *MainModel) UnifiedViewReturn(statusMessage string) (tea.Model, tea.Cmd) {
	if statusMessage != "" {
		m.statusBar.SetMessage(statusMessage)
	}
	m.currentView = ViewConfigMenu
	m.breadcrumbs.Clear()
	if m.viewRegistry != nil && m.viewRegistry.IsActive() {
		m.viewRegistry.Deactivate()
	}
	return m, nil
}

// handleViewChangeMsg handles ViewChangeMsg from sub-models
func (m *MainModel) handleViewChangeMsg(vcm ViewChangeMsg) (tea.Model, tea.Cmd) {
	if m.debugMode {
		log.Printf("[DEBUG] View change: %s -> %s", m.currentView, vcm.View)
	}
	m.currentView = vcm.View
	if vcm.View == ViewMainMenu {
		// Refresh menu items when returning from VM creation
		m.rebuildMenuList()
	}
	return m, nil
}

// handleVMStoppedMsg handles VMStoppedMsg from running model
func (m *MainModel) handleVMStoppedMsg(vsm VMStoppedMsg) (tea.Model, tea.Cmd) {
	m.statusBar.SetMessage(fmt.Sprintf("VM '%s' stopped: %s", vsm.VMName, vsm.Reason))
	// Return to main menu when VM stops
	m.currentView = ViewMainMenu
	m.rebuildMenuList()
	m.breadcrumbs.Clear()
	return m, nil
}

// handleVMSelected handles VMSelectedMsg from VMSelectModel.
// Creates the appropriate edit or delete model and activates it via registry.
func (m *MainModel) handleVMSelected(msg VMSelectedMsg) (tea.Model, tea.Cmd) {
	switch msg.Mode {
	case "edit":
		if editModel, err := NewVMEditModel(m.vmManager, msg.VMID); err == nil {
			editModel.SetSize(m.windowWidth-4, m.contentHeight()-2)
			if m.viewRegistry != nil && m.viewRegistry.GetDef(ViewVMEdit) != nil {
				m.viewRegistry.SetActiveModel(m.viewRegistry.GetDef(ViewVMEdit), editModel)
			}
			m.currentView = ViewVMEdit
			m.breadcrumbs.Clear()
			vmObj, _ := m.vmManager.GetVM(msg.VMID)
			if vmObj != nil {
				m.breadcrumbs.AddItem("Edit "+vmObj.Name, "vm_edit", 1)
			}
			return m, editModel.Init()
		}
		m.statusBar.SetMessage("Error creating edit form")
		return m, nil

	case "delete":
		if deleteModel, err := NewVMDeleteModel(m.vmManager, msg.VMID, m.debugMode); err == nil {
			deleteModel.SetSize(m.windowWidth-4, m.contentHeight()-2)
			if m.viewRegistry != nil && m.viewRegistry.GetDef(ViewVMDelete) != nil {
				m.viewRegistry.SetActiveModel(m.viewRegistry.GetDef(ViewVMDelete), deleteModel)
			}
			m.currentView = ViewVMDelete
			m.breadcrumbs.Clear()
			vmObj, _ := m.vmManager.GetVM(msg.VMID)
			if vmObj != nil {
				m.breadcrumbs.AddItem("Delete "+vmObj.Name, "vm_delete_confirm", 1)
			}
			return m, deleteModel.Init()
		}
		m.statusBar.SetMessage("Error creating delete form")
		return m, nil
	}

	return m, nil
}

// handleVMCreated handles VMCreatedMsg.
func (m *MainModel) handleVMCreated(msg VMCreatedMsg) (tea.Model, tea.Cmd) {
	m.rebuildMenuList()
	return m.UnifiedViewReturn(fmt.Sprintf("VM '%s' created successfully", msg.VMName))
}

// handleVMUpdated handles VMUpdatedMsg.
func (m *MainModel) handleVMUpdated(msg VMUpdatedMsg) (tea.Model, tea.Cmd) {
	m.rebuildMenuList()
	return m.UnifiedViewReturn(fmt.Sprintf("VM '%s' updated successfully", msg.VMName))
}

// handleVMDeleted handles VMDeletedMsg.
func (m *MainModel) handleVMDeleted(msg VMDeletedMsg) (tea.Model, tea.Cmd) {
	m.rebuildMenuList()
	return m.UnifiedViewReturn(fmt.Sprintf("VM '%s' deleted successfully", msg.VMName))
}

// handlePCIVFIOKernelApplied handles PCIVFIOKernelAppliedMsg.
func (m *MainModel) handlePCIVFIOKernelApplied(msg PCIVFIOKernelAppliedMsg) (tea.Model, tea.Cmd) {
	var status string
	if msg.Success {
		status = "vfio-pci.ids applied to kernel successfully"
	} else {
		status = "Apply to kernel failed: " + msg.Error
	}
	return m.UnifiedViewReturn(status)
}

// handleVCPUCPUKernelApplied handles VCPUCPUKernelAppliedMsg.
func (m *MainModel) handleVCPUCPUKernelApplied(msg VCPUCPUKernelAppliedMsg) (tea.Model, tea.Cmd) {
	var status string
	if msg.Success {
		status = "CPU isolation parameters applied to kernel successfully"
	} else {
		status = "Apply to kernel failed: " + msg.Error
	}
	return m.UnifiedViewReturn(status)
}

// handleLVCreateUpdated handles LVCreateUpdatedMsg.
func (m *MainModel) handleLVCreateUpdated(msg LVCreateUpdatedMsg) (tea.Model, tea.Cmd) {
	// Get status from the active registry model
	if m.viewRegistry != nil && m.viewRegistry.ActiveName() == ViewLVCreate {
		if svm, ok := m.viewRegistry.ActiveModel().(*LVCreateModel); ok {
			fm := svm.Form().Model().(*LVCreateFormModel)
			if fm.Preview() != "" {
				return m.UnifiedViewReturn(fm.Preview())
			}
		}
	}
	return m.UnifiedViewReturn("Logical volume created successfully")
}

// handleLBUCommit handles LBUCommitMsg.
func (m *MainModel) handleLBUCommit(msg LBUCommitMsg) (tea.Model, tea.Cmd) {
	var status string
	if msg.Success {
		status = "LBU commit: " + msg.Output
	} else {
		status = "LBU commit failed: " + msg.Output
	}
	m.statusBar.SetMessage(status)
	return m, nil
}

// handleReboot handles RebootMsg.
func (m *MainModel) handleReboot(msg RebootMsg) (tea.Model, tea.Cmd) {
	var status string
	if msg.Success {
		status = "Reboot: " + msg.Output
	} else {
		status = "Reboot failed: " + msg.Output
	}
	m.statusBar.SetMessage(status)
	return m, nil
}

// handlePowerOff handles PowerOffMsg.
func (m *MainModel) handlePowerOff(msg PowerOffMsg) (tea.Model, tea.Cmd) {
	var status string
	if msg.Success {
		status = "Power off: " + msg.Output
	} else {
		status = "Power off failed: " + msg.Output
	}
	m.statusBar.SetMessage(status)
	return m, nil
}

// handleSubViewMsg processes messages from sub-view command execution.
func (m *MainModel) handleSubViewMsg(nextMsg tea.Msg) (tea.Model, tea.Cmd) {
	// FileSelectedMsg and DiskAddedMsg are handled by the VMFormModel's
	// HandleMessage method via the MessageHandler interface.
	// See form/types.go for the interface definition.

	// Handle ViewChangeMsg specially (not in registry)
	if vcm, ok := nextMsg.(ViewChangeMsg); ok {
		m.currentView = vcm.View
		// Deactivate the registry when leaving a sub-view
		if m.viewRegistry != nil && m.viewRegistry.IsActive() {
			m.viewRegistry.Deactivate()
		}
		if vcm.View == ViewMainMenu {
			m.rebuildMenuList()
			m.breadcrumbs.Clear()
		}
		return m, nil
	}

	// Handle lvVGsLoadedMsg specially (no-op - handled by active view)
	if _, ok := nextMsg.(lvVGsLoadedMsg); ok {
		return m, nil
	}

	// Handle lvCreateErrorMsg specially (error messages don't change view)
	if em, ok := nextMsg.(lvCreateErrorMsg); ok {
		m.statusBar.SetMessage("LV create failed: " + em.err)
		return m, nil
	}

	// Route DiskAddedMsg to the active form via HandleMessage
	if dam, ok := nextMsg.(DiskAddedMsg); ok {
		if m.viewRegistry != nil && m.viewRegistry.IsActive() {
			// Try direct HandleMessage on the active model
			if hm, ok := m.viewRegistry.ActiveModel().(interface{ HandleMessage(tea.Msg) tea.Cmd }); ok {
				_ = hm.HandleMessage(dam)
				return m, nil
			}
			// VMEditModel/VMCreateModel wrap ScrollableForm which wraps VMFormModel
			// ScrollableForm.Model() returns VMFormModel which implements HandleMessage
			if getter, ok := m.viewRegistry.ActiveModel().(interface{ Model() form.FormModel }); ok {
				if hm, ok := getter.Model().(interface{ HandleMessage(tea.Msg) tea.Cmd }); ok {
					_ = hm.HandleMessage(dam)
				}
			}
		}
		return m, nil
	}

	// Route FileSelectedMsg to the active form via HandleMessage
	if fsm, ok := nextMsg.(FileSelectedMsg); ok {
		if m.viewRegistry != nil && m.viewRegistry.IsActive() {
			// Try direct HandleMessage on the active model
			if hm, ok := m.viewRegistry.ActiveModel().(interface{ HandleMessage(tea.Msg) tea.Cmd }); ok {
				_ = hm.HandleMessage(fsm)
				return m, nil
			}
			// VMEditModel/VMCreateModel wrap ScrollableForm which wraps VMFormModel
			// ScrollableForm.Model() returns VMFormModel which implements HandleMessage
			if getter, ok := m.viewRegistry.ActiveModel().(interface{ Model() form.FormModel }); ok {
				if hm, ok := getter.Model().(interface{ HandleMessage(tea.Msg) tea.Cmd }); ok {
					_ = hm.HandleMessage(fsm)
				}
			}
		}
		return m, nil
	}

	// Handle unified message types from sub-view command execution
	switch msg := nextMsg.(type) {
	case VMCreatedMsg:
		return m.handleVMCreated(msg)
	case VMUpdatedMsg:
		return m.handleVMUpdated(msg)
	case VMDeletedMsg:
		return m.handleVMDeleted(msg)
	case PCIVFIOKernelAppliedMsg:
		return m.handlePCIVFIOKernelApplied(msg)
	case VCPUCPUKernelAppliedMsg:
		return m.handleVCPUCPUKernelApplied(msg)
	case LVCreateUpdatedMsg:
		return m.handleLVCreateUpdated(msg)
	case LBUCommitMsg:
		return m.handleLBUCommit(msg)
	case RebootMsg:
		return m.handleReboot(msg)
	case PowerOffMsg:
		return m.handlePowerOff(msg)
	}

	// Special handling for interface-based messages (form.FormSavedMsg)
	if form.IsFormSavedMsg(nextMsg) {
		if status := form.FormSavedStatus(nextMsg); status != "" {
			m.statusBar.SetMessage(status)
		}
		m.currentView = ViewConfigMenu
		m.breadcrumbs.Clear()
		if m.viewRegistry != nil && m.viewRegistry.IsActive() {
			m.viewRegistry.Deactivate()
		}
		return m, nil
	}

	return m, nil
}
