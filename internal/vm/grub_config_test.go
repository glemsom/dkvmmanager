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
