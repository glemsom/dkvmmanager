# Backup & Restore

This guide explains how to protect DKVM Manager data (VM configurations, disk
volumes) and how to recover from a hardware failure or misconfiguration.

It follows the **3-2-1 rule**: keep at least three copies of your data, on two
different media, with one copy off-site.

---

## What to back up

| Data | Location | Persistence | Back up? |
|------|----------|-------------|----------|
| VM configurations | `/media/dkvmdata/dkvmmanager/config.yaml` | Data volume | **Yes** — the single most important file |
| VM disk volumes | LVM logical volumes (e.g. `/dev/vg0/my-vm-root`) | LVM | **Yes** — use LVM snapshots for crash-consistent copies |
| VM firmware (OVMF vars) | `/media/dkvmdata/vms/<id>/ovmf_vars.fd` | Data volume | **Yes** — stores UEFI boot entries, MAC addresses |
| VM start/stop scripts | `/media/dkvmdata/vms/<id>/start.sh`, `stop.sh` | Data volume | **Yes** |
| QEMU logs | `/media/dkvmdata/vms/<id>/qemu.log` | Data volume | Optional — useful for post-mortem |
| ISO images | `/media/dkvmdata/isos/` | Data volume | Optional — can be re-downloaded |
| Alpine overlay (OS config) | USB stick (`dkvm.apkovl.tar.gz`) | USB | **Yes** — network, SSH, GRUB settings |
| OS state | RAM (diskless mode) | None | **No** — resets on every boot |

---

## Back up VM configuration (config.yaml)

The central configuration file is `/media/dkvmdata/dkvmmanager/config.yaml`.
It contains every VM definition: CPU, memory, disks, network, scripts, and
firmware settings.

### Simple file copy (while host is running)

```bash
cp /media/dkvmdata/dkvmmanager/config.yaml \
   /mnt/backup/dkvmmanager/config-$(date +%F).yaml
```

The file is written atomically by DKVM Manager (with `write-file` library), so
a simple `cp` is safe even while the TUI is running.

### Automated backup with cron (Alpine)

```bash
# Install cronie if not present
apk add cronie

# Add a daily backup job
cat > /etc/periodic/daily/dkvm-backup << 'EOF'
#!/bin/sh
BACKUP_DST="/mnt/backup/dkvmmanager"
mkdir -p "$BACKUP_DST"
cp /media/dkvmdata/dkvmmanager/config.yaml \
   "$BACKUP_DST/config-$(date +%F).yaml"
# Keep only the last 30 daily backups
find "$BACKUP_DST" -name 'config-*.yaml' -mtime +30 -delete
EOF

chmod +x /etc/periodic/daily/dkvm-backup
rc-service crond start
rc-update add crond
```

