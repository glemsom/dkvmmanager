# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [0.1.16] - 2026-05-16

### Added
- **Centered modal sizing for mount point warning**: The mount point warning dialog is now centered in the terminal window using `lipgloss.Place`, with terminal dimensions stored via `SetSize` and environment variable fallback (`COLUMNS`/`LINES`). Includes comprehensive tests for View() and SetSize (`internal/tui/models/mount_point_warning.go`, `internal/tui/models/mount_point_warning_test.go`)
- **Terminal size validation with fallback**: Added `getInitialTerminalSize()` that queries terminal via `IoctlGetWinsize` with fallback to `COLUMNS`/`LINES` env vars, avoiding circular imports (`internal/tui/models/init.go`)
- **View registry deactivation**: The view registry is now properly deactivated when navigating between views to prevent stale state in sub-views (`internal/tui/models/key_handlers.go`, `internal/tui/models/message_handlers.go`)

### Changed
- **"Use Host Topology" toggle moved**: Removed from CPU topology form and relocated to the vCPU pinning form. The toggle is now saved to CPU topology config from the vCPU pinning form. Positions updated (toggle → use_host_topology → save → apply_kernel) (`internal/tui/models/cpu_topology_form.go`, `internal/tui/models/vcpu_pinning_form.go`, `internal/tui/models/vcpu_pinning_form_render.go`)
- **Pinning enabled toggle extracted**: Inline toggle rendering extracted into dedicated `renderPinningEnabledToggle()` method with ON/OFF style (`internal/tui/models/vcpu_pinning_form_render.go`)

### Fixed
- **Visual alignment of list item selection indicators**: Applied consistent `"> "` / `"  "` gutter to all rows in VMTable and list adapters instead of only the selected row, ensuring proper column alignment and visual hierarchy (`internal/tui/components/vm_table.go`, `internal/tui/models/list_adapter.go`)

## [0.1.15] - 2026-05-16

### Fixed
- **Start script output race**: Added channel synchronization in `executeStartScript()` to ensure script output goroutines finish before the function returns, preventing lost log output during start script execution (`internal/vm/vm_runner.go`)
- **VM status stuck on STARTING**: Added polling timeout fallback with 10s timeout in `pollStatus()` to handle VMs where QMP socket never appears, fixing permanently stuck "STARTING" status (`internal/vm/vm_runner.go`)

## [0.1.14] - 2026-05-15

### Fixed
- **Per-die core ID mapping**: On multi-die systems, host physical core IDs may not be zero-based within each die (e.g. die 0 has cores 0-7, die 1 has cores 10-17). QEMU's `-smp cores=N` requires core-ids in range 0:(N-1) per die. Added `buildDieLocalCoreMap()` to sort physical core IDs per die and assign 0-based local indices for correct `-device host-x86_64-cpu` declarations (`internal/vm/vm_runner_config.go`)

## [0.1.13] - 2026-05-15

### Fixed
- **QEMU duplicate APIC ID crash**: When using "Use host CPU topology", `-smp` previously provisioned all selected vCPUs automatically, causing QEMU to reject the explicit `-device host-x86_64-cpu` declarations with `CPU[N] with APIC ID N exists`. Fixed by setting `-smp 1,maxcpus=...` so only CPU 0 is auto-created and the remaining CPUs are added via explicit `-device` lines (`internal/vm/vm_runner_config.go`)
- **VM status stuck on STARTING**: Handle `VMStartedMsg` before registry dispatch to prevent VMs from permanently showing "STARTING" status after a fast QEMU exit (`internal/vm/manager.go`)

## [0.1.12]

### Added
- **Use host CPU topology toggle**: New `use_host_topology` toggle in the CPU topology form that allows inheriting the host's CPU topology automatically (`internal/tui/models/cpu_topology_form.go`, `internal/tui/models/cpu_topology_form_navigation.go`, `internal/tui/models/cpu_topology_form_render.go`, `internal/tui/models/cpu_topology_form_validation.go`, `internal/tui/models/cpu_topology_form_test.go`)
  - Added navigation support for the new toggle position
  - Added rendering for the toggle in the form view
  - Added validation and save logic to persist `UseHostTopology` in the `CPUTopology` model
  - Added comprehensive tests for toggle navigation, rendering, and validation
