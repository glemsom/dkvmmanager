# DKVM Manager User Guide

DKVM Manager is a terminal UI (TUI) for managing KVM/QEMU virtual machines on a single Linux host.
Built with Go and [BubbleTea](https://github.com/charmbracelet/bubbletea), it provides keyboard-driven
navigation through VM creation, hardware configuration, runtime monitoring, and system power operations.

See [CONTEXT.md](../CONTEXT.md) for the project domain glossary.

## Getting Started

- [Setup & Prerequisites](setup.md) — install requirements, configure hugepages, IOMMU, LVM
- [VM Management](vm-management.md) — create your first VM

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
| [Troubleshooting](troubleshooting.md) | Common issues and solutions |

## Quick Reference

| Key | Context | Action |
|-----|---------|--------|
| `q` / `Ctrl+C` | Top-level (no VM running) | Quit application |
| `q` | Running VM view | Stop VM and return to menu |
| `Esc` | Sub-view (form, dialog) | Cancel / return to parent tab |
| `Tab` | Top-level | Switch tabs: VMs, Configuration, Power |
| `↑/↓` / `j/k` | Any list | Navigate items |
| `Enter` / `Space` | Any list | Select item |
| `r` | VMs tab | Refresh VM list |

See [Keybindings](keybindings.md) for the full reference.
