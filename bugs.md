# Bug Report

Found by systematic code review. Items grouped by severity.

---

## CRITICAL — Deadlock Potential

### 1. `forwardViewLine` drop-oldest pattern can block forever

**File:** `internal/vm/vm_runner.go:280-289`

```go
select {
case ch <- line:
default:
    select {
    case <-ch:
    default:
    }
    ch <- line   // BLOCKS if no concurrent reader
}
```

When subscriber channel is full AND no goroutine is currently blocked reading it, the inner select takes `default` (no reader ready), then `ch <- line` blocks indefinitely because the buffer is full and nobody is receiving.

**Same bug in `readOutput` (line 1147-1156) and `readScriptOutput` (line 1180-1189)** — if `logChan` buffers fill and `persistLogLoop` is stuck (see above), they all deadlock.

**Deadlock chain:** QEMU fast output → readOutput blocks on logChan → persistLogLoop blocked on subscriber channel → subscriber blocked processing event → all three writers stuck.

**Fix:** Use non-blocking send after drain:

```go
select {
case ch <- line:
default:
    select {
    case <-ch:
        select {
        case ch <- line:
        default:
        }
    default:
    }
}
```

### 2. `connectQMP()` writes to `logChan` after persist loop exit

**File:** `internal/vm/vm_runner.go:1334`

After QMP connects, `connectQMP()` does `r.logChan <- "[QMP] Connected..."`. If `persistLogLoop` has already exited (via `closePersistLog()` called from `monitorProcess()`), there's no reader on `logChan`. A single write to buffered chan (256) is safe, but subsequent writes from `ApplyVCPUPinning` error path (line 1336) fill the buffer, then block forever.

**Fix:** Use same select-drop pattern or make goroutines check `r.running` before writing.

---

## HIGH — Data Race / Panic

### 3. `Stop()` nil-dereference on `r.cmd.Process`

**File:** `internal/vm/vm_runner.go:1065-1068`

```go
r.mu.Lock()
client := r.qmpClient
running := r.running
r.cmdProcess = r.cmd.Process   // r.cmd could be nil if Start() panics mid-way
```

If `Start()` sets `r.running = true` then panics before setting `r.cmd`, or if `ForceStop()` is called between `cmd.Start()` and `r.cmd = cmd` assignment (lines 1009-1038), `r.cmd` is nil → panic.

`ForceStop()` (line 1105) checks `r.cmd == nil` under lock and returns early. But `Stop()` does not.

**Fix:** Check `r.cmd != nil` before accessing `.Process`.

### 4. `ForceStop()` `r.cmd.Process` captured after unlock

**File:** `internal/vm/vm_runner.go:1103-1115`

```go
r.mu.Lock()
if !r.running || r.cmd == nil || r.cmd.Process == nil {
    r.mu.Unlock()
    return ...
}
proc := r.cmd.Process
r.mu.Unlock()
// proc.Kill() races with concurrent r.cmd = nil in monitorProcess
```

`proc` is a value copy of `*os.Process`, so Kill() won't crash, but `proc.Wait()` in `monitorProcess` and `proc.Kill()` here race on the same OS process handle.

### 5. `monitorProcess()` wraps `proc.Wait()` without concurrent-safety

When `Stop()`/`ForceStop()` calls `proc.Kill()` while `monitorProcess()` is in `r.cmd.Wait()`, there's a race between `Wait()` completion and `r.running = false` / `r.qmpClient = nil` inside `monitorProcess()`. The mutex only protects field assignment, not the Wait/Kill order.

### 6. `readStatCPUJiffies` fragile parsing

**File:** `internal/vm/proc.go:84-97`

Uses `strings.LastIndexByte(string(data), ')')` to find closing paren of comm field. If a process name contains `)` (valid in Linux, e.g. `qemu-system-x86_64) with funny args`), parsing finds the wrong `)` → parse corruption or panic on out-of-bounds.

Comm field can contain any byte except `\0` and `\n`. A `)` inside comm is unusual but not impossible.

---

## MEDIUM — Goroutine / Resource Leaks

### 7. `qmpWatchdog` goroutine leak

**File:** `internal/vm/vm_runner.go:1234-1273`

Only started in debug mode. Exits when `!running` or QMP connected. But if VM stops right after QMP connects, watchdog exits one tick later (≤2s). Minor but unnecessary goroutine with no cleanup on `Cleanup()`.

### 8. Stop-script goroutines fire-and-forget

**File:** `internal/vm/vm_runner.go:1536-1537, 1578-1579`

`readScriptOutput` goroutines for stop scripts are launched with `go r.readScriptOutput(...)` but never awaited. If the stop script is long-running, these goroutines persist after the runner is fully cleaned up. They still write to `r.logChan` which may have no reader → eventual block or drop.

### 9. `vm_create.go` sync `cmd()` call in Update

