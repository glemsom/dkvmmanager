package vm

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
)

func TestValidateOVMFFiles(t *testing.T) {
	// Create temp directory with OVMF files
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg)

	// Should fail - no OVMF files
	if err := runner.ValidateOVMFFiles(); err == nil {
		t.Error("Expected error for missing OVMF files")
	}

	// Create OVMF_CODE.fd only
	os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644)
	if err := runner.ValidateOVMFFiles(); err == nil {
		t.Error("Expected error for missing OVMF_VARS.fd")
	}

	// Create OVMF_VARS.fd
	os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644)
	if err := runner.ValidateOVMFFiles(); err != nil {
		t.Errorf("Expected success with both files present, got: %v", err)
	}
}

func TestBuildQEMUArgs(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create required OVMF files
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder:    dir,
		QEMUPath:      "/usr/bin/qemu-system-x86_64",
		NetworkBridge: "br0",
	}

	vm := &models.VM{
		ID:          "1",
		Name:        "test-vm",
		MAC:         "de:ad:be:ef:00:01",
		NetworkMode: "bridge",
		HardDisks:   []string{"/dev/sda"},
		CDROMs:      []string{"/path/to/install.iso"},
		VNCListen:   "",
	}

	runner := NewVMRunner(vm, cfg)
	runner.SetCPUTopology(models.CPUTopology{
		Enabled:      true,
		SelectedCPUs: []int{4, 5, 6, 7},
	})
	args := runner.buildQEMUArgs(vmDir)

	// Verify essential args are present
	argStr := string(joinArgs(args))

	checks := []struct {
		name  string
		value string
	}{
		{"VM name", "-name"},
		{"KVM accel", "accel=kvm"},
		{"Q35 machine", "q35"},
		{"QMP socket", "-qmp"},
		{"Memory prealloc", "-mem-prealloc"},
		{"Bridge networking", "bridge"},
		{"virtio-net", "virtio-net-pci"},
		{"scsi controller", "virtio-scsi-pci"},
		{"hard disk", "/dev/sda"},
		{"CDROM", "/path/to/install.iso"},
		{"OVMF code", "OVMF_CODE.fd"},
		{"OVMF vars", "OVMF_VARS.fd"},
		{"CPU host", "host"},
		{"no graphic", "-nographic"},
		{"hugepages", "hugetlb=on"},
		{"disable s3", "disable_s3=1"},
	}

	for _, check := range checks {
		if !containsString(argStr, check.value) {
			t.Errorf("Missing %s: expected '%s' in args", check.name, check.value)
		}
	}
}

func TestBuildQEMUArgsWithNAT(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:          "1",
		Name:        "test-vm",
		MAC:         "de:ad:be:ef:00:01",
		NetworkMode: "nat",
	}

	runner := NewVMRunner(vm, cfg)
	args := runner.buildQEMUArgs(vmDir)

	argStr := string(joinArgs(args))

	if !containsString(argStr, "user,id=hostnet0") {
		t.Error("Expected NAT networking: user,id=hostnet0")
	}
	if !containsString(argStr, "virtio-net-pci") {
		t.Error("Expected virtio-net-pci device")
	}
	if !containsString(argStr, "de:ad:be:ef:00:01") {
		t.Error("Expected MAC address in args")
	}
	if containsString(argStr, "bridge") {
		t.Error("Should not have bridge networking in NAT mode")
	}
}

func TestBuildQEMUArgsWithNATDefault(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder:    dir,
		QEMUPath:      "/usr/bin/qemu-system-x86_64",
		NetworkBridge: "br0",
	}

	// VM with empty NetworkMode should default to NAT
	vm := &models.VM{
		ID:   "1",
		Name: "test-vm",
		MAC:  "de:ad:be:ef:00:01",
	}

	runner := NewVMRunner(vm, cfg)
	args := runner.buildQEMUArgs(vmDir)

	argStr := string(joinArgs(args))

	if !containsString(argStr, "user,id=hostnet0") {
		t.Error("Expected default NAT networking when NetworkMode is empty")
	}
	if containsString(argStr, "bridge,") {
		t.Error("Should default to NAT, not bridge, when NetworkMode is empty")
	}
}

