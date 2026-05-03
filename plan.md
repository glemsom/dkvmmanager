# Implementation Plan: Apply CPU Kernel Parameters from vCPU Pinning Form

## Overview

Add an "Apply to Kernel" button to the vCPU Pinning form that writes three CPU isolation kernel parameters to `/media/usb/boot/grub/grub.cfg`:

1. `isolcpus=domain,managed_irq,<VM CPUs>`
2. `nohz_full=<VM CPUs>`
3. `rcu_nocbs=<VM CPUs>`

This mirrors the existing "Apply to Kernel" button in the PCI passthrough form (which writes `vfio-pci.ids`).

---

## Design Decisions

| # | Decision | Choice |
|---|---|---|
| 1 | Button count | Single button applies all 3 params atomically |
| 2 | CPU list source | Sorted, deduplicated `HostCPUID` values from `VCPUPinningGlobal.Mappings` |
| 3 | Pinning disabled behavior | Remove all 3 params from `grub.cfg` |
| 4 | Coexistence with PCI form | Decoupled — only touches the 3 CPU params, preserves `vfio-pci.ids` |
| 5 | `isolcpus` syntax | `isolcpus=domain,managed_irq,0,1,2,3` (comma-separated qualifiers, no second `=`) |
| 6 | Empty mappings | Same as disabled → remove params |
| 7 | Save first | Save pinning config to disk before applying to kernel (matches PCI form pattern) |
| 8 | Update function signature | `UpdateGrubCPUParams(isolcpus, nohzFull, rcuNoCBS, grubPath string)` — single read/write pass |
| 9 | Button placement | Always visible, below the Save button |
| 10 | Backup | Shared `grub.cfg.bak` file (same as VFIO function) |
| 11 | Feedback messages | Distinguish "applied" vs "removed" with green/red styling |

---

## File Changes

### 1. `internal/vm/grub_config.go`

Add `UpdateGrubCPUParams` function. Pattern: same as `UpdateGrubVFIOIDs` — single pass, line-by-line, regex removal + insertion on `linux` lines only.

**Regex patterns to remove existing params:**
- `isolcpus=[^\s]*` → strip old `isolcpus=` (entire param value after `=`)
- `\s*nohz_full=[^\s]*`
- `\s*rcu_nocbs=[^\s]*`

**Insertion logic (on each `linux` line):**
- If `isolcpus` value is non-empty → append ` isolcpus=domain,managed_irq,<cpulist>`
- If `nohz_full` value is non-empty → append ` nohz_full=<cpulist>`
- If `rcu_nocbs` value is non-empty → append ` rcu_nocbs=<cpulist>`
- If all three are empty → params are removed (already handled by regex strip)

**Function signature:**
```go
func UpdateGrubCPUParams(isolcpus, nohzFull, rcuNoCBS, grubPath string) error
```

### 2. `internal/vm/manager.go`

Add `ApplyCPUParamsToKernel()` method. Pattern: same as `ApplyVFIOIDsToKernel()`.

**Flow:**
1. Get `VCPUPinningGlobal` config
2. Extract `HostCPUID` from all `Mappings` → sorted, deduplicated → build comma-separated CPU string (e.g., `"0,1,2,3"`)
3. Build three param values:
   - `isolcpus = "domain,managed_irq,0,1,2,3"` (or `""` if no mappings)
   - `nohzFull = "0,1,2,3"` (or `""` if no mappings)
   - `rcuNoCBS = "0,1,2,3"` (or `""` if no mappings)
4. Resolve grub path from `m.cfg.GrubConfigPath` (fallback: `/media/usb/boot/grub/grub.cfg`)
5. Remount `/media/usb` as `rw` (using existing `detectMountPath` + `remountFilesystem`)
6. Call `UpdateGrubCPUParams(isolcpus, nohzFull, rcuNoCBS, grubPath)`
7. Defer remount to `ro`
8. Return error or nil

**Method signature:**
```go
func (m *Manager) ApplyCPUParamsToKernel() error
```

### 3. `internal/tui/models/pci_passthrough_form_handlers.go` (reference)

Note the existing message type pattern:
```go
case PCIVFIOKernelAppliedMsg:
    if msg.Success { ... }
```

We'll need a similar message type for CPU params.

### 4. `internal/tui/models/vcpu_pinning_form.go`

**Add new focus position:**
```go
const (
    vcpuPinningToggle vcpuPinningFocusKind = iota
    vcpuPinningSave
    vcpuPinningApplyKernel  // NEW
)
```

**Add new fields to `VCPUPinningFormModel`:**
```go
kernelMsg    string
kernelMsgErr bool
```

**Update navigation** in `handleKey` and `moveFocus` to include the new focus position.

