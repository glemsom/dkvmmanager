// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// ViewChangeMsg is a message indicating a view change
type ViewChangeMsg struct {
	View string
}

// VMCreatedMsg is a message indicating a VM was successfully created
type VMCreatedMsg struct {
	VMName string
}

// VMCreateModel is a thin wrapper around VMFormModel for creating new VMs
type VMCreateModel struct {
	form *VMFormModel
}

// NewVMCreateModel creates a new VM creation model
func NewVMCreateModel(vmManager *vm.Manager) *VMCreateModel {
	return &VMCreateModel{
		form: NewVMFormModel(vmManager),
	}
}

// Init initializes the model
func (m *VMCreateModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *VMCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	m.form = inner.(*VMFormModel)
	return m, cmd
}

// View returns the view for the model
func (m *VMCreateModel) View() string {
	return m.form.View()
}

// FileBrowserActive returns true if the form's file browser is active
func (m *VMCreateModel) FileBrowserActive() bool {
	return m.form.FileBrowserActive()
}
