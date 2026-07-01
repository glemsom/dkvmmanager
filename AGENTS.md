# DKVM Manager

TUI-based VM management for KVM/QEMU.

## Test
```bash
make test

# Run tests directly:
docker run --rm -w /build -v $(pwd):/build -e GOCACHE=/tmp/go-cache \
  --user $(id -u):$(id -g) \
  golang:1.26-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 \
  sh -c 'go test -v -run TestName ./internal/tui/models/...'
```

> `go test -race` needs CGO — use a non-alpine image with `CGO_ENABLED=1`.

## Build

```bash
make build   # via Docker (golang:1.26-alpine)

## Project layout

- `main.go` — entry point
- `internal/` — core packages (config, domain, hugepages, version, vm, tui)
- `examples/` — example scripts
- `VERSION` — current version

## Agent conventions

| Thing | Convention |
|-------|------------|
| Issues | GitHub Issues in `glemsom/dkvmmanager` — see `docs/agents/issue-tracker.md` |
| Triage labels | `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix` — see `docs/agents/triage-labels.md` |
| Domain docs | Single `CONTEXT.md` + `docs/adr/` — see `docs/agents/domain.md` |

See [CONTRIBUTING.md](./CONTRIBUTING.md) for full workflow.
