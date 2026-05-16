// Package vm provides virtual machine management functionality
package vm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

func TestGetCPUOptionsDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dkvmmanager")
	os.MkdirAll(configDir, 0755)
	configFile := filepath.Join(configDir, "vms.yaml")

	repo, err := NewRepository(configFile)
	if err != nil {
		t.Fatalf("NewRepository error: %v", err)
	}

	var opts models.CPUOptions
	if err := repo.GetConfig("cpu_options", &opts); err != nil {
		t.Fatalf("GetConfig error: %v", err)
	}

	// All defaults should be false/empty
	if opts.HideKVM {
		t.Errorf("HideKVM default should be false")
	}
	if opts.VendorID != "" {
		t.Errorf("VendorID default should be empty, got %q", opts.VendorID)
	}
	if opts.HVSpinlocks != "" {
		t.Errorf("HVSpinlocks default should be empty, got %q", opts.HVSpinlocks)
	}
}

func TestSaveAndLoadCPUOptions(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dkvmmanager")
	os.MkdirAll(configDir, 0755)
	configFile := filepath.Join(configDir, "vms.yaml")

	repo, err := NewRepository(configFile)
	if err != nil {
		t.Fatalf("NewRepository error: %v", err)
	}

	// Save CPU options
	opts := models.CPUOptions{
		HideKVM:     true,
		VendorID:    "AuthenticAMD",
		HVRelaxed:   true,
		HVTime:      true,
		HVSpinlocks: "0x1fff",
		X2APIC:      true,
	}

	if err := repo.SaveConfig("cpu_options", opts); err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}

	// Load CPU options
	var loaded models.CPUOptions
	if err := repo.GetConfig("cpu_options", &loaded); err != nil {
		t.Fatalf("GetConfig error: %v", err)
	}

	if loaded.HideKVM != true {
		t.Errorf("HideKVM = %v, want true", loaded.HideKVM)
	}
	if loaded.VendorID != "AuthenticAMD" {
		t.Errorf("VendorID = %q, want AuthenticAMD", loaded.VendorID)
	}
	if loaded.HVRelaxed != true {
		t.Errorf("HVRelaxed = %v, want true", loaded.HVRelaxed)
	}
	if loaded.HVTime != true {
		t.Errorf("HVTime = %v, want true", loaded.HVTime)
	}
	if loaded.HVSpinlocks != "0x1fff" {
		t.Errorf("HVSpinlocks = %q, want 0x1fff", loaded.HVSpinlocks)
	}
	if loaded.X2APIC != true {
		t.Errorf("X2APIC = %v, want true", loaded.X2APIC)
	}
	// Fields not set should remain at defaults
	if loaded.HVFrequency {
		t.Errorf("HVFrequency = %v, want false", loaded.HVFrequency)
	}
}

func TestCPUOptionsRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dkvmmanager")
	os.MkdirAll(configDir, 0755)
	configFile := filepath.Join(configDir, "vms.yaml")

	repo1, _ := NewRepository(configFile)

	// Save all options set to true
	opts := models.CPUOptions{
		HideKVM:                true,
		VendorID:               "GenuineIntel",
		HVFrequency:            true,
		HVRelaxed:              true,
		HVReset:                true,
		HVRuntime:              true,
		HVSpinlocks:            "0x1000",
		HVStimer:               true,
		HVSyncIC:               true,
		HVTime:                 true,
		HVVapic:                true,
		HVVPIndex:              true,
		HVNoNonarchCoresharing: true,
		HVTLBFlush:             true,
		HVTLBFlushExt:          true,
		HVIPI:                  true,
		HVAVIC:                 true,
		TopoExt:                true,
		L3Cache:                true,
		X2APIC:                 true,
		Migratable:             false,
		InvTSC:                 true,
	}

	if err := repo1.SaveConfig("cpu_options", opts); err != nil {
		t.Fatalf("SaveConfig error: %v", err)
	}

	// Fresh repository (simulates app restart)
	repo2, _ := NewRepository(configFile)
	var loaded models.CPUOptions
	if err := repo2.GetConfig("cpu_options", &loaded); err != nil {
		t.Fatalf("GetConfig error: %v", err)
	}

	if loaded.VendorID != "GenuineIntel" {
		t.Errorf("VendorID = %q, want GenuineIntel", loaded.VendorID)
	}
	if loaded.HVSpinlocks != "0x1000" {
		t.Errorf("HVSpinlocks = %q, want 0x1000", loaded.HVSpinlocks)
	}
	if !loaded.HideKVM {
		t.Errorf("HideKVM = false, want true")
	}
	if !loaded.HVRelaxed {
		t.Errorf("HVRelaxed = false, want true")
	}
	if loaded.Migratable {
		t.Errorf("Migratable = true, want false")
	}
}

func TestCPUOptionsCoexistsWithVMs(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dkvmmanager")
	os.MkdirAll(configDir, 0755)
	configFile := filepath.Join(configDir, "vms.yaml")

	repo, _ := NewRepository(configFile)

	// Save a VM
	repo.SaveVM(&models.VM{ID: "0", Name: "test-vm"})

	// Save CPU options
	opts := models.CPUOptions{HideKVM: true, HVRelaxed: true}
	repo.SaveConfig("cpu_options", opts)

	// Verify both exist
	vms, _ := repo.ListVMs()
	if len(vms) != 1 {
		t.Errorf("Expected 1 VM, got %d", len(vms))
	}

	var loaded models.CPUOptions
	repo.GetConfig("cpu_options", &loaded)
	if !loaded.HideKVM {
		t.Errorf("CPU options not preserved after saving VM")
	}
	if !loaded.HVRelaxed {
		t.Errorf("CPU options not preserved after saving VM")
	}

	// Save another VM and verify CPU options persist
	repo.SaveVM(&models.VM{ID: "1", Name: "test-vm-2"})

	vms, _ = repo.ListVMs()
	if len(vms) != 2 {
		t.Errorf("Expected 2 VMs, got %d", len(vms))
	}

	repo.GetConfig("cpu_options", &loaded)
	if !loaded.HideKVM {
		t.Errorf("CPU options lost after saving second VM")
	}
}
