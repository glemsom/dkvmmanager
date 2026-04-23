// Package vm provides tests for virtual machine management
package vm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
)

func TestGenerateMAC(t *testing.T) {
	mac := generateMAC()
	if len(mac) != 17 { // xx:xx:xx:xx:xx:xx
		t.Errorf("generateMAC() returned %q, expected format xx:xx:xx:xx:xx:xx", mac)
	}

	// Verify format: should be xx:xx:xx:xx:xx:xx
	for i, c := range mac {
		if i%3 == 2 {
			if c != ':' {
				t.Errorf("generateMAC() has invalid separator at position %d: %q", i, mac)
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("generateMAC() has invalid hex char at position %d: %q", i, mac)
			}
		}
	}

	// Verify local bit is set (second hex digit should have 2, 6, a, or e)
	if mac[1] != '2' && mac[1] != '6' && mac[1] != 'a' && mac[1] != 'e' {
		t.Errorf("generateMAC() local bit not set: %q", mac)
	}
}

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if mgr == nil {
		t.Error("NewManager() returned nil")
	}
	if mgr.cfg != cfg {
		t.Error("NewManager() did not set config")
	}
}

func TestListVMsEmptyDataFolder(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	vms, err := mgr.ListVMs()

	if err != nil {
		t.Errorf("ListVMs() returned error: %v", err)
	}
	if len(vms) != 0 {
		t.Errorf("ListVMs() returned %d VMs, want 0", len(vms))
	}
}

func TestListVMsWithMultipleDisks(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Create VM using new API
	vm := &models.VM{
		ID:        "0",
		Name:      "test-vm",
		HardDisks: []string{"/dev/sda", "/dev/sdb"},
		CDROMs:    []string{"/iso1.iso", "/iso2.iso"},
	}
	if err := mgr.SaveVM(vm); err != nil {
		t.Fatalf("SaveVM() returned error: %v", err)
	}

	vms, err := mgr.ListVMs()

	if err != nil {
		t.Errorf("ListVMs() returned error: %v", err)
	}
	if len(vms) != 1 {
		t.Fatalf("ListVMs() returned %d VMs, want 1", len(vms))
	}

	if len(vms[0].HardDisks) != 2 {
		t.Errorf("VM.HardDisks length = %v, want 2", len(vms[0].HardDisks))
	}
	if vms[0].HardDisks[0] != "/dev/sda" {
		t.Errorf("VM.HardDisks[0] = %v, want /dev/sda", vms[0].HardDisks[0])
	}
	if vms[0].HardDisks[1] != "/dev/sdb" {
		t.Errorf("VM.HardDisks[1] = %v, want /dev/sdb", vms[0].HardDisks[1])
	}

	if len(vms[0].CDROMs) != 2 {
		t.Errorf("VM.CDROMs length = %v, want 2", len(vms[0].CDROMs))
	}
}

func TestCreateVM(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	vm, err := mgr.CreateVM("new-vm")

	if err != nil {
		t.Errorf("CreateVM() returned error: %v", err)
	}
	if vm == nil {
		t.Fatal("CreateVM() returned nil VM")
	}
	if vm.Name != "new-vm" {
		t.Errorf("VM.Name = %v, want new-vm", vm.Name)
	}
	if vm.ID != "0" {
		t.Errorf("VM.ID = %v, want 0", vm.ID)
	}
	if vm.MAC == "" {
		t.Error("VM.MAC should not be empty")
	}

	// Verify the config file was created (YAML format)
	configFile := filepath.Join(tmpDir, "dkvmmanager", "vms.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("VM config file was not created")
	}
}

func TestCreateVMMultiple(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Create first VM
	vm1, err := mgr.CreateVM("vm-1")
	if err != nil {
		t.Fatalf("CreateVM() returned error: %v", err)
	}
	if vm1.ID != "0" {
		t.Errorf("First VM ID = %v, want 0", vm1.ID)
	}

	// Create second VM
	vm2, err := mgr.CreateVM("vm-2")
	if err != nil {
		t.Fatalf("CreateVM() returned error: %v", err)
	}
	if vm2.ID != "1" {
		t.Errorf("Second VM ID = %v, want 1", vm2.ID)
	}
}

func TestDeleteVM(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Create a VM first
	vm, err := mgr.CreateVM("to-delete")
	if err != nil {
		t.Fatalf("CreateVM() returned error: %v", err)
	}

	// Delete the VM
	err = mgr.DeleteVM(vm.ID)
	if err != nil {
		t.Errorf("DeleteVM() returned error: %v", err)
	}

	// Verify the VM is no longer in the list
	vms, err := mgr.ListVMs()
	if err != nil {
		t.Errorf("ListVMs() returned error: %v", err)
	}
	if len(vms) != 0 {
		t.Errorf("Expected 0 VMs after deletion, got %d", len(vms))
	}
}

func TestDeleteVMPreservesDataDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Create a VM first
	vm, err := mgr.CreateVM("to-delete")
	if err != nil {
		t.Fatalf("CreateVM() returned error: %v", err)
	}

	// Create a test file in the VM directory to verify it's preserved
	vmDir := filepath.Join(tmpDir, "vms", vm.ID)
	testFile := filepath.Join(vmDir, "test-disk.qcow2")
	if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete the VM
	err = mgr.DeleteVM(vm.ID)
	if err != nil {
		t.Errorf("DeleteVM() returned error: %v", err)
	}

	// Verify the VM is no longer in the list
	vms, err := mgr.ListVMs()
	if err != nil {
		t.Errorf("ListVMs() returned error: %v", err)
	}
	if len(vms) != 0 {
		t.Errorf("Expected 0 VMs after deletion, got %d", len(vms))
	}

	// Verify the VM directory still exists
	if _, err := os.Stat(vmDir); os.IsNotExist(err) {
		t.Error("VM directory was deleted (should be preserved)")
	}

	// Verify the test file still exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Test file in VM directory was deleted (should be preserved)")
	}
}

func TestGetVMNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, err = mgr.GetVM("999")

	if err == nil {
		t.Error("GetVM() should return error for non-existent VM")
	}
}

func TestSaveVMConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}

	vm := &models.VM{
		ID:        "7",
		Name:      "save-test-vm",
		HardDisks: []string{"/dev/sda", "/dev/sdb"},
		CDROMs:    []string{"/iso.iso"},
		MAC:       "11:22:33:44:55:66",
		VNCListen: "0.0.0.0:1",
		GPUROM:    "/rom/gpu.rom",
	}

	// Create VM directory before saving (saveVMConfig does not create it)
	vmDir := filepath.Join(tmpDir, "7")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatalf("Failed to create VM directory: %v", err)
	}

	err = mgr.SaveVM(vm)
	if err != nil {
		t.Errorf("saveVMConfig() returned error: %v", err)
	}

	// Verify the file was created (YAML format)
	configFile := filepath.Join(tmpDir, "dkvmmanager", "vms.yaml")
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read VM config: %v", err)
	}

	content := string(data)
	if content == "" {
		t.Error("VM config file is empty")
	}
}

func TestNetworkModePersistence(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Save a VM with network_mode set to "bridge"
	vm := &models.VM{
		ID:          "0",
		Name:        "net-test-vm",
		MAC:         "aa:bb:cc:dd:ee:ff",
		NetworkMode: "bridge",
		VNCListen:   "0.0.0.0:0",
	}
	if err := mgr.SaveVM(vm); err != nil {
		t.Fatalf("SaveVM() returned error: %v", err)
	}

	// Load the VM back and verify network_mode persisted
	loaded, err := mgr.GetVM("0")
	if err != nil {
		t.Fatalf("GetVM() returned error: %v", err)
	}
	if loaded.NetworkMode != "bridge" {
		t.Errorf("VM.NetworkMode = %q, want %q", loaded.NetworkMode, "bridge")
	}

	// Also test "nat" mode
	vm.NetworkMode = "nat"
	if err := mgr.SaveVM(vm); err != nil {
		t.Fatalf("SaveVM() returned error: %v", err)
	}

	loaded, err = mgr.GetVM("0")
	if err != nil {
		t.Fatalf("GetVM() returned error: %v", err)
	}
	if loaded.NetworkMode != "nat" {
		t.Errorf("VM.NetworkMode = %q, want %q", loaded.NetworkMode, "nat")
	}
}

func TestNetworkModePersistenceRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "dkvmmanager", "vms.yaml")

	// Save via repository
	repo1, err := NewRepository(configFile)
	if err != nil {
		t.Fatal(err)
	}

	vm := &models.VM{
		ID:          "0",
		Name:        "roundtrip-vm",
		NetworkMode: "bridge",
		MAC:         "11:22:33:44:55:66",
	}
	if err := repo1.SaveVM(vm); err != nil {
		t.Fatalf("SaveVM() error: %v", err)
	}

	// Load via a fresh repository (simulates app restart)
	repo2, err := NewRepository(configFile)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := repo2.GetVM("0")
	if err != nil {
		t.Fatalf("GetVM() error: %v", err)
	}
	if loaded.NetworkMode != "bridge" {
		t.Errorf("NetworkMode = %q after round-trip, want %q", loaded.NetworkMode, "bridge")
	}
}
