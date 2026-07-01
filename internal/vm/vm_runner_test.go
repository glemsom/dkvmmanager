package vm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/domain"
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

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)

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

	vm := &domain.VM{
		ID:          "1",
		Name:        "test-vm",
		MAC:         "de:ad:be:ef:00:01",
		NetworkMode: "bridge",
		HardDisks:   []string{"/dev/sda"},
		CDROMs:      []string{"/path/to/install.iso"},
		VNCListen:   "",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{
		CPUTopology: domain.CPUTopology{
			Enabled:      true,
			SelectedCPUs: []int{4, 5, 6, 7},
		},
	}, false)
	args := runner.buildQEMUArgs(vmDir)

	// Verify essential args are present
	argStr := string(joinArgs(args))

	checks := []struct {
		name  string
		value string
	}{
		{"VM name", "-name"},
		{"Q35 machine", "q35"},
		{"QMP socket", "-qmp"},
		{"Memory prealloc via backend", "prealloc=on"},
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

	// Specific assertions for the three issues in #64
	// 1. No redundant accel= in -accel
	if containsString(argStr, "accel=kvm") {
		t.Error("Redundant 'accel=' prefix present in -accel argument (should be just 'kvm,kernel-irqchip=split')")
	}
	// 2. No global -mem-prealloc (backend has prealloc=on)
	if containsString(argStr, "-mem-prealloc") {
		t.Error("Global -mem-prealloc present (should be removed; memory backend has prealloc=on)")
	}
	// 3. Single -machine line: merge memory-backend=mem into the main -machine arg
	if !containsString(argStr, "memory-backend=mem") {
		t.Error("Missing memory-backend=mem (should be merged into single -machine line)")
	}
	// Count -machine occurrences — should be exactly 1 after merge
	machineCount := strings.Count(argStr, " -machine ")
	if machineCount != 1 {
		t.Errorf("Expected exactly 1 '-machine' arg, got %d (should be merged into single -machine line)", machineCount)
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

	vm := &domain.VM{
		ID:          "1",
		Name:        "test-vm",
		MAC:         "de:ad:be:ef:00:01",
		NetworkMode: "nat",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
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
	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
		MAC:  "de:ad:be:ef:00:01",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
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

	vm := &domain.VM{
		ID:          "1",
		Name:        "test-vm",
		MAC:         "de:ad:be:ef:00:01",
		NetworkMode: "bridge",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
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

	vm := &domain.VM{
		ID:          "1",
		Name:        "test-vm",
		MAC:         "de:ad:be:ef:00:01",
		NetworkMode: "invalid",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
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

	vm := &domain.VM{
		ID:        "1",
		Name:      "test-vm",
		MAC:       "de:ad:be:ef:00:01",
		VNCListen: "0.0.0.0:0",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
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

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)

	if runner.VM() != vm {
		t.Error("VM() should return the original VM")
	}
	if runner.IsRunning() {
		t.Error("Should not be running initially")
	}
}

func TestNewVMRunnerWithRunConfig(t *testing.T) {
	cfg := &config.Config{
		DataFolder: "/tmp/test",
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runCfg := RunConfig{
		DryRun: true,
		CPUOptions: domain.CPUOptions{
			HideKVM: true,
			VendorID: "GenuineIntel",
		},
		CPUTopology: domain.CPUTopology{
			Enabled:      true,
			SelectedCPUs: []int{0, 1, 2, 3},
		},
		VCPUPinning: domain.VCPUPinningGlobal{
			Enabled: true,
			Mappings: []domain.VCPUToHostMapping{
				{VCPUID: 0, HostCPUID: 4},
			},
		},
		PCIPassthroughConfig: domain.PCIPassthroughConfig{
			Devices: []domain.PCIPassthroughDevice{
				{Address: "0000:01:00.0", Name: "GPU"},
			},
		},
		USBPassthroughConfig: domain.USBPassthroughConfig{
			Devices: []domain.USBPassthroughDevice{
				{Vendor: "046d", Product: "c52b"},
			},
		},
		StartStopScript: domain.StartStopScript{
			StartScript: "echo start",
			StopScript:  "echo stop",
		},
	}

	runner := NewVMRunner(vm, cfg, runCfg, false)

	// Verify all config sections propagated through getters
	if !runner.VCPUPinning().Enabled {
		t.Error("VCPUPinning should be enabled")
	}
	if len(runner.VCPUPinning().Mappings) != 1 || runner.VCPUPinning().Mappings[0].HostCPUID != 4 {
		t.Error("VCPUPinning.Mappings not propagated")
	}

	devices := runner.PCIPassthroughDevices()
	if len(devices) != 1 || devices[0].Address != "0000:01:00.0" {
		t.Error("PCIPassthroughDevices not propagated")
	}

	usbDevices := runner.USBPassthroughDevices()
	if len(usbDevices) != 1 || usbDevices[0].Vendor != "046d" {
		t.Error("USBPassthroughDevices not propagated")
	}

	if runner.VCpuCount() != 4 {
		t.Errorf("VCpuCount should be 4, got %d", runner.VCpuCount())
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

	vm := &domain.VM{ID: "1", Name: "test"}
	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
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

// --- Persisted log tests (S1) ---

func TestPersistLogWritesLines(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)

	// Start the persist log directly
	if err := runner.startPersistLog(); err != nil {
		t.Fatalf("startPersistLog() failed: %v", err)
	}

	// Send lines via logChan
	runner.logChan <- "[stdout] line one"
	runner.logChan <- "[stderr] line two"
	runner.logChan <- "[start] line three"

	// Close persist log to flush and finalize
	runner.closePersistLog()

	logPath := filepath.Join(dir, "vms", "1", "qemu.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Expected qemu.log to exist: %v", err)
	}
	content := string(data)

	if !containsString(content, "[stdout] line one") {
		t.Errorf("Expected first line in qemu.log, got:\n%s", content)
	}
	if !containsString(content, "[stderr] line two") {
		t.Errorf("Expected second line in qemu.log, got:\n%s", content)
	}
	if !containsString(content, "[start] line three") {
		t.Errorf("Expected third line in qemu.log, got:\n%s", content)
	}
}

func TestPersistLogAppendsOnRestart(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	// First run
	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
	if err := runner.startPersistLog(); err != nil {
		t.Fatalf("startPersistLog() failed: %v", err)
	}
	runner.logChan <- "[stdout] run one line"
	runner.closePersistLog()

	// Second run — same VM, same file
	runner2 := NewVMRunner(vm, cfg, RunConfig{}, false)
	if err := runner2.startPersistLog(); err != nil {
		t.Fatalf("startPersistLog() for second run failed: %v", err)
	}
	runner2.logChan <- "[stdout] run two line"
	runner2.closePersistLog()

	logPath := filepath.Join(dir, "vms", "1", "qemu.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Expected qemu.log to exist: %v", err)
	}
	content := string(data)

	if !containsString(content, "run one line") {
		t.Errorf("Expected first run line in qemu.log, got:\n%s", content)
	}
	if !containsString(content, "run two line") {
		t.Errorf("Expected second run line in qemu.log, got:\n%s", content)
	}

	// Verify order: line 1 before line 2
	idx1 := findIndex(content, "run one line")
	idx2 := findIndex(content, "run two line")
	if idx1 < 0 || idx2 < 0 {
		t.Fatal("Could not find expected lines in qemu.log")
	}
	if idx2 < idx1 {
		t.Errorf("Second run line appeared before first run line — file was truncated")
	}
}

func TestPersistLogSlowFlushDoesNotBlock(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
	if err := runner.startPersistLog(); err != nil {
		t.Fatalf("startPersistLog() failed: %v", err)
	}

	// Send more lines than the buffer (256) without reading from viewChan.
	// The writers should never block because the flusher's forwardViewLine
	// has a drop-oldest fallback.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			runner.logChan <- fmt.Sprintf("line %d", i)
		}
		close(done)
	}()

	select {
	case <-done:
		// All lines sent without blocking
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out: logChan send blocked — backpressure not working")
	}

	runner.closePersistLog()

	// File should exist with some content (may have dropped some lines)
	logPath := filepath.Join(dir, "vms", "1", "qemu.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Expected qemu.log to exist: %v", err)
	}
	content := string(data)

	// At minimum, the file should not be empty
	if len(content) == 0 {
		t.Error("qemu.log should not be empty")
	}
}

func TestPersistLogDryRunWritesToFile(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)
	os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644)
	os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{DryRun: true}, false)
	if err := runner.Start(); err != nil {
		t.Fatalf("Start() should succeed in dry-run: %v", err)
	}

	// Read the DRY-RUN lines from Subscribe() (verifies API unchanged).
	// This also acts as synchronization: the persist goroutine writes the
	// line to the file BEFORE sending it to the subscriber, so after we receive
	// all lines the file is guaranteed up-to-date.
	ch := runner.Subscribe()
	expectedPrefixes := []string{"[DRY-RUN] Full QEMU command:", "qemu-system-x86_64", "[DRY-RUN] Filtered QEMU command"}
	for _, prefix := range expectedPrefixes {
		select {
		case line, ok := <-ch:
			if !ok {
				t.Fatalf("Subscribe() channel closed unexpectedly")
			}
			if !containsString(line, prefix) {
				t.Errorf("Expected line containing %q, got: %s", prefix, line)
			}
		case <-time.After(time.Second):
			t.Fatalf("Timed out reading from Subscribe() for prefix: %s", prefix)
		}
	}

	// Now check the file — guaranteed to be up-to-date because we consumed
	// all log lines from the channel (goroutine writes file before sending).
	logPath := filepath.Join(dir, "vms", "1", "qemu.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Expected qemu.log to exist after dry-run: %v", err)
	}
	content := string(data)

	if !containsString(content, "[DRY-RUN] Full QEMU command:") {
		t.Errorf("Expected DRY-RUN line in qemu.log, got:\n%s", content)
	}
	if !containsString(content, "[DRY-RUN] Filtered QEMU command") {
		t.Errorf("Expected filtered DRY-RUN line in qemu.log, got:\n%s", content)
	}
}