func TestBuildQEMUArgsBridgeFallbackToNAT(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
		// NetworkBridge intentionally empty
	}

	vm := &models.VM{
		ID:          "1",
		Name:        "test-vm",
		MAC:         "de:ad:be:ef:00:01",
		NetworkMode: "bridge",
	}

	runner := NewVMRunner(vm, cfg)
	args := runner.buildQEMUArgs(vmDir)

	argStr := string(joinArgs(args))

	if !containsString(argStr, "user,id=hostnet0") {
		t.Error("Expected NAT fallback when bridge is selected but NetworkBridge is empty")
	}
	if containsString(argStr, "bridge,") {
		t.Error("Should not have bridge networking when NetworkBridge config is empty")
	}
}

func TestBuildQEMUArgsUnknownModeDefaultsToNAT(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:          "1",
		Name:        "test-vm",
		MAC:         "de:ad:be:ef:00:01",
		NetworkMode: "invalid",
	}

	runner := NewVMRunner(vm, cfg)
	args := runner.buildQEMUArgs(vmDir)

	argStr := string(joinArgs(args))

	if !containsString(argStr, "user,id=hostnet0") {
		t.Error("Expected NAT fallback for unknown NetworkMode")
	}
}

func TestBuildQEMUArgsWithVNC(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)
	os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644)
	os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:        "1",
		Name:      "test-vm",
		MAC:       "de:ad:be:ef:00:01",
		VNCListen: "0.0.0.0:0",
	}

	runner := NewVMRunner(vm, cfg)
	args := runner.buildQEMUArgs(vmDir)

	argStr := string(joinArgs(args))

	if !containsString(argStr, "-vga std") {
		t.Error("Expected VGA std when VNC is enabled")
	}
	if containsString(argStr, "-nographic") {
		t.Error("Should not have -nographic when VNC is enabled")
	}
}

func TestNewVMRunner(t *testing.T) {
	cfg := &config.Config{
		DataFolder: "/tmp/test",
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg)

	if runner.VM() != vm {
		t.Error("VM() should return the original VM")
	}
	if runner.IsRunning() {
		t.Error("Should not be running initially")
	}
}

func TestCleanupStaleSocket(t *testing.T) {
	dir := t.TempDir()
	socketPath := filepath.Join(dir, "dkvm-1.sock")

	// Create a stale socket file
	os.WriteFile(socketPath, []byte{}, 0600)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{ID: "1", Name: "test"}
	runner := NewVMRunner(vm, cfg)
	runner.socketPath = socketPath
	runner.Cleanup()

	if _, err := os.Stat(socketPath); err == nil {
		t.Error("Stale socket should have been removed")
	}
}

func TestFilterPassthroughArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no passthrough args",
			input:    []string{"-name", "test", "-device", "virtio-net-pci,netdev=hostnet0", "-cpu", "host"},
			expected: []string{"-name", "test", "-device", "virtio-net-pci,netdev=hostnet0", "-cpu", "host"},
		},
		{
			name:     "remove vfio-pci device",
			input:    []string{"-name", "test", "-device", "vfio-pci,host=0000:01:00.0,romfile=/path/to/rom", "-cpu", "host"},
			expected: []string{"-name", "test", "-cpu", "host"},
		},
		{
			name:     "remove usb-host device",
			input:    []string{"-name", "test", "-device", "usb-host,vendorid=0x1234,productid=0x5678", "-cpu", "host"},
			expected: []string{"-name", "test", "-cpu", "host"},
		},
		{
			name:     "remove drive with romfile",
			input:    []string{"-name", "test", "-drive", "if=pflash,format=raw,romfile=/path/to/rom,readonly=on", "-cpu", "host"},
			expected: []string{"-name", "test", "-cpu", "host"},
		},
		{
			name: "mixed passthrough and normal args",
			input: []string{
				"-name", "test",
				"-device", "virtio-net-pci,netdev=hostnet0",
				"-device", "vfio-pci,host=0000:01:00.0",
				"-device", "usb-host,vendorid=0x1234",
				"-drive", "if=pflash,format=raw,readonly=on,file=/path/to/code",
				"-drive", "if=pflash,format=raw,romfile=/path/to/rom",
				"-cpu", "host",
			},
			expected: []string{
				"-name", "test",
				"-device", "virtio-net-pci,netdev=hostnet0",
				"-drive", "if=pflash,format=raw,readonly=on,file=/path/to/code",
				"-cpu", "host",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterPassthroughArgs(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d: %v", len(tt.expected), len(result), result)
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Arg[%d]: expected %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}

// Helper functions

func joinArgs(args []string) []byte {
	result := ""
	for i, a := range args {
		if i > 0 {
			result += " "
		}
		result += a
	}
	return []byte(result)
}

func TestBuildQEMUArgsWithUSBPassthrough(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create required OVMF files
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg)
	runner.SetUSBPassthroughConfig(models.USBPassthroughConfig{
		Devices: []models.USBPassthroughDevice{
			{Vendor: "046d", Product: "c52b", Name: "Logitech Unifying Receiver", BusID: "1-1"},
			{Vendor: "045e", Product: "028e", Name: "Xbox Controller", BusID: "3-2"},
		},
	})

	args := runner.buildQEMUArgs(vmDir)
	argStr := string(joinArgs(args))

	// Verify USB passthrough args are present
	if !containsString(argStr, "usb-host,vendorid=0x046d,productid=0xc52b") {
		t.Error("Expected USB device 046d:c52b in args")
	}
	if !containsString(argStr, "usb-host,vendorid=0x045e,productid=0x028e") {
		t.Error("Expected USB device 045e:028e in args")
	}
}

func TestBuildQEMUArgsWithoutUSBPassthrough(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg)
	// No USB passthrough config set

	args := runner.buildQEMUArgs(vmDir)
	argStr := string(joinArgs(args))

	// Verify no USB passthrough args are present
	if containsString(argStr, "usb-host") {
		t.Error("Expected no usb-host args when no USB config is set")
	}
}

