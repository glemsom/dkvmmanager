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

// usbHeaderStyle is the form title style
var usbHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.Colors.Primary)

// --- FormModel interface methods ---

// RenderHeader returns the form header.
func (m *USBPassthroughFormModel) RenderHeader() string {
	var sb strings.Builder
	sb.WriteString(usbHeaderStyle.Render("USB Passthrough Configuration"))

	if m.scanErr != nil {
		sb.WriteString("\n")
		sb.WriteString(usbErrorStyle.Render(fmt.Sprintf("Warning: USB scan failed: %s", m.scanErr)))
		sb.WriteString("\n")
		sb.WriteString(usbWarnStyle.Render("Devices cannot be configured without a scan."))
	}

	if len(m.devices) == 0 {
		sb.WriteString("\n")
		sb.WriteString(usbMutedStyle.Render("No USB devices found on this system."))
	}

	return sb.String()
}

// RenderFooter returns the form footer.
func (m *USBPassthroughFormModel) RenderFooter() string {
	var parts []string

	// Save error at the bottom
	if errMsg, ok := m.errors["save"]; ok {
		parts = append(parts, "")
		parts = append(parts, usbErrorStyle.Render("Error: "+errMsg))
	}

	// Validation warnings
	for _, w := range m.warnings {
		parts = append(parts, "")
		parts = append(parts, usbWarnStyle.Render("Warning: "+w))
	}

	// Footer help text
	parts = append(parts, "")
	parts = append(parts, usbMutedStyle.Render("Tab Navigate  PgUp/PgDown Scroll  Space/Enter Toggle  ESC Cancel"))

	return strings.Join(parts, "\n")
}

// RenderPosition renders a single position for the form framework.
func (m *USBPassthroughFormModel) RenderPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
	switch pos.Kind {
	case form.FocusToggle:
		dev := m.getDeviceByID(pos.Key)
		if dev == nil {
			return "  ???"
		}
		selected := m.selected[pos.Key]
		return m.renderDeviceToggle(dev, selected, focused)

	case form.FocusButton:
		if pos.Key == "save" {
			saveText := usbMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
			if focused {
				saveText = usbSaveStyle.Render("[Space/Enter] Save") + "    " + usbMutedStyle.Render("[ESC] Cancel")
			}
			return "\n" + saveText
		}
	}

	return ""
}

// renderDeviceToggle renders a USB device as a toggle line.
func (m *USBPassthroughFormModel) renderDeviceToggle(dev *models.USBDevice, selected, focused bool) string {
	prefix := "  "
	if focused {
		prefix = usbFocusStyle.Render("> ")
	}

	// Toggle indicator
	var togglePart string
	if selected {
		if focused {
			togglePart = usbFocusStyle.Render("[X]")
		} else {
			togglePart = usbInputStyle.Render("[X]")
		}
	} else {
		if focused {
			togglePart = usbFocusStyle.Render("[ ]")
		} else {
			togglePart = usbMutedStyle.Render("[ ]")
		}
	}

	// Device name and IDs
	nameStr := usbLabelStyle.Render(dev.Name)
	idStr := usbMutedStyle.Render(fmt.Sprintf(" [%s:%s]", dev.Vendor, dev.Product))
	busStr := ""
	if dev.ID != "" {
		busStr = usbMutedStyle.Render(fmt.Sprintf(" (Bus %s)", dev.ID))
	}

	return prefix + togglePart + " " + nameStr + idStr + busStr
}
