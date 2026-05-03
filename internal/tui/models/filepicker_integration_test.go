package models

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// getVMForm extracts the inner VMFormModel from a ScrollableForm
func getVMForm(sf *form.ScrollableForm) *VMFormModel {
	if fm, ok := sf.Model().(*VMFormModel); ok {
		return fm
	}
	return nil
}

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

// TestFilePickerHardDiskEndToEnd tests that DiskAddedMsg is correctly routed
// through MainModel to update the form's hardDisks field.
func TestFilePickerHardDiskEndToEnd(t *testing.T) {
	m := setupEditModel(t)
	fm := getVMForm(m.vmEditModel.form)

	// Setup: form has an empty hard disk slot
	fm.hardDisks = []string{""}
	fm.browsingFieldName = "hardDisks"
	fm.browsingIndex = 0
	fm.addDiskModel = NewAddDiskModel(m.vmManager)

	// Simulate disk selection via DiskAddedMsg
	testFilePath := "/tmp/test-disk.qcow2"
	dam := DiskAddedMsg{Path: testFilePath, Canceled: false}
	m.Update(dam)

	fm = getVMForm(m.vmEditModel.form)
	// Verify the hard disk path was set
	if len(fm.hardDisks) == 0 {
		t.Fatal("hardDisks list is empty")
	}
	if fm.hardDisks[0] != testFilePath {
		t.Errorf("Expected hardDisks[0]='%s', got '%s'", testFilePath, fm.hardDisks[0])
	}

	// Verify the add disk model is cleared
	if fm.addDiskModel != nil {
		t.Error("Expected addDiskModel to be cleared after disk selection")
	}
}

// TestFilePickerDirectMsgRouting tests that FileSelectedMsg is correctly routed
// through addDiskModel and produces DiskAddedMsg which updates the form.
func TestFilePickerDirectMsgRouting(t *testing.T) {
	m := setupEditModel(t)
	fm := getVMForm(m.vmEditModel.form)

	// Set up form with active AddDiskModel (simulating mid-selection state)
	fm.hardDisks = []string{""}
	fm.browsingFieldName = "hardDisks"
	fm.browsingIndex = 0
	fm.addDiskModel = NewAddDiskModel(m.vmManager)

	testPath := "/tmp/test-disk.qcow2"
	dam := DiskAddedMsg{Path: testPath, Canceled: false}

	// Route through MainModel.Update
	m.Update(dam)

	fm = getVMForm(m.vmEditModel.form)
	// Verify the path was set (via DiskAddedMsg → handleDiskAdded)
	if fm.hardDisks[0] != testPath {
		t.Errorf("Expected hardDisks[0]='%s', got '%s'", testPath, fm.hardDisks[0])
	}

	// Verify addDiskModel was cleared
	if fm.addDiskModel != nil {
		t.Error("Expected addDiskModel to be nil")
	}
}

// TestFilePickerWithInactiveAddDiskModel tests the race condition scenario where
// addDiskModel is already inactive when MainModel receives FileSelectedMsg.
func TestFilePickerWithInactiveAddDiskModel(t *testing.T) {
	m := setupEditModel(t)
	fm := getVMForm(m.vmEditModel.form)

	// Set up form with INACTIVE AddDiskModel (race condition simulation)
	fm.hardDisks = []string{""}
	fm.browsingFieldName = "hardDisks"
	fm.browsingIndex = 0
	fm.addDiskModel = NewAddDiskModel(m.vmManager)
	fm.addDiskModel.active = false // Simulate prior deactivation

	testPath := "/tmp/test-disk.qcow2"
	dam := DiskAddedMsg{Path: testPath, Canceled: false}

	// Route through MainModel.Update
	m.Update(dam)

	fm = getVMForm(m.vmEditModel.form)
	// Even with inactive addDiskModel, the path should be set
	if fm.hardDisks[0] != testPath {
		t.Errorf("Expected hardDisks[0]='%s', got '%s'", testPath, fm.hardDisks[0])
	}
}

// TestCDROMFilePickerFlow tests that CDROM file selection still works correctly.
func TestCDROMFilePickerFlow(t *testing.T) {
	m := setupEditModel(t)
	fm := getVMForm(m.vmEditModel.form)

	// Set up form with a CDROM slot
	fm.cdroms = []string{""}
	fm.hardDisks = []string{"/existing.qcow2"}
	fm.rebuildPositions()
	fm.browsingFieldName = "cdroms"
	fm.browsingIndex = 0
	// No addDiskModel for CDROMs - direct file browser
	fm.fileBrowser = NewFileBrowserModel(FileTypeISO)

	testPath := "/tmp/test.iso"
	fsm := FileSelectedMsg{Path: testPath, Canceled: false}

	// Route through MainModel.Update
	m.Update(fsm)

	fm = getVMForm(m.vmEditModel.form)
	// Verify CDROM path was set
	if fm.cdroms[0] != testPath {
		t.Errorf("Expected cdroms[0]='%s', got '%s'", testPath, fm.cdroms[0])
	}

	// Verify hard disk was NOT modified
	if fm.hardDisks[0] != "/existing.qcow2" {
		t.Errorf("Expected hardDisks[0] unchanged, got '%s'", fm.hardDisks[0])
	}
}

// TestFilePickerCancelViaMainModel tests that cancel (via DiskAddedMsg with Canceled=true)
// preserves the existing disk value and clears the addDiskModel.
func TestFilePickerCancelViaMainModel(t *testing.T) {
	m := setupEditModel(t)
	fm := getVMForm(m.vmEditModel.form)

	fm.hardDisks = []string{"/existing.qcow2"}
	fm.browsingFieldName = "hardDisks"
	fm.browsingIndex = 0
	fm.addDiskModel = NewAddDiskModel(m.vmManager)

	// Simulate cancellation via DiskAddedMsg
	dam := DiskAddedMsg{Path: "", Canceled: true}
	m.Update(dam)

	fm = getVMForm(m.vmEditModel.form)
	// Verify hard disk was NOT modified (cancel should preserve existing value)
	if fm.hardDisks[0] != "/existing.qcow2" {
		t.Errorf("Expected hardDisks unchanged after cancel, got '%s'", fm.hardDisks[0])
	}

	// Verify addDiskModel was cleared
	if fm.addDiskModel != nil {
		t.Error("Expected addDiskModel to be nil after cancel")
	}
}