### 5. `internal/tui/models/vcpu_pinning_form_render.go`

**Add "Apply to Kernel" button rendering** below the Save button:
```
  [Space/Enter] Save    [ESC] Cancel
  [Space/Enter] Apply to Kernel    [ESC] Cancel
```

- When focused, use `vcpuPinningSaveStyle` (or similar focus style)
- When unfocused, use `vcpuPinningMutedStyle`
- Always visible regardless of pinning enabled/disabled state

**Add kernel feedback message display** below the button (matching PCI form pattern):
- Green/success style when `!m.kernelMsgErr && m.kernelMsg != ""`
- Red/error style when `m.kernelMsgErr`

**Update `focusedLineIndex()`** to account for the new button line.

### 6. `internal/tui/models/vcpu_pinning_form_validation.go`

Add `handleApplyKernel()` method. Pattern: same as PCI form's `handleApplyKernel()`.

**Flow:**
1. Clear errors and kernel message
2. Build config from current in-memory form state
3. Validate
4. Save config to disk first (`m.vmManager.SaveVCPUPinningGlobal(m.pinning)`)
5. Apply to kernel asynchronously:
   ```go
   return m, func() tea.Msg {
       err := m.vmManager.ApplyCPUParamsToKernel()
       if err != nil {
           return VCPUCPUKernelAppliedMsg{Success: false, Error: err.Error()}
       }
       return VCPUCPUKernelAppliedMsg{Success: true}
   }
   ```

### 7. `internal/tui/models/vcpu_pinning_form_handlers.go` (new file or add to existing)

Handle `VCPUCPUKernelAppliedMsg` in the form's `Update()` method:
```go
case VCPUCPUKernelAppliedMsg:
    if msg.Success {
        if m.pinning.Enabled && len(m.pinning.Mappings) > 0 {
            m.kernelMsg = "Kernel CPU isolation parameters applied to grub.cfg"
        } else {
            m.kernelMsg = "Kernel CPU isolation parameters removed from grub.cfg"
        }
        m.kernelMsgErr = false
    } else {
        m.kernelMsg = msg.Error
        m.kernelMsgErr = true
    }
    m.syncViewport()
    return m, nil
```

### 8. `internal/vm/grub_config_test.go`

Add tests for `UpdateGrubCPUParams`:
- Add all three parameters to a clean grub.cfg
- Update existing parameters (replace old values)
- Remove all three parameters (empty strings)
- Coexistence: ensure `vfio-pci.ids` is preserved when CPU params are added
- Coexistence: ensure CPU params are preserved when `vfio-pci.ids` is added afterward
- Handle corrupted/duplicate params on same line
- Params never appear on non-linux lines (e.g., initrd)

---

## Message Type

Add a new message type (can go in `vcpu_pinning_form.go` or a shared file):

```go
// VCPUCPUKernelAppliedMsg is sent when CPU kernel parameters have been applied to grub.cfg
type VCPUCPUKernelAppliedMsg struct {
    Success bool
    Error   string
}
```

---

## CPU List Building Logic

```go
// buildHostCPUList extracts sorted, deduplicated host CPU IDs from pinning mappings
func buildHostCPUList(mappings []models.VCPUToHostMapping) string {
    cpuSet := make(map[int]bool)
    for _, m := range mappings {
        cpuSet[m.HostCPUID] = true
    }
    cpus := make([]int, 0, len(cpuSet))
    for cpu := range cpuSet {
        cpus = append(cpus, cpu)
    }
    sort.Ints(cpus)
    
    if len(cpus) == 0 {
        return ""
    }
    var parts []string
    for _, c := range cpus {
        parts = append(parts, strconv.Itoa(c))
    }
    return strings.Join(parts, ",")
}
```

Then:
- `isolcpus value` = `"domain,managed_irq," + cpuList` (or `""` if empty)
- `nohz_full value` = `cpuList` (or `""` if empty)
- `rcu_nocbs value` = `cpuList` (or `""` if empty)

---

## Existing Files for Reference

- PCI form "Apply to Kernel": `internal/tui/models/pci_passthrough_form_validation.go` (`handleApplyKernel`)
- PCI form rendering: `internal/tui/models/pci_passthrough_form_render.go` (search for `pciApplyKernel`)
- Existing grub updater: `internal/vm/grub_config.go` (`UpdateGrubVFIOIDs`)
- Manager remount logic: `internal/vm/manager.go` (`ApplyVFIOIDsToKernel`, `detectMountPath`, `remountFilesystem`)
- PCI message type: search for `PCIVFIOKernelAppliedMsg`
- vCPU pinning form: `internal/tui/models/vcpu_pinning_form*.go`
- Grub config path: `/media/usb/boot/grub/grub.cfg` (from `config.go` and `manager.go`)
