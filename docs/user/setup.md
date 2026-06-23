# Setup & Prerequisites

Host setup guide covering all prerequisites before using DKVM Manager.

## Prerequisites

- Linux host with KVM/QEMU installed
- Root or sudo access for kernel configuration
- `/media/dkvmdata` storage volume available

> **You should know**: See [How DKVM Manager Works](../explanation/how-dkvm-manager-works.md) for hugepages, IOMMU, and LVM background.

---

## System Requirements

### KVM / QEMU

Install QEMU with x86_64 system emulation:

```bash
# Alpine
apk add qemu-system-x86_64 qemu-img qemu-modules

# Debian/Ubuntu
apt install qemu-system-x86 qemu-utils
```

Verify hardware virtualization is available (vmx for Intel, svm for AMD):

```bash
grep -E '(vmx|svm)' /proc/cpuinfo | head -1
```

Ensure `/dev/kvm` exists and your user has access (typically via `kvm` group):

```bash
ls -l /dev/kvm
groups $USER | grep -o '\bkvm\b'
```

> **Source**: `internal/config/config.go` — QEMU binary path defaults to `/usr/bin/qemu-system-x86_64`.

### /proc filesystem

DKVM reads host resource info from `/proc`:
- `/proc/meminfo` — total system memory for hugepage auto-config
- `/proc/<pid>/stat` and `/proc/<pid>/status` — QEMU process RSS and CPU time
- `/proc/sys/vm/nr_hugepages` — hugepage reservation

Ensure `/proc` is mounted (standard on Linux).

> **Source**: `internal/hugepages/hugepages.go` → `GetTotalSystemMemoryMB()`, `Check()`, `Ensure()`; `internal/vm/proc.go` → process metric readers.

---

## IOMMU & PCI Passthrough

### Enable IOMMU

Add to kernel command-line in GRUB (`/media/usb/boot/grub/grub.cfg`):

| CPU Vendor | Kernel Parameter   |
|------------|--------------------|
| Intel      | `intel_iommu=on`   |
| AMD        | `amd_iommu=on`     |

Also enable in BIOS/UEFI (VT-d for Intel, AMD-Vi for AMD).

### Verify IOMMU

```bash
dmesg | grep -i iommu
# Should show: DMAR: IOMMU enabled  (Intel)  or  AMD-Vi: IOMMU enabled
```

Check IOMMU groups:

```bash
ls /sys/kernel/iommu_groups/*/devices/
```

Each device in a group shares an IOMMU context. For passthrough, whole group must be passed together.

### vfio-pci Driver

Load the vfio-pci kernel module:

```bash
modprobe vfio-pci
```

For early binding (before host drivers claim devices), add to `/etc/modprobe.d/vfio.conf`:

```
options vfio-pci ids=10de:1b80,10de:10f0  # example: NVIDIA GPU + audio
```

Rebuild initramfs after changes.

DKVM Manager writes `vfio-pci.ids=` to the GRUB `linux` line automatically when you configure PCI passthrough in the TUI. It also creates a `.bak` backup before writing.

> **Source**: `internal/vm/grub_config.go` → `BuildVFIOIDs()`, `UpdateGrubVFIOIDs()`.

---

## Hugepages

QEMU uses 2 MB hugepages as VM memory backing. DKVM Manager checks availability at startup and can attempt allocation.

### Check current config

```bash
cat /proc/sys/vm/nr_hugepages        # number of 2MB pages reserved
grep Huge /proc/meminfo               # full hugepage stats
```

### Configure hugepages

**Temporary** (until reboot):

```bash
echo 4096 > /proc/sys/vm/nr_hugepages   # reserve 4096 × 2MB = 8 GB
```

**Permanent**: add to kernel command-line in GRUB:

```
hugepages=4096
```

### DKVM automatic check

DKVM's `hugepages.Ensure()` reads total system memory from `/proc/meminfo`, subtracts 4 GB reserved for the host OS, aligns down to a 2 MB boundary, and writes the required count to `/proc/sys/vm/nr_hugepages` if insufficient. Requires root.

On insufficient hugepages, DKVM shows a user-friendly error like:

```
insufficient hugepages: have 0, need 4096 (try: echo 4096 > /proc/sys/vm/nr_hugepages)
```

> **Source**: `internal/hugepages/hugepages.go` → `NewAutoConfig()`, `Ensure()`, `FormatError()`.

---

## LVM Setup

### Install LVM tools

```bash
# Alpine
apk add lvm2

# Debian/Ubuntu
apt install lvm2
```

### Typical DKVM data layout

1. Create a physical volume and volume group:

   ```bash
   pvcreate /dev/sdX
   vgcreate dkvm_vg /dev/sdX
   ```

2. Create logical volumes for VM disks via DKVM TUI or CLI:

   ```bash
   lvcreate -L 20G -n myvm_disk dkvm_vg
   ```

3. Volume path for VM config: `/dev/dkvm_vg/myvm_disk`

