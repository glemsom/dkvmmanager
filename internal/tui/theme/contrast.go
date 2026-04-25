package theme

import (
	"fmt"
	"math"
)

// ContrastResult holds the result of a contrast ratio check.
type ContrastResult struct {
	Name        string
	Foreground  string
	Background  string
	Ratio       float64
	Required    float64
	Passed      bool
	Level       string // "AA" or "AAA"
}

// hexToRGB converts a hex color string (#RRGGBB) to RGB components (0-255).
func hexToRGB(hex string) (r, g, b uint8) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0
	}
	r = hexToByte(hex[1:3])
	g = hexToByte(hex[3:5])
	b = hexToByte(hex[5:7])
	return
}

func hexToByte(s string) uint8 {
	var b uint8
	fmt.Sscanf(s, "%02x", &b)
	return b
}

// sRGBtoLinear converts an sRGB component (0-255) to linear light (0-1).
func sRGBtoLinear(c uint8) float64 {
	s := float64(c) / 255.0
	if s <= 0.03928 {
		return s / 12.92
	}
	return math.Pow((s+0.055)/1.055, 2.4)
}

// RelativeLuminance computes the relative luminance of an RGB color per WCAG 2.0.
func RelativeLuminance(r, g, b uint8) float64 {
	rl := 0.2126*sRGBtoLinear(r) + 0.7152*sRGBtoLinear(g) + 0.0722*sRGBtoLinear(b)
	return rl
}

// ContrastRatio computes the WCAG contrast ratio between two hex colors.
func ContrastRatio(fg, bg string) float64 {
	fr, fg_, fb := hexToRGB(fg)
	br, bg_, bb := hexToRGB(bg)

	l1 := RelativeLuminance(fr, fg_, fb)
	l2 := RelativeLuminance(br, bg_, bb)

	if l1 < l2 {
		l1, l2 = l2, l1
	}

	return (l1 + 0.05) / (l2 + 0.05)
}

// CheckContrast checks a single color pair against WCAG AA (4.5:1) for normal text.
func CheckContrast(name, fg, bg string) ContrastResult {
	ratio := ContrastRatio(fg, bg)
	required := 4.5 // WCAG AA for normal text
	return ContrastResult{
		Name:       name,
		Foreground: fg,
		Background: bg,
		Ratio:      math.Round(ratio*100) / 100,
		Required:   required,
		Passed:     ratio >= required,
		Level:      "AA",
	}
}

// CheckLargeTextContrast checks a color pair against WCAG AA (3:1) for large text.
func CheckLargeTextContrast(name, fg, bg string) ContrastResult {
	ratio := ContrastRatio(fg, bg)
	required := 3.0 // WCAG AA for large text (18pt+ or 14pt+ bold)
	return ContrastResult{
		Name:       name,
		Foreground: fg,
		Background: bg,
		Ratio:      math.Round(ratio*100) / 100,
		Required:   required,
		Passed:     ratio >= required,
		Level:      "AA (Large)",
	}
}

// Report formats a contrast check report.
func Report(results []ContrastResult) string {
	var s string
	passed := 0
	failed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}
	s += fmt.Sprintf("=== WCAG Contrast Report ===\n")
	s += fmt.Sprintf("Total: %d | Passed: %d | Failed: %d\n\n", len(results), passed, failed)

	for _, r := range results {
		status := "PASS"
		if !r.Passed {
			status = "FAIL"
		}
		s += fmt.Sprintf("[%s] %-40s %s on %s = %.2f:1 (req: %.1f:1 %s)\n",
			status, r.Name, r.Foreground, r.Background, r.Ratio, r.Required, r.Level)
	}
	return s
}
