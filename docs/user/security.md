# Security Considerations

Security posture of DKVM Manager across its operations — what is protected,
what is exposed, and what you should harden in your deployment.

> **You should know**: DKVM Manager is designed as a **dedicated virtualization
> host**. It assumes a single-user, single-purpose machine running Alpine Linux
> in diskless mode. Many security decisions are inherited from this architecture.
> See [How DKVM Manager Works](../explanation/how-dkvm-manager-works.md) for background.

---

## SSH Password

### Mechanism

DKVM Manager changes the Alpine root password via the **Set SSH Password** form
(Configuration tab → index 9). The form uses:

1. `chpasswd` — receives `USER:password` on stdin to update the local password database
2. `lbu commit` — persists the change to the Alpine diskless overlay

Username defaults to `$USER` from the environment, falling back to `root`.

> **Source**: `internal/tui/models/ssh_password_form_validation.go` → `apply()`.

### Security properties

| Aspect | Status |
|--------|--------|
| **Password in memory** | Cleartext in Go process memory during form lifetime; not persisted to disk by DKVM itself |
| **Password on wire** | `chpasswd` reads from a pipe (stdin) — no network exposure |
| **Hashing algorithm** | Delegated to the host's PAM/shadow configuration (DKVM does not control this) |
| **Short password allowed** | Minimum 6 characters (enforced by form validation) — weaker than many enterprise policies |
| **No account lockout** | Repeated `chpasswd` calls are not rate-limited by DKVM |
| **lbu commit scope** | Persists the entire Alpine overlay, not just the password — unintentional system state changes may be committed alongside |

### Recommendations

- Use a **strong password** (score 4–5 in the form's built-in strength meter:
  at least 10 characters with mixed case, digits, and symbols)
- If running in multi-user environments, restrict TUI access to trusted users —
  the form can change any password accessible to the running user
- Consider SSH key authentication instead of password-based login; DKVM does not
  manage SSH keys directly
- Review the Alpine `lbu` configuration (`/etc/lbu/lbu.conf`) to understand
  which files are included in the commit

### Cross-references

- [Scripts & SSH](scripts-and-ssh.md) — form usage walkthrough

---

## PCI Passthrough Security

### Mechanism

DKVM Manager assigns host PCI devices (GPUs, USB controllers, NICs) directly to
VMs via VFIO. The workflow:

1. **Select devices** in the PCI Passthrough form (Configuration tab → index 5)
2. **Apply to Kernel** writes `vfio-pci.ids=<vendor>:<device>,...` to the
   GRUB `linux` line in `/media/usb/boot/grub/grub.cfg`
3. On next host reboot, the kernel binds matching devices to `vfio-pci` at boot
   before host drivers claim them
4. QEMU is launched with `-device vfio-pci,host=<address>` arguments

> **Source**: `internal/vm/grub_config.go` → `UpdateGrubVFIOIDs()`;
> `internal/vm/vm_runner_config.go` → `buildQEMUArgs()`.

### Security properties

| Aspect | Status |
|--------|--------|
| **IOMMU DMA protection** | Enabled when `intel_iommu=on` / `amd_iommu=on` is in kernel cmdline (required for VFIO). Prevents DMA attacks from guest to host memory |
| **GPU reset attacks** | Some GPUs (especially NVIDIA consumer cards) do not fully reset across VM reboots — a malicious guest could leave the GPU in a state that leaks data or degrades stability on the next boot. **Function-Level Reset (FLR)** is GPU-dependent |
| **Device isolation** | VFIO binds the entire PCI device to the guest — the host loses access. Multi-function devices (e.g., GPU + audio controller) in the same IOMMU group are toggled together |
| **ROM file exposure** | GPU ROM files specified via `romfile=` in QEMU args must be readable by the DKVM user. A compromised ROM file could affect GPU initialization |
| **IOMMU group co-tenancy** | Devices sharing an IOMMU group are all assigned together. DKVM enforces strict group selection — toggling one device toggles the entire group |
| **GRUB modification** | `UpdateGrubVFIOIDs()` writes to the GRUB config and creates a `.bak` backup before modifying. No checksum validation of the backup exists |

