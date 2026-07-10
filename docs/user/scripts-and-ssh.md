# Scripts & SSH

Configure start/stop hook scripts and the SSH access password for your VMs.

## Prerequisites

- VM created (see [VM Management](vm-management.md))
- For custom scripts: shell scripts accessible from the host filesystem

> **You should know**: See [How DKVM Manager Works](../explanation/how-dkvm-manager-works.md) for start/stop script concepts.

---

## Start/Stop Script

### Access

Configuration tab → **Edit Start/Stop Script** (index 7).

> **Source**: `internal/tui/models/init.go` → ViewStartStopScript registration; `internal/tui/models/start_stop_script.go` → `NewStartStopScriptModel()`.

### Form layout

The form opens with a mode toggle and (in custom mode) path fields:

| Field | Type | Description |
|-------|------|-------------|
| **Mode** | Toggle | **Builtin** (auto-generate PCI passthrough script) or **Custom** (user-supplied scripts) |
| **Start Script** | Text input | Path to start script (custom mode only); defaults to `/media/dkvmdata/start.sh` |
| **Browse** | Button | Opens file browser to select start script |
| **Stop Script** | Text input | Path to stop script (custom mode only); defaults to `/media/dkvmdata/stop.sh` |
| **Browse** | Button | Opens file browser to select stop script |
| **Save** | Button | Persist configuration to repository |
| **Cancel** | Button | Discard changes, return to Configuration tab |

> **Source**: `internal/tui/models/start_stop_script_form.go` → `BuildPositions()`, `NewStartStopScriptFormModel()`; `internal/tui/models/start_stop_script_form_render.go`.

### Keybindings

Use `Tab`/`Shift+Tab` to navigate fields, `Space`/`Enter` to toggle the Mode or activate buttons, `PgUp`/`PgDown` to scroll, and `ESC` to discard and return.

See [Keybindings](keybindings.md) for the full reference.

> **Source**: `internal/tui/models/start_stop_script_form.go` → `HandleEnter()`, `handleArrowKey()`; `internal/tui/models/start_stop_script_form_render.go` → `renderTogglePosition()`, `renderButtonPosition()`.

### Execution flow

**Builtin mode**:
1. On VM start, `GenerateBuiltinScript()` produces a bash script that binds each configured PCI device to `vfio-pci` using `driver_override`
2. Script is written to `/tmp/dkvm-builtin-<vm-id>.sh`, executed, and deleted after completion
3. On VM stop, `GenerateBuiltinStopScript()` runs (currently a no-op — vfio-pci stays bound for subsequent runs)

**Custom mode**:
1. On VM start, the start script is invoked as `/bin/bash <script> start [device1] [device2] ...`
2. If the start script exits non-zero, QEMU is **not** started and an error is displayed
3. On VM stop, the stop script is invoked as `/bin/bash <script> stop [device1] [device2] ...`
4. Stop script failure is non-fatal — logged but does not block shutdown

**Log output**: Script stdout/stderr is captured and written to the VM log with `[start script stdout]` / `[start script stderr]` and `[stop script stdout]` / `[stop script stderr]` prefixes. In the running view, this appears as `[start]` / `[stop]` prefixed lines.

> **Source**: `internal/vm/vm_runner.go` → `executeStartScript()` (line ~1320), `executeStopScript()` (line ~1445); `internal/vm/script_generator.go` → `GenerateBuiltinScript()`, `GenerateBuiltinStopScript()`.

### Configuration storage

Script settings are stored under the `custom_script` key in the repository:

```yaml
custom_script:
  use_builtin: true
  start_script: ""
  stop_script: ""
```

Loaded by `LoadRunConfigFromRepo()` into `RunConfig.StartStopScript`.

> **Source**: `internal/domain/types.go` → `StartStopScript` struct; `internal/vm/run_config.go` → `LoadRunConfigFromRepo()`.

---

## SSH Password

### Access

Configuration tab → **Set SSH Password** (index 9).

> **Source**: `internal/tui/models/init.go` → ViewSSHPassword registration; `internal/tui/models/ssh_password.go` → `NewSSHPasswordModel()`.

### Form layout

| Field | Type | Description |
|-------|------|-------------|
| **New Password** | Text input (masked) | Enter new password; shown as `*` characters |
| **Confirm Password** | Text input (masked) | Re-enter password for confirmation |
| **Strength indicator** | Display | Bar + label showing password strength: Weak (1), Fair (2-3), Strong (4-5) |
| **Apply** | Button | Validate and apply password change |

> **Source**: `internal/tui/models/ssh_password_form.go` → `buildPositions()`, `RenderPosition()`.

### Keybindings

Use `Tab`/`Shift+Tab` to navigate fields, `Enter`/`Space` on a text field to move to the next field or on **Apply** to validate and apply, `Backspace`/`Delete` for text input, and `ESC` to cancel and return to the Configuration tab.

See [Keybindings](keybindings.md) for the full reference.

> **Source**: `internal/tui/models/ssh_password_form.go` → `handleKey()`; `internal/tui/models/ssh_password_form_validation.go`.

### Validation

Passwords are validated before applying:

| Rule | Error message |
|------|---------------|
| New Password is empty | "Password is required" |
| Confirm Password is empty | "Please confirm the password" |
| Password shorter than 6 characters | "Password must be at least 6 characters" |
| Passwords do not match | "Passwords do not match" |

Failed validation shows inline errors beneath each field and does not proceed.

> **Source**: `internal/tui/models/ssh_password_form_validation.go` → `validate()`.

### Apply mechanism

1. Password and confirmation pass validation
2. `chpasswd` is called with `USER:password` on stdin
3. On success, `lbu commit` persists the change to Alpine Linux diskless storage
4. On failure, an error message is displayed inline below the form
5. In **dry-run mode** (`dryRunMode` flag), `chpasswd` is not executed — a debug log message is emitted instead

> **Source**: `internal/tui/models/ssh_password_form_validation.go` → `apply()`.

### Password strength scoring

Strength is scored 0–5 based on:

| Criteria | Points |
|----------|--------|
| Length ≥ 8 | +1 |
| Length ≥ 10 | +1 |
| Contains lowercase letter | +1 |
| Contains uppercase letter | +1 |
| Contains digit or special character | +1 |

Score maps to labels: **Weak** (≤1, red), **Fair** (2–3, yellow), **Strong** (4–5, green).

> **Source**: `internal/tui/models/ssh_password_form_validation.go` → `passwordStrength()`, `strengthLabel()`.

---

> **Behind the scenes**: See [Architecture](../dev/architecture.md) for model hierarchy, message flow, and form framework details.

---

## See Also

- [VM Management](vm-management.md) — creating and editing VMs
- [Running VMs](running-vms.md) — VM lifecycle and script output in log viewer
- [Hardware Configuration](hardware-config.md) — PCI passthrough device configuration (used by builtin scripts)
- [Example Scripts](../../examples/README.md) — reference scripts for PCI passthrough
- [Keybindings](keybindings.md) — complete keyboard reference
- [Security Considerations](security.md) — SSH password and script security
