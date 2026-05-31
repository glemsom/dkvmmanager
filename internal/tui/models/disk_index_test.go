package models

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// TestSecondDiskInputGoesToCorrectSlot reproduces the bug where typing into
// the 2nd disk path field writes into the 1st disk slot instead.
func TestSecondDiskInputGoesToCorrectSlot(t *testing.T) {
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

	m := NewVMFormModel(mgr)

	// Start with one empty disk slot (default), add a second
	m.addItem("hardDisks_add")

	// Verify we have 2 disk slots
	if len(m.hardDisks) != 2 {
		t.Fatalf("Expected 2 hard disk slots, got %d", len(m.hardDisks))
	}

	// Rebuild positions so the new slot appears
	m.rebuildPositions()

	// Build positions and find the 2nd disk (hardDisks_1)
	positions := m.BuildPositions()
	var secondDiskPos *form.FocusPos
	for i := range positions {
		if positions[i].Key == "hardDisks_1" {
			secondDiskPos = &positions[i]
			break
		}
	}
	if secondDiskPos == nil {
		t.Fatal("Could not find hardDisks_1 position")
	}

	// Simulate typing a character into the 2nd disk slot
	m.HandleChar(*secondDiskPos, "/")

	// The 2nd disk slot should now have "/" and the 1st should remain empty
	if m.hardDisks[1] != "/" {
		t.Errorf("BUG: hardDisks[1] should be '/', got %q", m.hardDisks[1])
	}
	if m.hardDisks[0] != "" {
		t.Errorf("BUG: hardDisks[0] should remain empty, got %q (input went to wrong slot!)", m.hardDisks[0])
	}

	// Type more characters
	m.HandleChar(*secondDiskPos, "v")
	m.HandleChar(*secondDiskPos, "m")

	if m.hardDisks[1] != "/vm" {
		t.Errorf("hardDisks[1] should be '/vm', got %q", m.hardDisks[1])
	}
}

// TestGetValueForSecondDisk verifies getValue returns the correct disk value
func TestGetValueForSecondDisk(t *testing.T) {
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

	m := NewVMFormModel(mgr)
	m.hardDisks = []string{"/disk1.qcow2", "/disk2.qcow2"}
	m.rebuildPositions()

	positions := m.BuildPositions()
	for _, pos := range positions {
		if pos.Kind != form.FocusList {
			continue
		}
		val := m.getValue(pos)
		if pos.Key == "hardDisks_0" && val != "/disk1.qcow2" {
			t.Errorf("getValue(hardDisks_0) = %q, want %q", val, "/disk1.qcow2")
		}
		if pos.Key == "hardDisks_1" && val != "/disk2.qcow2" {
			t.Errorf("getValue(hardDisks_1) = %q, want %q", val, "/disk2.qcow2")
		}
	}
}

// TestSetValueForSecondDisk verifies setValue writes to the correct disk slot
func TestSetValueForSecondDisk(t *testing.T) {
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

	m := NewVMFormModel(mgr)
	m.hardDisks = []string{"/disk1.qcow2", "/disk2.qcow2"}
	m.rebuildPositions()

	positions := m.BuildPositions()
	for _, pos := range positions {
		if pos.Kind != form.FocusList {
			continue
		}
		if pos.Key == "hardDisks_1" {
			m.setValue(pos, "/new-disk2.qcow2")
			break
		}
	}

	if m.hardDisks[0] != "/disk1.qcow2" {
		t.Errorf("hardDisks[0] should be unchanged, got %q", m.hardDisks[0])
	}
	if m.hardDisks[1] != "/new-disk2.qcow2" {
		t.Errorf("hardDisks[1] should be '/new-disk2.qcow2', got %q", m.hardDisks[1])
	}
}

// TestCDROMIndexParsing verifies that CDROM indices are also parsed correctly
func TestCDROMIndexParsing(t *testing.T) {
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

	m := NewVMFormModel(mgr)
	m.cdroms = []string{"/iso1.iso", "/iso2.iso", "/iso3.iso"}
	m.rebuildPositions()

	positions := m.BuildPositions()
	for _, pos := range positions {
		if pos.Kind != form.FocusList {
			continue
		}
		val := m.getValue(pos)
		if pos.Key == "cdroms_0" && val != "/iso1.iso" {
			t.Errorf("getValue(cdroms_0) = %q, want %q", val, "/iso1.iso")
		}
		if pos.Key == "cdroms_1" && val != "/iso2.iso" {
			t.Errorf("getValue(cdroms_1) = %q, want %q", val, "/iso2.iso")
		}
		if pos.Key == "cdroms_2" && val != "/iso3.iso" {
			t.Errorf("getValue(cdroms_2) = %q, want %q", val, "/iso3.iso")
		}
	}
}
