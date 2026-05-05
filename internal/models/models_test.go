// Package models provides tests for data models
package models

import (
	"testing"
	"time"
)

func TestVMFields(t *testing.T) {
	vm := VM{
		ID:        "1",
		Name:      "test-vm",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		HardDisks: []string{"/dev/sda", "/dev/sdb"},
		CDROMs:    []string{"/isos/windows.iso"},
		GPUROM:    "/roms/gpu.rom",
		MAC:       "00:11:22:33:44:55",
		VNCListen: "0.0.0.0:0",
	}

	if vm.ID != "1" {
		t.Errorf("VM.ID = %v, want 1", vm.ID)
	}
	if vm.Name != "test-vm" {
		t.Errorf("VM.Name = %v, want test-vm", vm.Name)
	}
	if len(vm.HardDisks) != 2 {
		t.Errorf("VM.HardDisks length = %v, want 2", len(vm.HardDisks))
	}
	if len(vm.CDROMs) != 1 {
		t.Errorf("VM.CDROMs length = %v, want 1", len(vm.CDROMs))
	}
	if vm.GPUROM != "/roms/gpu.rom" {
		t.Errorf("VM.GPUROM = %v, want /roms/gpu.rom", vm.GPUROM)
	}
	if vm.MAC != "00:11:22:33:44:55" {
		t.Errorf("VM.MAC = %v, want 00:11:22:33:44:55", vm.MAC)
	}
	if vm.VNCListen != "0.0.0.0:0" {
		t.Errorf("VM.VNCListen = %v, want 0.0.0.0:0", vm.VNCListen)
	}
}

func TestPCIDeviceFields(t *testing.T) {
	pci := PCIDevice{
		Address:    "0000:01:00.0",
		Vendor:     "10de",
		Device:     "1b80",
		ClassCode:  "0300",
		Name:       "NVIDIA GeForce GTX 1080",
		IsGPU:      true,
		IsUSB:      false,
		IsBridge:   false,
		IOMMUGroup: 1,
	}

	if pci.Address != "0000:01:00.0" {
		t.Errorf("PCIDevice.Address = %v, want 0000:01:00.0", pci.Address)
	}
	if pci.Vendor != "10de" {
		t.Errorf("PCIDevice.Vendor = %v, want 10de", pci.Vendor)
	}
	if pci.Device != "1b80" {
		t.Errorf("PCIDevice.Device = %v, want 1b80", pci.Device)
	}
	if pci.ClassCode != "0300" {
		t.Errorf("PCIDevice.ClassCode = %v, want 0300", pci.ClassCode)
	}
	if pci.Name != "NVIDIA GeForce GTX 1080" {
		t.Errorf("PCIDevice.Name = %v, want NVIDIA GeForce GTX 1080", pci.Name)
	}
	if !pci.IsGPU {
		t.Errorf("PCIDevice.IsGPU = %v, want true", pci.IsGPU)
	}
	if pci.IsUSB {
		t.Errorf("PCIDevice.IsUSB = %v, want false", pci.IsUSB)
	}
	if pci.IsBridge {
		t.Errorf("PCIDevice.IsBridge = %v, want false", pci.IsBridge)
	}
	if pci.IOMMUGroup != 1 {
		t.Errorf("PCIDevice.IOMMUGroup = %v, want 1", pci.IOMMUGroup)
	}
}

func TestUSBDeviceFields(t *testing.T) {
	usb := USBDevice{
		ID:       "1-1",
		Vendor:   "046d",
		Product:  "c52b",
		Name:     "Logitech Unifying Receiver",
		Selected: false,
	}

	if usb.ID != "1-1" {
		t.Errorf("USBDevice.ID = %v, want 1-1", usb.ID)
	}
	if usb.Vendor != "046d" {
		t.Errorf("USBDevice.Vendor = %v, want 046d", usb.Vendor)
	}
	if usb.Product != "c52b" {
		t.Errorf("USBDevice.Product = %v, want c52b", usb.Product)
	}
	if usb.Name != "Logitech Unifying Receiver" {
		t.Errorf("USBDevice.Name = %v, want Logitech Unifying Receiver", usb.Name)
	}
	if usb.Selected {
		t.Errorf("USBDevice.Selected = %v, want false", usb.Selected)
	}
}

func TestCPUTopologyFields(t *testing.T) {
	topo := CPUTopology{
		Enabled:      true,
		SelectedCPUs: []int{4, 5, 6, 7, 8, 9, 10, 11},
	}

	if !topo.Enabled {
		t.Errorf("CPUTopology.Enabled = %v, want true", topo.Enabled)
	}
	if len(topo.SelectedCPUs) != 8 {
		t.Errorf("CPUTopology.SelectedCPUs length = %v, want 8", len(topo.SelectedCPUs))
	}
}

