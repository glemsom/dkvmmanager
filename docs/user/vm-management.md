# VM Management

Create, edit, and delete virtual machines through the DKVM Manager TUI.

## Prerequisites

- [Setup completed](setup.md) — KVM, hugepages, IOMMU configured
- `/media/dkvmdata` mounted and writable
- At least one VM disk path or LVM volume available

> **You should know**: See [How DKVM Manager Works](../explanation/how-dkvm-manager-works.md) for VM configuration, MAC addresses, and network modes.

## Navigation

### Accessing VM management

1. Press `Tab` to switch to the **Configuration** tab
2. Use `↑/↓` or `j/k` to highlight a menu item
3. Press `Enter` or `Space` to select

The Configuration menu contains:

| Index | Item | Description |
|-------|------|-------------|
| 0 | Add VM | Create a new VM |
| 1 | Edit VM | Modify an existing VM |
| 2 | Delete VM | Permanently remove a VM |
| 3 | Edit CPU Topology | Guest CPU socket/core/thread layout |
| 4 | Edit vCPU Pinning | Pin virtual CPUs to host cores |
| 5 | Edit PCI Passthrough | Assign host PCI devices to VMs |
| 6 | Edit USB Passthrough | Assign host USB devices to VMs |
| 7 | Edit Start/Stop Script | Hook scripts before/after QEMU |
| 8 | Edit CPU Options | CPU model, features, hypervisor flags |
| 9 | Set SSH Password | SSH access credential |
| 10 | Create Logical Volume | New LVM logical volume |
| last | Save changes | LBU commit to persist configuration |

> **Source**: `internal/tui/models/init.go` → `registerAllViews()`, `buildConfigListAdapter()`; `internal/tui/models/types.go` → view constants.

---

## Create a VM

### Opening the form

Configuration tab → **Add VM** (index 0).

> **Source**: `internal/tui/models/init.go` → `registerAllViews()` (ViewVMCreate); `internal/tui/models/vm_create.go` → `NewVMCreateModel()`.

### Form fields

The creation form opens with these fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| **VM Name** | Text input | *(empty)* | Name for the VM; alphanumeric, dash, underscore, space |
| **Hard Disks** | List | 1 empty slot | Disk paths, block devices, or LVM volumes |
| **CD/DVD Drives (ISOs)** | List | *(empty)* | ISO image paths for boot media |
| **MAC Address** | Text input | Auto-generated | MAC address (`xx:xx:xx:xx:xx:xx`); leave empty for auto |
| **VNC** | Toggle | ON | Enable/disable VNC console |
| **Network** | Toggle | NAT | Switch between NAT and Bridge modes |
| **TPM** | Toggle | OFF | Enable/disable software TPM |
| **Save** | Button | — | Validate and persist the VM |

> **Source**: `internal/tui/models/vm_form_model.go` → `NewVMFormModel()`, `rebuildPositions()`.

### Form keybindings

Use `Tab`/`Shift+Tab` to navigate fields, `↑/↓` to scroll when content exceeds the viewport, `Enter`/`Space` to activate toggles, buttons, or list items, `Backspace`/`Delete` for text input, and `ESC` to cancel and return to the Configuration tab.

See [Keybindings](keybindings.md) for the full reference.

### Save & validation

When **Save** is selected:
1. Form validates all fields
2. Empty trailing disk/CDROM slots are stripped
3. If validation fails, focus jumps to the first field with an error and an error message is shown
4. On success, VM is created via `vmManager.CreateVM()`, persisted via `vmManager.SaveVM()`, and a `VMCreatedMsg` is sent
5. The TUI returns to the Configuration tab with a status message

**Validation rules**:
- VM Name: required, must match `[a-zA-Z0-9_\- ]+`
- MAC Address: optional; if provided, must match `xx:xx:xx:xx:xx:xx` format
- TPM: if enabled, `swtpm` binary must exist at configured path

> **Source**: `internal/tui/models/vm_form_validation.go` → `validateAndSaveCmd()`, `saveNewVMCmd()`.

---

## Edit a VM

### Opening the edit form

Configuration tab → **Edit VM** → select VM from list → `Enter`.

If no VMs exist, the status bar shows "No VMs available to edit" and no action is taken.

> **Source**: `internal/tui/models/vm_selection.go` → `showVMSelection()`; `internal/tui/models/key_handlers.go` → `handleConfigMenuSelection()` (case 1).

### VM selection list

The VM selection view lists all existing VMs sorted by ID. Navigate with `↑/↓` / `j/k`, select with `Enter`/`Space`, or press `ESC` to cancel.

> **Source**: `internal/tui/models/vm_selection.go` → `renderVMSelectView()`.

### Edit form

The edit form is identical to the create form but **pre-filled** with the VM's existing values:

- VM Name, Hard Disks, CDROMs, MAC, VNC, Network, TPM are loaded from the stored config
- If no hard disks are configured, one empty slot is added
- Same keybindings and validation as create form

