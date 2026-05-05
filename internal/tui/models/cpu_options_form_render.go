// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// Style variables for the CPU options form.
var (
	cpuOptLabelStyle   = lipgloss.NewStyle().Foreground(styles.Colors.ForegroundDim)
	cpuOptFocusStyle   = styles.FormFocusStyle()
	cpuOptInputStyle   = styles.FormInputStyle()
	cpuOptErrorStyle   = styles.ErrorTextStyle()
	cpuOptMutedStyle   = styles.FormMutedStyle()
	cpuOptSaveStyle    = styles.FormSaveStyle()
	cpuOptSectionStyle = styles.DetailSectionStyle()
)

// --- FormModel Interface Implementation ---

// RenderHeader returns the form header.
func (m *CPUOptionsFormModel) RenderHeader() string {
	return cpuOptFocusStyle.Render("CPU Options")
}

// RenderFooter returns the form footer.
func (m *CPUOptionsFormModel) RenderFooter() string {
	var parts []string
	if m.statusMessage != "" {
		parts = append(parts, "")
		parts = append(parts, cpuOptErrorStyle.Render(m.statusMessage))
	}
	if errMsg, ok := m.errors["save"]; ok {
		parts = append(parts, "")
		parts = append(parts, cpuOptErrorStyle.Render("Error: "+errMsg))
	}
	parts = append(parts, "")
	parts = append(parts, cpuOptMutedStyle.Render("Tab/Shift+Tab Navigate  PgUp/PgDown Page  Space/Enter Toggle/Save  ESC Cancel"))
	return strings.Join(parts, "\n")
}

// RenderPosition returns the markup for a single position.
func (m *CPUOptionsFormModel) RenderPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
	switch pos.Kind {
	case form.FocusHeader:
		return cpuOptSectionStyle.Render(pos.Label)

	case form.FocusToggle:
		val := m.getToggleValue(pos.Key)
		return m.renderToggle(pos.Label, val, focused)

	case form.FocusText:
		val := m.getTextValue(pos.Key)
		cursor := m.effectiveCursor(pos.Key, val)
		if focused && cursorOffset >= 0 {
			cursor = cursorOffset
		}
		return m.renderTextInput(pos.Label, val, cursor, focused)

	case form.FocusButton:
		if pos.Key == "save" {
			saveText := cpuOptMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
			if focused {
				saveText = cpuOptSaveStyle.Render("[Space/Enter] Save") + "    " + cpuOptMutedStyle.Render("[ESC] Cancel")
			}
			return saveText
		}
	}
	return ""
}

// --- Backward-compatible rendering helpers ---

// renderAllLines produces the full list of output lines for the viewport (backward compat).
func (m *CPUOptionsFormModel) renderAllLines() []string {
	var lines []string

	// Header
	lines = append(lines, cpuOptFocusStyle.Render("CPU Options"))
	lines = append(lines, "")

	// Render all positions (including section headers)
	for i, pos := range m.positions {
		if pos.Kind == form.FocusHeader {
			// Blank line before section header
			lines = append(lines, "")
		}
		focused := (i == m.focusIndex)
		lines = m.renderPositionLine(lines, pos, focused)
	}

	// Save error at the bottom
	if errMsg, ok := m.errors["save"]; ok {
		lines = append(lines, "")
		lines = append(lines, cpuOptErrorStyle.Render("Error: "+errMsg))
	}

	// Footer
	lines = append(lines, "")
	lines = append(lines, cpuOptMutedStyle.Render("Tab/Shift+Tab Navigate  PgUp/PgDown Page  Space/Enter Toggle/Save  ESC Cancel"))

	return lines
}

