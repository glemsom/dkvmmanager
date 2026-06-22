# Terminal Capabilities: DKVM Host Console (tty1, TERM=linux)

## Scope

Capabilities of the Linux virtual console (tty1) on the DKVM host at `192.168.50.21`.
This is the physical/SPICE console accessible on the host.

## Environment

| Property | Value |
|----------|-------|
| Host | `192.168.50.21` — DKVM host |
| Kernel | `6.18.35-0-lts` (Alpine) |
| Console driver | `simpledrm` (UEFI framebuffer) |
| Console font | Kernel built-in VGA font (8×16, 256 glyphs, CP437 encoding) |
| Font packages | None installed (`kbd` package absent, no `.psf`/`.psfu` fonts) |
| TERM | `linux` |
| UTF-8 mode | Supported via `\033%%G` escape |

## Character Support

Test methodology: write Unicode characters to `/dev/tty1`, read back rendered
glyph indices from `/dev/vcsa1`. Kernel maps Unicode → VGA font glyph via
built-in translation table. Characters not in the 256-glyph font fall back to
`+` (0x2B).

### Box-drawing: supported (in CP437 VGA font)

These render correctly on tty1 without any font package:

| Glyph | Unicode | VGA glyph | Notes |
|-------|---------|-----------|-------|
| `─` | U+2500 | 0xC4 | Light horizontal — **used by lipgloss.NormalBorder** |
| `│` | U+2502 | 0xB3 | Light vertical — **used by lipgloss.NormalBorder** |
| `┌` | U+250C | 0xDA | Light down and right — **used by lipgloss.NormalBorder** |
| `┐` | U+2510 | 0xBF | Light down and left — **used by lipgloss.NormalBorder** |
| `└` | U+2514 | 0xC0 | Light up and right — **used by lipgloss.NormalBorder** |
| `┘` | U+2518 | 0xD9 | Light up and left — **used by lipgloss.NormalBorder** |
| `├` | U+251C | 0xC3 | Light vertical and right — **used by lipgloss.NormalBorder** |
| `┤` | U+2524 | 0xB4 | Light vertical and left — **used by lipgloss.NormalBorder** |
| `┬` | U+252C | 0xC2 | Light down and horizontal |
| `┴` | U+2534 | 0xC1 | Light up and horizontal |
| `┼` | U+253C | 0xC5 | Light vertical and horizontal |

All `lipgloss.NormalBorder()` characters render correctly on TERM=linux.

### Box-drawing: NOT supported (not in CP437)

These fall back to `+` and must not be used on TERM=linux:

| Glyph | Unicode | Fallback | Notes |
|-------|---------|----------|-------|
| `╭` | U+256D | `+` | Light arc down and right — **used by lipgloss.RoundedBorder** |
| `╮` | U+256E | `+` | Light arc down and left — **used by lipgloss.RoundedBorder** |
| `╯` | U+256F | `+` | Light arc up and left — **used by lipgloss.RoundedBorder** |
| `╰` | U+2570 | `+` | Light arc up and right — **used by lipgloss.RoundedBorder** |

### Other symbols: supported (in CP437)

| Glyph | Unicode | VGA glyph |
|-------|---------|-----------|
| `●` | U+25CF | 0x07 (bullet) |
| `○` | U+25CB | 0x08 (circle) |
| `◑` | U+25D1 | 0x09 (circle half) — only some fonts |

## Implication

`NormalBorder()` (straight box-drawing ┌┐└┘├┤─│) works natively on TERM=linux.
No ASCII fallback needed.

`RoundedBorder()` (arcs ╭╮╰╯) does NOT work — requires ASCII fallback or
replacement with straight variants.

**Decision (ADR pending):** Use straight borders everywhere for consistent
appearance across all terminals. Eliminates the TERM=linux border fallback.
