// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// --- Test Fixtures ---

// mockUSBDevices returns a deterministic set of USB devices for testing
func mockUSBDevices() []models.USBDevice {
	return []models.USBDevice{
		{Vendor: "046d", Product: "c52b", Name: "Logitech Unifying Receiver", ID: "1-1.2"},
		{Vendor: "8087", Product: "0a2b", Name: "Intel Bluetooth", ID: "1-4"},
		{Vendor: "0781", Product: "5567", Name: "SanDisk Cruzer Blade", ID: "2-1"},
	}
}

// newTestUSBForm creates a USB passthrough form with mock devices (no real VM manager needed)
func newTestUSBForm(t *testing.T) *USBPassthroughFormModel {
	t.Helper()
	vmManager := createTestVMManager(t)

	formModel, err := NewUSBPassthroughFormModel(vmManager)
	if err != nil {
		// In CI, real scanning may fail; construct form manually with mock devices
		formModel = &USBPassthroughFormModel{
			vmManager: vmManager,
			devices:   mockUSBDevices(),
			selected:  make(map[string]bool),
			errors:    make(map[string]string),
		}
		formModel.BuildPositions()
	}

	// Apply mock devices regardless of scan result
	formModel.devices = mockUSBDevices()
	formModel.BuildPositions()

	// Apply window size to enable viewport
	formModel.Update(tea.WindowSizeMsg{Width: 80, Height: 25})

	return formModel
}

// --- Interface Implementation Tests ---

// TestUSBPassthroughFormModelInterface verifies the form implements form.FormModel
func TestUSBPassthroughFormModelInterface(t *testing.T) {
	vmManager := createTestVMManager(t)

	m, err := NewUSBPassthroughFormModel(vmManager)
	if err != nil {
		// In CI, real scanning may fail; construct manually
		m = &USBPassthroughFormModel{
			vmManager: vmManager,
			devices:   mockUSBDevices(),
			selected:  make(map[string]bool),
			errors:    make(map[string]string),
		}
	}

	// Verify it implements form.FormModel
	var _ form.FormModel = m

	// Verify BuildPositions returns expected positions
	positions := m.BuildPositions()
	if len(positions) == 0 {
		t.Fatal("Expected at least one position")
	}

	// All positions should be FocusToggle except the last (save button)
	for i, pos := range positions {
		if i < len(positions)-1 {
			if pos.Kind != form.FocusToggle {
				t.Errorf("Position %d: expected FocusToggle, got %v", i, pos.Kind)
			}
		} else {
			if pos.Kind != form.FocusButton || pos.Key != "save" {
				t.Errorf("Last position: expected FocusButton with key 'save', got Kind=%v Key=%s", pos.Kind, pos.Key)
			}
		}
	}

	// Verify FocusToggle positions have correct keys
	for _, pos := range positions {
		if pos.Kind == form.FocusToggle {
			if pos.Key == "" {
				t.Errorf("FocusToggle position has empty key")
			}
		}
	}
}

// --- Navigation Tests ---

// TestUSBNavigationTabCycle verifies Tab cycles through all positions
func TestUSBNavigationTabCycle(t *testing.T) {
	m := newTestUSBForm(t)

	positions := m.BuildPositions()

	// Start at index 0
	m.SetFocusIndex(0)

	// Tab forward through all positions
	for i := 1; i < len(positions); i++ {
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = model.(*USBPassthroughFormModel)
		if m.focusIndex != i {
			t.Errorf("After Tab #%d, expected focusIndex=%d, got %d", i, i, m.focusIndex)
		}
	}

	// At last position, Tab should stay
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(*USBPassthroughFormModel)
	if m.focusIndex != len(positions)-1 {
		t.Errorf("Tab at last position should stay at %d, got %d", len(positions)-1, m.focusIndex)
	}
}

// TestUSBNavigationShiftTabBackward verifies Shift+Tab moves focus backward
func TestUSBNavigationShiftTabBackward(t *testing.T) {
	m := newTestUSBForm(t)

	// Tab forward once
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(*USBPassthroughFormModel)
	firstFocusable := m.focusIndex

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(*USBPassthroughFormModel)

	// Shift+Tab should go back
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = model.(*USBPassthroughFormModel)
	if m.focusIndex != firstFocusable {
		t.Errorf("After Shift+Tab, focusIndex = %d, want %d", m.focusIndex, firstFocusable)
	}
}

