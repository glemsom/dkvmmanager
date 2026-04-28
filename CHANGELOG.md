# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- GitHub releases now include the compiled binary — release-please skips release creation so GoReleaser exclusively owns the release with assets (`.github/workflows/release-please.yml`)

## [0.1.0] - 2026-04-28

### Added

- **Breadcrumbs navigation**: Shows current UI location (e.g., Configuration > Delete VM) in sub-views
- **Stripped logical volumes**: Option to create striped LVMs when Volume Group has multiple PVs (auto-enabled by default)

### Changed

- **PCI passthrough dialog**: Device lines now show PCI address first (bold/high-contrast) for quick identification: `[X] 0000:01:00.0 [GPU] NVIDIA GeForce GTX 1080 [10de:1b80] (IOMMU:1)`
- **PCI passthrough dialog**: Devices are now grouped by IOMMU group with visual headers showing device count and selection status. Toggling a device auto-selects/deselects all devices in its IOMMU group (strict mode).
- **TPM state persistence**: TPM state directory (`{vm}/tpm/`) is now preserved across VM restarts instead of being deleted
- **Graceful TPM shutdown**: TPM processes are now shut down gracefully via the swtpm control channel before SIGTERM
- **Orphaned swtpm detection**: Stale swtpm processes from previous runs are detected via PID file and killed before starting a new one
- **Status bar display**: Status bar now shows VM execution state (▶ Running / ■ Stopped) instead of VM counts
- **UI rendering**: Apply consistent background colors and full-width rendering for header, tab bar, and list items
- **UI rendering (Power tab)**: Adjust width calculation to leave 4-char margin on right side to prevent terminal auto-wrap from hiding the right border
- **List item rendering**: Replace fixed-width rendering with padding-based indentation for VM list and menu items ("Selected > " and unselected "  " prefixes for consistent visual alignment)

### Removed

- **PCI ROM path field**: Removed the per-device ROM path text input from the PCI passthrough dialog (ROM field preserved in data model for backward compatibility but not editable in UI)
- **TPM config screen**: Removed the "Edit TPM Binary" configuration screen from the TUI (TPM binary is now configured via config file only)
- **TPMSocketPath config**: Removed `tpm_socket_path` configuration option (socket path is auto-derived per-VM)

### Fixed

- Fixed line calculation for save button in CPU options form (`internal/tui/models/cpu_options_form_navigation.go`)

## [0.0.1] - 2025-04-19

### Added

- **vCPU pinning support**: VMs can now be pinned to specific vCPUs for optimal performance
- **Multi-die vCPU pinning**: Support for pinning VMs across multiple CPU dies
- **Multifunction PCI device passthrough**: Proper handling of multifunction PCI devices (e.g., GPUs with audio functions)
- **Copy OVMF firmware**: Option to copy OVMF firmware files to VM directories
- **Create Logical Volume form**: UI in Configuration tab for creating LVMs
- **Power off system**: Power off the host system from the Power menu
- **Reboot system**: Reboot the host system from the Power menu
- **Hugepages allocation check**: Validates hugepages allocation for GPU passthrough VMs
- **Lean context compression**: Integrated leanctx for improved context handling

### Fixed

- **Start script execution order**: Start/stop scripts now run synchronously before/after QEMU, ensuring VFIO devices are bound to `vfio-pci` before QEMU attempts passthrough
- VM name input: spaces can now be typed in text fields
- Multifunction PCI device address allocation for secondary functions
- Silent failures and inconsistent error wrapping in OVMF file operations
- Test summary counts when no tests match
- Golden file formatting in config tab

### Changed

- Added yq and jq to Docker image for improved scripting

<!-- Links -->
[Unreleased]: https://github.com/glemsom/dkvmmanager/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/glemsom/dkvmmanager/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/glemsom/dkvmmanager/tree/v0.0.1