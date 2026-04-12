package vm

import (
	"strings"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// TestGenerateBuiltinScriptEmpty tests script generation with no devices
func TestGenerateBuiltinScriptEmpty(t *testing.T) {
	script, err := GenerateBuiltinScript(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(script, "No PCI devices configured") {
		t.Error("Expected 'No PCI devices configured' message")
	}
}

// TestGenerateBuiltinScriptSingleDevice tests script generation with one device
func TestGenerateBuiltinScriptSingleDevice(t *testing.T) {
	devices := []models.PCIPassthroughDevice{
		{Address: "0000:01:00.0", Name: "NVIDIA GPU"},
	}

	script, err := GenerateBuiltinScript(devices)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain shebang
	if !strings.HasPrefix(script, "#!/bin/bash") {
		t.Error("Expected bash shebang")
	}

	// Should contain device address
	if !strings.Contains(script, "0000:01:00.0") {
		t.Error("Expected device address in script")
	}

	// Should contain driver_override command (new-style binding)
	if !strings.Contains(script, "driver_override") {
		t.Error("Expected driver_override command")
	}
}

// TestGenerateBuiltinScriptWithROM tests script generation with ROM path
func TestGenerateBuiltinScriptWithROM(t *testing.T) {
	devices := []models.PCIPassthroughDevice{
		{Address: "0000:01:00.0", Name: "NVIDIA GPU", ROMPath: "/roms/gpu.rom"},
	}

	script, err := GenerateBuiltinScript(devices)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain ROM path
	if !strings.Contains(script, "/roms/gpu.rom") {
		t.Error("Expected ROM path in script")
	}

	// Should contain ROM echo command
	if !strings.Contains(script, "echo 1") || !strings.Contains(script, "rom") {
		t.Error("Expected ROM enable command")
	}
}

// TestGenerateBuiltinScriptMultipleDevices tests script generation with multiple devices
func TestGenerateBuiltinScriptMultipleDevices(t *testing.T) {
	devices := []models.PCIPassthroughDevice{
		{Address: "0000:01:00.0", Name: "NVIDIA GPU"},
		{Address: "0000:02:00.0", Name: "USB Controller"},
	}

	script, err := GenerateBuiltinScript(devices)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should contain both addresses
	if !strings.Contains(script, "0000:01:00.0") {
		t.Error("Expected first device address")
	}
	if !strings.Contains(script, "0000:02:00.0") {
		t.Error("Expected second device address")
	}

	// Should have two driver_override commands per device (clear + set)
	// With 2 devices, that's 4 commands total
	count := strings.Count(script, "driver_override")
	if count < 4 {
		t.Errorf("Expected 2 vfio-pci bind commands, got %d", count)
	}
}

// TestGenerateBuiltinStopScript tests stop script generation
func TestGenerateBuiltinStopScript(t *testing.T) {
	script := GenerateBuiltinStopScript()

	if !strings.HasPrefix(script, "#!/bin/bash") {
		t.Error("Expected bash shebang")
	}

	if !strings.Contains(script, "No cleanup needed") {
		t.Error("Expected 'No cleanup needed' message")
	}
}

// TestGenerateBuiltinScriptUnbind tests that script uses driver_override
func TestGenerateBuiltinScriptUnbind(t *testing.T) {
	devices := []models.PCIPassthroughDevice{
		{Address: "0000:01:00.0", Name: "NVIDIA GPU"},
	}

	script, err := GenerateBuiltinScript(devices)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should use driver_override mechanism
	if !strings.Contains(script, "driver_override") {
		t.Error("Expected driver_override in script")
	}

	// Should also contain vfio-pci assignment
	if !strings.Contains(script, "vfio-pci") {
		t.Error("Expected vfio-pci in script")
	}
}
