package models

import (
	"testing"

	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

func TestFileBrowserActiveInitiallyFalse(t *testing.T) {
	m := setupTestForm(t)
	if m.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() to be false on fresh form")
	}
}

func TestOpenFilePickerCmdHardDisk(t *testing.T) {
	m := setupTestForm(t)
	// Focus on first hardDisk list item
	positions := m.BuildPositions()
	for i, p := range positions {
		if p.Kind == form.FocusList {
			m.focusIndex = i
			cmd := m.openFilePickerCmd(p)
			if m.addDiskModel == nil {
				t.Fatal("Expected addDiskModel to be set after opening picker for harddisk")
			}
			if m.fileBrowser != nil {
				t.Fatal("Expected fileBrowser to be nil for harddisk picker")
			}
			if m.browsingFieldName != "hardDisks" {
				t.Errorf("Expected browsingFieldName='hardDisks', got '%s'", m.browsingFieldName)
			}
			if m.browsingIndex != 0 {
				t.Errorf("Expected browsingIndex=0, got %d", m.browsingIndex)
			}
			// AddDiskModel.Init() returns nil (step 0 is source type selection)
			if cmd != nil {
				t.Error("Expected nil command from AddDiskModel Init() at step 0")
			}
			return
		}
	}
	t.Fatal("Could not find hardDisk list item position")
}

func TestOpenFilePickerCmdISO(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{""}
	m.rebuildPositions()

	positions := m.BuildPositions()
	for i, p := range positions {
		if p.Kind == form.FocusList && p.Key[:6] == "cdroms" {
			m.focusIndex = i
			cmd := m.openFilePickerCmd(p)
			if m.fileBrowser == nil {
				t.Fatal("Expected fileBrowser to be set after opening picker for ISO")
			}
			if m.addDiskModel != nil {
				t.Fatal("Expected addDiskModel to be nil for ISO picker")
			}
			if m.browsingFieldName != "cdroms" {
				t.Errorf("Expected browsingFieldName='cdroms', got '%s'", m.browsingFieldName)
			}
			// FileBrowserModel.Init() returns loadDirectory command
			if cmd == nil {
				t.Error("Expected non-nil command from FileBrowserModel Init()")
			}
			return
		}
	}
	t.Fatal("Could not find cdrom list item position")
}

func TestFileBrowserActiveAfterOpenHardDisk(t *testing.T) {
	m := setupTestForm(t)
	positions := m.BuildPositions()
	for _, p := range positions {
		if p.Kind == form.FocusList {
			m.openFilePickerCmd(p)
			if !m.FileBrowserActive() {
				t.Error("Expected FileBrowserActive() to be true after opening harddisk picker")
			}
			return
		}
	}
	t.Fatal("Could not find list item position")
}

func TestFileBrowserActiveAfterOpenISO(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{""}
	m.rebuildPositions()

	positions := m.BuildPositions()
	for _, p := range positions {
		if p.Kind == form.FocusList {
			m.openFilePickerCmd(p)
			if !m.FileBrowserActive() {
				t.Error("Expected FileBrowserActive() to be true after opening ISO picker")
			}
			return
		}
	}
	t.Fatal("Could not find list item position")
}

