# Runner owns the running-VM data plane

Status: accepted

For the live-metrics and persisted-log features, the runner is the single owner of three things: the metrics snapshot (QMP + `/proc/<pid>`), the on-disk `qemu.log`, and the log subscription API. The view is a pure renderer; the QMP client is owned by one runner for the life of one QEMU process. Status polling and metrics polling run on decoupled cadences (500 ms and 2 s) behind two independent `tea.Tick` commands.

## Considered Options

- **View-driven metrics, client-direct (M.1).** `VMRunningModel` calls `client.QueryCPUs()` / `QueryBlockStats()` / `QueryBalloon()` directly, holds the previous sample in the view, and computes deltas. Rejected because: the view has no mutex of its own, the `/proc` reader's error cases (`pid == 0`, `os.ErrNotExist` on early/late lifecycle) belong with the runner, and the (C.3)-shaped future would have to *move* this state back into the runner — forward-incompatible.
- **One tick, many queries, serialised (C.1).** Keep the existing 500 ms tick and just add `query-cpus` / `query-blockstats` / `query-balloon` to it. Rejected because: each query takes 50–150 ms on a busy VM; a single tick would let log-stream responsiveness degrade with VM load.
- **Per-stream persisted log files (b.2).** Split into `qemu-stdout.log`, `qemu-stderr.log`, `start.log`, `stop.log`. Rejected because: the user pain is "I lost the log"; splitting trades that for "which of the four files do I open?". The prefix scheme the runner already uses disambiguates streams inside one file.
- **Structured JSONL persisted log (b.3).** `qemu.log.jsonl` with `{"ts", "stream", "line"}` records. Deferred — the line model stays the same so we can migrate later, but YAGNI for v1.
- **JSON-line metric snapshots (T.3).** Define the `Metrics` type twice, once for wire and once for display. Rejected — verbose, no benefit while we have one consumer.
- **View-owned log persistence.** The TUI writes the file as a side effect of `VMLogMsg`. Rejected because: the file then exists only when the TUI was running, which is exactly the case we want to fix (post-mortem after the TUI has exited).

## Consequences

- **API change**: `runner.LogChan() <-chan string` becomes `runner.Subscribe() <-chan string`. The view must `Subscribe()` (which hands out a fresh buffered channel and drains any pre-existing buffer the runner may have staged) and then iterate. Only the TUI consumes this; the breaking change is small and well-contained.
- **Forward compatibility with (C.3)**: when we eventually move from view-driven metrics cadence to a runner-driven background poller, only the *implementation* of `Snapshot()` changes. The view's call site is untouched. Choosing (M.2) today is what makes this safe.
- **Concurrency discipline**: the 6 existing reader goroutines (QEMU stdout, QEMU stderr, start-stdout, start-stderr, stop-stdout, stop-stderr) all send to a single internal `chan string` consumed by a dedicated flusher goroutine that writes the persisted log with a `*bufio.Writer`. A slow disk must never block a QEMU pipe, so the file write is intentionally one level removed from the readers.
- **Deadlock prevention**: the v0.1.18 fix (VMRunning polling and log messages bypass the view registry dispatch) extends. The new `pollMetrics()` tick must also bypass the registry and feed `VMMetricsUpdateMsg` directly to `VMRunningModel.Update`.
- **PID accessor**: `VMRunner` gains a `PID() int` accessor (one-liner, under the existing `r.mu`) so the `/proc` reader can locate the QEMU process.
- **New QMP methods**: `QueryCPUs()` (distinct from the existing `QueryCPUsFast()` — needs CPU time in ns for delta math), `QueryBlockStats()`, `QueryBalloon()`. All typed. All serialised through the existing client mutex.
- **New value type**: `internal/vm/metrics.go` with the `Metrics` struct. Imported by the view for display; the view is otherwise free of QMP semantics.
