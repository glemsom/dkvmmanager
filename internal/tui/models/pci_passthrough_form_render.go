// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// pciLabelStyle is the label style for form fields
var pciLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

// pciFocusStyle is the focused field style
var pciFocusStyle = styles.FormFocusStyle()

// pciInputStyle is the input field style
var pciInputStyle = styles.FormInputStyle()

// pciErrorStyle is the error text style
var pciErrorStyle = styles.ErrorTextStyle()

// pciMutedStyle is the muted/form helper text style
var pciMutedStyle = styles.FormMutedStyle()

// pciSaveStyle is the save button style
var pciSaveStyle = styles.FormSaveStyle()

// pciGPUStyle is the GPU tag style
var pciGPUStyle = styles.FormFocusStyle()

// pciUSBStyle is the USB tag style
var pciUSBStyle = styles.FormFocusStyle()

// pciWarnStyle is the warning text style
var pciWarnStyle = styles.WarningTextStyle()

// renderAllLines produces the full list of output lines for the form
func (m *PCIPassthroughFormModel) renderAllLines() []string {
	var lines []string

	if m.scanErr != nil {
		lines = append(lines, pciErrorStyle.Render(fmt.Sprintf("Warning: PCI scan failed: %s", m.scanErr)))
		lines = append(lines, pciWarnStyle.Render("Devices cannot be configured without a scan."))
		lines = append(lines, "")
	}

	if len(m.devices) == 0 {
		lines = append(lines, pciMutedStyle.Render("No PCI devices found on this system."))
		lines = append(lines, "")
		saveText := pciMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
		if m.currentPos().kind == pciSave && m.focusIndex == len(m.positions)-1 {
			saveText = pciSaveStyle.Render("[Space/Enter] Save") + "    " + pciMutedStyle.Render("[ESC] Cancel")
		}
		lines = append(lines, saveText)
		return lines
	}

	// Render each position
	for i, pos := range m.positions {
		focused := (i == m.focusIndex)
		lines = m.renderPosition(lines, pos, focused)
	}

	// Save error at the bottom
	if errMsg, ok := m.errors["save"]; ok {
		lines = append(lines, "")
		lines = append(lines, pciErrorStyle.Render("Error: "+errMsg))
	}

	// Validation warnings
	for _, w := range m.warnings {
		lines = append(lines, "")
		lines = append(lines, pciWarnStyle.Render("Warning: "+w))
	}

	// Footer
	lines = append(lines, "")
	lines = append(lines, pciMutedStyle.Render("Tab Navigate  Space/Enter Toggle  ESC Cancel"))

	return lines
}

// renderPosition appends lines for one focus position
func (m *PCIPassthroughFormModel) renderPosition(lines []string, pos pciFocusPos, focused bool) []string {
	switch pos.kind {
	case pciToggle:
		dev := m.getDeviceByAddr(pos.deviceAddr)
		if dev == nil {
			lines = append(lines, "  ???")
			return lines
		}
		selected := m.selected[pos.deviceAddr]
		lines = append(lines, m.renderDeviceToggle(dev, selected, focused))
		return lines

	case pciROMPath:
		val := m.romPaths[pos.deviceAddr]
		key := fmt.Sprintf("rom_%s", pos.deviceAddr)
		cursor := m.effectiveCursor(key, val)
		lines = append(lines, m.renderROMPath(pos.deviceAddr, val, cursor, focused))
		return lines

	case pciSave:
		lines = append(lines, "")
		saveText := pciMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
		if focused {
			saveText = pciSaveStyle.Render("[Space/Enter] Save") + "    " + pciMutedStyle.Render("[ESC] Cancel")
		}
		lines = append(lines, saveText)
		return lines
	}

	return lines
}

// renderDeviceToggle renders a PCI device as a toggle line
func (m *PCIPassthroughFormModel) renderDeviceToggle(dev *models.PCIDevice, selected, focused bool) string {
	prefix := "  "
	if focused {
		prefix = pciFocusStyle.Render("> ")
	}

	// Toggle indicator
	var togglePart string
	if selected {
		if focused {
			togglePart = pciFocusStyle.Render("[X]")
		} else {
			togglePart = pciInputStyle.Render("[X]")
		}
	} else {
		if focused {
			togglePart = pciFocusStyle.Render("[ ]")
		} else {
			togglePart = pciMutedStyle.Render("[ ]")
		}
	}

	// Device type tag
	var tag string
	if dev.IsGPU {
		tag = pciGPUStyle.Render("[GPU]")
	} else if dev.IsUSB {
		tag = pciUSBStyle.Render("[USB]")
	}

	// IOMMU info
	iommuStr := ""
	if dev.IOMMUGroup >= 0 {
		iommuStr = pciMutedStyle.Render(fmt.Sprintf(" (IOMMU:%d)", dev.IOMMUGroup))
	}

	// Device name
	nameStr := pciLabelStyle.Render(dev.Name)
	addrStr := pciMutedStyle.Render(fmt.Sprintf(" %s", dev.Address))
	vendorDevStr := pciMutedStyle.Render(fmt.Sprintf(" [%s:%s]", dev.Vendor, dev.Device))

	return prefix + togglePart + " " + tag + " " + nameStr + addrStr + vendorDevStr + iommuStr
}

// renderROMPath renders a ROM path text input for a device
func (m *PCIPassthroughFormModel) renderROMPath(addr, value string, cursor int, focused bool) string {
	prefix := "    "
	if focused {
		prefix = pciFocusStyle.Render("    > ")
	}

	label := pciMutedStyle.Render("ROM: ")

	var valPart string
	if focused {
		if cursor < len(value) {
			before := value[:cursor]
			at := string(value[cursor])
			after := ""
			if cursor+1 < len(value) {
				after = value[cursor+1:]
			}
			valPart = pciInputStyle.Render(before) +
				lipgloss.NewStyle().Reverse(true).Render(at) +
				pciInputStyle.Render(after)
		} else {
			valPart = pciInputStyle.Render(value) + pciFocusStyle.Render("_")
		}
	} else {
		if value == "" {
			valPart = pciMutedStyle.Render("(optional ROM path)")
		} else {
			valPart = pciInputStyle.Render(value)
		}
	}

	return prefix + label + valPart
}
