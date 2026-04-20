package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
)

// TestScenarioQuitFromMainMenu verifies that 'q' in the main menu triggers quit
// through the full Update() pipeline (Phase 4 KeyMsg dispatch).
func TestScenarioQuitFromMainMenu(t *testing.T) {
	m := setupTestModelForScenarios(t)

	m = sendRunes(t, m, 'q')

	if !m.quitting {
		t.Error("Expected quitting=true after 'q' in main menu")
	}
	if m.currentView != ViewMainMenu {
		t.Errorf("Expected currentView=ViewMainMenu, got %s", m.currentView)
	}
}

// TestScenarioQuitViaCtrlC verifies Ctrl+C triggers quit through Update().
func TestScenarioQuitViaCtrlC(t *testing.T) {
	m := setupTestModelForScenarios(t)

	m = sendKeys(t, m, tea.KeyCtrlC)

	if !m.quitting {
		t.Error("Expected quitting=true after Ctrl+C")
	}
}

// TestScenarioTabSwitching verifies tab navigation through Update() Phase 4.
func TestScenarioTabSwitching(t *testing.T) {
	m := setupTestModelForScenarios(t)

	// Default tab is VMs
	if m.tabModel.GetActiveTab() != components.TabVMs {
		t.Fatal("Expected default tab to be TabVMs")
	}

	// Press Tab for Configuration
	m = sendKeys(t, m, tea.KeyTab)
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Errorf("Expected TabConfiguration after Tab, got %d", m.tabModel.GetActiveTab())
	}

	// Press Tab for Power
	m = sendKeys(t, m, tea.KeyTab)
	if m.tabModel.GetActiveTab() != components.TabPower {
		t.Errorf("Expected TabPower after second Tab, got %d", m.tabModel.GetActiveTab())
	}

	// Press Shift+Tab to return to Configuration
	m = sendKeys(t, m, tea.KeyShiftTab)
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Errorf("Expected TabConfiguration after Shift+Tab, got %d", m.tabModel.GetActiveTab())
	}
}

// TestScenarioTabKeyCycling verifies Tab/Shift+Tab cycling through Update().
func TestScenarioTabKeyCycling(t *testing.T) {
	m := setupTestModelForScenarios(t)

	m = sendKeys(t, m, tea.KeyTab)
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Error("Expected TabConfiguration after Tab")
	}

	m = sendKeys(t, m, tea.KeyTab)
	if m.tabModel.GetActiveTab() != components.TabPower {
		t.Error("Expected TabPower after second Tab")
	}

	m = sendKeys(t, m, tea.KeyShiftTab)
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Error("Expected TabConfiguration after Shift+Tab")
	}
}

// TestScenarioNavigateToVMCreate verifies the full flow: Config tab -> select "Add new VM" -> Enter
func TestScenarioNavigateToVMCreate(t *testing.T) {
	m := setupTestModelForScenarios(t)

	// Switch to Configuration tab
	m = sendKeys(t, m, tea.KeyTab)
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Fatal("Expected Configuration tab")
	}

	// configSelectedIndex=0 is "Add new VM", press Enter
	m = sendKeys(t, m, tea.KeyEnter)

	if m.currentView != ViewVMCreate {
		t.Errorf("Expected ViewVMCreate, got %s", m.currentView)
	}
	if m.vmCreateModel == nil {
		t.Error("Expected vmCreateModel to be initialized")
	}
}

// TestScenarioESCFromVMCreate verifies ESC returns to main menu from VM Create
// through the Phase 1 ESC intercept in Update().
func TestScenarioESCFromVMCreate(t *testing.T) {
	m := setupTestModelForScenarios(t)

	// Navigate to VM create
	m = sendKeys(t, m, tea.KeyTab)       // Config tab
	m = sendKeys(t, m, tea.KeyEnter) // "Add new VM"

	if m.currentView != ViewVMCreate {
		t.Fatalf("Expected ViewVMCreate, got %s", m.currentView)
	}

	// Press ESC through Update()
	m = sendKeys(t, m, tea.KeyEsc)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected ViewMainMenu after ESC, got %s", m.currentView)
	}
	if m.quitting {
		t.Error("Should not be quitting after ESC in sub-view")
	}
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Errorf("Expected tab to return to TabConfiguration, got %d", m.tabModel.GetActiveTab())
	}
}

