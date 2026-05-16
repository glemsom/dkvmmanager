package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWindowSizeMsg(t *testing.T) {
	m := setupTestModel(t)

	model, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = model.(*MainModel)

	if m.windowWidth != 100 {
		t.Errorf("Expected windowWidth to be 100, got %d", m.windowWidth)
	}
	if m.windowHeight != 30 {
		t.Errorf("Expected windowHeight to be 30, got %d", m.windowHeight)
	}

	// Verify view renders without panic
	view := m.View()
	if view == "" {
		t.Error("View should not be empty after window resize")
	}
}

func TestViewChangeMsg(t *testing.T) {
	m := setupTestModel(t)

	model, _ := m.Update(ViewChangeMsg{View: ViewConfigMenu})
	m = model.(*MainModel)

	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected view to be ViewConfigMenu, got %s", m.currentView)
	}
}

func TestViewChangeMsgToMainMenuRebuildsList(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.currentView = ViewConfigMenu

	model, _ := m.Update(ViewChangeMsg{View: ViewMainMenu})
	m = model.(*MainModel)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected view to be ViewMainMenu, got %s", m.currentView)
	}

	// Menu list should have been rebuilt with VMs only
	items := m.menuList.Items()
	if len(items) != 2 {
		t.Errorf("Expected 2 menu items after rebuild (VMs only), got %d", len(items))
	}
}

func TestVMCreatedMsg(t *testing.T) {
	m := setupTestModel(t)

	model, _ := m.Update(VMCreatedMsg{VMName: "new-test-vm"})
	m = model.(*MainModel)

	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected view to be ViewConfigMenu after VM created, got %s", m.currentView)
	}
	// Check status bar message (UnifiedViewReturn sets it there)
	expected := "VM 'new-test-vm' created successfully"
	if m.statusBar.GetMessage() != expected {
		t.Errorf("Expected status bar message '%s', got '%s'", expected, m.statusBar.GetMessage())
	}
}

func TestVMDeletedMsg(t *testing.T) {
	m := setupTestModel(t)

	model, _ := m.Update(VMDeletedMsg{VMName: "deleted-vm", VMID: "0"})
	m = model.(*MainModel)

	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected view to be ViewConfigMenu after VM deleted, got %s", m.currentView)
	}
	// Check status bar message (UnifiedViewReturn sets it there)
	expected := "VM 'deleted-vm' deleted successfully"
	if m.statusBar.GetMessage() != expected {
		t.Errorf("Expected status bar message '%s', got '%s'", expected, m.statusBar.GetMessage())
	}
}

func TestUnifiedViewReturn(t *testing.T) {
	m := setupTestModel(t)

	// Set up state that should be cleared
	m.currentView = ViewVMCreate
	m.breadcrumbs.AddItem("TestPath", "test", 0)

	// Create a registry and activate it with a fake active view
	// We directly set the activeView field to simulate an active registry
	m.viewRegistry = NewViewRegistry()
	m.viewRegistry.activeView = &ActiveView{} // Mark registry as "active"

	_, _ = m.UnifiedViewReturn("test message")

	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected view to be ViewConfigMenu, got %s", m.currentView)
	}
	if m.statusBar.GetMessage() != "test message" {
		t.Errorf("Expected status bar message 'test message', got '%s'", m.statusBar.GetMessage())
	}
	if len(m.breadcrumbs.GetItems()) != 0 {
		t.Errorf("Expected breadcrumbs to be cleared, got %d items", len(m.breadcrumbs.GetItems()))
	}
	if m.viewRegistry.IsActive() {
		t.Error("Expected view registry to be deactivated")
	}
}

func TestUnifiedViewReturnEmptyMessage(t *testing.T) {
	m := setupTestModel(t)
	m.currentView = ViewVMCreate

	// Empty message should not change status bar
	m.statusBar.SetMessage("existing message")
	_, _ = m.UnifiedViewReturn("")

	// Status bar should keep existing message (not be cleared)
	if m.statusBar.GetMessage() != "existing message" {
		t.Errorf("Expected status bar message to remain 'existing message', got '%s'", m.statusBar.GetMessage())
	}
	// But view should still change
	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected view to be ViewConfigMenu, got %s", m.currentView)
	}
}

func TestMessageHandlers_Registry(t *testing.T) {
	// Test that registry handlers are registered correctly
	handler := messageHandlers["VMCreatedMsg"]
	if handler == nil {
		t.Fatal("VMCreatedMsg handler not registered")
	}

	handler = messageHandlers["VMUpdatedMsg"]
	if handler == nil {
		t.Fatal("VMUpdatedMsg handler not registered")
	}

	handler = messageHandlers["VMDeletedMsg"]
	if handler == nil {
		t.Fatal("VMDeletedMsg handler not registered")
	}

	handler = messageHandlers["PCIVFIOKernelAppliedMsg"]
	if handler == nil {
		t.Fatal("PCIVFIOKernelAppliedMsg handler not registered")
	}

	handler = messageHandlers["VCPUCPUKernelAppliedMsg"]
	if handler == nil {
		t.Fatal("VCPUCPUKernelAppliedMsg handler not registered")
	}

	handler = messageHandlers["LBUCommitMsg"]
	if handler == nil {
		t.Fatal("LBUCommitMsg handler not registered")
	}
}

func TestMessageHandlers_Dispatch(t *testing.T) {
	m := setupTestModel(t)

	// Test that VMCreatedMsg dispatches through the registry to handleSubViewMsg
	model, _ := m.handleSubViewMsg(VMCreatedMsg{VMName: "registry-test-vm"})
	m = model.(*MainModel)

	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected view to be ViewConfigMenu, got %s", m.currentView)
	}

	expected := "VM 'registry-test-vm' created successfully"
	if m.statusBar.GetMessage() != expected {
		t.Errorf("Expected status bar message '%s', got '%s'", expected, m.statusBar.GetMessage())
	}
}

func TestMessageHandlers_PCIPassthroughKernelApplied(t *testing.T) {
	m := setupTestModel(t)

	// Test success case
	model, _ := m.handleSubViewMsg(PCIVFIOKernelAppliedMsg{Success: true})
	m = model.(*MainModel)

	if m.statusBar.GetMessage() != "vfio-pci.ids applied to kernel successfully" {
		t.Errorf("Unexpected status for success case: %s", m.statusBar.GetMessage())
	}

	// Test failure case
	model, _ = m.handleSubViewMsg(PCIVFIOKernelAppliedMsg{Success: false, Error: "test error"})
	m = model.(*MainModel)

	if m.statusBar.GetMessage() != "Apply to kernel failed: test error" {
		t.Errorf("Unexpected status for failure case: %s", m.statusBar.GetMessage())
	}
}

func TestMessageHandlers_LBUCommit(t *testing.T) {
	m := setupTestModel(t)

	// LBUCommitMsg should NOT change view (it's a status-only message)
	model, _ := m.handleSubViewMsg(LBUCommitMsg{Success: true, Output: "committed"})
	m = model.(*MainModel)

	// View should remain unchanged
	if m.currentView != ViewMainMenu {
		t.Errorf("Expected view to remain ViewMainMenu, got %s", m.currentView)
	}
	if m.statusBar.GetMessage() != "LBU commit: committed" {
		t.Errorf("Unexpected status bar message: %s", m.statusBar.GetMessage())
	}
}

