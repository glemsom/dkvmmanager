# Running VMs

Start, monitor, and stop virtual machines. The running view provides real-time log output, status updates, and performance metrics for an active VM.

## Prerequisites

- VM created (see [VM Management](vm-management.md))
- Only one VM runs at a time

## Concepts

- **Runner** (`VMRunner`): Owns the QEMU process lifecycle — start, monitor, stop. Created with the VM config, host config, and aggregated `RunConfig`. Runs QEMU as a child process, manages QMP socket connection, dispatches log output to subscribers, and persists logs to disk.

- **QMP client**: Connects to QEMU's QMP (QEMU Machine Protocol) Unix socket. Queries VM status (`query-status`), vCPU thread info (`query-cpus-fast`), block device stats (`query-blockstats`), and balloon size (`query-balloon`).

- **Persisted log**: All QEMU stdout/stderr, plus start/stop script output, is written to `<data>/vms/<id>/qemu.log`. Log is available for post-mortem review even after the VM stops.

- **Metrics snapshot**: CPU (per-vCPU and host process), memory (RSS), disk I/O rates, and balloon size. Refreshed every 2 seconds. Delta math converts raw cumulative counters to per-second rates.

- **Status poll**: VM binary status (`running`, `paused`, `stopped`, etc.) and vCPU thread IDs. Polled every 500ms via QMP. Falls back to process-is-alive check if QMP is unavailable for >5s.

- **Start/stop scripts**: Run before QEMU starts (start script) and after QEMU exits (stop script). Output prefixed with `[start]` / `[stop]` in the log.

> **Source**: `internal/vm/vm_runner.go` → `VMRunner`; `internal/vm/qmp_client.go` → `QMPClientInterface`; `internal/vm/metrics.go` → `Metrics`, `Snapshot()`; `internal/vm/proc.go` → `/proc` readers.

---

## Starting a VM

### Access

1. Press `Tab` to switch to the **VMs** tab
2. Use `↑/↓` or `j/k` to highlight a VM in the list
3. Press `Enter` or `Space` to start it

> **Source**: `internal/tui/models/key_handlers.go` → `handleVMSelection()`.

### Startup sequence

1. **VM list check**: Verifies no other VM is running. Shows status message if one is already active.
2. **Config loading**: Loads VM from repository, loads aggregated `RunConfig` (CPU topology, pinning, PCI/USB passthrough, CPU options).
3. **Runner creation**: `vm.NewVMRunner(vmObj, cfg, runCfg)` — builds the QEMU command line.
4. **View transition**: Switches to `ViewVMRunning` immediately — user sees `[STARTING]` badge while QEMU boots.
5. **Async start**: `runner.Start()` runs in a goroutine to avoid blocking the TUI event loop.
6. **Success**: `VMStartedMsg` arrives → log subscription begins, status/metrics polling starts.
7. **Failure**: `VMStartErrorMsg` arrives → returns to main menu with error message.

> **Source**: `internal/tui/models/key_handlers.go` → `handleVMSelection()`; `internal/tui/models/vm_running.go` → `startVMCommand()`, `VMStartedMsg`, `VMStartErrorMsg`.

### Start scripts

If a start script is configured (via [Scripts & SSH](scripts-and-ssh.md)), it runs before QEMU. Script output appears in the log with `[start]` prefix. If the script fails (non-zero exit), QEMU is not started and an error is shown.

---

## Running View Layout

The running view has two sections:

### Info Panel (top)

Displays VM metadata and live metrics:

| Line | Content |
|------|---------|
| **VM + Status** | VM name, status badge (`[RUNNING]`/`[STARTING]`/`[STOPPING]`/`[STOPPED]`), uptime |
| **Resources** | Memory (GB), vCPU count, pinning status |
| **PCI** | Passthrough device addresses (if configured) |
| **USB** | Passthrough device bus IDs (if configured) |
| **TPM** | "enabled" (if TPM is on) |
| **vCPU%** | Per-vCPU utilization and aggregate total (2s cadence) |
| **Host** | QEMU process CPU% and RSS (2s cadence) |
| **disk** | Per-disk read/write rates in B/s and IOPS (2s cadence) |
| **Balloon** | Current balloon size (if driver present) |

Lines only appear when their data is available — zero values for cold snapshots are hidden.

### Log Viewport (bottom)

Scrollable view of QEMU's stdout/stderr output:

- Lines prefixed with `[stdout]`, `[stderr]`, `[start]`, `[stop]`
- Auto-follows tail (scrolls to bottom on new output)
- Ring buffer holds last 500 lines in memory
- Seeded from persisted log file on startup, then subscribes to live stream
- Deduplication prevents overlap between seeded and live lines

> **Source**: `internal/tui/models/vm_running.go` → `renderInfoPanel()`, `renderLogContent()`, `seedAndSubscribe()`, `calculateInfoHeight()`.

---

## Log Viewer Details

### Log lifecycle

1. **Seeding**: On `VMStartedMsg`, `seedAndSubscribe()` reads the last 500 lines from the persisted `qemu.log` file and sends them as `LogSeedMsg`.
2. **Subscription**: After seeding, subscribes to the runner's live log channel via `runner.Subscribe()`. The staging buffer is drained atomically — no lines are lost.
3. **Live streaming**: Each `VMLogMsg` appends a line to the ring buffer. After each line, `waitForLog()` is re-issued to read the next.
4. **Dedup**: Last 20 lines in the buffer are checked for duplicates before appending (prevents overlap at the seed/live boundary).
5. **Persistence**: All lines are simultaneously written to `qemu.log` on disk by the runner's persistence goroutine.

