# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- VG dropdown in Create Logical Volume form - press Enter on Volume Group to open selection list
- Loading spinner state in LV create form
- VG dropdown open rendering in LV create form
  - Added `internal/hugepages/` module for hugepages allocation
  - Added hugepages validation in VM startup (`internal/vm/vm_runner.go`)
- Reboot system functionality from Power menu in TUI
  - Added `RebootMsg` type in `internal/tui/models/types.go`
  - Added `runReboot` handler in `internal/tui/models/debug.go`
  - Added key handler in `internal/tui/models/key_handlers.go`
- Power off system functionality from Power menu in TUI
  - Added `PowerOffMsg` type in `internal/tui/models/types.go`
  - Added `runPowerOff` handler in `internal/tui/models/debug.go`
  - Added key handler in `internal/tui/models/key_handlers.go`

### Changed
-

### Deprecated
-

### Removed
-

### Fixed
-

### Security
-

## [0.0.1] - 2026-04-12

### Added
- Initial release