> **Remember**: after adding cron, run `lbu commit` to persist the change (see
> [Alpine overlay backup](#alpine-overlay-backup-lbu-commit) below).

---

## Back up LVM volumes (live snapshots)

For VM disk volumes that live on LVM logical volumes, use **LVM snapshots** to
create crash-consistent copies without stopping the VM.

### Prerequisites

- The volume group containing the LV must have enough free space for the
  snapshot (typically 10–20 % of the LV size, depending on write activity).
- LVM snapshot support enabled in the kernel (default on Alpine).

### Create a snapshot

```bash
# List VGs and LVs
vgs
lvs

# Create a snapshot of a VM root volume (e.g. /dev/vg0/my-vm-root)
SNAPSHOT_SIZE="5G"   # adjust based on LV size and write rate
lvcreate -L "$SNAPSHOT_SIZE" -s -n my-vm-root-snap /dev/vg0/my-vm-root
```

### Back up from the snapshot

```bash
# Mount the snapshot (requires snapshot-merge target)
# Note: snapshots are read-only by default
mkdir -p /mnt/snap
mount -o ro /dev/vg0/my-vm-root-snap /mnt/snap

# Copy data to backup destination
rsync -aAXv /mnt/snap/ /mnt/backup/vms/my-vm-root/

# Unmount and remove the snapshot
umount /mnt/snap
lvremove -f /dev/vg0/my-vm-root-snap
```

### Scripted snapshot + backup

Save the following as `/root/scripts/backup-vm-lvm.sh`:

```bash
#!/bin/sh
# Usage: backup-vm-lvm.sh <vg>/<lv> <backup-dir>
# Example: backup-vm-lvm.sh vg0/my-vm-root /mnt/backup/vms

VG_LV="$1"
BACKUP_DIR="$2"

if [ -z "$VG_LV" ] || [ -z "$BACKUP_DIR" ]; then
    echo "Usage: $0 <vg>/<lv> <backup-dir>"
    exit 1
fi

VG="${VG_LV%/*}"
LV="${VG_LV#*/}"
SNAP_NAME="${LV}-snap-$(date +%s)"
SNAP_SIZE="5G"

# Create snapshot
lvcreate -L "$SNAP_SIZE" -s -n "$SNAP_NAME" "/dev/$VG/$LV"

# Mount and backup
mkdir -p "/mnt/snap-$SNAP_NAME"
mount -o ro "/dev/$VG/$SNAP_NAME" "/mnt/snap-$SNAP_NAME"
mkdir -p "$BACKUP_DIR/$VG/$LV"
rsync -aAXv "/mnt/snap-$SNAP_NAME/" "$BACKUP_DIR/$VG/$LV/"

# Cleanup
umount "/mnt/snap-$SNAP_NAME"
lvremove -f "/dev/$VG/$SNAP_NAME"
rmdir "/mnt/snap-$SNAP_NAME"
```

```bash
chmod +x /root/scripts/backup-vm-lvm.sh
```

---

## Back up the data volume (full-disk)

If the data volume (`/media/dkvmdata`) is itself an LVM logical volume, you can
snapshot the entire volume at once:

```bash
# Example: /media/dkvmdata is mounted from /dev/vg0/dkvmdata
lvcreate -L 10G -s -n dkvmdata-snap /dev/vg0/dkvmdata
mkdir -p /mnt/dkvmdata-snap
mount -o ro /dev/vg0/dkvmdata-snap /mnt/dkvmdata-snap

# Back up everything
rsync -aAXv /mnt/dkvmdata-snap/ /mnt/backup/dkvmdata/

umount /mnt/dkvmdata-snap
lvremove -f /dev/vg0/dkvmdata-snap
```

This captures VM configs, OVMF vars, logs, ISOs, and scripts in one go.

---

## Alpine overlay backup (lbu commit)

Alpine's diskless mode stores OS configuration (networking, SSH passwords,
GRUB settings) in RAM. To persist these changes to the USB stick, run:

```bash
lbu commit
```

This creates (or updates) `dkvm.apkovl.tar.gz` on the boot USB stick. The TUI
also does this automatically when you use **Save changes** in the Configuration
or Power tab.

### Back up the overlay file

Since the overlay lives on the USB stick, back it up separately:

```bash
# Find the USB boot partition (usually /media/usb or /media/<uuid>)
lsblk -o NAME,MOUNTPOINT

# Copy the overlay archive
cp /media/usb/dkvm.apkovl.tar.gz /mnt/backup/dkvm-overlay-$(date +%F).tar.gz
```

### Verify the overlay

```bash
tar -tzf /media/usb/dkvm.apkovl.tar.gz
```

This should list saved files such as `etc/network/interfaces`, `etc/shadow`,
etc.

---

## What does NOT need backup

| Item | Reason |
|------|--------|
| Alpine OS files (SquashFS on USB) | Can be re-created from Alpine installer ISO |
| RAM-based /tmp, /run, /proc | Volatile by design |
| `/media/dkvmdata/vms/<id>/qmp.sock` | QMP socket — recreated when VM starts |
| ISO images already backed up elsewhere | Optional; can be re-downloaded |
| OVMF code file | Part of the edk2-ovmf package |

---

## Restore procedure

Recover a DKVM host after hardware failure or accidental data loss. The
procedure assumes you have:

- A backup of `config.yaml`
- LVM snapshot backups (or raw file copies of the LV data)
- The Alpine overlay archive (`dkvm.apkovl.tar.gz`)

### Step 1 — Reinstall Alpine and boot

1. Boot the Alpine installer from a USB stick.
2. Run `setup-alpine` and configure diskless mode.
3. Boot into the fresh system.

### Step 2 — Restore the Alpine overlay

```bash
# Mount the boot USB stick
mount /dev/sdX1 /media/usb   # replace sdX1 with the correct device

# Restore the overlay from backup
cp /mnt/backup/dkvm-overlay-YYYY-MM-DD.tar.gz /media/usb/dkvm.apkovl.tar.gz

# Reboot to load the overlay
reboot
```

After reboot, your network bridge, SSH access, and GRUB parameters should be
restored.

### Step 3 — Mount the data volume

```bash
# Identify the data volume (e.g. /dev/vg0/dkvmdata)
lsblk

# Mount it
mkdir -p /media/dkvmdata
mount /dev/vg0/dkvmdata /media/dkvmdata

# Verify it looks correct
ls /media/dkvmdata/dkvmmanager/config.yaml
```

### Step 4 — Restore VM configuration

```bash
# If config.yaml is missing or corrupted, restore from backup
cp /mnt/backup/dkvmmanager/config-YYYY-MM-DD.yaml \
   /media/dkvmdata/dkvmmanager/config.yaml
```

### Step 5 — Recreate LVM volumes from snapshot backups

If you backed up LVM snapshot contents as files (via `rsync`), you need to
recreate the logical volumes and restore the data:

```bash
# Create a new LV of the same size (check original LV size from config.yaml)
lvcreate -L <size> -n <lv-name> <vg>

# Restore data
rsync -aAXv /mnt/backup/vms/<vg>/<lv>/ /dev/<vg>/<lv>
```

> **Note**: If you backed up the entire data volume via LVM snapshot and
> `rsync` (as shown in [Full-disk backup](#back-up-the-data-volume-full-disk)),
> you only need to re-mount it — all data inside is already intact.

### Step 6 — Start DKVM Manager

```bash
dkvmmanager
```

The TUI should read `config.yaml` and display your VMs. Start a VM and verify
it boots correctly.

---

## Test your backups

Periodically validate that backups can be restored:

1. Spare hardware or a VM: boot Alpine, mount backup media.
2. Run through the [Restore procedure](#restore-procedure) above.
3. Start a test VM and confirm it boots.

A backup that has never been restored is not a backup.

---

## See Also

- [Understanding LBU](../explanation/understanding-lbu.md) — Alpine diskless
  mode and overlay persistence
- [Storage](storage.md) — LVM volume creation and management
- [Power & Save](power-and-save.md) — saving configuration changes in the TUI
- [Setup](setup.md) — data volume mount point requirements
- [Alpine Wiki: Backup](https://wiki.alpinelinux.org/wiki/Alpine_local_backup) —
  official `lbu` documentation
- [Alpine Wiki: LVM](https://wiki.alpinelinux.org/wiki/LVM) — LVM on Alpine
