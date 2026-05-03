// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VCPUPinningModel is a thin wrapper around VCPUPinningFormModel using the ScrollableForm framework.
type VCPUPinningModel struct {
	form *form.ScrollableForm
}

// NewVCPUPinningModel creates a new vCPU pinning model
func NewVCPUPinningModel(vmManager *vm.Manager) (*VCPUPinningModel, error) {
	fm, err := NewVCPUPinningFormModel(vmManager)
	if err != nil {
		return nil, err
	}
	return &VCPUPinningModel{
		form: form.NewScrollableForm(fm),
	}, nil
}

// Init initializes the model
func (m *VCPUPinningModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *VCPUPinningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View returns the view for the model
func (m *VCPUPinningModel) View() string {
	return m.form.View()
}

// Form returns the underlying form model (for testing/internal access).
func (m *VCPUPinningModel) Form() *VCPUPinningFormModel {
	return m.form.Model().(*VCPUPinningFormModel)
}
