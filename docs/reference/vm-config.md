# VM Configuration Schema (`/media/dkvmdata/dkvmmanager/config.yaml`)

This file stores all VM definitions and global hardware configuration for DKVM Manager. It is read/written by the TUI when you create, edit, or delete VMs, and when you modify passthrough, CPU, or script settings.

> **⚠️ Two configuration files — don't confuse them.** DKVM Manager uses **two separate YAML files** for different purposes. This document describes the VM config file. See [App Config Schema](app-config.md) for the application settings file at `~/.dkvmmanager.yaml`.

> **Source:** `internal/vm/repository.go` — `Repository` (Viper-backed YAML store); `internal/domain/models.go` — data model structs; `internal/vm/run_config.go` — `LoadRunConfigFromRepo()`.

---

## Top-Level Structure

```yaml
vms:                  # VM definitions (map of VM ID → VM object)
  "<id>":             # VM ID string, e.g., "0", "1"
    ...

pci_passthrough:      # Global PCI passthrough configuration
  devices: [...]

usb_passthrough:      # Global USB passthrough configuration
  devices: [...]

cpu_topology:         # CPU topology allocation
  ...

vcpu_pinning:         # vCPU-to-host-CPU pinning
  ...

cpu_options:          # CPU performance flags
  ...

custom_script:        # Start/stop hook scripts
  ...
```

All top-level keys except `vms` are optional. Missing keys produce zero-valued defaults.

---

## `vms` — VM Definitions

Each VM is stored under a numeric string ID (e.g., `"0"`, `"1"`, … `"9"`). DKVM supports up to 10 concurrent VMs (IDs 0–9).

### VM Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `id` | `string` | yes | — | Numeric VM ID (e.g., `"0"`) |
| `name` | `string` | yes | — | Human-readable VM name |
| `created_at` | `string` (RFC3339) | auto | current time | Creation timestamp |
| `updated_at` | `string` (RFC3339) | auto | current time | Last update timestamp |
| `harddisks` | `[]string` | no | `[]` | Block device paths or file paths for disk images |
| `cdroms` | `[]string` | no | `[]` | ISO file paths for optical drives |
| `gpu_rom` | `string` | no | `""` | (Deprecated) VBIOS ROM file path; use per-device `rom_path` in `pci_passthrough` instead |
| `mac` | `string` | no | random | MAC address (auto-generated if omitted; locally-administered: bit 1 set) |
| `network_mode` | `string` | no | `"nat"` | Network mode: `"bridge"` or `"nat"` |
| `vnc_listen` | `string` | no | `""` | VNC listen address and port (e.g., `"0.0.0.0:0"` for dynamic port on all interfaces) |
| `tpm_enabled` | `bool` | no | `false` | Enable emulated TPM 2.0 via swtpm |

**Example VM entry:**

```yaml
vms:
  "0":
    id: "0"
    name: "windows-gaming"
    created_at: "2024-06-01T10:00:00Z"
    updated_at: "2024-06-15T14:30:00Z"
    harddisks:
      - /dev/dkvm_vg/windows_disk
    cdroms:
      - /media/dkvmdata/isos/virtio-win.iso
    mac: "02:1a:2b:3c:4d:5e"
    network_mode: bridge
    vnc_listen: "0.0.0.0:0"
    tpm_enabled: true
```

> **Source:** `internal/domain/models.go` — `VM` struct; `internal/vm/repository.go` — `SaveVM()`, `unmarshalVM()`.

---

## `pci_passthrough` — PCI Device Passthrough

Global list of PCI devices assigned to VMs. Each device entry:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `address` | `string` | yes | — | PCI bus address (e.g., `"0000:01:00.0"`) |
| `rom_path` | `string` | no | `""` | Optional VBIOS ROM file path for this device |
| `vendor` | `string` | auto | — | Vendor ID (e.g., `"10de"`) — populated from host scan |
| `device` | `string` | auto | — | Device ID (e.g., `"1b80"`) — populated from host scan |
| `name` | `string` | auto | — | Human-readable device name |
| `class_code` | `string` | auto | — | PCI class code (e.g., `"0300"` for VGA) |

**Example:**

```yaml
pci_passthrough:
  devices:
    - address: "0000:01:00.0"
      rom_path: /media/dkvmdata/vbios/gpu.rom
      vendor: "10de"
      device: "1b80"
      name: "NVIDIA GeForce GTX 1080"
      class_code: "0300"
    - address: "0000:01:00.1"
      vendor: "10de"
      device: "10f0"
      name: "NVIDIA HD Audio"
      class_code: "0403"
```

