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
// The caller is responsible for ensuring the filesystem is writable (remounted rw).
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

	// 3. Modify content using regex
	// Pattern: vfio-pci.ids= followed by non-whitespace characters
	re := regexp.MustCompile(`vfio-pci\.ids=[^\s]+`)

	var newContent string
	if vfioIDs == "" {
		// Remove the parameter if empty (clean up trailing space too)
		reWithSpace := regexp.MustCompile(`\s*vfio-pci\.ids=[^\s]+`)
		newContent = reWithSpace.ReplaceAllString(string(content), "")
		if debugMode {
			log.Printf("[DEBUG] UpdateGrubVFIOIDs: removed vfio-pci.ids parameter (was empty)")
		}
	} else {
		// Replace existing or add new
		replaced := re.ReplaceAllString(string(content), fmt.Sprintf("vfio-pci.ids=%s", vfioIDs))

		// If no replacement happened, we need to add the parameter to the linux line
		if replaced == string(content) {
			// Add to the end of the linux line
			// Use (?m) for multiline mode so ^ matches start of each line
			linuxLineRe := regexp.MustCompile(`(?m)^[\t ]*linux[^\n]*`)
			replaced = linuxLineRe.ReplaceAllString(string(content), fmt.Sprintf("$0 vfio-pci.ids=%s", vfioIDs))
			if debugMode {
				log.Printf("[DEBUG] UpdateGrubVFIOIDs: added new vfio-pci.ids parameter to linux line")
			}
		} else {
			if debugMode {
				log.Printf("[DEBUG] UpdateGrubVFIOIDs: replaced existing vfio-pci.ids parameter")
			}
		}
		newContent = replaced
	}

	// 4. Write back (requires rw mount on /media/usb)
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
