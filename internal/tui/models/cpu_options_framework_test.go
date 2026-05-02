package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// TestCPUOptionsForm_ImplementsFormModel verifies CPUOptionsFormModel satisfies the FormModel interface.
func TestCPUOptionsForm_ImplementsFormModel(t *testing.T) {
	vmManager := createTestVMManager(t)
	m := NewCPUOptionsFormModel(vmManager)

	// Compile-time check: must satisfy form.FormModel
	var _ form.FormModel = m
}

// TestCPUOptionsForm_BuildPositions verifies the form exposes correct focus positions.
func TestCPUOptionsForm_BuildPositions(t *testing.T) {
	vmManager := createTestVMManager(t)
	m := NewCPUOptionsFormModel(vmManager)

	positions := m.BuildPositions()

	// 22 toggles + 2 text fields + 1 save button = 25
	if len(positions) != 25 {
		t.Errorf("Expected 25 positions, got %d", len(positions))
	}

	// First position should be HideKVM toggle
	if positions[0].Kind != form.FocusToggle {
		t.Errorf("First position should be FocusToggle, got %d", positions[0].Kind)
	}
	if positions[0].Key != "HideKVM" {
		t.Errorf("First position key should be HideKVM, got %s", positions[0].Key)
	}

	// Second position should be VendorID text field
	if positions[1].Kind != form.FocusText {
		t.Errorf("Second position should be FocusText, got %d", positions[1].Kind)
	}
	if positions[1].Key != "VendorID" {
		t.Errorf("Second position key should be VendorID, got %s", positions[1].Key)
	}

	// Last position should be save button
	last := positions[len(positions)-1]
	if last.Kind != form.FocusButton {
		t.Errorf("Last position should be FocusButton, got %d", last.Kind)
	}
}

// TestCPUOptionsForm_Integration_Toggle verifies toggle behavior through the framework wrapper.
func TestCPUOptionsForm_Integration_Toggle(t *testing.T) {
	vmManager := createTestVMManager(t)
	wrapped := NewCPUOptionsModel(vmManager)

	// Initialize with window size
	wrapped.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Focus starts on HideKVM (first toggle)
	fm := wrapped.Form()
	if fm.CurrentIndex() != 0 {
		t.Fatalf("Expected focus index 0, got %d", fm.CurrentIndex())
	}
	if fm.getToggleValue("HideKVM") {
		t.Fatal("HideKVM should start false")
	}

	// Press Enter to toggle
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !fm.getToggleValue("HideKVM") {
		t.Error("HideKVM should be true after Enter")
	}

	// Toggle again
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if fm.getToggleValue("HideKVM") {
		t.Error("HideKVM should be false after second toggle")
	}
}

// TestCPUOptionsForm_Integration_Navigation verifies Tab/Shift+Tab navigation through the framework.
func TestCPUOptionsForm_Integration_Navigation(t *testing.T) {
	vmManager := createTestVMManager(t)
	wrapped := NewCPUOptionsModel(vmManager)

	wrapped.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Start at HideKVM (index 0)
	fm := wrapped.Form()
	if fm.CurrentIndex() != 0 {
		t.Fatalf("Expected focus 0, got %d", fm.CurrentIndex())
	}

	// Tab → VendorID (index 1)
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyTab})
	if fm.CurrentIndex() != 1 {
		t.Errorf("Expected focus 1 after Tab, got %d", fm.CurrentIndex())
	}

	// Tab → HVFrequency (index 2)
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyTab})
	if fm.CurrentIndex() != 2 {
		t.Errorf("Expected focus 2 after Tab, got %d", fm.CurrentIndex())
	}

	// Shift+Tab → VendorID (index 1)
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if fm.CurrentIndex() != 1 {
		t.Errorf("Expected focus 1 after Shift+Tab, got %d", fm.CurrentIndex())
	}

	// Up arrow → HideKVM (index 0)
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyUp})
	if fm.CurrentIndex() != 0 {
		t.Errorf("Expected focus 0 after Up, got %d", fm.CurrentIndex())
	}

	// Down arrow → VendorID (index 1)
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyDown})
	if fm.CurrentIndex() != 1 {
		t.Errorf("Expected focus 1 after Down, got %d", fm.CurrentIndex())
	}
}

// TestCPUOptionsForm_Integration_TextInput verifies text field editing through the framework.
func TestCPUOptionsForm_Integration_TextInput(t *testing.T) {
	vmManager := createTestVMManager(t)
	wrapped := NewCPUOptionsModel(vmManager)

	wrapped.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Navigate to VendorID (index 1)
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Type "AMD"
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}})
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})

	fm := wrapped.Form()
	if fm.getTextValue("VendorID") != "AMD" {
		t.Errorf("VendorID = %q, want AMD", fm.getTextValue("VendorID"))
	}

	// Backspace → "AM"
	_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if fm.getTextValue("VendorID") != "AM" {
		t.Errorf("VendorID = %q, want AM after backspace", fm.getTextValue("VendorID"))
	}
}

// TestCPUOptionsForm_Integration_Save verifies save behavior through the framework.
func TestCPUOptionsForm_Integration_Save(t *testing.T) {
	vmManager := createTestVMManager(t)
	wrapped := NewCPUOptionsModel(vmManager)

	wrapped.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	fm := wrapped.Form()

	// Set some values
	m := fm // *CPUOptionsFormModel
	m.options.HideKVM = true
	m.options.HVRelaxed = true
	m.options.VendorID = "TestVendor"

	// Navigate to save button (last position)
	positions := fm.BuildPositions()
	for i := 0; i < len(positions); i++ {
		_, _ = wrapped.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	// Press Enter to save
	_, cmd := wrapped.Update(tea.KeyMsg{Type: tea.KeyEnter})

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
		t.Error("Saved HideKVM should be true")
	}
	if !saved.HVRelaxed {
		t.Error("Saved HVRelaxed should be true")
	}
	if saved.VendorID != "TestVendor" {
		t.Errorf("Saved VendorID = %q, want TestVendor", saved.VendorID)
	}
}

// TestCPUOptionsUpdatedMsg_ImplementsFormSavedMsg verifies the message type works with the framework.
func TestCPUOptionsUpdatedMsg_ImplementsFormSavedMsg(t *testing.T) {
	msg := CPUOptionsUpdatedMsg{}

	// Must satisfy form.FormSavedMsg
	var _ form.FormSavedMsg = msg

	if msg.FormName() != "CPU Options" {
		t.Errorf("FormName = %q, want CPU Options", msg.FormName())
	}
	if msg.FormStatus() != "" {
		t.Errorf("FormStatus = %q, want empty string", msg.FormStatus())
	}
}

// cpuTestVM creates a temporary VM manager for CPU options tests.
func cpuTestVM(t *testing.T) *vm.Manager {
	return createTestVMManager(t)
}
