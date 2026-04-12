package models

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// setupEditModel creates a MainModel in ViewVMEdit for testing file picker flows.
func setupEditModel(t *testing.T) *MainModel {
	t.Helper()

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "dkvmmanager"), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := vm.NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create vmManager: %v", err)
	}

	vmObj, err := mgr.CreateVM("test-vm")
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}

	editModel, err := NewVMEditModel(mgr, vmObj.ID)
	if err != nil {
		t.Fatalf("Failed to create edit model: %v", err)
	}

	tabModel := components.NewTabModel()

	return &MainModel{
		currentView:  ViewVMEdit,
		vmManager:    mgr,
		vmEditModel:  editModel,
		tabModel:     tabModel,
		windowWidth:  80,
		windowHeight: 24,
		breadcrumbs:  components.NewBreadcrumbs(),
		statusBar:    components.NewStatusBar(),
	}
}

// TestFilePickerHardDiskEndToEnd tests the complete flow of selecting a hard disk
// file through the file picker via MainModel message routing.
func TestFilePickerHardDiskEndToEnd(t *testing.T) {
	m := setupEditModel(t)

	// Setup: form has an empty hard disk slot
	m.vmEditModel.form.hardDisks = []string{""}
	m.vmEditModel.form.rebuildPositions()

	// Find the first hardDisk list item position
	hdIdx := -1
	for i, p := range m.vmEditModel.form.positions {
		if p.kind == focusListItem && p.fieldName == "hardDisks" {
			hdIdx = i
			break
		}
	}
	if hdIdx < 0 {
		t.Fatal("Could not find hardDisk list item position")
	}
	m.vmEditModel.form.focusIndex = hdIdx

	// Step 1: Press Enter to open the file picker
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	_, _ = m.Update(enterKey)

	if m.vmEditModel.form.addDiskModel == nil {
		t.Fatal("Expected addDiskModel after opening picker")
	}

	// Step 2: Press Enter to select "Disk image file" (step 0 → step 1)
	_, cmd := m.Update(enterKey)
	if cmd != nil {
		msg := cmd()
		m.Update(msg) // Process DirectoryLoadedMsg
	}

	// Step 3: Simulate file selection via FileSelectedMsg
	testFilePath := "/tmp/test-disk.qcow2"
	fsm := FileSelectedMsg{Path: testFilePath, Canceled: false}
	m.Update(fsm)

	// CRITICAL ASSERTIONS: Verify the hard disk path was set
	if len(m.vmEditModel.form.hardDisks) == 0 {
		t.Fatal("hardDisks list is empty")
	}
	if m.vmEditModel.form.hardDisks[0] != testFilePath {
		t.Errorf("Expected hardDisks[0]='%s', got '%s'", testFilePath, m.vmEditModel.form.hardDisks[0])
	}

	// Verify the file picker is closed
	if m.vmEditModel.form.addDiskModel != nil {
		t.Error("Expected addDiskModel to be cleared after file selection")
	}
	if m.vmEditModel.form.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() to be false")
	}
}

// TestFilePickerDirectMsgRouting tests that FileSelectedMsg is correctly routed
// through addDiskModel and produces DiskAddedMsg which updates the form.
func TestFilePickerDirectMsgRouting(t *testing.T) {
	m := setupEditModel(t)

	// Set up form with active AddDiskModel (simulating mid-selection state)
	m.vmEditModel.form.hardDisks = []string{""}
	m.vmEditModel.form.browsingFieldName = "hardDisks"
	m.vmEditModel.form.browsingIndex = 0
	m.vmEditModel.form.addDiskModel = NewAddDiskModel(m.vmManager)

	testPath := "/tmp/test-disk.qcow2"
	fsm := FileSelectedMsg{Path: testPath, Canceled: false}

	// Route through MainModel.Update - cmd is consumed internally by handleSubViewOutput
	m.Update(fsm)

	// Verify the path was set (via DiskAddedMsg → handleDiskAdded)
	if m.vmEditModel.form.hardDisks[0] != testPath {
		t.Errorf("Expected hardDisks[0]='%s', got '%s'", testPath, m.vmEditModel.form.hardDisks[0])
	}

	// Verify addDiskModel was cleared
	if m.vmEditModel.form.addDiskModel != nil {
		t.Error("Expected addDiskModel to be nil")
	}
}

