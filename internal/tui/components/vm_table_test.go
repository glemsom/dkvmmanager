package components

import (
	"strings"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

func TestNewVMTable(t *testing.T) {
	vms := []models.VM{
		{ID: "vm-001", Name: "Test VM 1", MAC: "52:54:00:00:00:01"},
		{ID: "vm-002", Name: "Test VM 2", MAC: "52:54:00:00:00:02"},
	}

	table := NewVMTable(vms, 80, 10)

	if table == nil {
		t.Fatal("NewVMTable returned nil")
	}

	if len(table.vms) != 2 {
		t.Errorf("Expected 2 VMs, got %d", len(table.vms))
	}
}

func TestVMTableView(t *testing.T) {
	vms := []models.VM{
		{ID: "vm-001", Name: "Test VM", MAC: "52:54:00:00:00:01"},
	}

	table := NewVMTable(vms, 80, 10)
	view := table.View()

	if view == "" {
		t.Error("View() returned empty string")
	}

	if !strings.Contains(view, "Test VM") {
		t.Error("View() does not contain VM name")
	}

	if !strings.Contains(view, "vm-001") {
		t.Error("View() does not contain VM ID")
	}
}

func TestVMTableSelectedVM(t *testing.T) {
	vms := []models.VM{
		{ID: "vm-001", Name: "VM One"},
		{ID: "vm-002", Name: "VM Two"},
	}

	table := NewVMTable(vms, 80, 10)

	selected := table.SelectedVM()
	if selected == nil {
		t.Fatal("SelectedVM() returned nil for non-empty table")
	}

	if selected.ID != "vm-001" {
		t.Errorf("Expected first VM (vm-001), got %s", selected.ID)
	}
}

func TestVMTableSelectedVMEmpty(t *testing.T) {
	table := NewVMTable([]models.VM{}, 80, 10)

	selected := table.SelectedVM()
	if selected != nil {
		t.Error("SelectedVM() should return nil for empty table")
	}
}

func TestVMTableSetVMs(t *testing.T) {
	initial := []models.VM{
		{ID: "vm-001", Name: "Initial VM"},
	}

	table := NewVMTable(initial, 80, 10)

	updated := []models.VM{
		{ID: "vm-002", Name: "Updated VM 1"},
		{ID: "vm-003", Name: "Updated VM 2"},
	}

	table.SetVMs(updated)

	if len(table.vms) != 2 {
		t.Errorf("Expected 2 VMs after SetVMs, got %d", len(table.vms))
	}

	view := table.View()
	if !strings.Contains(view, "Updated VM 1") {
		t.Error("View() does not contain updated VM name")
	}
}

func TestVMTableSetSize(t *testing.T) {
	vms := []models.VM{
		{ID: "vm-001", Name: "Test VM"},
	}

	table := NewVMTable(vms, 80, 10)

	// Should not panic
	table.SetSize(120, 20)
	table.SetSize(40, 5)
}

func TestVMTableCursor(t *testing.T) {
	vms := []models.VM{
		{ID: "vm-001", Name: "VM One"},
		{ID: "vm-002", Name: "VM Two"},
	}

	table := NewVMTable(vms, 80, 10)

	if table.Cursor() != 0 {
		t.Errorf("Expected initial cursor 0, got %d", table.Cursor())
	}
}

func TestVMTableDisksColumn(t *testing.T) {
	vms := []models.VM{
		{ID: "vm-001", Name: "Test VM", HardDisks: []string{"/dev/sda", "/dev/sdb"}},
	}

	table := NewVMTable(vms, 80, 10)
	view := table.View()

	if !strings.Contains(view, "2") {
		t.Error("View() should show disk count of 2")
	}
}

func TestVMTableEmptyMAC(t *testing.T) {
	vms := []models.VM{
		{ID: "vm-001", Name: "Test VM"},
	}

	table := NewVMTable(vms, 80, 10)
	view := table.View()

	if !strings.Contains(view, "-") {
		t.Error("View() should show '-' for empty MAC")
	}
}
