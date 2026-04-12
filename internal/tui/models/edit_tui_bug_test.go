package models

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
)

// setupTestModelWithTwoVMs creates a MainModel with exactly 2 VMs
func setupTestModelWithTwoVMs(t *testing.T) *MainModel {
	t.Helper()

	tmpDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmpDir, "dkvmmanager"), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	m, err := NewMainModelWithConfig(cfg)
	if err != nil {
		t.Fatalf("Failed to create MainModel: %v", err)
	}

	// Create exactly 2 VMs
	_, err = m.vmManager.CreateVM("test-vm-1")
	if err != nil {
		t.Fatalf("Failed to create VM 1: %v", err)
	}
	_, err = m.vmManager.CreateVM("test-vm-2")
	if err != nil {
		t.Fatalf("Failed to create VM 2: %v", err)
	}

	// Rebuild menu
	m.rebuildMenuList()

	return m
}

// TestEditVMPreservesAllVMsInMenu tests that editing a VM via the form
// does not corrupt the other VM entries in the repository
func TestEditVMPreservesAllVMsInMenu(t *testing.T) {
	m := setupTestModelWithTwoVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	// Verify initial state: 2 VMs in menu
	if len(m.menuItems) != 2 {
		t.Fatalf("Expected 2 menu items initially, got %d", len(m.menuItems))
	}

	// Get the first VM for editing
	vms, err := m.vmManager.ListVMs()
	if err != nil {
		t.Fatalf("Failed to list VMs: %v", err)
	}
	if len(vms) != 2 {
		t.Fatalf("Expected 2 VMs from manager, got %d", len(vms))
	}

	// Create edit model for the first VM
	editModel, err := NewVMEditModel(m.vmManager, vms[0].ID)
	if err != nil {
		t.Fatalf("Failed to create edit model: %v", err)
	}

	// Set the VM name via the form's internal fields (same package access)
	editModel.form.vmName = "test-vm-1-edited"

	// Trigger save via validateAndSave
	_, cmd := editModel.form.validateAndSave()
	if cmd == nil {
		t.Fatal("Expected command from save, got nil")
	}

	msg := cmd()
	updatedMsg, ok := msg.(VMUpdatedMsg)
	if !ok {
		t.Fatalf("Expected VMUpdatedMsg, got %T", msg)
	}
	if updatedMsg.VMName != "test-vm-1-edited" {
		t.Errorf("Expected VM name 'test-vm-1-edited', got '%s'", updatedMsg.VMName)
	}

	// Simulate MainModel handling the update message
	m.statusMessage = "VM '" + updatedMsg.VMName + "' updated successfully"
	m.currentView = ViewConfigMenu
	m.rebuildMenuList()

	// CRITICAL CHECK: Menu should still show 2 VMs
	if len(m.menuItems) != 2 {
		t.Errorf("BUG REPRODUCED: Expected 2 menu items after edit save, got %d", len(m.menuItems))
		for i, item := range m.menuItems {
			t.Logf("  Menu[%d]: %s (Type: %s, VMID: %s)", i, item.Title, item.Type, item.VMID)
		}

		vms, _ := m.vmManager.ListVMs()
		t.Logf("Manager has %d VMs:", len(vms))
		for _, vm := range vms {
			t.Logf("  VM: %s (ID: %s)", vm.Name, vm.ID)
		}
	}

	// Verify the edit was saved
	editedVMID := vms[0].ID
	vms, _ = m.vmManager.ListVMs()
	foundEdited := false
	for _, vm := range vms {
		if vm.ID == editedVMID {
			foundEdited = true
			if vm.Name != "test-vm-1-edited" {
				t.Errorf("Expected edited VM name 'test-vm-1-edited', got '%s'", vm.Name)
			}
		}
	}
	if !foundEdited {
		t.Error("Edited VM not found")
	}
}

// TestFullEditFlowViaUpdate tests the complete flow through the MainModel Update method
func TestFullEditFlowViaUpdate(t *testing.T) {
	m := setupTestModelWithTwoVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	// Verify initial state
	if len(m.menuItems) != 2 {
		t.Fatalf("Expected 2 menu items initially, got %d", len(m.menuItems))
	}

	// Get the first VM
	vms, err := m.vmManager.ListVMs()
	if err != nil {
		t.Fatalf("Failed to list VMs: %v", err)
	}

	// Create edit model
	editModel, err := NewVMEditModel(m.vmManager, vms[0].ID)
	if err != nil {
		t.Fatalf("Failed to create edit model: %v", err)
	}

	m.vmEditModel = editModel
	m.currentView = ViewVMEdit

	// Modify and save directly via form
	m.vmEditModel.form.vmName = "edited-via-update"
	_, cmd := m.vmEditModel.form.validateAndSave()
	if cmd == nil {
		t.Fatal("Expected command from save, got nil")
	}

	// Execute the command and process through MainModel Update
	msg := cmd()
	model, _ := m.Update(msg)
	m = model.(*MainModel)

	// CRITICAL CHECK: Both VMs should still be in the menu
	if len(m.menuItems) != 2 {
		t.Errorf("BUG: Expected 2 menu items after full edit flow, got %d", len(m.menuItems))
		for i, item := range m.menuItems {
			t.Logf("  Menu[%d]: %s (Type: %s, VMID: %s)", i, item.Title, item.Type, item.VMID)
		}

		vms, _ := m.vmManager.ListVMs()
		t.Logf("Manager has %d VMs:", len(vms))
		for _, vm := range vms {
			t.Logf("  VM: %s (ID: %s)", vm.Name, vm.ID)
		}
	}
}