// renderPositionLine appends lines for one focus position to the lines slice.
func (m *CPUOptionsFormModel) renderPositionLine(lines []string, pos form.FocusPos, focused bool) []string {
	switch pos.Kind {
	case form.FocusHeader:
		lines = append(lines, cpuOptSectionStyle.Render(pos.Label))
		return lines

	case form.FocusToggle:
		val := m.getToggleValue(pos.Key)
		lines = append(lines, m.renderToggle(pos.Label, val, focused))
		return lines

	case form.FocusText:
		val := m.getTextValue(pos.Key)
		cursor := m.effectiveCursor(pos.Key, val)
		lines = append(lines, m.renderTextInput(pos.Label, val, cursor, focused))
		return lines

	case form.FocusButton:
		if pos.Key == "save" {
			saveText := cpuOptMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
			if focused {
				saveText = cpuOptSaveStyle.Render("[Space/Enter] Save") + "    " + cpuOptMutedStyle.Render("[ESC] Cancel")
			}
			lines = append(lines, saveText)
			return lines
		}
	}

	return lines
}

// renderToggle renders a toggle as "[ON] Description" or "[OFF] Description"
func (m *CPUOptionsFormModel) renderToggle(desc string, value bool, focused bool) string {
	prefix := "  "
	if focused {
		prefix = cpuOptFocusStyle.Render("> ")
	}

	var togglePart string
	if value {
		if focused {
			togglePart = cpuOptFocusStyle.Render("[ON]")
		} else {
			togglePart = cpuOptInputStyle.Render("[ON]")
		}
	} else {
		if focused {
			togglePart = cpuOptFocusStyle.Render("[OFF]")
		} else {
			togglePart = cpuOptMutedStyle.Render("[OFF]")
		}
	}

	return prefix + togglePart + " " + cpuOptLabelStyle.Render(desc)
}

// renderTextInput renders a text input as "[ value ] Description"
func (m *CPUOptionsFormModel) renderTextInput(desc, value string, cursor int, focused bool) string {
	prefix := "  "
	if focused {
		prefix = cpuOptFocusStyle.Render("> ")
	}

	var inputPart string
	if focused {
		if cursor < len(value) {
			before := value[:cursor]
			at := string(value[cursor])
			after := ""
			if cursor+1 < len(value) {
				after = value[cursor+1:]
			}
			inputPart = cpuOptInputStyle.Render("[ "+before) +
				lipgloss.NewStyle().Reverse(true).Render(at) +
				cpuOptInputStyle.Render(after+" ]")
		} else {
			inputPart = cpuOptInputStyle.Render("[ "+value) + cpuOptFocusStyle.Render("_ ]")
		}
	} else {
		if value == "" {
			inputPart = cpuOptMutedStyle.Render("[ (empty) ]")
		} else {
			inputPart = cpuOptInputStyle.Render("[ " + value + " ]")
		}
	}

	return prefix + inputPart + " " + cpuOptLabelStyle.Render(desc)
}

// fieldLabel returns the user-friendly description for a field name
func (m *CPUOptionsFormModel) fieldLabel(name string) string {
	labels := map[string]string{
		"HideKVM":                "Hides VM from guest",
		"VendorID":               "Custom hypervisor vendor ID",
		"HVFrequency":            "Expose TSC/APIC frequencies",
		"HVRelaxed":              "Relaxed timing checks",
		"HVReset":                "Guest reset capability",
		"HVRuntime":              "Hypervisor runtime info",
		"HVSpinlocks":            "Paravirtualized spinlocks",
		"HVStimer":               "Synthetic timers",
		"HVSyncIC":               "Synthetic interrupt controller",
		"HVTime":                 "Reference TSC page",
		"HVVapic":                "Exit-less EOI processing",
		"HVVPIndex":              "Virtual CPU index",
		"HVNoNonarchCoresharing": "SMT perf counter isolation",
		"HVTLBFlush":             "Paravirtualized TLB flush",
		"HVTLBFlushExt":          "Extended TLB flush ranges",
		"HVIPI":                  "Paravirtualized IPI",
		"HVAVIC":                 "Hyper-V nested APIC virt",
		"TopoExt":                "AMD topology extension",
		"L3Cache":                "Expose host L3 cache info",
		"X2APIC":                 "x2APIC mode (>255 vCPUs)",
		"Migratable":             "Expose all host features (no live migration)",
		"InvTSC":                 "Invariant TSC",
		"RTCUTC":                 "Use UTC time for RTC",
		"CPUPM":                  "Allow guest C/P-state control",
	}
	if lbl, ok := labels[name]; ok {
		return lbl
	}
	return name
}
