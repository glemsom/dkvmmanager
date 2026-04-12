package vm

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
)

// TestEditVMPreservesOtherVMs tests that editing one VM doesn't remove other VMs
func TestEditVMPreservesOtherVMs(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Create two VMs
	vm1, err := mgr.CreateVM("vm-one")
	if err != nil {
		t.Fatalf("Failed to create VM 1: %v", err)
	}

	vm2, err := mgr.CreateVM("vm-two")
	if err != nil {
		t.Fatalf("Failed to create VM 2: %v", err)
	}

	// Verify both exist
	vms, err := mgr.ListVMs()
	if err != nil {
		t.Fatalf("ListVMs() error: %v", err)
	}
	if len(vms) != 2 {
		t.Fatalf("Expected 2 VMs, got %d", len(vms))
	}

	// Now edit VM 1 (simulating the edit flow)
	vm1.Name = "vm-one-edited"
	if err := mgr.SaveVM(vm1); err != nil {
		t.Fatalf("SaveVM() error: %v", err)
	}

	// Verify BOTH VMs still exist after edit
	vms, err = mgr.ListVMs()
	if err != nil {
		t.Fatalf("ListVMs() after edit error: %v", err)
	}
	if len(vms) != 2 {
		t.Errorf("Expected 2 VMs after editing one, got %d", len(vms))
		for _, vm := range vms {
			t.Logf("  VM: %s (ID: %s)", vm.Name, vm.ID)
		}
	}

	// Verify the edit was saved
	found := false
	for _, vm := range vms {
		if vm.ID == vm1.ID {
			found = true
			if vm.Name != "vm-one-edited" {
				t.Errorf("Expected edited VM name 'vm-one-edited', got '%s'", vm.Name)
			}
		}
	}
	if !found {
		t.Error("Edited VM not found in list")
	}

	// Verify VM2 still exists with correct name
	found2 := false
	for _, vm := range vms {
		if vm.ID == vm2.ID {
			found2 = true
			if vm.Name != "vm-two" {
				t.Errorf("Expected VM2 name 'vm-two', got '%s'", vm.Name)
			}
		}
	}
	if !found2 {
		t.Error("VM2 not found in list after editing VM1")
	}
}

// TestEditVMMultipleTimes tests editing the same VM multiple times
func TestEditVMMultipleTimes(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Create two VMs
	vm1, err := mgr.CreateVM("vm-one")
	if err != nil {
		t.Fatalf("Failed to create VM 1: %v", err)
	}
	vm2, err := mgr.CreateVM("vm-two")
	if err != nil {
		t.Fatalf("Failed to create VM 2: %v", err)
	}

	// Edit VM1 multiple times
	for i := 0; i < 3; i++ {
		vm1.Name = fmt.Sprintf("vm-one-edit-%d", i)
		if err := mgr.SaveVM(vm1); err != nil {
			t.Fatalf("SaveVM() error on edit %d: %v", i, err)
		}

		vms, err := mgr.ListVMs()
		if err != nil {
			t.Fatalf("ListVMs() error on edit %d: %v", i, err)
		}
		if len(vms) != 2 {
			t.Errorf("Expected 2 VMs after edit %d, got %d", i, len(vms))
		}
	}

	// Verify VM2 is untouched
	vms, err := mgr.ListVMs()
	if err != nil {
		t.Fatalf("ListVMs() error: %v", err)
	}
	found2 := false
	for _, vm := range vms {
		if vm.ID == vm2.ID {
			found2 = true
			if vm.Name != "vm-two" {
				t.Errorf("Expected VM2 name 'vm-two', got '%s'", vm.Name)
			}
		}
	}
	if !found2 {
		t.Error("VM2 not found after multiple edits of VM1")
	}
}
