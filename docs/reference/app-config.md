# Application Configuration Reference (`~/.dkvmmanager.yaml`)

DKVM Manager reads application-level settings from `~/.dkvmmanager.yaml`. If the file does not exist, all defaults shown below are used. Paths support `~` expansion.

> **⚠️ Two configuration files — don't confuse them.** DKVM Manager uses **two separate YAML files** for different purposes. See [VM Config Schema](vm-config.md) for the per-VM configuration file at `/media/dkvmdata/dkvmmanager/config.yaml`.
>
> Most users **never need to edit `~/.dkvmmanager.yaml`** — the defaults work well on standard DKVM hosts.

> **Source:** `internal/config/config.go` — `Config` struct, `DefaultConfig()`, `Load()`.

---

## Field Table

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `data_folder` | `string` | `/media/dkvmdata` | Data folder for VM storage — configs, logs, ISOs, scripts, LVs |
| `vms_config_file` | `string` | `/media/dkvmdata/dkvmmanager/config.yaml` | VM repository YAML path |
| `reserved_mem_mb` | `int` | `4096` | Memory (MB) reserved for the host OS; used for hugepage calculation |
| `bios_code` | `string` | `/usr/share/OVMF/OVMF_CODE.fd` | OVMF UEFI firmware code file — copied per-VM for QEMU |
| `bios_vars` | `string` | `/usr/share/OVMF/OVMF_VARS.fd` | OVMF UEFI firmware variables template — copied per-VM |
| `network_bridge` | `string` | `br0` | Default network bridge name for bridge-mode networking |
| `qemu_path` | `string` | `/usr/bin/qemu-system-x86_64` | QEMU system emulator binary path |
| `tpm_binary` | `string` | `/usr/bin/swtpm` | Software TPM binary path (used when VM TPM is enabled) |
| `log_file` | `string` | `/var/log/dkvm.log` | Application-level log output path |
| `grub_config_path` | `string` | `/media/usb/boot/grub/grub.cfg` | GRUB configuration file — edited by PCI passthrough and vCPU pinning forms |

---

## Field Details

### `data_folder`

Root directory for all VM data. Must be a mount point (DKVM checks this at startup). The following subdirectories are expected:

| Subdirectory | Purpose |
|--------------|---------|
| `dkvmmanager/` | VM configuration YAML |
| `vms/<id>/` | Per-VM data (QEMU log, OVMF files, scripts) |
| `isos/` | ISO images for VM boot |

**Example:**

```yaml
data_folder: /media/dkvmdata
```

> **Source:** `internal/config/config.go` — `Config.DataFolder`; `internal/tui/models/mount_point_warning.go` — mount point check.

---

### `vms_config_file`

Path to the VM repository YAML file. This file is read/written by the TUI when you create or edit VMs. See [VM Config Schema](vm-config.md) for its structure.

**Example:**

```yaml
vms_config_file: /media/dkvmdata/dkvmmanager/config.yaml
```

> **Source:** `internal/vm/repository.go` — `NewRepository()`.

---

### `reserved_mem_mb`

Amount of system memory (in MB) reserved for the host operating system. DKVM subtracts this from total system memory when auto-calculating hugepage requirements.

**Example:**

```yaml
reserved_mem_mb: 4096
```

> **Source:** `internal/hugepages/hugepages.go` — `NewAutoConfig()`.

---

### `bios_code`

Path to the OVMF UEFI firmware code file. DKVM copies this file into each VM's data directory during VM creation. Used as the `-drive if=pflash,format=raw,readonly=on,file=...` argument for QEMU's OVMF_CODE.

**Example:**

```yaml
bios_code: /usr/share/OVMF/OVMF_CODE.fd
```

---

### `bios_vars`

Path to the OVMF UEFI firmware variables template file. DKVM copies this file into each VM's data directory during VM creation. Used as the `-drive if=pflash,format=raw,file=...` argument for QEMU's OVMF_VARS.

**Example:**

```yaml
bios_vars: /usr/share/OVMF/OVMF_VARS.fd
```

---

### `network_bridge`

Default network bridge name used for VM bridge-mode networking. QEMU connects guest NICs to this bridge via the `-netdev bridge` option.

**Example:**

```yaml
network_bridge: br0
```

---

### `qemu_path`

Path to the QEMU system emulator binary. DKVM executes this binary when starting VMs.

**Example:**

```yaml
qemu_path: /usr/bin/qemu-system-x86_64
```

---

### `tpm_binary`

Path to the swtpm binary. Used when a VM has `tpm_enabled: true` — DKVM launches swtpm as a companion process to provide an emulated TPM 2.0 device.

**Example:**

```yaml
tpm_binary: /usr/bin/swtpm
```

---

### `log_file`

Path for application-level log output. DKVM writes startup info, warnings, and errors here. Not to be confused with per-VM QEMU logs stored at `/media/dkvmdata/vms/<id>/qemu.log`.

**Example:**

```yaml
log_file: /var/log/dkvm.log
```

---

### `grub_config_path`

Path to the GRUB configuration file that DKVM edits to apply kernel command-line parameters. Used by the PCI passthrough and vCPU pinning forms to update `linux` lines with `vfio-pci.ids`, `isolcpus`, `nohz_full`, and `rcu_nocbs`.

A `.bak` backup is created before each write.

**Example:**

```yaml
grub_config_path: /media/usb/boot/grub/grub.cfg
```

> **Source:** `internal/vm/grub_config.go` — `BuildVFIOIDs()`, `UpdateGrubVFIOIDs()`, `UpdateGrubCPUParams()`.

---

## Complete Example

```yaml
# ~/.dkvmmanager.yaml
data_folder: /media/dkvmdata
vms_config_file: /media/dkvmdata/dkvmmanager/config.yaml
reserved_mem_mb: 4096
bios_code: /usr/share/OVMF/OVMF_CODE.fd
bios_vars: /usr/share/OVMF/OVMF_VARS.fd
network_bridge: br0
qemu_path: /usr/bin/qemu-system-x86_64
tpm_binary: /usr/bin/swtpm
log_file: /var/log/dkvm.log
grub_config_path: /media/usb/boot/grub/grub.cfg
```

---

## See Also

- [VM Config Schema](vm-config.md) — the per-VM YAML file structure
- [CLI Flags Reference](cli-flags.md) — command-line flags
- [Setup & Prerequisites](../user/setup.md) — system requirements and first launch
- [User Guide Index](../user/README.md) — all documentation
