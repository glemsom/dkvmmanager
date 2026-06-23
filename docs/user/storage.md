# Storage

Create LVM logical volumes and assign disks to virtual machines. The TUI provides an integrated LV creation form and a three-step disk selection flow (file, block device, or LVM volume).

## Prerequisites

- [Setup completed](setup.md) — LVM tools installed (`vgs`, `lvcreate`, `lvs`)
- Volume group(s) available on the host
- VM created (see [VM Management](vm-management.md))

## Concepts

- **Logical Volume (LV)**: A block storage device backed by an LVM Volume Group (VG). LVs serve as disk backends for VMs.
- **Volume Group (VG)**: A pool of physical storage from one or more Physical Volumes (PVs). Free space in a VG determines how large an LV can be.
- **Disk paths**: In the VM form, each hard disk entry is a path — a regular file (`.img`, `.qcow2`), a block device (`/dev/sda`), or an LVM LV (`/dev/<vg>/<lv>`).
- **Disk source types**: Three options when adding a disk — disk image file, block device, or LVM logical volume. Each opens a dedicated lister or file browser.

---

## Create Logical Volume

### Access

Configuration tab → **Create Logical Volume** (index 10).

> **Source**: `internal/tui/models/init.go` → `registerAllViews()` (ViewLVCreate); `internal/tui/models/lv_create.go` → `NewLVCreateModel()`.

### Form fields

The LV creation form opens with these fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| **Volume Group** | Dropdown | First VG (or empty if none) | LVM Volume Group to allocate from |
| **Volume Name** | Text input | `my-data-volume` | Name for the logical volume; alphanumeric, dash, underscore |
| **Size** | Text input | `100` | Numeric size for the volume |
| **Unit** | Cycle | `GiB` | Size unit — cycles between GiB, TiB, MiB |
| **Thin pool** | Toggle | OFF | Create a thin-provisioned pool (`--type thin`) |
| **Stripped** | Toggle | Auto-ON if VG has >1 PV | Stripe across multiple PVs (`--stripes N`); only visible when selected VG has >1 PV |
| **Contiguous** | Toggle | OFF | Allocate contiguous extents (`--contiguous y`) |
| **Read-only** | Toggle | OFF | Create read-only volume (`-p r`) |
| **Create** | Button | — | Validate and execute `lvcreate` |
| **Cancel** | Button | — | Return to Configuration tab |

> **Source**: `internal/tui/models/lv_create_form.go` → `LVCreateFormModel`, `BuildPositions()`.

### Volume Group selection

When the Volume Group field is focused:

- `Enter` opens a dropdown listing all discovered VGs with free space
- `↑/↓` / `←/→` navigate the dropdown
- `Enter` confirms the selection and closes the dropdown
- Selecting a VG with more than one PV auto-enables Stripped

**VG discovery** runs `vgs --noheadings -o vg_name,vg_size,vg_free,lv_count,pv_count --units g --separator <TAB>`. If the separator-based output is empty (some environments ignore `--separator`), a fallback command without separator runs and whitespace-parses the output.

> **Source**: `internal/tui/models/lv_create_form.go` → `loadVolumeGroupsCmd()`, `parseVGSOutput()`, `splitVGSLine()`.

### Form keybindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Navigate between fields |
| `↑/↓` / `←/→` | Navigate VG dropdown (when open) or cycle Unit |
| `Enter` on VG field | Open/close VG dropdown |
| `Enter` on toggle | Toggle value (ON/OFF) |
| `Enter` on Create | Validate and run `lvcreate` |
| `Enter` on any field | Same as pressing Create (submit) |
| `Space` on toggle | Toggle value |
| `Space` on Create/Cancel | Same as Enter |
| `Backspace` | Delete character in text fields (Volume Name, Size) |

> **Source**: `internal/tui/models/lv_create_form.go` → `HandleEnter()`, `HandleChar()`, `HandleBackspace()`, `HandleSpace()`.

### Validation & command building

When **Create** is selected:

1. **Volume Group**: Required — must be selected
2. **Volume Name**: Required — must match `[a-zA-Z0-9_-]+`, max 128 characters
3. **Size**: Required — must be a positive number; checked against VG free space (if free > 0 and size exceeds it, an error is shown)

The `lvcreate` command is built from the form:

```
lvcreate -L <size><unit> -n <name> <vg>
          [--type thin] [--stripes N] [--contiguous y] [-p r]
```

- `G` for GiB, `T` for TiB, `M` for MiB
- Stripe count equals the number of PVs in the selected VG (minimum 2)

> **Source**: `internal/tui/models/lv_create_form.go` → `validate()`, `buildCommand()`.

### Dry-run vs. actual execution

The form has two modes controlled by the global `dryRunMode` flag:

