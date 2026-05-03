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

// validateAndSaveCmd persists the PCI passthrough config and returns a form result + tea.Cmd.
func (m *PCIPassthroughFormModel) validateAndSaveCmd() (form.FormResult, tea.Cmd) {
	m.errors = make(map[string]string)
	m.warnings = nil

	// Build config from selected devices
	var devices []models.PCIPassthroughDevice
	for _, dev := range m.devices {
		if !m.selected[dev.Address] {
			continue
		}
		devices = append(devices, models.PCIPassthroughDevice{
			Address:   dev.Address,
			ROMPath:   "", // ROM field removed from UI; preserved for backward compatibility
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
		return form.ResultNone, nil
	}

	cfg := models.PCIPassthroughConfig{
		Devices: devices,
	}

	if err := m.vmManager.SavePCIPassthroughConfig(cfg); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		return form.ResultNone, nil
	}

	// Store warnings to display after successful save
	m.warnings = warnings

	return form.ResultSave, func() tea.Msg {
		return PCIPassthroughUpdatedMsg{}
	}
}

// handleApplyKernelCmd applies the current PCI passthrough config to grub.cfg.
func (m *PCIPassthroughFormModel) handleApplyKernelCmd() tea.Cmd {
	m.errors = make(map[string]string)
	m.kernelMsg = ""
	m.kernelMsgErr = false

	// Build config from selected devices (same as save)
	var devices []models.PCIPassthroughDevice
	for _, dev := range m.devices {
		if !m.selected[dev.Address] {
			continue
		}
		devices = append(devices, models.PCIPassthroughDevice{
			Address:   dev.Address,
			ROMPath:   "",
			Vendor:    dev.Vendor,
			Device:    dev.Device,
			Name:      dev.Name,
			ClassCode: dev.ClassCode,
		})
	}

	// Validate before applying
	_, valErrors := vm.ValidatePCIDevices(devices)
	if len(valErrors) > 0 {
		m.kernelMsg = strings.Join(valErrors, "; ")
		m.kernelMsgErr = true
		return nil
	}

	// Build the PCI passthrough config and save it first
	cfg := models.PCIPassthroughConfig{
		Devices: devices,
	}
	if err := m.vmManager.SavePCIPassthroughConfig(cfg); err != nil {
		m.kernelMsg = fmt.Sprintf("Failed to save config: %v", err)
		m.kernelMsgErr = true
		return nil
	}

	// Apply to kernel grub.cfg asynchronously
	return func() tea.Msg {
		err := m.vmManager.ApplyVFIOIDsToKernel()
		if err != nil {
			return PCIVFIOKernelAppliedMsg{
				Success: false,
				Error:   err.Error(),
			}
		}
		return PCIVFIOKernelAppliedMsg{
			Success: true,
		}
	}
}
