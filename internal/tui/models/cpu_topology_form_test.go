package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// TestCPUTopologyFormInit tests form initialization
func TestCPUTopologyFormInit(t *testing.T) {
	vmManager := createTestVMManager(t)

	form, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if form.focusIndex != 0 {
		t.Errorf("Initial focusIndex = %d, want 0", form.focusIndex)
	}
}

// TestCPUTopologyFormToggle tests toggling core selection
func TestCPUTopologyFormToggle(t *testing.T) {
	vmManager := createTestVMManager(t)

	form, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if form.scanErr != nil {
		t.Skip("Skipping toggle test: CPU scan failed (expected in CI)")
	}

	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Find a toggle position
	found := false
	for i, pos := range form.positions {
		if pos.kind == cpuTopoToggle {
			form.focusIndex = i
			found = true
			break
		}
	}
	if !found {
		t.Skip("No cores to toggle")
	}

	pos := form.currentPos()
	key := coreKey(pos.dieID, pos.coreID)
	initialSelected := form.coreSelected[key]

	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*CPUTopologyFormModel)

	if form.coreSelected[key] == initialSelected {
		t.Error("Toggle did not change selection state")
	}
}

// TestCPUTopologyFormNavigation tests Tab navigation
func TestCPUTopologyFormNavigation(t *testing.T) {
	vmManager := createTestVMManager(t)

	form, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if form.scanErr != nil {
		t.Skip("Skipping navigation test: CPU scan failed (expected in CI)")
	}

	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if form.focusIndex != 0 {
		t.Errorf("Initial focusIndex = %d, want 0", form.focusIndex)
	}

	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*CPUTopologyFormModel)
	if form.focusIndex != 1 {
		t.Errorf("After Tab, focusIndex = %d, want 1", form.focusIndex)
	}

	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	form = model.(*CPUTopologyFormModel)
	if form.focusIndex != 0 {
		t.Errorf("After Shift+Tab, focusIndex = %d, want 0", form.focusIndex)
	}
}

// TestCPUTopologyFormSave tests saving the CPU topology config
func TestCPUTopologyFormSave(t *testing.T) {
	vmManager := createTestVMManager(t)

	form, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if form.scanErr != nil {
		t.Skip("Skipping save test: CPU scan failed (expected in CI)")
	}

	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Ensure at least one core is selected for VM (default has most as VM)
	// Toggle a HOST core to VM if needed
	for i, pos := range form.positions {
		if pos.kind == cpuTopoToggle {
			form.focusIndex = i
			form.Update(tea.KeyMsg{Type: tea.KeySpace})
			break
		}
	}

	// Navigate to save button (last position)
	form.focusIndex = len(form.positions) - 1

	model, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*CPUTopologyFormModel)

	if cmd == nil {
		t.Fatal("Expected command after save, got nil")
	}

	msg := cmd()
	if _, ok := msg.(CPUTopologyUpdatedMsg); !ok {
		t.Errorf("Expected CPUTopologyUpdatedMsg, got %T", msg)
	}

	// Verify saved config
	savedTopo, err := vmManager.GetCPUTopology()
	if err != nil {
		t.Fatalf("Failed to load CPU topology: %v", err)
	}
	if !savedTopo.Enabled {
		t.Errorf("Saved CPUTopology.Enabled = false, want true")
	}
}

// TestCPUTopologyFormDefaultHostCore verifies default init: 1 core as HOST, rest as VM
func TestCPUTopologyFormDefaultHostCore(t *testing.T) {
	vmManager := createTestVMManager(t)

	form, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if form.scanErr != nil {
		t.Skip("Skipping default host core test: CPU scan failed (expected in CI)")
	}

	totalCores := form.hostTopo.TotalCores
	if totalCores <= 1 {
		t.Skip("Skipping: single-core system has no cores to default to VM")
	}

	// Count HOST (unselected) cores — should be exactly 1
	hostCount := 0
	for _, pos := range form.positions {
		if pos.kind == cpuTopoToggle && !form.coreSelected[coreKey(pos.dieID, pos.coreID)] {
			hostCount++
		}
	}

	if hostCount != 1 {
		t.Errorf("Default HOST core count = %d, want 1", hostCount)
	}

	// First core (die 0, core 0) should be the HOST core
	if form.coreSelected[coreKey(0, 0)] {
		t.Errorf("First core (die 0, core 0) should default to HOST, but is VM")
	}
}

// TestCPUTopologyFormZeroHostWarning verifies the warning renders when all cores are VM
func TestCPUTopologyFormZeroHostWarning(t *testing.T) {
	vmManager := createTestVMManager(t)

	form, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if form.scanErr != nil {
		t.Skip("Skipping zero-host warning test: CPU scan failed (expected in CI)")
	}

	form.Update(tea.WindowSizeMsg{Width: 80, Height: 80})

	// Toggle all HOST cores to VM
	for i, pos := range form.positions {
		if pos.kind == cpuTopoToggle {
			key := coreKey(pos.dieID, pos.coreID)
			if !form.coreSelected[key] {
				form.focusIndex = i
				form.Update(tea.KeyMsg{Type: tea.KeySpace})
			}
		}
	}

	// Verify hostCoreCount is 0
	if form.hostCoreCount() != 0 {
		t.Fatalf("Expected hostCoreCount = 0 after toggling all cores to VM, got %d", form.hostCoreCount())
	}

	// Verify the warning text appears in the view
	view := form.View()
	if !strings.Contains(view, "No cores reserved for host") {
		t.Errorf("Expected zero-host warning in view, but it was not found.\nView:\n%s", view)
	}
}

// TestFormatCacheSize tests the formatCacheSize helper
func TestFormatCacheSize(t *testing.T) {
	tests := []struct {
		kb       int
		expected string
	}{
		{32768, "32M"},
		{98304, "96M"},
		{1024, "1M"},
		{1048576, "1G"},
		{512, "512K"},
	}

	for _, tt := range tests {
		result := formatCacheSize(tt.kb)
		if result != tt.expected {
			t.Errorf("formatCacheSize(%d) = %q, want %q", tt.kb, result, tt.expected)
		}
	}
}

// TestCoreKey tests the coreKey helper
func TestCoreKey(t *testing.T) {
	if coreKey(0, 3) != "0:3" {
		t.Errorf("coreKey(0, 3) = %q, want '0:3'", coreKey(0, 3))
	}
	if coreKey(1, 8) != "1:8" {
		t.Errorf("coreKey(1, 8) = %q, want '1:8'", coreKey(1, 8))
	}
}

// TestCPUTopologyModelWrapper tests the CPUTopologyModel wrapper
func TestCPUTopologyModelWrapper(t *testing.T) {
	vmManager := createTestVMManager(t)

	model, err := NewCPUTopologyModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyModel returned error: %v", err)
	}

	if model.form == nil {
		t.Fatal("Expected form to be non-nil")
	}

	_ = model.View()
}

// Ensure vm package is used (required by createTestVMManager)
var _ = vm.NewManager

// Ensure models package is used
var _ = models.CPUTopology{}
