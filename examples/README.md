# Example Scripts

Reference scripts for PCI passthrough configuration with DKVM Manager.

---

## `amd_ryzen_start_stop.sh`

AMD Ryzen GPU passthrough start/stop script.

**Purpose:** Binds and unbinds PCI devices for GPU passthrough on AMD Ryzen systems. Handles amdgpu driver cycling required by AMD GPUs.

**Prerequisites:**
- AMD Ryzen system with integrated GPU (iGPU) + discrete GPU (dGPU)
- IOMMU enabled (VT-d/AMD-Vi)
- `vfio-pci` kernel module loaded

**Usage:**

```bash
./amd_ryzen_start_stop.sh start 0000:01:00.0 0000:02:00.0
./amd_ryzen_start_stop.sh stop  0000:01:00.0 0000:02:00.0
```

**What it does:**
- `start`: Protects iGPU from amdgpu via `driver_override`, reloads amdgpu for dGPU, unbinds current drivers, binds devices to `vfio-pci`
- `stop`: Unbinds devices from `vfio-pci`, cleans up iGPU driver overrides

**Configuration in DKVM Manager:**
Set Configuration → **Edit Start/Stop Script** → Mode: **Custom**, point to this script path.

> **Note:** This script is AMD-specific. The `driver_override` approach for iGPU protection may need adjustment for other vendors.

---

## `custom_pci_example.sh`

Minimal example start/stop script for PCI passthrough.

**Purpose:** Demonstrates the basic pattern for binding/unbinding PCI devices to `vfio-pci`. Simpler and vendor-agnostic compared to the AMD Ryzen script.

**Prerequisites:**
- PCI devices suitable for passthrough
- `vfio-pci` kernel module loaded

**Usage:**

```bash
./custom_pci_example.sh start 0000:01:00.0 0000:02:00.0
./custom_pci_example.sh stop  0000:01:00.0 0000:02:00.0
```

**What it does:**
- `start`: Unbinds each device from its current driver, binds to `vfio-pci`
- `stop`: Unbinds each device from `vfio-pci` (driver re-assignment left to the system)

**Configuration in DKVM Manager:**
Set Configuration → **Edit Start/Stop Script** → Mode: **Custom**, point to this script path.

---

## Script Arguments

Both scripts receive the same arguments from DKVM Manager:

| Position | Value | Example |
|----------|-------|---------|
| `$1` | Command: `start` or `stop` | `start` |
| `$2+` | PCI device addresses (0000:BB:DD.F format) | `0000:01:00.0` |

---

## DKVM Manager Integration

These scripts are used as custom start/stop hooks. Configure them in the TUI:

1. Switch to **Configuration** tab
2. Select **Edit Start/Stop Script** (index 7)
3. Toggle Mode to **Custom**
4. Browse to or type the script path
5. Save

The runner executes the script before QEMU starts and after it stops. See [Scripts & SSH](../docs/user/scripts-and-ssh.md) for full details.

---

## See Also

- [Scripts & SSH](../docs/user/scripts-and-ssh.md) — how to configure scripts in DKVM Manager
- [Hardware Configuration](../docs/user/hardware-config.md) — PCI passthrough device configuration
- [Setup & Prerequisites](../docs/user/setup.md) — IOMMU and vfio-pci setup
