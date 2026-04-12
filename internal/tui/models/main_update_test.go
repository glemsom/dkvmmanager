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

	if m.statusMessage == "" {
		t.Error("Expected status message to be set")
	}
	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected view to be ViewConfigMenu after VM created, got %s", m.currentView)
	}
}

func TestVMDeletedMsg(t *testing.T) {
	m := setupTestModel(t)

	model, _ := m.Update(VMDeletedMsg{VMName: "deleted-vm", VMID: "0"})
	m = model.(*MainModel)

	if m.statusMessage == "" {
		t.Error("Expected status message to be set")
	}
	if m.currentView != ViewConfigMenu {
		t.Errorf("Expected view to be ViewConfigMenu after VM deleted, got %s", m.currentView)
	}
}

