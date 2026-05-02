package vm

import (
	"fmt"
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
	// 1. Read current content
	content, err := os.ReadFile(grubPath)
	if err != nil {
		return fmt.Errorf("read grub.cfg: %w", err)
	}

	// 2. Backup existing file
	backupPath := grubPath + ".bak"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("backup grub.cfg: %w", err)
	}

	// 3. Modify content using regex
	// Pattern: vfio-pci.ids= followed by non-whitespace characters
	re := regexp.MustCompile(`vfio-pci\.ids=[^\s]+`)

	var newContent string
	if vfioIDs == "" {
		// Remove the parameter if empty (clean up trailing space too)
		reWithSpace := regexp.MustCompile(`\s*vfio-pci\.ids=[^\s]+`)
		newContent = reWithSpace.ReplaceAllString(string(content), "")
	} else {
		// Replace existing or add new
		replaced := re.ReplaceAllString(string(content), fmt.Sprintf("vfio-pci.ids=%s", vfioIDs))

		// If no replacement happened, we need to add the parameter to the linux line
		if replaced == string(content) {
			// Add to the end of the linux line
			// Use (?m) for multiline mode so ^ matches start of each line
			linuxLineRe := regexp.MustCompile(`(?m)^[\t ]*linux[^\n]*`)
			replaced = linuxLineRe.ReplaceAllString(string(content), fmt.Sprintf("$0 vfio-pci.ids=%s", vfioIDs))
		}
		newContent = replaced
	}

	// 4. Write back (requires rw mount on /media/usb)
	return os.WriteFile(grubPath, []byte(newContent), 0644)
}