- **ForceCPUID0x80000026 toggle**: New CPU option to force CPUID leaf 0x80000026 (`internal/models/models.go`, `internal/tui/models/fields/cpu_options.go`, `internal/vm/repository.go`)
- **Asymmetric vCPU topology mapper**: New `GenerateAsymmetricCPUDevices` function for building per-die CPU device declarations with correct APIC IDs (`internal/vm/vcpu_topology_mapper.go`)

### Fixed
- **Multi-die vCPU pinning**: Allow asymmetric topologies with multiple dies instead of rejecting them as errors (`internal/vm/vm_runner_pinning.go`)

### Removed
- **vCPU threads display**: Removed thread ID display from the VM running view (`internal/tui/models/vm_running.go`)

## [0.1.11] - 2026-05-05

### Added
- **PCI bridge detection**: New `IsBridge` field on `PCIDevice` to identify PCI-to-PCI bridges (class code `0604`), enabling better device classification in the PCI passthrough form (`internal/models/models.go`, `internal/vm/pci_scanner.go`, `internal/tui/models/pci_passthrough_form.go`)

## [0.1.10] - 2026-05-04

### Added
- **Mouse scroll support**: `tea.MouseMsg` handling in all scrollable forms (CPU, PCI, USB, SSH, scripts, vCPU pinning) for trackpad and mouse wheel scrolling

### Fixed
- **Multi-line scroll-to-focus**: ScrollableForm now tracks actual rendered line counts per position so scroll-to-focus works correctly when positions render to multiple lines (e.g. button rows with blank separators) (`internal/tui/models/form/form.go`, `internal/tui/models/form/form_test.go`)

## [0.1.9] - 2026-05-04

### Added
- **Mount point warning**: Warn when `/media/dkvmdata` is not a mount point, helping users detect misconfigured storage (`internal/tui/models/mount_point_warning*.go`)

### Fixed
- **Mount point warning colors**: Use `styles.Colors.Foreground` for consistent theming in the mount point warning view (`internal/tui/models/mount_point_warning*.go`)

## [0.1.8] - 2026-05-03

### Changed
- **CPU Options form**: Migrated to ScrollableForm framework for improved navigation and consistency (`internal/tui/models/cpu_options_form_*.go`)
- **CPU Topology form**: Migrated to ScrollableForm framework (`internal/tui/models/cpu_topology_form*.go`)
- **PCI Passthrough form**: Migrated to ScrollableForm framework (`internal/tui/models/pci_passthrough_form*.go`)
- **VM form**: Migrated to ScrollableForm framework (`internal/tui/models/vm_form*.go`)
- **Main TUI**: Simplified message routing through ScrollableForm (`internal/tui/main.go`)

### Added
- **Dropdown navigation**: Left/right arrow key handling in dropdown fields for easier value selection (`internal/tui/models/form_dropdown.go`)

## [0.1.7](https://github.com/glemsom/dkvmmanager/compare/v0.1.5...v0.1.7) - 2026-05-03

### Added
- **vCPU Pinning Apply to Kernel**: New "Apply to Kernel" button in the vCPU pinning form that writes CPU isolation parameters (`isolcpus`, `nohz_full`, `rcu_nocbs`) to `grub.cfg`, enabling persistent kernel-level CPU isolation for pinned VMs (`internal/vm/grub_config.go`, `internal/vm/grub_config_test.go`, `internal/vm/manager.go`, `internal/tui/models/vcpu_pinning_form*.go`, `internal/tui/models/message_handlers.go`)

### Fixed
- **PCI passthrough Apply to Kernel**: Remount `/media/usb` as `rw` before modifying `grub.cfg` and restore to `ro` afterward, since DKVM Hypervisor keeps the USB filesystem read-only by default (`internal/vm/manager.go`)

