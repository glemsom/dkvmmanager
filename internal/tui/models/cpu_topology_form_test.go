package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// TestCPUTopologyFormInit tests form initialization
func TestCPUTopologyFormInit(t *testing.T) {
	vmManager := createTestVMManager(t)

	formModel, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if formModel.focusIndex != 0 {
		t.Errorf("Initial focusIndex = %d, want 0", formModel.focusIndex)
	}
}

// TestCPUTopologyFormToggle tests toggling core selection
func TestCPUTopologyFormToggle(t *testing.T) {
	vmManager := createTestVMManager(t)

	formModel, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if formModel.scanErr != nil {
		t.Skip("Skipping toggle test: CPU scan failed (expected in CI)")
	}

	formModel.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Find a toggle position
	found := false
	for i, pos := range formModel.positions {
		if pos.Kind == form.FocusToggle {
			formModel.focusIndex = i
			found = true
			break
		}
	}
	if !found {
		t.Skip("No cores to toggle")
	}

	pos := formModel.currentPos()
	key := coreKey(pos.dieID, pos.coreID)
	initialSelected := formModel.coreSelected[key]

	model, _ := formModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	formModel = model.(*CPUTopologyFormModel)

	if formModel.coreSelected[key] == initialSelected {
		t.Error("Toggle did not change selection state")
	}
}

// TestCPUTopologyFormNavigation tests Tab navigation
func TestCPUTopologyFormNavigation(t *testing.T) {
	vmManager := createTestVMManager(t)

	formModel, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if formModel.scanErr != nil {
		t.Skip("Skipping navigation test: CPU scan failed (expected in CI)")
	}

	formModel.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if formModel.focusIndex != 0 {
		t.Errorf("Initial focusIndex = %d, want 0", formModel.focusIndex)
	}

	model, _ := formModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	formModel = model.(*CPUTopologyFormModel)
	if formModel.focusIndex != 1 {
		t.Errorf("After Tab, focusIndex = %d, want 1", formModel.focusIndex)
	}

	model, _ = formModel.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	formModel = model.(*CPUTopologyFormModel)
	if formModel.focusIndex != 0 {
		t.Errorf("After Shift+Tab, focusIndex = %d, want 0", formModel.focusIndex)
	}
}

// TestCPUTopologyFormSave tests saving the CPU topology config
func TestCPUTopologyFormSave(t *testing.T) {
	vmManager := createTestVMManager(t)

	formModel, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if formModel.scanErr != nil {
		t.Skip("Skipping save test: CPU scan failed (expected in CI)")
	}

	formModel.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Ensure at least one core is selected for VM (default has most as VM)
	// Toggle a HOST core to VM if needed
	for i, pos := range formModel.positions {
		if pos.Kind == form.FocusToggle {
			formModel.focusIndex = i
			formModel.Update(tea.KeyMsg{Type: tea.KeySpace})
			break
		}
	}

	// Navigate to save button (last position)
	formModel.focusIndex = len(formModel.positions) - 1

	model, cmd := formModel.Update(tea.KeyMsg{Type: tea.KeyEnter})
	formModel = model.(*CPUTopologyFormModel)

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

	formModel, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if formModel.scanErr != nil {
		t.Skip("Skipping default host core test: CPU scan failed (expected in CI)")
	}

	totalCores := formModel.hostTopo.TotalCores
	if totalCores <= 1 {
		t.Skip("Skipping: single-core system has no cores to default to VM")
	}

	// Count HOST (unselected) cores — should be exactly 1
	hostCount := 0
	for _, pos := range formModel.positions {
		if pos.Kind == form.FocusToggle {
			d := pos.Data.(cpuTopoFocusData)
			key := coreKey(d.dieID, d.coreID)
			if !formModel.coreSelected[key] {
				hostCount++
			}
		}
	}

	if hostCount != 1 {
		t.Errorf("Default HOST core count = %d, want 1", hostCount)
	}

	// First core (die 0, core 0) should be the HOST core
	if formModel.coreSelected[coreKey(0, 0)] {
		t.Errorf("First core (die 0, core 0) should default to HOST, but is VM")
	}
}

// TestCPUTopologyFormZeroHostWarning verifies the warning renders when all cores are VM
func TestCPUTopologyFormZeroHostWarning(t *testing.T) {
	vmManager := createTestVMManager(t)

	formModel, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	if formModel.scanErr != nil {
		t.Skip("Skipping zero-host warning test: CPU scan failed (expected in CI)")
	}

	formModel.Update(tea.WindowSizeMsg{Width: 80, Height: 80})

	// Toggle all HOST cores to VM
	for i, pos := range formModel.positions {
		if pos.Kind == form.FocusToggle {
			d := pos.Data.(cpuTopoFocusData)
			key := coreKey(d.dieID, d.coreID)
			if !formModel.coreSelected[key] {
				formModel.focusIndex = i
				formModel.Update(tea.KeyMsg{Type: tea.KeySpace})
			}
		}
	}

	// Verify hostCoreCount is 0
	if formModel.hostCoreCount() != 0 {
		t.Fatalf("Expected hostCoreCount = 0 after toggling all cores to VM, got %d", formModel.hostCoreCount())
	}

	// Verify the warning text appears in the view
	view := formModel.View()
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

// TestCPUTopologyFormModelInterface verifies the form implements form.FormModel
func TestCPUTopologyFormModelInterface(t *testing.T) {
	vmManager := createTestVMManager(t)

	formModel, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyFormModel returned error: %v", err)
	}

	// Verify it implements form.FormModel
	var _ form.FormModel = formModel

	// Verify BuildPositions returns expected positions
	positions := formModel.BuildPositions()
	if len(positions) == 0 {
		t.Fatal("Expected at least one position")
	}

	// Last position should be save button
	lastPos := positions[len(positions)-1]
	if lastPos.Kind != form.FocusButton || lastPos.Key != "save" {
		t.Errorf("Expected last position to be save button, got Kind=%v Key=%s", lastPos.Kind, lastPos.Key)
	}
}

// Ensure vm package is used (required by createTestVMManager)
var _ = vm.NewManager

// Ensure models package is used
var _ = models.CPUTopology{}
