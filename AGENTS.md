# DKVM Manager

DKVM Manager is a TUI-based virtual machine management tool for KVM/QEMU.

## Development

This project uses Go modules and follows standard Go practices.

**Essential references:**
- **Development workflow, testing, and build instructions:** See [CONTRIBUTING.md](./CONTRIBUTING.md)

## Quick Start

```bash
# Build
make build

# Test
make test
```

## Project Structure

- `main.go` - Entry point
- `internal/` - Core packages (config, vm, tui, etc.)
- `examples/` - Example scripts

## Version
For current version, see [VERSION](./VERSION)

## Agent skills

### Issue tracker

Issues live as GitHub Issues in `glemsom/dkvmmanager`. See `docs/agents/issue-tracker.md`.

### Triage labels

Default five-canonical label vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context — one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.