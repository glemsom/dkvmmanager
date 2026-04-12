// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// validateAndSave persists the PCI passthrough config
func (m *PCIPassthroughFormModel) validateAndSave() (tea.Model, tea.Cmd) {
	m.errors = make(map[string]string)
	m.warnings = nil

	// Build config from selected devices
	var devices []models.PCIPassthroughDevice
	for _, dev := range m.devices {
		if !m.selected[dev.Address] {
			continue
		}
		romPath := m.romPaths[dev.Address]
		devices = append(devices, models.PCIPassthroughDevice{
			Address:   dev.Address,
			ROMPath:   romPath,
			Vendor:    dev.Vendor,
			Device:    dev.Device,
			Name:      dev.Name,
			ClassCode: dev.ClassCode,
		})
	}

	// Validate before saving
	warnings, valErrors := vm.ValidatePCIDevices(devices)
	if len(valErrors) > 0 {
		m.errors["save"] = strings.Join(valErrors, "; ")
		m.syncViewport()
		return m, nil
	}

	cfg := models.PCIPassthroughConfig{
		Devices: devices,
	}

	if err := m.vmManager.SavePCIPassthroughConfig(cfg); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		m.syncViewport()
		return m, nil
	}

	// Store warnings to display after successful save
	m.warnings = warnings

	return m, func() tea.Msg {
		return PCIPassthroughUpdatedMsg{}
	}
}
