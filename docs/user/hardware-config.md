# Hardware Configuration

Configure CPU topology, vCPU pinning, CPU feature flags, and PCI/USB device passthrough for VMs.

## Prerequisites

- VM created (see [vm-management.md](vm-management.md))
- [Setup completed](setup.md) — KVM, hugepages, IOMMU configured
- For PCI passthrough: IOMMU enabled, vfio-pci driver loaded (see [setup.md](setup.md))
- For vCPU pinning: CPU topology must be configured first

> **You should know**: See [How DKVM Manager Works](../explanation/how-dkvm-manager-works.md) for CPU topology, vCPU pinning, and IOMMU background.

## Navigation

### Accessing hardware forms

1. Press `Tab` to switch to the **Configuration** tab
2. Use `↑/↓` or `j/k` to highlight a menu item
3. Press `Enter` or `Space` to select

The Configuration menu contains hardware-related items:

| Index | Item | Description |
|-------|------|-------------|
| 3 | Edit CPU Topology | Guest CPU socket/core/thread layout |
| 4 | Edit vCPU Pinning | Pin virtual CPUs to host cores |
| 5 | Edit PCI Passthrough | Assign host PCI devices to VMs |
| 6 | Edit USB Passthrough | Assign host USB devices to VMs |
| 8 | Edit CPU Options | CPU model, features, hypervisor flags |

> **Source**: `internal/tui/models/init.go` → `registerAllViews()`, `buildConfigListAdapter()`.

---

## CPU Topology

Configure which host physical cores are allocated to VMs vs. reserved for the host OS.

### Opening the form

Configuration tab → **Edit CPU Topology** (index 3).

> **Source**: `internal/tui/models/cpu_topology.go` → `NewCPUTopologyModel()`; `internal/tui/models/cpu_topology_form.go` → `NewCPUTopologyFormModel()`.

### Form layout

The form displays host CPU topology grouped by die, with each physical core shown as a toggle:

```
CPU Topology
Host: 2 dies, 12 cores, 24 threads

Die 0 — L3 Cache: 32M
  [HOST] Core 0  [2 threads: 0,12]
> [ VM ] Core 1  [2 threads: 1,13]
  [HOST] Core 2  [2 threads: 2,14]

Die 1 — L3 Cache: 32M
  [ VM ] Core 0  [2 threads: 6,18]
  [ VM ] Core 1  [2 threads: 7,19]

Summary: 3 cores for VMs, 9 for host
[Space/Enter] Save    [ESC] Cancel
```

- **`[HOST]`**: Core reserved for host OS — not available to VMs
- **`[ VM ]`**: Core allocated to VMs — all its threads become selectable vCPUs
- Per-die L3 cache size shown when available
- Thread IDs listed per core (e.g., `[2 threads: 0,12]` — HT sibling pair)

### Keybindings

Use `Tab`/`↓` and `Shift+Tab`/`↑` to move between cores, `Space`/`Enter` to toggle a core or activate the Save button, `PgUp`/`PgDown` to scroll, and `ESC` to cancel.

See [Keybindings](keybindings.md) for the full reference.

### Validation

- At least one core must be allocated for VMs
- Warning shown if zero cores reserved for host ("system may become unresponsive")
- Selected CPU IDs (threads) are sorted and saved to `cpu_topology` config

> **Source**: `internal/tui/models/cpu_topology_form_validation.go` → `validateAndSaveCmd()`.

### Behind the scenes

- DKVM scans host CPU topology via `/sys/devices/system/cpu/` — dies, cores, threads, and L3 cache
- Configured CPU topology drives vCPU pinning (auto-computed from selected cores)
- Saved to `cpu_topology` key in the repository config

> **Source**: `internal/vm/cpu_scanner.go` → `CPUScanner.ScanTopology()`.

---

## vCPU Pinning

Bind guest vCPUs to specific host CPUs, with automatic alignment to host topology.

