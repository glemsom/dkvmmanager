# Implementation Plan: vfio-pci.ids Kernel Parameter Support

## Overview
Add support for setting the `vfio-pci.ids` kernel parameter by modifying `/media/usb/boot/grub/grub.cfg`. This parameter specifies which PCI devices should be bound to the vfio-pci driver at kernel boot time.

## Progress

### Phase 1 - Core Implementation ✅ COMPLETE (Verified 2026-05-02)

All files created/modified as planned. Verification confirmed all 11 deliverables are present and correctly implemented.

#### Files Created
- [x] **`internal/vm/grub_config.go`** — `BuildVFIOIDs()` and `UpdateGrubVFIOIDs()` with backup, regex replace/add/remove
- [x] **`internal/vm/grub_config_test.go`** — Unit tests for BuildVFIOIDs (empty, single, multi-device) and UpdateGrubVFIOIDs (replace, add, remove, error, structure preservation)

#### Files Modified
- [x] **`internal/config/config.go`** — Added `GrubConfigPath` field (default: `/media/usb/boot/grub/grub.cfg`), viper defaults, save
- [x] **`internal/vm/manager.go`** — Added `ApplyVFIOIDsToKernel()` method that reads PCI config, builds IDs, calls `UpdateGrubVFIOIDs`
- [x] **`internal/tui/models/pci_passthrough.go`** — Added `PCIVFIOKernelAppliedMsg` type
- [x] **`internal/tui/models/pci_passthrough_form.go`** — Added `pciApplyKernel` focus kind, `kernelMsg`/`kernelMsgErr` status fields
- [x] **`internal/tui/models/pci_passthrough_form_handlers.go`** — Renamed `handleEnter` → `handleEnterOrApply`, added `PCIVFIOKernelAppliedMsg` handler
- [x] **`internal/tui/models/pci_passthrough_form_validation.go`** — Added `handleApplyKernel()` method (validates, saves config, applies to grub asynchronously)
- [x] **`internal/tui/models/pci_passthrough_form_navigation.go`** — Added `pciApplyKernel` position, updated `focusedLineIndex`
- [x] **`internal/tui/models/pci_passthrough_form_render.go`** — Added `pciApplyStyle`, render `pciApplyKernel` button, kernel status messages, updated footer
- [x] **`internal/tui/models/message_handlers.go`** — Added handler for `PCIVFIOKernelAppliedMsg` in main model

### Implementation Notes

- `UpdateGrubVFIOIDs` does **not** handle remounting — that's the caller's responsibility (must remount `/media/usb` rw before calling)
- When `vfioIDs` is empty, the parameter is removed from the linux line (with whitespace cleanup)
- The Apply button runs asynchronously via a BubbleTea command, returning `PCIVFIOKernelAppliedMsg`
- Both Save and Apply first persist the PCI config to YAML before modifying grub.cfg
- Navigation: Save → Apply to Kernel (tab/down to move between them)

## Investigation Summary

### Current State
- PCI passthrough devices are configured via `PCIPassthroughDevice` model with `Vendor` and `Device` fields (e.g., `1002:7550`)
- The grub.cfg at `/media/usb/boot/grub/grub.cfg` contains kernel parameters including `vfio-pci.ids=1002:7550,1002:ab40,...`
- PCI passthrough works via a start script that binds devices to vfio-pci at runtime
- The configuration is persisted via the Repository pattern using Viper (YAML config)

### DKVM System Findings (SSH: root@192.168.50.21)
- `/media/usb` is mounted `ro` (read-only) by default
- grub.cfg format uses tabs and space-separated kernel parameters
- `vfio-pci.ids` appears at the end of the linux line before `initrd`
- Backup files exist: `grub.cfg.bak`, `grub.cfg.old`

## Implementation Plan

### Files to Create

1. **`internal/vm/grub_config.go`** - grub.cfg modification logic
   - `UpdateGrubVFIOIDs(vfioIDs string, grubPath string) error` - main entry point
   - Backup existing grub.cfg before modification
   - Remount /media/usb as rw if needed
   - Use regex to find and replace `vfio-pci.ids` parameter

