// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// View implements tea.Model
func (m *CPUOptionsFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// SetSize updates the form dimensions
func (m *CPUOptionsFormModel) SetSize(w, h int) {
	m.contentW = w
	m.contentH = h
	if !m.ready {
		m.vp = viewport.New(w, h)
		m.ready = true
	} else {
		m.vp.Width = w
		m.vp.Height = h
	}
	m.syncViewport()
}

// --- Rendering ---

var (
	cpuOptLabelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cpuOptFocusStyle   = styles.FormFocusStyle()
	cpuOptInputStyle   = styles.FormInputStyle()
	cpuOptErrorStyle   = styles.ErrorTextStyle()
	cpuOptMutedStyle   = styles.FormMutedStyle()
	cpuOptSaveStyle    = styles.FormSaveStyle()
	cpuOptSectionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Bold(true)
)

// renderAllLines produces the full list of output lines for the form
func (m *CPUOptionsFormModel) renderAllLines() []string {
	var lines []string

	// Section 1: Hypervisor Stealth
	lines = append(lines, cpuOptSectionStyle.Render("== Hypervisor Stealth =="))
	lines = append(lines, "")

	for i, pos := range m.positions[0:2] {
		focused := (i == m.focusIndex)
		lines = m.renderPosition(lines, pos, focused)
	}

	lines = append(lines, "")

	// Section 2: Hyper-V Enlightenments
	lines = append(lines, cpuOptSectionStyle.Render("== Hyper-V Enlightenments =="))
	lines = append(lines, "")

	for i, pos := range m.positions[2:17] {
		focused := (i+2 == m.focusIndex)
		lines = m.renderPosition(lines, pos, focused)
	}

	lines = append(lines, "")

	// Section 3: Advanced CPU Features
	lines = append(lines, cpuOptSectionStyle.Render("== Advanced CPU Features =="))
	lines = append(lines, "")

	for i, pos := range m.positions[17:23] {
		focused := (i+17 == m.focusIndex)
		lines = m.renderPosition(lines, pos, focused)
	}

	// Save button (last position)
	lines = append(lines, "")
	focused := (len(m.positions)-1 == m.focusIndex)
	lines = m.renderPosition(lines, m.positions[len(m.positions)-1], focused)

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

// renderPosition appends lines for one focus position
func (m *CPUOptionsFormModel) renderPosition(lines []string, pos cpuOptFocusPos, focused bool) []string {
	switch pos.kind {
	case cpuOptToggle:
		desc := m.fieldLabel(pos.fieldName)
		val := m.getToggleValue(pos.fieldName)
		lines = append(lines, m.renderToggle(desc, val, focused))
		return lines

	case cpuOptText:
		desc := m.fieldLabel(pos.fieldName)
		val := m.getTextValue(pos.fieldName)
		key := cpuOptPosKey(pos)
		cursor := m.effectiveCursor(key, val)
		lines = append(lines, m.renderTextInput(desc, val, cursor, focused))
		return lines

	case cpuOptSave:
		saveText := cpuOptMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
		if focused {
			saveText = cpuOptSaveStyle.Render("[Space/Enter] Save") + "    " + cpuOptMutedStyle.Render("[ESC] Cancel")
		}
		lines = append(lines, saveText)
		return lines
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
		"HVRelaxed":              "Disable watchdog timeouts",
		"HVReset":                "Guest reset capability",
		"HVRuntime":              "Track stolen CPU time",
		"HVSpinlocks":            "Paravirtualized spinlocks",
		"HVStimer":               "Synthetic timers",
		"HVSyncIC":               "Synthetic interrupt controller",
		"HVTime":                 "Fast clocksources",
		"HVVapic":                "Exit-less EOI processing",
		"HVVPIndex":              "Virtual CPU index",
		"HVNoNonarchCoresharing": "No non-arch core sharing",
		"HVTLBFlush":             "Paravirtualized TLB flush",
		"HVTLBFlushExt":          "Extended TLB flush ranges",
		"HVIPI":                  "Paravirtualized IPI",
		"HVAVIC":                 "Hardware APIC virtualization",
		"TopoExt":                "AMD topology extension",
		"L3Cache":                "Expose host L3 cache info",
		"X2APIC":                 "x2APIC mode (>255 vCPUs)",
		"Migratable":             "Disable vCPU migration",
		"InvTSC":                 "Invariant TSC",
		"RTCUTC":                 "Use UTC time for RTC",
	}
	if lbl, ok := labels[name]; ok {
		return lbl
	}
	return name
}
