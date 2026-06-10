// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"log"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
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
		if km, ok := msg.(tea.KeyPressMsg); ok && km.String() == "esc" {
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
		// Deactivate the registry when leaving a sub-view
		if m.viewRegistry != nil && m.viewRegistry.IsActive() {
			m.viewRegistry.Deactivate()
		}
		if vcm.View == ViewMainMenu {
			// Refresh menu items when returning from VM creation
			m.rebuildMenuList()
		}
		return m, nil
	}

	// Handle VM created messages from sub-models (delegated to handleSubViewMsg)
	// Note: VMCreatedMsg is handled by handleSubViewMsg through the registry

	// Handle VM stopped messages from running model
	if vsm, ok := msg.(VMStoppedMsg); ok {
		m.statusBar.SetMessage(fmt.Sprintf("VM '%s' stopped: %s", vsm.VMName, vsm.Reason))
		// Clear running VM ID and legacy field
		m.runningVMID = ""
		m.vmRunningModel = nil
		// Deactivate registry entry for VMRunning
		if m.viewRegistry != nil && m.viewRegistry.ActiveName() == ViewVMRunning {
			m.viewRegistry.Deactivate()
		}
		// Return to main menu when VM stops
		m.currentView = ViewMainMenu
		m.rebuildMenuList()
		m.breadcrumbs.Clear()
		return m, nil
	}

	// Handle VM start failure from async start command
	if vse, ok := msg.(VMStartErrorMsg); ok {
		m.statusBar.SetMessage("Error starting VM: " + vse.Err.Error())
		m.runningVMID = ""
		m.vmRunningModel = nil
		// Deactivate registry entry for VMRunning
		if m.viewRegistry != nil && m.viewRegistry.ActiveName() == ViewVMRunning {
			m.viewRegistry.Deactivate()
		}
		m.currentView = ViewMainMenu
		m.rebuildMenuList()
		m.breadcrumbs.Clear()
		return m, nil
	}

	// Handle VM started (runner is now available, start log/status/exit watchers)
	if vsm, ok := msg.(VMStartedMsg); ok {
		if m.vmRunningModel != nil {
			m.vmRunningModel.runner = vsm.Runner
			m.vmRunningModel.status = "starting" // will be updated by initialStatus
			m.vmRunningModel.pollingSince = time.Now()
			return m, tea.Batch(
				m.vmRunningModel.waitForLog(),
				m.vmRunningModel.waitForVMExit(),
				m.vmRunningModel.pollStatus(),
				m.vmRunningModel.initialStatus(),
			)
		}
		return m, nil
	}

	// Handle exit request from VM running view
	if _, ok := msg.(VMRunningViewExitMsg); ok {
		return m.returnFromSubView()
	}

	// Handle sub-view completion messages directly (before registry dispatch)
	// so they always route to the main model regardless of active view.
	switch msg := msg.(type) {
	case VMCreatedMsg:
		return HandleVMCreatedMsg(m, msg)
	case VMUpdatedMsg:
		return HandleVMUpdatedMsg(m, msg)
	case VMDeletedMsg:
		return HandleVMDeletedMsg(m, msg)
	case LBUCommitMsg:
		return HandleLBUCommitMsg(m, msg)
	case RebootMsg:
		return HandleRebootMsg(m, msg)
	case PowerOffMsg:
		return HandlePowerOffMsg(m, msg)
	case PCIVFIOKernelAppliedMsg:
		return HandlePCIVFIOKernelAppliedMsg(m, msg)
	case VCPUCPUKernelAppliedMsg:
		return HandleVCPUCPUKernelAppliedMsg(m, msg)
	case LVCreateUpdatedMsg:
		return HandleLVCreateUpdatedMsg(m, msg)
	}

	// VMRunning-specific messages (polling/log) must bypass the registry dispatch
	// because the registry calls cmd() synchronously and feeds the result through
	// handleSubViewMsg, which breaks the command chain for Tick-based polling and
	// blocking channel reads. Route them directly to the fallback path instead.
	if m.currentView == ViewVMRunning && m.vmRunningModel != nil {
		switch msg.(type) {
		case VMStatusUpdateMsg, VMLogMsg, VMMetricsUpdateMsg:
			model, cmd := m.vmRunningModel.Update(msg)
			if vrm, ok := model.(*VMRunningModel); ok {
				m.vmRunningModel = vrm
			}
			return m, cmd
		}
	}

	// Registry-based dispatch
	// This ensures async messages (e.g. lvVGsLoadedMsg) reach registered views.
	if m.viewRegistry != nil && m.viewRegistry.IsActive() {
		activeModel := m.viewRegistry.ActiveModel()
		if activeModel != nil {
			newModel, cmd := activeModel.Update(msg)
			if svm, ok := newModel.(SubViewModel); ok {
				m.viewRegistry.Active().Model = svm
			} else {
				return newModel, cmd
			}
			if cmd != nil {
				nextMsg := cmd()
				return m.handleSubViewOutput(nextMsg)
			}
			return m, nil
		}
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
			if _, ok := nextMsg.(VMDeletedMsg); ok {
				// Delegate to handleSubViewMsg for consistent handling
				return m.handleSubViewMsg(nextMsg)
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
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
	// Registry-based size forwarding
	if m.viewRegistry != nil && m.viewRegistry.IsActive() {
		m.viewRegistry.ActiveModel().SetSize(msg.Width-4, m.contentHeight()-2)
		return
	}

	// Fallback: resize VMRunning view (registry handles all other views)
	if m.currentView == ViewVMRunning && m.vmRunningModel != nil {
		m.vmRunningModel.SetSize(msg.Width-4, m.contentHeight()-2)
	}
}

// handleKeyPress handles keyboard input
func (m *MainModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	keyStr := msg.String()

	// Global quit keys - check if a VM is running first
	switch keyStr {
	case "ctrl+c", "q":
		// Check if a VM is running and warn user
		if m.vmRunningModel != nil && m.vmRunningModel.Runner() != nil && m.vmRunningModel.Runner().IsRunning() {
			m.statusBar.SetMessage("A VM is running. Press 'q' in the VM view to stop it first.")
			return m, nil
		}
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
	if keyStr == "enter" || keyStr == "space" {
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

	// Try registry-based activation for form views
	if m.viewRegistry != nil {
		if sub, err := m.viewRegistry.ActivateByConfigIndex(m.configSelectedIndex, m); err == nil {
			sub.SetSize(m.windowWidth-4, m.contentHeight()-2)
			def := m.viewRegistry.ActiveDef()
			m.currentView = def.Name
			m.breadcrumbs.Clear()
			m.breadcrumbs.AddItem("Configuration", "config", 1)
			m.breadcrumbs.AddItem(def.BreadcrumbLabel, def.Name, 1)
			return m, sub.Init()
		}
	}

	switch m.configSelectedIndex {
	case 1: // Edit VM (not in registry — requires VM selection first)
		return m.showVMSelection()
	case 2: // Delete VM (not in registry — requires VM selection first)
		return m.showVMSelectionForDeletion()
	}

	// "Save changes" is always the last item in the config list
	if len(m.configList.Items()) > 0 && m.configSelectedIndex == len(m.configList.Items())-1 {
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

		// Load aggregated RunConfig from repo and host discovery
		runCfg := vm.LoadRunConfigFromRepo(m.configRepo)
		// Override HostCPUTopology with the model's discovery (supports testing mocks)
		if hostTopo, err := m.hostDiscovery.ScanCPUTopology(); err == nil {
			runCfg.HostCPUTopology = hostTopo
		}
		// Create runner with RunConfig (replaces individual Set* calls)
		runner := vm.NewVMRunner(vmObj, m.cfg, runCfg)

		// Create running model immediately (runner will be set async)
		vmRunningModel := NewVMRunningModel(vmObj, nil) // nil runner
		vmRunningModel.SetSize(m.windowWidth-4, m.contentHeight()-2)
		m.vmRunningModel = vmRunningModel
		m.currentView = ViewVMRunning
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Start VM", "vms", 1)
		m.breadcrumbs.AddItem("Start", "vm_start", 1)
		m.breadcrumbs.AddItem(vmObj.Name, "vm_running", 1)

		// Register as active view in the registry
		if m.viewRegistry != nil {
			m.viewRegistry.SetActiveModel(m.viewRegistry.GetDef(ViewVMRunning), vmRunningModel)
		}

		// Optimistically update status bar
		m.statusBar.SetStats(len(m.menuItems), 1)
		m.runningVMID = vmObj.ID

		// Start VM asynchronously — view is already visible
		if debugMode {
			log.Printf("[DEBUG] VM starting async: %s (ID: %s)", vmObj.Name, vmObj.ID)
		}

		return m, tea.Batch(
			m.vmRunningModel.Init(),               // polls with nil runner → shows "[STARTING]"
			startVMCommand(runner, vmObj.Name, vmObj.ID),
		)
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

// isSubViewActive returns true if a sub-view is active
func (m *MainModel) isSubViewActive() bool {
	if m.viewRegistry != nil && m.viewRegistry.IsActive() {
		return true
	}
	switch m.currentView {
	case ViewVMDelete, ViewVMSelect, ViewVMRunning:
		return true
	}
	return false
}

// isFileBrowserActiveInSubView returns true if a file browser is active within the current sub-view
func (m *MainModel) isFileBrowserActiveInSubView() bool {
	if m.viewRegistry != nil && m.viewRegistry.IsActive() {
		return m.viewRegistry.ActiveModel().FileBrowserActive()
	}
	return false
}

// returnFromSubView handles ESC in a sub-view, returning to the parent tab
func (m *MainModel) returnFromSubView() (tea.Model, tea.Cmd) {
	m.currentView = ViewMainMenu
	m.breadcrumbs.Clear()

	// Registry-based: read parent tab from active view
	if m.viewRegistry != nil && m.viewRegistry.ActiveDef() != nil {
		def := m.viewRegistry.ActiveDef()
		m.tabModel.SetActiveTab(def.ParentTab)
		// VMRunning is special: keep the model if VM is still running
		// so status updates continue to arrive. Only deactivate if stopped.
		if def.Name == ViewVMRunning {
			if m.vmRunningModel != nil && m.vmRunningModel.Runner() != nil && m.vmRunningModel.Runner().IsRunning() {
				if m.runningVMID == "" {
					m.runningVMID = m.vmRunningModel.Runner().VM().ID
				}
			} else {
				m.vmRunningModel = nil
				m.runningVMID = ""
				m.viewRegistry.Deactivate()
			}
		} else {
			m.viewRegistry.Deactivate()
		}
	} else {
		// VMSelect: returns to Configuration tab
		m.tabModel.SetActiveTab(components.TabConfiguration)
		// Clear VMSelect state
		m.vmSelectList = list.Model{}
		m.vmListForSelection = nil
		m.selectionMode = ""
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

	// Try registry-based dispatch first
	if m.viewRegistry != nil && m.viewRegistry.IsActive() {
		model := m.viewRegistry.ActiveModel()
		if model != nil {
			newModel, modelCmd := model.Update(msg)
			if svm, ok := newModel.(SubViewModel); ok {
				m.viewRegistry.Active().Model = svm
			} else {
				// Sub-model returned a different tea.Model (e.g. during exit)
				return newModel, modelCmd
			}
			if modelCmd != nil {
				nextMsg := modelCmd()
				return m.handleSubViewOutput(nextMsg)
			}
			return m, nil
		}
	}

	switch m.currentView {
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
		if km, ok := msg.(tea.KeyPressMsg); ok && (km.String() == "enter" || km.String() == "space") {
			selectedIndex := m.vmSelectList.Index()
			if selectedIndex >= 0 && selectedIndex < len(m.vmListForSelection) {
				selectedVM := m.vmListForSelection[selectedIndex]
				if m.selectionMode == "delete" {
					if deleteModel, err := NewVMDeleteModel(m.vmManager, selectedVM.ID); err == nil {
						m.vmDeleteModel = deleteModel
						if m.viewRegistry != nil && m.viewRegistry.GetDef(ViewVMDelete) != nil {
							m.viewRegistry.SetActiveModel(m.viewRegistry.GetDef(ViewVMDelete), deleteModel)
						}
						m.currentView = ViewVMDelete
						m.breadcrumbs.AddItem("Delete "+selectedVM.Name, "vm_delete_confirm", 1)
						return m, nil
					}
				} else if m.selectionMode == "edit" {
					if editModel, err := NewVMEditModel(m.vmManager, selectedVM.ID); err == nil {
						editModel.form.SetSize(m.windowWidth-4, m.contentHeight()-2)
						if m.viewRegistry != nil && m.viewRegistry.GetDef(ViewVMEdit) != nil {
							m.viewRegistry.SetActiveModel(m.viewRegistry.GetDef(ViewVMEdit), editModel)
						}
						m.currentView = ViewVMEdit
						m.breadcrumbs.AddItem("Edit "+selectedVM.Name, "vm_edit", 1)
						return m, nil
					}
				}
			}
		}
		m.vmSelectList, cmd = m.vmSelectList.Update(msg)
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
// (from delegateToSubView). Messages are now handled through the framework's
// HandleMessage interface, so this just delegates to handleSubViewMsg.
func (m *MainModel) handleSubViewOutput(nextMsg tea.Msg) (tea.Model, tea.Cmd) {
	return m.handleSubViewMsg(nextMsg)
}


