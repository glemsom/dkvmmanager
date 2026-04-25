package theme

import (
	"testing"
)

// TestWCAGContrastAllCombinations verifies that all foreground/background
// combinations used across the TUI meet WCAG AA contrast requirements.
// WCAG AA requires 4.5:1 for normal text, 3:1 for large text (18pt+ / 14pt+ bold).
// Since TUI font sizes are user-controlled, we aim for 4.5:1 across the board.
func TestWCAGContrastAllCombinations(t *testing.T) {
	theme := NewDarkTheme()

	// Helper: extract hex from lipgloss.Color string.
	// lipgloss.Color("#RRGGBB") stores the string as-is.
	results := []ContrastResult{}

	check := func(name, fg, bg string) {
		results = append(results, CheckContrast(name, fg, bg))
	}

	// ---- Base text on base backgrounds ----
	check("Foreground on Background", string(theme.Foreground), string(theme.Background))
	check("Foreground on FocusedBackground", string(theme.Foreground), string(theme.FocusedBackground))
	check("Foreground on UnfocusedBackground", string(theme.Foreground), string(theme.UnfocusedBackground))
	check("ForegroundDim on Background", string(theme.ForegroundDim), string(theme.Background))
	check("ForegroundDim on FocusedBackground", string(theme.ForegroundDim), string(theme.FocusedBackground))
	check("Muted on Background", string(theme.Muted), string(theme.Background))
	check("Muted on FocusedBackground", string(theme.Muted), string(theme.FocusedBackground))

	// ---- Accent colors on backgrounds ----
	check("Primary on Background", string(theme.Primary), string(theme.Background))
	check("Primary on FocusedBackground", string(theme.Primary), string(theme.FocusedBackground))
	check("Secondary on Background", string(theme.Secondary), string(theme.Background))
	check("Success on Background", string(theme.Success), string(theme.Background))
	check("Warning on Background", string(theme.Warning), string(theme.Background))
	check("Error on Background", string(theme.Error), string(theme.Background))

	// ---- Inverse text (bg color on fg color) ----
	check("Background on Primary (inverse/badge text)", string(theme.Background), string(theme.Primary))
	check("Background on Secondary (badge text)", string(theme.Background), string(theme.Secondary))

	// ---- Section headers / labels on dim backgrounds ----
	check("Primary on PrimaryDim (section headers)", string(theme.Primary), string(theme.PrimaryDim))
	check("Secondary on SecondaryDim", string(theme.Secondary), string(theme.SecondaryDim))
	check("Success on SuccessDim (running status bg)", string(theme.Success), string(theme.SuccessDim))
	check("Warning on WarningDim", string(theme.Warning), string(theme.WarningDim))
	check("Error on ErrorDim (error status bg)", string(theme.Error), string(theme.ErrorDim))

	// ---- Status indicator colors ----
	check("Muted on Background (stopped indicator)", string(theme.Muted), string(theme.Background))
	check("Stopped (Muted) on Background", string(theme.Muted), string(theme.Background))

	// ---- Border colors on backgrounds ----
	check("Border on Background", string(theme.Border), string(theme.Background))
	check("FocusBorder on Background", string(theme.FocusBorder), string(theme.Background))
	check("Primary (border) on Background", string(theme.Primary), string(theme.Background))

	// ---- List items ----
	check("ForegroundDim on HoverBackground (list normal)", string(theme.ForegroundDim), string(theme.HoverBackground))
	check("Primary on HoverBackground (list selected)", string(theme.Primary), string(theme.HoverBackground))

	// ---- Button text ----
	check("Background on Primary (primary button text)", string(theme.Background), string(theme.Primary))
	check("Background on Secondary (secondary button text)", string(theme.Background), string(theme.Secondary))
	check("Muted on Background (disabled button text)", string(theme.Muted), string(theme.Background))

	// ---- Progress bar ----
	check("Success (fill) on Background (progress bar)", string(theme.Success), string(theme.Background))
	check("Muted (empty) on Background (progress bar)", string(theme.Muted), string(theme.Background))
	check("Foreground on Muted (progress bar percentage)", string(theme.Foreground), string(theme.Muted))

	// ---- Form elements ----
	check("Primary on Background (form input)", string(theme.Primary), string(theme.Background))
	check("Muted on Background (form label)", string(theme.Muted), string(theme.Background))
	check("Success on Background (form save)", string(theme.Success), string(theme.Background))

	// ---- Detail panel ----
	check("Muted on Background (detail label)", string(theme.Muted), string(theme.Background))
	check("Foreground on Background (detail value)", string(theme.Foreground), string(theme.Background))

	// ---- Code style ----
	check("Secondary on Background (code style)", string(theme.Secondary), string(theme.Background))

	// ---- Table ----
	check("Primary on Background (table header)", string(theme.Primary), string(theme.Background))
	check("Foreground on Background (table row)", string(theme.Foreground), string(theme.Background))

	// Run the report
	report := Report(results)
	t.Log("\n" + report)

	// Check for failures
	failed := []ContrastResult{}
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, r)
		}
	}

	if len(failed) > 0 {
		t.Errorf("\n%d contrast check(s) failed against WCAG AA (4.5:1):\n", len(failed))
		for _, r := range failed {
			t.Errorf("  FAIL: %s — %s on %s = %.2f:1 (need %.1f:1)",
				r.Name, r.Foreground, r.Background, r.Ratio, r.Required)
		}
	}
}

