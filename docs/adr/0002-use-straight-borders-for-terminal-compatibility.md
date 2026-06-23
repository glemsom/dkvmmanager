# Use straight borders for terminal compatibility

**Status:** accepted

**Date:** 2026-06-23

**Driver:** @dkvmmanager (terminal analysis)

## Context

DKVM Manager runs on a dedicated host whose primary console is the Linux virtual
console (tty1, `TERM=linux`). On this terminal the kernel VGA font (CP437
encoding, 256 glyphs) is used — no external font packages are installed.

The project uses [lipgloss](https://github.com/charmbracelet/lipgloss) v2 for
terminal styling, which provides two border presets:

| Border type     | Glyphs                      | Unicode codepoints |
|-----------------|-----------------------------|--------------------|
| `NormalBorder`  | Straight box-drawing ┌┐└┘  | U+2500–U+257F      |
| `RoundedBorder` | Arcs ╭╮╰╯                  | U+256D–U+2570      |

Full character-level testing (see
[terminal-capabilities.md](../terminal-capabilities.md)) confirmed:

- All `NormalBorder` characters exist in CP437 and render on `TERM=linux`.
- `RoundedBorder` arc characters do **not** exist in CP437. On `TERM=linux`
  they fall back to `+` (0x2B), producing a visually broken border.

When the TUI was first written, `RoundedBorder` was used in several places.
An ASCII-fallback layer (`internal/tui/styles/ascii_fallback.go`) was added to
detect `TERM=linux` and swap to `NormalBorder`. This introduced a runtime
branch on every border render and still produced visual inconsistency — in
a mixed TERM environment (e.g. SSH from a modern terminal) some borders would
round and others would not.

## Decision

Use `lipgloss.NormalBorder()` everywhere. No ASCII fallback is needed.

- `styles.RoundedBorder()` was kept as a compatibility wrapper that simply
  returns `lipgloss.NormalBorder()`.
- `styles.GetBorder()` passes through any border unchanged (no runtime branch).
- All border styles in `internal/tui/styles/styles.go` use `NormalBorder()`.
- The runtime `TERM=linux` branch in `ascii_fallback.go` is retained only for
  mode icons, status symbols, and disk bullets — places where the Unicode
  alternatives (●, ○, ◑) are outside CP437.

## Consequences

- **Consistent appearance** across all terminals (SSH, SPICE console, tty1).
  No more mixed rounded/straight borders.
- **Simpler code.** The `ascii_fallback.go` border wrappers are empty pass-
  throughs. No conditional logic per render.
- **Slightly less aesthetic** on modern terminals that support the arc
  characters, but visually negligible — straight box-drawing has been the
  terminal standard for decades.
- **Eliminates a class of rendering bugs.** Users no longer see `++++` where
  borders should be.

## Related

- [Terminal Capabilities](../terminal-capabilities.md) — full character-level
  analysis of the DKVM host console.
- `internal/tui/styles/styles.go` — border style definitions.
- `internal/tui/styles/ascii_fallback.go` — compatibility wrappers.
