# Frequently Asked Questions

Common questions about DKVM Manager.

---

## Can I run multiple VMs at the same time?

No. DKVM Manager is a single-host, single-VM-at-a-time tool. One QEMU process runs at most. Starting a VM while another is running returns an error.

> **Source:** `internal/vm/vm_runner.go` — single runner instance per session

---

## Where is my VM config stored?

All VM configurations are in `/media/dkvmdata/dkvmmanager/config.yaml`. This is a YAML file read at startup and updated when you create, edit, or delete VMs.

> **Source:** `internal/config/config.go` — `VmsConfigFile` default

---

## Why do I need to run `lbu commit`?

DKVM runs on Alpine Linux in **diskless mode**. The entire filesystem lives in RAM (overlay). Changes to configuration, SSH passwords, and scripts are lost on reboot unless persisted.

`lbu commit` writes the overlay changes to persistent storage. See [Power & Save](power-and-save.md) for details.

> **Source:** `internal/tui/models/power_and_save.go` — LBU commit logic

---

## Do I need the DKVM hypervisor to use DKVM Manager?

Yes. DKVM Manager is part of the [DKVM](https://github.com/glemsom/dkvm) stack. It depends on:
- The DKVM data folder layout (`/media/dkvmdata`)
- The DKVM diskless boot and overlay persistence model
- The DKVM mount point convention (filesystem label `dkvmdata`)

It is not a standalone general-purpose QEMU management tool.

---

## What terminal emulator do I need?

Any 80×25 terminal that supports basic ANSI escape codes and UTF-8 works:

- **Linux console TTY** (TERM=linux) — uses CP437 box-drawing characters
- **xterm**, **kitty**, **alacritty**, **GNOME Terminal**, **Konsole** — all work
- **tmux** / **screen** — work inside a terminal that meets the above requirements

> **Source:** `docs/terminal-capabilities.md` — full analysis of terminal requirements

---

## How do I SSH into the host?

Use Configuration → **Set SSH Password** (index 9) to set or change the host password. This runs `chpasswd` and persists with `lbu commit`. Then SSH in normally:

```bash
ssh user@host-ip
```

> **Source:** [Scripts & SSH](scripts-and-ssh.md) → SSH Password section

---

## Can I use DKVM Manager without IOMMU?

Yes. CPU topology configuration, VM creation, editing, and management all work without IOMMU. Only **PCI passthrough** requires IOMMU support (VT-d / AMD-Vi).

> **Source:** `docs/user/setup.md` → IOMMU section

---

## What happens if I Ctrl+C while a VM is running?

Pressing `Ctrl+C` sends SIGKILL to the QEMU process — equivalent to yanking the power cord. The VM stops immediately without graceful shutdown.

**Use `q` instead** for graceful shutdown (`system_powerdown` via QMP). See [Running VMs](running-vms.md).

---

## Does DKVM Manager support UEFI or legacy BIOS?

DKVM Manager uses **OVMF (UEFI)** firmware only. Each VM gets a copy of `OVMF_CODE.fd` and `OVMF_VARS.fd`. Legacy BIOS/SeaBIOS is not supported.

> **Source:** `internal/config/config.go` — `BiosCode`, `BiosVars` defaults

---

## See Also

- [Troubleshooting](troubleshooting.md) — common issues and solutions
- [User Guide Index](README.md) — all documentation
