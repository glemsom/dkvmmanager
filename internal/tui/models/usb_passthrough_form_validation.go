// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// validateAndSaveCmd persists the USB passthrough config and returns a form result + tea.Cmd.
func (m *USBPassthroughFormModel) validateAndSaveCmd() (form.FormResult, tea.Cmd) {
	m.errors = make(map[string]string)
	m.warnings = nil

	// Build config from selected devices
	var devices []models.USBPassthroughDevice
	for _, dev := range m.devices {
		key := usbDeviceKey(dev.Vendor, dev.Product)
		if !m.selected[key] {
			continue
		}
		devices = append(devices, models.USBPassthroughDevice{
			Vendor:  dev.Vendor,
			Product: dev.Product,
			Name:    dev.Name,
			BusID:   dev.ID,
		})
	}

	// Validate before saving
	warnings, valErrors := vm.ValidateUSBDevices(devices)
	if len(valErrors) > 0 {
		m.errors["save"] = strings.Join(valErrors, "; ")
		return form.ResultNone, nil
	}

	cfg := models.USBPassthroughConfig{
		Devices: devices,
	}

	if err := m.vmManager.SaveUSBPassthroughConfig(cfg); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		return form.ResultNone, nil
	}

	// Store warnings to display after successful save
	m.warnings = warnings

	return form.ResultSave, func() tea.Msg {
		return USBPassthroughUpdatedMsg{}
	}
}