## [0.1.5](https://github.com/glemsom/dkvmmanager/compare/v0.1.4...v0.1.5) - 2026-05-03

### Added
- **PCI passthrough Apply to Kernel**: New "Apply to Kernel" button in the PCI passthrough form that writes selected device VFIO IDs to `grub.cfg`'s `vfio-pci.ids` kernel parameter, enabling persistent kernel-level VFIO binding (`internal/vm/grub_config.go`, `internal/vm/grub_config_test.go`, `internal/tui/models/pci_passthrough_form_*.go`)
- **GRUB VFIO config utilities**: `UpdateGrubVFIOIDs` function to safely update `vfio-pci.ids` in grub.cfg with backup support, parameter removal, and whitespace cleanup (`internal/vm/grub_config.go`)

### Changed
- **PCI passthrough validation form**: Extended with kernel apply button, status messages, and asynchronous command handling (`internal/tui/models/pci_passthrough_form_*.go`)
- **VM manager**: Added `ApplyVFIOIDsToKernel` method to orchestrate grub.cfg updates from PCI passthrough config (`internal/vm/manager.go`)
- **Config**: Extended config save with additional logging (`internal/config/config.go`)

### Fixed
- Ensure VM form is focused when opened (new/edit mode)
- **CPU Power Management toggle**: Fix `-overcommit cpu-pm=on` now respects the CPUPM setting (was previously hardcoded and always enabled)
- **GRUB VFIO config**: Rewrite `UpdateGrubVFIOIDs` to process lines individually, ensuring `vfio-pci.ids=` appears exactly once per linux line and never on non-linux lines (e.g. initrd lines) (`internal/vm/grub_config.go`)

## [0.1.4](https://github.com/glemsom/dkvmmanager/compare/v0.1.3...v0.1.4) - 2026-05-01

### Fixed

- Prevent TUI freeze during VM startup by running `runner.Start()` in a goroutine instead of blocking the BubbleTea event loop

## [0.1.3](https://github.com/glemsom/dkvmmanager/compare/v0.1.2...v0.1.3) (2026-04-30)

### Added

- **CPU Power Management toggle**: New `CPUPM` option in CPU options form to enable CPU power management passthrough (`cpu-pm=on`) to QEMU (`internal/models/models.go`, `internal/vm/vm_runner_config.go`, `internal/tui/models/cpu_options_form_*.go`)

## [0.1.2](https://github.com/glemsom/dkvmmanager/compare/v0.1.1...v0.1.2) (2026-04-30)

### Changed

- **Panel borders**: Replace rounded borders with normal (straight) borders in layered panels and titled panels for consistent visual alignment

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
[Unreleased]: https://github.com/glemsom/dkvmmanager/compare/v0.1.16...HEAD
[0.1.16]: https://github.com/glemsom/dkvmmanager/compare/v0.1.15...v0.1.16
[0.1.15]: https://github.com/glemsom/dkvmmanager/compare/v0.1.14...v0.1.15
[0.1.14]: https://github.com/glemsom/dkvmmanager/compare/v0.1.13...v0.1.14
[0.1.13]: https://github.com/glemsom/dkvmmanager/compare/v0.1.12...v0.1.13
[0.1.12]: https://github.com/glemsom/dkvmmanager/compare/v0.1.11...v0.1.12
[0.1.11]: https://github.com/glemsom/dkvmmanager/compare/v0.1.10...v0.1.11
[0.1.10]: https://github.com/glemsom/dkvmmanager/compare/v0.1.9...v0.1.10
[0.1.9]: https://github.com/glemsom/dkvmmanager/compare/v0.1.8...v0.1.9
[0.1.8]: https://github.com/glemsom/dkvmmanager/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/glemsom/dkvmanager/compare/v0.1.5...v0.1.7
[0.1.5]: https://github.com/glemsom/dkvmmanager/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/glemsom/dkvmmanager/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/glemsom/dkvmmanager/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/glemsom/dkvmmanager/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/glemsom/dkvmmanager/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/glemsom/dkvmmanager/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/glemsom/dkvmmanager/tree/v0.0.1
