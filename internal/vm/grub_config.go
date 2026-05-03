package vm

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// BuildVFIOIDs builds the vfio-pci.ids parameter value from PCI passthrough devices.
// Format: "vendor1:device1,vendor2:device2,..."
func BuildVFIOIDs(devices []models.PCIPassthroughDevice) string {
	if len(devices) == 0 {
		return ""
	}
	var ids []string
	for _, d := range devices {
		ids = append(ids, fmt.Sprintf("%s:%s", d.Vendor, d.Device))
	}
	return strings.Join(ids, ",")
}

// UpdateGrubVFIOIDs updates the vfio-pci.ids parameter in the grub.cfg file.
// It creates a backup before modification and writes the updated content.
// The caller must ensure the filesystem is writable (e.g., remounted rw for /media/usb).
func UpdateGrubVFIOIDs(vfioIDs, grubPath string) error {
	// Debug logging
	if debugMode {
		log.Printf("[DEBUG] UpdateGrubVFIOIDs: grubPath=%s, vfioIDs=%q", grubPath, vfioIDs)
	}

	// 1. Read current content
	content, err := os.ReadFile(grubPath)
	if err != nil {
		if debugMode {
			log.Printf("[DEBUG] UpdateGrubVFIOIDs: failed to read grub.cfg: %v", err)
		}
		return fmt.Errorf("read grub.cfg: %w", err)
	}

	if debugMode {
		log.Printf("[DEBUG] UpdateGrubVFIOIDs: read %d bytes from %s", len(content), grubPath)
	}

	// 2. Backup existing file
	backupPath := grubPath + ".bak"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		if debugMode {
			log.Printf("[DEBUG] UpdateGrubVFIOIDs: failed to create backup: %v", err)
		}
		return fmt.Errorf("backup grub.cfg: %w", err)
	}

	if debugMode {
		log.Printf("[DEBUG] UpdateGrubVFIOIDs: backup created at %s", backupPath)
	}

	// 3. Modify content: process line-by-line to guarantee vfio-pci.ids= appears
	// exactly once per linux line (no duplicates) and never on non-linux lines.
	vfioRe := regexp.MustCompile(`\s*vfio-pci\.ids=[^\s]*`)
	linuxLineRe := regexp.MustCompile(`(?i)^([\t ]*linux[ \t])`)

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		isLinux := linuxLineRe.MatchString(line)
		// Remove ALL existing vfio-pci.ids= occurrences from this line
		cleaned := vfioRe.ReplaceAllString(line, "")
		// If this is a linux line and we have IDs, append exactly one vfio-pci.ids=
		if isLinux && vfioIDs != "" {
			cleaned = cleaned + " vfio-pci.ids=" + vfioIDs
		}
		lines[i] = cleaned
	}
	newContent := strings.Join(lines, "\n")

	if vfioIDs == "" {
		if debugMode {
			log.Printf("[DEBUG] UpdateGrubVFIOIDs: removed all vfio-pci.ids parameters")
		}
	} else {
		if debugMode {
			log.Printf("[DEBUG] UpdateGrubVFIOIDs: set vfio-pci.ids=%s on linux line(s)", vfioIDs)
		}
	}

	// 4. Write back
	if err := os.WriteFile(grubPath, []byte(newContent), 0644); err != nil {
		if debugMode {
			log.Printf("[DEBUG] UpdateGrubVFIOIDs: failed to write updated grub.cfg: %v", err)
		}
		return err
	}

	if debugMode {
		log.Printf("[DEBUG] UpdateGrubVFIOIDs: successfully updated %s", grubPath)
	}

	return nil
}
