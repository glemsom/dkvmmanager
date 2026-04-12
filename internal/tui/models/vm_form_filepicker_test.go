package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOpenFilePickerHardDisk(t *testing.T) {
	m := setupTestForm(t)
	m.focusIndex = 2 // First hardDisk list item

	pos := m.currentPos()
	if pos.kind != focusListItem || pos.fieldName != "hardDisks" {
		t.Fatalf("Expected focus on hardDisks list item, got kind=%d fieldName=%s", pos.kind, pos.fieldName)
	}

	_, cmd := m.openFilePicker(pos)

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
	// AddDiskModel.Init() returns nil
	if cmd != nil {
		t.Error("Expected nil command from AddDiskModel Init()")
	}
}

func TestOpenFilePickerISO(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{""}
	m.rebuildPositions()

	// Find the cdrom list item position
	cdromIdx := -1
	for i, p := range m.positions {
		if p.kind == focusListItem && p.fieldName == "cdroms" {
			cdromIdx = i
			break
		}
	}
	if cdromIdx < 0 {
		t.Fatal("Could not find cdrom list item position")
	}
	m.focusIndex = cdromIdx

	pos := m.currentPos()
	_, cmd := m.openFilePicker(pos)

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
}

func TestFileBrowserActiveInitiallyFalse(t *testing.T) {
	m := setupTestForm(t)
	if m.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() to be false on fresh form")
	}
}

func TestFileBrowserActiveAfterOpenHardDisk(t *testing.T) {
	m := setupTestForm(t)
	m.focusIndex = 2 // First hardDisk item
	pos := m.currentPos()
	m.openFilePicker(pos)

	if !m.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() to be true after opening harddisk picker")
	}
}

func TestFileBrowserActiveAfterOpenISO(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{""}
	m.rebuildPositions()

	cdromIdx := -1
	for i, p := range m.positions {
		if p.kind == focusListItem && p.fieldName == "cdroms" {
			cdromIdx = i
			break
		}
	}
	m.focusIndex = cdromIdx
	pos := m.currentPos()
	m.openFilePicker(pos)

	if !m.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() to be true after opening ISO picker")
	}
}

