package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func setupBlockDeviceModel(t *testing.T, devices []BlockDevice) *BlockDeviceModel {
	t.Helper()
	return &BlockDeviceModel{
		devices:       devices,
		selectedIndex: 0,
		active:        true,
	}
}

func TestNewBlockDeviceModel(t *testing.T) {
	m := NewBlockDeviceModel()
	if m == nil {
		t.Fatal("NewBlockDeviceModel() returned nil")
	}
	if !m.active {
		t.Error("Expected active to be true")
	}
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", m.selectedIndex)
	}
}

func TestBlockDeviceModelHandleKeyPressUp(t *testing.T) {
	devices := []BlockDevice{
		{Path: "/dev/sda", Name: "sda", Size: "500G", Type: "disk"},
		{Path: "/dev/sdb", Name: "sdb", Size: "1T", Type: "disk"},
		{Path: "/dev/sdc", Name: "sdc", Size: "256G", Type: "disk"},
	}
	m := setupBlockDeviceModel(t, devices)
	m.selectedIndex = 2

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(*BlockDeviceModel)
	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after up, got %d", m.selectedIndex)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(*BlockDeviceModel)
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after up, got %d", m.selectedIndex)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(*BlockDeviceModel)
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 (bounded at 0), got %d", m.selectedIndex)
	}
}

func TestBlockDeviceModelHandleKeyPressDown(t *testing.T) {
	devices := []BlockDevice{
		{Path: "/dev/sda", Name: "sda", Size: "500G", Type: "disk"},
		{Path: "/dev/sdb", Name: "sdb", Size: "1T", Type: "disk"},
	}
	m := setupBlockDeviceModel(t, devices)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(*BlockDeviceModel)
	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after down, got %d", m.selectedIndex)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(*BlockDeviceModel)
	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 (bounded at len-1), got %d", m.selectedIndex)
	}
}

func TestBlockDeviceModelHandleKeyPressESC(t *testing.T) {
	m := setupBlockDeviceModel(t, []BlockDevice{
		{Path: "/dev/sda", Name: "sda", Size: "500G", Type: "disk"},
	})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(*BlockDeviceModel)

	if m.active {
		t.Error("Expected active to be false after ESC")
	}
	if cmd == nil {
		t.Fatal("Expected command after ESC")
	}
	msg := cmd()
	fsm, ok := msg.(FileSelectedMsg)
	if !ok {
		t.Fatalf("Expected FileSelectedMsg, got %T", msg)
	}
	if !fsm.Canceled {
		t.Error("Expected Canceled to be true")
	}
	if fsm.Path != "" {
		t.Errorf("Expected empty path, got '%s'", fsm.Path)
	}
}

func TestBlockDeviceModelHandleKeyPressCtrlC(t *testing.T) {
	m := setupBlockDeviceModel(t, []BlockDevice{
		{Path: "/dev/sda", Name: "sda", Size: "500G", Type: "disk"},
	})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(*BlockDeviceModel)

	if m.active {
		t.Error("Expected active to be false after Ctrl+C")
	}
	if cmd == nil {
		t.Fatal("Expected command after Ctrl+C")
	}
	msg := cmd()
	fsm, ok := msg.(FileSelectedMsg)
	if !ok {
		t.Fatalf("Expected FileSelectedMsg, got %T", msg)
	}
	if !fsm.Canceled {
		t.Error("Expected Canceled to be true")
	}
}

func TestBlockDeviceModelHandleEnter(t *testing.T) {
	devices := []BlockDevice{
		{Path: "/dev/sda", Name: "sda", Size: "500G", Type: "disk"},
		{Path: "/dev/sdb", Name: "sdb", Size: "1T", Type: "disk"},
	}
	m := setupBlockDeviceModel(t, devices)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*BlockDeviceModel)

	if m.active {
		t.Error("Expected active to be false after Enter")
	}
	if m.selectedPath != "/dev/sda" {
		t.Errorf("Expected selectedPath '/dev/sda', got '%s'", m.selectedPath)
	}
	if cmd == nil {
		t.Fatal("Expected command after Enter")
	}
	msg := cmd()
	fsm, ok := msg.(FileSelectedMsg)
	if !ok {
		t.Fatalf("Expected FileSelectedMsg, got %T", msg)
	}
	if fsm.Canceled {
		t.Error("Expected Canceled to be false")
	}
	if fsm.Path != "/dev/sda" {
		t.Errorf("Expected path '/dev/sda', got '%s'", fsm.Path)
	}
}

