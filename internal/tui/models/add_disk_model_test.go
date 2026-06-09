package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

func setupAddDiskTest(t *testing.T) (*AddDiskModel, *vm.Manager) {
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
		t.Fatalf("Failed to create VM manager: %v", err)
	}

	return NewAddDiskModel(mgr), mgr
}

func TestNewAddDiskModel(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	if m.step != 0 {
		t.Errorf("Expected step 0, got %d", m.step)
	}
	if !m.active {
		t.Error("Expected active to be true")
	}
	if m.fileBrowser == nil {
		t.Error("Expected fileBrowser to be created")
	}
	if m.blockDevice == nil {
		t.Error("Expected blockDevice to be created")
	}
	if m.lvmVolume == nil {
		t.Error("Expected lvmVolume to be created")
	}
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", m.selectedIndex)
	}
}

func TestAddDiskModelInit(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestAddDiskModelHandleKeyPressUp(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	m.selectedIndex = 1
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	m = updated.(*AddDiskModel)

	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after up, got %d", m.selectedIndex)
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}

	updated, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	m = updated.(*AddDiskModel)

	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 (bounded), got %d", m.selectedIndex)
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestAddDiskModelHandleKeyPressDown(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = updated.(*AddDiskModel)

	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after down, got %d", m.selectedIndex)
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}

	updated, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = updated.(*AddDiskModel)

	// Now there are 3 options: Disk image file, Block device, LVM Logical Volume
	// So selectedIndex should be 2 after second down
	if m.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex 2 (bounded to 2), got %d", m.selectedIndex)
	}
	if cmd != nil {
		t.Error("Expected nil command")
	}
}

func TestAddDiskModelHandleKeyPressESC(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = updated.(*AddDiskModel)

	if m.active {
		t.Error("Expected active to be false after ESC")
	}
	if cmd == nil {
		t.Fatal("Expected command after ESC")
	}

	msg := cmd()
	dam, ok := msg.(DiskAddedMsg)
	if !ok {
		t.Fatalf("Expected DiskAddedMsg, got %T", msg)
	}
	if !dam.Canceled {
		t.Error("Expected Canceled to be true")
	}
}

func TestAddDiskModelHandleKeyPressCtrlC(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	m = updated.(*AddDiskModel)

	if m.active {
		t.Error("Expected active to be false after Ctrl+C")
	}
	if cmd == nil {
		t.Fatal("Expected command after Ctrl+C")
	}

	msg := cmd()
	dam, ok := msg.(DiskAddedMsg)
	if !ok {
		t.Fatalf("Expected DiskAddedMsg, got %T", msg)
	}
	if !dam.Canceled {
		t.Error("Expected Canceled to be true")
	}
}

func TestAddDiskModelHandleEnterStep0File(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(*AddDiskModel)

	if m.sourceType != DiskSourceFile {
		t.Errorf("Expected sourceType DiskSourceFile, got %d", m.sourceType)
	}
	if m.step != 1 {
		t.Errorf("Expected step 1, got %d", m.step)
	}
	if cmd == nil {
		t.Error("Expected Init command from fileBrowser")
	}
}

func TestAddDiskModelHandleEnterStep0Device(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	m.selectedIndex = 1
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(*AddDiskModel)

	if m.sourceType != DiskSourceDevice {
		t.Errorf("Expected sourceType DiskSourceDevice, got %d", m.sourceType)
	}
	if m.step != 2 {
		t.Errorf("Expected step 2, got %d", m.step)
	}
	if cmd == nil {
		t.Error("Expected Init command from blockDevice")
	}
}

func TestAddDiskModelHandleEnterStep0LVM(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	m.selectedIndex = 2
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = updated.(*AddDiskModel)

	if m.sourceType != DiskSourceLVM {
		t.Errorf("Expected sourceType DiskSourceLVM, got %d", m.sourceType)
	}
	if m.step != 3 {
		t.Errorf("Expected step 3, got %d", m.step)
	}
	if cmd == nil {
		t.Error("Expected Init command from lvmVolume")
	}
}