### DMA and IOMMU

Without IOMMU (or with `iommu=pt`), DMA remapping is disabled and the guest can
potentially read/write arbitrary host physical memory. **DKVM Manager requires
IOMMU for PCI passthrough but does not verify the kernel cmdline at runtime** —
ensure `intel_iommu=on` or `amd_iommu=on` is present in your bootloader config.

### Recommendations

- Verify IOMMU is active: `dmesg | grep -i "dmar\|iommu\|VFIO"`
- Prefer AMD GPUs (RX 5000 series and newer) which have better reset behavior
- For NVIDIA, research the specific card's reset capability before assigning it
  to VMs that will be rebooted frequently
- Review IOMMU groups (`find /sys/kernel/iommu_groups/ -type l`) before assigning
  devices — avoid sharing groups with host-critical devices (storage, network)
- Treat GPU ROM files from untrusted sources with the same scrutiny as firmware
- Keep the `.bak` GRUB backup until you've confirmed the new config boots correctly

### Cross-references

- [Hardware Configuration](hardware-config.md#pci-passthrough) — form usage and IOMMU groups
- [Setup & Prerequisites](setup.md) — IOMMU and GRUB configuration

---

## Configuration File Permissions

### Files

DKVM Manager uses two configuration files:

| File | Path | Contents |
|------|------|----------|
| App config | `~/.dkvmmanager.yaml` | QEMU binary path, data folder, network bridge, reserved memory |
| VM repository | `/media/dkvmdata/dkvmmanager/config.yaml` | All VM definitions, PCI/USB passthrough config, CPU topology, scripts |

> **Source**: `internal/config/config.go` → `Load()`, `Save()`;
> `internal/vm/repository.go` → `NewRepository()`.

### Security properties

| Aspect | Status |
|--------|--------|
| **File permissions** | DKVM does **not** set explicit permissions — files inherit the umask of the running process |
| **Sensitive data in VM config** | VM names, disk paths, MAC addresses, PCI addresses, ROM paths, script paths. **No plaintext passwords are stored** (the SSH password lives in the Alpine shadow file, which is in the lbu overlay) |
| **Write access** | Any user with write access to `config.yaml` can add/delete VMs, reconfigure PCI passthrough, and change scripts that execute as the DKVM user |
| **Read access** | Reveals VM topology: which PCI devices are assigned, disk layout, network configuration |

### Recommendations

- Restrict write access to the DKVM user only:
  ```bash
  chmod 600 ~/.dkvmmanager.yaml
  chmod 600 /media/dkvmdata/dkvmmanager/config.yaml
  ```
- Ensure `/media/dkvmdata` is mounted with restrictive permissions if the
  storage volume is shared or network-attached
- Audit group membership — users in the same group as the DKVM user may have
  read access depending on umask

### Cross-references

- [App Config Schema](../reference/app-config.md) — field reference
- [VM Config Schema](../reference/vm-config.md) — per-VM YAML structure

---

## Log File Exposure

### Files

| Log | Path | Contains |
|-----|------|----------|
| App log | `/var/log/dkvm.log` | Startup info, warnings, errors, debug messages |
| Per-VM QEMU log | `/media/dkvmdata/vms/<id>/qemu.log` | QEMU stdout/stderr, script output, console output |

> **Source**: `internal/config/config.go` → `LogFile`;
> `internal/vm/vm_runner.go` → `startPersistLog()`.

### Security properties

| Aspect | Status |
|--------|--------|
| **Console output in logs** | `qemu.log` captures QEMU stdout/stderr including guest OS console output. If a VM boots to a login prompt, failed login attempts and usernames may be logged |
| **Script output** | Start/stop script stdout/stderr is written to `qemu.log` with `[start]`/`[stop]` prefixes. Secrets printed by scripts (e.g., `echo $API_KEY`) are logged |
| **File permissions** | DKVM does not set explicit permissions on log files — they inherit process umask |
| **Log rotation** | No automatic rotation. `qemu.log` is created fresh on each VM start; the previous log is overwritten |
| **App log location** | `/var/log/dkvm.log` — typically requires root or `adm` group to read (system-dependent) |

### Recommendations

- Set restrictive permissions on the VM data directory:
  ```bash
  chmod 700 /media/dkvmdata/vms
  ```
- Avoid emitting secrets from start/stop scripts — if necessary, redirect
  sensitive output to `/dev/null` or a dedicated secret file
- Monitor `qemu.log` size for long-running VMs with verbose guests
- Consider a logrotate configuration for `/var/log/dkvm.log` if the app log
  grows large in debug mode (debug output is not rate-limited)

### Cross-references

- [Scripts & SSH](scripts-and-ssh.md) — script execution and log output
- [Running VMs](running-vms.md) — log viewer access

---

## Command Execution & Privileges

### Commands executed by DKVM Manager

| Command | Context | Privilege needed |
|---------|---------|-----------------|
| `chpasswd` | SSH password form | Root (modifies `/etc/shadow`) |
| `lbu commit` | SSH password form, System tab | Root (writes to overlay) |
| `/sbin/poweroff` | System tab | Root |
| `/sbin/reboot` | System tab | Root |
| `lvcreate` | LVM volume creation | Root or `lvm` group |
| `vgs` / `lvs` | LVM scanning | Root or `lvm` group |
| `lsblk` | Disk scanning | Any user (reads `/sys/block`) |
| `lspci` | PCI device scanning | Any user |
| `lsusb` | USB device scanning | Any user |
| `mount -o remount` | Filesystem remount for GRUB writes | Root |
| `qemu-system-x86_64` | VM creation (copy OVMF) | Read access to OVMF files |
| `qemu-system-x86_64` | VM execution | Root or `kvm` group + VFIO access |
| `/bin/bash <script>` | Start/stop scripts | Inherits DKVM user privileges |
| `swtpm` | TPM emulation | DKVM user |

> **Sources**: `internal/tui/models/ssh_password_form_validation.go` →
> `apply()`; `internal/tui/models/debug.go` → `runLBUCommit()`, `runReboot()`,
> `runPowerOff()`; `internal/tui/models/lv_create_form.go` → `buildCommand()`;
> `internal/vm/vm_runner.go` → `Start()`.

### Security properties

| Aspect | Status |
|--------|--------|
| **Privilege model** | DKVM runs as a single user with broad system access. There is no internal privilege separation — every form can execute any command the DKVM user can |
| **Input sanitization** | LVM volume names are validated against a regex (`[a-zA-Z0-9_-]{1,128}`). Size values are parsed as floats. PCI device addresses are validated against known scan results. **No shell escaping layer exists** — commands are built via `exec.Command()` with a separate args array, avoiding shell injection |
| **Script execution** | Custom start/stop scripts run as `/bin/bash <script>` with full DKVM user privileges. Any process accessible to the DKVM user can be launched |
| **Dry-run mode** | When `--dry-run` flag is active, `chpasswd`, `lbu commit`, `/sbin/reboot`, and `/sbin/poweroff` are **not executed** — only log messages are emitted. lvcreate and qemu-system-x86_64 still execute normally |

### Recommendations

- Run DKVM Manager as a **dedicated user** (not root) where possible, granting
  only the specific capabilities needed via `sudo` or `setcap`
- For `chpasswd`, `lbu commit`, `poweroff`, and `reboot`: configure passwordless
  `sudo` for the DKVM user rather than running the TUI as root
- Review custom start/stop scripts before enabling them — they run with full
  DKVM user privileges
- Test with `--dry-run` before production changes to the SSH password or system
  commands
- Avoid storing sensitive data in start/stop script paths or script content
  that may be visible in the VM config YAML

### Cross-references

- [Scripts & SSH](scripts-and-ssh.md) — script execution flow
- [Power & Save](power-and-save.md) — System tab operations

---

## Network Exposure (VNC)

### Mechanism

DKVM Manager exposes VM graphical console via VNC when the **VNC Enabled**
toggle is on (in the VM creation/edit form). The QEMU argument is:

```
-vnc 0.0.0.0:0
```

The value comes from the VM's `vnc_listen` field, defaulting to `"0.0.0.0:0"`
which binds to **all interfaces on the first available TCP port** (usually 5900).

> **Source**: `internal/vm/vm_runner_config.go` → `buildQEMUArgs()`;
> `internal/tui/models/vm_form.go` → `vncEnabled` toggle.

### Security properties

| Aspect | Status |
|--------|--------|
| **Binding address** | `0.0.0.0` — accessible from **any network interface** on the host, including public-facing NICs |
| **Authentication** | **None**. QEMU VNC has no password, no TLS, no SASL authentication configured by DKVM Manager |
| **Encryption** | **None**. VNC traffic is unencrypted — keyboard input, screen contents, and clipboard data travel in cleartext |
| **Port assignment** | `:0` means QEMU picks the first available port starting at 5900. The actual port is visible in `ps aux` and the VM details view |

### Recommendations

- **Disable VNC** if you do not need graphical console access (default is
  `-nographic` for headless operation)
- If VNC is required, bind to `127.0.0.1` instead of `0.0.0.0` and use SSH
  tunneling:
  ```bash
  ssh -L 5900:127.0.0.1:5900 user@dkvm-host
  ```
  Then connect the VNC client to `localhost:5900`
- Consider a firewall rule blocking external access to the VNC port range
  (5900–5999):
  ```bash
  iptables -A INPUT -p tcp --dport 5900:5999 -j DROP
  ```
- If using a bridged network, ensure the bridge does not expose the VNC port
  to untrusted VLANs

### Cross-references

- [VM Management](vm-management.md) — enabling/disabling VNC in the VM form
- [Running VMs](running-vms.md) — viewing VNC connection info

---

## QMP Socket

### Mechanism

QEMU's QMP (QEMU Machine Protocol) control socket is created at:

```
/media/dkvmdata/vms/<id>/qmp.sock
```

It is a Unix domain socket started with `server=on,wait=off` — QEMU creates the
socket and DKVM connects to it. Through QMP, DKVM can:

- Query VM status (`query-status`)
- Collect CPU, block, network, and balloon metrics
- Pause/Resume the VM (`stop` / `cont`)
- Power down the VM (`system_powerdown`)
- Shut down QEMU (`quit`)

> **Source**: `internal/vm/qmp_client.go` → `NewQMPClient()`, `Execute()`;
> `internal/vm/vm_runner_config.go` → `buildQEMUArgs()`.

### Security properties

| Aspect | Status |
|--------|--------|
| **Socket type** | Unix domain socket — only accessible from the local host |
| **Authentication** | **None**. QMP capability negotiation (`qmp_capabilities`) is performed but no authentication handshake exists |
| **Access control** | Controlled by filesystem permissions on the socket file. Any user with read/write access to the socket can send arbitrary QMP commands |
| **Commands available** | Full QEMU control: VM pause/resume, device hotplug (if implemented), system powerdown, and QEMU exit |
| **Impact of compromise** | An attacker with socket access can: force-shutdown VMs, pause VMs indefinitely, or trigger `quit` to kill the QEMU process |

### Recommendations

- The socket lives inside `/media/dkvmdata/vms/<id>/` — restrict directory
  permissions:
  ```bash
  chmod 700 /media/dkvmdata/vms
  ```
- Do not expose the data folder via NFS or other network filesystems without
  understanding that Unix socket files become accessible to NFS clients
- Monitor for unexpected QMP connections (unusual socket activity)

### Cross-references

- [How DKVM Manager Works](../explanation/how-dkvm-manager-works.md#qmp-the-qemu-control-channel) — QMP architecture
- [App Config Schema](../reference/app-config.md) — data folder location

---

## See Also

- [Setup & Prerequisites](setup.md) — initial host hardening
- [Troubleshooting](troubleshooting.md) — debugging permission issues
- [How DKVM Manager Works](../explanation/how-dkvm-manager-works.md) — architecture overview
- [Understanding LBU](../explanation/understanding-lbu.md) — Alpine diskless mode and `lbu commit`
- [App Config Schema](../reference/app-config.md) — file locations
- [VM Config Schema](../reference/vm-config.md) — per-VM YAML structure
