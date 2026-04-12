// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// validateAndSave checks all fields and either saves or shows errors
func (m *VMFormModel) validateAndSave() (tea.Model, tea.Cmd) {
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

	if len(m.errors) > 0 {
		// Focus the first field with an error
		for i, p := range m.positions {
			k := posKey(p)
			if _, hasErr := m.errors[k]; hasErr {
				m.focusIndex = i
				break
			}
		}
		m.syncViewport()
		return m, nil
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
		return m.saveNewVM(disks, cdroms)
	}
	return m.updateExistingVM(disks, cdroms)
}

// saveNewVM creates a new VM and persists it
func (m *VMFormModel) saveNewVM(disks, cdroms []string) (tea.Model, tea.Cmd) {
	createdVM, err := m.vmManager.CreateVM(m.vmName)
	if err != nil {
		m.errors["vmName"] = fmt.Sprintf("Failed to create VM: %v", err)
		m.syncViewport()
		return m, nil
	}

	createdVM.HardDisks = disks
	createdVM.CDROMs = cdroms
	createdVM.MAC = m.macAddress
	createdVM.NetworkMode = m.networkMode
	if m.vncEnabled {
		createdVM.VNCListen = "0.0.0.0:0"
	} else {
		createdVM.VNCListen = ""
	}

	if err := m.vmManager.SaveVM(createdVM); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		m.syncViewport()
		return m, nil
	}

	return m, func() tea.Msg { return VMCreatedMsg{VMName: m.vmName} }
}

// updateExistingVM updates an existing VM and persists it
func (m *VMFormModel) updateExistingVM(disks, cdroms []string) (tea.Model, tea.Cmd) {
	m.vm.Name = m.vmName
	m.vm.HardDisks = disks
	m.vm.CDROMs = cdroms
	m.vm.MAC = m.macAddress
	m.vm.NetworkMode = m.networkMode
	if m.vncEnabled {
		m.vm.VNCListen = "0.0.0.0:0"
	} else {
		m.vm.VNCListen = ""
	}

	if err := m.vmManager.SaveVM(m.vm); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		m.syncViewport()
		return m, nil
	}

	return m, func() tea.Msg { return VMUpdatedMsg{VMName: m.vmName} }
}