### Opening the form

Configuration tab → **Edit vCPU Pinning** (index 4).

> **Source**: `internal/tui/models/vcpu_pinning.go` → `NewVCPUPinningModel()`; `internal/tui/models/vcpu_pinning_form.go` → `NewVCPUPinningFormModel()`.

### Form layout

The form is read-only (pinning is auto-computed from CPU topology — edit core allocation in CPU Topology instead):

```
vCPU Pinning
Host: 2 dies, 12 cores, 24 threads
Current Allocation:
  Die 0: 2 cores (vCPUs 0-7) -> Host CPUs auto
  Die 1: 1 core (vCPUs 8-11) -> Host CPUs auto
Current Mappings (auto-computed from topology):
  vCPU 0 (die 0, siblings 0,1) -> Host CPU 1 (die 0, siblings 1,13)
  vCPU 1 (die 0, siblings 0,1) -> Host CPU 13 (die 0, siblings 1,13)
  ...
  Die mapping: OK (guest die 0 -> host die 0)
  Sibling alignment: OK
Summary: 12 vCPUs pinned
topology-aware

> [ON]  vCPU Auto Pinning
  [ON]  Use Host Topology

[Space/Enter] Save    [ESC] Cancel

[Space/Enter] Apply to Kernel    [ESC] Cancel
```

### Toggles

| Toggle | Default | Description |
|--------|---------|-------------|
| **vCPU Auto Pinning** | Depends on config | When ON, pinning is active. Save recomputes mappings from current topology. |
| **Use Host Topology** | Depends on config | When ON, guest vCPU topology mirrors host die/core layout. Saved to CPU topology config. |

### Keybindings

Use `Tab`/`↓` and `Shift+Tab`/`↑` to move between toggles and buttons, `Space`/`Enter` to toggle a value or activate a button, `PgUp`/`PgDown` to scroll, and `ESC` to cancel.

See [Keybindings](keybindings.md) for the full reference.

### Apply to Kernel

The **Apply to Kernel** button writes CPU isolation parameters to `/media/usb/boot/grub/grub.cfg`:

| Parameter | Purpose |
|-----------|---------|
| `isolcpus=` | Isolate host CPUs from scheduler |
| `nohz_full=` | Disable timer ticks on isolated CPUs |
| `rcu_nocbs=` | Offload RCU callbacks from isolated CPUs |

A `.bak` backup is created before writing. The filesystem must be writable (remount rw if needed). A status message shows success or error after the async operation completes.

> **Source**: `internal/vm/grub_config.go` → `UpdateGrubCPUParams()`; `internal/tui/models/vcpu_pinning_form_validation.go` → `handleApplyKernelCmd()`.

### Behind the scenes

- Pinning mappings are auto-computed from CPU topology: each selected host thread becomes a guest vCPU
- Guest die and sibling topology are derived from the host die/core layout
- Die mapping and sibling alignment checks are displayed inline
- Results are saved to `vcpu_pinning` config key

> **Source**: `internal/vm/cpu_scanner.go` → `ComputePinningFromTopology()`.

---

## CPU Options

Configure QEMU CPU model flags, Hyper-V enlightenments, hypervisor stealth, and advanced CPU features.

### Opening the form

Configuration tab → **Edit CPU Options** (index 8).

> **Source**: `internal/tui/models/cpu_options.go` → `NewCPUOptionsModel()`; `internal/tui/models/cpu_options_form.go` → `NewCPUOptionsFormModel()`.

### Form layout

The form has three sections plus per-die L3 cache overrides (shown when host topology is available):

