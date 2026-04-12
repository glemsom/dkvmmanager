package vm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// TestSaveAndReloadFromFreshRepository tests saving then reading from a new repository
func TestSaveAndReloadFromFreshRepository(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dkvmmanager")
	os.MkdirAll(configDir, 0755)
	configFile := filepath.Join(configDir, "vms.yaml")

	// First session: create 2 VMs
	repo1, _ := NewRepository(configFile)
	repo1.SaveVM(&models.VM{ID: "0", Name: "vm-zero", HardDisks: []string{"/dev/sda"}, MAC: "aa:bb:cc:dd:ee:ff", VNCListen: "0.0.0.0:0"})
	repo1.SaveVM(&models.VM{ID: "1", Name: "vm-one", HardDisks: []string{"/dev/sdb"}, MAC: "11:22:33:44:55:66", VNCListen: "0.0.0.0:1"})

	vms, _ := repo1.ListVMs()
	if len(vms) != 2 {
		t.Fatalf("Session 1: expected 2 VMs, got %d", len(vms))
	}

	// Second session: fresh repository (simulates app restart)
	repo2, _ := NewRepository(configFile)
	vms, _ = repo2.ListVMs()
	if len(vms) != 2 {
		t.Fatalf("Session 2 (fresh): expected 2 VMs, got %d", len(vms))
	}

	// Edit VM 0 in session 2
	repo2.SaveVM(&models.VM{ID: "0", Name: "vm-zero-edited", HardDisks: []string{"/dev/sdc"}, MAC: "aa:bb:cc:dd:ee:ff", VNCListen: "0.0.0.0:0"})

	vms, _ = repo2.ListVMs()
	if len(vms) != 2 {
		t.Errorf("Session 2 (after edit): expected 2 VMs, got %d", len(vms))
	}

	// Third session: fresh repository again
	repo3, _ := NewRepository(configFile)
	vms, _ = repo3.ListVMs()
	if len(vms) != 2 {
		t.Errorf("Session 3 (fresh): expected 2 VMs, got %d", len(vms))
		for _, vm := range vms {
			t.Logf("  VM: %s (ID: %s)", vm.Name, vm.ID)
		}
	}
}

// TestMultipleSequentialSaves tests rapid sequential saves
func TestMultipleSequentialSaves(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dkvmmanager")
	os.MkdirAll(configDir, 0755)
	configFile := filepath.Join(configDir, "vms.yaml")

	repo, _ := NewRepository(configFile)

	// Create 2 VMs
	repo.SaveVM(&models.VM{ID: "0", Name: "vm-0"})
	repo.SaveVM(&models.VM{ID: "1", Name: "vm-1"})

	// Edit VM 0 ten times rapidly
	for i := 0; i < 10; i++ {
		repo.SaveVM(&models.VM{ID: "0", Name: "vm-0-edited"})
		vms, err := repo.ListVMs()
		if err != nil {
			t.Fatalf("ListVMs error on iteration %d: %v", i, err)
		}
		if len(vms) != 2 {
			t.Errorf("Iteration %d: expected 2 VMs, got %d", i, len(vms))
		}
	}

	// Final check
	vms, _ := repo.ListVMs()
	if len(vms) != 2 {
		t.Errorf("Final: expected 2 VMs, got %d", len(vms))
	}
}

// TestRealYAMLFormat tests with YAML format matching the actual system file
func TestRealYAMLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dkvmmanager")
	os.MkdirAll(configDir, 0755)
	configFile := filepath.Join(configDir, "vms.yaml")

	// Write a YAML file matching the actual system format
	yamlContent := `vms:
    "0":
        cdroms: []
        cpu_opts: ""
        created_at: "2026-03-25T20:13:54+01:00"
        gpu_rom: ""
        harddisks:
            - ""
        id: "0"
        mac: 6a:68:e5:95:d7:5f
        name: test22
        updated_at: "2026-03-27T17:01:48+01:00"
        vnc_listen: 0.0.0.0:0
    "1":
        cdroms: []
        cpu_opts: ""
        created_at: "2026-03-26T19:13:50+01:00"
        gpu_rom: ""
        harddisks: []
        id: "1"
        mac: ca:9b:65:12:ca:db
        name: test33
        updated_at: "2026-03-27T17:00:18+01:00"
        vnc_listen: 0.0.0.0:0
`
	os.WriteFile(configFile, []byte(yamlContent), 0644)

	// Load with repository
	repo, err := NewRepository(configFile)
	if err != nil {
		t.Fatalf("NewRepository error: %v", err)
	}

	vms, err := repo.ListVMs()
	if err != nil {
		t.Fatalf("ListVMs error: %v", err)
	}
	if len(vms) != 2 {
		t.Fatalf("Expected 2 VMs from real YAML, got %d", len(vms))
	}

	// Edit VM 0
	var vm0 *models.VM
	for i := range vms {
		if vms[i].ID == "0" {
			vm0 = &vms[i]
			break
		}
	}
	if vm0 == nil {
		t.Fatal("VM 0 not found")
	}

	vm0.Name = "test22-edited"
	err = repo.SaveVM(vm0)
	if err != nil {
		t.Fatalf("SaveVM error: %v", err)
	}

	// Verify both VMs still exist
	vms, err = repo.ListVMs()
	if err != nil {
		t.Fatalf("ListVMs after edit error: %v", err)
	}
	if len(vms) != 2 {
		t.Errorf("Expected 2 VMs after edit, got %d", len(vms))
		data, _ := os.ReadFile(configFile)
		t.Logf("YAML after edit:\n%s", string(data))
		for _, vm := range vms {
			t.Logf("  VM: %s (ID: %s)", vm.Name, vm.ID)
		}
	}
}
