package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// TestCustomScriptFormInit tests form initialization with default config
func TestCustomScriptFormInit(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)

	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Default is custom mode (UseBuiltin = false) when no config saved
	if form.config.UseBuiltin {
		t.Errorf("Default UseBuiltin should be false when not configured")
	}

	// Initial focus should be on toggle
	if form.focusIndex != 0 {
		t.Errorf("Initial focusIndex = %d, want 0", form.focusIndex)
	}

	// Should have toggle position
	if len(form.positions) == 0 {
		t.Errorf("Expected positions to be populated")
	}
}

// TestCustomScriptFormToggleBuiltin tests toggling between builtin and custom mode
func TestCustomScriptFormToggleBuiltin(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Initial: custom mode (UseBuiltin = false by default), so path positions should show
	// In custom mode: toggle, start_path, start_browse, stop_path, stop_browse, save, cancel = 7 positions
	if len(form.positions) != 7 {
		t.Errorf("Expected 7 positions in custom mode initially, got %d", len(form.positions))
	}

	// Toggle to builtin mode via Enter key
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*StartStopScriptFormModel)

	// After toggle: should be builtin mode
	if !form.config.UseBuiltin {
		t.Errorf("UseBuiltin should be true after toggle")
	}

	// In builtin mode: toggle, save, cancel = 3 positions
	if len(form.positions) != 3 {
		t.Errorf("Expected 3 positions in builtin mode, got %d", len(form.positions))
	}

	// Toggle back to custom
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*StartStopScriptFormModel)

	if form.config.UseBuiltin {
		t.Errorf("UseBuiltin should be false after second toggle")
	}

	// Back to custom mode: 7 positions
	if len(form.positions) != 7 {
		t.Errorf("Expected 7 positions in custom mode, got %d", len(form.positions))
	}
}

// TestCustomScriptFormNavigation tests Tab navigation between fields
func TestCustomScriptFormNavigation(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Start at toggle (index 0)
	if form.currentPos().fieldName != "toggle" {
		t.Fatalf("Expected initial field toggle, got %s", form.currentPos().fieldName)
	}

	// In custom mode initially: 7 positions (0-6)
	// Tab moves to index 1
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*StartStopScriptFormModel)
	if form.focusIndex != 1 {
		t.Errorf("Expected focusIndex 1 after Tab, got %d", form.focusIndex)
	}

	// Up arrow at index 1 should go to index 0
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyUp})
	form = model.(*StartStopScriptFormModel)
	if form.focusIndex != 0 {
		t.Errorf("Expected focusIndex 0 after Up, got %d", form.focusIndex)
	}

	// Down arrow should go to index 1
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyDown})
	form = model.(*StartStopScriptFormModel)
	if form.focusIndex != 1 {
		t.Errorf("Expected focusIndex 1 after Down, got %d", form.focusIndex)
	}
}

// TestCustomScriptFormRenderBuiltin tests rendering in builtin mode
func TestCustomScriptFormRenderBuiltin(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Set size and render
	form.SetSize(78, 20)
	view := form.View()

	// Should contain key text
	if !strings.Contains(view, "Custom Start/Stop Script") {
		t.Error("Expected 'Custom Start/Stop Script' in view")
	}
	// Should show single-line toggle
	if !strings.Contains(view, "[Builtin]") {
		t.Error("Expected '[Builtin]' in view")
	}
	if !strings.Contains(view, "[Custom]") {
		t.Error("Expected '[Custom]' in view")
	}
	if !strings.Contains(view, "Save") {
		t.Error("Expected 'Save' in view")
	}
	if !strings.Contains(view, "Cancel") {
		t.Error("Expected 'Cancel' in view")
	}
}

// TestCustomScriptFormRenderCustom tests rendering in custom mode
func TestCustomScriptFormRenderCustom(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Switch to custom mode
	form.config.UseBuiltin = false
	form.rebuildPositions()
	form.SetSize(78, 20)
	view := form.View()

	// Should show script paths
	if !strings.Contains(view, "Start Script") {
		t.Error("Expected 'Start Script' in custom mode view")
	}
	if !strings.Contains(view, "Stop Script") {
		t.Error("Expected 'Stop Script' in custom mode view")
	}
}