```
CPU Options

== Hypervisor Stealth ==
  [OFF] Hides VM from guest
  [     ] Custom hypervisor vendor ID

== Hyper-V Enlightenments ==
  [OFF] Expose TSC/APIC frequencies
  [OFF] Relaxed timing checks
  [OFF] Guest reset capability
  [ON]  Hypervisor runtime info
  [0x1FF] Paravirtualized spinlocks
  [OFF] Synthetic timers
  [ON]  Synthetic interrupt controller
  [OFF] Reference TSC page
  [OFF] Exit-less EOI processing
  [OFF] Virtual CPU index
  [OFF] SMT perf counter isolation
  [OFF] Paravirtualized TLB flush
  [OFF] Extended TLB flush ranges
  [OFF] Paravirtualized IPI
  [OFF] Hyper-V nested APIC virt

== Advanced CPU Features ==
  [OFF] AMD topology extension
  [OFF] Expose host L3 cache info
  [OFF] x2APIC mode (>255 vCPUs)
  [OFF] Expose all host features (no live migration)
  [OFF] Invariant TSC
  [OFF] Force AMD CPUID
  [OFF] Use UTC time for RTC
  [OFF] Allow guest C/P-state control

[Space/Enter] Save    [ESC] Cancel
```

### Hypervisor Stealth section

| Field | Type | Description |
|-------|------|-------------|
| **Hides VM from guest** (`HideKVM`) | Toggle | Hides KVM signature from guest (`kvm=off`) |
| **Custom hypervisor vendor ID** (`VendorID`) | Text | 12-character vendor ID string (e.g., `GenuineIntel`) |

### Hyper-V Enlightenments section

Toggle-based features for Windows guests:

| Field | Key | Description |
|-------|-----|-------------|
| Expose TSC/APIC frequencies | `HVFrequency` | `hv-frequencies` |
| Relaxed timing checks | `HVRelaxed` | `hv-relaxed` |
| Guest reset capability | `HVReset` | `hv-reset` |
| Hypervisor runtime info | `HVRuntime` | `hv-runtime` |
| Paravirtualized spinlocks | `HVSpinlocks` | `hv-spinlocks=<value>` (text input) |
| Synthetic timers | `HVStimer` | `hv-stimer` |
| Synthetic interrupt controller | `HVSyncIC` | `hv-synic` |
| Reference TSC page | `HVTime` | `hv-time` |
| Exit-less EOI processing | `HVVapic` | `hv-vapic` |
| Virtual CPU index | `HVVPIndex` | `hv-vpindex` |
| SMT perf counter isolation | `HVNoNonarchCoresharing` | `hv-no-nonarch-coresharing` |
| Paravirtualized TLB flush | `HVTLBFlush` | `hv-tlbflush` |
| Extended TLB flush ranges | `HVTLBFlushExt` | `hv-tlbflush-ext` |
| Paravirtualized IPI | `HVIPI` | `hv-ipi` |
| Hyper-V nested APIC virt | `HVAVIC` | `hv-avic` |

### Advanced CPU Features section

| Field | Key | Description |
|-------|-----|-------------|
| AMD topology extension | `TopoExt` | `topoext=on` |
| Expose host L3 cache info | `L3Cache` | `host-cache-info=on` |
| x2APIC mode (>255 vCPUs) | `X2APIC` | `x2apic=on` |
| Expose all host features | `Migratable` | `migratable=no` (disables live migration) |
| Invariant TSC | `InvTSC` | `invtsc=on` |
| Force AMD CPUID | `ForceCPUID0x80000026` | Forces AMD CPUID leaf |
| Use UTC time for RTC | `RTCUTC` | `rtc base=utc` |
| Allow guest C/P-state control | `CPUPM` | `guest-cpupm=on` |

### Per-Die L3 Cache Override

When host topology scan succeeds, additional text fields appear per die:

- **Die N L3 cache size** — Override L3 cache size (e.g., `32M`, `96M`). Detected size shown when available.
- **Die N L3 cache associativity** — Override associativity (e.g., `12`). Detected value shown when available.

These override what QEMU reports to the guest. Only shown when `Expose host L3 cache info` (`L3Cache`) is ON.

### Keybindings

