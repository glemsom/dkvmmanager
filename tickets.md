# Tickets: Bug fixes from systematic code review

Fixes all 19 bugs identified in `bugs.md`. Grouped by concern, ordered by severity. Each ticket is independent — work the frontier in any order.

## Fix VM runner channel deadlocks

**Severity:** CRITICAL
**Bugs addressed:** 1, 2 (from `bugs.md`)

**What to build:** No more VM runner hard-block when QEMU output floods subscriber/log channels, or when `connectQMP` writes after persist loop exit.

The `forwardViewLine` drop-oldest pattern (`vm_runner.go:280-289`) has a blocking path: when the subscriber channel is full AND no goroutine is reading from it, the inner select takes `default`, then `ch <- line` blocks forever because nobody is receiving. Same pattern in `readOutput` (line 1147-1156) and `readScriptOutput` (line 1180-1189) where `logChan` can fill if `persistLogLoop` is stuck.

After QMP connects, `connectQMP()` does `r.logChan <- "[QMP] Connected..."` (line 1334) and `r.logChan <- fmt.Sprintf(...)` on the error path (line 1336). If `persistLogLoop` has already exited, no reader exists on `logChan` — these writes can fill the buffer (256) and block forever.

**Blocked by:** None — can start immediately

- [x] Fix `forwardViewLine` select pattern: after draining one item, use a non-blocking attempt to write
- [x] Fix `readOutput` and `readScriptOutput` same select-drain-drop pattern
- [x] Guard all `r.logChan <-` writes in `connectQMP` with select-drop or running check
- [x] Add tests proving no deadlock with full channels and no reader

## Fix VM runner process lifecycle concurrency

**Severity:** HIGH (bugs 3-5) + LOW (bug 17)
**Bugs addressed:** 3, 4, 5, 17

**What to build:** No nil-deref panic in `Stop()` when `r.cmd` not yet set. No data race between `ForceStop()`/`Kill()` and `monitorProcess()`/`Wait()`. No race window on `swtpmProcess` between `startTPM` and `cleanupTPM`.

`Stop()` (line 1065-1068) accesses `r.cmd.Process` without nil-checking `r.cmd`. If `Start()` sets `r.running = true` then panics before setting `r.cmd`, or if `ForceStop()` is called between `cmd.Start()` and `r.cmd = cmd`, `r.cmd` is nil → panic.

`ForceStop()` (1103-1115) captures `proc := r.cmd.Process` under lock then unlocks, but `proc.Kill()` can race with `r.cmd.Wait()` in `monitorProcess()` on the same OS process handle.

`startTPM` (782-795) reads `r.swtpmProcess` under lock, but between unlock and `proc.Signal(0)` check, `cleanupTPM()` from another path could have killed it.

**Blocked by:** None — can start immediately

- [ ] Add `r.cmd != nil` check before `r.cmd.Process` access in `Stop()`
- [ ] Introduce a process-lifecycle state machine or `sync.WaitGroup` to serialize `Kill`/`Wait` calls
- [ ] Fix `ForceStop()` to not race with `monitorProcess()` — use an intermediate channel or atomic flag
- [ ] Fix `startTPM` race: re-check process aliveness under lock or use a done channel
- [ ] Add concurrency tests: concurrent Stop/ForceStop, Start+immediate-Stop races

## Fix goroutine & channel lifecycle cleanup

**Severity:** MEDIUM (bugs 7, 8) + LOW (bug 18)
**Bugs addressed:** 7, 8, 18

**What to build:** No goroutine leaks from `qmpWatchdog` or stop-script readers. No double-close panic on `persistLogLoop`.

`qmpWatchdog` (1234-1273) is only started in debug mode. It exits when `!running` or QMP connected. But if `Cleanup()` is called while it's still running, the goroutine continues until its next tick — no explicit cancellation.

`readScriptOutput` goroutines for stop scripts (1536-1537, 1578-1579) are launched with `go r.readScriptOutput(...)` but never awaited. If the stop script is long-running, these goroutines persist after the runner is fully cleaned up.

`closePersistLog` (184-191) uses a select-based idempotency guard. If two callers call it concurrently, both see `default` (not yet closed) and both attempt `close(r.persistQuit)` — second close panics.

**Blocked by:** None — can start immediately

- [ ] Add cancellation context or done-channel check to `qmpWatchdog`
- [ ] Track and `Wait()` on stop-script reader goroutines before cleanup completes
- [ ] Replace select-based guard in `closePersistLog` with `sync.Once`
- [ ] Add tests: concurrent closePersistLog calls, goroutine lifecycle after Cleanup

## Fix /proc/stat parsing robustness

**Severity:** HIGH
**Bug addressed:** 6

**What to build:** `readStatCPUJiffies` correctly finds the closing `)` even when the process comm field contains `)` characters. No parse corruption or panic.