| Mode | Behavior |
|------|----------|
| **Dry-run** (`dryRunMode=true`) | Does not execute `lvcreate`. Shows a preview of the command that *would* be run. Form closes immediately with an `LVCreateUpdatedMsg`. |
| **Live** (`dryRunMode=false`) | Executes `lvcreate` via `exec.Command`. On success, sends `LVCreateUpdatedMsg`. On failure, displays the error inline on the Size field. |

> **Source**: `internal/tui/models/lv_create_form.go` → `createCmd()`; `internal/tui/models/types.go` → `dryRunMode`.

### Post-creation

On success, the TUI returns to the Configuration tab with a status message. The newly created LV is immediately available at `/dev/<vg>/<name>` and can be selected when adding a disk to a VM.

---

## Disk Selection in VM Form

When adding a hard disk in the VM create/edit form, the disk selector provides a three-step flow to choose the disk source type and path.

### Access

VM form → press `Enter` on a **Hard Disk** slot → disk selector opens.

> **Source**: `internal/tui/models/vm_form.go` → `HandleEnter()`; `internal/tui/models/disk_selector.go` → `AddDiskModel`.

### Step 0 — Source type selection

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
| **Disk image file** | Browse for `.img`, `.raw`, `.qcow2`, `.qcow`, `.vmdk`, `.vdi`, `.vhdx` files |
| **Block device** | Select from host block devices (e.g., `/dev/sda`, `/dev/nvme0n1`) |
| **LVM Logical Volume** | Select from LVM logical volumes |

**Keybindings**: `↑/↓` / `j/k` navigate options, `Enter` / `Space` select, `ESC` cancel.

> **Source**: `internal/tui/models/disk_selector.go` → `renderSourceSelect()`, `handleEnter()`.

### Step 1 — Disk image file browser

Selecting "Disk image file" opens a file browser filtered for disk image files:

- Filters: `.img`, `.raw`, `.qcow2`, `.qcow`, `.vmdk`, `.vdi`, `.vhdx`
- Only disk image files and directories are shown
- Hidden files (starting with `.`) are excluded
- Directories listed first, then files alphabetically
- Starts in the user's home directory (`$HOME`)

**Keybindings**: `↑/↓` / `j/k` navigate, `Enter` enter directory or select file, `Backspace` go to parent, `ESC` cancel.

> **Source**: `internal/tui/models/file_browser.go` → `NewFileBrowserModel(FileTypeDiskImage)`, `listDirectory()`, `isDiskImage()`.

### Step 2 — Block device lister

Selecting "Block device" lists available host block devices:

```
Select Block Device

Available block devices:

> sda  256G  disk
  sdb  1TB   disk  [RO]

↑/↓ Navigate  Space/Enter Select  ESC Cancel
```

- Discovers devices via `lsblk -o NAME,SIZE,TYPE,RO -n -p`
- Falls back to reading `/sys/block` if `lsblk` is unavailable
- Only `disk` and `part` type devices are shown
- Read-only devices are marked `[RO]`

**Keybindings**: `↑/↓` / `j/k` navigate, `Enter` / `Space` select, `ESC` cancel.

> **Source**: `internal/tui/models/disk_selector_scanner.go` → `listBlockDevices()`, `parseLSBlkOutput()`, `listSysBlock()`.

### Step 3 — LVM volume lister

Selecting "LVM Logical Volume" lists available LVM logical volumes:

```
Select LVM Volume

Available LVM Logical Volumes:

> vg0/my-data-volume  100g  linear  ✓
  vg0/my-thin-pool    200g  pool    ✓
  vg1/snapshot-vol     50g  snapshot  ✗

↑/↓ Navigate  Space/Enter Select  ESC Cancel
```

- Discovers volumes via `lvs --noheadings -o lv_name,vg_name,lv_size,lv_attr --units g --separator <TAB>`
- Sorted by VG name, then LV name
- Snapshot origins (names starting with `[`) are excluded
- Writable volumes show `✓`; read-only volumes show `✗`
- Volume type derived from LV attributes: `linear`, `pool`, `thin`, `snapshot`

**Keybindings**: `↑/↓` / `j/k` navigate, `Enter` / `Space` select, `ESC` cancel.

> **Source**: `internal/tui/models/lvm_volume.go` → `LVMVolumeModel`, `listLVMVolumes()`, `parseLVSOutput()`.

### Selection completed

Once a disk is selected (from any source type), the path is inserted into the VM form's hard disk list. The disk selector closes automatically.

> **Source**: `internal/tui/models/disk_selector_handlers.go` → `handleFileSelected()`; message flow via `FileSelectedMsg` → `DiskAddedMsg`.

---

> **Behind the scenes**: See [Architecture](../dev/architecture.md) for model hierarchy, message flow, and form framework details.

---

## See Also

- [VM Management](vm-management.md) — creating and editing VMs, adding disks
- [Setup](setup.md) — LVM prerequisites and host configuration
- [Running VMs](running-vms.md) — VM lifecycle
- [Hardware Configuration](hardware-config.md) — CPU topology, pinning, passthrough
