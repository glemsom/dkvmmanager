package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
)

func TestHandleKeyPressQuit(t *testing.T) {
	m := setupTestModel(t)

	// Test 'q' key
	model, cmd := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = model.(*MainModel)

	if !m.quitting {
		t.Error("Expected quitting to be true after 'q' key")
	}
	if cmd == nil {
		t.Error("Expected quit command after 'q' key")
	}
}

func TestHandleKeyPressCtrlC(t *testing.T) {
	m := setupTestModel(t)

	model, cmd := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = model.(*MainModel)

	if !m.quitting {
		t.Error("Expected quitting to be true after ctrl+c")
	}
	if cmd == nil {
		t.Error("Expected quit command after ctrl+c")
	}
}

func TestHandleKeyPressEscapeMainMenu(t *testing.T) {
	m := setupTestModel(t)
	m.currentView = ViewMainMenu

	model, cmd := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(*MainModel)

	if !m.quitting {
		t.Error("Expected quitting to be true after ESC in main menu")
	}
	if cmd == nil {
		t.Error("Expected quit command after ESC in main menu")
	}
}

func TestHandleKeyPressEscapeSubView(t *testing.T) {
	m := setupTestModel(t)
	m.currentView = ViewVMCreate
	m.vmCreateModel = NewVMCreateModel(m.vmManager)

	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(*MainModel)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected view to be ViewMainMenu after ESC in sub-view, got %s", m.currentView)
	}
	if m.quitting {
		t.Error("Should not be quitting after ESC in sub-view")
	}
}

func TestHandleKeyPressRefresh(t *testing.T) {
	m := setupTestModelWithVMs(t)
	initialCount := len(m.menuList.Items())

	// Add another VM
	_, _ = m.vmManager.CreateVM("refresh-test-vm")

	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	m = model.(*MainModel)

	newCount := len(m.menuList.Items())
	if newCount <= initialCount {
		t.Errorf("Expected more menu items after refresh, got %d (was %d)", newCount, initialCount)
	}
}

func TestHandleKeyPressEnterConfigTab(t *testing.T) {
	m := setupTestModel(t)
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.configSelectedIndex = 0 // "Add new VM"

	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*MainModel)

	if m.currentView != ViewVMCreate {
		t.Errorf("Expected to navigate to VM create view, got %s", m.currentView)
	}
	if m.vmCreateModel == nil {
		t.Error("Expected vmCreateModel to be initialized")
	}
}

func TestHandleKeyPressDelegatesToList(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.currentView = ViewMainMenu

	initialIndex := m.menuList.Index()

	// Send down arrow key
	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*MainModel)

	newIndex := m.menuList.Index()
	if newIndex == initialIndex && len(m.menuList.Items()) > 1 {
		// List should have moved (unless at end)
		t.Logf("List index: %d -> %d (items: %d)", initialIndex, newIndex, len(m.menuList.Items()))
	}
}

func TestHandleKeyPressVMSelectDelegation(t *testing.T) {
	m := setupTestModelWithVMs(t)

	// Enter VM select mode
	m.currentView = ViewVMSelect
	m.selectionMode = "edit"
	m.vmListForSelection, _ = m.vmManager.ListVMs()

	// Send a key - should not panic
	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*MainModel)

	// Just verify it doesn't crash
	if m == nil {
		t.Error("Model should not be nil after key press in VM select")
	}
}

func TestVMSelectEnterDeleteMode(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	// Enter VM selection for deletion
	model, _ := m.showVMSelectionForDeletion()
	m = model.(*MainModel)

	if m.currentView != ViewVMSelect {
		t.Fatalf("Expected ViewVMSelect, got %s", m.currentView)
	}
	if len(m.vmListForSelection) == 0 {
		t.Fatal("Expected VMs to be available for selection")
	}

	// Simulate pressing Enter on the selected VM
	model, _ = m.delegateToSubView(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*MainModel)

	if m.currentView != ViewVMDelete {
		t.Errorf("Expected ViewVMDelete after Enter in delete mode, got %s", m.currentView)
	}
	if m.vmDeleteModel == nil {
		t.Error("Expected vmDeleteModel to be initialized")
	}
}