// TestScenarioESCFromVMEdit verifies ESC returns to main menu from VM Edit.
func TestScenarioESCFromVMEdit(t *testing.T) {
	m := setupTestModelForScenarios(t)

	// Set up VM edit view directly
	vms, _ := m.vmManager.ListVMs()
	editModel, err := NewVMEditModel(m.vmManager, vms[0].ID)
	if err != nil {
		t.Fatalf("Failed to create edit model: %v", err)
	}
	m.vmEditModel = editModel
	m.currentView = ViewVMEdit

	m = sendKeys(t, m, tea.KeyEsc)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected ViewMainMenu after ESC, got %s", m.currentView)
	}
	if m.quitting {
		t.Error("Should not be quitting after ESC in sub-view")
	}
}

// TestScenarioESCFromVMDelete verifies ESC returns to main menu from VM Delete.
func TestScenarioESCFromVMDelete(t *testing.T) {
	m := setupTestModelForScenarios(t)

	vms, _ := m.vmManager.ListVMs()
	deleteModel, err := NewVMDeleteModel(m.vmManager, vms[0].ID)
	if err != nil {
		t.Fatalf("Failed to create delete model: %v", err)
	}
	m.vmDeleteModel = deleteModel
	m.currentView = ViewVMDelete

	m = sendKeys(t, m, tea.KeyEsc)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected ViewMainMenu after ESC, got %s", m.currentView)
	}
	if m.quitting {
		t.Error("Should not be quitting after ESC in sub-view")
	}
}

// TestScenarioVMCreatedMsg verifies VMCreatedMsg handling through Update() Phase 2.
func TestScenarioVMCreatedMsg(t *testing.T) {
	m := setupTestModelForScenarios(t)

	initialCount := len(m.menuItems)

	model, _ := m.Update(VMCreatedMsg{VMName: "brand-new-vm"})
	m = model.(*MainModel)

	if !strings.Contains(m.statusMessage, "brand-new-vm") {
		t.Errorf("Expected status message to contain VM name, got '%s'", m.statusMessage)
	}
	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected ViewConfigMenu after VMCreatedMsg, got %s", m.currentView)
	}
	_ = initialCount
}

// TestScenarioVMUpdatedMsg verifies VMUpdatedMsg handling through Update() Phase 2.
func TestScenarioVMUpdatedMsg(t *testing.T) {
	m := setupTestModelForScenarios(t)

	model, _ := m.Update(VMUpdatedMsg{VMName: "edited-vm"})
	m = model.(*MainModel)

	if !strings.Contains(m.statusMessage, "edited-vm") {
		t.Errorf("Expected status message to contain VM name, got '%s'", m.statusMessage)
	}
	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected ViewConfigMenu after VMUpdatedMsg, got %s", m.currentView)
	}
}

// TestScenarioVMDeletedMsg verifies VMDeletedMsg handling through Update() Phase 2.
func TestScenarioVMDeletedMsg(t *testing.T) {
	m := setupTestModelForScenarios(t)

	model, _ := m.Update(VMDeletedMsg{VMName: "old-vm", VMID: "99"})
	m = model.(*MainModel)

	if !strings.Contains(m.statusMessage, "old-vm") {
		t.Errorf("Expected status message to contain VM name, got '%s'", m.statusMessage)
	}
	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected ViewConfigMenu after VMDeletedMsg, got %s", m.currentView)
	}
}

