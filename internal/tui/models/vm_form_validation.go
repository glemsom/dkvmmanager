// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// validateAndSaveCmd checks all fields and returns a cmd that sends the appropriate message
func (m *VMFormModel) validateAndSaveCmd() tea.Cmd {
	m.errors = make(map[string]string)

	// Name
	m.vmName = strings.TrimSpace(m.vmName)
	if m.vmName == "" {
		m.errors["vmName"] = "VM name cannot be empty"
	} else if !nameRegex.MatchString(m.vmName) {
		m.errors["vmName"] = "Only alphanumeric, dash, underscore, and space allowed"
	}

	// MAC (optional - allow empty means auto-generated)
	if strings.TrimSpace(m.macAddress) != "" && !macRegex.MatchString(m.macAddress) {
		m.errors["macAddress"] = "Expected format: xx:xx:xx:xx:xx:xx"
	}

	// TPM binary check if TPM is enabled
	if m.tpmEnabled {
		cfg := m.vmManager.GetConfig()
		if _, err := os.Stat(cfg.TPMBinary); err != nil {
			m.errors["tpmEnabled"] = fmt.Sprintf("TPM enabled but swtpm binary not found at %s", cfg.TPMBinary)
		}
	}

	if len(m.errors) > 0 {
		// Focus the first field with an error
		for i, p := range m.positions {
			if _, hasErr := m.errors[p.Key]; hasErr {
				m.focusIndex = i
				break
			}
		}
		return nil
	}

	// Strip empty trailing disk slot
	disks := make([]string, 0, len(m.hardDisks))
	for _, d := range m.hardDisks {
		if strings.TrimSpace(d) != "" {
			disks = append(disks, d)
		}
	}

	cdroms := make([]string, 0, len(m.cdroms))
	for _, c := range m.cdroms {
		if strings.TrimSpace(c) != "" {
			cdroms = append(cdroms, c)
		}
	}

	if m.mode == FormCreate {
		return m.saveNewVMCmd(disks, cdroms)
	}
	return m.updateExistingVMCmd(disks, cdroms)
}

// saveNewVMCmd creates a new VM and persists it, returning a cmd
func (m *VMFormModel) saveNewVMCmd(disks, cdroms []string) tea.Cmd {
	createdVM, err := m.vmManager.CreateVM(m.vmName)
	if err != nil {
		m.errors["vmName"] = fmt.Sprintf("Failed to create VM: %v", err)
		return nil
	}

	createdVM.HardDisks = disks
	createdVM.CDROMs = cdroms
	createdVM.MAC = m.macAddress
	createdVM.NetworkMode = m.networkMode
	createdVM.TPMEnabled = m.tpmEnabled
	if m.vncEnabled {
		createdVM.VNCListen = "0.0.0.0:0"
	} else {
		createdVM.VNCListen = ""
	}

	if err := m.vmManager.SaveVM(createdVM); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		return nil
	}

	return func() tea.Msg { return VMCreatedMsg{VMName: m.vmName} }
}

// updateExistingVMCmd updates an existing VM and persists it, returning a cmd
func (m *VMFormModel) updateExistingVMCmd(disks, cdroms []string) tea.Cmd {
	m.vm.Name = m.vmName
	m.vm.HardDisks = disks
	m.vm.CDROMs = cdroms
	m.vm.MAC = m.macAddress
	m.vm.NetworkMode = m.networkMode
	m.vm.TPMEnabled = m.tpmEnabled
	if m.vncEnabled {
		m.vm.VNCListen = "0.0.0.0:0"
	} else {
		m.vm.VNCListen = ""
	}

	if err := m.vmManager.SaveVM(m.vm); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		return nil
	}

	return func() tea.Msg { return VMUpdatedMsg{VMName: m.vmName} }
}