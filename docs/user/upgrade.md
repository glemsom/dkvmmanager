# Upgrade & Migration

How to upgrade DKVM Manager, check your current version, and migrate to a new
host.

---

## How to Check Current Version

### Installed binary

```bash
dkvmmanager --version
```

Output shows the semantic version, git commit, and build date:

```
DKVM Manager v0.2.0 (commit: a1b2c3d, built: 2026-06-22T12:00:00Z)
```

> **Note**: The `--version` flag is injected at build time via `-ldflags`.
> Binaries downloaded from GitHub Releases always include version information.
> If the flag shows `dev` / `none` / `unknown`, the binary was built locally
> without passing version metadata.

### Source checkout

The canonical version is in the `VERSION` file at the repository root:

```bash
cat VERSION
# 0.2.0
```

The Go constant `version.Version` in `internal/version/version.go` must match
the `VERSION` file. Run `make test` to verify they agree.

---

## Upgrade Binary from GitHub Releases

Releases are published on the
[GitHub Releases](https://github.com/glemsom/dkvmmanager/releases) page.

### Find the latest release

Use `gh` CLI or visit the releases page:

```bash
gh release list --repo glemsom/dkvmmanager --limit 5
```

### Download and replace

1. Download the binary for your platform (linux/amd64):

   ```bash
   VERSION="v0.x.x"
   curl -LO "https://github.com/glemsom/dkvmmanager/releases/download/${VERSION}/dkvmmanager"
   ```

2. Verify the checksum (recommended):

   ```bash
   curl -LO "https://github.com/glemsom/dkvmmanager/releases/download/${VERSION}/checksums.txt"
   sha256sum -c checksums.txt --ignore-missing
   ```

3. Replace the old binary:

   ```bash
   cp dkvmmanager /usr/local/bin/dkvmmanager
   chmod +x /usr/local/bin/dkvmmanager
   ```

4. Verify the new version:

   ```bash
   dkvmmanager --version
   ```

> **Tip**: You can also download archives. Each release includes a `.tar.gz`
> with the binary, `LICENSE`, `README`, and `CHANGELOG.md`.

> **Source**: `.goreleaser.yml` — release artifact configuration.

---

## Alpine Package Upgrade

If DKVM Manager was installed via the Alpine package repository:

```bash
apk upgrade dkvmmanager
```

The package manager handles binary replacement and dependency updates
automatically. No manual steps required.

> **Note**: Alpine packages may lag behind GitHub Releases. For the latest
> version, use the binary download method above.

---

## Version Compatibility Notes

### Config file format

DKVM Manager stores all VM configuration in
`/media/dkvmdata/dkvmmanager/config.yaml`.

| Version Change | Config Compatibility |
|----------------|----------------------|
| **Patch bump** (e.g. `0.1.0` → `0.1.1`) | Fully compatible — no changes to config format |
| **Minor bump** (e.g. `0.1.0` → `0.2.0`) | Backward compatible — new optional fields may be added; old configs still load |
| **Major bump** (e.g. `0.2.0` → `1.0.0`) | May include breaking changes — review the changelog before upgrading |

### Before upgrading

1. **Read the changelog** for the target version — look for `### Changed` or
   `### Fixed` sections that mention config or data format.
2. **Back up your config**:

   ```bash
   cp /media/dkvmdata/dkvmmanager/config.yaml /media/dkvmdata/dkvmmanager/config.yaml.bak
   ```

3. **Test on a non-production host** if possible.

### Detecting incompatibility

DKVM Manager validates the config file at startup. If the format is
incompatible, an error message appears in the TUI detailing the issue. The
TUI will not launch until the config is fixed or removed.

> **Source**: `internal/vm/repository.go` — config load and validation.

---

## Migrating to a New Host

Moving DKVM Manager to a different physical host involves transferring the
data volume and reinstalling the binary.

### Step 1 — Prepare the new host

Complete the [Setup & Prerequisites](setup.md) on the new machine:

- KVM/QEMU installed
- IOMMU enabled (if using PCI passthrough)
- Hugepages configured
- LVM tools installed
- Terminal with 80×25 minimum size

### Step 2 — Mount the DKVM data volume

Attach the storage device containing your DKVM data to the new host and mount
it:

```bash
# If using a filesystem labeled dkvmdata (DKVM convention)
mount LABEL=dkvmdata /media/dkvmdata

# Otherwise, mount by device
mount /dev/sdX1 /media/dkvmdata
```

Verify the mount:

```bash
mountpoint /media/dkvmdata
# /media/dkvmdata is a mount point
```

Your VM configurations, logs, ISOs, and scripts are now available on the new
host.

### Step 3 — Install DKVM Manager

Install the same version you were running on the old host:

```bash
# Option A: Download from GitHub Releases
VERSION="v0.2.0"
curl -LO "https://github.com/glemsom/dkvmmanager/releases/download/${VERSION}/dkvmmanager"
cp dkvmmanager /usr/local/bin/dkvmmanager
chmod +x /usr/local/bin/dkvmmanager

# Option B: Install via Alpine package
apk add dkvmmanager
```

### Step 4 — Re-apply host-specific configuration

Some settings are host-specific and need reconfiguration:

| Setting | How to reconfigure |
|---------|-------------------|
| PCI passthrough devices | Configuration tab → Edit PCI Passthrough (IOMMU groups differ per host) |
| vCPU pinning | Configuration tab → Edit vCPU Pinning (CPU topology differs per host) |
| CPU options | Configuration tab → Edit CPU Options |
| Network bridge | `~/.dkvmmanager.yaml` → `network_bridge` (default: `br0`) |
| SSH password | Configuration tab → Set SSH Password |

### Step 5 — Verify

Start DKVM Manager and confirm your VMs appear in the VMs tab:

```bash
dkvmmanager
```

Test-start a VM to verify QEMU launches correctly.

### Step 6 — Clean up old host (optional)

Once the new host is verified:

- Power down the old host
- Keep the old config as a backup, or remove it

---

## What Changes Between Versions

See the [CHANGELOG](../../CHANGELOG.md) for a complete list of additions,
fixes, and breaking changes in each release.

The changelog follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and groups changes by:

| Section | Meaning |
|---------|---------|
| **Added** | New features |
| **Fixed** | Bug fixes |
| **Changed** | Behaviour changes, documentation, internal refactoring |
| **Removed** | Removed features or APIs (breaking) |

Release tags follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html):

- **Patch** (`0.1.0` → `0.1.1`): bug fixes, no breaking changes
- **Minor** (`0.1.0` → `0.2.0`): new features, backward compatible
- **Major** (`0.2.0` → `1.0.0`): may include breaking changes

---

## See Also

- [Setup & Prerequisites](setup.md) — host requirements and first installation
- [FAQ](faq.md) — frequently asked questions
- [Troubleshooting](troubleshooting.md) — common issues and solutions
- [User Guide Index](README.md) — all documentation
