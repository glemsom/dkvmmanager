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

// TestEditFormHasFocus verifies that the edit form starts with focus enabled,
// so that visual indicators ("> " prefix, cursor, [Del] button) are rendered.
func TestEditFormHasFocus(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "dkvmmanager"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := vm.NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create vmManager: %v", err)
	}

	_, err = mgr.CreateVM("test-vm")
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}

	vms, err := mgr.ListVMs()
	if err != nil {
		t.Fatalf("Failed to list VMs: %v", err)
	}

	editModel, err := NewVMEditModel(mgr, vms[0].ID)
	if err != nil {
		t.Fatalf("Failed to create edit model: %v", err)
	}

	// The form must have focus for visual indicators to render
	if !editModel.form.focused {
		t.Error("BUG: VMFormModel.focused should be true in edit mode, but got false. " +
			"This causes no visual feedback when navigating form fields with arrow keys.")
	}
}

// TestCreateFormHasFocus verifies that the create form also starts with focus enabled.
func TestCreateFormHasFocus(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "dkvmmanager"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := vm.NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create vmManager: %v", err)
	}

	createModel := NewVMCreateModel(mgr)

	if !createModel.form.focused {
		t.Error("BUG: VMFormModel.focused should be true in create mode, but got false.")
	}
}

// TestEditFormArrowKeyNavigation verifies that arrow keys move focus between
// form fields and the rendered output reflects the new focus position.
func TestEditFormArrowKeyNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "dkvmmanager"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := vm.NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create vmManager: %v", err)
	}

	_, err = mgr.CreateVM("test-vm")
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}

	vms, err := mgr.ListVMs()
	if err != nil {
		t.Fatalf("Failed to list VMs: %v", err)
	}

	editModel, err := NewVMEditModel(mgr, vms[0].ID)
	if err != nil {
		t.Fatalf("Failed to create edit model: %v", err)
	}

	// Initialize viewport dimensions
	editModel.form.SetSize(76, 28)

	// Capture the initial view - should show focus indicator on first field
	initialView := editModel.View()
	if !strings.Contains(initialView, "> ") {
		t.Error("Initial view should contain focus indicator '> ' prefix")
	}

	initialFocusIndex := editModel.form.focusIndex

	// Press down arrow
	model, _ := editModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	editModel = model.(*VMEditModel)

	// Focus should have moved
	if editModel.form.focusIndex <= initialFocusIndex {
		t.Errorf("Focus index should have increased after pressing down. Before: %d, After: %d",
			initialFocusIndex, editModel.form.focusIndex)
	}

	// View should still show focus indicator on the new field
	afterView := editModel.View()
	if !strings.Contains(afterView, "> ") {
		t.Error("View after arrow down should still contain focus indicator '> ' prefix")
	}

	// Press down again
	model, _ = editModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	editModel = model.(*VMEditModel)

	// Focus should have moved again
	if editModel.form.focusIndex <= initialFocusIndex+1 {
		t.Errorf("Focus index should have increased again. Expected > %d, Got: %d",
			initialFocusIndex+1, editModel.form.focusIndex)
	}

	// Tab key should also work
	model, _ = editModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	editModel = model.(*VMEditModel)

	// Focus should keep moving
	if editModel.form.focusIndex <= initialFocusIndex+2 {
		t.Errorf("Focus index should have increased after tab. Expected > %d, Got: %d",
			initialFocusIndex+2, editModel.form.focusIndex)
	}
}

// TestEditFormArrowNavigationViaMainModel tests the full flow through MainModel
// to ensure arrow keys are properly delegated to the edit form.
func TestEditFormArrowNavigationViaMainModel(t *testing.T) {
	m := setupTestModelWithTwoVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	// Navigate: Config Tab -> Edit VM -> Select first VM
	m.tabModel.SetActiveTab(1) // TabConfiguration
	m.configSelectedIndex = 1  // "Edit VM"
	model, _ := m.handleMenuSelection()
	m = model.(*MainModel)

	if m.currentView != ViewVMSelect {
		t.Fatalf("Expected ViewVMSelect, got %s", m.currentView)
	}

	// Select first VM to enter edit mode
	model, _ = m.delegateToSubView(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(*MainModel)

	if m.currentView != ViewVMEdit {
		t.Fatalf("Expected ViewVMEdit, got %s", m.currentView)
	}

	if m.vmEditModel == nil {
		t.Fatal("vmEditModel should not be nil")
	}

	// The form must have focus
	if !m.vmEditModel.form.focused {
		t.Error("BUG: Edit form should have focused=true when entering edit view")
	}

	// Capture initial focus index
	initialFocusIndex := m.vmEditModel.form.focusIndex

	// Send down arrow through MainModel Update
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*MainModel)

	// Focus should have moved
	if m.vmEditModel.form.focusIndex <= initialFocusIndex {
		t.Errorf("Focus index should increase after down arrow. Before: %d, After: %d",
			initialFocusIndex, m.vmEditModel.form.focusIndex)
	}

	// The rendered view should show focus indicators
	view := m.vmEditModel.View()
	if !strings.Contains(view, "> ") {
		t.Error("Rendered view should contain '> ' focus indicator")
	}

	// Verify list items show [Del] button when focused
	// Press down enough times to reach a hard disk list item (position 2 typically)
	for i := 0; i < 5; i++ {
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = model.(*MainModel)
	}

	view = m.vmEditModel.View()
	// When a list item is focused, it should show [Del] button
	if !strings.Contains(view, "[Del]") {
		t.Error("Focused list item should show [Del] button indicator")
	}
}