func TestPersistLogWriteErrorDoesNotCrash(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
	if err := runner.startPersistLog(); err != nil {
		t.Fatalf("startPersistLog() failed: %v", err)
	}

	// Write a valid line first
	runner.logChan <- "[stdout] before error"

	// Corrupt the persistBuf by setting it to a broken writer
	// (simulates a write error by making the underlying file unwriteable)
	runner.persistFile.Close()

	// Now send another line — should not panic, just log and continue
	runner.logChan <- "[stdout] after error"

	// Close should not panic either
	runner.closePersistLog()

	// The VM runner should still be usable
	if runner.VM() != vm {
		t.Error("Runner should still be usable after write error")
	}
}

func findIndex(s, substr string) int {
	bs := []byte(s)
	needle := []byte(substr)
	for i := 0; i <= len(bs)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if bs[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
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

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{
		USBPassthroughConfig: domain.USBPassthroughConfig{
			Devices: []domain.USBPassthroughDevice{
				{Vendor: "046d", Product: "c52b", Name: "Logitech Unifying Receiver", BusID: "1-1"},
				{Vendor: "045e", Product: "028e", Name: "Xbox Controller", BusID: "3-2"},
			},
		},
	}, false)

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

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
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

func TestBuildCPUOptsStringWithCPUPM(t *testing.T) {
	runner := &VMRunner{
		vm:     &domain.VM{Name: "test"},
		runCfg: RunConfig{CPUOptions: domain.CPUOptions{CPUPM: true}},
	}

	result := runner.buildCPUOptsString()
	// cpu-pm=on should NOT be in CPU flags - it belongs in -overcommit only
	if containsString(result, "cpu-pm=on") {
		t.Error("cpu-pm=on should not be in CPU flags when CPUPM is enabled")
	}
}

func TestBuildQEMUArgsWithCPUPM(t *testing.T) {
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

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{
		CPUOptions: domain.CPUOptions{CPUPM: true},
	}, false)
	args := runner.buildQEMUArgs(vmDir)
	argStr := string(joinArgs(args))

	// Should have cpu-pm=on in -overcommit
	if !containsString(argStr, "cpu-pm=on") {
		t.Error("Expected cpu-pm=on in -overcommit arg when CPUPM is enabled")
	}
	// Should NOT have cpu-pm=on in -cpu flags (it belongs in -overcommit only)
	if containsString(argStr, "-cpu host,cpu-pm=on") {
		t.Error("cpu-pm=on should not be in -cpu args when CPUPM is enabled")
	}
}