func TestDryRunFiltersUSBPassthrough(t *testing.T) {
	args := []string{
		"-name", "test-vm",
		"-device", "usb-host,vendorid=0x046d,productid=0xc52b",
		"-device", "virtio-net-pci,netdev=hostnet0",
		"-device", "usb-host,vendorid=0x045e,productid=0x028e",
		"-cpu", "host",
	}

	filtered := filterPassthroughArgs(args)
	argStr := string(joinArgs(filtered))

	if containsString(argStr, "usb-host") {
		t.Error("Expected usb-host args to be filtered out in dry-run")
	}
	if !containsString(argStr, "virtio-net-pci") {
		t.Error("Expected non-passthrough device to remain")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && searchBytes([]byte(s), []byte(substr))
}

func searchBytes(haystack, needle []byte) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestBuildQEMUArgsWithCPUTopology(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)
	os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644)
	os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg)
	runner.SetCPUTopology(models.CPUTopology{
		Enabled:      true,
		SelectedCPUs: []int{4, 5, 6, 7},
	})
	args := runner.buildQEMUArgs(vmDir)
	argStr := string(joinArgs(args))

	// Should have -smp with correct count
	if !containsString(argStr, "-smp 4,sockets=1,cores=4,threads=1") {
		t.Error("Expected -smp with 4 vCPUs")
	}
}

func TestBuildQEMUArgsWithoutCPUTopology(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)
	os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644)
	os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:   "1",
		Name: "test-vm",
		// No CPU topology enabled
	}

	runner := NewVMRunner(vm, cfg)
	args := runner.buildQEMUArgs(vmDir)
	argStr := string(joinArgs(args))

	// Should NOT have -smp
	if containsString(argStr, "-smp") {
		t.Error("Should not have -smp when CPU topology is not enabled")
	}
}

func TestBuildCPUOptsString(t *testing.T) {
	runner := &VMRunner{
		vm: &models.VM{Name: "test"},
		cpuOptions: models.CPUOptions{
			HideKVM:   true,
			HVRelaxed: true,
			TopoExt:   true,
		},
	}

	result := runner.buildCPUOptsString()

	if !containsString(result, "kvm=off") {
		t.Error("Expected kvm=off flag")
	}
	if !containsString(result, "+hv-relaxed") {
		t.Error("Expected +hv-relaxed flag")
	}
	if !containsString(result, "+topoext") {
		t.Error("Expected +topoext flag")
	}
}

func TestBuildCPUOptsStringEmpty(t *testing.T) {
	runner := &VMRunner{
		vm:         &models.VM{Name: "test"},
		cpuOptions: models.CPUOptions{},
	}

	result := runner.buildCPUOptsString()
	if result != "" {
		t.Errorf("Expected empty string for default CPU options, got %q", result)
	}
}

