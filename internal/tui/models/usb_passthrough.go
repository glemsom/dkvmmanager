// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// USBPassthroughUpdatedMsg is a message indicating USB passthrough config was saved
type USBPassthroughUpdatedMsg struct{}

// USBPassthroughModel is a thin wrapper around USBPassthroughFormModel
type USBPassthroughModel struct {
	form *USBPassthroughFormModel
}

// NewUSBPassthroughModel creates a new USB passthrough model
func NewUSBPassthroughModel(vmManager *vm.Manager) (*USBPassthroughModel, error) {
	form, err := NewUSBPassthroughFormModel(vmManager)
	if err != nil {
		return nil, err
	}
	return &USBPassthroughModel{form: form}, nil
}

// Init initializes the model
func (m *USBPassthroughModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *USBPassthroughModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	m.form = inner.(*USBPassthroughFormModel)
	return m, cmd
}

// View returns the view for the model
func (m *USBPassthroughModel) View() string {
	return m.form.View()
}
