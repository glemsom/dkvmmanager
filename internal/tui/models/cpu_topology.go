// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// CPUTopologyUpdatedMsg is a message indicating CPU topology config was saved
type CPUTopologyUpdatedMsg struct{}

// CPUTopologyModel is a thin wrapper around CPUTopologyFormModel
type CPUTopologyModel struct {
	form *CPUTopologyFormModel
}

// NewCPUTopologyModel creates a new CPU topology model
func NewCPUTopologyModel(vmManager *vm.Manager) (*CPUTopologyModel, error) {
	form, err := NewCPUTopologyFormModel(vmManager)
	if err != nil {
		return nil, err
	}
	return &CPUTopologyModel{form: form}, nil
}

// Init initializes the model
func (m *CPUTopologyModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *CPUTopologyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	m.form = inner.(*CPUTopologyFormModel)
	return m, cmd
}

// View returns the view for the model
func (m *CPUTopologyModel) View() string {
	return m.form.View()
}