func TestHostCPUTopologyFields(t *testing.T) {
	topo := HostCPUTopology{
		Dies: []CPUDie{
			{
				ID:          0,
				Cores:       4,
				Threads:     2,
				LogicalCPUs: []int{0, 1, 2, 3, 4, 5, 6, 7},
				L3CacheKB:   32768,
			},
			{
				ID:          1,
				Cores:       4,
				Threads:     2,
				LogicalCPUs: []int{8, 9, 10, 11, 12, 13, 14, 15},
				L3CacheKB:   32768,
			},
		},
		TotalCores:     8,
		TotalCPUs:      16,
		ThreadsPerCore: 2,
	}

	if len(topo.Dies) != 2 {
		t.Errorf("HostCPUTopology.Dies length = %v, want 2", len(topo.Dies))
	}
	if topo.TotalCores != 8 {
		t.Errorf("HostCPUTopology.TotalCores = %v, want 8", topo.TotalCores)
	}
	if topo.TotalCPUs != 16 {
		t.Errorf("HostCPUTopology.TotalCPUs = %v, want 16", topo.TotalCPUs)
	}
	if topo.ThreadsPerCore != 2 {
		t.Errorf("HostCPUTopology.ThreadsPerCore = %v, want 2", topo.ThreadsPerCore)
	}
	if topo.Dies[0].L3CacheKB != 32768 {
		t.Errorf("HostCPUTopology.Dies[0].L3CacheKB = %v, want 32768", topo.Dies[0].L3CacheKB)
	}
	if len(topo.Dies[1].LogicalCPUs) != 8 {
		t.Errorf("HostCPUTopology.Dies[1].LogicalCPUs length = %v, want 8", len(topo.Dies[1].LogicalCPUs))
	}
}

func TestCPUOptionsFields(t *testing.T) {
	opts := CPUOptions{
		HideKVM:                true,
		VendorID:               "AuthenticAMD",
		HVFrequency:            true,
		HVRelaxed:              true,
		HVReset:                true,
		HVRuntime:              true,
		HVSpinlocks:            "0x1000",
		HVStimer:               true,
		HVSyncIC:               true,
		HVTime:                 true,
		HVVapic:                true,
		HVVPIndex:              true,
		HVNoNonarchCoresharing: true,
		HVTLBFlush:             true,
		HVTLBFlushExt:          true,
		HVIPI:                  true,
		HVAVIC:                 true,
		TopoExt:                true,
		L3Cache:                true,
		X2APIC:                 true,
		Migratable:             false,
		InvTSC:                 true,
	}

	if !opts.HideKVM {
		t.Errorf("CPUOptions.HideKVM = %v, want true", opts.HideKVM)
	}
	if opts.VendorID != "AuthenticAMD" {
		t.Errorf("CPUOptions.VendorID = %v, want AuthenticAMD", opts.VendorID)
	}
	if !opts.HVFrequency {
		t.Errorf("CPUOptions.HVFrequency = %v, want true", opts.HVFrequency)
	}
	if !opts.HVRelaxed {
		t.Errorf("CPUOptions.HVRelaxed = %v, want true", opts.HVRelaxed)
	}
	if !opts.HVReset {
		t.Errorf("CPUOptions.HVReset = %v, want true", opts.HVReset)
	}
	if !opts.HVRuntime {
		t.Errorf("CPUOptions.HVRuntime = %v, want true", opts.HVRuntime)
	}
	if opts.HVSpinlocks != "0x1000" {
		t.Errorf("CPUOptions.HVSpinlocks = %v, want 0x1000", opts.HVSpinlocks)
	}
	if !opts.HVStimer {
		t.Errorf("CPUOptions.HVStimer = %v, want true", opts.HVStimer)
	}
	if !opts.HVSyncIC {
		t.Errorf("CPUOptions.HVSyncIC = %v, want true", opts.HVSyncIC)
	}
	if !opts.HVTime {
		t.Errorf("CPUOptions.HVTime = %v, want true", opts.HVTime)
	}
	if !opts.HVVapic {
		t.Errorf("CPUOptions.HVVapic = %v, want true", opts.HVVapic)
	}
	if !opts.HVVPIndex {
		t.Errorf("CPUOptions.HVVPIndex = %v, want true", opts.HVVPIndex)
	}
	if !opts.HVNoNonarchCoresharing {
		t.Errorf("CPUOptions.HVNoNonarchCoresharing = %v, want true", opts.HVNoNonarchCoresharing)
	}
	if !opts.HVTLBFlush {
		t.Errorf("CPUOptions.HVTLBFlush = %v, want true", opts.HVTLBFlush)
	}
	if !opts.HVTLBFlushExt {
		t.Errorf("CPUOptions.HVTLBFlushExt = %v, want true", opts.HVTLBFlushExt)
	}
	if !opts.HVIPI {
		t.Errorf("CPUOptions.HVIPI = %v, want true", opts.HVIPI)
	}
	if !opts.HVAVIC {
		t.Errorf("CPUOptions.HVAVIC = %v, want true", opts.HVAVIC)
	}
	if !opts.TopoExt {
		t.Errorf("CPUOptions.TopoExt = %v, want true", opts.TopoExt)
	}
	if !opts.L3Cache {
		t.Errorf("CPUOptions.L3Cache = %v, want true", opts.L3Cache)
	}
	if !opts.X2APIC {
		t.Errorf("CPUOptions.X2APIC = %v, want true", opts.X2APIC)
	}
	if opts.Migratable {
		t.Errorf("CPUOptions.Migratable = %v, want false", opts.Migratable)
	}
	if !opts.InvTSC {
		t.Errorf("CPUOptions.InvTSC = %v, want true", opts.InvTSC)
	}
}