// TestCustomScriptFormRebuildPositions tests rebuildPositions updates positions correctly
func TestCustomScriptFormRebuildPositions(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Builtin mode: toggle, save, cancel
	form.config.UseBuiltin = true
	form.rebuildPositions()
	if len(form.positions) != 3 {
		t.Errorf("Builtin mode: expected 3 positions, got %d", len(form.positions))
	}

	// Custom mode: toggle, start_path, start_browse, stop_path, stop_browse, save, cancel
	form.config.UseBuiltin = false
	form.rebuildPositions()
	if len(form.positions) != 7 {
		t.Errorf("Custom mode: expected 7 positions, got %d", len(form.positions))
	}

	// Focus index clamped when out of bounds
	form.focusIndex = 10
	form.rebuildPositions()
	if form.focusIndex != 0 {
		t.Errorf("Expected focusIndex clamped to 0, got %d", form.focusIndex)
	}
}

// TestCustomScriptFormEffectiveCursor tests cursor position calculation
func TestCustomScriptFormEffectiveCursor(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Default (no offset): end of string
	val := form.effectiveCursor("test", "hello")
	if val != 5 {
		t.Errorf("Expected cursor at end (5), got %d", val)
	}

	// Set offset
	form.setCursorOffset("test", 2)
	val = form.effectiveCursor("test", "hello")
	if val != 2 {
		t.Errorf("Expected cursor at 2, got %d", val)
	}

	// Offset beyond length clamped
	form.setCursorOffset("test", 10)
	val = form.effectiveCursor("test", "hello")
	if val != 5 {
		t.Errorf("Expected cursor clamped to 5, got %d", val)
	}

	// Negative offset returns end
	form.setCursorOffset("test", -1)
	val = form.effectiveCursor("test", "hello")
	if val != 5 {
		t.Errorf("Expected cursor at end for -1, got %d", val)
	}
}

// TestCustomScriptFormViewContent tests viewport content updates
func TestCustomScriptFormViewContent(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	form.SetSize(78, 20)

	// Should have rendered lines
	if len(form.renderedLines) == 0 {
		t.Error("Expected rendered lines after SetSize")
	}
}

// createTestVMManagerForScript creates a temporary VM manager for testing custom script form
func createTestVMManagerForScript(t *testing.T) *vm.Manager {
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

	// Pre-populate some PCI config for display in builtin mode
	pciCfg := models.PCIPassthroughConfig{
		Devices: []models.PCIPassthroughDevice{
			{Address: "0000:01:00.0", Name: "NVIDIA GPU"},
		},
	}
	if err := mgr.SavePCIPassthroughConfig(pciCfg); err != nil {
		t.Logf("Warning: could not save PCI config: %v", err)
	}

	return mgr
}

// TestEnterOnStartBrowseCreatesFileBrowser tests that Enter on start script browse button opens file browser
func TestEnterOnStartBrowseCreatesFileBrowser(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// In custom mode: positions are [toggle, start_path, start_browse, stop_path, stop_browse, save, cancel]
	// Focus on start_browse (index 2)
	form.focusIndex = 2

	// Verify focused on start_browse
	if form.positions[form.focusIndex].kind != startStopScriptStartBrowse {
		t.Fatalf("Expected focus on start_browse, got %s", form.positions[form.focusIndex].fieldName)
	}

	// Press Enter - should create file browser
	model, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*StartStopScriptFormModel)

	if form.fileBrowser == nil {
		t.Fatal("Expected fileBrowser to be created after Enter on start_browse")
	}
	// FileBrowserModel.Init() returns a command to load the initial directory
	if cmd == nil {
		t.Error("Expected non-nil command from fileBrowser Init()")
	}
}