**File:** `internal/tui/models/vm_create.go:53,60,73`

`cmd()` is called synchronously inside `Update()` to produce a message, then fed immediately into another `Update`. This turns async commands into synchronous calls, blocking the event loop for the duration of the command (e.g., disk scanning). If the command blocks on I/O, the entire TUI freezes.

Same pattern in several Update handlers.

---

## MEDIUM — Visual/TUI Bugs

### 10. Status tick loop continues after VM stopped

**File:** `internal/tui/models/vm_running.go:332-336`

`VMStatusUpdateMsg` handler always returns `m.pollStatus()` even when `m.status` is `"stopping"` or `"stopped"` (the status value is not updated but the tick IS re-armed). Status ticks keep firing until `VMStoppedMsg` arrives and breaks the chain. Extra ticks are harmless but wasteful.

True fix: stop re-arming when status is terminal.

### 11. Stopping status styled as STARTING

**File:** `internal/tui/models/vm_running.go:508`

```go
case "stopping", "finish":
    statusStr = statusStarting.Render("[STOPPING]")
```

Stopping uses `statusStarting` (warning/yellow color). Should use a distinct stopping style (maybe error/orange) for visual clarity.

### 12. Memory display format inconsistent

**File:** `internal/tui/models/vm_running.go:528-534`

```go
memGB := memMB / 1024
if memMB%1024 == 0 {
    // "8 GB"
} else {
    // fmt.Sprintf("%.1f GB", float64(memMB)/1024.0)
}
```

8193 MB → `8193/1024 = 8.0009...` → displays "8.0 GB". 8192 MB → "8 GB". Inconsistent decimal precision. Minor but looks sloppy.

### 13. `effectiveWidth()` masks terminal resize below 80 cols

**File:** `internal/tui/models/view.go:318-323`

`effectiveWidth()` clamps to 80, but `renderVMsTabWithWidth()` (line 152) checks `m.windowWidth < 80` (the actual value, not clamped). If terminal is 60 wide, render uses `listWidth = 36`, but `effectiveWidth()` reports 80 for layout purposes. Inconsistent layout math across tabs.

### 14. `PadToScreen` creates massive string allocations per frame

**File:** `internal/tui/styles/colors.go:340-360`

Each `View()` call pads content to full terminal height (e.g. 25+ lines) by creating new strings. With 2 FPS updates from metrics, this creates garbage pressure proportional to terminal size × framerate.

---

## LOW — Edge Cases & Cleanup

### 15. `buildConfigListAdapter` assumes items[0] and items[len-1] exist

**File:** `internal/tui/models/init.go:231,240`

If registry has no config menu items, `items[0]` panics (index out of range). Currently always ≥1 item but fragile to future changes.

### 16. Global viper in config.go pollutes test isolation

**File:** `internal/config/config.go:61-84`

Uses global `viper` singleton for defaults. Tests running concurrently with config loading race on global state. The repository correctly uses `viper.New()`, but config.Load() doesn't.

### 17. `startTPM` race window on `swtpmProcess`

**File:** `internal/vm/vm_runner.go:782-795`

After socket appears, `startTPM` verifies swtpm is alive using `proc.Signal(0)` (captured under lock). But between unlock and this check, `cleanupTPM()` from another path could have killed it. The check uses stale `proc`.

### 18. `persistLogLoop` closes `persistQuit` race in `closePersistLog`

**File:** `internal/vm/vm_runner.go:184-191`

Double-close guard uses select on channel, but `persistLogLoop` might read the close signal before `closePersistLog` returns. The goroutine then flushes and closes subscriber channels. If `closePersistLog` then calls `persistWg.Wait()`, the goroutine is already done — correct, but the select-based close guard could race if two callers call `closePersistLog` concurrently. Both would see `default` (not yet closed) and both would attempt `close(r.persistQuit)` — second close panics.

**Fix:** Use `sync.Once` for close.

### 19. `FormatError` on `result` nil

**File:** `internal/hugepages/hugepages.go:174-180`

If `result` is nil, `FormatError` panics on `result.AvailablePages`. Callers (vm_runner.go:957) pass `result` from `Ensure()` which can return nil on error. Although Start() checks error first, a helper or future caller could crash.

---

## Summary

| Severity | Count | Key issues |
|----------|-------|------------|
| CRITICAL | 2     | `forwardViewLine` deadlock, `connectQMP` blocked write after persist exit |
| HIGH     | 4     | `Stop()` nil deref, `ForceStop()` race, `monitorProcess`/`Wait` race, fragile `/proc/stat` parse |
| MEDIUM   | 5     | Goroutine leaks, sync cmd() in Update, status tick leak, stopping style, memory format |
| LOW      | 4     | Slice bounds, global viper, `swtpmProcess` race, double-close risk |
