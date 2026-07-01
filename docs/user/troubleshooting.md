# Troubleshooting

Common issues when using DKVM Manager, with symptoms, causes, and solutions.

---

## 1. Mount point check fails

**Symptom:** Warning at startup: *"/media/dkvmdata is not a mount point"*

**Cause:** DKVM Manager checks that `/media/dkvmdata` is a real mount point (device ID differs from parent). A plain directory passes the same device ID check and triggers the warning.

**Solution:** Create a filesystem with `LABEL=dkvmdata` and mount it:

```bash
mkfs.ext4 -L dkvmdata /dev/sdX
mount /dev/sdX /media/dkvmdata
```

For testing, skip the check with `-skip-mount-check`:

```bash
./dkvmmanager -skip-mount-check
```

> **Source:** `internal/tui/models/mount_point_warning.go` → `isMountPoint()`

---

## 2. Insufficient hugepages

**Symptom:** Error message at startup: *"insufficient hugepages: have N, need M"*

**Cause:** QEMU uses 2 MB hugepages for VM memory backing. The host has fewer reserved than needed.

**Solution:** Reserve more hugepages temporarily:

```bash
echo 4096 > /proc/sys/vm/nr_hugepages
```

Or permanently via kernel command-line in GRUB (`/media/usb/boot/grub/grub.cfg`):

```
hugepages=4096
```

DKVM Manager's `hugepages.Ensure()` can auto-configure if run as root.

> **Source:** `internal/hugepages/hugepages.go` → `NewAutoConfig()`, `Ensure()`, `FormatError()`

---

## 3. VM fails to start

**Symptom:** VM creation succeeds but `Enter` on the VM in the list doesn't start it, or QEMU exits immediately.

**Common causes:**

| Cause | Check |
|-------|-------|
| QEMU binary not found | `which qemu-system-x86_64` — should return a path |
| OVMF firmware missing | `ls /usr/share/OVMF/OVMF_CODE.fd` — configure path in `~/.dkvmmanager.yaml` |
| swtpm not installed | `which swtpm` — needed when VM TPM is enabled |
| Data folder permissions | `ls -la /media/dkvmdata/` — must be readable by the user running DKVM Manager |
| VM name conflicts | Check `config.yaml` for duplicate names |

> **Source:** `internal/vm/vm_runner.go` → `Start()`; `internal/config/config.go` → default paths

---

## 4. QMP connection timeout

**Symptom:** `[STARTING]` badge never transitions to `[RUNNING]`. Log shows QEMU started but QMP connection failed.

**Cause:** QEMU's QMP socket didn't become ready within the timeout period. Possible reasons:
- QEMU process crashed immediately (see previous section)
- Socket path is on a slow filesystem
- Resource contention at startup

**Solution:**
1. Check `debug.log` for QEMU command-line and output
2. Run with `-dry-run` to see the QEMU command without executing:
   ```bash
   ./dkvmmanager -dry-run
   ```
3. Try launching the generated QEMU command manually to see stderr

> **Source:** `internal/vm/qmp_client.go` → QMP connection logic

---

## 5. Empty VM list after creation

**Symptom:** Created a VM but the VMs tab shows nothing.

**Cause:** DKVM Manager reads VM configuration from `/media/dkvmdata/dkvmmanager/config.yaml`. If this file is missing, inaccessible, or not a mount point, the list stays empty.

**Solution:**
1. Verify the data folder: `ls -la /media/dkvmdata/dkvmmanager/config.yaml`
2. Ensure `/media/dkvmdata` is a mount point (see issue #1)
3. Check file permissions — must be readable by your user
4. Press `r` on the VMs tab to refresh

> **Source:** `internal/vm/manager.go` → VM registry loading

---

## 6. LBU commit not found

**Symptom:** "Save changes" in Configuration tab shows an error about `lbu` not found.

**Cause:** DKVM Manager runs on Alpine Linux diskless mode. The `lbu` command is part of Alpine's `lbu` package and may not be installed on non-Alpine systems or minimal Alpine installs.

**Solution:**
```bash
apk add lbu
```

> **Source:** `internal/tui/models/debug.go` → LBU commit logic

---

## 7. PCI passthrough: device not found

**Symptom:** In the PCI passthrough configuration screen, expected devices don't appear or fail to bind.

**Causes and solutions:**

| Problem | Check |
|---------|-------|
| IOMMU not enabled | `dmesg | grep -i iommu` — should show IOMMU enabled |
| vfio-pci not loaded | `lsmod | grep vfio-pci` — load with `modprobe vfio-pci` |
| BIOS setting disabled | Enable VT-d (Intel) or AMD-Vi (AMD) in BIOS/UEFI |
| Device in wrong IOMMU group | `ls /sys/kernel/iommu_groups/*/devices/` — whole group must be passed together |

> **Source:** `docs/user/setup.md` → IOMMU section; `internal/vm/discovery.go` → `ScanPCIDevices()`

---

## 8. Terminal too small

**Symptom:** Warning: *"Terminal too small"* at startup.

**Cause:** DKVM Manager requires an 80×25 terminal. Your terminal is smaller.

**Solution:**
- Resize the terminal window to at least 80 columns × 25 rows
- Adjust font size to fit more characters
- On the DKVM host console, check font configuration

The warning is non-fatal — you can continue, but some views may not render correctly.

> **Source:** `internal/tui/tui.go` → `validateAndLogTerminalSize()`

---

## See Also

- [Setup & Prerequisites](setup.md) — system requirements and configuration
- [FAQ](faq.md) — frequently asked questions
- [Tutorial](tutorial.md) — step-by-step guide to your first VM
