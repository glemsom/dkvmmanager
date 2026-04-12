package models

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/config"
)

// TestFullTUIEditFlowReproducingBug simulates the exact user flow:
// Configuration Tab -> Edit VM -> Select VM -> Save changes
// Then verify VMs tab shows both VMs
func TestFullTUIEditFlowReproducingBug(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "dkvmmanager"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	m, err := NewMainModelWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	m.windowWidth = 80
	m.windowHeight = 30

	// Create 2 VMs
	m.vmManager.CreateVM("vm-one")
	m.vmManager.CreateVM("vm-two")
	m.rebuildMenuList()

	if len(m.menuItems) != 2 {
		t.Fatalf("Initial: expected 2 menu items, got %d", len(m.menuItems))
	}

	// Step 1: Switch to Configuration tab
	m.tabModel.SetActiveTab(1) // TabConfiguration

	// Step 2: Select "Edit VM" (index 1)
	m.configSelectedIndex = 1
	model, _ := m.handleMenuSelection()
	m = model.(*MainModel)

	if m.currentView != ViewVMSelect {
		t.Fatalf("Expected ViewVMSelect, got %s", m.currentView)
	}

	// Step 3: Press Enter to select first VM (enter edit mode)
	model, _ = m.delegateToSubView(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*MainModel)

	if m.currentView != ViewVMEdit {
		t.Fatalf("Expected ViewVMEdit, got %s", m.currentView)
	}

	// Step 4: Modify the VM name via form and save
	m.vmEditModel.form.vmName = "vm-one-edited"
	_, cmd := m.vmEditModel.form.validateAndSave()
	if cmd == nil {
		t.Fatal("Expected command from save, got nil")
	}

	// Execute the command to get VMUpdatedMsg
	msg := cmd()

	// Step 5: Process the message through MainModel Update
	model, _ = m.Update(msg)
	m = model.(*MainModel)

	// Step 6: Switch to VMs tab
	m.tabModel.SetActiveTab(0) // TabVMs

	// The menu list should have both VMs
	if len(m.menuItems) != 2 {
		t.Errorf("BUG: After edit save, expected 2 menu items, got %d", len(m.menuItems))
		for i, item := range m.menuItems {
			t.Logf("  Menu[%d]: %s (Type: %s)", i, item.Title, item.Type)
		}
	}

	// Also verify via renderVMsTab
	view := m.renderVMsTab()
	if view == "" {
		t.Error("renderVMsTab returned empty string")
	}
}

// TestDoubleVMCreatedMsg tests if processing VMCreatedMsg twice causes issues
func TestDoubleVMCreatedMsg(t *testing.T) {
	m := setupTestModelWithVMs(t)

	// Verify initial state
	if len(m.menuItems) != 2 {
		t.Fatalf("Expected 2 menu items, got %d", len(m.menuItems))
	}

	// Simulate the double-processing scenario:
	// 1. handleSubViewMsg processes VMCreatedMsg (calls rebuildMenuList)
	m.statusMessage = "VM 'test-vm-1' created successfully"
	m.currentView = ViewConfigMenu
	m.rebuildMenuList()

	// 2. Then VMCreatedMsg is also processed by outer Update handler
	model, _ := m.Update(VMCreatedMsg{VMName: "test-vm-1"})
	m = model.(*MainModel)

	// Should still have 2 VMs
	if len(m.menuItems) != 2 {
		t.Errorf("After double VMCreatedMsg, expected 2 menu items, got %d", len(m.menuItems))
	}
}

// TestEditSaveThenSwitchToVMsTab tests the complete flow including tab switching
func TestEditSaveThenSwitchToVMsTab(t *testing.T) {
	m := setupTestModelWithTwoVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	// Simulate editing a VM and saving
	vms, _ := m.vmManager.ListVMs()
	editModel, _ := NewVMEditModel(m.vmManager, vms[0].ID)

	// Set values via form fields and save
	editModel.form.vmName = "edited-vm"
	editModel.form.hardDisks = []string{""}
	editModel.form.cdroms = []string{}
	editModel.form.macAddress = "11:22:33:44:55:66"
	editModel.form.vncEnabled = true

	_, cmd := editModel.form.validateAndSave()
	if cmd == nil {
		t.Fatal("validateAndSave returned nil command")
	}
	msg := cmd()

	// Process through main model Update
	m.Update(msg)

	// Switch to VMs tab
	m.tabModel.SetActiveTab(0)

	// Check menu items
	if len(m.menuItems) != 2 {
		t.Errorf("Expected 2 menu items, got %d", len(m.menuItems))
		for i, item := range m.menuItems {
			t.Logf("  Menu[%d]: %s", i, item.Title)
		}
	}
}
