package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/fields"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// TestCPUOptionsFieldRegistry tests the central field registry
func TestCPUOptionsFieldRegistry(t *testing.T) {
	// Verify all fields in CPUOptions are represented in the registry
	expectedFields := []struct {
		name string
		kind fields.FieldKind
	}{
		{"HideKVM", fields.FieldToggle},
		{"VendorID", fields.FieldText},
		{"HVFrequency", fields.FieldToggle},
		{"HVRelaxed", fields.FieldToggle},
		{"HVReset", fields.FieldToggle},
		{"HVRuntime", fields.FieldToggle},
		{"HVSpinlocks", fields.FieldText},
		{"HVStimer", fields.FieldToggle},
		{"HVSyncIC", fields.FieldToggle},
		{"HVTime", fields.FieldToggle},
		{"HVVapic", fields.FieldToggle},
		{"HVVPIndex", fields.FieldToggle},
		{"HVNoNonarchCoresharing", fields.FieldToggle},
		{"HVTLBFlush", fields.FieldToggle},
		{"HVTLBFlushExt", fields.FieldToggle},
		{"HVIPI", fields.FieldToggle},
		{"HVAVIC", fields.FieldToggle},
		{"TopoExt", fields.FieldToggle},
		{"L3Cache", fields.FieldToggle},
		{"X2APIC", fields.FieldToggle},
		{"Migratable", fields.FieldToggle},
		{"InvTSC", fields.FieldToggle},
		{"ForceCPUID0x80000026", fields.FieldToggle},
		{"RTCUTC", fields.FieldToggle},
		{"CPUPM", fields.FieldToggle},
	}

	for _, exp := range expectedFields {
		found := false
		for _, f := range fields.CPUOptionsFields {
			if f.Name == exp.name {
				found = true
				if f.Kind != exp.kind {
					t.Errorf("Field %s: expected kind %d, got %d", exp.name, exp.kind, f.Kind)
				}
				break
			}
		}
		if !found {
			t.Errorf("Field %s not found in registry", exp.name)
		}
	}
}

// TestCPUOptionsFormInit tests form initialization with default options
func TestCPUOptionsFormInit(t *testing.T) {
	vmManager := createTestVMManager(t)
	form := NewCPUOptionsFormModel(vmManager.Repository())

	if form.focusIndex != 1 {
		t.Errorf("Initial focusIndex = %d, want 1 (first interactive element after header)", form.focusIndex)
	}

	// Default options should all be false/empty using reflection getters
	if form.getToggleValue("HideKVM") {
		t.Errorf("Default HideKVM should be false")
	}
	if form.getTextValue("VendorID") != "" {
		t.Errorf("Default VendorID should be empty")
	}
}

// TestCPUOptionsFormToggle tests toggle behavior using table-driven tests
func TestCPUOptionsFormToggle(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		pressKeys []tea.KeyMsg
		want      bool
	}{
		{"HideKVM single toggle", "HideKVM", []tea.KeyMsg{tea.KeyPressMsg{Code: tea.KeyEnter}}, true},
		{"HideKVM double toggle", "HideKVM", []tea.KeyMsg{tea.KeyPressMsg{Code: tea.KeyEnter}, tea.KeyPressMsg{Code: tea.KeyEnter}}, false},
		{"HVRelaxed toggle", "HVRelaxed", []tea.KeyMsg{tea.KeyPressMsg{Code: tea.KeyEnter}}, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vmManager := createTestVMManager(t)
			form := NewCPUOptionsFormModel(vmManager.Repository())

			// Navigate to the field
			for i, pos := range form.positions {
				if pos.Key == tc.field {
					form.focusIndex = i
					break
				}
			}

			// Press keys
			for _, msg := range tc.pressKeys {
				model, _ := form.Update(msg)
				form = model.(*CPUOptionsFormModel)
			}

			if form.getToggleValue(tc.field) != tc.want {
				t.Errorf("Field %s = %v, want %v", tc.field, form.getToggleValue(tc.field), tc.want)
			}
		})
	}
}

// TestCPUOptionsFormNavigation tests Tab navigation
func TestCPUOptionsFormNavigation(t *testing.T) {
	vmManager := createTestVMManager(t)
	form := NewCPUOptionsFormModel(vmManager.Repository())

	// Start at HideKVM (index 1, after header at 0)
	if form.currentPos().fieldName != "HideKVM" {
		t.Fatalf("Expected initial field HideKVM, got %s", form.currentPos().fieldName)
	}

	// Tab moves to VendorID (index 2)
	model, _ := form.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	form = model.(*CPUOptionsFormModel)
	if form.currentPos().fieldName != "VendorID" {
		t.Errorf("Expected VendorID after Tab, got %s", form.currentPos().fieldName)
	}
}