// TestEnterOnStopBrowseCreatesFileBrowser tests that Enter on stop script browse button opens file browser
func TestEnterOnStopBrowseCreatesFileBrowser(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// In custom mode: positions are [toggle, start_path, start_browse, stop_path, stop_browse, save, cancel]
	// Focus on stop_browse (index 4)
	form.focusIndex = 4

	// Verify focused on stop_browse
	if form.positions[form.focusIndex].kind != startStopScriptStopBrowse {
		t.Fatalf("Expected focus on stop_browse, got %s", form.positions[form.focusIndex].fieldName)
	}

	// Press Enter - should create file browser
	model, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*StartStopScriptFormModel)

	if form.fileBrowser == nil {
		t.Fatal("Expected fileBrowser to be created after Enter on stop_browse")
	}
	if cmd == nil {
		t.Error("Expected non-nil command from fileBrowser Init()")
	}
}

// TestFileSelectedStartPath tests that selecting a file sets the start script path
func TestFileSelectedStartPath(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Set up file browser was active for start script
	form.fileBrowser = NewFileBrowserModel(FileTypeAll)
	form.focusIndex = 2 // start_browse

	// Simulate file selection
	msg := FileSelectedMsg{Path: "/path/to/start-script.sh", Canceled: false}
	form.handleFileSelected(msg)

	if form.fileBrowser != nil {
		t.Error("Expected fileBrowser to be cleared after selection")
	}
	if form.config.StartScript != "/path/to/start-script.sh" {
		t.Errorf("Expected StartScript='/path/to/start-script.sh', got '%s'", form.config.StartScript)
	}
}

// TestFileSelectedStopPath tests that selecting a file sets the stop script path
func TestFileSelectedStopPath(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Set up file browser was active for stop script
	form.fileBrowser = NewFileBrowserModel(FileTypeAll)
	form.focusIndex = 4 // stop_browse

	// Simulate file selection
	msg := FileSelectedMsg{Path: "/path/to/stop-script.sh", Canceled: false}
	form.handleFileSelected(msg)

	if form.fileBrowser != nil {
		t.Error("Expected fileBrowser to be cleared after selection")
	}
	if form.config.StopScript != "/path/to/stop-script.sh" {
		t.Errorf("Expected StopScript='/path/to/stop-script.sh', got '%s'", form.config.StopScript)
	}
}

// TestFileSelectedCanceled tests that canceling file selection clears the browser
func TestFileSelectedCanceled(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Pre-set a path
	form.config.StartScript = "/existing/script.sh"
	form.fileBrowser = NewFileBrowserModel(FileTypeAll)
	form.focusIndex = 2 // start_browse

	// Cancel selection
	msg := FileSelectedMsg{Path: "", Canceled: true}
	form.handleFileSelected(msg)

	if form.fileBrowser != nil {
		t.Error("Expected fileBrowser to be cleared after cancel")
	}
	// Path should be unchanged
	if form.config.StartScript != "/existing/script.sh" {
		t.Errorf("Expected StartScript unchanged, got '%s'", form.config.StartScript)
	}
}

// TestStartStopScriptKeyDelegationToFileBrowser tests that keys are delegated to active file browser
func TestStartStopScriptKeyDelegationToFileBrowser(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Activate file browser
	form.fileBrowser = NewFileBrowserModel(FileTypeAll)

	// Send ESC key - should be delegated to file browser
	_, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Error("Expected file browser to return a command on ESC")
	}
	// Execute the command to get the FileSelectedMsg
	msg := cmd()
	fsm, ok := msg.(FileSelectedMsg)
	if !ok {
		t.Fatalf("Expected FileSelectedMsg, got %T", msg)
	}
	if !fsm.Canceled {
		t.Error("Expected ESC to produce canceled FileSelectedMsg")
	}
}

// TestStartStopScriptViewShowsFileBrowserWhenActive tests that the View shows file browser when active
func TestStartStopScriptViewShowsFileBrowserWhenActive(t *testing.T) {
	vmManager := createTestVMManagerForScript(t)
	form, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	form.SetSize(78, 20)

	// Without file browser, should show form
	view := form.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
	if !strings.Contains(view, "Custom Start/Stop Script") {
		t.Error("Expected form content when file browser not active")
	}

	// With active file browser, should show file browser view
	form.fileBrowser = NewFileBrowserModel(FileTypeAll)
	view = form.View()
	if view == "" {
		t.Error("Expected non-empty file browser view")
	}
	// File browser view shows directory contents
	if !strings.Contains(view, "/") {
		t.Error("Expected file browser to show path")
	}
}
