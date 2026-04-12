// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VMUpdatedMsg is a message indicating a VM was successfully updated
type VMUpdatedMsg struct {
	VMName string
}

// VMEditModel is a thin wrapper around VMFormModel for editing existing VMs
type VMEditModel struct {
	form *VMFormModel
}

// NewVMEditModel creates a new VM edit model
func NewVMEditModel(vmManager *vm.Manager, vmID string) (*VMEditModel, error) {
	// Load the VM
	vms, err := vmManager.ListVMs()
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	var targetVM *models.VM
	for i := range vms {
		if vms[i].ID == vmID {
			targetVM = &vms[i]
			break
		}
	}

	if targetVM == nil {
		return nil, fmt.Errorf("VM not found: %s", vmID)
	}

	return &VMEditModel{
		form: NewVMFormModelEdit(vmManager, targetVM),
	}, nil
}

// Init initializes the model
func (m *VMEditModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *VMEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if f, ok := inner.(*VMFormModel); ok {
		m.form = f
	}
	return m, cmd
}

// View returns the view for the model
func (m *VMEditModel) View() string {
	return m.form.View()
}

// FileBrowserActive returns true if the form's file browser is active
func (m *VMEditModel) FileBrowserActive() bool {
	return m.form.FileBrowserActive()
}