> **Source:** `internal/domain/models.go` — `PCIPassthroughDevice`, `PCIPassthroughConfig` structs.

---

## `usb_passthrough` — USB Device Passthrough

Global list of USB devices assigned to VMs.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `vendor` | `string` | yes | — | USB vendor ID (e.g., `"046d"`) |
| `product` | `string` | yes | — | USB product ID (e.g., `"c52b"`) |
| `name` | `string` | no | `""` | Human-readable device name |
| `bus_id` | `string` | no | `""` | Bus device identifier (e.g., `"1-1"`) |

**Example:**

```yaml
usb_passthrough:
  devices:
    - vendor: "046d"
      product: "c52b"
      name: "Logitech Unifying Receiver"
      bus_id: "1-1"
    - vendor: "1532"
      product: "0083"
      name: "Razer DeathAdder"
      bus_id: "1-3"
```

> **Source:** `internal/domain/models.go` — `USBPassthroughDevice`, `USBPassthroughConfig` structs.

---

## `cpu_topology` — CPU Topology Allocation

Controls how host CPU topology is exposed to VMs.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | `bool` | no | `false` | Whether custom topology is enabled |
| `selected_cpus` | `[]int` | no | `[]` | Logical CPU IDs allocated to VMs |
| `use_host_topology` | `bool` | no | `false` | Use host topology layout (dies/sockets) vs flat layout |

**Example:**

```yaml
cpu_topology:
  enabled: true
  selected_cpus: [0, 1, 2, 3, 8, 9, 10, 11]
  use_host_topology: true
```

> **Source:** `internal/domain/models.go` — `CPUTopology` struct.

---

## `vcpu_pinning` — vCPU-to-Host-CPU Pinning

Maps guest vCPUs to host logical CPUs.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | `bool` | no | `false` | Whether vCPU pinning is active |
| `mappings` | `[]VCPUToHostMapping` | no | `[]` | Array of vCPU-to-host-CPU pin mappings |

Each mapping:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `vcpu_id` | `int` | yes | — | Guest vCPU index |
| `host_cpu_id` | `int` | yes | — | Host logical CPU ID |

When pinning is enabled, DKVM also updates the kernel command-line via GRUB (`isolcpus=`, `nohz_full=`, `rcu_nocbs=`) for the mapped host CPUs.

**Example:**

```yaml
vcpu_pinning:
  enabled: true
  mappings:
    - vcpu_id: 0
      host_cpu_id: 0
    - vcpu_id: 1
      host_cpu_id: 1
    - vcpu_id: 2
      host_cpu_id: 2
    - vcpu_id: 3
      host_cpu_id: 3
```

> **Source:** `internal/domain/models.go` — `VCPUPinningGlobal`, `VCPUToHostMapping` structs.

---

## `cpu_options` — CPU Performance Flags

Hypervisor CPU features and optimizations exposed to the guest.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `hide_kvm` | `bool` | `false` | Hide KVM hypervisor signature from guest |
| `vendor_id` | `string` | `""` | Custom CPU vendor ID string (e.g., `"GenuineIntel"`) |
| `hv_frequency` | `bool` | `false` | Hyper-V frequency MSRs (hv_frequencies) |
| `hv_relaxed` | `bool` | `false` | Hyper-V relaxed timer (hv_relaxed) |
| `hv_reset` | `bool` | `false` | Hyper-V reset (hv_reset) |
| `hv_runtime` | `bool` | `false` | Hyper-V runtime (hv_runtime) |
| `hv_spinlocks` | `string` | `""` | Hyper-V spinlock retry count (e.g., `"0x1fff"`) |
| `hv_stimer` | `bool` | `false` | Hyper-V synthetic timers (hv_stimer) |
| `hv_synic` | `bool` | `false` | Hyper-V synthetic interrupt controller (hv_synic) |
| `hv_time` | `bool` | `false` | Hyper-V time (hv_time) |
| `hv_vapic` | `bool` | `false` | Hyper-V virtual APIC (hv_vapic) |
| `hv_vpindex` | `bool` | `false` | Hyper-V VP index (hv_vpindex) |
| `hv_no_nonarch_coresharing` | `bool` | `false` | Disable non-architectural core sharing for Hyper-V |
| `hv_tlbflush` | `bool` | `false` | Hyper-V TLB flush (hv_tlbflush) |
| `hv_tlbflush_ext` | `bool` | `false` | Hyper-V extended TLB flush (hv_tlbflush_ext) |
| `hv_ipi` | `bool` | `false` | Hyper-V IPI (hv_ipi) |
| `hv_avic` | `bool` | `false` | Hyper-V virtual APIC (hv_avic) |
| `topoext` | `bool` | `false` | Enable CPU topology extensions (topoext) |
| `l3_cache` | `bool` | `false` | Expose L3 cache topology to guest |
| `l3_cache_size_die` | `map[int]string` | `{}` | L3 cache size per die (e.g., `0: "0x8000001D"`) |
| `l3_cache_assoc_die` | `map[int]int` | `{}` | L3 cache associativity (ways) per die |
| `x2apic` | `bool` | `false` | Enable x2APIC |
| `migratable` | `bool` | `true` | Allow VM migration (clear CPUID bits that prevent migration) |
| `invtsc` | `bool` | `false` | Expose invariant TSC to guest |
| `force_cpuid_0x80000026` | `bool` | `false` | Force enabling of CPUID leaf 0x80000026 (AMD L3 topology) |
| `rtc_utc` | `bool` | `true` | RTC in UTC mode (vs localtime) |
| `cpu_pm` | `bool` | `false` | Enable CPU power management (halt, mwait) |

