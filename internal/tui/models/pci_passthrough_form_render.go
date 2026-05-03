// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
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

// pciApplyStyle is the apply-to-kernel button style
var pciApplyStyle = styles.FormSaveStyle().Background(styles.Colors.Warning)

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

// --- FormModel interface methods ---

// RenderHeader returns the form header.
func (m *PCIPassthroughFormModel) RenderHeader() string {
	var sb strings.Builder
	sb.WriteString(pciHeaderStyle.Render("PCI Passthrough Configuration"))

	if m.scanErr != nil {
		sb.WriteString("\n")
		sb.WriteString(pciErrorStyle.Render(fmt.Sprintf("Warning: PCI scan failed: %s", m.scanErr)))
		sb.WriteString("\n")
		sb.WriteString(pciWarnStyle.Render("Devices cannot be configured without a scan."))
	}

	if len(m.devices) == 0 {
		sb.WriteString("\n")
		sb.WriteString(pciMutedStyle.Render("No PCI devices found on this system."))
	}

	return sb.String()
}

// RenderFooter returns the form footer.
func (m *PCIPassthroughFormModel) RenderFooter() string {
	var parts []string

	// Save error at the bottom
	if errMsg, ok := m.errors["save"]; ok {
		parts = append(parts, "")
		parts = append(parts, pciErrorStyle.Render("Error: "+errMsg))
	}

	// Kernel apply status message
	if m.kernelMsg != "" {
		parts = append(parts, "")
		if m.kernelMsgErr {
			parts = append(parts, pciErrorStyle.Render("Error: "+m.kernelMsg))
		} else {
			parts = append(parts, pciSaveStyle.Render(m.kernelMsg))
		}
	}

	// Validation warnings
	for _, w := range m.warnings {
		parts = append(parts, "")
		parts = append(parts, pciWarnStyle.Render("Warning: "+w))
	}

	// Footer help text
	parts = append(parts, "")
	parts = append(parts, pciMutedStyle.Render("Tab Navigate  Space/Enter Toggle/Action  ESC Cancel"))

	return strings.Join(parts, "\n")
}

// RenderPosition renders a single position for the form framework.
func (m *PCIPassthroughFormModel) RenderPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
	switch pos.Kind {
	case form.FocusHeader:
		return m.renderGroupHeader(pos)

	case form.FocusToggle:
		fd := pos.Data.(pciFocusData)
		dev := m.getDeviceByAddr(fd.Address)
		if dev == nil {
			return "  ???"
		}
		selected := m.selected[fd.Address]
		return m.renderDeviceToggle(dev, selected, focused)

	case form.FocusButton:
		if pos.Key == "save" {
			saveText := pciMutedStyle.Render("[Space/Enter] Save") + "    " + pciMutedStyle.Render("[ESC] Cancel")
			if focused {
				saveText = pciSaveStyle.Render("[Space/Enter] Save") + "    " + pciMutedStyle.Render("[ESC] Cancel")
			}
			return "\n" + saveText
		}

		if pos.Key == "apply_kernel" {
			applyText := pciMutedStyle.Render("[Space/Enter] Apply to Kernel") + "    " + pciMutedStyle.Render("[ESC] Cancel")
			if focused {
				applyText = pciApplyStyle.Render("[Space/Enter] Apply to Kernel") + "    " + pciMutedStyle.Render("[ESC] Cancel")
			}
			return "\n" + applyText
		}
	}

	return ""
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
func (m *PCIPassthroughFormModel) renderGroupHeader(pos form.FocusPos) string {
	fd := pos.Data.(pciFocusData)
	groupNum := fd.GroupNum

	// Find devices in this group from the IOMMU group index
	devices, ok := m.iommuGroups[groupNum]
	if !ok {
		return ""
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
