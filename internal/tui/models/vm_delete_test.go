// Package models provides tests for the VM deletion model
package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// setupDeleteTest creates a VM manager with a single VM for delete testing
func setupDeleteTest(t *testing.T) (*vm.Manager, string) {
	t.Helper()

	tmpDir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(tmpDir, "dkvmmanager"), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := vm.NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create VM manager: %v", err)
	}

	createdVM, err := mgr.CreateVM("test-vm")
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}

	return mgr, createdVM.ID
}

func TestNewVMDeleteModel(t *testing.T) {
	mgr, vmID := setupDeleteTest(t)

	model, err := NewVMDeleteModel(mgr, vmID)
	if err != nil {
		t.Fatalf("NewVMDeleteModel() returned error: %v", err)
	}
	if model == nil {
		t.Fatal("NewVMDeleteModel() returned nil")
	}
	if model.vm == nil {
		t.Error("Expected vm to be set")
	}
	if model.vm.ID != vmID {
		t.Errorf("Expected VM ID %s, got %s", vmID, model.vm.ID)
	}
	if model.vm.Name != "test-vm" {
		t.Errorf("Expected VM name 'test-vm', got '%s'", model.vm.Name)
	}
	if model.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", model.selectedIndex)
	}
}

func TestNewVMDeleteModelNotFound(t *testing.T) {
	mgr, _ := setupDeleteTest(t)

	model, err := NewVMDeleteModel(mgr, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent VM")
	}
	if model != nil {
		t.Error("Expected nil model for nonexistent VM")
	}
}

func TestVMDeleteModelInit(t *testing.T) {
	mgr, vmID := setupDeleteTest(t)

	model, err := NewVMDeleteModel(mgr, vmID)
	if err != nil {
		t.Fatalf("NewVMDeleteModel() returned error: %v", err)
	}

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init() should return nil command")
	}
}

func TestVMDeleteModelView(t *testing.T) {
	mgr, vmID := setupDeleteTest(t)

	model, err := NewVMDeleteModel(mgr, vmID)
	if err != nil {
		t.Fatalf("NewVMDeleteModel() returned error: %v", err)
	}

	view := model.View()
	if !strings.Contains(view, "WARNING") {
		t.Error("View should contain warning message")
	}
	if !strings.Contains(view, "test-vm") {
		t.Error("View should contain VM name")
	}
	if !strings.Contains(view, "No") {
		t.Error("View should contain 'No' option")
	}
	if !strings.Contains(view, "Yes") {
		t.Error("View should contain 'Yes' option")
	}
}

func TestVMDeleteModelSelectNo(t *testing.T) {
	mgr, vmID := setupDeleteTest(t)

	model, err := NewVMDeleteModel(mgr, vmID)
	if err != nil {
		t.Fatalf("NewVMDeleteModel() returned error: %v", err)
	}

	// selectedIndex is 0 (No) by default, press Enter
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = updatedModel

	if cmd == nil {
		t.Fatal("Expected command after pressing Enter on No")
	}

	msg := cmd()
	vcm, ok := msg.(ViewChangeMsg)
	if !ok {
		t.Fatalf("Expected ViewChangeMsg, got %T", msg)
	}
	if vcm.View != ViewConfigMenu {
		t.Errorf("Expected view %s, got %s", ViewConfigMenu, vcm.View)
	}
}

func TestVMDeleteModelSelectYes(t *testing.T) {
	mgr, vmID := setupDeleteTest(t)

	model, err := NewVMDeleteModel(mgr, vmID)
	if err != nil {
		t.Fatalf("NewVMDeleteModel() returned error: %v", err)
	}

	// Navigate to Yes (index 1)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(*VMDeleteModel)

	// Press Enter on Yes
	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = updatedModel

	if cmd == nil {
		t.Fatal("Expected command after pressing Enter on Yes")
	}

	msg := cmd()
	vdm, ok := msg.(VMDeletedMsg)
	if !ok {
		t.Fatalf("Expected VMDeletedMsg, got %T", msg)
	}
	if vdm.VMName != "test-vm" {
		t.Errorf("Expected VM name 'test-vm', got '%s'", vdm.VMName)
	}
	if vdm.VMID != vmID {
		t.Errorf("Expected VM ID '%s', got '%s'", vmID, vdm.VMID)
	}

	// Verify VM is actually deleted
	_, err = mgr.GetVM(vmID)
	if err == nil {
		t.Error("Expected VM to be deleted from manager")
	}
}

func TestVMDeleteModelNavigation(t *testing.T) {
	mgr, vmID := setupDeleteTest(t)

	model, err := NewVMDeleteModel(mgr, vmID)
	if err != nil {
		t.Fatalf("NewVMDeleteModel() returned error: %v", err)
	}

	// Initially at index 0 (No)
	if model.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", model.selectedIndex)
	}

	// Press down to go to Yes
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	m := updated.(*VMDeleteModel)
	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after down, got %d", m.selectedIndex)
	}

	// Press up to go back to No
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(*VMDeleteModel)
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after up, got %d", m.selectedIndex)
	}

	// Press up again - should stay at 0
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(*VMDeleteModel)
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 (bounded), got %d", m.selectedIndex)
	}
}
