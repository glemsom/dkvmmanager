package models

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// TestCPUOptionsFormInit tests form initialization with default options
func TestCPUOptionsFormInit(t *testing.T) {
	vmManager := createTestVMManager(t)
	form := NewCPUOptionsFormModel(vmManager)

	if form.focusIndex != 1 {
		t.Errorf("Initial focusIndex = %d, want 1 (first interactive element after header)", form.focusIndex)
	}

	// Default options should all be false/empty
	if form.options.HideKVM {
		t.Errorf("Default HideKVM should be false")
	}
	if form.options.VendorID != "" {
		t.Errorf("Default VendorID should be empty")
	}
}

// TestCPUOptionsFormToggle tests toggle behavior
func TestCPUOptionsFormToggle(t *testing.T) {
	vmManager := createTestVMManager(t)
	form := NewCPUOptionsFormModel(vmManager)

	// Focus is on HideKVM (index 1, after header at 0)
	pos := form.currentPos()
	if pos.fieldName != "HideKVM" {
		t.Fatalf("Expected current position to be HideKVM, got %s", pos.fieldName)
	}
	if pos.kind != cpuOptToggle {
		t.Fatalf("Expected current position to be toggle, got %d", pos.kind)
	}

	// Toggle HideKVM via Enter key
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*CPUOptionsFormModel)

	if !form.options.HideKVM {
		t.Errorf("HideKVM should be true after toggle")
	}

	// Toggle again
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*CPUOptionsFormModel)

	if form.options.HideKVM {
		t.Errorf("HideKVM should be false after second toggle")
	}
}

// TestCPUOptionsFormNavigation tests Tab navigation between fields
func TestCPUOptionsFormNavigation(t *testing.T) {
	vmManager := createTestVMManager(t)
	form := NewCPUOptionsFormModel(vmManager)

	// Start at HideKVM (index 1, after header at 0)
	if form.currentPos().fieldName != "HideKVM" {
		t.Fatalf("Expected initial field HideKVM, got %s", form.currentPos().fieldName)
	}

	// Tab moves to VendorID (index 2)
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*CPUOptionsFormModel)
	if form.currentPos().fieldName != "VendorID" {
		t.Errorf("Expected VendorID after Tab, got %s", form.currentPos().fieldName)
	}

	// Tab moves to section header (index 3)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*CPUOptionsFormModel)
	if form.currentPos().fieldName != "header_hyperv" {
		t.Errorf("Expected header_hyperv after Tab, got %s", form.currentPos().fieldName)
	}

	// Tab moves to HVFrequency (index 4)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*CPUOptionsFormModel)
	if form.currentPos().fieldName != "HVFrequency" {
		t.Errorf("Expected HVFrequency after Tab, got %s", form.currentPos().fieldName)
	}

	// Shift+Tab moves back to header_hyperv
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	form = model.(*CPUOptionsFormModel)
	if form.currentPos().fieldName != "header_hyperv" {
		t.Errorf("Expected header_hyperv after Shift+Tab, got %s", form.currentPos().fieldName)
	}
}

// TestCPUOptionsFormTextEditing tests text field editing
func TestCPUOptionsFormTextEditing(t *testing.T) {
	vmManager := createTestVMManager(t)
	form := NewCPUOptionsFormModel(vmManager)

	// Move to VendorID (index 2)
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*CPUOptionsFormModel)

	// Type "AMD"
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	form = model.(*CPUOptionsFormModel)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}})
	form = model.(*CPUOptionsFormModel)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	form = model.(*CPUOptionsFormModel)

	if form.options.VendorID != "AMD" {
		t.Errorf("VendorID = %q, want AMD", form.options.VendorID)
	}

	// Backspace removes last character
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	form = model.(*CPUOptionsFormModel)

	if form.options.VendorID != "AM" {
		t.Errorf("VendorID = %q, want AM after backspace", form.options.VendorID)
	}
}

