package styles

import (
	"os"
	"strings"
)

// asciiModeIcons maps mode names to ASCII fallback symbols for TERM=linux.
var asciiModeIcons = map[string]string{
	"Ready":   "-",
	"Editing": "*",
	"Loading": "-",
	"Running": ">",
	"Stopped": "#",
	"Error":   "!",
}

// isTermLinux reports whether the terminal is the Linux console (TERM=linux).
var statusAsciiSymbols = map[string]string{
	"running": "*",
	"stopped": "o",
	"error":   "*",
	"unknown": "o",
}

var statusUnicodeSymbols = map[string]string{
	"running": "●",
	"stopped": "○",
	"error":   "●",
	"unknown": "○",
}

// statusSymbol returns the status symbol for the given status, using ASCII
// fallback on TERM=linux (vgacon) where Unicode glyphs may not render.
func statusSymbol(status string) string {
	if isTermLinux() {
		if s, ok := statusAsciiSymbols[status]; ok {
			return s
		}
		return "o"
	}
	if s, ok := statusUnicodeSymbols[status]; ok {
		return s
	}
	return "○"
}

// isTermLinux reports whether the terminal is the Linux console (TERM=linux).
func isTermLinux() bool {
	term := os.Getenv("TERM")
	return term == "linux" || strings.HasPrefix(term, "linux-")
}

// DiskBullet returns the bullet symbol for hard disk listings, using ASCII
// fallback (*) on TERM=linux where Unicode bullets (●) may not render.
func DiskBullet() string {
	if isTermLinux() {
		return "*"
	}
	return "●"
}

// CDROMBullet returns the bullet symbol for CDROM listings, using ASCII
// fallback (o) on TERM=linux where Unicode (◑) may not render.
func CDROMBullet() string {
	if isTermLinux() {
		return "o"
	}
	return "◑"
}

// GetModeIcon returns the mode icon for the given mode, using ASCII fallback
// on TERM=linux (vgacon) where Unicode glyphs may not render.
func GetModeIcon(mode string) string {
	if isTermLinux() {
		if icon, ok := asciiModeIcons[mode]; ok {
			return icon
		}
		return "-"
	}
	if icon, ok := ModeIcons[mode]; ok {
		return icon
	}
	return "◌"
}