On save, the VM is updated in-place via `vmManager.SaveVM()` and a `VMUpdatedMsg` is sent.

> **Source**: `internal/tui/models/vm_form_model.go` → `NewVMFormModelEdit()`; `internal/tui/models/vm_form_validation.go` → `updateExistingVMCmd()`; `internal/tui/models/vm_edit.go` → `NewVMEditModel()`.

---

## Delete a VM

### Opening delete confirmation

Configuration tab → **Delete VM** → select VM from list → `Enter`.

> **Source**: `internal/tui/models/key_handlers.go` → `handleConfigMenuSelection()` (case 2); `internal/tui/models/vm_selection.go` → `showVMSelectionForDeletion()`.

### Confirmation dialog

A confirmation dialog appears:

```
WARNING: This action cannot be undone!

Are you sure you want to delete VM '<name>' (ID: <id>)?

> No
  Yes

↑/↓ Navigate  Space/Enter Select  ESC Cancel
```

Navigate with `↑/↓` / `j/k`, confirm with `Enter`/`Space`, or press `ESC` to cancel (same as selecting No).

**Behavior**:
- Selecting **No** returns to the Configuration tab (no action)
- Selecting **Yes** calls `vmManager.DeleteVM()`, sends `VMDeletedMsg`, and returns to Configuration tab with a status message
- On error, an error message is displayed inline below the options

> **Source**: `internal/tui/models/vm_delete.go` → `NewVMDeleteModel()`, `View()`, `handleKeyPress()`.

---

## File Browser

Used when selecting ISO images for CD/DVD drives. Activated by pressing `Enter` on a CDROM list item in the create/edit form.

### Navigation

Navigate with `↑/↓` / `j/k`, enter a directory or select a file with `Enter`/`Space`, go to the parent directory with `Backspace`, or press `ESC` to cancel.

### Filtering

- ISO mode: only `.iso` files are shown (plus directories for navigation)
- Hidden files (starting with `.`) are excluded
- Directories listed first, then files alphabetically

### Starting directory

Defaults to the user's home directory (`$HOME`), falls back to `/` if home is unavailable.

> **Source**: `internal/tui/models/file_browser.go` → `NewFileBrowserModel()`, `listDirectory()`, `isISOFile()`.

---

## Disk Selection (AddDiskModel)

Used when adding a hard disk to a VM. Activated by pressing `Enter` on a hard disk list item in the create/edit form.

### Three-step flow

**Step 0 — Source type selection**:
```
Add Hard Disk

Select source type:

> Disk image file
  Block device
  LVM Logical Volume

Space/Enter Select  ESC Cancel
```

| Option | Description |
|--------|-------------|
| Disk image file | Browse for `.img`, `.raw`, `.qcow2`, `.qcow`, `.vmdk`, `.vdi`, `.vhdx` files |
| Block device | Select from host block devices (e.g., `/dev/sda`, `/dev/nvme0n1`) |
| LVM Logical Volume | Select from LVM logical volumes |

**Step 1 — File browser** (disk image file selected): Opens a file browser filtered for disk images (`.img`, `.raw`, `.qcow2`, `.qcow`, `.vmdk`, `.vdi`, `.vhdx`) and block devices.

**Step 2 — Block device lister** (block device selected): Lists available block devices with size, type, and read-only status. Navigate with `↑/↓` / `j/k`, select with `Enter`/`Space`, or cancel with `ESC`.

**Step 3 — LVM volume lister** (LVM selected): Lists LVM logical volumes discovered via `lvs --noheadings`.

> **Source**: `internal/tui/models/disk_selector.go` → `AddDiskModel`, `NewAddDiskModel()`, `renderSourceSelect()`; `internal/tui/models/disk_selector_scanner.go` (block device listing); `internal/tui/models/lvm_volume.go` (LVM volume listing).

### Block device listing

```
Select Block Device

Available block devices:

> sda  256G  disk
  sdb  1TB   disk  [RO]
```

Read-only devices are marked `[RO]`. Navigate with `↑/↓` / `j/k`, select with `Enter`/`Space`, or cancel with `ESC`. The model runs `lsblk`-equivalent logic to discover devices.

> **Source**: `internal/tui/models/disk_selector.go` → `BlockDeviceModel`, `loadDevices()`.

---

> **Behind the scenes**: See [Architecture](../dev/architecture.md) for model hierarchy, message flow, and form framework details.

---

## See Also

- [Hardware Configuration](hardware-config.md) — CPU topology, vCPU pinning, PCI/USB passthrough
- [Storage](storage.md) — LVM logical volume creation and disk management
- [Scripts & SSH](scripts-and-ssh.md) — Start/stop hook scripts and SSH password
- [Running VMs](running-vms.md) — Start, monitor, and stop VMs
- [Keybindings](keybindings.md) — Complete keyboard reference
