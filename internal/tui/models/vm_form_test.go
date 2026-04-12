package models

import (
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

			_, _ = m.validateAndSave()

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

			_, _ = m.validateAndSave()

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

			_, _ = m.validateAndSave()

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

	_, _ = m.validateAndSave()

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

	_, _ = m.validateAndSave()

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