func TestSetCPUOptions(t *testing.T) {
	vm := &models.VM{ID: "1", Name: "test"}
	cfg := &config.Config{DataFolder: "/tmp", QEMUPath: "/usr/bin/qemu-system-x86_64"}
	runner := NewVMRunner(vm, cfg)

	opts := models.CPUOptions{
		HideKVM:  true,
		VendorID: "TestVendor",
	}
	runner.SetCPUOptions(opts)

	if !runner.cpuOptions.HideKVM {
		t.Error("SetCPUOptions failed to set HideKVM")
	}
	if runner.cpuOptions.VendorID != "TestVendor" {
		t.Error("SetCPUOptions failed to set VendorID")
	}
}

func TestBuildQEMUArgsWithHostTopology(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)
	os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644)
	os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:   "1",
		Name: "test-vm",
	}

	// Create host topology: 2 dies, 4 cores per die, 2 threads per core
	hostTopo := models.HostCPUTopology{
		TotalCores:      8,
		TotalCPUs:     16,
		ThreadsPerCore: 2,
		Dies: []models.CPUDie{
			{ID: 0, Threads: 2, Cores: 4},
			{ID: 1, Threads: 2, Cores: 4},
		},
	}

	runner := NewVMRunner(vm, cfg)
	runner.SetHostCPUTopology(hostTopo)
	runner.SetCPUTopology(models.CPUTopology{
		Enabled:         true,
		SelectedCPUs:    []int{0, 1, 2, 3},
		UseHostTopology: true,
	})

	args := runner.buildQEMUArgs(vmDir)
	argStr := string(joinArgs(args))

	// Should use -smp with maxcpus and topology params
	if !containsString(argStr, "maxcpus=16") {
		t.Error("Expected maxcpus=16 in -smp args")
	}
	if !containsString(argStr, "dies=2") {
		t.Error("Expected dies=2 in -smp args")
	}
	if !containsString(argStr, "cores=4") {
		t.Error("Expected cores=4 in -smp args")
	}
	if !containsString(argStr, "threads=2") {
		t.Error("Expected threads=2 in -smp args")
	}

	// Should have host-x86_64-cpu devices for each selected CPU
	if !containsString(argStr, "host-x86_64-cpu") {
		t.Error("Expected host-x86_64-cpu device in args")
	}
}

func TestBuildQEMUArgsWithHostTopologyInvalidCPU(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)
	os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644)
	os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644)


	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &models.VM{
		ID:   "1",
		Name: "test-vm",
	}

	// Create host topology: 2 dies, 4 cores per die, 2 threads per core
	hostTopo := models.HostCPUTopology{
		TotalCores:      8,
		TotalCPUs:     16,
		ThreadsPerCore: 2,
		Dies: []models.CPUDie{
			{ID: 0, Threads: 2, Cores: 4},
			{ID: 1, Threads: 2, Cores: 4},
		},
	}

	runner := NewVMRunner(vm, cfg)
	runner.SetHostCPUTopology(hostTopo)
	// Include invalid CPU (99) in selected CPUs
	runner.SetCPUTopology(models.CPUTopology{
		Enabled:         true,
		SelectedCPUs:    []int{0, 99},
		UseHostTopology: true,
	})

	// Should not panic - just skip invalid CPU and continue
	args := runner.buildQEMUArgs(vmDir)
	if len(args) == 0 {
		t.Error("Expected args to be generated even with invalid CPU")
	}

	argStr := string(joinArgs(args))
	// Should still have -smp args
	if !containsString(argStr, "-smp") {
		t.Error("Expected -smp args even when some CPUs are invalid")
	}
}

// --- TPM Persistence Tests ---

