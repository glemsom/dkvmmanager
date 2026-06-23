# CLI Flags Reference

Command-line flags for DKVM Manager.

All flags are defined in `main.go` and parsed by Go's `flag` package.

---

## Flag Table

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-debug` | `bool` | `false` | Enable debug mode with verbose logging to `debug.log` |
| `-dry-run` | `bool` | `false` | Build QEMU command but don't execute it |
| `-test <scenario>` | `string` | `""` | Run a test scenario and exit |
| `-skip-mount-check` | `bool` | `false` | Skip `/media/dkvmdata` mount point check |

> **Source:** `main.go` lines 18–26

---

## `-debug`

Enables verbose debug output. When set:

- `tea.LogToFile()` redirects all `log.*` output to `debug.log` (tries CWD, then HOME, then `/tmp`)
- BubbleTea's **AltScreen is disabled** — the TUI renders on stderr so debug log output is visible without switching screen buffers
- All log output is suppressed from the terminal to avoid TUI corruption

```bash
./dkvmmanager -debug
```

> **Source:** `main.go` → `setupDebugLog()`; `internal/tui/tui.go` → `Run()`

---

## `-dry-run`

Builds the QEMU command-line for a VM but does not execute it. The command is written to the VM log. Useful for:

- Verifying QEMU arguments before running
- Debugging PCI passthrough device addresses
- Testing script execution flow without launching QEMU

```bash
./dkvmmanager -dry-run
```

> **Source:** `internal/vm/vm_runner.go` — checked at VM start; `internal/tui/models/init.go` → `SetDryRunMode()`

---

## `-test <scenario>`

Runs a specific test scenario and exits. Available scenarios:

| Scenario | What it does |
|----------|-------------|
| `main_menu` | Launches the main menu TUI for manual testing |
| `vm_create` | Opens the VM creation form directly |

```bash
./dkvmmanager -test main_menu
./dkvmmanager -test vm_create
```

> **Source:** `internal/tui/tui.go` → `Run()` — scenario dispatch

---

## `-skip-mount-check`

Skips the startup check that verifies `/media/dkvmdata` is a real mount point. Use for:

- Testing on systems without the DKVM data volume
- Development environments
- Running from a plain directory

```bash
./dkvmmanager -skip-mount-check
```

A warning is still logged to `debug.log` but no modal appears.

> **Source:** `internal/tui/models/mount_point_warning.go` → `isMountPoint()`

---

## See Also

- [Setup & Prerequisites](../user/setup.md) — system requirements and first launch
- [Troubleshooting](../user/troubleshooting.md) — common issues
- [User Guide Index](../user/README.md) — all documentation