// TestUSBNavigationUpArrow verifies Up arrow moves focus backward
func TestUSBNavigationUpArrow(t *testing.T) {
	m := newTestUSBForm(t)

	m.SetFocusIndex(1)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = model.(*USBPassthroughFormModel)

	if m.focusIndex >= 1 {
		t.Errorf("After Up from index 1, focus should move back, got %d", m.focusIndex)
	}
}

// TestUSBNavigationDownArrow verifies Down arrow moves focus forward
func TestUSBNavigationDownArrow(t *testing.T) {
	m := newTestUSBForm(t)

	m.SetFocusIndex(0)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*USBPassthroughFormModel)

	if m.focusIndex <= 0 {
		t.Errorf("After Down from index 0, focus should move forward, got %d", m.focusIndex)
	}
}

// TestUSBNavigationBoundaryUp verifies Up at first position stays at 0
func TestUSBNavigationBoundaryUp(t *testing.T) {
	m := newTestUSBForm(t)

	m.SetFocusIndex(0)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = model.(*USBPassthroughFormModel)

	if m.focusIndex != 0 {
		t.Errorf("Up at position 0 should stay at 0, got %d", m.focusIndex)
	}
}

// TestUSBNavigationBoundaryDown verifies Down at last position stays at end
func TestUSBNavigationBoundaryDown(t *testing.T) {
	m := newTestUSBForm(t)

	positions := m.BuildPositions()
	m.SetFocusIndex(len(positions) - 1)
	lastIdx := m.focusIndex
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*USBPassthroughFormModel)

	if m.focusIndex != lastIdx {
		t.Errorf("Down at last position %d should stay, got %d", lastIdx, m.focusIndex)
	}
}

// --- Toggle Tests ---

// TestUSBToggleViaEnterKey verifies Enter key toggles device on/off
func TestUSBToggleViaEnterKey(t *testing.T) {
	m := newTestUSBForm(t)

	positions := m.BuildPositions()

	// Find first toggle
	for i, pos := range positions {
		if pos.Kind == form.FocusToggle {
			m.SetFocusIndex(i)
			break
		}
	}

	addr := positions[m.focusIndex].Key

	// Enter to select
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.selected[addr] {
		t.Error("Device should be selected after Enter")
	}

	// Enter to deselect
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.selected[addr] {
		t.Error("Device should be deselected after second Enter")
	}
}

// TestUSBToggleViaSpaceKey verifies Space key toggles device on/off
func TestUSBToggleViaSpaceKey(t *testing.T) {
	m := newTestUSBForm(t)

	positions := m.BuildPositions()

	// Find first toggle
	for i, pos := range positions {
		if pos.Kind == form.FocusToggle {
			m.SetFocusIndex(i)
			break
		}
	}

	addr := positions[m.focusIndex].Key

	// Space to select
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !m.selected[addr] {
		t.Error("Device should be selected after Space")
	}
}

// TestUSBToggleDevice verifies toggleDevice toggles selection correctly
func TestUSBToggleDevice(t *testing.T) {
	m := newTestUSBForm(t)

	key := usbDeviceKey("046d", "c52b")

	// Initially not selected
	if m.selected[key] {
		t.Error("Device should not be initially selected")
	}

	// Toggle on
	m.toggleDevice(key)
	if !m.selected[key] {
		t.Error("Device should be selected after toggle")
	}

	// Toggle off
	m.toggleDevice(key)
	if m.selected[key] {
		t.Error("Device should be deselected after second toggle")
	}
}

// --- Render Tests ---

// TestUSBRenderDeviceToggle verifies device toggle renders correctly
func TestUSBRenderDeviceToggle(t *testing.T) {
	m := newTestUSBForm(t)
	m.syncViewport()

	lines := m.renderAllLines()

	// Find a line containing a known device name
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Logitech Unifying Receiver") {
			found = true
			// Verify vendor:product format is present
			if !strings.Contains(line, "046d:c52b") {
				t.Errorf("Device line should contain vendor:product key, got: %s", line)
			}
			break
		}
	}

	if !found {
		t.Error("Expected device 'Logitech Unifying Receiver' in rendered output")
	}
}