2. **`internal/vm/grub_config_test.go`** - unit tests for grub modification logic

### Files to Modify

1. **`internal/vm/manager.go`**
   - Add `ApplyVFIOIDsToKernel() error` method
   - Add `GrubConfigPath` field access

2. **`internal/tui/models/pci_passthrough_form.go`**
   - Add new focus position type `pciApplyKernel`
   - Add `showApplyKernel` flag

3. **`internal/tui/models/pci_passthrough_form_handlers.go`**
   - Handle "Apply to Kernel" button press
   - Call manager's `ApplyVFIOIDsToKernel()` with async command

4. **`internal/tui/models/pci_passthrough_form_navigation.go`**
   - Add navigation to the new Apply button

5. **`internal/tui/models/pci_passthrough_form_render.go`**
   - Render "Apply to Kernel" button alongside "Save"
   - Show status messages (success/error)

6. **`internal/config/config.go`**
   - Add `GrubConfigPath` field with default `/media/usb/boot/grub/grub.cfg`

### Key Design Decisions

| Decision | Choice | Reasoning |
|----------|--------|-----------|
| **When to update grub.cfg** | Manual button | Safer, user controlled, remount requires root |
| **vfio-pci.ids source** | All PCI devices | Consistent with start script binding |
| **Mount handling** | Shell `mount -o remount,rw` | Standard Linux approach |
| **grub.cfg modification** | Regex replace with backup | Simple, targeted, recoverable |
| **Config path** | Configurable | Flexible for different setups |

### Code Sketch

```go
// grub_config.go
package vm

import (
    "fmt"
    "os"
    "regexp"
    "strings"
    
    "github.com/glemsom/dkvmmanager/internal/models"
)

// BuildVFIOIDs builds the vfio-pci.ids parameter value from PCI devices.
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
// It creates a backup before modification.
func UpdateGrubVFIOIDs(vfioIDs, grubPath string) error {
    // 1. Read current content
    content, err := os.ReadFile(grubPath)
    if err != nil {
        return fmt.Errorf("read grub.cfg: %w", err)
    }
    
    // 2. Backup existing file
    backupPath := grubPath + ".bak"
    if err := os.WriteFile(backupPath, content, 0644); err != nil {
        return fmt.Errorf("backup failed: %w", err)
    }
    
    // 3. Modify content using regex
    // Pattern: vfio-pci.ids= followed by non-whitespace characters
    re := regexp.MustCompile(`vfio-pci\.ids=[^\s]+`)
    var newContent string
    
    if vfioIDs == "" {
        // Remove the parameter if empty
        newContent = re.ReplaceAllString(string(content), "")
    } else {
        // Replace existing or add new
        replaced := re.ReplaceAllString(string(content), fmt.Sprintf("vfio-pci.ids=%s", vfioIDs))
        
        // If no replacement happened, we need to add the parameter to the linux line
        if replaced == string(content) {
            // Add to the end of the linux line (before initrd)
            linuxLineRe := regexp.MustCompile(`(^.*linux[^\n]*?)(\n)`)
            replaced = linuxLineRe.ReplaceAllString(string(content), fmt.Sprintf("$1 vfio-pci.ids=%s$2", vfioIDs))
        }
        newContent = replaced
    }
    
    // 4. Write back (requires rw mount)
    return os.WriteFile(grubPath, []byte(newContent), 0644)
}
```

### UI Flow

1. User configures PCI passthrough devices in the PCI form
2. User clicks "Save" to persist configuration
3. User clicks "Apply to Kernel" to update grub.cfg
4. System remounts /media/usb as rw
5. System updates grub.cfg with new vfio-pci.ids
6. User sees success/error message

### Testing Strategy

1. Unit test `UpdateGrubVFIOIDs` with mock file content
2. Test `BuildVFIOIDs` with various device configurations
3. Integration test: verify grub.cfg format is preserved
4. Manual test on DKVM system

## Future Considerations

### Other Kernel Parameters (Separate Feature)
- `isolcpus=domain,managed_irq=VM CPUs`
- `nohz_full=VM CPUs`
- `rcu_nocbs=VM CPUs`

These should be a separate feature as they relate to CPU isolation, not PCI passthrough.