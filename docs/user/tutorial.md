# Tutorial: Create and Run Your First VM

This tutorial walks you through creating and running your first virtual machine
with DKVM Manager — from launch to a running guest.

**Time:** ~10 minutes
**Audience:** Linux user with KVM-capable hardware, new to DKVM Manager.

---

## Prerequisites

Before starting, ensure your host meets the minimum requirements:

- Linux with KVM/QEMU (verify: `ls /dev/kvm`)
- `/media/dkvmdata` mount point available (or use `-skip-mount-check` for testing)
- 80×25 terminal

> If anything is missing, see [Setup & Prerequisites](setup.md) for detailed
> installation steps.

---

## Step 1 — Build and launch DKVM Manager

Build the binary:

```bash
make build
```

Launch the TUI:

```bash
./dkvmmanager
```

If `/media/dkvmdata` is not a mount point, a warning appears. Press `Enter` or
`ESC` to dismiss and continue (for this tutorial, the warning is harmless).

You should see the main screen with three tabs:

| Tab | Content |
|-----|---------|
| **VMs** | List of virtual machines (empty) |
| **Configuration** | Menu for creating and configuring VMs |
| **Power** | Power off and reboot options |

---

## Step 2 — Explore the TUI

Try these navigation keys (press `q` to quit, but only when no VM is running):

- `Tab` / `→` — switch to next tab
- `↑/↓` / `j/k` — navigate lists

Spend a moment switching between tabs to get familiar with the layout.

The **VMs** tab shows a list of VMs (currently empty) with a detail panel on
the right. The **Configuration** tab lists all management actions. The **Power**
tab has power-off and reboot buttons.

---

## Step 3 — Create a VM

Now create your first VM.

1. Press `Tab` to switch to the **Configuration** tab.
2. Make sure **Add VM** is highlighted (index 0) and press `Enter`.

The VM creation form opens. It has these fields:

| Field | What to enter |
|-------|---------------|
| **VM Name** | `my-first-vm` (alphanumeric, spaces, dashes, underscores) |
| **Hard Disks** | Leave one empty slot — you'll add a disk |
| **CD/DVD Drives** | Leave empty — no ISO needed |
| **MAC Address** | Auto-generated — leave as-is |
| **VNC** | `ON` — enables remote console access |
| **Network** | `NAT` — default network mode |
| **TPM** | `OFF` — not needed for this tutorial |

3. Press `Tab` to move between fields. Toggle VNC, Network, and TPM with
   `Space` or `Enter`.
4. Navigate to the **Save** button and press `Enter`.

The form validates and creates the VM. A status message confirms success, and
you return to the Configuration tab.

> **What happened?** DKVM Manager saved the VM configuration as YAML in
> `/media/dkvmdata/dkvmmanager/config.yaml`. The VM now exists in the
> repository.

---

## Step 4 — Start the VM

1. Press `Tab` to switch to the **VMs** tab.
2. Your VM `my-first-vm` appears in the list.
3. Press `Enter` to start it.

The screen switches to the **running view**. You'll see:

- `[STARTING]` badge — QEMU is booting, QMP not yet connected
- A log viewport showing QEMU output
- An info panel at the top with VM metadata

After a few seconds, the badge changes to `[RUNNING]`. The info panel now shows
live metrics (updated every 2 seconds):

- **vCPU%** — per-vCPU utilization
- **Host** — QEMU process CPU% and RSS memory
- **disk** — per-disk I/O rates

The log viewport streams QEMU's console output. Scroll with `↑/↓` or
`PgUp`/`PgDn`.

---

## Step 5 — Stop the VM

Press `q` in the running view to stop the VM gracefully.

The badge changes to `[STOPPING]` while QEMU shuts down. When the VM stops,
you're returned to the main menu automatically.

> **Try also:** `Ctrl+C` force-kills the QEMU process (SIGKILL). Use `q` for
> normal shutdown.

---

## Step 6 — Save changes (LBU commit)

DKVM Manager runs on Alpine Linux diskless mode. Changes made to the
configuration are lost on reboot unless persisted with `lbu commit`.

1. Switch to the **Configuration** tab.
2. Scroll to the bottom of the menu.
3. Select **Save changes** (the last item).

The status bar confirms the LBU commit. Your VM configuration is now
persistent.

> **Why is this needed?** DKVM runs from a RAM-based overlay. `lbu commit`
> writes the overlay changes to persistent storage. See
> [Power & Save](power-and-save.md) for details.

---

## Step 7 — What's next?

You've completed the full VM lifecycle: create, start, monitor, stop, and save.

Try these next:

| Guide | What you'll learn |
|-------|-------------------|
| [VM Management](vm-management.md) | Editing VMs, adding disks, deleting VMs |
| [Hardware Configuration](hardware-config.md) | CPU topology, vCPU pinning, PCI/USB passthrough |
| [Storage](storage.md) | Creating LVM logical volumes for VM disks |
| [Scripts & SSH](scripts-and-ssh.md) | Start/stop hook scripts and SSH password |
| [Running VMs](running-vms.md) | Live metrics, log viewer, force stop |
| [Power & Save](power-and-save.md) | Power off, reboot, LBU commit |
| [Troubleshooting](troubleshooting.md) | Common issues and solutions |

---

## Keybindings reference

See [Keybindings](keybindings.md) for the complete keyboard reference.