func TestVMSelectEnterEditMode(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	// Enter VM selection for editing
	model, _ := m.showVMSelection()
	m = model.(*MainModel)

	if m.currentView != ViewVMSelect {
		t.Fatalf("Expected ViewVMSelect, got %s", m.currentView)
	}
	if len(m.vmListForSelection) == 0 {
		t.Fatal("Expected VMs to be available for selection")
	}

	// Simulate pressing Enter on the selected VM
	model, _ = m.delegateToSubView(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*MainModel)

	if m.currentView != ViewVMEdit {
		t.Errorf("Expected ViewVMEdit after Enter in edit mode, got %s", m.currentView)
	}
	if m.vmEditModel == nil {
		t.Error("Expected vmEditModel to be initialized")
	}
}

func TestVMSelectBreadcrumbsAfterDeleteEnter(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	// Enter VM selection for deletion
	model, _ := m.showVMSelectionForDeletion()
	m = model.(*MainModel)

	initialBreadcrumbs := m.breadcrumbs.Len()

	// Simulate pressing Enter on the selected VM
	model, _ = m.delegateToSubView(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*MainModel)

	if m.breadcrumbs.Len() <= initialBreadcrumbs {
		t.Errorf("Expected breadcrumbs to increase after navigating to delete confirmation, got %d (was %d)", m.breadcrumbs.Len(), initialBreadcrumbs)
	}
}

func TestUpdateESCFromVMEdit(t *testing.T) {
	m := setupTestModelWithVMs(t)

	vms, err := m.vmManager.ListVMs()
	if err != nil {
		t.Fatalf("Failed to list VMs: %v", err)
	}

	editModel, err := NewVMEditModel(m.vmManager, vms[0].ID)
	if err != nil {
		t.Fatalf("Failed to create edit model: %v", err)
	}

	m.vmEditModel = editModel
	m.currentView = ViewVMEdit

	// Press ESC through Update() - the real message path
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(*MainModel)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected view to be ViewMainMenu after ESC in Edit VM via Update(), got %s", m.currentView)
	}
	if m.quitting {
		t.Error("Should not be quitting after ESC in sub-view")
	}
}

func TestUpdateESCFromVMCreate(t *testing.T) {
	m := setupTestModel(t)
	m.currentView = ViewVMCreate
	m.vmCreateModel = NewVMCreateModel(m.vmManager)

	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(*MainModel)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected view to be ViewMainMenu after ESC in Create VM via Update(), got %s", m.currentView)
	}
	if m.quitting {
		t.Error("Should not be quitting after ESC in sub-view")
	}
}

func TestStartStopScriptBrowseOpensFileBrowser(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 25
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.configSelectedIndex = 7 // Edit Start/Stop Script

	model, _ := m.handleMenuSelection()
	m = model.(*MainModel)

	if m.currentView != ViewStartStopScript {
		t.Fatalf("Expected ViewStartStopScript, got %s", m.currentView)
	}
	if m.startStopScriptFormModel == nil {
		t.Fatal("Expected startStopScriptFormModel to be initialized")
	}

	// Focus start_browse in custom mode: [toggle, start_path, start_browse, ...]
	m.startStopScriptFormModel.focusIndex = 2
	if m.startStopScriptFormModel.positions[m.startStopScriptFormModel.focusIndex].kind != startStopScriptStartBrowse {
		t.Fatalf("Expected focus on start_browse, got %s", m.startStopScriptFormModel.positions[m.startStopScriptFormModel.focusIndex].fieldName)
	}

	model, _ = m.delegateToSubView(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*MainModel)

	if m.startStopScriptFormModel.fileBrowser == nil {
		t.Fatal("Expected file browser to be created after pressing Enter on browse")
	}
	if !m.startStopScriptFormModel.fileBrowser.active {
		t.Error("Expected file browser to be active")
	}
	if len(m.startStopScriptFormModel.fileBrowser.files) == 0 {
		t.Error("Expected file browser directory listing to be loaded")
	}
}
