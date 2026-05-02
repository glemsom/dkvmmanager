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

	// Press down arrow twice to reach first hard disk list item (position 2)
	// (position 1 is the hardDisks_label which doesn't show indicators)
	for i := 0; i < 2; i++ {
		model, _ := editModel.Update(tea.KeyMsg{Type: tea.KeyDown})
		editModel = model.(*VMEditModel)
	}

	// Focus should have moved to position 2 (list item)
	if editModel.form.focusIndex != initialFocusIndex+2 {
		t.Errorf("Focus index should be %d after 2 down presses, got %d",
			initialFocusIndex+2, editModel.form.focusIndex)
	}

	// List items show [Del] when focused, not '> '
	afterView := editModel.View()
	if !strings.Contains(afterView, "[Del]") {
		t.Error("Focused list item should show [Del] button indicator")
	}

	// Press down to reach +Add button, then Tab to verify Tab works
	model, _ := editModel.Update(tea.KeyMsg{Type: tea.KeyDown})
	editModel = model.(*VMEditModel)
	model, _ = editModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	editModel = model.(*VMEditModel)

	// Focus should have moved forward by Tab
	if editModel.form.focusIndex <= initialFocusIndex+2 {
		t.Errorf("Focus index should have advanced after down+tab. Got: %d", editModel.form.focusIndex)
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

	// Verify list items show [Del] button when focused
	// First navigate to the hard disk list item (position 2) from start
	for i := 0; i < 2; i++ {
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = model.(*MainModel)
	}

	view := m.vmEditModel.View()
	// When a list item is focused, it should show [Del] button
	if !strings.Contains(view, "[Del]") {
		t.Error("Focused list item should show [Del] button indicator")
	}

	// Now navigate to macAddress text field (position 6) to check '> '
	for i := 0; i < 4; i++ {
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = model.(*MainModel)
	}

	// Focus should have moved to a text field
	if m.vmEditModel.form.focusIndex <= 2 {
		t.Errorf("Focus index should have moved past list items. Got: %d", m.vmEditModel.form.focusIndex)
	}

	view = m.vmEditModel.View()
	if !strings.Contains(view, "> ") {
		t.Error("Rendered view should contain '> ' focus indicator on text field")
	}
}
