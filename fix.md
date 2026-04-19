# Plan: Fix vCPU Topology/Pinning Follow-up Issues

## Goal
Address the review findings so topology-aware vCPU setup is reliable, persisted, test-covered, and debuggable.

---

## 1) Persist `UseHostTopology` in repository (Critical)

### Files
- `internal/vm/repository_topology.go`

### Changes
- In `GetCPUTopology()`:
  - Read `use_host_topology` from config map into `topo.UseHostTopology`.
- In `SaveCPUTopology()`:
  - Include `"use_host_topology": topo.UseHostTopology` in saved data.

### Verification
- Add/update repository tests to assert round-trip persistence for:
  - `enabled`
  - `selected_cpus`
  - `use_host_topology`

---

## 2) Add diagnostics when topology mapping fails (Medium)

### Files
- `internal/vm/vm_runner_config.go`
- (optionally) `internal/vm/vm_runner.go` if logging helper is needed

### Changes
- Replace silent skip in topology mapping loop with explicit warning output.
- Preferred behavior:
  - Record warning with enough context: VM name, selected host CPU ID, error text.
  - Continue processing remaining CPUs (non-fatal).

### Verification
- Unit test for invalid CPU IDs in `SelectedCPUs` when `UseHostTopology=true`:
  - Build args still succeeds.
  - Invalid mapping does not panic.
  - Expected warning signal/path is exercised.

---

## 3) Add missing tests for topology-aware path (High)

### Files
- `internal/vm/cpu_scanner_test.go` (or new focused test file)
- `internal/vm/vm_runner_test.go`
- `internal/vm/repository_topology_test.go` (or existing repository tests)

### Test Cases

#### 3.1 `CPUIndexToTopology` tests
- Returns correct `(dieID, coreID, threadID)` for valid CPU IDs.
- Returns error for unknown CPU ID.

#### 3.2 `buildQEMUArgs` topology-aware tests (`UseHostTopology=true`)
- Includes `-smp <n>,maxcpus=<hostTotal>,sockets=1,dies=<d>,cores=<c>,threads=<t>`.
- Includes one `-device host-x86_64-cpu,...` per valid selected CPU.
- Correctly falls back to flat `-smp` when host topology is missing/invalid or flag is false.

#### 3.3 Repository persistence tests
- Save + load retains `UseHostTopology=true` and `false`.

---

## 4) Cleanup and guardrails (Low)

### Changes
- Ensure `debug.log` is not committed.
- Optionally add/update ignore rule if needed.

### Verification
- `git status` clean of unintended artifacts.

---

## 5) Validation sequence

1. Run targeted tests for modified packages.
2. Run full test suite.
3. (Optional) Race run for touched packages.

> Per project workflow, test/build execution should be delegated to the `tester` subagent.

---

## 6) Definition of done

- `UseHostTopology` persists across save/load.
- Topology mapping failures are visible via warnings (no silent drops).
- New tests cover CPU mapping, QEMU args topology path, and repository persistence.
- No stray files included in commit.
- Existing and new tests pass.