// TestCPUOptionsFormTextEditing tests text field editing
func TestCPUOptionsFormTextEditing(t *testing.T) {
	vmManager := createTestVMManager(t)
	form := NewCPUOptionsFormModel(vmManager.Repository())

	// Move to VendorID (index 2)
	model, _ := form.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	form = model.(*CPUOptionsFormModel)

	// Type "AMD"
	model, _ = form.Update(tea.KeyPressMsg{Code: 'A', Text: "A"})
	form = model.(*CPUOptionsFormModel)
	model, _ = form.Update(tea.KeyPressMsg{Code: 'M', Text: "M"})
	form = model.(*CPUOptionsFormModel)
	model, _ = form.Update(tea.KeyPressMsg{Code: 'D', Text: "D"})
	form = model.(*CPUOptionsFormModel)

	if form.getTextValue("VendorID") != "AMD" {
		t.Errorf("VendorID = %q, want AMD", form.getTextValue("VendorID"))
	}

	// Backspace removes last character
	model, _ = form.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	form = model.(*CPUOptionsFormModel)

	if form.getTextValue("VendorID") != "AM" {
		t.Errorf("VendorID = %q, want AM after backspace", form.getTextValue("VendorID"))
	}
}

// TestCPUOptionsFormSave tests saving via Enter on save button
func TestCPUOptionsFormSave(t *testing.T) {
	vmManager := createTestVMManager(t)
	form := NewCPUOptionsFormModel(vmManager.Repository())

	// Set some options using reflection setters
	form.setBoolField("HideKVM", true)
	form.setBoolField("HVRelaxed", true)
	form.setStringField("VendorID", "TestVendor")

	// Navigate to save button (last position)
	form.focusIndex = len(form.positions) - 1

	// Press Enter to save
	model, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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
	var saved models.CPUOptions
	vmManager.Repository().GetConfig("cpu_options", &saved)
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
	form := NewCPUOptionsFormModel(vmManager.Repository())

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
		model, _ := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
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
		options:       &models.CPUOptions{},
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
	form := &CPUOptionsFormModel{
		options:       &models.CPUOptions{},
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}

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
		options:       &models.CPUOptions{},
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	form.positions = form.BuildPositions()

	// 23 toggles + 2 text fields + 1 save button + 3 section headers = 29
	expectedCount := 29
	if len(form.positions) != expectedCount {
		t.Errorf("Expected %d positions, got %d", expectedCount, len(form.positions))
	}
}

// TestCPUOptionsFormWithL3CacheSizeDie tests per-die L3 cache size fields in form
func TestCPUOptionsFormWithL3CacheSizeDie(t *testing.T) {
	vmManager := createTestVMManager(t)
	repo := vmManager.Repository()

	// Save host topology with 2 dies so the form shows per-die L3 fields
	// The form scans topology at runtime, so we mock it by creating a form and
	// directly providing hostTopo.

	f := NewCPUOptionsFormModelWithTopo(repo, models.HostCPUTopology{
		Dies: []models.CPUDie{
			{ID: 0, L3CacheKB: 32768},
			{ID: 1, L3CacheKB: 98304},
		},
	}, nil)

	// Find die L3 cache fields
	var die0Idx, die1Idx int = -1, -1
	for i, p := range f.positions {
		if p.Key == "L3CacheSizeDie0" {
			die0Idx = i
		}
		if p.Key == "L3CacheSizeDie1" {
			die1Idx = i
		}
	}

	if die0Idx < 0 {
		t.Error("Expected L3CacheSizeDie0 field in positions")
	}
	if die1Idx < 0 {
		t.Error("Expected L3CacheSizeDie1 field in positions")
	}

	if die0Idx >= 0 {
		// Check it's a text field
		if f.positions[die0Idx].Kind != form.FocusText {
			t.Errorf("L3CacheSizeDie0 kind = %d, want FocusText", f.positions[die0Idx].Kind)
		}
		// Check label mentions detected size
		label := f.positions[die0Idx].Label
		if !strings.Contains(label, "Die 0") || !strings.Contains(label, "32M") {
			t.Errorf("L3CacheSizeDie0 label = %q, should mention Die 0 and 32M", label)
		}
	}

	// Test editing die 0 L3 cache size (one char at a time)
	f.focusIndex = die0Idx
	f.handleCharInput("3")
	f.handleCharInput("2")
	f.handleCharInput("M")
	if f.getTextValue("L3CacheSizeDie0") != "32M" {
		t.Errorf("After typing '32M', L3CacheSizeDie0 = %q, want 32M", f.getTextValue("L3CacheSizeDie0"))
	}

	// Verify it's stored in the map
	if f.options.L3CacheSizeDie[0] != "32M" {
		t.Errorf("L3CacheSizeDie[0] = %q, want 32M", f.options.L3CacheSizeDie[0])
	}

	// Test empty clears from map
	f.focusIndex = die0Idx
	// Clear the field
	for i := 0; i < 3; i++ {
		f.handleBackspaceKey()
	}
	if f.getTextValue("L3CacheSizeDie0") != "" {
		t.Errorf("After clearing, L3CacheSizeDie0 = %q, want empty", f.getTextValue("L3CacheSizeDie0"))
	}
	// Entry should be deleted from map, not just empty string
	if _, exists := f.options.L3CacheSizeDie[0]; exists {
		t.Errorf("L3CacheSizeDie[0] should be deleted after clearing")
	}
}