func TestHandleFileSelectedCmdSetsPath(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{""}
	m.browsingFieldName = "cdroms"
	m.browsingIndex = 0
	m.fileBrowser = NewFileBrowserModel(FileTypeISO)

	msg := FileSelectedMsg{Path: "/path/to/image.iso", Canceled: false}
	cmd := m.handleFileSelectedCmd(msg)

	if m.fileBrowser != nil {
		t.Error("Expected fileBrowser to be cleared after selection")
	}
	if m.cdroms[0] != "/path/to/image.iso" {
		t.Errorf("Expected cdroms[0]='/path/to/image.iso', got '%s'", m.cdroms[0])
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestHandleFileSelectedCmdCanceled(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{"/existing/path.iso"}
	m.browsingFieldName = "cdroms"
	m.browsingIndex = 0
	m.fileBrowser = NewFileBrowserModel(FileTypeISO)

	msg := FileSelectedMsg{Path: "", Canceled: true}
	cmd := m.handleFileSelectedCmd(msg)

	if m.fileBrowser != nil {
		t.Error("Expected fileBrowser to be cleared after cancel")
	}
	if m.cdroms[0] != "/existing/path.iso" {
		t.Errorf("Expected cdroms[0] unchanged='/existing/path.iso', got '%s'", m.cdroms[0])
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestHandleDiskAddedCmdSetsPath(t *testing.T) {
	m := setupTestForm(t)
	m.hardDisks = []string{""}
	m.browsingFieldName = "hardDisks"
	m.browsingIndex = 0
	m.addDiskModel = NewAddDiskModel(m.vmManager)

	msg := DiskAddedMsg{Path: "/dev/sda", Canceled: false}
	cmd := m.handleDiskAddedCmd(msg)

	if m.addDiskModel != nil {
		t.Error("Expected addDiskModel to be cleared after selection")
	}
	if m.hardDisks[0] != "/dev/sda" {
		t.Errorf("Expected hardDisks[0]='/dev/sda', got '%s'", m.hardDisks[0])
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestHandleDiskAddedCmdCanceled(t *testing.T) {
	m := setupTestForm(t)
	m.hardDisks = []string{"/existing/disk.qcow2"}
	m.browsingFieldName = "hardDisks"
	m.browsingIndex = 0
	m.addDiskModel = NewAddDiskModel(m.vmManager)

	msg := DiskAddedMsg{Path: "", Canceled: true}
	cmd := m.handleDiskAddedCmd(msg)

	if m.addDiskModel != nil {
		t.Error("Expected addDiskModel to be cleared after cancel")
	}
	if m.hardDisks[0] != "/existing/disk.qcow2" {
		t.Errorf("Expected hardDisks[0] unchanged, got '%s'", m.hardDisks[0])
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestHandleEnterOpensPickerOnListItem(t *testing.T) {
	m := setupTestForm(t)
	positions := m.BuildPositions()
	for _, p := range positions {
		if p.Kind == form.FocusList {
			result, cmd := m.HandleEnter(p)
			if m.addDiskModel == nil {
				t.Fatal("Expected HandleEnter to open file picker for harddisk list item")
			}
			if result != form.ResultNone {
				t.Errorf("Expected ResultNone, got %d", result)
			}
			// AddDiskModel.Init() returns nil at step 0
			_ = cmd
			return
		}
	}
	t.Fatal("Could not find list item position")
}

func TestHandleEnterMovesFocusOnTextField(t *testing.T) {
	m := setupTestForm(t)
	// Focus on vmName text field (first position)
	positions := m.BuildPositions()
	if len(positions) == 0 {
		t.Fatal("No positions")
	}
	m.focusIndex = 0

	initialIndex := m.focusIndex
	result, _ := m.HandleEnter(positions[0])

	if m.focusIndex == initialIndex {
		t.Error("Expected Enter on text field to move focus to next field")
	}
	if result != form.ResultNone {
		t.Errorf("Expected ResultNone, got %d", result)
	}
	if m.fileBrowser != nil {
		t.Error("Expected no file browser for text field Enter")
	}
	if m.addDiskModel != nil {
		t.Error("Expected no add disk model for text field Enter")
	}
}

func TestVMCreateModelFileBrowserActive(t *testing.T) {
	m := setupTestForm(t)
	create := &VMCreateModel{form: form.NewScrollableForm(m)}

	if create.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() false initially")
	}

	m.fileBrowser = NewFileBrowserModel(FileTypeISO)
	if !create.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() true after setting fileBrowser")
	}
}

func TestVMEditModelFileBrowserActive(t *testing.T) {
	m := setupTestForm(t)
	edit := &VMEditModel{form: form.NewScrollableForm(m)}

	if edit.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() false initially")
	}

	m.addDiskModel = NewAddDiskModel(m.vmManager)
	if !edit.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() true after setting addDiskModel")
	}
}

func TestMultipleISOsFilePicker(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{"/first.iso", ""}
	m.rebuildPositions()

	// Find second cdrom item by counting cdrom positions
	positions := m.BuildPositions()
	cdromCount := 0
	for _, p := range positions {
		if p.Kind == form.FocusList && len(p.Key) > 6 && p.Key[:7] == "cdroms_" {
			cdromCount++
			if cdromCount == 2 {
				cmd := m.openFilePickerCmd(p)
				if m.browsingFieldName != "cdroms" {
					t.Errorf("Expected browsingFieldName='cdroms', got '%s'", m.browsingFieldName)
				}
				if m.browsingIndex != 1 {
					t.Errorf("Expected browsingIndex=1, got %d", m.browsingIndex)
				}
				// FileBrowserModel.Init() returns non-nil
				if cmd == nil {
					t.Error("Expected non-nil command from file picker Init()")
				}

				// Select file for second slot
				msg := FileSelectedMsg{Path: "/second.iso", Canceled: false}
				m.handleFileSelectedCmd(msg)

				if m.cdroms[0] != "/first.iso" {
					t.Errorf("Expected first cdrom unchanged, got '%s'", m.cdroms[0])
				}
				if m.cdroms[1] != "/second.iso" {
					t.Errorf("Expected second cdrom='/second.iso', got '%s'", m.cdroms[1])
				}
				return
			}
		}
	}
	t.Fatal("Could not find second cdrom position")
}

// TestFilePickerViaMainModel tests file picker routing through MainModel
// using DiskAddedMsg for hard disk selection (the correct message type for AddDiskModel)
func TestFilePickerViaMainModel(t *testing.T) {
	m := setupEditModel(t)
	fm := getVMForm(m.vmEditModel.form)

	// Activate file picker for hardDisks
	positions := fm.BuildPositions()
	for _, p := range positions {
		if p.Kind == form.FocusList {
			fm.openFilePickerCmd(p)
			break
		}
	}

	if !fm.FileBrowserActive() {
		t.Fatal("Expected file browser to be active")
	}

	// Simulate disk added via DiskAddedMsg through MainModel
	// (hardDisks use AddDiskModel which produces DiskAddedMsg, not FileSelectedMsg)
	testFilePath := "/tmp/test-disk.qcow2"
	dam := DiskAddedMsg{Path: testFilePath, Canceled: false}
	m.Update(dam)

	fm = getVMForm(m.vmEditModel.form)
	if fm.hardDisks[0] != testFilePath {
		t.Errorf("Expected hardDisks[0]='%s', got '%s'", testFilePath, fm.hardDisks[0])
	}
}