func TestBlockDeviceModelHandleEnterEmpty(t *testing.T) {
	m := setupBlockDeviceModel(t, nil)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(*BlockDeviceModel)

	if !m.active {
		t.Error("Expected active to remain true with no devices")
	}
	if cmd != nil {
		t.Error("Expected nil command with no devices")
	}
}

func TestBlockDeviceModelUpdateInactive(t *testing.T) {
	m := setupBlockDeviceModel(t, []BlockDevice{
		{Path: "/dev/sda", Name: "sda", Size: "500G", Type: "disk"},
	})
	m.active = false

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(*BlockDeviceModel)

	if cmd != nil {
		t.Error("Expected nil command for inactive model")
	}
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex unchanged, got %d", m.selectedIndex)
	}
}

func TestBlockDeviceModelViewContainsHeader(t *testing.T) {
	m := setupBlockDeviceModel(t, []BlockDevice{
		{Path: "/dev/sda", Name: "sda", Size: "500G", Type: "disk"},
	})

	view := m.View()
	if !strings.Contains(view, "Select Block Device") {
		t.Error("View should contain 'Select Block Device'")
	}
}

func TestBlockDeviceModelViewEmpty(t *testing.T) {
	m := setupBlockDeviceModel(t, nil)

	view := m.View()
	if !strings.Contains(view, "(no block devices found)") {
		t.Error("View should contain '(no block devices found)' for empty device list")
	}
}

func TestBlockDeviceModelViewDevice(t *testing.T) {
	m := setupBlockDeviceModel(t, []BlockDevice{
		{Path: "/dev/sda", Name: "sda", Size: "500G", Type: "disk"},
	})

	view := m.View()
	if !strings.Contains(view, "sda") {
		t.Error("View should contain device name 'sda'")
	}
	if !strings.Contains(view, "500G") {
		t.Error("View should contain device size '500G'")
	}
	if !strings.Contains(view, "disk") {
		t.Error("View should contain device type 'disk'")
	}
}

func TestBlockDeviceModelViewReadOnly(t *testing.T) {
	m := setupBlockDeviceModel(t, []BlockDevice{
		{Path: "/dev/sda", Name: "sda", Size: "500G", Type: "disk", ReadOnly: true},
	})

	view := m.View()
	if !strings.Contains(view, "[RO]") {
		t.Error("View should contain '[RO]' for read-only device")
	}
}

func TestBlockDeviceModelViewSelected(t *testing.T) {
	m := setupBlockDeviceModel(t, []BlockDevice{
		{Path: "/dev/sda", Name: "sda", Size: "500G", Type: "disk"},
		{Path: "/dev/sdb", Name: "sdb", Size: "1T", Type: "disk"},
	})

	view := m.View()
	lines := strings.Split(view, "\n")
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "> ") {
			found = true
			if !strings.Contains(trimmed, "sda") {
				t.Error("Selected line should contain 'sda'")
			}
		}
	}
	if !found {
		t.Error("View should show '>' prefix for selected device")
	}
}

func TestBlockDeviceModelGetSelectedPath(t *testing.T) {
	m := &BlockDeviceModel{
		selectedPath: "/dev/sda",
	}
	if m.GetSelectedPath() != "/dev/sda" {
		t.Errorf("Expected '/dev/sda', got '%s'", m.GetSelectedPath())
	}
}

func TestParseLSBlkOutputBasic(t *testing.T) {
	m := NewBlockDeviceModel()
	output := "sda      500G disk 0\nsda1     450G part 0\nnvme0n1    1T disk 0"

	devices, err := m.parseLSBlkOutput(output)
	if err != nil {
		t.Fatalf("parseLSBlkOutput() error: %v", err)
	}
	if len(devices) != 3 {
		t.Fatalf("Expected 3 devices, got %d", len(devices))
	}
	if devices[0].Name != "nvme0n1" {
		t.Errorf("Expected first device 'nvme0n1' (sorted), got '%s'", devices[0].Name)
	}
	if devices[0].Size != "1T" {
		t.Errorf("Expected size '1T', got '%s'", devices[0].Size)
	}
	if devices[0].Type != "disk" {
		t.Errorf("Expected type 'disk', got '%s'", devices[0].Type)
	}
	if devices[0].Path != "/dev/nvme0n1" {
		t.Errorf("Expected path '/dev/nvme0n1', got '%s'", devices[0].Path)
	}
}