**Example:**

```yaml
cpu_options:
  hide_kvm: true
  vendor_id: "GenuineIntel"
  hv_frequency: true
  hv_relaxed: true
  hv_time: true
  hv_vapic: true
  hv_vpindex: true
  hv_spinlocks: "0x1fff"
  migratable: false
  rtc_utc: true
```

> **Source:** `internal/domain/models.go` — `CPUOptions` struct.

---

## `custom_script` — Start/Stop Hook Scripts

Controls scripts executed before VM start and after VM stop.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `use_builtin` | `bool` | no | `true` | If `true`, use builtin scripts; if `false`, use custom scripts below |
| `start_script` | `string` | no | `""` | Path to custom start script (absolute path or relative to VM data dir) |
| `stop_script` | `string` | no | `""` | Path to custom stop script (absolute path or relative to VM data dir) |

**Example:**

```yaml
custom_script:
  use_builtin: false
  start_script: /media/dkvmdata/vms/0/start.sh
  stop_script: /media/dkvmdata/vms/0/stop.sh
```

> **Source:** `internal/domain/models.go` — `StartStopScript` struct.

---

## Complete Example

A comprehensive example showing all sections together:

```yaml
vms:
  "0":
    id: "0"
    name: "windows-gaming"
    created_at: "2024-06-01T10:00:00Z"
    updated_at: "2024-06-20T09:15:00Z"
    harddisks:
      - /dev/dkvm_vg/windows_disk
    cdroms:
      - /media/dkvmdata/isos/virtio-win.iso
    mac: "02:1a:2b:3c:4d:5e"
    network_mode: bridge
    vnc_listen: "0.0.0.0:0"
    tpm_enabled: true

pci_passthrough:
  devices:
    - address: "0000:01:00.0"
      vendor: "10de"
      device: "1b80"
      name: "NVIDIA GeForce GTX 1080"
      class_code: "0300"

usb_passthrough:
  devices:
    - vendor: "046d"
      product: "c52b"
      name: "Logitech Unifying Receiver"

cpu_topology:
  enabled: true
  selected_cpus: [0, 1, 2, 3, 8, 9, 10, 11]
  use_host_topology: true

vcpu_pinning:
  enabled: true
  mappings:
    - vcpu_id: 0
      host_cpu_id: 0
    - vcpu_id: 1
      host_cpu_id: 1
    - vcpu_id: 2
      host_cpu_id: 2
    - vcpu_id: 3
      host_cpu_id: 3

cpu_options:
  hide_kvm: true
  vendor_id: "GenuineIntel"
  hv_relaxed: true
  hv_time: true
  hv_vapic: true
  hv_vpindex: true
  migratable: false

custom_script:
  use_builtin: false
  start_script: /media/dkvmdata/vms/0/start.sh
  stop_script: /media/dkvmdata/vms/0/stop.sh
```

---

## See Also

- [App Config Schema](app-config.md) — the `~/.dkvmmanager.yaml` application settings file
- [CLI Flags Reference](cli-flags.md) — command-line flags
- [Setup & Prerequisites](../user/setup.md) — system requirements and first launch
- [Hardware Configuration](../user/hardware-config.md) — how PCI, USB, CPU settings are configured via the TUI
- [User Guide Index](../user/README.md) — all documentation
