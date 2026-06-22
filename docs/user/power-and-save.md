# Power & Save

Power off, reboot the system, and persist configuration changes (LBU commit). All operations run asynchronously — the TUI remains responsive while commands execute.

## Prerequisites

- [Setup completed](setup.md) — `/media/dkvmdata` mounted, KVM configured
- `lbu` (Local Backup Utility) installed for saving configuration changes

## Concepts

- **LBU (Local Backup Utility)**: Persists configuration changes to the writable overlay. Without `lbu commit`, changes made in the Configuration tab are lost on reboot or power off.
- **Power off**: Graceful system shutdown via `/sbin/poweroff`. Terminates all running VMs and shuts down the host.
- **Reboot**: System restart via `/sbin/reboot`. Same termination behavior as power off.

> **Source**: `internal/tui/models/debug.go` → `runLBUCommit()`, `runPowerOff()`, `runReboot()`

---

## Power Tab

### Access

Press `Tab` to switch to the **Power** tab. Two options are available:

| Index | Item | Action |
|-------|------|--------|
| 0 | Power off system | Shut down the host |
| 1 | Reboot system | Restart the host |

> **Source**: `internal/tui/models/init.go` → `buildPowerListAdapter()`

### Power Off System

- **Access**: Power tab → **Power off system** (index 0) → `Enter`
- Executes `/sbin/poweroff` asynchronously
- Status bar shows result: "Power off: Power off initiated" on success, or "Power off failed: …" on error
- All running VMs are terminated

> **Source**: `internal/tui/models/key_handlers.go` → `handlePowerSelection()` (case 0); `internal/tui/models/debug.go` → `runPowerOff()`; `internal/tui/models/message_handlers.go` → `HandlePowerOffMsg()`

### Reboot System

- **Access**: Power tab → **Reboot system** (index 1) → `Enter`
- Executes `/sbin/reboot` asynchronously
- Status bar shows result: "Reboot: Reboot initiated" on success, or "Reboot failed: …" on error
- All running VMs are terminated

> **Source**: `internal/tui/models/key_handlers.go` → `handlePowerSelection()` (case 1); `internal/tui/models/debug.go` → `runReboot()`; `internal/tui/models/message_handlers.go` → `HandleRebootMsg()`

### Dry-Run Mode

When dry-run mode is enabled (`--dry-run` flag), power and reboot commands are not executed. The status bar shows a simulation message (e.g., "Would execute: reboot").

> **Source**: `internal/tui/models/debug.go` → `dryRunMode` checks

---

## Save Changes (LBU Commit)

### Access

Configuration tab → scroll to bottom → **Save changes** (always the last item in the Configuration menu).

> **Source**: `internal/tui/models/init.go` → `buildConfigListAdapter()` (last item); `internal/tui/models/key_handlers.go` → `handleConfigMenuSelection()` (last item check)

### Behavior

- Executes `lbu commit` asynchronously
- Status bar shows result: "LBU commit: …" with stdout on success, or "LBU commit failed: …" on error
- Commits **all** changes made since last reboot or previous `lbu commit`:
  - VM configurations (created, edited, deleted)
  - CPU topology, vCPU pinning, PCI/USB passthrough settings
  - Start/stop scripts, SSH password, CPU options
  - LVM logical volume creation
- Without save, all changes are lost on exit or power operation

> **Source**: `internal/tui/models/debug.go` → `runLBUCommit()`; `internal/tui/models/message_handlers.go` → `HandleLBUCommitMsg()`

### When to Save

- After creating, editing, or deleting a VM
- After modifying any hardware configuration
- After setting start/stop scripts or SSH password
- Before powering off or rebooting
- Before quitting the application (if changes should persist)

---

## Warnings

- **Power off / reboot will terminate running VMs** — ensure VMs are stopped first via [Running VMs](running-vms.md)
- **Always save changes before exit or power operation** — unsaved changes are lost
- **"Save changes" only appears in the Configuration tab** — not in VMs or Power tabs
- Power operations execute system-level commands (`/sbin/poweroff`, `/sbin/reboot`) — they do not go through the VM lifecycle; VMs are terminated by the host shutdown, not gracefully stopped

---

## Architecture Notes

### Tab Navigation

The Power tab is one of three top-level tabs:

| Tab | Purpose |
|-----|---------|
| VMs | List and start VMs |
| Configuration | VM and hardware config, scripts, SSH, storage, save |
| Power | Power off and reboot |

`Tab` cycles through them; each has its own list model (`powerList`, `configList`, `menuList`).

> **Source**: `internal/tui/models/init.go` → `NewMainModelWithConfig()`; `internal/tui/components/tab_model.go`

### Message Flow

```
User selects Power Off / Reboot / Save changes
  → handleMenuSelection() → handlePowerSelection() / handleConfigMenuSelection()
    → runPowerOff() / runReboot() / runLBUCommit()
      → async command execution (exec.Command)
        → PowerOffMsg / RebootMsg / LBUCommitMsg
          → HandlePowerOffMsg / HandleRebootMsg / HandleLBUCommitMsg
            → statusBar.SetMessage()
```

All three operations return `tea.Model, tea.Cmd` with no view change — the user stays on the current screen. The status bar provides feedback when the async message arrives.

> **Source**: `internal/tui/models/key_handlers.go` → `handleMenuSelection()`; `internal/tui/models/message_handlers.go` → handler registry

---

## See Also

- [VM Management](vm-management.md) — creating and editing VMs (changes need saving)
- [Running VMs](running-vms.md) — stopping VMs before power operations
- [Hardware Configuration](hardware-config.md) — CPU, PCI/USB settings (changes need saving)
- [Scripts & SSH](scripts-and-ssh.md) — hook scripts and SSH password (changes need saving)
- [Storage](storage.md) — LVM volume creation (changes need saving)
- [Keybindings](keybindings.md) — complete keyboard reference