Use `Tab`/`↓` and `Shift+Tab`/`↑` to navigate fields, `Space`/`Enter` to toggle values or activate the Save button, `Backspace`/`Delete` for text input, `PgUp`/`PgDown` to scroll, and `ESC` to cancel.

See [Keybindings](keybindings.md) for the full reference.

### Validation

- **VendorID**: Must be exactly 12 characters or empty
- On save, options are persisted to `cpu_options` config key

> **Source**: `internal/tui/models/cpu_options_form_validation.go` → `validateField()`; `internal/tui/models/cpu_options_form_handlers.go` → `HandleEnter()`.

### Behind the scenes

- Options stored in `models.CPUOptions` struct and saved to `cpu_options` config key
- Text fields use reflection-based get/set via `getStringField()` / `setStringField()`
- Per-die L3 cache fields mapped to `L3CacheSizeDie` (map[int]string) and `L3CacheAssocDie` (map[int]int)
- Field registry: `internal/tui/models/fields/cpu_options.go` → `CPUOptionsFields`

> **Source**: `internal/tui/models/cpu_options_form_navigation.go` → `BuildPositions()`; `internal/tui/models/fields/cpu_options.go`.

---

## PCI Passthrough

Pass host PCI devices (GPUs, USB controllers, NICs) directly to VMs.

### Opening the form

Configuration tab → **Edit PCI Passthrough** (index 5).

> **Source**: `internal/tui/models/pci_passthrough.go` → `NewPCIPassthroughModel()`; `internal/tui/models/pci_passthrough_form.go` → `NewPCIPassthroughFormModel()`.

### Form layout

Devices are grouped by IOMMU group, with group headers showing selection status:

```
PCI Passthrough Configuration

── IOMMU Group 1 (3 devices, 2 selected) ──
  [X] 0000:01:00.0 [GPU] NVIDIA GeForce GTX 1080 [10de:1b80] (IOMMU:1)
  [ ] 0000:01:00.1 [    ] NVIDIA GP108 High Def Audio [10de:10f0] (IOMMU:1)

── IOMMU Group 5 (2 devices, all selected) ──
  [X] 0000:04:00.0 [USB] ASMedia USB 3.0 Controller [1b21:1142] (IOMMU:5)

── Ungrouped Devices (1 devices) ──
  [ ] 0000:00:1f.3 [    ] Intel Audio Device [8086:a171]

[Space/Enter] Save    [ESC] Cancel

[Space/Enter] Apply to Kernel    [ESC] Cancel
```

- **PCI address** shown in bold for quick scanning (e.g., `0000:01:00.0`)
- **Device tags**: `[GPU]`, `[USB]` for quick identification
- **Vendor:Device IDs** shown in muted style (e.g., `[10de:1b80]`)
- **IOMMU group** number shown per device when assigned
- Devices grouped by IOMMU group with selection summary in header

### Keybindings

Use `Tab`/`↓` and `Shift+Tab`/`↑` to navigate devices and buttons, `Space`/`Enter` to toggle a device selection or activate a button, `PgUp`/`PgDown` to scroll, and `ESC` to cancel. Group headers are display-only — focus moves only between device toggles and buttons.

See [Keybindings](keybindings.md) for the full reference.

### Strict IOMMU group selection

When toggling a device that belongs to a multi-device IOMMU group, **all devices in the group are toggled together**. This is enforced by the `toggleDevice()` method — selecting a GPU also selects its audio controller in the same IOMMU group.

Single-device groups and ungrouped devices (IOMMU group negative) toggle independently.

> **Source**: `internal/tui/models/pci_passthrough_form_navigation.go` → `toggleDevice()`.

### Apply to Kernel

The **Apply to Kernel** button writes `vfio-pci.ids=` to the GRUB `linux` line in `/media/usb/boot/grub/grub.cfg`. This ensures devices are bound to vfio-pci at boot before host drivers claim them.

