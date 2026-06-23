# DKVM Manager

[![Go version](https://img.shields.io/github/go-mod/go-version/glemsom/dkvmmanager)](https://github.com/glemsom/dkvmmanager/blob/main/go.mod)
[![License](https://img.shields.io/github/license/glemsom/dkvmmanager)](LICENSE)
[![Latest release](https://img.shields.io/github/v/release/glemsom/dkvmmanager)](https://github.com/glemsom/dkvmmanager/releases/latest)
[![Build status](https://img.shields.io/github/actions/workflow/status/glemsom/dkvmmanager/release-please.yml?branch=main)](https://github.com/glemsom/dkvmmanager/actions)

Terminal UI for managing KVM/QEMU virtual machines on a single Linux host.

> **Part of [DKVM](https://github.com/glemsom/dkvm)** — a KVM-based hypervisor for single-host virtualization. DKVM Manager is the TUI frontend for the DKVM stack.

Built with Go and [BubbleTea](https://github.com/charmbracelet/bubbletea), offering keyboard-driven navigation through VM creation, hardware configuration, runtime monitoring, and system power operations.

## Features

- **VM lifecycle** — create, edit, delete, start, and stop VMs from the TUI
- **Live monitoring** — real-time log viewer, per-vCPU metrics, disk I/O rates, RSS tracking
- **Hardware configuration** — CPU topology, vCPU pinning, PCI/USB passthrough, Hyper-V enlightenments
- **Storage** — LVM logical volume creation and disk selection (file, block device, LVM volume)
- **Scripts & SSH** — start/stop hook scripts and SSH password management
- **Power management** — system power off, reboot, and LBU commit for persisting configuration

## Installation

### From GitHub Releases (recommended)

Download the latest pre-built binary for Linux amd64:

```bash
curl -LO https://github.com/glemsom/dkvmmanager/releases/latest/download/dkvmmanager-linux-amd64
chmod +x dkvmmanager-linux-amd64
sudo mv dkvmmanager-linux-amd64 /usr/local/bin/dkvmmanager
```

Verify the installation:

```bash
dkvmmanager --help
```

### Build from Source

Requires Go 1.26+ and Docker:

```bash
make build
```

The binary is produced at `./dkvmmanager`.

### Alpine Linux (DKVM host)

On the [DKVM](https://github.com/glemsom/dkvm) hypervisor (or any Alpine Linux system), download the pre-built binary:

```bash
curl -LO https://github.com/glemsom/dkvmmanager/releases/latest/download/dkvmmanager-linux-amd64
chmod +x dkvmmanager-linux-amd64
sudo mv dkvmmanager-linux-amd64 /usr/local/bin/dkvmmanager
```

## Quick Start

```bash
# Run (after installation)
dkvmmanager

# With debug logging
dkvmmanager -debug

# Dry-run (build QEMU command without executing)
dkvmmanager -dry-run
```

See [CLI Flags Reference](docs/reference/cli-flags.md) for all available flags.

### Minimum Requirements

- Linux host with KVM/QEMU
- 80×25 terminal
- `/media/dkvmdata` storage volume (or `-skip-mount-check` for testing)

## Documentation

### User Guide

| Document | Description |
|----------|-------------|
| [Setup & Prerequisites](docs/user/setup.md) | Install requirements — KVM, IOMMU, hugepages, LVM, GRUB, data folder |
| [VM Management](docs/user/vm-management.md) | Create, edit, and delete VMs through the TUI |
| [Hardware Configuration](docs/user/hardware-config.md) | CPU topology, vCPU pinning, CPU options, PCI/USB passthrough |
| [Scripts & SSH](docs/user/scripts-and-ssh.md) | Start/stop hook scripts and SSH password configuration |
| [Storage](docs/user/storage.md) | LVM logical volume creation and disk management |
| [Running VMs](docs/user/running-vms.md) | VM runtime — log viewer, status, metrics, stopping |
| [Power & Save](docs/user/power-and-save.md) | Power off, reboot, and saving configuration changes (LBU commit) |
| [Keybindings](docs/user/keybindings.md) | Complete keyboard reference for the TUI |
| [FAQ](docs/user/faq.md) | Frequently asked questions |
| [Troubleshooting](docs/user/troubleshooting.md) | Common issues and solutions |
| [CLI Flags Reference](docs/reference/cli-flags.md) | All command-line flags and options |
| [Example Scripts](examples/README.md) | Reference scripts for PCI passthrough |
| [User Guide Index](docs/user/README.md) | Overview of all user documentation |

### Developer Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/dev/architecture.md) | Package map, runner lifecycle, view registry, form framework, testing patterns |
| [CONTEXT.md](CONTEXT.md) | Domain glossary — ubiquitous language for the project |
| [ADR 0001](docs/adr/0001-runner-owned-running-vm-data-plane.md) | Key architectural decision — runner owns the running-VM data plane |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution workflow, commit conventions, release process |
| [CHANGELOG.md](CHANGELOG.md) | Release history, version changes, and migration notes |

### Agent Documentation

| Document | Description |
|----------|-------------|
| [Domain](docs/agents/domain.md) | Agent conventions for maintaining domain documentation |
| [Issue Tracker](docs/agents/issue-tracker.md) | GitHub Issues workflow for agents |
| [Triage Labels](docs/agents/triage-labels.md) | Issue triage label definitions |

## Keybindings Quick Reference

| Key | Context | Action |
|-----|---------|--------|
| `q` / `Ctrl+C` | Top-level (no VM) | Quit |
| `Tab` / `→` | Top-level | Next tab |
| `↑/↓` / `j/k` | Any list | Navigate |
| `Enter` / `Space` | Any list | Select |
| `Esc` | Sub-view | Cancel / return |

See [Keybindings](docs/user/keybindings.md) for the full reference.

## Build & Test

```bash
make build          # Build via Docker (golang:1.26-alpine)
make test           # Full test suite via Docker
make test-short     # Skip integration tests
```

## Project Layout

```
dkvmmanager/
├── main.go                    # Entry point
├── internal/
│   ├── vm/                    # Data plane — runner, QMP client, manager, metrics
│   ├── tui/                   # BubbleTea TUI entry point
│   │   ├── models/            # View plane — BubbleTea models, forms, key handlers
│   │   ├── components/        # Reusable UI components
│   │   └── styles/            # Lipgloss style definitions
│   ├── config/                # Configuration file loading
│   ├── models/                # Shared domain types (VM struct)
│   ├── hugepages/             # Hugepage detection and configuration
│   └── version/               # Version constant
├── docs/                      # Documentation
│   ├── user/                  # End-user documentation
│   ├── dev/                   # Developer documentation
│   ├── adr/                   # Architecture Decision Records
│   └── agents/                # Agent workflow documentation
├── examples/                  # Example scripts
├── VERSION                    # Current version
├── CHANGELOG.md               # Release changelog
├── CONTEXT.md                 # Domain glossary
├── CONTRIBUTING.md            # Contribution workflow
└── Makefile                   # Build targets
```

## License

MIT
