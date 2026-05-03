package vm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

func TestBuildVFIOIDs(t *testing.T) {
	tests := []struct {
		name     string
		devices  []models.PCIPassthroughDevice
		expected string
	}{
		{
			name:     "empty devices",
			devices:  []models.PCIPassthroughDevice{},
			expected: "",
		},
		{
			name: "single device",
			devices: []models.PCIPassthroughDevice{
				{Vendor: "1002", Device: "7550"},
			},
			expected: "1002:7550",
		},
		{
			name: "multiple devices",
			devices: []models.PCIPassthroughDevice{
				{Vendor: "1002", Device: "7550"},
				{Vendor: "1002", Device: "ab40"},
				{Vendor: "8086", Device: "15d4"},
			},
			expected: "1002:7550,1002:ab40,8086:15d4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := BuildVFIOIDs(tc.devices)
			if result != tc.expected {
				t.Errorf("BuildVFIOIDs() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestUpdateGrubVFIOIDs_ReplaceExisting(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet vfio-pci.ids=1002:7550,1002:ab40
	initrd /initrd.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	newIDs := "10de:1b80,10de:10f0"
	if err := UpdateGrubVFIOIDs(newIDs, grubPath); err != nil {
		t.Fatalf("UpdateGrubVFIOIDs() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "vfio-pci.ids=10de:1b80,10de:10f0") {
		t.Errorf("Expected new vfio-pci.ids not found. Got:\n%s", string(content))
	}

	// Verify backup was created
	backupPath := grubPath + ".bak"
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Backup file not created: %v", err)
	}
	if string(backup) != original {
		t.Error("Backup file does not match original content")
	}
}

func TestUpdateGrubVFIOIDs_AddNew(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet
	initrd /initrd.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	newIDs := "10de:1b80"
	if err := UpdateGrubVFIOIDs(newIDs, grubPath); err != nil {
		t.Fatalf("UpdateGrubVFIOIDs() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "vfio-pci.ids=10de:1b80") {
		t.Errorf("Expected vfio-pci.ids parameter added. Got:\n%s", string(content))
	}
}

func TestUpdateGrubVFIOIDs_Remove(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet vfio-pci.ids=1002:7550
	initrd /initrd.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateGrubVFIOIDs("", grubPath); err != nil {
		t.Fatalf("UpdateGrubVFIOIDs() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(content), "vfio-pci.ids") {
		t.Errorf("vfio-pci.ids should have been removed. Got:\n%s", string(content))
	}
}

func TestUpdateGrubVFIOIDs_ReadError(t *testing.T) {
	err := UpdateGrubVFIOIDs("1002:7550", "/nonexistent/path/grub.cfg")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestUpdateGrubVFIOIDs_DeduplicatesSameLine(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	// Simulate a corrupted grub.cfg with vfio-pci.ids appearing twice on the same line
	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet vfio-pci.ids=1002:7550 vfio-pci.ids=10de:1b80
	initrd /initrd.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateGrubVFIOIDs("8086:15d4", grubPath); err != nil {
		t.Fatalf("UpdateGrubVFIOIDs() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}

	// Count occurrences of vfio-pci.ids=
	count := strings.Count(string(content), "vfio-pci.ids=")
	if count != 1 {
		t.Errorf("Expected exactly 1 vfio-pci.ids= occurrence, got %d. Got:\n%s", count, string(content))
	}

	if !strings.Contains(string(content), "vfio-pci.ids=8086:15d4") {
		t.Errorf("Expected vfio-pci.ids=8086:15d4. Got:\n%s", string(content))
	}
}

func TestUpdateGrubVFIOIDs_MultipleLinuxLines(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	// Multiple boot entries with vfio-pci.ids on each
	original := `set default=0
set timeout=5

menuentry 'DKVM Normal' {
	linux /vmlinuz root=/dev/sda1 ro quiet vfio-pci.ids=1002:7550
	initrd /initrd.img
}

menuentry 'DKVM Recovery' {
	linux /vmlinuz root=/dev/sda1 ro quiet single vfio-pci.ids=1002:7550
	initrd /initrd-recovery.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateGrubVFIOIDs("10de:1b80", grubPath); err != nil {
		t.Fatalf("UpdateGrubVFIOIDs() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}

	// Each linux line should have exactly one vfio-pci.ids=
	lines := strings.Split(string(content), "\n")
	linuxLineCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "linux") {
			linuxLineCount++
			count := strings.Count(line, "vfio-pci.ids=")
			if count != 1 {
				t.Errorf("Linux line has %d vfio-pci.ids= occurrences (expected 1): %s", count, line)
			}
		}
	}
	if linuxLineCount < 2 {
		t.Errorf("Expected at least 2 linux lines, got %d", linuxLineCount)
	}
}

func TestUpdateGrubVFIOIDs_RemovesFromNonLinuxLine(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	// vfio-pci.ids accidentally on an initrd line
	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet
	initrd /initrd.img vfio-pci.ids=1002:7550
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateGrubVFIOIDs("10de:1b80", grubPath); err != nil {
		t.Fatalf("UpdateGrubVFIOIDs() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}

	// Should have been moved to the linux line, removed from initrd line
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "initrd") && strings.Contains(line, "vfio-pci.ids=") {
			t.Errorf("vfio-pci.ids= should not appear on initrd line: %s", line)
		}
	}

	if !strings.Contains(string(content), "vfio-pci.ids=10de:1b80") {
		t.Errorf("Expected vfio-pci.ids=10de:1b80 on linux line. Got:\n%s", string(content))
	}
}

func TestUpdateGrubVFIOIDs_PreservesStructure(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet vfio-pci.ids=1002:7550
	initrd /initrd.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateGrubVFIOIDs("8086:15d4", grubPath); err != nil {
		t.Fatalf("UpdateGrubVFIOIDs() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}

	// Verify structure is preserved
	for _, expected := range []string{
		"set default=0",
		"set timeout=5",
		"menuentry 'DKVM'",
		"initrd /initrd.img",
		"vfio-pci.ids=8086:15d4",
	} {
		if !strings.Contains(string(content), expected) {
			t.Errorf("Expected %q to be preserved. Got:\n%s", expected, string(content))
		}
	}
}

func TestUpdateGrubCPUParams_AddAllThree(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet
	initrd /initrd.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateGrubCPUParams("domain,managed_irq,0,1,2,3", "0,1,2,3", "0,1,2,3", grubPath); err != nil {
		t.Fatalf("UpdateGrubCPUParams() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}
	result := string(content)

	if !strings.Contains(result, "isolcpus=domain,managed_irq,0,1,2,3") {
		t.Errorf("Expected isolcpus parameter not found. Got:\n%s", result)
	}
	if !strings.Contains(result, "nohz_full=0,1,2,3") {
		t.Errorf("Expected nohz_full parameter not found. Got:\n%s", result)
	}
	if !strings.Contains(result, "rcu_nocbs=0,1,2,3") {
		t.Errorf("Expected rcu_nocbs parameter not found. Got:\n%s", result)
	}
}

func TestUpdateGrubCPUParams_ReplaceExisting(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet isolcpus=domain,managed_irq,0,1 nohz_full=0,1 rcu_nocbs=0,1
	initrd /initrd.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateGrubCPUParams("domain,managed_irq,2,3", "2,3", "2,3", grubPath); err != nil {
		t.Fatalf("UpdateGrubCPUParams() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}
	result := string(content)

	if !strings.Contains(result, "isolcpus=domain,managed_irq,2,3") {
		t.Errorf("Expected new isolcpus not found. Got:\n%s", result)
	}
	if strings.Contains(result, "isolcpus=domain,managed_irq,0,1") {
		t.Errorf("Old isolcpus should have been replaced. Got:\n%s", result)
	}
}

func TestUpdateGrubCPUParams_RemoveAll(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet isolcpus=domain,managed_irq,0,1,2,3 nohz_full=0,1,2,3 rcu_nocbs=0,1,2,3
	initrd /initrd.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateGrubCPUParams("", "", "", grubPath); err != nil {
		t.Fatalf("UpdateGrubCPUParams() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}

	for _, param := range []string{"isolcpus=", "nohz_full=", "rcu_nocbs="} {
		if strings.Contains(string(content), param) {
			t.Errorf("%s should have been removed. Got:\n%s", param, string(content))
		}
	}
}

func TestUpdateGrubCPUParams_CoexistenceWithVFIO(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet vfio-pci.ids=1002:7550
	initrd /initrd.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	// Add CPU params
	if err := UpdateGrubCPUParams("domain,managed_irq,0,1", "0,1", "0,1", grubPath); err != nil {
		t.Fatalf("UpdateGrubCPUParams() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}
	result := string(content)

	// Both vfio-pci.ids and CPU params should be present
	if !strings.Contains(result, "vfio-pci.ids=1002:7550") {
		t.Errorf("vfio-pci.ids should be preserved after adding CPU params. Got:\n%s", result)
	}
	if !strings.Contains(result, "isolcpus=domain,managed_irq,0,1") {
		t.Errorf("isolcpus should be present. Got:\n%s", result)
	}

	// Now add VFIO params - both should coexist
	if err := UpdateGrubVFIOIDs("8086:15d4", grubPath); err != nil {
		t.Fatalf("UpdateGrubVFIOIDs() error: %v", err)
	}

	content, err = os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}
	result = string(content)

	if !strings.Contains(result, "isolcpus=domain,managed_irq,0,1") {
		t.Errorf("isolcpus should be preserved after adding VFIO params. Got:\n%s", result)
	}
	if !strings.Contains(result, "vfio-pci.ids=8086:15d4") {
		t.Errorf("vfio-pci.ids should be updated. Got:\n%s", result)
	}
}

func TestUpdateGrubCPUParams_DuplicateParamsOnSameLine(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	// Simulate corrupted grub.cfg with duplicate params
	original := `set default=0
set timeout=5

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet isolcpus=domain,managed_irq,0,1 isolcpus=domain,managed_irq,2,3 nohz_full=0,1 nohz_full=2,3
	initrd /initrd.img
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateGrubCPUParams("domain,managed_irq,4,5", "4,5", "4,5", grubPath); err != nil {
		t.Fatalf("UpdateGrubCPUParams() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}
	result := string(content)

	// Each param should appear exactly once
	for _, param := range []string{"isolcpus=", "nohz_full=", "rcu_nocbs="} {
		count := strings.Count(result, param)
		if count != 1 {
			t.Errorf("Expected exactly 1 occurrence of %s, got %d. Got:\n%s", param, count, result)
		}
	}
}

func TestUpdateGrubCPUParams_NeverOnNonLinuxLines(t *testing.T) {
	dir := t.TempDir()
	grubPath := filepath.Join(dir, "grub.cfg")

	// Params accidentally on non-linux lines
	original := `set default=0
set timeout=5 isolcpus=domain,managed_irq,0,1

menuentry 'DKVM' {
	linux /vmlinuz root=/dev/sda1 ro quiet
	initrd /initrd.img nohz_full=0,1 rcu_nocbs=0,1
}
`
	if err := os.WriteFile(grubPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := UpdateGrubCPUParams("domain,managed_irq,2,3", "2,3", "2,3", grubPath); err != nil {
		t.Fatalf("UpdateGrubCPUParams() error: %v", err)
	}

	content, err := os.ReadFile(grubPath)
	if err != nil {
		t.Fatal(err)
	}

	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "linux") {
			for _, param := range []string{"isolcpus=", "nohz_full=", "rcu_nocbs="} {
				if strings.Contains(line, param) {
					t.Errorf("%s should not appear on non-linux line: %s", param, line)
				}
			}
		}
	}

	// Params should appear on the linux line
	if !strings.Contains(string(content), "isolcpus=domain,managed_irq,2,3") {
		t.Errorf("Expected isolcpus on linux line. Got:\n%s", string(content))
	}
}

func TestBuildHostCPUList(t *testing.T) {
	tests := []struct {
		name     string
		mappings []models.VCPUToHostMapping
		expected string
	}{
		{
			name:     "empty mappings",
			mappings: []models.VCPUToHostMapping{},
			expected: "",
		},
		{
			name: "single mapping",
			mappings: []models.VCPUToHostMapping{
				{VCPUID: 0, HostCPUID: 4},
			},
			expected: "4",
		},
		{
			name: "multiple mappings sorted",
			mappings: []models.VCPUToHostMapping{
				{VCPUID: 0, HostCPUID: 6},
				{VCPUID: 1, HostCPUID: 4},
				{VCPUID: 2, HostCPUID: 5},
			},
			expected: "4,5,6",
		},
		{
			name: "deduplicates same host CPU",
			mappings: []models.VCPUToHostMapping{
				{VCPUID: 0, HostCPUID: 2},
				{VCPUID: 1, HostCPUID: 2},
				{VCPUID: 2, HostCPUID: 4},
			},
			expected: "2,4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := buildHostCPUList(tc.mappings)
			if result != tc.expected {
				t.Errorf("buildHostCPUList() = %q, want %q", result, tc.expected)
			}
		})
	}
}
