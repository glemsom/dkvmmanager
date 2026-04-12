package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
)

func TestOnTabChanged(t *testing.T) {
	m := setupTestModel(t)

	// Switch to Configuration tab - breadcrumbs should be cleared (no tab-level breadcrumbs)
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.onTabChanged()

	if m.breadcrumbs.Len() != 0 {
		t.Errorf("Expected 0 breadcrumb items, got %d", m.breadcrumbs.Len())
	}

	// Switch to Power tab
	m.tabModel.SetActiveTab(components.TabPower)
	m.onTabChanged()

	if m.breadcrumbs.Len() != 0 {
		t.Errorf("Expected 0 breadcrumb items after tab change, got %d", m.breadcrumbs.Len())
	}

	// Switch to VMs tab
	m.tabModel.SetActiveTab(components.TabVMs)
	m.onTabChanged()

	if m.breadcrumbs.Len() != 0 {
		t.Errorf("Expected 0 breadcrumb items after tab change, got %d", m.breadcrumbs.Len())
	}
}

func TestOnTabChangedSyncsCursor(t *testing.T) {
	m := setupTestModelWithVMs(t)

	// Set VMs list to index 1
	m.menuList.Select(1)
	m.selectedIndex = 1

	// Switch away and back to VMs tab
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.onTabChanged()
	m.tabModel.SetActiveTab(components.TabVMs)
	m.onTabChanged()

	// Cursor should be restored to selectedIndex (1)
	if m.menuList.Index() != 1 {
		t.Errorf("Expected VMs list cursor at 1 after tab switch, got %d", m.menuList.Index())
	}
	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1, got %d", m.selectedIndex)
	}
}

func TestOnTabChangedConfigCursor(t *testing.T) {
	m := setupTestModel(t)

	// Set config list to index 2
	m.configList.Select(2)
	m.configSelectedIndex = 2

	// Switch away and back to Configuration tab
	m.tabModel.SetActiveTab(components.TabVMs)
	m.onTabChanged()
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.onTabChanged()

	// Config cursor should be restored to configSelectedIndex (2)
	if m.configList.Index() != 2 {
		t.Errorf("Expected config list cursor at 2 after tab switch, got %d", m.configList.Index())
	}
	if m.configSelectedIndex != 2 {
		t.Errorf("Expected configSelectedIndex 2, got %d", m.configSelectedIndex)
	}
}

func TestEnterWorksAfterTabSwitch(t *testing.T) {
	m := setupTestModelWithVMs(t)

	// Switch to Configuration tab via keyboard
	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(*MainModel)

	// Press Enter immediately (without arrow keys) - should not panic
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*MainModel)

	if m == nil {
		t.Error("Model should not be nil after Enter on Configuration tab")
	}
}

func TestTabNavigation(t *testing.T) {
	m := setupTestModel(t)

	// Default tab should be VMs
	if m.tabModel.GetActiveTab() != components.TabVMs {
		t.Error("Expected default tab to be TabVMs")
	}

	// Tab key should switch to Configuration
	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(*MainModel)
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Error("Expected tab to be TabConfiguration after Tab key")
	}

	// Tab key again should switch to Power
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(*MainModel)
	if m.tabModel.GetActiveTab() != components.TabPower {
		t.Error("Expected tab to be TabPower after second Tab key")
	}

	// Shift+Tab should go back
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = model.(*MainModel)
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Error("Expected tab to be TabConfiguration after Shift+Tab")
	}
}

func TestArrowKeysWorkAfterTabSwitch(t *testing.T) {
	m := setupTestModelWithVMs(t)

	// Initial state: VMs tab, menuList cursor at 0
	if m.menuList.Index() != 0 {
		t.Fatalf("Expected menuList cursor at 0 initially, got %d", m.menuList.Index())
	}

	// Press right arrow to switch to Configuration tab
	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRight})
	m = model.(*MainModel)

	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Fatal("Expected tab to be Configuration after right arrow")
	}

	// configList cursor should be at 0
	if m.configList.Index() != 0 {
		t.Errorf("Expected configList cursor at 0 after tab switch, got %d", m.configList.Index())
	}

	// Press down arrow - should immediately move cursor on configList
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*MainModel)

	if m.configList.Index() != 1 {
		t.Errorf("Expected configList cursor at 1 after down arrow, got %d", m.configList.Index())
	}
	if m.configSelectedIndex != 1 {
		t.Errorf("Expected configSelectedIndex at 1, got %d", m.configSelectedIndex)
	}

	// Switch to Power tab
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRight})
	m = model.(*MainModel)

	if m.tabModel.GetActiveTab() != components.TabPower {
		t.Fatal("Expected tab to be Power after second right arrow")
	}

	// powerList cursor should be at 0
	if m.powerList.Index() != 0 {
		t.Errorf("Expected powerList cursor at 0 after switch, got %d", m.powerList.Index())
	}

	// Down arrow on Power tab
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*MainModel)

	if m.powerList.Index() != 1 {
		t.Errorf("Expected powerList cursor at 1 after down arrow, got %d", m.powerList.Index())
	}
}

func TestTabSwitchRestoresCursor(t *testing.T) {
	m := setupTestModel(t)

	// Navigate on Configuration tab: press down twice
	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyTab}) // Switch to Config
	m = model.(*MainModel)
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown}) // cursor → 1
	m = model.(*MainModel)
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown}) // cursor → 2
	m = model.(*MainModel)

	if m.configList.Index() != 2 {
		t.Fatalf("Expected configList cursor at 2, got %d", m.configList.Index())
	}

	// Switch away to Power and back to Configuration
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRight}) // Power
	m = model.(*MainModel)
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyLeft}) // Back to Config
	m = model.(*MainModel)

	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Fatal("Expected Configuration tab after left arrow")
	}

	// Cursor should be restored to 2
	if m.configList.Index() != 2 {
		t.Errorf("Expected configList cursor restored to 2, got %d", m.configList.Index())
	}

	// Arrow keys should work immediately after restore
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*MainModel)

	if m.configList.Index() != 3 {
		t.Errorf("Expected configList cursor at 3 after down arrow on restored tab, got %d", m.configList.Index())
	}
}

func TestVMsTabArrowKeysAfterSwitch(t *testing.T) {
	m := setupTestModelWithVMs(t)

	// Initial cursor should be at 0
	if m.menuList.Index() != 0 {
		t.Fatalf("Expected menuList cursor at 0, got %d", m.menuList.Index())
	}

	// Switch to Configuration and back via Tab
	model, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyTab}) // Config
	m = model.(*MainModel)
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyTab}) // Power
	m = model.(*MainModel)
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyTab}) // VMs (wraps)
	m = model.(*MainModel)

	if m.tabModel.GetActiveTab() != components.TabVMs {
		t.Fatal("Expected VMs tab")
	}

	// Cursor should be restored to 0
	if m.menuList.Index() != 0 {
		t.Errorf("Expected menuList cursor at 0 after return, got %d", m.menuList.Index())
	}

	// Down arrow should work immediately
	model, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*MainModel)

	if m.menuList.Index() != 1 {
		t.Errorf("Expected menuList cursor at 1 after down arrow, got %d", m.menuList.Index())
	}
}