// TestScenarioLBUCommitDryRun verifies LBU commit through Update() with dry-run mode.
func TestScenarioLBUCommitDryRun(t *testing.T) {
	m := setupTestModelForScenarios(t)
	dryRunMode = true
	defer func() { dryRunMode = false }()

	// Switch to Configuration tab, select "Save changes" (index 11)
	m = sendKeys(t, m, tea.KeyTab)
	m.configSelectedIndex = 11

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*MainModel)

	if cmd == nil {
		t.Fatal("Expected tea.Cmd for LBU commit")
	}

	msg := cmd()
	lcm, ok := msg.(LBUCommitMsg)
	if !ok {
		t.Fatalf("Expected LBUCommitMsg, got %T", msg)
	}

	// Feed the LBUCommitMsg back through Update()
	model, _ = m.Update(lcm)
	m = model.(*MainModel)

	if !strings.Contains(m.statusBar.Render(m.windowWidth), "Would execute: lbu commit") {
		t.Error("Expected status bar to show dry-run output")
	}
}

// TestScenarioWindowSizeMsg verifies WindowSizeMsg handling through Update() Phase 4.
func TestScenarioWindowSizeMsg(t *testing.T) {
	m := setupTestModelForScenarios(t)

	model, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = model.(*MainModel)

	if m.windowWidth != 120 {
		t.Errorf("Expected windowWidth=120, got %d", m.windowWidth)
	}
	if m.windowHeight != 40 {
		t.Errorf("Expected windowHeight=40, got %d", m.windowHeight)
	}

	// View should render without panic
	view := m.View()
	if view == "" {
		t.Error("View should not be empty after WindowSizeMsg")
	}
}

// TestScenarioRefreshKey verifies 'r' refreshes the VM list through Update().
func TestScenarioRefreshKey(t *testing.T) {
	m := setupTestModelForScenarios(t)
	initialCount := len(m.menuItems)

	// Add a VM behind the scenes
	_, _ = m.vmManager.CreateVM("refresh-scenario-vm")

	m = sendRunes(t, m, 'r')

	newCount := len(m.menuItems)
	if newCount <= initialCount {
		t.Errorf("Expected more menu items after refresh, got %d (was %d)", newCount, initialCount)
	}
}

// TestScenarioConfigTabMenuItems verifies navigating config tab items through Update().
func TestScenarioConfigTabMenuItems(t *testing.T) {
	m := setupTestModelForScenarios(t)

	// Switch to Configuration tab
	m = sendKeys(t, m, tea.KeyTab)

	// Select "Edit CPU Options" (index 8)
	m.configSelectedIndex = 8
	m = sendKeysWithCmd(t, m, tea.KeyEnter)

	if m.currentView != ViewCPUOptions {
		t.Errorf("Expected ViewCPUOptions, got %s", m.currentView)
	}
	if m.cpuOptionsModel == nil {
		t.Error("Expected cpuOptionsModel to be initialized")
	}
}

// TestScenarioESCFromCPUOptions verifies ESC returns from CPU Options sub-view.
func TestScenarioESCFromCPUOptions(t *testing.T) {
	m := setupTestModelForScenarios(t)

	m.currentView = ViewCPUOptions
	m.cpuOptionsModel = NewCPUOptionsModel(m.vmManager)

	m = sendKeys(t, m, tea.KeyEsc)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected ViewMainMenu after ESC, got %s", m.currentView)
	}
}

// TestScenarioComingSoonItems verifies menu selection doesn't panic.
// Note: Custom Scripts (item 6) and SSH Password (item 8) are now implemented.
func TestScenarioComingSoonItems(t *testing.T) {
	m := setupTestModelForScenarios(t)
	m.tabModel.SetActiveTab(components.TabConfiguration)

	// Just verify no panic on menu selection
	for i := 0; i <= 9; i++ {
		m.configSelectedIndex = i
		model, _ := m.handleMenuSelection()
		_ = model.(*MainModel)
	}
	// Skip index 10 (Create LV) here because it triggers a vgs command.
	m.configSelectedIndex = 11
	model, _ := m.handleMenuSelection()
	_ = model.(*MainModel)
}

// TestScenarioViewChangeMsg verifies ViewChangeMsg handling through Update() Phase 2.
func TestScenarioViewChangeMsg(t *testing.T) {
	m := setupTestModelForScenarios(t)

	model, _ := m.Update(ViewChangeMsg{View: ViewConfigMenu})
	m = model.(*MainModel)

	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected ViewConfigMenu, got %s", m.currentView)
	}
}

