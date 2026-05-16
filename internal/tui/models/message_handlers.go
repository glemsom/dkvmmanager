// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"log"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
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

// handleVMStoppedMsg handles VMStoppedMsg from running model
func (m *MainModel) handleVMStoppedMsg(vsm VMStoppedMsg) (tea.Model, tea.Cmd) {
	m.statusBar.SetMessage(fmt.Sprintf("VM '%s' stopped: %s", vsm.VMName, vsm.Reason))
	// Return to main menu when VM stops
	m.currentView = ViewMainMenu
	m.rebuildMenuList()
	m.breadcrumbs.Clear()
	return m, nil
}

// MessageHandler is a function that processes a message and returns the model/cmd.
type MessageHandler func(m *MainModel, msg tea.Msg) (tea.Model, tea.Cmd)

// messageHandlers holds registered handlers for specific message types.
var messageHandlers = make(map[string]MessageHandler)

// RegisterMessageHandler registers a handler for a message type (by name).
// Package init functions should call this for their message types.
func RegisterMessageHandler(msgName string, handler MessageHandler) {
	messageHandlers[msgName] = handler
}

// init registers all message handlers at package initialization time.
func init() {
	// Register handlers for messages that use the unified pattern
	RegisterMessageHandler("VMCreatedMsg", HandleVMCreatedMsg)
	RegisterMessageHandler("VMUpdatedMsg", HandleVMUpdatedMsg)
	RegisterMessageHandler("VMDeletedMsg", HandleVMDeletedMsg)
	RegisterMessageHandler("PCIVFIOKernelAppliedMsg", HandlePCIVFIOKernelAppliedMsg)
	RegisterMessageHandler("VCPUCPUKernelAppliedMsg", HandleVCPUCPUKernelAppliedMsg)
	RegisterMessageHandler("LVCreateUpdatedMsg", HandleLVCreateUpdatedMsg)
	RegisterMessageHandler("LBUCommitMsg", HandleLBUCommitMsg)
	RegisterMessageHandler("RebootMsg", HandleRebootMsg)
	RegisterMessageHandler("PowerOffMsg", HandlePowerOffMsg)
}

// HandleVMCreatedMsg handles VMCreatedMsg using unified pattern.
func HandleVMCreatedMsg(m *MainModel, msg tea.Msg) (tea.Model, tea.Cmd) {
	vcm := msg.(VMCreatedMsg)
	m.rebuildMenuList()
	return m.UnifiedViewReturn(fmt.Sprintf("VM '%s' created successfully", vcm.VMName))
}

// HandleVMUpdatedMsg handles VMUpdatedMsg using unified pattern.
func HandleVMUpdatedMsg(m *MainModel, msg tea.Msg) (tea.Model, tea.Cmd) {
	vcm := msg.(VMUpdatedMsg)
	m.rebuildMenuList()
	return m.UnifiedViewReturn(fmt.Sprintf("VM '%s' updated successfully", vcm.VMName))
}

// HandleVMDeletedMsg handles VMDeletedMsg using unified pattern.
func HandleVMDeletedMsg(m *MainModel, msg tea.Msg) (tea.Model, tea.Cmd) {
	vdm := msg.(VMDeletedMsg)
	m.rebuildMenuList()
	return m.UnifiedViewReturn(fmt.Sprintf("VM '%s' deleted successfully", vdm.VMName))
}

// HandlePCIVFIOKernelAppliedMsg handles PCIVFIOKernelAppliedMsg with conditional status.
func HandlePCIVFIOKernelAppliedMsg(m *MainModel, msg tea.Msg) (tea.Model, tea.Cmd) {
	kam := msg.(PCIVFIOKernelAppliedMsg)
	var status string
	if kam.Success {
		status = "vfio-pci.ids applied to kernel successfully"
	} else {
		status = "Apply to kernel failed: " + kam.Error
	}
	return m.UnifiedViewReturn(status)
}

// HandleVCPUCPUKernelAppliedMsg handles VCPUCPUKernelAppliedMsg with conditional status.
func HandleVCPUCPUKernelAppliedMsg(m *MainModel, msg tea.Msg) (tea.Model, tea.Cmd) {
	kam := msg.(VCPUCPUKernelAppliedMsg)
	var status string
	if kam.Success {
		status = "CPU isolation parameters applied to kernel successfully"
	} else {
		status = "Apply to kernel failed: " + kam.Error
	}
	return m.UnifiedViewReturn(status)
}

// HandleLVCreateUpdatedMsg handles LVCreateUpdatedMsg with optional preview status.
func HandleLVCreateUpdatedMsg(m *MainModel, msg tea.Msg) (tea.Model, tea.Cmd) {
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

// HandleLBUCommitMsg handles LBUCommitMsg completion (no view change).
func HandleLBUCommitMsg(m *MainModel, msg tea.Msg) (tea.Model, tea.Cmd) {
	lcm := msg.(LBUCommitMsg)
	var status string
	if lcm.Success {
		status = "LBU commit: " + lcm.Output
	} else {
		status = "LBU commit failed: " + lcm.Output
	}
	m.statusBar.SetMessage(status)
	return m, nil
}

// HandleRebootMsg handles RebootMsg completion (no view change).
func HandleRebootMsg(m *MainModel, msg tea.Msg) (tea.Model, tea.Cmd) {
	rm := msg.(RebootMsg)
	var status string
	if rm.Success {
		status = "Reboot: " + rm.Output
	} else {
		status = "Reboot failed: " + rm.Output
	}
	m.statusBar.SetMessage(status)
	return m, nil
}

// HandlePowerOffMsg handles PowerOffMsg completion (no view change).
func HandlePowerOffMsg(m *MainModel, msg tea.Msg) (tea.Model, tea.Cmd) {
	pom := msg.(PowerOffMsg)
	var status string
	if pom.Success {
		status = "Power off: " + pom.Output
	} else {
		status = "Power off failed: " + pom.Output
	}
	m.statusBar.SetMessage(status)
	return m, nil
}

// handleSubViewMsg processes messages from sub-view command execution.
// It dispatches to registered handlers first, then handles special cases.
func (m *MainModel) handleSubViewMsg(nextMsg tea.Msg) (tea.Model, tea.Cmd) {
	// Note: FileSelectedMsg and DiskAddedMsg are handled by the VMFormModel's
	// HandleMessage method, which is delegated to by ScrollableForm.Update.
	// No manual routing needed here.

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

	// Check registry for dynamic handler using type name
	// Handle both pointer and non-pointer message types
	msgType := reflect.TypeOf(nextMsg)
	typeName := msgType.Name()
	if typeName == "" && msgType.Kind() == reflect.Ptr {
		typeName = msgType.Elem().Name()
	}
	if handler, ok := messageHandlers[typeName]; ok {
		return handler(m, nextMsg)
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
