# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Stripped logical volumes**: Option to create striped LVMs when Volume Group has multiple PVs (auto-enabled by default)

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

- VM name input: spaces can now be typed in text fields
- Multifunction PCI device address allocation for secondary functions
- Silent failures and inconsistent error wrapping in OVMF file operations
- Test summary counts when no tests match
- Golden file formatting in config tab

### Changed

- Added yq and jq to Docker image for improved scripting

<!-- Links -->
[Unreleased]: https://github.com/glemsom/dkvmmanager/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/glemsom/dkvmmanager/tree/v0.0.1