// TestCPUOptionsFormSave tests saving via Enter on save button
func TestCPUOptionsFormSave(t *testing.T) {
	vmManager := createTestVMManager(t)
	form := NewCPUOptionsFormModel(vmManager)

	// Set some options
	form.options.HideKVM = true
	form.options.HVRelaxed = true
	form.options.VendorID = "TestVendor"

	// Navigate to save button (last position)
	form.focusIndex = len(form.positions) - 1

	// Press Enter to save
	model, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*CPUOptionsFormModel)

	// Command should return CPUOptionsUpdatedMsg
	if cmd == nil {
		t.Fatal("Expected command after save, got nil")
	}

	msg := cmd()
	if _, ok := msg.(CPUOptionsUpdatedMsg); !ok {
		t.Errorf("Expected CPUOptionsUpdatedMsg, got %T", msg)
	}

	// Verify saved options
	saved, _ := vmManager.GetCPUOptions()
	if !saved.HideKVM {
		t.Errorf("Saved HideKVM = false, want true")
	}
	if !saved.HVRelaxed {
		t.Errorf("Saved HVRelaxed = false, want true")
	}
	if saved.VendorID != "TestVendor" {
		t.Errorf("Saved VendorID = %q, want TestVendor", saved.VendorID)
	}
}

// TestCPUOptionsFormAllToggles tests toggling all boolean fields
func TestCPUOptionsFormAllToggles(t *testing.T) {
	vmManager := createTestVMManager(t)
	form := NewCPUOptionsFormModel(vmManager)

	toggleFields := []string{
		"HideKVM", "HVFrequency", "HVRelaxed", "HVReset", "HVRuntime",
		"HVStimer", "HVSyncIC", "HVTime", "HVVapic", "HVVPIndex",
		"HVNoNonarchCoresharing", "HVTLBFlush", "HVTLBFlushExt", "HVIPI",
		"HVAVIC", "TopoExt", "L3Cache", "X2APIC", "Migratable", "InvTSC",
		"RTCUTC", "CPUPM",
	}

	for _, field := range toggleFields {
		form.focusIndex = form.findIndexByName(field)
		if form.focusIndex < 0 {
			t.Errorf("Field %s not found in positions", field)
			continue
		}

		// Get initial value
		initialVal := form.getToggleValue(field)

		// Toggle
		model, _ := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
		form = model.(*CPUOptionsFormModel)

		newVal := form.getToggleValue(field)
		if newVal == initialVal {
			t.Errorf("Field %s did not toggle", field)
		}
	}
}

// findIndexByName returns the focus index for a field name
func (m *CPUOptionsFormModel) findIndexByName(name string) int {
	for i, p := range m.positions {
		if p.Key == name {
			return i
		}
	}
	return -1
}

// TestCPUOptionsUpdateMsg tests that Update returns correct types
func TestCPUOptionsUpdateMsg(t *testing.T) {
	form := &CPUOptionsFormModel{
		options:       models.CPUOptions{},
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	form.positions = form.BuildPositions()

	// WindowSizeMsg initializes viewport
	model, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	form = model.(*CPUOptionsFormModel)

	if !form.ready {
		t.Errorf("Form should be ready after WindowSizeMsg")
	}
}

// TestCPUOptionsFieldLabels tests that all fields have labels
func TestCPUOptionsFieldLabels(t *testing.T) {
	form := &CPUOptionsFormModel{}

	fields := []string{
		"HideKVM", "VendorID", "HVFrequency", "HVRelaxed", "HVReset",
		"HVRuntime", "HVSpinlocks", "HVStimer", "HVSyncIC", "HVTime",
		"HVVapic", "HVVPIndex", "HVNoNonarchCoresharing", "HVTLBFlush",
		"HVTLBFlushExt", "HVIPI", "HVAVIC", "TopoExt", "L3Cache",
		"X2APIC", "Migratable", "InvTSC", "RTCUTC", "CPUPM",
	}

	for _, field := range fields {
		label := form.fieldLabel(field)
		if label == field {
			t.Errorf("Field %s has no human-readable label", field)
		}
	}
}

// TestCPUOptionsPositionsCount tests that BuildPositions creates correct number of positions
func TestCPUOptionsPositionsCount(t *testing.T) {
	form := &CPUOptionsFormModel{
		options:       models.CPUOptions{},
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	form.positions = form.BuildPositions()

	// 22 toggles + 2 text fields + 1 save button + 3 section headers = 28
	expectedCount := 28
	if len(form.positions) != expectedCount {
		t.Errorf("Expected %d positions, got %d", expectedCount, len(form.positions))
	}
}

// createTestVMManager creates a temporary VM manager for testing
func createTestVMManager(t *testing.T) *vm.Manager {
	t.Helper()

	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "data")
	os.MkdirAll(dataDir, 0755)

	cfg := &config.Config{
		DataFolder:    dataDir,
		VMsConfigFile: filepath.Join(dataDir, "vms.yaml"),
	}

	mgr, err := vm.NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create test VM manager: %v", err)
	}

	return mgr
}