func TestVMStatusFields(t *testing.T) {
	status := VMStatus{
		VMID:       "1",
		Name:       "test-vm",
		Running:    true,
		PID:        12345,
		StartedAt:  time.Now(),
		CPUThreads: []int{100, 101, 102, 103},
		USBDevices: []string{"1-1", "1-2"},
		PCIDevices: []string{"0000:01:00.0"},
	}

	if status.VMID != "1" {
		t.Errorf("VMStatus.VMID = %v, want 1", status.VMID)
	}
	if status.Name != "test-vm" {
		t.Errorf("VMStatus.Name = %v, want test-vm", status.Name)
	}
	if !status.Running {
		t.Errorf("VMStatus.Running = %v, want true", status.Running)
	}
	if status.PID != 12345 {
		t.Errorf("VMStatus.PID = %v, want 12345", status.PID)
	}
	if len(status.CPUThreads) != 4 {
		t.Errorf("VMStatus.CPUThreads length = %v, want 4", len(status.CPUThreads))
	}
	if len(status.USBDevices) != 2 {
		t.Errorf("VMStatus.USBDevices length = %v, want 2", len(status.USBDevices))
	}
	if len(status.PCIDevices) != 1 {
		t.Errorf("VMStatus.PCIDevices length = %v, want 1", len(status.PCIDevices))
	}
}

func TestStartStopScriptFields(t *testing.T) {
	script := StartStopScript{
		StartScript: "/path/to/start.sh",
		StopScript:  "/path/to/stop.sh",
	}

	if script.StartScript != "/path/to/start.sh" {
		t.Errorf("StartStopScript.StartScript = %v, want /path/to/start.sh", script.StartScript)
	}
	if script.StopScript != "/path/to/stop.sh" {
		t.Errorf("StartStopScript.StopScript = %v, want /path/to/stop.sh", script.StopScript)
	}
}

func TestVMEmptyArrays(t *testing.T) {
	vm := VM{}

	if vm.HardDisks != nil {
		t.Errorf("VM.HardDisks should be nil, got %v", vm.HardDisks)
	}
	if vm.CDROMs != nil {
		t.Errorf("VM.CDROMs should be nil, got %v", vm.CDROMs)
	}
}

func TestPCIPassthroughDeviceFields(t *testing.T) {
	dev := PCIPassthroughDevice{
		Address:   "0000:01:00.0",
		ROMPath:   "/roms/gpu.rom",
		Vendor:    "10de",
		Device:    "1b80",
		Name:      "NVIDIA GeForce GTX 1080",
		ClassCode: "0300",
	}

	if dev.Address != "0000:01:00.0" {
		t.Errorf("PCIPassthroughDevice.Address = %v, want 0000:01:00.0", dev.Address)
	}
	if dev.ROMPath != "/roms/gpu.rom" {
		t.Errorf("PCIPassthroughDevice.ROMPath = %v, want /roms/gpu.rom", dev.ROMPath)
	}
	if dev.Vendor != "10de" {
		t.Errorf("PCIPassthroughDevice.Vendor = %v, want 10de", dev.Vendor)
	}
	if dev.Device != "1b80" {
		t.Errorf("PCIPassthroughDevice.Device = %v, want 1b80", dev.Device)
	}
	if dev.Name != "NVIDIA GeForce GTX 1080" {
		t.Errorf("PCIPassthroughDevice.Name = %v, want NVIDIA GeForce GTX 1080", dev.Name)
	}
	if dev.ClassCode != "0300" {
		t.Errorf("PCIPassthroughDevice.ClassCode = %v, want 0300", dev.ClassCode)
	}
}

func TestPCIPassthroughConfigFields(t *testing.T) {
	cfg := PCIPassthroughConfig{
		Devices: []PCIPassthroughDevice{
			{Address: "0000:01:00.0", Vendor: "10de", Device: "1b80", Name: "GPU"},
			{Address: "0000:03:00.0", Vendor: "10de", Device: "10f0", Name: "Audio"},
		},
	}

	if len(cfg.Devices) != 2 {
		t.Errorf("PCIPassthroughConfig.Devices length = %v, want 2", len(cfg.Devices))
	}
	if cfg.Devices[0].Address != "0000:01:00.0" {
		t.Errorf("PCIPassthroughConfig.Devices[0].Address = %v, want 0000:01:00.0", cfg.Devices[0].Address)
	}
}
