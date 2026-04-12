// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

func (m *MainModel) init() tea.Cmd {
	return nil
}

func (m *MainModel) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Debug log all messages
	if debugMode {
		log.Printf("[DEBUG] Update received: %T", msg)
	}

	// Handle ESC for sub-views before routing to sub-view Update blocks.
	// Sub-view models (VMCreate, VMEdit) don't handle ESC themselves;
	// it must be intercepted here to call returnFromSubView().
	// Note: VMRunning view does NOT allow ESC to return - backgrounding is disabled.
	// If file browser is active in the sub-view, let ESC pass through
	// so the form can close the browser instead of exiting the sub-view.
	if m.isSubViewActive() {
		if km, ok := msg.(tea.KeyMsg); ok && km.String() == "esc" {
			if !m.isFileBrowserActiveInSubView() {
				if m.currentView == ViewVMRunning && m.vmRunningModel != nil {
					// VMRunning view: never allow ESC to leave - backgrounding disabled
					if m.vmRunningModel != nil && m.vmRunningModel.Runner() != nil && m.vmRunningModel.Runner().IsRunning() {
						m.statusBar.SetMessage("VM is still running. Press 'q' to stop it.")
					} else {
						m.statusBar.SetMessage("Press 'q' to exit the VM view.")
					}
					return m, nil
				}
				return m.returnFromSubView()
			}
		}
	}

	// Handle view change messages from sub-models
	if vcm, ok := msg.(ViewChangeMsg); ok {
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

	// Handle VM created messages from sub-models
	if vcm, ok := msg.(VMCreatedMsg); ok {
		m.statusMessage = fmt.Sprintf("VM '%s' created successfully", vcm.VMName)
		m.currentView = ViewConfigMenu
		m.rebuildMenuList()
		m.breadcrumbs.Clear()
		return m, nil
	}

	// Handle VM updated messages from sub-models
	if vcm, ok := msg.(VMUpdatedMsg); ok {
		m.statusMessage = fmt.Sprintf("VM '%s' updated successfully", vcm.VMName)
		m.currentView = ViewConfigMenu
		m.rebuildMenuList()
		m.breadcrumbs.Clear()
		return m, nil
	}

	// Handle VM deleted messages from sub-models
	if vdm, ok := msg.(VMDeletedMsg); ok {
		m.statusMessage = fmt.Sprintf("VM '%s' deleted successfully", vdm.VMName)
		m.currentView = ViewConfigMenu
		m.rebuildMenuList()
		return m, nil
	}

	// Handle VM stopped messages from running model
	if vsm, ok := msg.(VMStoppedMsg); ok {
		m.statusBar.SetMessage(fmt.Sprintf("VM '%s' stopped: %s", vsm.VMName, vsm.Reason))
		return m, nil
	}

	// Handle exit request from VM running view
	if _, ok := msg.(VMRunningViewExitMsg); ok {
		return m.returnFromSubView()
	}

	// Handle LBU commit completion
	if lcm, ok := msg.(LBUCommitMsg); ok {
		if lcm.Success {
			m.statusBar.SetMessage("LBU commit: " + lcm.Output)
		} else {
			m.statusBar.SetMessage("LBU commit failed: " + lcm.Output)
		}
		return m, nil
	}

	// Handle reboot completion
	if rm, ok := msg.(RebootMsg); ok {
		if rm.Success {
			m.statusBar.SetMessage("Reboot: " + rm.Output)
		} else {
			m.statusBar.SetMessage("Reboot failed: " + rm.Output)
		}
		return m, nil
	}

	// Handle power off completion
	if pom, ok := msg.(PowerOffMsg); ok {
		if pom.Success {
			m.statusBar.SetMessage("Power off: " + pom.Output)
		} else {
			m.statusBar.SetMessage("Power off failed: " + pom.Output)
		}
		return m, nil
	}

	// If we're in VM creation view, delegate to that model
	if m.currentView == ViewVMCreate && m.vmCreateModel != nil {
		// Forward file/disk selection messages to sub-view
		if fsm, ok := msg.(FileSelectedMsg); ok {
			// Route through AddDiskModel when it exists (harddisk flow).
			// Note: addDiskModel may be inactive at this point because
			// handleFileSelected sets active=false during cmd execution
			// in a prior Update call. We still route through it to get
			// the DiskAddedMsg conversion.
			if m.vmCreateModel.form.addDiskModel != nil {
				inner, cmd := m.vmCreateModel.form.addDiskModel.Update(fsm)
				if adm, ok := inner.(*AddDiskModel); ok {
					m.vmCreateModel.form.addDiskModel = adm
				}
				if cmd != nil {
					return m.handleSubViewMsg(cmd())
				}
				return m, nil
			}
			// Otherwise handle directly (CDROM flow)
			inner, cmd := m.vmCreateModel.form.Update(fsm)
			if f, ok := inner.(*VMFormModel); ok {
				m.vmCreateModel.form = f
			}
			if cmd != nil {
				return m.handleSubViewMsg(cmd())
			}
			return m, nil
		}
		// Handle DiskAddedMsg directly — addDiskModel may already be
		// inactive/cleared by the time this message arrives.
		if dam, ok := msg.(DiskAddedMsg); ok {
			inner, cmd := m.vmCreateModel.form.Update(dam)
			if f, ok := inner.(*VMFormModel); ok {
				m.vmCreateModel.form = f
			}
			if cmd != nil {
				return m.handleSubViewMsg(cmd())
			}
			return m, nil
		}
		model, cmd := m.vmCreateModel.Update(msg)
		// Update the sub-model reference
		if vcm, ok := model.(*VMCreateModel); ok {
			m.vmCreateModel = vcm
		}
		// Execute the command to get any resulting messages
		if cmd != nil {
			nextMsg := cmd()
			return m.handleSubViewOutput(nextMsg)
		}
		return m, nil
	}

	// If we're in VM edit view, delegate to that model
	if m.currentView == ViewVMEdit && m.vmEditModel != nil {
		// Forward file/disk selection messages to sub-view
		if fsm, ok := msg.(FileSelectedMsg); ok {
			// Route through AddDiskModel when it exists (harddisk flow).
			// Note: addDiskModel may be inactive at this point because
			// handleFileSelected sets active=false during cmd execution
			// in a prior Update call. We still route through it to get
			// the DiskAddedMsg conversion.
			if m.vmEditModel.form.addDiskModel != nil {
				inner, cmd := m.vmEditModel.form.addDiskModel.Update(fsm)
				if adm, ok := inner.(*AddDiskModel); ok {
					m.vmEditModel.form.addDiskModel = adm
				}
				if cmd != nil {
					return m.handleSubViewMsg(cmd())
				}
				return m, nil
			}
			// Otherwise handle directly (CDROM flow)
			inner, cmd := m.vmEditModel.form.Update(fsm)
			if f, ok := inner.(*VMFormModel); ok {
				m.vmEditModel.form = f
			}
			if cmd != nil {
				return m.handleSubViewMsg(cmd())
			}
			return m, nil
		}
		// Handle DiskAddedMsg directly — addDiskModel may already be
		// inactive/cleared by the time this message arrives.
		if dam, ok := msg.(DiskAddedMsg); ok {
			inner, cmd := m.vmEditModel.form.Update(dam)
			if f, ok := inner.(*VMFormModel); ok {
				m.vmEditModel.form = f
			}
			if cmd != nil {
				return m.handleSubViewMsg(cmd())
			}
			return m, nil
		}
		model, cmd := m.vmEditModel.Update(msg)
		if vem, ok := model.(*VMEditModel); ok {
			m.vmEditModel = vem
		}
		if cmd != nil {
			nextMsg := cmd()
			return m.handleSubViewOutput(nextMsg)
		}
		return m, nil
	}

	// If we're in VM running view, delegate to that model
	if m.currentView == ViewVMRunning && m.vmRunningModel != nil {
		model, cmd := m.vmRunningModel.Update(msg)
		if vrm, ok := model.(*VMRunningModel); ok {
			m.vmRunningModel = vrm
		}
		return m, cmd
	}

	// If we're in VM delete view, delegate to that model
	if m.currentView == ViewVMDelete && m.vmDeleteModel != nil {
		model, cmd := m.vmDeleteModel.Update(msg)
		if vdm, ok := model.(*VMDeleteModel); ok {
			m.vmDeleteModel = vdm
		}
		if cmd != nil {
			nextMsg := cmd()
			if vcm, ok := nextMsg.(ViewChangeMsg); ok {
				m.currentView = vcm.View
				if vcm.View == ViewMainMenu {
					m.rebuildMenuList()
				}
				return m, nil
			}
			if vdm, ok := nextMsg.(VMDeletedMsg); ok {
				m.statusMessage = fmt.Sprintf("VM '%s' deleted successfully", vdm.VMName)
				m.currentView = ViewConfigMenu
				m.rebuildMenuList()
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		// Forward to active sub-view so viewport can resize
		m.forwardWindowSizeToSubView(msg)
		// List sizing is done in render methods based on contentHeight()
		return m, nil
	}

	return m, nil
}

// forwardWindowSizeToSubView sends the window size to the active sub-view
func (m *MainModel) forwardWindowSizeToSubView(msg tea.WindowSizeMsg) {
	contentH := m.contentHeight()
	switch m.currentView {
	case ViewVMCreate:
		if m.vmCreateModel != nil {
			m.vmCreateModel.form.SetSize(msg.Width-4, contentH-2)
		}
	case ViewVMEdit:
		if m.vmEditModel != nil {
			m.vmEditModel.form.SetSize(msg.Width-4, contentH-2)
		}
	case ViewCPUOptions:
		if m.cpuOptionsModel != nil {
			m.cpuOptionsModel.form.SetSize(msg.Width-4, contentH-2)
		}
	case ViewPCIPassthrough:
		if m.pciPassthroughModel != nil {
			m.pciPassthroughModel.form.SetSize(msg.Width-4, contentH-2)
		}
	case ViewUSBPassthrough:
		if m.usbPassthroughModel != nil {
			m.usbPassthroughModel.form.SetSize(msg.Width-4, contentH-2)
		}
	case ViewCPUTopology:
		if m.cpuTopologyModel != nil {
			m.cpuTopologyModel.form.SetSize(msg.Width-4, contentH-2)
		}
	case ViewVCPUPinning:
		if m.vcpuPinningModel != nil {
			m.vcpuPinningModel.form.SetSize(msg.Width-4, contentH-2)
		}
	case ViewSSHPassword:
		if m.sshPasswordModel != nil {
			m.sshPasswordModel.form.SetSize(msg.Width-4, contentH-2)
		}
	case ViewVMRunning:
		if m.vmRunningModel != nil {
			m.vmRunningModel.SetSize(msg.Width-4, contentH-2)
		}
	}
}

// handleKeyPress handles keyboard input
func (m *MainModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	// Global quit keys
	switch keyStr {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	}

	// If we're in a sub-view (create/edit/delete/select), ESC returns to parent tab
	if m.isSubViewActive() {
		if keyStr == "esc" {
			return m.returnFromSubView()
		}
		// Delegate other keys to sub-view (existing delegation code)
		return m.delegateToSubView(msg)
	}

	// Tab navigation (only when not in sub-view)
	if m.tabModel.HandleKeyInput(keyStr) {
		m.onTabChanged()
		return m, nil
	}

	// ESC at top level quits
	if keyStr == "esc" {
		m.quitting = true
		return m, tea.Quit
	}

	// Enter or Space key
	if keyStr == "enter" || keyStr == " " {
		return m.handleMenuSelection()
	}

	// Refresh
	if keyStr == "r" {
		m.rebuildMenuList()
		return m, nil
	}

	// Delegate to active tab's list
	var cmd tea.Cmd
	switch m.tabModel.GetActiveTab() {
	case components.TabVMs:
		m.menuList, cmd = m.menuList.Update(msg)
		m.selectedIndex = m.menuList.Index()
	case components.TabConfiguration:
		m.configList, cmd = m.configList.Update(msg)
		m.configSelectedIndex = m.configList.Index()
	case components.TabPower:
		m.powerList, cmd = m.powerList.Update(msg)
	}

	return m, cmd
}

// handleMenuSelection handles menu item selection
func (m *MainModel) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.tabModel.GetActiveTab() {
	case components.TabConfiguration:
		return m.handleConfigMenuSelection()
	case components.TabVMs:
		return m.handleVMSelection()
	case components.TabPower:
		return m.handlePowerSelection()
	}
	return m, nil
}

// handleConfigMenuSelection handles selections in the Configuration tab
func (m *MainModel) handleConfigMenuSelection() (tea.Model, tea.Cmd) {
	if debugMode {
		log.Printf("[DEBUG] handleConfigMenuSelection called, selectedIndex: %d", m.configSelectedIndex)
	}

	switch m.configSelectedIndex {
	case 0: // Add new VM
		m.vmCreateModel = NewVMCreateModel(m.vmManager)
		m.vmCreateModel.form.SetSize(m.windowWidth-4, m.contentHeight()-2)
		m.currentView = ViewVMCreate
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Configuration", "config", 1)
		m.breadcrumbs.AddItem("Add VM", "vm_create", 1)
		return m, nil
	case 1: // Edit VM
		return m.showVMSelection()
	case 2: // Delete VM
		return m.showVMSelectionForDeletion()
	case 3: // Edit CPU Topology
		cpuModel, err := NewCPUTopologyModel(m.vmManager)
		if err != nil {
			m.statusBar.SetMessage("Error loading CPU topology: " + err.Error())
			return m, nil
		}
		m.cpuTopologyModel = cpuModel
		m.cpuTopologyModel.form.SetSize(m.windowWidth-4, m.contentHeight()-2)
		m.currentView = ViewCPUTopology
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Configuration", "config", 1)
		m.breadcrumbs.AddItem("CPU Topology", "cpu_topology", 1)
		return m, nil
	case 4: // Edit vCPU Pinning
		vcpuModel, err := NewVCPUPinningModel(m.vmManager)
		if err != nil {
			m.statusBar.SetMessage("Error loading vCPU pinning: " + err.Error())
			return m, nil
		}
		m.vcpuPinningModel = vcpuModel
		m.vcpuPinningModel.form.SetSize(m.windowWidth-4, m.contentHeight()-2)
		m.currentView = ViewVCPUPinning
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Configuration", "config", 1)
		m.breadcrumbs.AddItem("vCPU Pinning", "vcpu_pinning", 1)
		return m, nil
	case 5: // Edit PCI Passthrough
		pciModel, err := NewPCIPassthroughModel(m.vmManager)
		if err != nil {
			m.statusBar.SetMessage("Error loading PCI passthrough: " + err.Error())
			return m, nil
		}
		m.pciPassthroughModel = pciModel
		m.pciPassthroughModel.form.SetSize(m.windowWidth-4, m.contentHeight()-2)
		m.currentView = ViewPCIPassthrough
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Configuration", "config", 1)
		m.breadcrumbs.AddItem("PCI Passthrough", "pci_passthrough", 1)
		return m, nil
	case 6: // Edit USB Passthrough
		usbModel, err := NewUSBPassthroughModel(m.vmManager)
		if err != nil {
			m.statusBar.SetMessage("Error loading USB passthrough: " + err.Error())
			return m, nil
		}
		m.usbPassthroughModel = usbModel
		m.usbPassthroughModel.form.SetSize(m.windowWidth-4, m.contentHeight()-2)
		m.currentView = ViewUSBPassthrough
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Configuration", "config", 1)
		m.breadcrumbs.AddItem("USB Passthrough", "usb_passthrough", 1)
		return m, nil
	case 7: // Edit Start/Stop Script
		form, err := NewStartStopScriptFormModel(m.vmManager)
		if err != nil {
			m.statusMessage = "Error loading start/stop script form: " + err.Error()
		} else {
			m.startStopScriptFormModel = form
			m.startStopScriptFormModel.SetSize(m.windowWidth-4, m.contentHeight()-2)
			m.currentView = ViewStartStopScript
			m.breadcrumbs.Clear()
			m.breadcrumbs.AddItem("Configuration", "config", 1)
			m.breadcrumbs.AddItem("Start/Stop Script", "start_stop_script", 1)
		}
		return m, nil
	case 8: // Edit CPU Options
		m.cpuOptionsModel = NewCPUOptionsModel(m.vmManager)
		m.cpuOptionsModel.form.SetSize(m.windowWidth-4, m.contentHeight()-2)
		m.currentView = ViewCPUOptions
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Configuration", "config", 1)
		m.breadcrumbs.AddItem("CPU Options", "cpu_options", 1)
		return m, nil
	case 9: // Set SSH Password
		m.sshPasswordModel = NewSSHPasswordModel()
		m.sshPasswordModel.form.SetSize(m.windowWidth-4, m.contentHeight()-2)
		m.currentView = ViewSSHPassword
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Configuration", "config", 1)
		m.breadcrumbs.AddItem("Set SSH Password", "ssh_password", 1)
		return m, nil
	case 10: // LBU COMMIT
		return m, runLBUCommit()
	}
	return m, nil
}

// handleVMSelection handles selections in the VMs tab
func (m *MainModel) handleVMSelection() (tea.Model, tea.Cmd) {
	if m.selectedIndex >= len(m.menuItems) {
		return m, nil
	}
	item := m.menuItems[m.selectedIndex]

	if debugMode {
		log.Printf("[DEBUG] handleVMSelection: %s (%s)", item.Title, item.Type)
	}

	switch item.Type {
	case "VM":
		// Check if a VM is already running
		if m.vmRunningModel != nil && m.vmRunningModel.Runner() != nil && m.vmRunningModel.Runner().IsRunning() {
			m.statusBar.SetMessage("A VM is already running. Stop it before starting another.")
			return m, nil
		}

		// Load VM config
		vmObj, err := m.vmManager.GetVM(item.VMID)
		if err != nil {
			m.statusBar.SetMessage("Error loading VM: " + err.Error())
			return m, nil
		}

		// Create runner
		runner := vm.NewVMRunner(vmObj, m.cfg)
		// Load and inject PCI passthrough config
		if pciCfg, err := m.vmManager.GetPCIPassthroughConfig(); err == nil {
			runner.SetPCIPassthroughConfig(pciCfg)
		}
		// Load and inject USB passthrough config
		if usbCfg, err := m.vmManager.GetUSBPassthroughConfig(); err == nil {
			runner.SetUSBPassthroughConfig(usbCfg)
		}
		// Load and inject CPU options for feature flags
		if cpuOpts, err := m.vmManager.GetCPUOptions(); err == nil {
			runner.SetCPUOptions(cpuOpts)
		}
		// Load and inject start/stop script config
		if scriptCfg, err := m.vmManager.GetStartStopScript(); err == nil {
			runner.SetStartStopScript(scriptCfg)
		}
		if err := runner.Start(); err != nil {
			m.statusBar.SetMessage("Error starting VM: " + err.Error())
			return m, nil
		}

		// Create running model
		m.vmRunningModel = NewVMRunningModel(vmObj, runner)
		m.vmRunningModel.SetSize(m.windowWidth-4, m.contentHeight()-2)
		m.vmRunningModel.startTime = time.Now()
		m.currentView = ViewVMRunning
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Start VM", "vms", 1)
		m.breadcrumbs.AddItem("Start", "vm_start", 1)
		m.breadcrumbs.AddItem(vmObj.Name, "vm_running", 1)

		if debugMode {
			log.Printf("[DEBUG] VM started: %s (ID: %s)", vmObj.Name, vmObj.ID)
		}

		return m, m.vmRunningModel.Init()
	}
	return m, nil
}

// handlePowerSelection handles selections in the Power tab
func (m *MainModel) handlePowerSelection() (tea.Model, tea.Cmd) {
	selectedIndex := m.powerList.Index()
	switch selectedIndex {
	case 0:
		// Reboot system
		return m, runReboot()
	case 1:
		// Power off system
		return m, runPowerOff()
	}
	return m, nil
}

// isSubViewActive returns true if a sub-view (create/edit/delete/select/running) is active
func (m *MainModel) isSubViewActive() bool {
	switch m.currentView {
	case ViewVMCreate, ViewVMEdit, ViewVMDelete, ViewVMSelect, ViewCPUOptions, ViewVMRunning, ViewPCIPassthrough, ViewUSBPassthrough, ViewCPUTopology, ViewVCPUPinning, ViewSSHPassword, ViewStartStopScript:
		return true
	}
	return false
}

// isFileBrowserActiveInSubView returns true if a file browser is active within the current sub-view
func (m *MainModel) isFileBrowserActiveInSubView() bool {
	switch m.currentView {
	case ViewVMCreate:
		if m.vmCreateModel != nil {
			return m.vmCreateModel.FileBrowserActive()
		}
	case ViewVMEdit:
		if m.vmEditModel != nil {
			return m.vmEditModel.FileBrowserActive()
		}
	case ViewStartStopScript:
		if m.startStopScriptFormModel != nil && m.startStopScriptFormModel.fileBrowser != nil && m.startStopScriptFormModel.fileBrowser.active {
			return true
		}
	}
	return false
}

// returnFromSubView handles ESC in a sub-view, returning to the parent tab
func (m *MainModel) returnFromSubView() (tea.Model, tea.Cmd) {
	prevView := m.currentView
	m.currentView = ViewMainMenu
	m.breadcrumbs.Clear()

	// Determine which tab to return to
	switch prevView {
	case ViewVMCreate, ViewVMEdit, ViewVMDelete, ViewVMSelect, ViewCPUOptions, ViewPCIPassthrough, ViewUSBPassthrough, ViewCPUTopology, ViewVCPUPinning, ViewSSHPassword, ViewStartStopScript:
		m.tabModel.SetActiveTab(components.TabConfiguration)
	case ViewVMRunning:
		m.vmRunningModel = nil
		m.tabModel.SetActiveTab(components.TabVMs)
	default:
		m.tabModel.SetActiveTab(components.TabVMs)
	}

	m.rebuildMenuList()
	return m, nil
}

// onTabChanged updates status bar and breadcrumbs when the tab changes.
// It also ensures the active tab's list cursor is properly positioned so
// keyboard input is immediately responsive without requiring an arrow key press.
func (m *MainModel) onTabChanged() {
	m.breadcrumbs.Clear()
	m.statusBar.SetMode("Ready")

	// Ensure the active tab's list has a properly positioned cursor
	// and sync the selected index so Enter works immediately.
	switch m.tabModel.GetActiveTab() {
	case components.TabVMs:
		m.menuList.Select(m.selectedIndex)
	case components.TabConfiguration:
		m.configList.Select(m.configSelectedIndex)
	case components.TabPower:
		m.powerList.Select(0)
	}
}

// delegateToSubView handles key events when a sub-view is active
func (m *MainModel) delegateToSubView(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.currentView {
	case ViewVMCreate:
		if m.vmCreateModel != nil {
			vmModel, vmCmd := m.vmCreateModel.Update(msg)
			if vmCreate, ok := vmModel.(*VMCreateModel); ok {
				m.vmCreateModel = vmCreate
			}
			// Execute the command to get any resulting messages
			if vmCmd != nil {
				nextMsg := vmCmd()
				return m.handleSubViewOutput(nextMsg)
			}
		}
	case ViewVMEdit:
		if m.vmEditModel != nil {
			vmModel, vmCmd := m.vmEditModel.Update(msg)
			if vmEdit, ok := vmModel.(*VMEditModel); ok {
				m.vmEditModel = vmEdit
			}
			if vmCmd != nil {
				nextMsg := vmCmd()
				return m.handleSubViewOutput(nextMsg)
			}
		}
	case ViewVMDelete:
		if m.vmDeleteModel != nil {
			vmModel, vmCmd := m.vmDeleteModel.Update(msg)
			if vmDelete, ok := vmModel.(*VMDeleteModel); ok {
				m.vmDeleteModel = vmDelete
			}
			if vmCmd != nil {
				nextMsg := vmCmd()
				return m.handleSubViewOutput(nextMsg)
			}
		}
	case ViewVMSelect:
		m.ensureVMSelectList()
		// Handle Enter or Space key to navigate to delete/edit confirmation
		if km, ok := msg.(tea.KeyMsg); ok && (km.String() == "enter" || km.String() == " ") {
			selectedIndex := m.vmSelectList.Index()
			if selectedIndex >= 0 && selectedIndex < len(m.vmListForSelection) {
				selectedVM := m.vmListForSelection[selectedIndex]
				if m.selectionMode == "delete" {
					if deleteModel, err := NewVMDeleteModel(m.vmManager, selectedVM.ID); err == nil {
						m.vmDeleteModel = deleteModel
						m.currentView = ViewVMDelete
						m.breadcrumbs.AddItem("Delete "+selectedVM.Name, "vm_delete_confirm", 1)
						return m, nil
					}
				} else if m.selectionMode == "edit" {
					if editModel, err := NewVMEditModel(m.vmManager, selectedVM.ID); err == nil {
						m.vmEditModel = editModel
						m.vmEditModel.form.SetSize(m.windowWidth-4, m.contentHeight()-2)
						m.currentView = ViewVMEdit
						m.breadcrumbs.AddItem("Edit "+selectedVM.Name, "vm_edit", 1)
						return m, nil
					}
				}
			}
		}
		m.vmSelectList, cmd = m.vmSelectList.Update(msg)
	case ViewCPUOptions:
		if m.cpuOptionsModel != nil {
			cpuModel, cpuCmd := m.cpuOptionsModel.Update(msg)
			if cpuOpts, ok := cpuModel.(*CPUOptionsModel); ok {
				m.cpuOptionsModel = cpuOpts
			}
			if cpuCmd != nil {
				nextMsg := cpuCmd()
				return m.handleSubViewOutput(nextMsg)
			}
		}
	case ViewPCIPassthrough:
		if m.pciPassthroughModel != nil {
			pciModel, pciCmd := m.pciPassthroughModel.Update(msg)
			if pciPassthrough, ok := pciModel.(*PCIPassthroughModel); ok {
				m.pciPassthroughModel = pciPassthrough
			}
			if pciCmd != nil {
				nextMsg := pciCmd()
				return m.handleSubViewOutput(nextMsg)
			}
		}
	case ViewUSBPassthrough:
		if m.usbPassthroughModel != nil {
			usbModel, usbCmd := m.usbPassthroughModel.Update(msg)
			if usbPassthrough, ok := usbModel.(*USBPassthroughModel); ok {
				m.usbPassthroughModel = usbPassthrough
			}
			if usbCmd != nil {
				nextMsg := usbCmd()
				return m.handleSubViewOutput(nextMsg)
			}
		}
	case ViewCPUTopology:
		if m.cpuTopologyModel != nil {
			cpuModel, cpuCmd := m.cpuTopologyModel.Update(msg)
			if cpuTopo, ok := cpuModel.(*CPUTopologyModel); ok {
				m.cpuTopologyModel = cpuTopo
			}
			if cpuCmd != nil {
				nextMsg := cpuCmd()
				return m.handleSubViewOutput(nextMsg)
			}
		}
	case ViewVCPUPinning:
		if m.vcpuPinningModel != nil {
			vcpuModel, vcpuCmd := m.vcpuPinningModel.Update(msg)
			if vcpuPinning, ok := vcpuModel.(*VCPUPinningModel); ok {
				m.vcpuPinningModel = vcpuPinning
			}
			if vcpuCmd != nil {
				nextMsg := vcpuCmd()
				return m.handleSubViewOutput(nextMsg)
			}
		}
	case ViewSSHPassword:
		if m.sshPasswordModel != nil {
			sshModel, sshCmd := m.sshPasswordModel.Update(msg)
			if sshPw, ok := sshModel.(*SSHPasswordModel); ok {
				m.sshPasswordModel = sshPw
			}
			if sshCmd != nil {
				nextMsg := sshCmd()
				return m.handleSubViewOutput(nextMsg)
			}
		}
	case ViewStartStopScript:
		// Handle keys directly in start/stop script form model
		_, cmd := m.StartStopScriptFormUpdate(msg)
		// Execute the command to initialize the file browser
		if cmd != nil {
			nextMsg := cmd()
			return m.handleSubViewOutput(nextMsg)
		}
		return m, nil
	case ViewVMRunning:
		if m.vmRunningModel != nil {
			vrm, vrmCmd := m.vmRunningModel.Update(msg)
			if runningModel, ok := vrm.(*VMRunningModel); ok {
				m.vmRunningModel = runningModel
			}
			return m, vrmCmd
		}
	}

	return m, cmd
}

// handleSubViewOutput processes messages produced by sub-view command execution
// (from delegateToSubView). Unlike handleSubViewMsg, this method properly routes
// FileSelectedMsg through addDiskModel to convert it to DiskAddedMsg.
func (m *MainModel) handleSubViewOutput(nextMsg tea.Msg) (tea.Model, tea.Cmd) {
	switch nextMsg.(type) {
	case DirectoryLoadedMsg:
		// Directory loaded in file browser - refresh the form view
		if m.startStopScriptFormModel != nil {
			m.startStopScriptFormModel.syncViewport()
		}
		return m, nil
	case FileSelectedMsg:
		// Route FileSelectedMsg through addDiskModel when it exists
		// (hard disk flow). addDiskModel converts it to DiskAddedMsg.
		// For CDROMs (no addDiskModel), route through form directly.
		// For Start/Stop Script form, route through the form model directly.
		switch m.currentView {
		case ViewVMCreate:
			if m.vmCreateModel != nil {
				if m.vmCreateModel.form.addDiskModel != nil {
					inner, cmd := m.vmCreateModel.form.addDiskModel.Update(nextMsg)
					if adm, ok := inner.(*AddDiskModel); ok {
						m.vmCreateModel.form.addDiskModel = adm
					}
					if cmd != nil {
						return m.handleSubViewOutput(cmd())
					}
					return m, nil
				}
				inner, cmd := m.vmCreateModel.form.Update(nextMsg)
				if f, ok := inner.(*VMFormModel); ok {
					m.vmCreateModel.form = f
				}
				if cmd != nil {
					return m.handleSubViewOutput(cmd())
				}
			}
		case ViewVMEdit:
			if m.vmEditModel != nil {
				if m.vmEditModel.form.addDiskModel != nil {
					inner, cmd := m.vmEditModel.form.addDiskModel.Update(nextMsg)
					if adm, ok := inner.(*AddDiskModel); ok {
						m.vmEditModel.form.addDiskModel = adm
					}
					if cmd != nil {
						return m.handleSubViewOutput(cmd())
					}
					return m, nil
				}
				inner, cmd := m.vmEditModel.form.Update(nextMsg)
				if f, ok := inner.(*VMFormModel); ok {
					m.vmEditModel.form = f
				}
				if cmd != nil {
					return m.handleSubViewOutput(cmd())
				}
			}
		case ViewStartStopScript:
			if m.startStopScriptFormModel != nil {
				inner, cmd := m.startStopScriptFormModel.Update(nextMsg)
				if f, ok := inner.(*StartStopScriptFormModel); ok {
					m.startStopScriptFormModel = f
				}
				if cmd != nil {
					return m.handleSubViewOutput(cmd())
				}
			}
		}
		return m, nil
	}

	// For all other messages (DiskAddedMsg, VMUpdatedMsg, etc.),
	// delegate to handleSubViewMsg
	return m.handleSubViewMsg(nextMsg)
}