// TestCPUOptionsFormWithL3CacheAssocDie tests per-die L3 cache associativity fields in form
func TestCPUOptionsFormWithL3CacheAssocDie(t *testing.T) {
	vmManager := createTestVMManager(t)
	repo := vmManager.Repository()

	f := NewCPUOptionsFormModelWithTopo(repo, models.HostCPUTopology{
		Dies: []models.CPUDie{
			{ID: 0, L3CacheKB: 32768},
			{ID: 1, L3CacheKB: 98304},
		},
	}, nil)

	// Find die L3 cache associativity fields
	var die0Idx, die1Idx int = -1, -1
	for i, p := range f.positions {
		if p.Key == "L3CacheAssocDie0" {
			die0Idx = i
		}
		if p.Key == "L3CacheAssocDie1" {
			die1Idx = i
		}
	}

	if die0Idx < 0 {
		t.Error("Expected L3CacheAssocDie0 field in positions")
	}
	if die1Idx < 0 {
		t.Error("Expected L3CacheAssocDie1 field in positions")
	}

	if die0Idx >= 0 {
		// Check it's a text field
		if f.positions[die0Idx].Kind != form.FocusText {
			t.Errorf("L3CacheAssocDie0 kind = %d, want FocusText", f.positions[die0Idx].Kind)
		}
		// Check label mentions die
		label := f.positions[die0Idx].Label
		if !strings.Contains(label, "Die 0") {
			t.Errorf("L3CacheAssocDie0 label = %q, should mention Die 0", label)
		}
	}

	// Test editing die 0 L3 cache associativity
	f.focusIndex = die0Idx
	f.handleCharInput("8")
	if f.getTextValue("L3CacheAssocDie0") != "8" {
		t.Errorf("After typing '8', L3CacheAssocDie0 = %q, want 8", f.getTextValue("L3CacheAssocDie0"))
	}

	// Verify it's stored in the map
	if f.options.L3CacheAssocDie[0] != 8 {
		t.Errorf("L3CacheAssocDie[0] = %d, want 8", f.options.L3CacheAssocDie[0])
	}

	// Test clearing by backspace
	f.focusIndex = die0Idx
	f.handleBackspaceKey()
	if f.getTextValue("L3CacheAssocDie0") != "" {
		t.Errorf("After backspace, L3CacheAssocDie0 = %q, want empty", f.getTextValue("L3CacheAssocDie0"))
	}
	// Entry should be deleted from map, not just empty string
	if _, exists := f.options.L3CacheAssocDie[0]; exists {
		t.Errorf("L3CacheAssocDie[0] should be deleted after clearing")
	}
}