func TestParseLSBlkOutputPartitions(t *testing.T) {
	m := NewBlockDeviceModel()
	output := "sda      500G disk 0\nsda1     450G part 0\nsda2     50G  part 0"

	devices, err := m.parseLSBlkOutput(output)
	if err != nil {
		t.Fatalf("parseLSBlkOutput() error: %v", err)
	}
	if len(devices) != 3 {
		t.Fatalf("Expected 3 devices, got %d", len(devices))
	}

	for _, dev := range devices {
		if dev.Type != "disk" && dev.Type != "part" {
			t.Errorf("Expected type 'disk' or 'part', got '%s'", dev.Type)
		}
	}
}

func TestParseLSBlkOutputReadOnly(t *testing.T) {
	m := NewBlockDeviceModel()
	output := "sda      500G disk 1\nnvme0n1    1T disk 0"

	devices, err := m.parseLSBlkOutput(output)
	if err != nil {
		t.Fatalf("parseLSBlkOutput() error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("Expected 2 devices, got %d", len(devices))
	}
	if devices[0].Name != "nvme0n1" {
		t.Errorf("Expected first device 'nvme0n1', got '%s'", devices[0].Name)
	}
	if devices[1].Name != "sda" {
		t.Errorf("Expected second device 'sda', got '%s'", devices[1].Name)
	}
	if !devices[1].ReadOnly {
		t.Error("Expected sda to be read-only")
	}
	if devices[0].ReadOnly {
		t.Error("Expected nvme0n1 to not be read-only")
	}
}

func TestParseLSBlkOutputIgnoresLoop(t *testing.T) {
	m := NewBlockDeviceModel()
	output := "sda      500G disk 0\nloop0     100M loop 0\nloop1     200M loop 0"

	devices, err := m.parseLSBlkOutput(output)
	if err != nil {
		t.Fatalf("parseLSBlkOutput() error: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("Expected 1 device (loop ignored), got %d", len(devices))
	}
	if devices[0].Name != "sda" {
		t.Errorf("Expected 'sda', got '%s'", devices[0].Name)
	}
}

func TestParseLSBlkOutputEmpty(t *testing.T) {
	m := NewBlockDeviceModel()

	devices, err := m.parseLSBlkOutput("")
	if err != nil {
		t.Fatalf("parseLSBlkOutput() error: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("Expected 0 devices for empty input, got %d", len(devices))
	}
}

func TestParseLSBlkOutputSorted(t *testing.T) {
	m := NewBlockDeviceModel()
	output := "sdb      1T   disk 0\nsda      500G disk 0\nnvme0n1   256G disk 0"

	devices, err := m.parseLSBlkOutput(output)
	if err != nil {
		t.Fatalf("parseLSBlkOutput() error: %v", err)
	}
	if len(devices) != 3 {
		t.Fatalf("Expected 3 devices, got %d", len(devices))
	}
	if devices[0].Name != "nvme0n1" {
		t.Errorf("Expected first 'nvme0n1', got '%s'", devices[0].Name)
	}
	if devices[1].Name != "sda" {
		t.Errorf("Expected second 'sda', got '%s'", devices[1].Name)
	}
	if devices[2].Name != "sdb" {
		t.Errorf("Expected third 'sdb', got '%s'", devices[2].Name)
	}
}

func TestParseLSBlkOutputMalformedLines(t *testing.T) {
	m := NewBlockDeviceModel()
	output := "sda      500G disk 0\nincomplete\ntoo short\nnvme0n1    1T disk 0"

	devices, err := m.parseLSBlkOutput(output)
	if err != nil {
		t.Fatalf("parseLSBlkOutput() error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("Expected 2 devices (malformed skipped), got %d", len(devices))
	}
	if devices[0].Name != "nvme0n1" {
		t.Errorf("Expected first 'nvme0n1', got '%s'", devices[0].Name)
	}
	if devices[1].Name != "sda" {
		t.Errorf("Expected second 'sda', got '%s'", devices[1].Name)
	}
}

func TestBlockDeviceModelViewError(t *testing.T) {
	m := setupBlockDeviceModel(t, nil)
	m.errorMsg = "permission denied"

	view := m.View()
	if !strings.Contains(view, "Error: permission denied") {
		t.Error("View should contain error message when errorMsg is set")
	}
}