func TestCleanupPreservesTPMState(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}

	tpmDir := filepath.Join(vmDir, "tpm")
	if err := os.MkdirAll(tpmDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create fake TPM state files
	stateFile := filepath.Join(tpmDir, "tpm2-00.permall")
	if err := os.WriteFile(stateFile, []byte("fake-tpm-state"), 0600); err != nil {
		t.Fatal(err)
	}
	ctrlSock := filepath.Join(tpmDir, "ctrl")
	if err := os.WriteFile(ctrlSock, []byte{}, 0600); err != nil {
		t.Fatal(err)
	}
	pidPath := filepath.Join(tpmDir, "swtpm.pid")
	if err := os.WriteFile(pidPath, []byte("12345"), 0600); err != nil {
		t.Fatal(err)
	}

	// Start a fake background process to simulate swtpm
	fakeProc, err := os.StartProcess("/bin/sleep", []string{"sleep", "60"}, &os.ProcAttr{})
	if err != nil {
		t.Skipf("Cannot start fake process: %v (test skipped)", err)
	}
	defer fakeProc.Kill()

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}
	vm := &models.VM{ID: "1", Name: "test"}
	runner := NewVMRunner(vm, cfg)
	// Inject fake swtpm process
	runner.swtpmProcess = fakeProc

	// Create a transient socket file (like swtpm would)
	tpmSock := filepath.Join(vmDir, "tpm.sock")
	os.WriteFile(tpmSock, []byte{}, 0600)

	// Run cleanup
	runner.cleanupTPM()

	// Wait a moment for cleanup to complete
	time.Sleep(100 * time.Millisecond)

	// Transient files should be removed
	if _, err := os.Stat(tpmSock); err == nil {
		t.Error("tpm.sock should have been removed")
	}
	if _, err := os.Stat(ctrlSock); err == nil {
		t.Error("ctrl socket should have been removed")
	}
	if _, err := os.Stat(pidPath); err == nil {
		t.Error("swtpm.pid should have been removed")
	}

	// TPM state directory should STILL exist
	if _, err := os.Stat(tpmDir); err != nil {
		t.Errorf("TPM state dir should still exist: %v", err)
	}

	// State files should still exist
	if _, err := os.Stat(stateFile); err != nil {
		t.Errorf("TPM state file should still exist: %v", err)
	}
}

func TestStartTPMErrorDoesNotDeleteState(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}

	tpmDir := filepath.Join(vmDir, "tpm")
	if err := os.MkdirAll(tpmDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create fake TPM state files
	stateFile := filepath.Join(tpmDir, "tpm2-00.permall")
	if err := os.WriteFile(stateFile, []byte("fake-tpm-state"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
		TPMBinary:  "/nonexistent/swtpm-binary", // will fail to start
	}
	vm := &models.VM{ID: "1", Name: "test"}
	runner := NewVMRunner(vm, cfg)

	// startTPM should fail
	err := runner.startTPM(vmDir)
	if err == nil {
		t.Fatal("Expected startTPM to fail with invalid binary")
	}

	// TPM state dir must still exist
	if _, err := os.Stat(tpmDir); err != nil {
		t.Errorf("TPM state dir should still exist after start error: %v", err)
	}

	// State files must still exist
	if _, err := os.Stat(stateFile); err != nil {
		t.Errorf("TPM state file should still exist after start error: %v", err)
	}
}

func TestStartTPMOrphanKill(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		t.Fatal(err)
	}

	tpmDir := filepath.Join(vmDir, "tpm")
	if err := os.MkdirAll(tpmDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create a fake "orphan" process
	fakeProc, err := os.StartProcess("/bin/sleep", []string{"sleep", "60"}, &os.ProcAttr{})
	if err != nil {
		t.Skipf("Cannot start fake process: %v (test skipped)", err)
	}
	defer fakeProc.Kill()

	// Write PID file pointing to the fake orphan
	pidPath := filepath.Join(tpmDir, "swtpm.pid")
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(fakeProc.Pid)), 0600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
		TPMBinary:  "/nonexistent/swtpm-binary", // will fail to start
	}
	vm := &models.VM{ID: "1", Name: "test"}
	runner := NewVMRunner(vm, cfg)

	// startTPM should fail (bad binary) but should have killed the orphan first
	err = runner.startTPM(vmDir)
	if err == nil {
		t.Fatal("Expected startTPM to fail with invalid binary")
	}

	// The orphan process should have been killed
	// Give it a moment to actually die
	time.Sleep(50 * time.Millisecond)
	if err := fakeProc.Signal(syscall.Signal(0)); err == nil {
		t.Error("Orphan process should have been killed by orphan detection")
	}

	// PID file should be removed (orphan detection removes it)
	if _, err := os.Stat(pidPath); err == nil {
		t.Error("Stale PID file should have been removed after orphan kill")
	}
}