// TestCPUOptionsFormL3CachePersistence tests L3 cache size/assoc persistence after save
func TestCPUOptionsFormL3CachePersistence(t *testing.T) {
	vmManager := createTestVMManager(t)
	repo := vmManager.Repository()

	f := NewCPUOptionsFormModelWithTopo(repo, models.HostCPUTopology{
		Dies: []models.CPUDie{
			{ID: 0, L3CacheKB: 32768, L3CacheAssoc: 16},
			{ID: 1, L3CacheKB: 98304, L3CacheAssoc: 12},
		},
	}, nil)

	// Navigate to L3CacheSizeDie0 and set value
	var size0Idx int = -1
	for i, p := range f.positions {
		if p.Key == "L3CacheSizeDie0" {
			size0Idx = i
			break
		}
	}
	if size0Idx < 0 {
		t.Fatal("L3CacheSizeDie0 field not found")
	}

	// Set L3 cache size via setTextValue (simulates user typing)
	f.focusIndex = size0Idx
	f.handleCharInput("3")
	f.handleCharInput("2")
	f.handleCharInput("M")

	if f.getTextValue("L3CacheSizeDie0") != "32M" {
		t.Fatalf("Before save: L3CacheSizeDie0 = %q, want 32M", f.getTextValue("L3CacheSizeDie0"))
	}

	// Navigate to save button (last position) and save
	f.focusIndex = len(f.positions) - 1
	model, cmd := f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	f = model.(*CPUOptionsFormModel)

	if cmd == nil {
		t.Fatal("Expected command after save, got nil")
	}
	msg := cmd()
	if _, ok := msg.(CPUOptionsUpdatedMsg); !ok {
		t.Errorf("Expected CPUOptionsUpdatedMsg, got %T", msg)
	}

	// Verify saved options directly via repo
	var saved models.CPUOptions
	if err := repo.GetConfig("cpu_options", &saved); err != nil {
		t.Fatalf("Failed to GetConfig: %v", err)
	}

	// Check L3 cache size persisted
	if saved.L3CacheSizeDie == nil {
		t.Fatal("L3CacheSizeDie is nil after save")
	}
	if saved.L3CacheSizeDie[0] != "32M" {
		t.Errorf("L3CacheSizeDie[0] = %q, want 32M AFTER SAVE", saved.L3CacheSizeDie[0])
	}

	// Also check L3 cache assoc
	// Set L3CacheAssocDie0
	var assoc0Idx int = -1
	for i, p := range f.positions {
		if p.Key == "L3CacheAssocDie0" {
			assoc0Idx = i
			break
		}
	}
	if assoc0Idx < 0 {
		t.Fatal("L3CacheAssocDie0 field not found")
	}

	f.focusIndex = assoc0Idx
	f.handleCharInput("1")
	f.handleCharInput("6")

	if f.getTextValue("L3CacheAssocDie0") != "16" {
		t.Fatalf("Before save: L3CacheAssocDie0 = %q, want 16", f.getTextValue("L3CacheAssocDie0"))
	}

	// Save again
	f.focusIndex = len(f.positions) - 1
	model, cmd = f.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	f = model.(*CPUOptionsFormModel)

	if cmd == nil {
		t.Fatal("Expected command after second save, got nil")
	}
	msg = cmd()
	if _, ok := msg.(CPUOptionsUpdatedMsg); !ok {
		t.Errorf("Expected CPUOptionsUpdatedMsg, got %T", msg)
	}

	// Re-read from repo
	if err := repo.GetConfig("cpu_options", &saved); err != nil {
		t.Fatalf("Failed to GetConfig: %v", err)
	}

	if saved.L3CacheAssocDie == nil {
		t.Fatal("L3CacheAssocDie is nil after save")
	}
	if saved.L3CacheAssocDie[0] != 16 {
		t.Errorf("L3CacheAssocDie[0] = %d, want 16 AFTER SAVE", saved.L3CacheAssocDie[0])
	}

	// Simulate re-entering form: create a new form and verify values
	f2 := NewCPUOptionsFormModelWithTopo(repo, models.HostCPUTopology{
		Dies: []models.CPUDie{
			{ID: 0, L3CacheKB: 32768, L3CacheAssoc: 16},
			{ID: 1, L3CacheKB: 98304, L3CacheAssoc: 12},
		},
	}, nil)

	if f2.getTextValue("L3CacheSizeDie0") != "32M" {
		t.Errorf("After re-open: L3CacheSizeDie0 = %q, want 32M", f2.getTextValue("L3CacheSizeDie0"))
	}
	if f2.getTextValue("L3CacheAssocDie0") != "16" {
		t.Errorf("After re-open: L3CacheAssocDie0 = %q, want 16", f2.getTextValue("L3CacheAssocDie0"))
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