A `.bak` backup is created before writing. The filesystem must be writable (remount rw if needed). A status message shows success or error after the async operation.

> `vfio-pci.ids=` syntax: comma-separated vendor:device pairs (e.g., `vfio-pci.ids=10de:1b80,10de:10f0`)

> **Source**: `internal/vm/grub_config.go` → `BuildVFIOIDs()`, `UpdateGrubVFIOIDs()`; `internal/tui/models/pci_passthrough_form_validation.go` → `handleApplyKernelCmd()`.

### Validation

- Devices are validated via `vm.ValidatePCIDevices()` before save
- Validation errors (e.g., conflicting devices) block save and display as inline errors
- Non-fatal warnings (e.g., IOMMU group partial selection warning) display below the footer after save
- PCI bridges (switch ports, root ports) are filtered out — only end devices are selectable

> **Source**: `internal/tui/models/pci_passthrough_form.go` → `filterPCIBridges()`; `internal/tui/models/pci_passthrough_form_validation.go` → `validateAndSaveCmd()`.

### Behind the scenes

- Devices scanned via `hostDiscovery.ScanPCIDevices()` (reads `/sys/bus/pci/devices/`)
- Config stored in `pci_passthrough` key as `PCIPassthroughConfig{Devices: []PCIPassthroughDevice}`
- Bridges filtered by `IsBridge` flag from PCI class code detection
- IOMMU groups indexed by `buildIOMMUGroups()` for fast group-based operations

> **Source**: `internal/vm/discovery.go` → `ScanPCIDevices()`.

---

## USB Passthrough

Pass host USB devices to VMs by vendor:product ID.

### Opening the form

Configuration tab → **Edit USB Passthrough** (index 6).

> **Source**: `internal/tui/models/usb_passthrough.go` → `NewUSBPassthroughModel()`; `internal/tui/models/usb_passthrough_form.go` → `NewUSBPassthroughFormModel()`.

### Form layout

Devices listed as a flat list of toggles:

```
USB Passthrough Configuration

  [X] Logitech G502 Mouse [046d:c332] (Bus 001)
  [ ] Corsair K70 Keyboard [1b1c:1b09] (Bus 003)
  [X] SanDisk USB Flash Drive [0781:5583] (Bus 002)

[Space/Enter] Save    [ESC] Cancel
```

- Each device shows: name, vendor:product ID, and bus ID
- Toggle `[X]` for selected, `[ ]` for deselected
- Focus indicator: `>` prefix on focused line

### Keybindings

Use `Tab`/`↓` and `Shift+Tab`/`↑` to navigate devices and buttons, `Space`/`Enter` to toggle a device selection or activate the Save button, `PgUp`/`PgDown` to scroll, and `ESC` to cancel.

See [Keybindings](keybindings.md) for the full reference.

### Validation

- Devices are validated via `vm.ValidateUSBDevices()` before save
- Validation errors block save and display as inline errors
- Non-fatal warnings display below the footer after successful save

> **Source**: `internal/tui/models/usb_passthrough_form_validation.go` → `validateAndSaveCmd()`.

### Behind the scenes

- Devices scanned via `hostDiscovery.ScanUSBDevices()` (reads `/sys/bus/usb/devices/`)
- Config stored in `usb_passthrough` key as `USBPassthroughConfig{Devices: []USBPassthroughDevice}`
- Device key: `vendor:product` (e.g., `046d:c332`)

> **Source**: `internal/vm/discovery.go` → `ScanUSBDevices()`.

---

> **Behind the scenes**: See [Architecture](../dev/architecture.md) for model hierarchy, message flow, and GRUB configuration integration details.

---

## See Also

- [VM Management](vm-management.md) — Create, edit, and delete VMs
- [Setup](setup.md) — IOMMU, vfio-pci, hugepages, LVM
- [Running VMs](running-vms.md) — Start, monitor, and stop VMs
- [Keybindings](keybindings.md) — Complete keyboard reference
