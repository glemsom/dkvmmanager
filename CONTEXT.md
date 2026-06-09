# DKVM Manager

Terminal UI for KVM/QEMU virtual machine management on a single host.

## Language

**Runner**:
A `VMRunner` instance owns the full lifecycle of one running QEMU process: it starts the process, holds the QMP client, streams log lines, runs start/stop scripts, owns the metrics snapshot, and tears everything down on stop. A fresh runner is created per VM session; the runner is not reused across runs of the same VM.
_Avoid_: controller, driver. (Manager is the multi-VM registry and is a different thing — see "Manager" below.)

**Manager**:
A `vm.Manager` owns the multi-VM registry: listing, lookup, creation, and YAML persistence of VM configurations. It does not own running processes; that is the runner's job.
_Avoid_: runner (different scope — per-process vs. multi-VM), orchestrator.

**QMP client**:
A typed wrapper around the QEMU Monitor Protocol over a Unix socket. The QMP client is the only path for guest-side introspection and control (status, vCPU state, block devices, balloon). It is owned by exactly one runner for the lifetime of one QEMU process and is never shared between VMs.
_Avoid_: "monitor", "QMP socket" (the socket is the wire; the client is the wrapper).

**Persisted log**:
A per-VM file at `<DataFolder>/vms/<id>/qemu.log` written by the runner. It mirrors the runner's internal log channel — QEMU stdout, QEMU stderr, and start/stop script output, all prefixed (e.g. `[stderr] `, `[start] `) and in arrival order — so log history survives view re-entry and VM exit. The persisted log is the source of truth for post-mortem; the in-memory log channel is a live tail.
_Avoid_: "log file" (too generic), "QEMU log" (confuses the channel with the file), "vm log".

**Log subscription**:
The runner's `Subscribe() <-chan string` API, which hands out a fresh buffered channel and lets the caller drain it before receiving new lines. Replaces the older `LogChan() <-chan string` global channel. The drain-on-subscribe semantic is what lets the view recover missed lines from the persisted log without duplicating them.
_Avoid_: "LogChan" (the old name), "log channel" (suggests the old global channel).

**Metrics snapshot**:
A point-in-time value object returned by `runner.Snapshot()` containing both QMP-derived guest metrics (per-vCPU time, per-block I/O counters, balloon size) and host `/proc/<pid>`-derived metrics for the QEMU process (RSS, CPU time). Snapshots are produced on a 2 s cadence, decoupled from the 500 ms status poll.
_Avoid_: "live metrics" (that is the *cadence*; the value object is a snapshot), "stats", "telemetry".

**Status poll**:
The fast (500 ms) QMP-derived read that powers the existing `[RUNNING]` / `[STARTING]` / `[STOPPING]` badge and the vCPU thread-ID list. Distinct from the metrics snapshot: status is cheap, binary-ish, and shared with the view's existing key path; metrics is heavier, numerical, and lives on its own cadence.
_Avoid_: "metrics poll" (status is not metrics), "health check" (carries external-service connotations we don't have).

## Package boundaries

- `internal/vm` is the *data plane* for running VMs: the runner, the QMP client, the metrics snapshot, the persisted log. Anything that talks to QEMU or the host kernel about a running process lives here.
- `internal/tui/models` is the *view plane*: BubbleTea models, key handlers, message dispatch, rendering. The view plane imports `internal/vm`; `internal/vm` does not import the view plane.
- The runner is the only object that crosses the boundary in both directions: it exposes `Snapshot()`, `Subscribe()`, `PID()`, `QMPClient()` for the view, and consumes the QMP client, `/proc`, and the disk for itself.
