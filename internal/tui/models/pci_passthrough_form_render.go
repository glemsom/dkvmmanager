// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// pciLabelStyle is the label style for form fields
var pciLabelStyle = lipgloss.NewStyle().Foreground(styles.Colors.ForegroundDim)

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

// pciHeaderStyle is the IOMMU group header style
var pciHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.Colors.Primary)

// pciAddrStyle is the PCI address style (bold/high-contrast for quick scanning)
var pciAddrStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.Colors.Primary)

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
	case pciGroupHeader:
		lines = append(lines, m.renderGroupHeader(pos))
		return lines

	case pciToggle:
		dev := m.getDeviceByAddr(pos.deviceAddr)
		if dev == nil {
			lines = append(lines, "  ???")
			return lines
		}
		selected := m.selected[pos.deviceAddr]
		lines = append(lines, m.renderDeviceToggle(dev, selected, focused))
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

// renderDeviceToggle renders a PCI device as a toggle line.
// Format: [X] 0000:01:00.0 [GPU] NVIDIA GeForce GTX 1080 [10de:1b80] (IOMMU:1)
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

	// PCI address first (bold, high-contrast for quick scanning)
	addrStr := pciAddrStyle.Render(dev.Address)

	// Device type tag
	var tag string
	if dev.IsGPU {
		tag = pciGPUStyle.Render("[GPU]")
	} else if dev.IsUSB {
		tag = pciUSBStyle.Render("[USB]")
	}

	// Device name
	nameStr := pciLabelStyle.Render(dev.Name)
	vendorDevStr := pciMutedStyle.Render(fmt.Sprintf(" [%s:%s]", dev.Vendor, dev.Device))

	// IOMMU info
	iommuStr := ""
	if dev.IOMMUGroup >= 0 {
		iommuStr = pciMutedStyle.Render(fmt.Sprintf(" (IOMMU:%d)", dev.IOMMUGroup))
	}

	return prefix + togglePart + " " + addrStr + " " + tag + " " + nameStr + vendorDevStr + iommuStr
}

// renderGroupHeader renders an IOMMU group header line.
// Format: ── IOMMU Group 1 (2 devices, all selected) ──
func (m *PCIPassthroughFormModel) renderGroupHeader(pos pciFocusPos) string {
	groupNum := pos.groupNum

	// Find all devices in this group (consecutive toggles after this header)
	headerIdx := -1
	for i, p := range m.positions {
		if p.kind == pciGroupHeader && p.groupNum == groupNum {
			headerIdx = i
			break
		}
	}
	if headerIdx < 0 {
		return ""
	}

	// Walk forward collecting device pointers from this group
	var devices []*models.PCIDevice
	for i := headerIdx + 1; i < len(m.positions); i++ {
		if m.positions[i].kind == pciToggle {
			d := m.getDeviceByAddr(m.positions[i].deviceAddr)
			if d != nil {
				devices = append(devices, d)
			}
		} else {
			break // Hit next header or save button
		}
	}

	// Count selected devices in this group
	selectedCount := 0
	for _, d := range devices {
		if m.selected[d.Address] {
			selectedCount++
		}
	}

	// Build label
	var label string
	if groupNum < 0 {
		label = "Ungrouped Devices"
	} else {
		label = fmt.Sprintf("IOMMU Group %d", groupNum)
	}

	// Selection status suffix
	status := fmt.Sprintf("(%d devices)", len(devices))
	if selectedCount > 0 {
		if selectedCount == len(devices) {
			status = fmt.Sprintf("(%d devices, all selected)", len(devices))
		} else {
			status = fmt.Sprintf("(%d devices, %d selected)", len(devices), selectedCount)
		}
	}

	return pciHeaderStyle.Render("── " + label + " " + status + " ──")
}
