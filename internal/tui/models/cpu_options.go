// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// CPUOptionsUpdatedMsg is a message indicating CPU options were successfully saved
type CPUOptionsUpdatedMsg struct{}

// IsFormSaved implements form.FormSavedMsg.
func (CPUOptionsUpdatedMsg) IsFormSaved() {}

// FormName implements form.FormSavedMsg.
func (CPUOptionsUpdatedMsg) FormName() string { return "CPU Options" }

// FormStatus implements form.FormSavedMsg.
func (CPUOptionsUpdatedMsg) FormStatus() string { return "" }

// CPUOptionsModel is a thin wrapper around CPUOptionsFormModel using the ScrollableForm framework.
type CPUOptionsModel struct {
	form *form.ScrollableForm
}

// NewCPUOptionsModel creates a new CPU options model
func NewCPUOptionsModel(vmManager *vm.Manager) *CPUOptionsModel {
	fm := NewCPUOptionsFormModel(vmManager)
	return &CPUOptionsModel{
		form: form.NewScrollableForm(fm),
	}
}

// Init initializes the model
func (m *CPUOptionsModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *CPUOptionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View returns the view for the model
func (m *CPUOptionsModel) View() string {
	return m.form.View()
}

// Form returns the underlying form model (for testing/internal access).
func (m *CPUOptionsModel) Form() *CPUOptionsFormModel {
	return m.form.Model().(*CPUOptionsFormModel)
}
