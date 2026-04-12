package models

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
)

// setupTestModel creates a MainModel with a temporary config directory
func setupTestModel(t *testing.T) *MainModel {
	t.Helper()

	tmpDir := t.TempDir()

	// Create required subdirectories
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

	return m
}

// setupTestModelWithVMs creates a MainModel with pre-existing VMs
func setupTestModelWithVMs(t *testing.T) *MainModel {
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

	// Add VMs with deterministic MAC addresses for test stability
	vm1, err := m.vmManager.CreateVMWithMAC("test-vm-1", "9a:9b:1d:a6:65:c8")
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	_ = vm1

	vm2, err := m.vmManager.CreateVMWithMAC("test-vm-2", "12:bc:7d:b5:49:51")
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}
	_ = vm2

	// Rebuild menu to include VMs
	m.rebuildMenuList()

	return m
}

func TestNewMainModelWithConfig(t *testing.T) {
	m := setupTestModel(t)

	if m == nil {
		t.Fatal("NewMainModelWithConfig returned nil")
	}

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected initial view to be ViewMainMenu, got %s", m.currentView)
	}

	if m.cfg == nil {
		t.Error("Expected cfg to be set")
	}

	if m.vmManager == nil {
		t.Error("Expected vmManager to be set")
	}

	if m.tabModel == nil {
		t.Error("Expected tabModel to be set")
	}

	if m.statusBar == nil {
		t.Error("Expected statusBar to be set")
	}

	if m.breadcrumbs == nil {
		t.Error("Expected breadcrumbs to be set")
	}
}

func TestMainModelMenuListInitialized(t *testing.T) {
	m := setupTestModelWithVMs(t)

	items := m.menuList.Items()
	// Should have 2 VMs only (Config/Power/Shell removed — handled by tabs)
	if len(items) != 2 {
		t.Errorf("Expected 2 menu items (VMs only), got %d", len(items))
	}
}

func TestMainModelConfigListInitialized(t *testing.T) {
	m := setupTestModel(t)

	items := m.configList.Items()
	if len(items) != 11 {
		t.Errorf("Expected 11 config menu items, got %d", len(items))
	}

	// Check first item
	if items[0].FilterValue() != "Add new VM" {
		t.Errorf("Expected first config item to be 'Add new VM', got '%s'", items[0].FilterValue())
	}
}

func TestMainModelInit(t *testing.T) {
	m := setupTestModel(t)

	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil command")
	}
}

func TestSetDebugMode(t *testing.T) {
	// Just verify it doesn't panic
	SetDebugMode(true)
	SetDebugMode(false)
}

func TestMenuItemStruct(t *testing.T) {
	item := MenuItem{
		Title:    "Test VM",
		Type:     "VM",
		VMID:     "0",
		Disabled: false,
	}

	if item.Title != "Test VM" {
		t.Errorf("Expected title 'Test VM', got '%s'", item.Title)
	}
	if item.Type != "VM" {
		t.Errorf("Expected type 'VM', got '%s'", item.Type)
	}
	if item.VMID != "0" {
		t.Errorf("Expected VMID '0', got '%s'", item.VMID)
	}
	if item.Disabled {
		t.Error("Expected Disabled to be false")
	}
}

func TestViewConstants(t *testing.T) {
	// Verify view constants are distinct
	views := map[string]bool{
		ViewMainMenu:      true,
		ViewConfigMenu:    true,
		ViewVMMenu:        true,
		ViewLogViewer:     true,
		ViewFirstRunSetup: true,
		ViewPowerMenu:     true,
		ViewVMCreate:      true,
		ViewVMEdit:        true,
		ViewVMSelect:      true,
		ViewVMDelete:      true,
	}

	if len(views) != 10 {
		t.Errorf("Expected 10 unique view constants, got %d", len(views))
	}
}

func TestVMDeletedMsgStruct(t *testing.T) {
	msg := VMDeletedMsg{
		VMName: "test-vm",
		VMID:   "42",
	}

	if msg.VMName != "test-vm" {
		t.Errorf("Expected VMName 'test-vm', got '%s'", msg.VMName)
	}
	if msg.VMID != "42" {
		t.Errorf("Expected VMID '42', got '%s'", msg.VMID)
	}
}

// Ensure the VM model used in tests is valid
var _ models.VM = models.VM{}
