// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/glemsom/dkvmmanager/internal/domain"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// validateAndSaveCmd persists the USB passthrough config and returns a form result + tea.Cmd.
func (m *USBPassthroughFormModel) validateAndSaveCmd() (form.FormResult, tea.Cmd) {
	m.errors = make(map[string]string)
	m.warnings = nil

	// Build config from selected devices
	var devices []domain.USBPassthroughDevice
	for _, dev := range m.devices {
		key := usbDeviceKey(dev.Vendor, dev.Product)
		if !m.selected[key] {
			continue
		}
		devices = append(devices, domain.USBPassthroughDevice{
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

	cfg := domain.USBPassthroughConfig{
		Devices: devices,
	}

	if err := m.repo.SaveConfig("usb_passthrough", cfg); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		return form.ResultNone, nil
	}

	// Store warnings to display after successful save
	m.warnings = warnings

	return form.ResultSave, func() tea.Msg {
		return USBPassthroughUpdatedMsg{}
	}
}
