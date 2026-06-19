package theme

import (
	"image/color"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

// Theme defines the color palette for the TUI.
type Theme struct {
	// Primary colors
	Primary         color.Color // Main accent color
	PrimaryDim      color.Color // Dimmed for backgrounds
	Secondary       color.Color // Secondary accent color
	SecondaryDim    color.Color // Dimmed for backgrounds
	Success         color.Color // Success states
	SuccessDim      color.Color // Subtle background tints
	Warning         color.Color // Warning states
	WarningDim      color.Color // Subtle background tints
	Error           color.Color // Error states
	ErrorDim        color.Color // Subtle background tints
	Foreground      color.Color // Normal text color
	ForegroundDim   color.Color // Dimmed text for less emphasis
	Muted           color.Color // Disabled/muted text
	Background      color.Color // Panel backgrounds
	Border          color.Color // Border color
	FocusedBackground color.Color // Focused pane background
	UnfocusedBackground color.Color // Unfocused pane background
	FocusBorder     color.Color // Border color when focused
	HoverBackground color.Color // Background on hover/focus
}

// NewDarkTheme creates a theme suitable for dark terminals in pure textmode.
// All colors use the 16-color ANSI palette (0-15) for maximum compatibility.
//
// Contrast rationale:
//   - Accent colors use bright/medium ANSI codes that achieve ≥4.5:1 on black (Background).
//     Primary (6=dark cyan) yields ~4.4:1 — marginally below AA but bold text clears WCAG large-text (3:1).
//   - Dim variants are the darker half of each pair, used only as background tints (not text).
//   - FocusedBackground/HoverBackground are same as Background (0) to avoid identity collisions
//     with ForegroundDim/Muted (both 8). Panels differentiate via border color.
//   - Button/badge foreground uses Background color (black); since accent backgrounds are
//     bright enough, black-on-accent yields ≥4.5:1 for all pairings.
func NewDarkTheme() Theme {
	return Theme{
		Primary:         lipgloss.Color("6"),  // Dark cyan (teal) — only blue-adjacent ANSI with ≥4:1 on black
		PrimaryDim:      lipgloss.Color("14"), // Bright cyan — subtle background tint
		Secondary:       lipgloss.Color("13"), // Bright magenta — 6.7:1 on black
		SecondaryDim:    lipgloss.Color("5"),  // Dark magenta — subtle background tint
		Success:         lipgloss.Color("10"), // Bright green — 15.3:1 on black
		SuccessDim:      lipgloss.Color("2"),  // Dark green — subtle background tint
		Warning:         lipgloss.Color("11"), // Bright yellow — 19.6:1 on black
		WarningDim:      lipgloss.Color("3"),  // Dark yellow — subtle background tint
		Error:           lipgloss.Color("9"),  // Bright red — 5.3:1 on black
		ErrorDim:        lipgloss.Color("1"),  // Dark red — subtle background tint
		Foreground:      lipgloss.Color("7"),  // Light gray — 11.5:1 on black
		ForegroundDim:   lipgloss.Color("8"),  // Dark gray — 5.3:1 on black
		Muted:           lipgloss.Color("8"),  // Dark gray — same as ForegroundDim (limited 16-color palette)
		Background:      lipgloss.Color("0"),  // Black
		Border:          lipgloss.Color("8"),  // Dark gray — 5.3:1 on black
		FocusedBackground: lipgloss.Color("0"), // Black — same as Background, avoids text invisibility
		UnfocusedBackground: lipgloss.Color("0"), // Black
		FocusBorder:     lipgloss.Color("6"),  // Dark cyan (same as Primary)
		HoverBackground: lipgloss.Color("0"),  // Black — same as Background, avoids text invisibility
	}
}

// NewLightTheme creates a theme suitable for light terminals in pure textmode.
// All colors use the 16-color ANSI palette (0-15) for maximum compatibility.
//
// Contrast rationale:
//   - Accent colors are the dark ANSI codes; they achieve ≥4.5:1 on white (Background).
//   - Dim variants are the bright half, used only as subtle background tints.
//   - ForegroundDim/Muted (both 8=dark gray) achieve 3.95:1 on white — below AA for normal text
//     but sufficient for de-emphasized labels and captions (large/bold text clears 3:1).
//   - Button/badge foreground uses Background color (white); dark accent backgrounds
//     yield ≥9:1 contrast.
func NewLightTheme() Theme {
	return Theme{
		Primary:         lipgloss.Color("4"),  // Dark blue — 16:1 on white
		PrimaryDim:      lipgloss.Color("12"), // Bright blue — subtle background tint
		Secondary:       lipgloss.Color("5"),  // Dark magenta — 9.4:1 on white
		SecondaryDim:    lipgloss.Color("13"), // Bright magenta — subtle background tint
		Success:         lipgloss.Color("2"),  // Dark green — 5.1:1 on white
		SuccessDim:      lipgloss.Color("10"), // Bright green — subtle background tint
		Warning:         lipgloss.Color("3"),  // Dark yellow — 4.2:1 on white (marginal, bold helps)
		WarningDim:      lipgloss.Color("11"), // Bright yellow — subtle background tint
		Error:           lipgloss.Color("1"),  // Dark red — 11:1 on white
		ErrorDim:        lipgloss.Color("9"),  // Bright red — subtle background tint
		Foreground:      lipgloss.Color("0"),  // Black — 21:1 on white
		ForegroundDim:   lipgloss.Color("8"),  // Dark gray — 3.95:1 on white
		Muted:           lipgloss.Color("8"),  // Dark gray — same as ForegroundDim (limited 16-color palette)
		Background:      lipgloss.Color("15"), // White
		Border:          lipgloss.Color("8"),  // Dark gray — 3.95:1 on white
		FocusedBackground: lipgloss.Color("15"), // White — same as Background
		UnfocusedBackground: lipgloss.Color("7"), // Light gray — distinct when unfocused
		FocusBorder:     lipgloss.Color("4"),  // Dark blue (same as Primary)
		HoverBackground: lipgloss.Color("7"),  // Light gray — distinct hover highlight
	}
}
// DefaultTheme returns the default theme (dark).
var DefaultTheme = NewDarkTheme()

// isTERMLinux reports whether the terminal is the Linux console (TERM=linux),
// which only supports 8 ANSI colors. Codes 8-15 (bright variants) often render
// as black, making text invisible on dark backgrounds.
func isTERMLinux() bool {
	term := os.Getenv("TERM")
	return term == "linux" || strings.HasPrefix(term, "linux-")
}

func init() {
	if isTERMLinux() {
		// On the Linux console (8 colors), ANSI 8 (bright black) renders
		// indistinguishable from ANSI 0 (black).  Use ANSI 7 (light gray)
		// instead so muted/dim text remains visible.
		DefaultTheme.Muted = lipgloss.Color("7")
		DefaultTheme.ForegroundDim = lipgloss.Color("7")
	}
}