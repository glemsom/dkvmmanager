# DKVM Manager User Guide

DKVM Manager is a terminal UI (TUI) for managing KVM/QEMU virtual machines on a single Linux host.
Built with Go and [BubbleTea](https://github.com/charmbracelet/bubbletea), it provides keyboard-driven
navigation through VM creation, hardware configuration, runtime monitoring, and system power operations.

See [CONTEXT.md](../CONTEXT.md) for the project domain glossary.

## Getting Started

- [Setup & Prerequisites](setup.md) — install requirements, configure hugepages, IOMMU, LVM
- [VM Management](vm-management.md) — create your first VM

## Learning Path

New to DKVM Manager? Follow this recommended reading order — each document builds on the previous one.

1. **[Tutorial](tutorial.md)** — create your first VM in 10 minutes
2. **[Setup & Prerequisites](setup.md)** — host preparation: KVM, IOMMU, hugepages, LVM
3. **[VM Management](vm-management.md)** — deeper VM operations: create, edit, delete
4. **[Hardware Configuration](hardware-config.md)** — CPU topology, pinning, PCI/USB passthrough
5. **[Running VMs](running-vms.md)** — runtime monitoring, logs, metrics
6. **[Storage](storage.md)** — LVM logical volume creation and disk management
7. **[Power & Save](power-and-save.md)** — power off, reboot, and saving configuration changes

This complements the alphabetical index below by providing a purpose-driven reading order.

## Documentation Index

| Section | Description |
|---------|-------------|
| [Setup](setup.md) | Host prerequisites: KVM, IOMMU, hugepages, LVM, data folder |
| [VM Management](vm-management.md) | Create, edit, and delete VMs through the TUI |
| [Hardware Configuration](hardware-config.md) | CPU topology, vCPU pinning, CPU options, PCI/USB passthrough |
| [Scripts & SSH](scripts-and-ssh.md) | Start/stop hook scripts and SSH password configuration |
| [Storage](storage.md) | LVM logical volume creation and disk management |
| [Running VMs](running-vms.md) | VM runtime: log viewer, status, metrics, stopping |
| [Power & Save](power-and-save.md) | Power off, reboot, and saving configuration changes (LBU commit) |
| [Keybindings](keybindings.md) | Complete keyboard reference for the TUI |
| [FAQ](faq.md) | Frequently asked questions |
| [Upgrade & Migration](upgrade.md) | Upgrade binary, Alpine package, version compatibility, migrating to a new host |
| [Troubleshooting](troubleshooting.md) | Common issues and solutions |
| [Backup & Restore](backup-and-restore.md) | Protect VM configs, LVM volumes, and Alpine overlay; restore after failure |

## Understanding DKVM Manager

- [How DKVM Manager Works](../explanation/how-dkvm-manager-works.md) — architecture overview for end users: two-layer config, runner lifecycle, QMP, hugepages, metrics
- [Understanding LBU](../explanation/understanding-lbu.md) — why `lbu commit` exists, Alpine diskless mode, what persists and what doesn't

## Quick Reference

| Key | Context | Action |
|-----|---------|--------|
| `q` / `Ctrl+C` | Top-level (no VM running) | Quit application |
| `Tab` | Top-level | Switch tabs: VMs, Configuration, Power |
| `↑/↓` / `j/k` | Any list | Navigate items |
| `Enter` / `Space` | Any list | Select item |
| `Esc` | Sub-view (form, dialog) | Cancel / return to parent tab |

See [Keybindings](keybindings.md) for the full reference.