// TestUSBRenderFocusIndicator verifies focused device has '> ' prefix
func TestUSBRenderFocusIndicator(t *testing.T) {
	m := newTestUSBForm(t)

	// Focus on first toggle
	m.SetFocusIndex(0)
	m.syncViewport()

	lines := m.renderAllLines()

	// First device line should have focus indicator
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && strings.Contains(trimmed, "Logitech Unifying Receiver") &&
			!strings.HasPrefix(trimmed, "Tab ") && !strings.Contains(trimmed, "Save") &&
			!strings.Contains(trimmed, "USB Passthrough") {
			found = true
			if !strings.Contains(line, "> ") {
				t.Errorf("Focused device line should contain '> ' indicator, got: %s", line)
			}
			break
		}
	}

	if !found {
		t.Error("Expected focused device line in rendered output")
	}
}

// TestUSBViewShowsNoDevicesMessage verifies "no devices" message when scan finds nothing
func TestUSBViewShowsNoDevicesMessage(t *testing.T) {
	vmManager := createTestVMManager(t)

	m := &USBPassthroughFormModel{
		vmManager: vmManager,
		devices:   []models.USBDevice{},
		selected:  make(map[string]bool),
		errors:    make(map[string]string),
	}
	m.BuildPositions()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 25})
	m.syncViewport()

	view := m.View()
	if !strings.Contains(view, "No USB devices found") {
		t.Errorf("Expected 'No USB devices found' message, got:\n%s", view)
	}
}

// TestUSBOneLinePerDevice verifies each device occupies exactly one toggle line
func TestUSBOneLinePerDevice(t *testing.T) {
	m := newTestUSBForm(t)

	positions := m.BuildPositions()

	// Count toggle positions — should equal number of devices
	toggleCount := 0
	for _, pos := range positions {
		if pos.Kind == form.FocusToggle {
			toggleCount++
		}
	}

	if toggleCount != len(m.devices) {
		t.Errorf("Expected %d toggle positions (one per device), got %d", len(m.devices), toggleCount)
	}

	// Also verify save button exists
	saveCount := 0
	for _, pos := range positions {
		if pos.Kind == form.FocusButton && pos.Key == "save" {
			saveCount++
		}
	}
	if saveCount != 1 {
		t.Errorf("Expected 1 save button position, got %d", saveCount)
	}
}

// --- Save Tests ---

// TestUSBFormEmptyDevices verifies behavior when no devices are found
func TestUSBFormEmptyDevices(t *testing.T) {
	vmManager := createTestVMManager(t)

	m := &USBPassthroughFormModel{
		vmManager: vmManager,
		devices:   []models.USBDevice{},
		selected:  make(map[string]bool),
		errors:    make(map[string]string),
	}
	positions := m.BuildPositions()

	// Should only have save button
	if len(positions) != 1 {
		t.Errorf("Expected 1 position (save button), got %d", len(positions))
	}
	if positions[0].Kind != form.FocusButton || positions[0].Key != "save" {
		t.Errorf("Expected save button as first position, got Kind=%v Key=%s", positions[0].Kind, positions[0].Key)
	}
}

// TestUSBFormBuildPositionsKeyFormat verifies device keys use vendor:product format
func TestUSBFormBuildPositionsKeyFormat(t *testing.T) {
	m := newTestUSBForm(t)

	positions := m.BuildPositions()

	for _, pos := range positions {
		if pos.Kind == form.FocusToggle {
			// Key should be in vendor:product format
			parts := strings.Split(pos.Key, ":")
			if len(parts) != 2 {
				t.Errorf("Toggle key should be in vendor:product format, got: %s", pos.Key)
			}
		}
	}
}

// TestUSBFormCurrentIndex verifies CurrentIndex/SetFocusIndex work correctly
func TestUSBFormCurrentIndex(t *testing.T) {
	m := newTestUSBForm(t)

	m.SetFocusIndex(2)
	if m.CurrentIndex() != 2 {
		t.Errorf("CurrentIndex should be 2 after SetFocusIndex(2), got %d", m.CurrentIndex())
	}
}

// TestUSBFormPositionCountMatchesDevices verifies BuildPositions count matches device count
func TestUSBFormPositionCountMatchesDevices(t *testing.T) {
	m := newTestUSBForm(t)

	positions := m.BuildPositions()

	// Should have len(devices) toggles + 1 save button
	expected := len(m.devices) + 1
	if len(positions) != expected {
		t.Errorf("Expected %d positions (%d devices + 1 save), got %d", expected, len(m.devices), len(positions))
	}
}