`readStatCPUJiffies` (`proc.go:84-97`) uses `strings.LastIndexByte(string(data), ')')` to find the closing paren of the comm field. If a process name contains `)` (valid in Linux), parsing finds the wrong `)` leading to parse corruption or out-of-bounds panic.

**Blocked by:** None — can start immediately

- [x] Parse `/proc/[pid]/stat` by finding comm field start (`(`) then scanning forward byte-by-byte to the matching `)`
- [x] Or use `strings.IndexByte` after the first `(` to find the first `)` that terminates comm
- [x] Add tests with process names containing `)`, special characters, empty comm field

## Fix TUI event-loop blocking & allocation churn

**Severity:** MEDIUM
**Bugs addressed:** 9, 14

**What to build:** `Update()` handlers no longer block the TUI event loop with synchronous `cmd()` calls. `PadToScreen` reuses buffers instead of allocating per frame.

`vm_create.go:53,60,73` calls `cmd()` synchronously inside `Update()` and feeds the result immediately into another `Update`. This turns async commands into synchronous calls. If the command blocks on I/O (disk scanning), the entire TUI freezes.

`PadToScreen` (`styles/colors.go:340-360`) pads content to full terminal height each `View()` call by creating new strings. With 2 FPS updates from metrics, this creates garbage pressure proportional to terminal size × framerate.

**Blocked by:** None — can start immediately

- [ ] Refactor `Update()` handlers to send the `cmd()` result as a message through bubbletea's messaging system instead of calling synchronously
- [ ] Refactor `PadToScreen` to use a strings.Builder or pre-allocated buffer, reusing across calls
- [ ] Add tests: verify Update returns immediately without blocking I/O; benchmark PadToScreen allocation count

## Fix TUI display bugs

**Severity:** MEDIUM
**Bugs addressed:** 10, 11, 12, 13

**What to build:** Status tick stops when VM is terminal. `[STOPPING]` uses distinct style from `[STARTING]`. Memory display consistent precision. Layout math consistent at terminals below 80 cols.

`VMStatusUpdateMsg` handler (`vm_running.go:332-336`) always re-arms `m.pollStatus()` even when `m.status` is `"stopping"` or `"stopped"`. Extra ticks fire until `VMStoppedMsg` arrives.

Stopping status (`vm_running.go:508`) uses `statusStarting` (warning/yellow) for `"stopping"` / `"finish"`. Should use a distinct style.

Memory display (`vm_running.go:528-534`): `memMB % 1024 == 0` uses "8 GB", otherwise `"%.1f GB"` — 8193 MB → "8.0 GB" but 8192 MB → "8 GB". Inconsistent precision.

`effectiveWidth()` (`view.go:318-323`) clamps to 80, but `renderVMsTabWithWidth()` (line 152) checks `m.windowWidth < 80` (actual, not clamped). If terminal is 60 wide, layout math is inconsistent across tabs.

**Blocked by:** None — can start immediately

- [x] Stop re-arming status tick when status is terminal ("stopping", "stopped")
- [x] Add a distinct style (error/orange) for stopping/finish status rendering
- [x] Normalize memory display: always use "%.1f GB" or a helper that gives clean integer strings for exact GB
- [x] Fix `effectiveWidth()`/`renderVMsTabWithWidth()` to use consistent width value
- [x] Add tests: status tick not re-armed after stop, memory format consistency at boundary values
  - TestVMRunningModelStatusUpdateStopsPollWhenStopping
  - TestVMRunningModelStatusUpdateStopsPollWhenStopped
  - TestVMRunningModelStatusUpdateReArmsPollWhenNonTerminal
  - TestVMRunningModelMemoryFormatConsistent
  - TestVMRunningModelViewStoppingStyle
  - TestVMRunningModelViewNarrowWidth

## Fix edge case panics & test pollution

**Severity:** LOW
**Bugs addressed:** 15, 16, 19

**What to build:** No panic on empty config menu items. No test pollution from global viper state. No nil deref in `FormatError`.

`buildConfigListAdapter` (`init.go:231,240`) assumes `items[0]` and `items[len-1]` exist. If registry has no config menu items, this panics with index out of range.

`config.Load()` (`config.go:61-84`) uses global `viper` singleton for defaults. Tests running concurrently with config loading race on global state.

`FormatError` (`hugepages.go:174-180`) panics on `result.AvailablePages` if `result` is nil. Callers (`vm_runner.go:957`) pass `result` from `Ensure()` which can return nil on error.

**Blocked by:** None — can start immediately

- [x] Guard `buildConfigListAdapter` with `len(items) == 0` check
- [x] Change `config.Load()` to use `viper.New()` local instance instead of global `viper`
- [x] Add nil check for `result` in `FormatError` before accessing fields
- [x] Add tests: empty config registry, concurrent config loading in tests, nil result to FormatError
