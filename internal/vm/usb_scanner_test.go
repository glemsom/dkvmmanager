package vm

import (
	"testing"
)

func TestUSBScannerParseLsusbOutput(t *testing.T) {
	scanner := &USBScanner{
		sysfsPath: "/nonexistent",
		lsusbPath: "lsusb",
	}

	lsusbOutput := `Bus 001 Device 001: ID 1d6b:0002 Linux Foundation 2.0 root hub
Bus 001 Device 005: ID 046d:c52b Logitech, Inc. Unifying Receiver
Bus 002 Device 003: ID 8087:0024 Intel Corp. Rate Matching Hub
Bus 003 Device 002: ID 045e:028e Microsoft Corp. Xbox360 Controller`

	devices := scanner.parseLsusbOutput(lsusbOutput)

	if len(devices) != 4 {
		t.Fatalf("Expected 4 devices, got %d", len(devices))
	}

	// Linux root hub
	dev := devices[0]
	if dev.Vendor != "1d6b" {
		t.Errorf("Vendor = %s, want 1d6b", dev.Vendor)
	}
	if dev.Product != "0002" {
		t.Errorf("Product = %s, want 0002", dev.Product)
	}
	if dev.Name != "Linux Foundation 2.0 root hub" {
		t.Errorf("Name = %s, want Linux Foundation 2.0 root hub", dev.Name)
	}

	// Logitech receiver
	dev = devices[1]
	if dev.Vendor != "046d" {
		t.Errorf("Vendor = %s, want 046d", dev.Vendor)
	}
	if dev.Product != "c52b" {
		t.Errorf("Product = %s, want c52b", dev.Product)
	}
	if dev.Name != "Logitech, Inc. Unifying Receiver" {
		t.Errorf("Name = %s, want Logitech, Inc. Unifying Receiver", dev.Name)
	}

	// Xbox controller
	dev = devices[3]
	if dev.Vendor != "045e" {
		t.Errorf("Vendor = %s, want 045e", dev.Vendor)
	}
	if dev.Product != "028e" {
		t.Errorf("Product = %s, want 028e", dev.Product)
	}
	if dev.Name != "Microsoft Corp. Xbox360 Controller" {
		t.Errorf("Name = %s, want Microsoft Corp. Xbox360 Controller", dev.Name)
	}
}

func TestUSBScannerParseEmptyOutput(t *testing.T) {
	scanner := &USBScanner{
		sysfsPath: "/nonexistent",
		lsusbPath: "lsusb",
	}

	devices := scanner.parseLsusbOutput("")
	if len(devices) != 0 {
		t.Errorf("Expected 0 devices for empty output, got %d", len(devices))
	}
}

func TestUSBScannerParseMalformedLines(t *testing.T) {
	scanner := &USBScanner{
		sysfsPath: "/nonexistent",
		lsusbPath: "lsusb",
	}

	output := `Bus 001 Device 005: ID 046d:c52b Logitech, Inc. Unifying Receiver
some garbage line
Bus invalid
Bus 002 Device 003: ID 8087:0024 Intel Corp. Rate Matching Hub`

	devices := scanner.parseLsusbOutput(output)

	if len(devices) != 2 {
		t.Errorf("Expected 2 devices with malformed lines, got %d", len(devices))
	}
}

func TestValidateUSBDevices(t *testing.T) {
	// Test with empty list (should pass)
	warnings, errors := ValidateUSBDevices(nil)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for nil devices, got %v", errors)
	}
	if len(warnings) != 0 {
		t.Errorf("Expected no warnings for nil devices, got %v", warnings)
	}
}
