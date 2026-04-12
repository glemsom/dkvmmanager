// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// PCIPassthroughUpdatedMsg is a message indicating PCI passthrough config was saved
type PCIPassthroughUpdatedMsg struct{}

// PCIPassthroughModel is a thin wrapper around PCIPassthroughFormModel
type PCIPassthroughModel struct {
	form *PCIPassthroughFormModel
}

// NewPCIPassthroughModel creates a new PCI passthrough model
func NewPCIPassthroughModel(vmManager *vm.Manager) (*PCIPassthroughModel, error) {
	form, err := NewPCIPassthroughFormModel(vmManager)
	if err != nil {
		return nil, err
	}
	return &PCIPassthroughModel{form: form}, nil
}

// Init initializes the model
func (m *PCIPassthroughModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *PCIPassthroughModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	m.form = inner.(*PCIPassthroughFormModel)
	return m, cmd
}

// View returns the view for the model
func (m *PCIPassthroughModel) View() string {
	return m.form.View()
}
