# How DKVM Manager Works

This document explains the architecture and data flow of DKVM Manager from an
end-user perspective. It answers *why* things work the way they do, not just
*how* to use them.

## Why a dedicated host architecture?

DKVM Manager is designed to run on a **dedicated virtualization host** — a
physical machine whose sole purpose is running VMs. This is different from
running QEMU on a laptop or workstation where the hypervisor shares resources
with desktop applications.

The host runs Alpine Linux from a USB stick in **diskless mode** (see
[Understanding LBU](understanding-lbu.md)). The operating system is entirely
in RAM, meaning:

- The OS is stateless — a reboot returns to a clean known state.
- All persistent data (VM configs, disk images, logs) lives on `/media/dkvmdata`.
- There is no package manager state, no configuration drift, no journald logs
  to fill up the root filesystem.

## The two-layer architecture

DKVM Manager separates two concerns that are often conflated in other
virtualization tools:

### 1. Application layer (`~/.dkvmmanager.yaml`)

This file tells DKVM *where* things are on the host. It contains paths to the
QEMU binary, OVMF firmware, the data folder, the GRUB config file, and a few
tuning parameters like reserved memory for hugepage calculation.

**You almost never need to edit this.** The defaults are correct for a standard
DKVM host. See [Setup: Application Configuration](../user/setup.md) for details.

### 2. VM repository layer (`/media/dkvmdata/dkvmmanager/config.yaml`)

This file contains every VM: its name, CPU topology, RAM, disk attachments,
PCI/USB devices, boot order, scripts, SSH password — everything specific to a
virtual machine. When you use the TUI to create or edit a VM, this is the file
being modified.

The TUI is a **configuration authoring tool** for this file. It validates input,
shows available hardware (PCI devices, LVM volumes), and writes structured YAML.
It never modifies `~/.dkvmmanager.yaml`.

## The runner lifecycle

When you start a VM, DKVM Manager creates a **runner** — an in-process object
that owns the QEMU process for the lifetime of the VM:

```
TUI → NewVMRunner() → Start() → spawn QEMU → connect QMP → begin polling
```

The runner does three things concurrently:

| Activity | Cadence | What it reads |
|----------|---------|---------------|
| Status polling | Every 500 ms | QEMU running state and vCPU thread IDs via QMP |
| Metrics polling | Every 2 s | CPU usage, disk I/O, memory balloon via QMP + `/proc/<pid>` |
| Log streaming | Event-driven | QEMU stdout/stderr via pipe, script output via buffer |

**Why three separate loops?** Because they have different requirements. Status
(up/down/paused) must feel responsive — 500 ms is unnoticeable latency. Metrics
(CPU%, disk I/O) are expensive to collect and change slowly — 2 s is plenty.
Logs should arrive as they happen, not on a timer.

The TUI subscribes to these data streams and renders them. If you close the TUI
while a VM is running, the runner keeps running — the VM stays up. You can
re-attach later.

## QMP: the QEMU control channel

QEMU exposes a [QEMU Machine Protocol (QMP)](https://wiki.qemu.org/Features/QMP)
socket — a JSON-based control channel. DKVM Manager uses it to:

- Query VM status (`query-status`, `query-cpus-fast`)
- Collect block device statistics (`query-blockstats`)
- Read memory balloon info (`query-balloon`)
- Issue power commands (`system_powerdown`)

QMP is textual and human-readable. You can explore it manually by connecting
to the QMP socket while a VM is running:

```bash
# While DKVM is running a VM
socat -,echo=0,icanon=0 unix-connect:/media/dkvmdata/vms/<id>/qmp.sock
```

Type `{"execute":"qmp_capabilities"}` then `{"execute":"query-status"}`.

## Why hugepages?

Every VM guest memory allocation goes through the host CPU's Translation
Lookaside Buffer (TLB). With standard 4 KB pages, a VM with 8 GB of RAM needs
2 million page table entries — the TLB misses constantly. With 2 MB hugepages,
that drops to 4,096 entries.

DKVM Manager detects total system memory at startup, reserves 4 GB for the host
OS, and attempts to allocate the rest as 2 MB hugepages via
`/proc/sys/vm/nr_hugepages`. This is why `root` access is needed for the
initial setup.

## The metrics pipeline

DKVM Manager collects two kinds of metrics:

**QMP-sourced** (disk I/O, balloon memory): Queried through the QMP socket
using `query-blockstats` and `query-balloon`. These are instantaneous counters
that QEMU maintains internally.

**`/proc`-sourced** (CPU time, RSS): Read from `/proc/<pid>/stat` and
`/proc/<pid>/status`. These are kernel counters.

For CPU percentage, the runner reads the process's `utime` + `stime` (in clock
ticks), waits 2 seconds, reads again, and computes delta / elapsed time. This
gives an accurate CPU% without adding instrumentation to QEMU.

## See Also

- [Understanding LBU](understanding-lbu.md) — why `lbu commit` exists and how
  Alpine diskless mode works
- [Architecture](../dev/architecture.md) — developer-oriented architecture
  documentation
- [ADR-0001: Runner owns the running-VM data plane](../adr/0001-runner-owned-running-vm-data-plane.md)
  — why the runner design was chosen