// TestScenarioViewChangeToMainMenuRebuilds verifies menu rebuild on return.
func TestScenarioViewChangeToMainMenuRebuilds(t *testing.T) {
	m := setupTestModelForScenarios(t)
	m.currentView = ViewConfigMenu

	model, _ := m.Update(ViewChangeMsg{View: ViewMainMenu})
	m = model.(*MainModel)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected ViewMainMenu, got %s", m.currentView)
	}
	if len(m.menuList.Items()) != 2 {
		t.Errorf("Expected 2 menu items after rebuild, got %d", len(m.menuList.Items()))
	}
}

// TestScenarioFullEditFlowViaUpdate exercises the complete edit flow:
// Config tab -> Edit VM -> Select VM -> Modify -> Save -> Verify menu preserved
func TestScenarioFullEditFlowViaUpdate(t *testing.T) {
	m := setupTestModelForScenarios(t)

	if len(m.menuItems) != 2 {
		t.Fatalf("Expected 2 menu items initially, got %d", len(m.menuItems))
	}

	// Get the first VM
	vms, _ := m.vmManager.ListVMs()

	// Create edit model and set it up
	editModel, err := NewVMEditModel(m.vmManager, vms[0].ID)
	if err != nil {
		t.Fatalf("Failed to create edit model: %v", err)
	}
	m.vmEditModel = editModel
	m.currentView = ViewVMEdit

	// Modify and save
	m.vmEditModel.form.vmName = "edited-via-scenario"
	_, cmd := m.vmEditModel.form.validateAndSave()
	if cmd == nil {
		t.Fatal("Expected command from save")
	}

	// Process VMUpdatedMsg through Update()
	msg := cmd()
	model, _ := m.Update(msg)
	m = model.(*MainModel)

	// Both VMs should still be in menu
	if len(m.menuItems) != 2 {
		t.Errorf("Expected 2 menu items after edit, got %d", len(m.menuItems))
	}
	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected ViewConfigMenu after edit save, got %s", m.currentView)
	}
}

// TestScenarioVMStoppedMsg verifies VMStoppedMsg handling through Update() Phase 2.
func TestScenarioVMStoppedMsg(t *testing.T) {
	m := setupTestModelForScenarios(t)

	// Simulate being in the running VM view
	m.currentView = ViewVMRunning

	model, _ := m.Update(VMStoppedMsg{VMName: "running-vm", Reason: "exited"})
	m = model.(*MainModel)

	// VMStoppedMsg returns to main menu
	if m.currentView != ViewMainMenu {
		t.Errorf("Expected currentView=ViewMainMenu after VM stopped, got %s", m.currentView)
	}

	// Status bar should mention the stopped VM
	view := m.statusBar.Render(m.windowWidth)
	if !strings.Contains(view, "running-vm") {
		t.Errorf("Expected status bar to mention VM name, got '%s'", view)
	}

	// Running count should be 0
	if !strings.Contains(view, "0 running") {
		t.Errorf("Expected status bar to show 0 running, got '%s'", view)
	}
}

// TestScenarioDeactivateDryRun verifies the Update pipeline works correctly
// when returning from sub-views via ESC with the full 4-phase cascade.
func TestScenarioDeactivateDryRun(t *testing.T) {
	m := setupTestModelForScenarios(t)

	// Start at main menu, go to config, open PCI passthrough, then ESC back
	m = sendKeys(t, m, tea.KeyTab) // Config tab
	m.configSelectedIndex = 5 // PCI Passthrough
	model, _ := m.handleMenuSelection()
	m = model.(*MainModel)

	if m.currentView != ViewPCIPassthrough {
		t.Fatalf("Expected ViewPCIPassthrough, got %s", m.currentView)
	}

	// ESC through Update() — tests Phase 1 intercept
	m = sendKeys(t, m, tea.KeyEsc)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected ViewMainMenu after ESC, got %s", m.currentView)
	}
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Errorf("Expected tab to be TabConfiguration after ESC, got %d", m.tabModel.GetActiveTab())
	}
}