func TestHandleFileSelectedSetsPath(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{""}
	m.browsingFieldName = "cdroms"
	m.browsingIndex = 0

	// Simulate an active file browser
	m.fileBrowser = NewFileBrowserModel(FileTypeISO)

	msg := FileSelectedMsg{Path: "/path/to/image.iso", Canceled: false}
	_, cmd := m.handleFileSelected(msg)

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

func TestHandleFileSelectedCanceled(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{"/existing/path.iso"}
	m.browsingFieldName = "cdroms"
	m.browsingIndex = 0

	m.fileBrowser = NewFileBrowserModel(FileTypeISO)

	msg := FileSelectedMsg{Path: "", Canceled: true}
	_, _ = m.handleFileSelected(msg)

	if m.fileBrowser != nil {
		t.Error("Expected fileBrowser to be cleared after cancel")
	}
	// Path should be unchanged
	if m.cdroms[0] != "/existing/path.iso" {
		t.Errorf("Expected cdroms[0] unchanged='/existing/path.iso', got '%s'", m.cdroms[0])
	}
}

func TestHandleDiskAddedSetsPath(t *testing.T) {
	m := setupTestForm(t)
	m.hardDisks = []string{""}
	m.browsingFieldName = "hardDisks"
	m.browsingIndex = 0

	m.addDiskModel = NewAddDiskModel(m.vmManager)

	msg := DiskAddedMsg{Path: "/dev/sda", Canceled: false}
	_, cmd := m.handleDiskAdded(msg)

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

func TestHandleDiskAddedCanceled(t *testing.T) {
	m := setupTestForm(t)
	m.hardDisks = []string{"/existing/disk.qcow2"}
	m.browsingFieldName = "hardDisks"
	m.browsingIndex = 0

	m.addDiskModel = NewAddDiskModel(m.vmManager)

	msg := DiskAddedMsg{Path: "", Canceled: true}
	_, _ = m.handleDiskAdded(msg)

	if m.addDiskModel != nil {
		t.Error("Expected addDiskModel to be cleared after cancel")
	}
	if m.hardDisks[0] != "/existing/disk.qcow2" {
		t.Errorf("Expected hardDisks[0] unchanged, got '%s'", m.hardDisks[0])
	}
}

func TestHandleEnterOpensPickerOnListItem(t *testing.T) {
	m := setupTestForm(t)
	m.focusIndex = 2 // First hardDisk item

	pos := m.currentPos()
	if pos.kind != focusListItem {
		t.Fatalf("Expected focusListItem, got %d", pos.kind)
	}

	_, cmd := m.handleEnter()

	if m.addDiskModel == nil {
		t.Fatal("Expected handleEnter to open file picker for harddisk list item")
	}
	// AddDiskModel.Init() returns nil
	_ = cmd
}

func TestKeyDelegationToFileBrowser(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{""}
	m.browsingFieldName = "cdroms"
	m.browsingIndex = 0

	// Create and activate file browser
	m.fileBrowser = NewFileBrowserModel(FileTypeISO)
	// Simulate Init having run (load directory)
	m.fileBrowser.loadDirectory()

	// Send ESC key - should be delegated to file browser
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
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

func TestKeyDelegationToAddDiskModel(t *testing.T) {
	m := setupTestForm(t)
	m.hardDisks = []string{""}
	m.browsingFieldName = "hardDisks"
	m.browsingIndex = 0

	m.addDiskModel = NewAddDiskModel(m.vmManager)

	// Send ESC key in step 0 - should be delegated to add disk model
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("Expected add disk model to return a command on ESC")
	}
	msg := cmd()
	dam, ok := msg.(DiskAddedMsg)
	if !ok {
		t.Fatalf("Expected DiskAddedMsg, got %T", msg)
	}
	if !dam.Canceled {
		t.Error("Expected ESC to produce canceled DiskAddedMsg")
	}
}

func TestViewShowsFileBrowserWhenActive(t *testing.T) {
	m := setupTestForm(t)
	m.contentW = 80
	m.contentH = 24
	m.vp.Width = 80
	m.vp.Height = 24
	m.ready = true

	// Without file browser, should show form
	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	// With active file browser, should show file browser view
	m.fileBrowser = NewFileBrowserModel(FileTypeISO)
	view = m.View()
	// File browser view contains "Select ISO Image"
	if view == "" {
		t.Error("Expected non-empty file browser view")
	}
}

func TestViewShowsAddDiskModelWhenActive(t *testing.T) {
	m := setupTestForm(t)
	m.contentW = 80
	m.contentH = 24
	m.vp.Width = 80
	m.vp.Height = 24
	m.ready = true

	m.addDiskModel = NewAddDiskModel(m.vmManager)
	view := m.View()
	// AddDiskModel step 0 view contains "Add Hard Disk"
	if view == "" {
		t.Error("Expected non-empty add disk view")
	}
}

func TestVMCreateModelFileBrowserActive(t *testing.T) {
	m := setupTestForm(t)
	create := &VMCreateModel{form: m}

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
	edit := &VMEditModel{form: m}

	if edit.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() false initially")
	}

	m.addDiskModel = NewAddDiskModel(m.vmManager)
	if !edit.FileBrowserActive() {
		t.Error("Expected FileBrowserActive() true after setting addDiskModel")
	}
}

func TestHandleEnterMovesFocusOnTextField(t *testing.T) {
	m := setupTestForm(t)
	m.focusIndex = 0 // vmName text field

	initialIndex := m.focusIndex
	m.handleEnter()

	if m.focusIndex == initialIndex {
		t.Error("Expected Enter on text field to move focus to next field")
	}
	if m.fileBrowser != nil {
		t.Error("Expected no file browser for text field Enter")
	}
	if m.addDiskModel != nil {
		t.Error("Expected no add disk model for text field Enter")
	}
}

func TestMultipleISOsFilePicker(t *testing.T) {
	m := setupTestForm(t)
	m.cdroms = []string{"/first.iso", ""}
	m.rebuildPositions()

	// Find second cdrom item
	cdromIdx := -1
	for i, p := range m.positions {
		if p.kind == focusListItem && p.fieldName == "cdroms" && p.listIndex == 1 {
			cdromIdx = i
			break
		}
	}
	if cdromIdx < 0 {
		t.Fatal("Could not find second cdrom position")
	}
	m.focusIndex = cdromIdx
	pos := m.currentPos()
	m.openFilePicker(pos)

	if m.browsingIndex != 1 {
		t.Errorf("Expected browsingIndex=1, got %d", m.browsingIndex)
	}

	// Select file for second slot
	msg := FileSelectedMsg{Path: "/second.iso", Canceled: false}
	m.handleFileSelected(msg)

	if m.cdroms[0] != "/first.iso" {
		t.Errorf("Expected first cdrom unchanged, got '%s'", m.cdroms[0])
	}
	if m.cdroms[1] != "/second.iso" {
		t.Errorf("Expected second cdrom='/second.iso', got '%s'", m.cdroms[1])
	}
}
