package vm

import (
	"testing"
)

func TestPCIScannerParseLspciOutput(t *testing.T) {
	scanner := &PCIScanner{
		sysfsPath: "/nonexistent",
		lspciPath: "lspci",
	}

	lspciOutput := `0000:00:02.0 VGA compatible controller [0300]: Intel Corporation HD Graphics 530 [8086:1912] (rev 06)
0000:01:00.0 VGA compatible controller [0300]: Advanced Micro Devices, Inc. [AMD/ATI] Navi 48 [1002:7550] (rev c0)
0000:01:00.1 Audio device [0403]: Advanced Micro Devices, Inc. [AMD/ATI] Navi 48 HDMI/DP Audio [1002:ab28]
0000:00:1f.3 Audio device [0403]: Intel Corporation 100 Series/C230 Series Chipset Family HD Audio Controller [8086:a170] (rev 31)
0000:00:14.0 USB controller [0c03]: Intel Corporation 100 Series/C230 Series Chipset Family USB xHCI [8086:a12f] (rev 31)
0000:03:00.0 Non-Volatile memory controller [0108]: Samsung Electronics Co Ltd NVMe SSD Controller SM981/PM981/PM983 [144d:a808]`

	devices := scanner.parseLspciOutput(lspciOutput)

	if len(devices) != 6 {
		t.Fatalf("Expected 6 devices, got %d", len(devices))
	}

	// Intel GPU
	intelGPU := devices[0]
	if intelGPU.Address != "0000:00:02.0" {
		t.Errorf("Intel GPU address = %s, want 0000:00:02.0", intelGPU.Address)
	}
	if intelGPU.Vendor != "8086" {
		t.Errorf("Intel GPU vendor = %s, want 8086", intelGPU.Vendor)
	}
	if intelGPU.Device != "1912" {
		t.Errorf("Intel GPU device = %s, want 1912", intelGPU.Device)
	}
	if intelGPU.ClassCode != "0300" {
		t.Errorf("Intel GPU class = %s, want 0300", intelGPU.ClassCode)
	}
	if !intelGPU.IsGPU {
		t.Error("Intel GPU IsGPU should be true")
	}
	if intelGPU.IsUSB {
		t.Error("Intel GPU IsUSB should be false")
	}

	// AMD GPU
	amdGPU := devices[1]
	if amdGPU.Address != "0000:01:00.0" {
		t.Errorf("AMD GPU address = %s, want 0000:01:00.0", amdGPU.Address)
	}
	if amdGPU.Vendor != "1002" {
		t.Errorf("AMD GPU vendor = %s, want 1002", amdGPU.Vendor)
	}
	if amdGPU.Device != "7550" {
		t.Errorf("AMD GPU device = %s, want 7550", amdGPU.Device)
	}
	if !amdGPU.IsGPU {
		t.Error("AMD GPU IsGPU should be true")
	}

	// AMD Audio
	amdAudio := devices[2]
	if amdAudio.Address != "0000:01:00.1" {
		t.Errorf("AMD Audio address = %s, want 0000:01:00.1", amdAudio.Address)
	}
	if amdAudio.ClassCode != "0403" {
		t.Errorf("AMD Audio class = %s, want 0403", amdAudio.ClassCode)
	}
	if amdAudio.IsGPU {
		t.Error("AMD Audio IsGPU should be false")
	}
	if amdAudio.IsUSB {
		t.Error("AMD Audio IsUSB should be false")
	}

	// USB controller
	usb := devices[4]
	if usb.Address != "0000:00:14.0" {
		t.Errorf("USB address = %s, want 0000:00:14.0", usb.Address)
	}
	if !usb.IsUSB {
		t.Error("USB controller IsUSB should be true")
	}
	if usb.IsGPU {
		t.Error("USB controller IsGPU should be false")
	}

	// NVMe
	nvme := devices[5]
	if nvme.ClassCode != "0108" {
		t.Errorf("NVMe class = %s, want 0108", nvme.ClassCode)
	}
	if nvme.IsGPU {
		t.Error("NVMe IsGPU should be false")
	}
	if nvme.IsUSB {
		t.Error("NVMe IsUSB should be false")
	}
}

func TestPCIScannerParseEmptyOutput(t *testing.T) {
	scanner := &PCIScanner{
		sysfsPath: "/nonexistent",
		lspciPath: "lspci",
	}

	devices := scanner.parseLspciOutput("")
	if len(devices) != 0 {
		t.Errorf("Expected 0 devices from empty output, got %d", len(devices))
	}
}

func TestPCIScannerParseGarbageOutput(t *testing.T) {
	scanner := &PCIScanner{
		sysfsPath: "/nonexistent",
		lspciPath: "lspci",
	}

	devices := scanner.parseLspciOutput("some garbage\nnot a pci line\n")
	if len(devices) != 0 {
		t.Errorf("Expected 0 devices from garbage output, got %d", len(devices))
	}
}

func TestIsMultifunction(t *testing.T) {
	tests := []struct {
		addr1    string
		addr2    string
		expected bool
	}{
		{"0000:01:00.0", "0000:01:00.1", true},  // Same bus:device
		{"0000:01:00.0", "0000:01:00.2", true},  // Same bus:device
		{"0000:01:00.0", "0000:02:00.0", false}, // Different bus:device
		{"0000:01:00.0", "0000:01:01.0", false}, // Different device
		{"0000:00:02.0", "0000:01:00.0", false}, // Completely different
		{"invalid", "0000:01:00.0", false},      // Invalid address
	}

	for _, tt := range tests {
		result := IsMultifunction(tt.addr1, tt.addr2)
		if result != tt.expected {
			t.Errorf("IsMultifunction(%q, %q) = %v, want %v",
				tt.addr1, tt.addr2, result, tt.expected)
		}
	}
}

func TestReadIOMMUGroupNoSysfs(t *testing.T) {
	scanner := &PCIScanner{
		sysfsPath: "/nonexistent/sys/bus/pci/devices",
		lspciPath: "lspci",
	}

	group := scanner.readIOMMUGroup("0000:01:00.0")
	if group != -1 {
		t.Errorf("Expected -1 for nonexistent sysfs path, got %d", group)
	}
}

func TestGetIOMMUGroupDevicesNoSysfs(t *testing.T) {
	addrs := GetIOMMUGroupDevices(1)
	if addrs != nil {
		t.Errorf("Expected nil for nonexistent sysfs path, got %v", addrs)
	}
}

func TestGetIOMMUGroupDevicesNegative(t *testing.T) {
	addrs := GetIOMMUGroupDevices(-1)
	if addrs != nil {
		t.Errorf("Expected nil for negative IOMMU group, got %v", addrs)
	}
}