// TestWCAGContrastLargeText verifies that all combinations meet at least
// the large-text WCAG AA threshold (3:1). This is a softer check for
// decorative/bold elements.
func TestWCAGContrastLargeText(t *testing.T) {
	theme := NewDarkTheme()

	results := []ContrastResult{}

	check := func(name, fg, bg string) {
		results = append(results, CheckLargeTextContrast(name, fg, bg))
	}

	// All text combinations (same as above, but checking large text threshold)
	check("Foreground on Background", string(theme.Foreground), string(theme.Background))
	check("ForegroundDim on Background", string(theme.ForegroundDim), string(theme.Background))
	check("Muted on Background", string(theme.Muted), string(theme.Background))
	check("Primary on PrimaryDim", string(theme.Primary), string(theme.PrimaryDim))
	check("Success on SuccessDim", string(theme.Success), string(theme.SuccessDim))
	check("Error on ErrorDim", string(theme.Error), string(theme.ErrorDim))
	check("Warning on WarningDim", string(theme.Warning), string(theme.WarningDim))
	check("Secondary on SecondaryDim", string(theme.Secondary), string(theme.SecondaryDim))
	check("Foreground on Muted (progress bar %)", string(theme.Foreground), string(theme.Muted))

	report := Report(results)
	t.Log("\n" + report)

	failed := []ContrastResult{}
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, r)
		}
	}

	if len(failed) > 0 {
		t.Errorf("\n%d contrast check(s) failed against WCAG AA Large Text (3:1):\n", len(failed))
		for _, r := range failed {
			t.Errorf("  FAIL: %s — %s on %s = %.2f:1 (need %.1f:1)",
				r.Name, r.Foreground, r.Background, r.Ratio, r.Required)
		}
	}
}

// TestWCAGContrastLightTheme verifies the light theme meets WCAG AA.
func TestWCAGContrastLightTheme(t *testing.T) {
	theme := NewLightTheme()

	results := []ContrastResult{}

	check := func(name, fg, bg string) {
		results = append(results, CheckContrast(name, fg, bg))
	}

	// Base text
	check("Foreground on Background", string(theme.Foreground), string(theme.Background))
	check("ForegroundDim on Background", string(theme.ForegroundDim), string(theme.Background))
	check("Muted on Background", string(theme.Muted), string(theme.Background))

	// Accent colors
	check("Primary on Background", string(theme.Primary), string(theme.Background))
	check("Secondary on Background", string(theme.Secondary), string(theme.Background))
	check("Success on Background", string(theme.Success), string(theme.Background))
	check("Warning on Background", string(theme.Warning), string(theme.Background))
	check("Error on Background", string(theme.Error), string(theme.Background))

	// Inverse / badges
	check("Background on Primary", string(theme.Background), string(theme.Primary))
	check("Background on Secondary", string(theme.Background), string(theme.Secondary))

	// Section headers
	check("Primary on PrimaryDim", string(theme.Primary), string(theme.PrimaryDim))
	check("Success on SuccessDim", string(theme.Success), string(theme.SuccessDim))
	check("Warning on WarningDim", string(theme.Warning), string(theme.WarningDim))
	check("Error on ErrorDim", string(theme.Error), string(theme.ErrorDim))

	// Progress bar
	check("Foreground on Muted", string(theme.Foreground), string(theme.Muted))

	report := Report(results)
	t.Log("\n" + report)

	failed := []ContrastResult{}
	for _, r := range results {
		if !r.Passed {
			failed = append(failed, r)
		}
	}

	if len(failed) > 0 {
		t.Errorf("\n%d contrast check(s) failed for light theme against WCAG AA (4.5:1):\n", len(failed))
		for _, r := range failed {
			t.Errorf("  FAIL: %s — %s on %s = %.2f:1 (need %.1f:1)",
				r.Name, r.Foreground, r.Background, r.Ratio, r.Required)
		}
	}
}
