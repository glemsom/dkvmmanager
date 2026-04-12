// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// CPUOptionsUpdatedMsg is a message indicating CPU options were successfully saved
type CPUOptionsUpdatedMsg struct{}

// CPUOptionsModel is a thin wrapper around CPUOptionsFormModel
type CPUOptionsModel struct {
	form *CPUOptionsFormModel
}

// NewCPUOptionsModel creates a new CPU options model
func NewCPUOptionsModel(vmManager *vm.Manager) *CPUOptionsModel {
	return &CPUOptionsModel{
		form: NewCPUOptionsFormModel(vmManager),
	}
}

// Init initializes the model
func (m *CPUOptionsModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *CPUOptionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	m.form = inner.(*CPUOptionsFormModel)
	return m, cmd
}

// View returns the view for the model
func (m *CPUOptionsModel) View() string {
	return m.form.View()
}