func TestAddDiskModelHandleFileSelected(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	testPath := "/tmp/test-disk.qcow2"
	fsm := FileSelectedMsg{Path: testPath, Canceled: false}
	updated, cmd := m.Update(fsm)
	m = updated.(*AddDiskModel)

	if m.path != testPath {
		t.Errorf("Expected path '%s', got '%s'", testPath, m.path)
	}
	if m.active {
		t.Error("Expected active to be false")
	}
	if cmd == nil {
		t.Fatal("Expected command after FileSelectedMsg")
	}

	msg := cmd()
	dam, ok := msg.(DiskAddedMsg)
	if !ok {
		t.Fatalf("Expected DiskAddedMsg, got %T", msg)
	}
	if dam.Path != testPath {
		t.Errorf("Expected path '%s', got '%s'", testPath, dam.Path)
	}
	if dam.Canceled {
		t.Error("Expected Canceled to be false")
	}
}

func TestAddDiskModelHandleFileSelectedCanceled(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	fsm := FileSelectedMsg{Path: "", Canceled: true}
	updated, cmd := m.Update(fsm)
	m = updated.(*AddDiskModel)

	if m.active {
		t.Error("Expected active to be false")
	}
	if cmd == nil {
		t.Fatal("Expected command after canceled FileSelectedMsg")
	}

	msg := cmd()
	dam, ok := msg.(DiskAddedMsg)
	if !ok {
		t.Fatalf("Expected DiskAddedMsg, got %T", msg)
	}
	if !dam.Canceled {
		t.Error("Expected Canceled to be true")
	}
}

func TestAddDiskModelUpdateInactive(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	m.active = false
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_ = updated

	if cmd != nil {
		t.Error("Expected nil command for inactive model")
	}
}

func TestAddDiskModelUpdateFileSelectedWhenInactive(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	m.active = false
	testPath := "/tmp/test-disk.qcow2"
	fsm := FileSelectedMsg{Path: testPath, Canceled: false}
	updated, cmd := m.Update(fsm)
	m = updated.(*AddDiskModel)

	if cmd == nil {
		t.Fatal("Expected command even when inactive")
	}

	msg := cmd()
	dam, ok := msg.(DiskAddedMsg)
	if !ok {
		t.Fatalf("Expected DiskAddedMsg, got %T", msg)
	}
	if dam.Path != testPath {
		t.Errorf("Expected path '%s', got '%s'", testPath, dam.Path)
	}
}

func TestAddDiskModelUpdateDelegatesToFileBrowser(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	m.step = 1
	m.fileBrowser = NewFileBrowserModel(FileTypeDiskImage)
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = updated.(*AddDiskModel)

	if m.step != 1 {
		t.Errorf("Expected step to remain 1, got %d", m.step)
	}
}

func TestAddDiskModelUpdateDelegatesToBlockDevice(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	m.step = 2
	m.blockDevice = NewBlockDeviceModel()
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m = updated.(*AddDiskModel)

	if m.step != 2 {
		t.Errorf("Expected step to remain 2, got %d", m.step)
	}
}

func TestAddDiskModelViewStep0(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Add Hard Disk") {
		t.Error("View should contain 'Add Hard Disk'")
	}
	if !strings.Contains(viewContent, "Select source type") {
		t.Error("View should contain 'Select source type'")
	}
	if !strings.Contains(viewContent, "Enter Select") {
		t.Error("View should contain 'Enter Select'")
	}
}

func TestAddDiskModelViewStep0ContainsOptions(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Disk image file") {
		t.Error("View should contain 'Disk image file'")
	}
	if !strings.Contains(viewContent, "Block device") {
		t.Error("View should contain 'Block device'")
	}
	if !strings.Contains(viewContent, "LVM Logical Volume") {
		t.Error("View should contain 'LVM Logical Volume'")
	}
}

func TestAddDiskModelViewStep1(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	m.step = 1
	m.fileBrowser = NewFileBrowserModel(FileTypeDiskImage)

	fbViewContent := m.fileBrowser.View().Content
	viewContent := m.View().Content

	if viewContent == "" {
		t.Error("View should not be empty at step 1")
	}
	_ = fbViewContent
}

func TestAddDiskModelViewStep2(t *testing.T) {
	m, _ := setupAddDiskTest(t)

	m.step = 2
	m.blockDevice = NewBlockDeviceModel()

	bdViewContent := m.blockDevice.View().Content
	viewContent := m.View().Content

	if viewContent == "" {
		t.Error("View should not be empty at step 2")
	}
	_ = bdViewContent
}