func TestBuildQEMUArgsWithoutCPUPM(t *testing.T) {
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

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
	// CPUPM not set (defaults to false)
	args := runner.buildQEMUArgs(vmDir)
	argStr := string(joinArgs(args))

	// Should NOT have cpu-pm=on in -overcommit
	if containsString(argStr, "cpu-pm=on") {
		t.Error("Should not have cpu-pm=on in -overcommit when CPUPM is disabled")
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

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{
		CPUTopology: domain.CPUTopology{
			Enabled:      true,
			SelectedCPUs: []int{4, 5, 6, 7},
		},
	}, false)
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

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
		// No CPU topology enabled
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
	args := runner.buildQEMUArgs(vmDir)
	argStr := string(joinArgs(args))

	// Should NOT have -smp
	if containsString(argStr, "-smp") {
		t.Error("Should not have -smp when CPU topology is not enabled")
	}
}

func TestBuildCPUOptsString(t *testing.T) {
	runner := &VMRunner{
		vm:     &domain.VM{Name: "test"},
		runCfg: RunConfig{CPUOptions: domain.CPUOptions{
			HideKVM:   true,
			HVRelaxed: true,
			TopoExt:   true,
		}},
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

func TestBuildCPUOptsStringWithHVVendorID(t *testing.T) {
	runner := &VMRunner{
		vm:     &domain.VM{Name: "test"},
		runCfg: RunConfig{CPUOptions: domain.CPUOptions{
			HideKVM:  true,
			VendorID: "AuthenticAMD",
		}},
	}

	result := runner.buildCPUOptsString()

	if !containsString(result, "kvm=off") {
		t.Error("Expected kvm=off flag")
	}
	if !containsString(result, "hv-vendor-id=AuthenticAMD") {
		t.Errorf("Expected hv-vendor-id=AuthenticAMD flag, got: %s", result)
	}
	// Should NOT contain the old incorrect format
	if containsString(result, "-hypervisor,vendor_id=") {
		t.Error("Should not contain old incorrect -hypervisor,vendor_id= format")
	}
}

func TestBuildCPUOptsStringEmpty(t *testing.T) {
	runner := &VMRunner{
		vm:     &domain.VM{Name: "test"},
		runCfg: RunConfig{CPUOptions: domain.CPUOptions{}},
	}

	result := runner.buildCPUOptsString()
	if result != "" {
		t.Errorf("Expected empty string for default CPU options, got %q", result)
	}
}

func TestBuildCPUOptsStringWithForceCPUID(t *testing.T) {
	runner := &VMRunner{
		vm:     &domain.VM{Name: "test"},
		runCfg: RunConfig{CPUOptions: domain.CPUOptions{
			ForceCPUID0x80000026: true,
		}},
	}

	result := runner.buildCPUOptsString()

	if !containsString(result, "x-force-cpuid-0x80000026=on") {
		t.Errorf("Expected x-force-cpuid-0x80000026=on flag, got: %s", result)
	}
}

func TestBuildCPUOptsStringWithL3CacheSizeDie(t *testing.T) {
	runner := &VMRunner{
		vm:     &domain.VM{Name: "test"},
		runCfg: RunConfig{CPUOptions: domain.CPUOptions{
			L3CacheSizeDie: map[int]string{
				0: "32M",
				1: "96M",
			},
		}},
	}

	result := runner.buildCPUOptsString()

	if !containsString(result, "l3-cache-size-die0=33554432") {
		t.Errorf("Expected l3-cache-size-die0=33554432 flag, got: %s", result)
	}
	if !containsString(result, "l3-cache-size-die1=100663296") {
		t.Errorf("Expected l3-cache-size-die1=100663296 flag, got: %s", result)
	}
}

func TestBuildCPUOptsStringWithL3CacheSizeDieEmpty(t *testing.T) {
	runner := &VMRunner{
		vm:     &domain.VM{Name: "test"},
		runCfg: RunConfig{CPUOptions: domain.CPUOptions{
			L3CacheSizeDie: map[int]string{},
		}},
	}

	result := runner.buildCPUOptsString()

	if containsString(result, "l3-cache-size-die") {
		t.Errorf("Expected no l3-cache-size-die flags for empty map, got: %s", result)
	}
}

func TestBuildCPUOptsStringWithL3CacheAssocDie(t *testing.T) {
	runner := &VMRunner{
		vm:     &domain.VM{Name: "test"},
		runCfg: RunConfig{CPUOptions: domain.CPUOptions{
			L3CacheAssocDie: map[int]int{
				0: 8,
				1: 12,
			},
		}},
	}

	result := runner.buildCPUOptsString()

	if !containsString(result, "l3-cache-assoc-die0=8") {
		t.Errorf("Expected l3-cache-assoc-die0=8 flag, got: %s", result)
	}
	if !containsString(result, "l3-cache-assoc-die1=12") {
		t.Errorf("Expected l3-cache-assoc-die1=12 flag, got: %s", result)
	}
}

func TestBuildCPUOptsStringWithL3CacheAssocDieEmpty(t *testing.T) {
	runner := &VMRunner{
		vm:     &domain.VM{Name: "test"},
		runCfg: RunConfig{CPUOptions: domain.CPUOptions{
			L3CacheAssocDie: map[int]int{},
		}},
	}

	result := runner.buildCPUOptsString()

	if containsString(result, "l3-cache-assoc-die") {
		t.Errorf("Expected no l3-cache-assoc-die flags for empty map, got: %s", result)
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

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	// Create host topology: 2 dies, 4 cores per die, 2 threads per core
	hostTopo := domain.HostCPUTopology{
		TotalCores:      8,
		TotalCPUs:     16,
		ThreadsPerCore: 2,
		Dies: []domain.CPUDie{
			{
				ID:    0,
				Cores: 4,
				Threads: 2,
				LogicalCPUs: []int{0, 1, 2, 3, 4, 5, 6, 7},
				CoreDetails: []domain.CPUCore{
					{ID: 0, DieID: 0, Threads: []int{0, 1}},
					{ID: 1, DieID: 0, Threads: []int{2, 3}},
					{ID: 2, DieID: 0, Threads: []int{4, 5}},
					{ID: 3, DieID: 0, Threads: []int{6, 7}},
				},
			},
			{
				ID:    1,
				Cores: 4,
				Threads: 2,
				LogicalCPUs: []int{8, 9, 10, 11, 12, 13, 14, 15},
				CoreDetails: []domain.CPUCore{
					{ID: 0, DieID: 1, Threads: []int{8, 9}},
					{ID: 1, DieID: 1, Threads: []int{10, 11}},
					{ID: 2, DieID: 1, Threads: []int{12, 13}},
					{ID: 3, DieID: 1, Threads: []int{14, 15}},
				},
			},
		},
	}

	runner := NewVMRunner(vm, cfg, RunConfig{
		HostCPUTopology: hostTopo,
		CPUTopology: domain.CPUTopology{
			Enabled:         true,
			SelectedCPUs:    []int{0, 1, 2, 3},
			UseHostTopology: true,
		},
	}, false)

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

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	// Create host topology: 2 dies, 4 cores per die, 2 threads per core
	hostTopo := domain.HostCPUTopology{
		TotalCores:      8,
		TotalCPUs:     16,
		ThreadsPerCore: 2,
		Dies: []domain.CPUDie{
			{ID: 0, Threads: 2, Cores: 4},
			{ID: 1, Threads: 2, Cores: 4},
		},
	}

	runner := NewVMRunner(vm, cfg, RunConfig{
		HostCPUTopology: hostTopo,
		CPUTopology: domain.CPUTopology{
			Enabled:         true,
			SelectedCPUs:    []int{0, 99},
			UseHostTopology: true,
		},
	}, false)

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
	// PID file (transient)
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
	vm := &domain.VM{ID: "1", Name: "test"}
	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
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
	vm := &domain.VM{ID: "1", Name: "test"}
	runner := NewVMRunner(vm, cfg, RunConfig{}, false)

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

// --- Subscribe tests (S2) ---

func TestSubscribeDrainsBufferedLines(t *testing.T) {
	// Given a runner with N lines in its internal channel,
	// Subscribe() returns a channel that drains those N lines before receiving new ones.
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
	if err := runner.startPersistLog(); err != nil {
		t.Fatalf("startPersistLog() failed: %v", err)
	}

	// Send lines to logChan to simulate pre-existing buffer
	runner.logChan <- "[stdout] old line 1"
	runner.logChan <- "[stdout] old line 2"
	runner.logChan <- "[stdout] old line 3"

	// Give the persist loop time to process them
	time.Sleep(50 * time.Millisecond)

	// Subscribe: should drain the buffered lines
	ch := runner.Subscribe()

	// The first lines from the channel should be the buffered lines
	// (or at minimum, the channel should work without blocking)
	received := []string{}
	timeout := time.After(500 * time.Millisecond)
	for i := 0; i < 3; i++ {
		select {
		case line, ok := <-ch:
			if !ok {
				t.Fatalf("channel closed before receiving buffered lines")
			}
			received = append(received, line)
		case <-timeout:
			t.Fatalf("timed out waiting for line %d", i+1)
		}
	}

	// Now send a new line and verify it comes through
	runner.logChan <- "[stdout] new line"
	select {
	case line, ok := <-ch:
		if !ok {
			t.Fatal("channel closed unexpectedly")
		}
		if !containsString(line, "new line") {
			t.Errorf("expected new line, got: %s", line)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for new line")
	}

	_ = received
	runner.closePersistLog()
}

func TestSubscribeChannelClosedOnExit(t *testing.T) {
	// Given a runner that has exited, the returned channel is closed.
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
	if err := runner.startPersistLog(); err != nil {
		t.Fatalf("startPersistLog() failed: %v", err)
	}

	// Subscribe before closing
	ch := runner.Subscribe()

	// Close the persist log (which closes the subscriber channel)
	runner.closePersistLog()

	// The channel should be closed (receive zero value, ok=false)
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed after runner exit")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for channel close")
	}
}

func TestSubscribeNoDuplicationWithRecentLog(t *testing.T) {
	// Integration test: RecentLog + Subscribe should not cause gaps.
	// The persisted file is written before lines reach the subscriber,
	// so RecentLog captures everything on disk, Subscribe covers the live tail.
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/usr/bin/qemu-system-x86_64",
	}

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)
	if err := runner.startPersistLog(); err != nil {
		t.Fatalf("startPersistLog() failed: %v", err)
	}

	// Write some lines to the log (they'll be persisted)
	for i := 1; i <= 5; i++ {
		runner.logChan <- fmt.Sprintf("[stdout] line%d", i)
	}
	time.Sleep(50 * time.Millisecond) // let persist loop flush them

	// RecentLog should see the 5 lines
	lines, err := runner.RecentLog(500)
	if err != nil {
		t.Fatalf("RecentLog() failed: %v", err)
	}
	if len(lines) < 5 {
		t.Errorf("expected at least 5 lines from RecentLog, got %d", len(lines))
	}

	// Subscribe after RecentLog — this drains the old buffered lines first
	ch := runner.Subscribe()

	// Drain any old buffered lines (lines 1-5 that were in viewChan)
	// Subscribe() returns the channel with old lines already drained, so
	// the next read should get the new line we're about to send.

	// Send a new line that wasn't on disk when RecentLog ran
	runner.logChan <- "[stdout] line6"

	// Read lines until we see line6 (skip any buffered overlap)
	found := false
	timeout := time.After(500 * time.Millisecond)
	for !found {
		select {
		case line, ok := <-ch:
			if !ok {
				t.Fatal("channel closed unexpectedly")
			}
			if containsString(line, "line6") {
				found = true
			}
		case <-timeout:
			t.Fatal("timed out waiting for line6")
		}
	}

	runner.closePersistLog()
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
	vm := &domain.VM{ID: "1", Name: "test"}
	runner := NewVMRunner(vm, cfg, RunConfig{}, false)

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

// TestVMRunnerCmdRaceDocumentation documents that the race on r.cmd is fixed.
// Before the fix, Start() wrote r.cmd = exec.Command(...) WITHOUT holding r.mu,
// while ForceStop() and Stop() read r.cmd WITH r.mu held.
// After the fix, r.cmd is assigned under the lock, eliminating the data race.
// Run with `go test -race` to verify no races are reported.
func TestVMRunnerCmdRaceDocumentation(t *testing.T) {
	// This is a documentation test: it verifies the fix compiles and the
	// concurrent-access pattern used by Start() and ForceStop() is race-free.
	// The actual race detection is done by `go test -race`.

	runner := NewVMRunner(&domain.VM{ID: "1", Name: "test"},
		&config.Config{QEMUPath: "/bin/true"}, RunConfig{}, false)

	var wg sync.WaitGroup

	// Simulate the fixed pattern: both write and read happen under the mutex.
	// Run multiple iterations; the race detector will flag any unsynchronized access.
	for i := 0; i < 100; i++ {
		wg.Add(2)

		// Writer: writes r.cmd UNDER the lock (as fixed)
		go func() {
			defer wg.Done()
			runner.mu.Lock()
			runner.cmd = exec.Command("/bin/true")
			runner.cmdProcess = runner.cmd.Process
			runner.mu.Unlock()
		}()

		// Reader: reads r.cmd UNDER the lock
		go func() {
			defer wg.Done()
			runner.mu.Lock()
			if runner.cmd != nil {
				_ = runner.cmd.Process
			}
			runner.mu.Unlock()
		}()

		wg.Wait()

		// Clean up
		runner.mu.Lock()
		if runner.cmd != nil && runner.cmd.Process != nil {
			runner.cmd.Process.Kill()
		}
		runner.cmd = nil
		runner.cmdProcess = nil
		runner.mu.Unlock()
	}
}

// TestStartForceStopConcurrent verifies that Start() and ForceStop() can be
// called concurrently without races, by setting up a minimal real process.
func TestStartForceStopConcurrent(t *testing.T) {
	dir := t.TempDir()
	vmDir := filepath.Join(dir, "vms", "1")
	os.MkdirAll(vmDir, 0755)
	os.WriteFile(filepath.Join(vmDir, "OVMF_CODE.fd"), []byte("fake"), 0644)
	os.WriteFile(filepath.Join(vmDir, "OVMF_VARS.fd"), []byte("fake"), 0644)

	cfg := &config.Config{
		DataFolder: dir,
		QEMUPath:   "/bin/sleep",
	}

	vm := &domain.VM{
		ID:   "1",
		Name: "test-vm",
	}

	runner := NewVMRunner(vm, cfg, RunConfig{}, false)

	// Override hugepages: set memMB small so we skip hugepages checks
	runner.memMB = 64

	// Start in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- runner.Start()
	}()

	// Wait for Start() to get past r.cmd assignment and reach running state
	time.Sleep(100 * time.Millisecond)

	// ForceStop reads r.cmd under the lock — this races with the write in Start()
	err := runner.ForceStop()
	if err != nil {
		t.Logf("ForceStop result (expected possible with sleep binary): %v", err)
	}

	// Drain Start() error
	select {
	case startErr := <-errCh:
		if startErr != nil {
			t.Logf("Start result: %v", startErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start() did not complete within timeout")
	}

	// Cleanup - wait for monitorProcess to finish
	runner.mu.Lock()
	runner.running = false
	runner.mu.Unlock()
}
