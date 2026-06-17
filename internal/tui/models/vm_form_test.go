package models

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// setupTestForm creates a VMFormModel with a temporary vmManager for validation tests
func setupTestForm(t *testing.T) *VMFormModel {
	t.Helper()

	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "dkvmmanager"), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
		NetworkBridge: "virbr0",
	}

	mgr, err := vm.NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create vmManager: %v", err)
	}

	return NewVMFormModel(mgr)
}

func TestVMNameValidationValidNames(t *testing.T) {
	validNames := []string{
		"myvm",
		"my-vm",
		"my_vm",
		"MyVM1",
		"my vm",
		"My Gaming VM 1",
		"vm 2024",
		"A",
		"123",
		"test-vm_name 1",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			m := setupTestForm(t)
			m.vmName = name
			m.hardDisks = []string{"/tmp/test-disk.qcow2"}

			_ = m.validateAndSaveCmd()

			if err, ok := m.errors["vmName"]; ok {
				t.Errorf("Expected name %q to be valid, got error: %s", name, err)
			}
		})
	}
}

func TestVMNameValidationInvalidNames(t *testing.T) {
	invalidNames := []struct {
		name   string
		reason string
	}{
		{"vm!name", "exclamation mark"},
		{"vm@home", "at sign"},
		{"vm#1", "hash"},
		{"vm.name", "dot"},
		{"vm/name", "slash"},
		{"vm\\name", "backslash"},
		{"vm:name", "colon"},
		{"vm;name", "semicolon"},
		{"vm(name)", "parentheses"},
		{"vm[name]", "brackets"},
		{"vm{name}", "braces"},
		{"vm$name", "dollar sign"},
		{"vm%name", "percent"},
		{"vm^name", "caret"},
		{"vm&name", "ampersand"},
		{"vm*name", "asterisk"},
		{"vm=name", "equals"},
		{"vm+name", "plus"},
		{"vm|name", "pipe"},
		{"vm<name>", "angle brackets"},
		{"vm?name", "question mark"},
		{"vm~name", "tilde"},
		{"vm`name", "backtick"},
		{"vm'name", "single quote"},
		{"vm\"name", "double quote"},
		{"vm,name", "comma"},
	}

	for _, tc := range invalidNames {
		t.Run(tc.reason, func(t *testing.T) {
			m := setupTestForm(t)
			m.vmName = tc.name
			m.hardDisks = []string{"/tmp/test-disk.qcow2"}

			_ = m.validateAndSaveCmd()

			if _, ok := m.errors["vmName"]; !ok {
				t.Errorf("Expected name %q (%s) to be invalid, but no error was set", tc.name, tc.reason)
			}
		})
	}
}

func TestVMNameValidationEmptyAfterTrim(t *testing.T) {
	emptyAfterTrim := []string{
		"",
		"   ",
		"\t",
		"\n",
	}

	for _, name := range emptyAfterTrim {
		t.Run("empty_after_trim", func(t *testing.T) {
			m := setupTestForm(t)
			m.vmName = name
			m.hardDisks = []string{"/tmp/test-disk.qcow2"}

			_ = m.validateAndSaveCmd()

			err, ok := m.errors["vmName"]
			if !ok {
				t.Fatal("Expected vmName error for empty/whitespace-only name")
			}
			if err != "VM name cannot be empty" {
				t.Errorf("Expected 'VM name cannot be empty', got %q", err)
			}
		})
	}
}

func TestVMNameValidationTrimmedBeforeSave(t *testing.T) {
	m := setupTestForm(t)
	m.vmName = "  my vm  "
	m.hardDisks = []string{"/tmp/test-disk.qcow2"}

	_ = m.validateAndSaveCmd()

	if m.vmName != "my vm" {
		t.Errorf("Expected vmName to be trimmed to 'my vm', got %q", m.vmName)
	}

	if err, ok := m.errors["vmName"]; ok {
		t.Errorf("Expected trimmed name to be valid, got error: %s", err)
	}
}

func TestVMNameValidationErrorMessage(t *testing.T) {
	m := setupTestForm(t)
	m.vmName = "bad!name"
	m.hardDisks = []string{"/tmp/test-disk.qcow2"}

	_ = m.validateAndSaveCmd()

	err, ok := m.errors["vmName"]
	if !ok {
		t.Fatal("Expected vmName error for invalid name")
	}

	expected := "Only alphanumeric, dash, underscore, and space allowed"
	if err != expected {
		t.Errorf("Expected error %q, got %q", expected, err)
	}
}

func TestNetworkModeDefaultIsNAT(t *testing.T) {
	m := setupTestForm(t)
	if m.networkMode != "nat" {
		t.Errorf("Expected default network mode to be 'nat', got %q", m.networkMode)
	}
}

func TestNetworkModeToggle(t *testing.T) {
	m := setupTestForm(t)

	if m.networkMode != "nat" {
		t.Fatalf("Expected initial mode 'nat', got %q", m.networkMode)
	}

	m.toggleValue("networkMode")
	if m.networkMode != "bridge" {
		t.Errorf("Expected mode 'bridge' after toggle, got %q", m.networkMode)
	}

	m.toggleValue("networkMode")
	if m.networkMode != "nat" {
		t.Errorf("Expected mode 'nat' after second toggle, got %q", m.networkMode)
	}
}

func TestRemoveListAtMultiDigitIndex(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		disks    []string
		wantRemoved string // the value that should be removed
	}{
		{
			name:        "index 0 (single digit)",
			key:         "hardDisks_0",
			disks:       []string{"disk0", "disk1", "disk2"},
			wantRemoved: "disk0",
		},
		{
			name:        "index 10 (multi-digit)",
			key:         "hardDisks_10",
			disks:       []string{"disk0", "disk1", "disk2", "disk3", "disk4", "disk5", "disk6", "disk7", "disk8", "disk9", "disk10", "disk11"},
			wantRemoved: "disk10",
		},
		{
			name:        "index 42 (multi-digit not ending in 0)",
			key:         "hardDisks_42",
			disks: func() []string {
				d := make([]string, 43)
				for i := 0; i < 43; i++ {
					d[i] = fmt.Sprintf("disk%d", i)
				}
				return d
			}(),
			wantRemoved: "disk42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setupTestForm(t)
			m.hardDisks = make([]string, len(tt.disks))
			copy(m.hardDisks, tt.disks)

			// Need at least 2 items or removeListAt won't actually remove (it resets to empty)
			if len(m.hardDisks) <= 1 {
				m.hardDisks = append(m.hardDisks, "extra")
			}

			m.removeListAt("hardDisks", tt.key)

			// Check that the removed value is no longer present
			for _, d := range m.hardDisks {
				if d == tt.wantRemoved {
					t.Errorf("Value %q should have been removed but was found in hardDisks: %v", tt.wantRemoved, m.hardDisks)
				}
			}

			// Verify length decreased by 1
			wantLen := len(tt.disks) - 1
			if len(m.hardDisks) != wantLen {
				t.Errorf("Expected hardDisks length %d, got %d. hardDisks: %v", wantLen, len(m.hardDisks), m.hardDisks)
			}
		})
	}
}
