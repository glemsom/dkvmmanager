# Understanding LBU and Alpine Diskless Mode

This document explains *why* DKVM Manager uses Alpine Linux in diskless mode,
what `lbu commit` does, and how changes flow through the system.

## What is Alpine diskless mode?

Alpine Linux can run in three modes:

| Mode | Root filesystem | Persistence | Use case |
|------|----------------|-------------|----------|
| **Diskless** | RAM (tmpfs) | Nothing persists unless explicitly saved | Embedded, kiosks, virtualization hosts |
| **Data** | RAM + persistent `/var` | Configs saved, system resets on reboot | Routers, appliances |
| **Disk** (sys) | Physical disk | Full persistent install | Workstations, servers |

DKVM Manager uses **diskless mode**. The entire OS loads into RAM at boot from
a SquashFS image on the USB stick. The root filesystem is empty on every boot.

## Why diskless for a virtualization host?

Three reasons:

### 1. Deterministic state

A virtualization host should always start from a known state. If a kernel
update, a package install, or a configuration tweak causes instability, a
reboot restores a clean system. There is no configuration drift, no orphaned
packages, no half-applied updates.

### 2. No disk I/O for the OS

The root filesystem is in RAM. All disk spindles (or flash) can be dedicated to
VM storage — either as LVM physical volumes or as direct-attached storage. The
OS never competes with VMs for I/O.

### 3. Boot from USB, survive USB failure

The Alpine image lives on a USB stick. Once booted, the OS runs entirely in
RAM. You can **remove the USB stick** and the host keeps running. If the USB
stick fails, insert a replacement and `lbu commit` to save the current
configuration.

## How changes flow

The TUI can make changes that need to survive reboot:

- VM configurations (written to `/media/dkvmdata/dkvmmanager/config.yaml`)
- Network bridge setup
- GRUB kernel parameter changes (PCI passthrough IDs, CPU isolation)
- SSH password changes
- Custom scripts

Some of these (VM configs, scripts) live on the persistent data volume at
`/media/dkvmdata` — they survive automatically. Others (networking, SSH
passwords) modify files in `/etc` on the root filesystem — which is RAM-backed
and **volatile**.

### What `lbu commit` does

`lbu commit` (Local Backup Utility) saves the current state of tracked files
from RAM back to the USB stick:

```
RAM (/etc, /root, /var)  →  lbu commit  →  USB stick (dkvm.apkovl.tar.gz)
                                 ↓
                        Reboot loads overlay
```

Alpine's `lbu` maintains a list of files to back up (in
`/etc/local.d/` and `/etc/apk/protected_paths.d/`). `lbu commit` creates a
`.apkovl.tar.gz` archive on the USB stick. On next boot, Alpine extracts this
overlay on top of the SquashFS image, restoring saved configuration.

In the TUI, **Save changes** in the Configuration tab (or Power tab) runs
`lbu commit` for you. The status bar shows success or failure.

### What does NOT need `lbu commit`

- VM configurations (they are on `/media/dkvmdata`, which is persistent)
- LVM volumes and their contents
- ISO images downloaded to `/media/dkvmdata/isos/`
- Log files in `/media/dkvmdata/vms/<id>/`

### What DOES need `lbu commit`

- The network bridge configuration (`/etc/network/interfaces`)
- SSH password changes (`/etc/shadow`)
- GRUB kernel parameter changes (via `grub.cfg` on the USB stick — though this
  file lives on the USB partition, not in RAM)
- Custom start/stop scripts saved to `/root/` or similar paths
- Any package installations done outside the TUI

## The `/media/dkvmdata` data volume

The data volume is the permanent home for VM state. It is expected to be a
separate mount point (DKVM warns if it is not). This volume holds:

```
/media/dkvmdata/
├── dkvmmanager/
│   └── config.yaml           # All VM configurations
├── vms/
│   └── <vm-id>/
│       ├── qemu.log          # Persisted QEMU log
│       ├── qmp.sock          # QMP socket (live only)
│       ├── ovmf_code.fd      # Per-VM OVMF firmware copy
│       ├── ovmf_vars.fd      # Per-VM OVMF variables
│       ├── start.sh          # Start hook script
│       └── stop.sh           # Stop hook script
└── isos/
    └── *.iso                 # ISO images for VM boot
```

Because this is a separate mount, it can be backed up independently, snapshotted
via LVM, or migrated to a new host by re-mounting at the same path.

## See Also

- [How DKVM Manager Works](how-dkvm-manager-works.md) — architecture overview
- [Power & Save](../user/power-and-save.md) — saving changes in the TUI
- [Setup: Data Folder](../user/setup.md#data-folder--mount-point) —
  `/media/dkvmdata` requirements
- [Alpine Wiki: Backup](https://wiki.alpinelinux.org/wiki/Alpine_local_backup) —
  official `lbu` documentation