// TestFilePickerWithInactiveAddDiskModel tests the race condition scenario where
// addDiskModel is already inactive when MainModel receives FileSelectedMsg.
func TestFilePickerWithInactiveAddDiskModel(t *testing.T) {
	m := setupEditModel(t)

	// Set up form with INACTIVE AddDiskModel (race condition simulation)
	m.vmEditModel.form.hardDisks = []string{""}
	m.vmEditModel.form.browsingFieldName = "hardDisks"
	m.vmEditModel.form.browsingIndex = 0
	m.vmEditModel.form.addDiskModel = NewAddDiskModel(m.vmManager)
	m.vmEditModel.form.addDiskModel.active = false // Simulate prior deactivation

	testPath := "/tmp/test-disk.qcow2"
	fsm := FileSelectedMsg{Path: testPath, Canceled: false}

	// Route through MainModel.Update
	m.Update(fsm)

	// Even with inactive addDiskModel, the path should be set
	if m.vmEditModel.form.hardDisks[0] != testPath {
		t.Errorf("Expected hardDisks[0]='%s', got '%s'", testPath, m.vmEditModel.form.hardDisks[0])
	}
}

// TestCDROMFilePickerFlow tests that CDROM file selection still works correctly.
func TestCDROMFilePickerFlow(t *testing.T) {
	m := setupEditModel(t)

	// Set up form with a CDROM slot
	m.vmEditModel.form.cdroms = []string{""}
	m.vmEditModel.form.hardDisks = []string{"/existing.qcow2"}
	m.vmEditModel.form.rebuildPositions()
	m.vmEditModel.form.browsingFieldName = "cdroms"
	m.vmEditModel.form.browsingIndex = 0
	// No addDiskModel for CDROMs - direct file browser
	m.vmEditModel.form.fileBrowser = NewFileBrowserModel(FileTypeISO)

	testPath := "/tmp/test.iso"
	fsm := FileSelectedMsg{Path: testPath, Canceled: false}

	// Route through MainModel.Update
	m.Update(fsm)

	// Verify CDROM path was set
	if m.vmEditModel.form.cdroms[0] != testPath {
		t.Errorf("Expected cdroms[0]='%s', got '%s'", testPath, m.vmEditModel.form.cdroms[0])
	}

	// Verify hard disk was NOT modified
	if m.vmEditModel.form.hardDisks[0] != "/existing.qcow2" {
		t.Errorf("Expected hardDisks[0] unchanged, got '%s'", m.vmEditModel.form.hardDisks[0])
	}
}

// TestFilePickerCancelViaMainModel tests ESC cancellation through the full
// MainModel message routing path.
func TestFilePickerCancelViaMainModel(t *testing.T) {
	m := setupEditModel(t)

	m.vmEditModel.form.hardDisks = []string{"/existing.qcow2"}
	m.vmEditModel.form.browsingFieldName = "hardDisks"
	m.vmEditModel.form.browsingIndex = 0
	m.vmEditModel.form.addDiskModel = NewAddDiskModel(m.vmManager)

	// Advance to step 1 (file browser)
	m.vmEditModel.form.addDiskModel.step = 1
	m.vmEditModel.form.addDiskModel.fileBrowser = NewFileBrowserModel(FileTypeDiskImage)
	m.vmEditModel.form.addDiskModel.fileBrowser.active = true

	// Verify file browser is active
	if !m.vmEditModel.form.FileBrowserActive() {
		t.Fatal("FileBrowserActive should be true in step 1")
	}

	// Simulate ESC via MainModel (do NOT call form.Update first, as it modifies state)
	// MainModel handles the command internally via handleSubViewOutput
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Verify hard disk was NOT modified (cancel should preserve existing value)
	if m.vmEditModel.form.hardDisks[0] != "/existing.qcow2" {
		t.Errorf("Expected hardDisks unchanged after cancel, got '%s'", m.vmEditModel.form.hardDisks[0])
	}

	// Verify addDiskModel was cleared
	if m.vmEditModel.form.addDiskModel != nil {
		t.Error("Expected addDiskModel to be nil after cancel")
	}

	// Verify file browser is no longer active
	if m.vmEditModel.form.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() to be false after cancel")
	}
}