### Log prefixes

| Prefix | Source |
|--------|--------|
| `[stdout]` | QEMU standard output |
| `[stderr]` | QEMU standard error |
| `[start]` | Start script output |
| `[stop]` | Stop script output |

> **Source**: `internal/vm/vm_runner.go` → `Subscribe()`, `RecentLog()`, persist goroutine.

---

## Status Badge

The status badge in the info panel shows one of four states:

| Badge | Color | Meaning |
|-------|-------|---------|
| `[STARTING]` | Yellow/Warning | QEMU is booting, QMP not connected yet |
| `[RUNNING]` | Green/Success | QMP reports status as `running`, `paused`, `postmigrate`, or `prelaunch` |
| `[STOPPING]` | Yellow/Warning | User pressed `q` to stop; shutdown sent via QMP |
| `[STOPPED]` | Red/Error | VM has exited |

Status is polled every 500ms via `pollStatus()`. If QMP is unavailable but the process is alive for >5 seconds, status falls back to `running` (handles slow QMP socket creation).

> **Source**: `internal/tui/models/vm_running.go` → `pollStatus()`, `initialStatus()`, `renderInfoPanel()`.

---

## Metrics Display

Metrics are updated every 2 seconds independently of the 500ms status poll.

### Per-vCPU utilization

```
vCPU%: #0: 12.5%  #1: 3.2%  #2: 8.1%  total: 23.8%
```

CPU time for each vCPU thread is read from `/proc/<qemu-pid>/task/<tid>/stat` (utime + stime fields). Delta against the previous snapshot gives utilization as a percentage. Aggregate total is the sum of all per-vCPU percentages.

### Host process metrics

```
Host: CPU 24.1%  RSS 1.2 GiB
```

- **CPU**: QEMU process CPU% from `/proc/<pid>/stat` (utime + stime jiffies, delta math).
- **RSS**: Resident memory from `/proc/<pid>/status` (VmRSS field). Shown in IEC binary units (KiB, MiB, GiB).

Hidden when values are zero (cold snapshot).

### Per-disk I/O

```
disk drive0: r: 1.2 MiB/s · 345 IOPS  w: 512 KiB/s · 120 IOPS
```

- Raw cumulative r/w bytes and operation counts from QMP `query-blockstats`
- Delta math converts to per-second rates (B/s and IOPS)
- One line per block device, rendered even when idle (zero rates)

### Balloon

```
Balloon: 512 MiB
```

Current balloon size from QMP `query-balloon`. Hidden when zero (no balloon driver — a normal guest configuration).

> **Source**: `internal/vm/metrics.go` → `Metrics`, `VCPUStat`, `BlockStat`; `internal/vm/vm_runner.go` → `Snapshot()`; `internal/vm/proc.go` → `readThreadCPUTime()`, `readProcessCPUJiffies()`, `readProcessRSS()`; `internal/tui/models/vm_running.go` → `pollMetrics()`.

---

## Stopping a VM

### Graceful stop

Press `q` in the running view:

1. Runner calls `runner.Stop()` — sends `system_powerdown` via QMP
2. Status changes to `[STOPPING]`
3. QEMU performs guest shutdown
4. Stop script executes (output prefixed `[stop]` in log)
5. Runner's `Done()` channel closes
6. `VMStoppedMsg` arrives → status changes to `[STOPPED]`
7. View returns to main menu automatically, status bar shows reason

### Force kill

Press `Ctrl+C` in the running view:

1. Runner calls `runner.ForceStop()` — sends SIGKILL to QEMU process
2. Stop script still executes
3. Same cleanup sequence as graceful stop

### Post-stop

After VM stops:
- Persisted log file remains for review
- VM list refreshes (status bar updated)
- Another VM can be started immediately

> **Source**: `internal/tui/models/vm_running.go` → `handleKeyPress()` (q / ctrl+c cases), `waitForVMExit()`; `internal/vm/vm_runner.go` → `Stop()`, `ForceStop()`.

---

## ESC Behavior (Safety)

ESC does **NOT** leave the running view while a VM is active. This prevents accidentally returning to the menu while a VM runs in the background.

| State | ESC behavior |
|-------|-------------|
| VM running | Status message: "VM is still running. Press 'q' to stop it." |
| VM stopped | Status message: "Press 'q' to exit the VM view." |

Press `q` to intentionally stop the VM, or wait for the VM to stop naturally.

> **Source**: `internal/tui/models/key_handlers.go` → `update()` ESC handling for `ViewVMRunning`.

---

## Keybindings

- Press `q` to stop a running VM gracefully, or to exit the view when the VM has stopped.
- Press `Ctrl+C` to force-kill the QEMU process.
- Use `↑/↓` / `j/k`, `PgUp`/`PgDn`, or `Home`/`End` to scroll the log viewport.
- `ESC` shows a reminder — it does not leave the running view while a VM is active.

See [Keybindings](keybindings.md) for the full reference.

> **Source**: `internal/tui/models/vm_running.go` → `handleKeyPress()`; viewport delegates to `charm.land/bubbles/v2/viewport`.

---

> **Behind the scenes**: See [Architecture](../dev/architecture.md) for VM startup sequence, polling cadence, and message routing details.

---

## See Also

- [VM Management](vm-management.md) — creating VMs
- [Scripts & SSH](scripts-and-ssh.md) — start/stop scripts
- [Hardware Configuration](hardware-config.md) — CPU pinning, PCI/USB passthrough
- [Keybindings](keybindings.md) — complete keyboard reference
