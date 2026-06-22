# Keybindings Reference

Complete keyboard reference for DKVM Manager TUI.

## Conventions

- `↑/↓/←/→` — arrow keys
- `j/k` — vim-style navigation (alternatives to `↓/↑`)
- `Enter` / `Space` — both act as "select" in most contexts
- `Ctrl+C` — equivalent to `q` for quit; also force-kills VM in running view
- `ESC` — cancel / return to parent (except running VM view where it's blocked)

> **Source**: `internal/tui/models/key_handlers.go` → `handleKeyPress()`; `internal/tui/models/form/keybinds.go` → `DefaultKeyBindings()`.

---

## Global (Top-Level)

| Key | Context | Action |
|-----|---------|--------|
| `q` / `Ctrl+C` | Any top-level screen (no VM running) | Quit application |
| `q` / `Ctrl+C` | Any top-level screen (VM running) | Warning: "A VM is running. Press 'q' in the VM view to stop it first." |
| `ESC` | Top-level | Quit application |
| `Tab` / `→` | Top-level | Next tab (VMs → Configuration → Power → VMs) |
| `Shift+Tab` / `←` | Top-level | Previous tab |
| `↑/↓` / `j/k` | Any list | Navigate list items |
| `Enter` / `Space` | Any list | Select highlighted item |
| `r` | VMs tab | Refresh VM list from disk |

> **Source**: `internal/tui/models/key_handlers.go` → `handleKeyPress()` (q, ctrl+c, esc, r, tab navigation); `internal/tui/components/tabs.go` → `HandleKeyInput()`.

---

## VMs Tab

The main VM list with dual-pane view (list + detail panel).

| Key | Action |
|-----|--------|
| `↑/↓` / `j/k` | Navigate VM list |
| `Enter` / `Space` | Start selected VM (opens running view) |
| `r` | Refresh VM list |
| `Tab` / `→` | Switch to Configuration tab |
| `Shift+Tab` / `←` | Switch to Power tab |

> **Source**: `internal/tui/models/key_handlers.go` → `handleKeyPress()`, `handleVMSelection()`.

---

## Configuration Tab

The configuration menu listing all VM management and hardware forms.

| Key | Action |
|-----|--------|
| `↑/↓` / `j/k` | Navigate config menu |
| `Enter` / `Space` | Open selected form / action |
| `Tab` / `→` | Switch to Power tab |
| `Shift+Tab` / `←` | Switch to VMs tab |

**Menu items** (ordered):

| Index | Item | View opened |
|-------|------|-------------|
| 0 | Add VM | VM create form |
| 1 | Edit VM | VM selection list → edit form |
| 2 | Delete VM | VM selection list → delete confirmation |
| 3 | Edit CPU Topology | CPU topology form |
| 4 | Edit vCPU Pinning | vCPU pinning form |
| 5 | Edit PCI Passthrough | PCI passthrough form |
| 6 | Edit USB Passthrough | USB passthrough form |
| 7 | Edit Start/Stop Script | Start/stop script form |
| 8 | Edit CPU Options | CPU options form |
| 9 | Set SSH Password | SSH password form |
| 10 | Create Logical Volume | LV create form |
| last | Save changes | LBU commit (no UI — runs in background) |

> **Source**: `internal/tui/models/init.go` → `registerAllViews()`, `buildConfigListAdapter()`; `internal/tui/models/key_handlers.go` → `handleConfigMenuSelection()`.

---

## Power Tab

| Key | Action |
|-----|--------|
| `↑/↓` / `j/k` | Navigate power menu |
| `Enter` / `Space` | Execute selected action |
| `Tab` / `→` | Switch to VMs tab |
| `Shift+Tab` / `←` | Switch to Configuration tab |

| Index | Item | Action |
|-------|------|--------|
| 0 | Power off system | Execute system power off |
| 1 | Reboot system | Execute system reboot |

> **Source**: `internal/tui/models/key_handlers.go` → `handlePowerSelection()`; `internal/tui/models/init.go` → `buildPowerListAdapter()`.

---

## Sub-View Navigation (General)

When inside any sub-view (form, dialog, selection list):

| Key | Action |
|-----|--------|
| `ESC` | Return to parent tab (cancel current operation) |

**Exception**: Running VM view — ESC does NOT leave. Shows reminder instead. See [Running VM View](#running-vm-view).

If a file browser or disk selector is active within a form, ESC closes the browser/selector first. A second ESC then returns from the form.

> **Source**: `internal/tui/models/key_handlers.go` → `update()` ESC handling, `isFileBrowserActiveInSubView()`.

---

## VM Create/Edit Form

The shared form used for both creating and editing VMs. Uses the `ScrollableForm` framework.

### Navigation

| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `↑/↓` / `PgUp/PgDn` | Scroll within form (when content exceeds viewport) |

### Text fields (VM Name, MAC Address)

| Key | Action |
|-----|--------|
| Typing (any printable char) | Insert character at cursor |
| `Enter` | Move to next editable field |
| `Backspace` | Delete character before cursor |
| `Delete` | Delete character at cursor |

### Toggle fields (VNC, Network, TPM)

| Key | Action |
|-----|--------|
| `Enter` / `Space` | Toggle value (ON↔OFF, NAT↔Bridge) |

### List fields (Hard Disks, CD/DVD Drives)

| Key | Action |
|-----|--------|
| `Enter` / `Space` | Open file picker / disk selector for the item |
| `Delete` (item focused) | Remove the item from the list |
| `Enter` on **[+ Add Disk]** | Add new empty disk slot |
| `Enter` on **[+ Add CDROM]** | Add new empty CDROM slot |

### Save button

| Key | Action |
|-----|--------|
| `Enter` / `Space` on **Save** | Validate form and persist VM (create or update) |

### Cancel

| Key | Action |
|-----|--------|
| `ESC` | Cancel form, return to Configuration tab |

> **Source**: `internal/tui/models/vm_form.go` → `HandleEnter()`, `HandleChar()`, `HandleBackspace()`, `HandleDelete()`; `internal/tui/models/vm_form_ui.go` → `RenderPosition()`; `internal/tui/models/form/keybinds.go` → `DefaultKeyBindings()`.

---

## File Browser

Used for selecting ISO images (in VM form) and disk image files (in Add Disk flow).

| Key | Action |
|-----|--------|
| `↑/↓` / `j/k` | Navigate files/directories |
| `Enter` / `Space` | Enter directory or select file |
| `Backspace` | Go to parent directory |
| `ESC` / `Ctrl+C` | Cancel (no selection) |

> **Source**: `internal/tui/models/file_browser.go` → `handleKeyPress()`.

---

## Disk Selector (AddDiskModel)

Multi-step flow for adding a hard disk to a VM.

### Step 0 — Source type selection

| Key | Action |
|-----|--------|
| `↑/↓` / `j/k` | Navigate source types |
| `Enter` / `Space` | Select source type |
| `ESC` | Cancel |

### Step 1 — File browser (disk image)

Same keybindings as [File Browser](#file-browser) but filtered for disk image extensions (`.img`, `.raw`, `.qcow2`, `.qcow`, `.vmdk`, `.vdi`, `.vhdx`) and block devices.

### Step 2 — Block device lister

| Key | Action |
|-----|--------|
| `↑/↓` / `j/k` | Navigate block devices |
| `Enter` / `Space` | Select device |
| `ESC` / `Ctrl+C` | Cancel |

### Step 3 — LVM volume lister

| Key | Action |
|-----|--------|
| `↑/↓` / `j/k` | Navigate LVM volumes |
| `Enter` / `Space` | Select volume |
| `ESC` / `Ctrl+C` | Cancel |

> **Source**: `internal/tui/models/disk_selector_handlers.go` → `AddDiskModel.handleKeyPress()`, `BlockDeviceModel.handleKeyPress()`; `internal/tui/models/lvm_volume.go` → `handleKeyPress()`.

---

## VM Delete Confirmation

| Key | Action |
|-----|--------|
| `↑/↓` / `j/k` | Switch between No and Yes |
| `Enter` / `Space` | Confirm selection |
| `ESC` | Cancel (same as No — returns to Configuration tab) |

> **Source**: `internal/tui/models/vm_delete.go` → `handleKeyPress()`.

---

## VM Selection List

Used when choosing a VM to edit or delete.

| Key | Action |
|-----|--------|
| `↑/↓` / `j/k` | Navigate VM list |
| `Enter` / `Space` | Select VM → opens edit form or delete confirmation |
| `ESC` | Cancel, return to Configuration tab |

> **Source**: `internal/tui/models/vm_selection.go` → `renderVMSelectView()`; `internal/tui/models/key_handlers.go` → `delegateToSubView()` (ViewVMSelect case).

---

## Running VM View

### While VM is running

| Key | Action |
|-----|--------|
| `q` | Stop VM gracefully (QMP `system_powerdown`), return to menu when stopped |
| `Ctrl+C` | Force kill QEMU process (SIGKILL), return to menu |
| `↑/↓` / `j/k` | Scroll log one line |
| `PgUp` / `PgDn` | Scroll log one page |
| `Home` / `End` | Jump to top/bottom of log |
| `ESC` | Show reminder: "VM is still running. Press 'q' to stop it." |

### After VM has stopped

| Key | Action |
|-----|--------|
| `q` | Exit view, return to main menu |
| `ESC` | Show reminder: "Press 'q' to exit the VM view." |
| `↑/↓` / `PgUp` / `PgDn` | Scroll log (available for post-mortem review) |

> **Source**: `internal/tui/models/vm_running.go` → `handleKeyPress()`; viewport delegates to `charm.land/bubbles/v2/viewport`.

---

## Mount Point Warning

Shown at startup if `/media/dkvmdata` is not a mount point.

| Key | Action |
|-----|--------|
| `Enter` / `Space` / `ESC` | Dismiss warning and continue |

> **Source**: `internal/tui/models/mount_point_warning.go` → `handleKeyPress()`.

---

## Form-Specific Keys

All forms share the `ScrollableForm` framework with consistent navigation:

| Key | All forms | Notes |
|-----|-----------|-------|
| `Tab` / `Shift+Tab` | Navigate fields | All forms |
| `↑/↓` / `PgUp/PgDn` | Scroll viewport | All forms |
| `Enter` / `Space` on toggle | Toggle value | Applies to forms with toggle fields |
| `Enter` on button | Activate button | Save, Apply, Add, etc. |
| `ESC` | Cancel / return | All forms |

### Text-entry forms

Forms with text input (VM form, CPU Options, SSH Password, LV Create, Start/Stop Script) additionally support:

| Key | Action |
|-----|--------|
| Typing | Insert character at cursor |
| `Backspace` | Delete before cursor |
| `Delete` | Delete at cursor |
| `Enter` on text field | Move to next field (VM form, LV Create); or no-op |

### Toggle-only forms

Forms without text input (vCPU Pinning, PCI Passthrough, USB Passthrough, CPU Topology):
- `HandleChar`, `HandleBackspace`, `HandleDelete` are no-ops
- Navigation is via `Tab`/`Shift+Tab` and `Enter`/`Space` on toggles

> **Source**: Various `*_form_handlers.go` files; `internal/tui/models/form/form.go` → `update()` dispatch logic.

---

## Source References

| File | Covers |
|------|--------|
| `internal/tui/models/key_handlers.go` | Main key dispatch, global keys, tab switching, menu selection, VM start |
| `internal/tui/models/vm_form.go` | VM create/edit form key handling (Enter, Char, Backspace, Delete) |
| `internal/tui/components/tabs.go` | Tab bar rendering and tab key input (`Tab`/`Shift+Tab`) |
| `internal/tui/models/file_browser.go` | File browser navigation keys |
| `internal/tui/models/vm_delete.go` | Delete confirmation dialog keys |
| `internal/tui/models/vm_running.go` | Running VM view keys (stop, scroll) |
| `internal/tui/models/vm_selection.go` | VM selection list keys |
| `internal/tui/models/disk_selector.go` | Disk selector model and sub-models |
| `internal/tui/models/disk_selector_handlers.go` | AddDiskModel and BlockDeviceModel key handling |
| `internal/tui/models/mount_point_warning.go` | Mount point warning dismiss keys |
| `internal/tui/models/form/keybinds.go` | Form framework key binding definitions |
| `internal/tui/models/form/form.go` | ScrollableForm key dispatch to form models |

---

## See Also

- [Setup & Prerequisites](setup.md)
- [VM Management](vm-management.md)
- [Running VMs](running-vms.md)
- [Power & Save](power-and-save.md)
