package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
)

// TestYAMLIntegrityAfterSave tests the actual YAML file after save operations
func TestYAMLIntegrityAfterSave(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dkvmmanager")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(configDir, "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Create two VMs
	vm1, err := mgr.CreateVM("test-vm-1")
	if err != nil {
		t.Fatalf("CreateVM 1: %v", err)
	}
	vm2, err := mgr.CreateVM("test-vm-2")
	if err != nil {
		t.Fatalf("CreateVM 2: %v", err)
	}

	// Read YAML after creation
	data, _ := os.ReadFile(cfg.VMsConfigFile)
	t.Logf("YAML after creating 2 VMs:\n%s", string(data))

	// Verify 2 VMs via ListVMs
	vms, _ := mgr.ListVMs()
	t.Logf("ListVMs after create: %d VMs", len(vms))

	// Now edit VM1 (simulating the edit model flow)
	// First, load VM1 (like NewVMEditModel does)
	allVMs, _ := mgr.ListVMs()
	var targetVM *models.VM
	for i := range allVMs {
		if allVMs[i].ID == vm1.ID {
			targetVM = &allVMs[i]
			break
		}
	}
	if targetVM == nil {
		t.Fatal("Could not find VM1")
	}

	// Edit the name (like VMEditModel.saveVM does)
	targetVM.Name = "test-vm-1-edited"
	targetVM.HardDisks = []string{"/dev/sda"}
	targetVM.CDROMs = []string{}
	targetVM.MAC = "11:22:33:44:55:66"
	targetVM.VNCListen = "0.0.0.0:0"
	targetVM.GPUROM = ""

	// Save
	if err := mgr.SaveVM(targetVM); err != nil {
		t.Fatalf("SaveVM: %v", err)
	}

	// Read YAML after edit
	data, _ = os.ReadFile(cfg.VMsConfigFile)
	t.Logf("YAML after editing VM1:\n%s", string(data))

	// Verify BOTH VMs still exist
	vms, err = mgr.ListVMs()
	if err != nil {
		t.Fatalf("ListVMs error: %v", err)
	}
	t.Logf("ListVMs after edit: %d VMs", len(vms))
	for _, vm := range vms {
		t.Logf("  VM: %s (ID: %s)", vm.Name, vm.ID)
	}

	if len(vms) != 2 {
		t.Errorf("Expected 2 VMs after edit, got %d", len(vms))
	}

	// Verify VM2 is unchanged
	found2 := false
	for _, vm := range vms {
		if vm.ID == vm2.ID {
			found2 = true
			if vm.Name != "test-vm-2" {
				t.Errorf("VM2 name changed! Got '%s', want 'test-vm-2'", vm.Name)
			}
		}
	}
	if !found2 {
		t.Errorf("VM2 disappeared after editing VM1!")
	}
}

// TestSaveVMWithEmptySlices tests saving VMs with empty harddisks/cdroms
func TestSaveVMWithEmptySlices(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dkvmmanager")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(configDir, "vms.yaml"),
	}

	mgr, _ := NewManager(cfg)

	// Create VM with disks
	vm1, _ := mgr.CreateVM("vm-with-disks")
	vm1.HardDisks = []string{"/dev/sda", "/dev/sdb"}
	vm1.CDROMs = []string{"/iso/test.iso"}
	mgr.SaveVM(vm1)

	// Create VM without disks
	vm2, _ := mgr.CreateVM("vm-no-disks")
	vm2.HardDisks = []string{}
	vm2.CDROMs = []string{}
	mgr.SaveVM(vm2)

	// Now edit VM1 to remove all disks
	vm1.HardDisks = []string{}
	vm1.CDROMs = []string{}
	mgr.SaveVM(vm1)

	// Verify both VMs still exist
	vms, _ := mgr.ListVMs()
	if len(vms) != 2 {
		t.Errorf("Expected 2 VMs, got %d", len(vms))
		for _, vm := range vms {
			t.Logf("  VM: %s (ID: %s, Disks: %v)", vm.Name, vm.ID, vm.HardDisks)
		}
	}
}

// TestRepositorySavePreservesOtherVMs specifically tests the repository SaveVM
func TestRepositorySavePreservesOtherVMs(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dkvmmanager")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configFile := filepath.Join(configDir, "vms.yaml")
	repo, err := NewRepository(configFile)
	if err != nil {
		t.Fatal(err)
	}

	// Save VM 0
	vm0 := &models.VM{
		ID:        "0",
		Name:      "vm-zero",
		HardDisks: []string{"/dev/sda"},
		CDROMs:    []string{"/iso/a.iso"},
		MAC:       "aa:bb:cc:dd:ee:ff",
		VNCListen: "0.0.0.0:0",
	}
	repo.SaveVM(vm0)

	// Save VM 1
	vm1 := &models.VM{
		ID:        "1",
		Name:      "vm-one",
		HardDisks: []string{"/dev/sdb"},
		CDROMs:    []string{"/iso/b.iso"},
		MAC:       "11:22:33:44:55:66",
		VNCListen: "0.0.0.0:1",
	}
	repo.SaveVM(vm1)

	// Verify 2 VMs
	vms, _ := repo.ListVMs()
	if len(vms) != 2 {
		t.Fatalf("Expected 2 VMs, got %d", len(vms))
	}

	// Read YAML
	data, _ := os.ReadFile(configFile)
	fmt.Printf("YAML with 2 VMs:\n%s\n", string(data))

	// Now edit VM 0
	vm0.Name = "vm-zero-edited"
	vm0.HardDisks = []string{"/dev/sdc"}
	repo.SaveVM(vm0)

	// Read YAML after edit
	data, _ = os.ReadFile(configFile)
	fmt.Printf("YAML after editing VM 0:\n%s\n", string(data))

	// Verify BOTH VMs still exist
	vms, err = repo.ListVMs()
	if err != nil {
		t.Fatalf("ListVMs error: %v", err)
	}
	if len(vms) != 2 {
		t.Errorf("Expected 2 VMs after edit, got %d", len(vms))
		for _, vm := range vms {
			t.Logf("  VM: %s (ID: %s)", vm.Name, vm.ID)
		}
	}
}
