# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Changed
- **User docs**: Added `docs/user/storage.md` covering LVM logical volume creation
  and disk management for VMs.
  ([#100](https://github.com/glemsom/dkvmmanager/issues/100))

## [0.2.0] - 2026-06-22

### Added
- **Host L3 cache associativity detection**: New `L3CacheAssoc` field on CPU
  topology that shows detected L3 cache associativity per die in the CPU options
  form. The CPU scanner now populates associativity alongside cache size.
  ([#86](https://github.com/glemsom/dkvmmanager/issues/86))
  (`internal/models/models.go`, `internal/vm/cpu_scanner.go`,
  `internal/tui/models/cpu_options_form_navigation.go`)

### Fixed
- **Status bar position and styling**: Status bar now sits at screen bottom;
  faint (dim) style applied only to chrome borders, not to form content
  (`internal/tui/models/view.go`)
- **Config view rendering after resize**: `WindowSizeMsg` handler now runs before
  registry dispatch, fixing config view rendering when terminal is resized
  ([#84](https://github.com/glemsom/dkvmmanager/issues/84))
  (`internal/tui/models/key_handlers.go`)
- **SSH Password Apply button format**: Uses standard
  `[Space/Enter] Apply [ESC] Cancel` format
  ([#89](https://github.com/glemsom/dkvmmanager/issues/89))
  (`internal/tui/models/ssh_password_form.go`)
- **CPU Topology save hint**: Uses `[Space/Enter] Save` instead of `[Enter] Save`
  ([#90](https://github.com/glemsom/dkvmmanager/issues/90))
  (`internal/tui/models/cpu_topology_form_render.go`)
- **Start/Stop Script Cancel highlighting**: Cancel button highlights Cancel
  instead of Save
  ([#91](https://github.com/glemsom/dkvmmanager/issues/91))
  (`internal/tui/models/start_stop_script_form_render.go`)
- **Standardised footer help text**: All 8 forms now use consistent
  `[Space/Enter] <Action>  [ESC] Cancel` footer format
  ([#92](https://github.com/glemsom/dkvmmanager/issues/92))
  (9 files across CPU, PCI, USB, SSH, script, vCPU pinning, VM forms)
- **Symmetric vCPU Pinning toggle rendering**: `[ON]/[OFF]` toggle now renders
  symmetrically in the vCPU pinning form
  (`internal/tui/models/vcpu_pinning_form_render.go`,
  `internal/tui/models/vcpu_pinning_form_test.go`,
  `docs/terminal-capabilities.md`)
- **[Space/Enter] consistency and cancel highlighting**: Ensured all forms use
  `[Space/Enter]` for save/apply actions and Cancel buttons highlight correctly
  (`internal/tui/models/cpu_topology_form.go`,
  `internal/tui/models/start_stop_script_form_render.go`)

### Changed
- **CPU scanner refactored**: Extracted `findL3CacheIndexPath`, fixed variable
  naming, improved fallback logic
  (`internal/vm/cpu_scanner.go`)

## [0.1.31] - 2026-06-21

### Added
- **Per-die L3 cache size override**: New `L3CacheSizeDie` CPU option to set
  `l3-cache-size-die<N>=<size>` QEMU flags for asymmetric AMD V-cache CPUs
  (9950X3D, EPYC). The CPU options form now shows per-die L3 cache size
  text fields with detected sizes from the host topology scanner.
  ([#79](https://github.com/glemsom/dkvmmanager/issues/79))
  (`internal/models/models.go`, `internal/vm/vm_runner_config.go`,
  `internal/tui/models/cpu_options_form.go`,
  `internal/tui/models/cpu_options_form_navigation.go`,
  `internal/tui/models/cpu_options_form_render.go`)
- **Per-die L3 cache associativity override**: New `L3CacheAssocDie` CPU option
  to set `l3-cache-assoc-die<N>=<value>` QEMU flags for asymmetric AMD V-cache
  CPUs (9950X3D, EPYC). The CPU options form now shows per-die L3 cache
  associativity text fields alongside the size fields.
  ([#80](https://github.com/glemsom/dkvmmanager/issues/80))

### Fixed
- **Version mismatch**: Synced `internal/version/version.go` with `VERSION` file
  and added `make bump-version` target + documentation to prevent future drift.
  (`internal/version/version.go`, `Makefile`, `CONTRIBUTING.md`)
- **Force CPUID 0x80000026 toggle was dead**: The `ForceCPUID0x80000026` CPU
  option was modeled and exposed in the TUI as 'Force AMD CPUID' but
  `buildCPUOptsString()` never emitted the corresponding QEMU property.
  Fixed to append `x-force-cpuid-0x80000026=on` to `-cpu host,...`.
  ([#78](https://github.com/glemsom/dkvmmanager/issues/78))
  (`internal/vm/vm_runner_config.go`)

## [0.1.30] - 2026-06-21

### Fixed
- **Data race in dry-run path**: Removed redundant `persistBuf.Flush()` call in dry-run path to fix a data race ([#76](https://github.com/glemsom/dkvmmanager/issues/76)) (`internal/vm/vm_runner.go`)
- **Straight borders for TERM=linux**: Replaced rounded borders with straight box-drawing characters (CP437-compatible) for Linux console support, fixing '+' fallback rendering on VGA console ([#74](https://github.com/glemsom/dkvmmanager/issues/74)) (`internal/tui/components/vm_cards.go`, `internal/tui/components/vm_details.go`, `internal/tui/styles/ascii_fallback.go`, `internal/tui/styles/colors.go`, `internal/tui/styles/styles.go`)

## [0.1.29] - 2026-06-20

### Fixed
- **VM metrics disappear after QMP connects**: Fixed `Snapshot()` to use `query-cpus-fast` instead of deprecated `query-cpus` (removed in QEMU 7.1+), preventing early return that dropped all metrics including host RSS/CPU (`internal/vm/vm_runner.go`, `internal/tui/models/key_handlers.go`)

## [0.1.28] - 2026-06-19

### Added
- **ASCII fallback for TERM=linux (vgacon)**: VM card borders, panel borders, status symbols, mode icons, and disk/CDROM bullets now fall back to ASCII on the Linux console where Unicode glyphs may not render (`internal/tui/styles/ascii_fallback.go`, `internal/tui/styles/ascii_fallback_test.go`, `internal/tui/components/vm_cards.go`, `internal/tui/components/statusbar.go`, `internal/tui/styles/colors.go`, `internal/tui/styles/styles.go`)

### Fixed
- **Invisible muted text on Linux console**: ANSI color 8 (bright black) renders as black on TERM=linux; fall back to ANSI 7 (light gray) for muted/dim text (`internal/tui/theme/theme.go`)
- **ASCII fallback for Unicode symbols in status bar**: Mode indicator and running/stopped status now use ASCII fallback on TERM=linux (`internal/tui/components/statusbar.go`, `internal/tui/styles/colors.go`)
- **Corrected colorprofile import path in test file**: Fixed `test_lipgloss2.go` import path for colorprofile (`test_lipgloss2.go`)

### Changed
- **Power menu reordered**: Power off listed before Reboot for safety-first ordering (`internal/tui/models/init.go`, `internal/tui/models/key_handlers.go`)

## [0.1.25] - 2026-06-10

### Added
- **Persisted QEMU log per VM**: The runner now writes a persistent log to `<DataFolder>/vms/<id>/qemu.log` mirroring QEMU stdout, QEMU stderr, and start/stop script output, so log history survives view re-entry and VM exit (S1) (`internal/vm/vm_runner.go`, `internal/vm/vm_runner_test.go`)
- **View replays persisted log on entry**: The running view now replays the persisted log when the user re-enters it, draining the runner's subscription channel so live lines continue without duplication (S2) (`internal/tui/models/vm_running.go`, `internal/vm/vm_runner.go`, `internal/vm/vm_runner_test.go`)
- **Per-vCPU CPU% metrics**: Live per-vCPU CPU utilisation is now plumbed from QMP through the metrics snapshot and displayed in the running view (S3) (`internal/vm/metrics.go`, `internal/vm/metrics_test.go`, `internal/vm/proc.go`, `internal/vm/proc_test.go`, `internal/vm/qmp_client.go`, `internal/vm/qmp_client_test.go`, `internal/vm/vm_runner.go`, `internal/tui/models/vm_running.go`, `internal/tui/models/vm_running_test.go`)
- **Host-side QEMU process CPU% and RSS**: The running view now shows the QEMU process's host-side CPU% and RSS, derived from `/proc/<pid>` (S4) (`internal/vm/proc.go`, `internal/vm/proc_test.go`, `internal/vm/metrics.go`, `internal/vm/metrics_test.go`, `internal/vm/vm_runner.go`, `internal/tui/models/vm_running.go`, `internal/tui/models/vm_running_test.go`)
- **Per-disk IOPS/B/s and balloon in the running view**: Per-block IOPS and bytes/s plus the current balloon size are now part of the metrics snapshot and rendered in the running view (S5) (`internal/vm/metrics.go`, `internal/vm/metrics_test.go`, `internal/vm/qmp_client.go`, `internal/vm/qmp_client_test.go`, `internal/vm/vm_runner.go`, `internal/tui/models/vm_running.go`, `internal/tui/models/vm_running_test.go`)

### Changed
- **Domain glossary and ADR for the runner-owned data plane**: New `CONTEXT.md` glossary locks the terms Runner / Manager / QMP client / Persisted log / Log subscription / Metrics snapshot / Status poll, and ADR 0001 records the decision that the runner owns the running-VM data plane (`CONTEXT.md`, `docs/adr/0001-runner-owned-running-vm-data-plane.md`)

### Fixed
- **Start error cleanup of persisted log**: The dry-run check now runs before hugepages setup, and on `Start` errors the persisted log file is removed to avoid stale state on the next run (`internal/vm/vm_runner.go`, `internal/vm/vm_runner_test.go`)
- **TestHandleVMSelection race with async persisted log**: Stabilised the test against the async persisted-log write so it no longer flakes (`internal/tui/models/main_test.go`)

## [0.1.24] - 2026-06-09

### Added
- **RunConfig struct and LoadRunConfigFromRepo helper**: New run configuration persistence for VMs with `LoadRunConfigFromRepo` helper function (`internal/vm/run_config.go`, `internal/vm/run_config_test.go`)
- **RunConfig wired through NewVMRunner**: `RunConfig` struct is now passed through `NewVMRunner` and used during VM startup (closes #38) (`internal/tui/models/key_handlers.go`, `internal/vm/vm_runner.go`, `internal/vm/vm_runner_config.go`, etc.)
- **Charmbracelet library migration from v1 to v2**: Migrated `bubbles`, `bubbletea`, and `lipgloss` libraries to their v2 APIs across the entire codebase
- **Key event handling migrated from Bubble Tea v1 to v2**: Updated all key event handlers to use the new Bubble Tea v2 `tea.KeyMsg` API

### Changed
- **Charmbracelet v2 migration audit**: Added audit document tracking v2 migration status (closes #42) (`docs/charmbracelet-v2-migration-audit.md`)

### Fixed
- **Mouse event forwarding tests for Bubble Tea v2**: Added tests for proper mouse event routing in the form framework (closes #45) (`internal/tui/models/form/form_test.go`)
- **Bubbles table and viewport v2 API usage**: Updated table and viewport usage to match v2 API (closes #46) (`internal/tui/components/vm_table.go`, `internal/tui/models/vm_running_test.go`)
- **View tests for Lip Gloss v2 ANSI rendering**: Updated test assertions to match Lip Gloss v2 ANSI rendering output (`internal/tui/components/breadcrumbs_test.go`, `internal/tui/components/statusbar_test.go`, `internal/tui/components/tabs_test.go`)

## [0.1.22] - 2026-05-31

### Fixed
- **File browser message routing**: DiskAddedMsg and FileSelectedMsg are now properly routed to active forms (VMCreateModel, VMEditModel) when the file browser returns control, ensuring disk selection and file browsing work correctly in the VM creation and edit flows

## [0.1.21] - 2026-05-31

## [0.1.19] - 2026-05-19

### Changed
- **MessageHandler interface extracted**: Moved the `handleMessage` interface from `form/form.go` to `form/types.go` as a public `MessageHandler` interface, with standardized godoc comments on all implementing form models (`internal/tui/models/form/form.go`, `internal/tui/models/form/types.go`, `internal/tui/models/cpu_options_form_handlers.go`, `internal/tui/models/lv_create_form.go`, `internal/tui/models/message_handlers.go`, `internal/tui/models/pci_passthrough_form_handlers.go`, `internal/tui/models/ssh_password_form.go`, `internal/tui/models/vcpu_pinning_form_handlers.go`, `internal/tui/models/vm_form.go`)

## [0.1.18] - 2026-05-18

### Fixed
- **VM running status detection and states**: Added `prelaunch`, `paused`, and `postmigrate` to the set of statuses displayed as `[RUNNING]`, and `finish` as `[STOPPING]`. Uptime now reads from the runner's `StartTime()` instead of a redundant local field. Tightened QMP fallback timeout from 10s to 5s for faster status detection (`internal/tui/models/vm_running.go`, `internal/tui/models/vm_running_test.go`)
- **Registry dispatch deadlock with VMRunning polling**: VMRunning polling and log messages now bypass the view registry dispatch to avoid a deadlock where `cmd()` chaining broke Tick-based polling and blocking channel reads (`internal/tui/models/key_handlers.go`)
- **Start script skip condition**: Fixed `executeStartScript()` to correctly skip when neither builtin nor custom script is configured (`internal/vm/vm_runner.go`)
- **Redundant startTime field removed**: Removed unused `startTime` field initialization from `handleVMSelection()` and cleaned up stale imports (`internal/tui/models/key_handlers.go`, `internal/tui/models/vm_running_test.go`)

## [0.1.17] - 2026-05-17

### Added
- **Section headers in CPU topology form**: Core toggles are now grouped by die with section headers showing die label and L3 cache info for better visual organization (`internal/tui/models/cpu_topology_form.go`, `internal/tui/models/form/form.go`, `internal/tui/models/form/types.go`)
- **`--skip-mount-check` CLI flag**: New flag to bypass mount point warning for testing purposes (`main.go`, `internal/tui/tui.go`, `internal/tui/models/init.go`)

### Changed
- **Status message handling refactored**: Replaced statusMessage field with `statusBar.SetMessage()` calls for consistent status message management across the application (`internal/tui/models/key_handlers.go`, `internal/tui/models/message_handlers.go`, `internal/tui/models/types.go`)
- **StatusBar component integration**: Message handlers now use the statusBar component for unified status display

### Fixed
- **Debug log output isolation**: Debug log output is now properly isolated from the TUI display using `tea.LogToFile` and `tea.WithOutput(os.Stderr)` in debug mode, preventing log output from corrupting the TUI (`internal/tui/tui.go`, `main.go`)
- **Debug log flushing and AltScreen handling**: Fixed debug log flushing to ensure logs are written before TUI exit (`internal/tui/tui.go`)
- **Removed unnecessary log.Sync() calls**: Eliminated redundant log synchronization before TUI starts for cleaner startup
- **List adapter test spacing**: Fixed test assertions to match single-space `'> '` prefix instead of `'>  '` for selected item styling

### Removed
- **`plan.md`**: Completed work documentation file removed

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
[Unreleased]: https://github.com/glemsom/dkvmmanager/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/glemsom/dkvmmanager/compare/v0.1.31...v0.2.0
[0.1.31]: https://github.com/glemsom/dkvmmanager/compare/v0.1.30...v0.1.31
[0.1.30]: https://github.com/glemsom/dkvmmanager/compare/v0.1.29...v0.1.30
[0.1.29]: https://github.com/glemsom/dkvmmanager/compare/v0.1.28...v0.1.29
[0.1.28]: https://github.com/glemsom/dkvmmanager/compare/v0.1.25...v0.1.28
[0.1.25]: https://github.com/glemsom/dkvmmanager/compare/v0.1.24...v0.1.25
[0.1.24]: https://github.com/glemsom/dkvmmanager/compare/v0.1.23...v0.1.24
[0.1.22]: https://github.com/glemsom/dkvmmanager/compare/v0.1.21...v0.1.22
[0.1.21]: https://github.com/glemsom/dkvmmanager/compare/v0.1.19...v0.1.21
[0.1.19]: https://github.com/glemsom/dkvmmanager/compare/v0.1.18...v0.1.19
[0.1.18]: https://github.com/glemsom/dkvmmanager/compare/v0.1.17...v0.1.18
[0.1.17]: https://github.com/glemsom/dkvmmanager/compare/v0.1.16...v0.1.17
[0.1.16]: https://github.com/glemsom/dkvmmanager/compare/v0.1.15...v0.1.16
[0.1.15]: https://github.com/glemsom/dkvmmanager/compare/v0.1.14...v0.1.15
[0.1.14]: https://github.com/glemsom/dkvmmanager/compare/v0.1.13...v0.1.14
[0.1.13]: https://github.com/glemsom/dkvmmanager/compare/v0.1.12...v0.1.13
[0.1.12]: https://github.com/glemsom/dkvmmanager/compare/v0.1.11...v0.1.12
[0.1.11]: https://github.com/glemsom/dkvmmanager/compare/v0.1.10...v0.1.11
[0.1.10]: https://github.com/glemsom/dkvmmanager/compare/v0.1.9...v0.1.10
[0.1.9]: https://github.com/glemsom/dkvmmanager/compare/v0.1.8...v0.1.9
[0.1.8]: https://github.com/glemsom/dkvmmanager/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/glemsom/dkvmmanager/compare/v0.1.5...v0.1.7
[0.1.5]: https://github.com/glemsom/dkvmmanager/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/glemsom/dkvmmanager/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/glemsom/dkvmmanager/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/glemsom/dkvmmanager/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/glemsom/dkvmmanager/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/glemsom/dkvmmanager/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/glemsom/dkvmmanager/tree/v0.0.1