DKVM's LVM volume lister runs `lvs --noheadings` to discover available volumes and displays them in a selectable list.

> **Source**: `internal/tui/models/lvm_volume.go` → `listLVMVolumes()`, `parseLVSOutput()`.

---

## Data Folder & Mount Point

DKVM Manager expects `/media/dkvmdata` to be a **mount point** (not a plain directory). The DKVM hypervisor auto-mounts filesystems with `LABEL=dkvmdata`.

### What lives there

| Path | Purpose |
|------|---------|
| `/media/dkvmdata/dkvmmanager/config.yaml` | VM configuration YAML — see [VM Config Schema](../reference/vm-config.md) |
| `/media/dkvmdata/vms/<id>/qemu.log` | Per-VM persisted log |
| `/media/dkvmdata/vms/<id>/` | Per-VM data (scripts, configs) |
| `/media/dkvmdata/isos/` | ISO images for VM boot |

### Mount point check

At startup, DKVM compares the device ID of `/media/dkvmdata` with its parent directory. If they match (not a mount point), a warning modal appears:

> *"/media/dkvmdata is not a mount point. The DKVM hypervisor will auto-mount filesystems with LABEL=dkvmdata. To resolve this, create a filesystem with the label 'dkvmdata' and restart."*

Press Enter, Space, or ESC to dismiss and continue anyway (but VM settings won't persist).

Skip the check with `-skip-mount-check` for testing.

> **Source**: `internal/tui/models/mount_point_warning.go` → `isMountPoint()`, `MountPointWarningModel`; `internal/tui/models/init.go` → `NewMainModelWithConfig()`.

---

## GRUB Configuration

DKVM Manager edits `/media/usb/boot/grub/grub.cfg` (configurable via `grub_config_path`) to manage kernel command-line parameters on `linux` lines.

### Parameters managed

| Parameter | Purpose | Set via TUI |
|-----------|---------|-------------|
| `vfio-pci.ids=` | Bind PCI devices to vfio-pci at boot | PCI Passthrough form |
| `isolcpus=` | Isolate host CPUs from scheduler | vCPU Pinning form |
| `nohz_full=` | Disable timer ticks on isolated CPUs | vCPU Pinning form |
| `rcu_nocbs=` | Offload RCU callbacks from isolated CPUs | vCPU Pinning form |

Each edit creates a `.bak` backup before writing.

The filesystem must be writable (remount rw if needed):

```bash
mount -o remount,rw /media/usb
```

> **Source**: `internal/vm/grub_config.go` → `UpdateGrubVFIOIDs()`, `UpdateGrubCPUParams()`; `internal/config/config.go` → `GrubConfigPath`.

---

## First Launch

### CLI flags

See [CLI Flags Reference](../reference/cli-flags.md) for all available flags.

### Minimum terminal

DKVM requires 80×25 terminal. Warns on smaller sizes but allows continuing.

See [Terminal Capabilities](../terminal-capabilities.md) for detailed analysis of terminal compatibility (box-drawing, CP437 support, color profiles).

### Debug mode details

When `-debug` is set:
- `tea.LogToFile()` redirects all `log.*` output to `debug.log` (tries CWD, then HOME, then `/tmp`)
- AltScreen is disabled so debug output is visible without switching buffers
- All log output suppressed from terminal to avoid TUI corruption

> **Source**: `main.go` → flag definitions, `setupDebugLog()`; `internal/tui/tui.go` → `Run()`, `validateAndLogTerminalSize()`.

---

## Application Configuration (`~/.dkvmmanager.yaml`)

DKVM reads application-level settings from `~/.dkvmmanager.yaml`. If the file does not

> **⚠️ Two configuration files — don't confuse them.** DKVM Manager uses **two separate YAML files** for different purposes:
>
> | File | Purpose | Edited by |
> |------|---------|----------|
> | `~/.dkvmmanager.yaml` | **Application settings** (QEMU path, firmware paths, data folder location, GRUB config path, etc.) | User manually (rarely needed) |
> | `/media/dkvmdata/dkvmmanager/config.yaml` | **VM configurations** (all VMs, their hardware, scripts, etc.) | TUI automatically when you create/edit VMs |
>
> Most users **never need to edit `~/.dkvmmanager.yaml`** — the defaults work well on standard DKVM hosts.
> The TUI reads/writes `/media/dkvmdata/dkvmmanager/config.yaml` directly when you manage VMs.
> See [VM Config Schema](../reference/vm-config.md) for the full VM file structure.
exist, all defaults below are used. Paths support `~` expansion.

See the [App Config Schema](../reference/app-config.md) for the complete field reference with types, defaults, descriptions, and examples.
---

## See Also

- [VM Management](vm-management.md) — create your first VM
- [Hardware Configuration](hardware-config.md) — CPU topology, vCPU pinning, PCI/USB passthrough
- [Storage](storage.md) — LVM logical volume creation
- [Keybindings](keybindings.md) — keyboard reference
