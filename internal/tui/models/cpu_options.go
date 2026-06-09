// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "charm.land/bubbletea/v2"
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
func NewCPUOptionsModel(repo *vm.Repository) *CPUOptionsModel {
	fm := NewCPUOptionsFormModel(repo)
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
func (m *CPUOptionsModel) View() tea.View {
	return m.form.View()
}

// SetSize forwards window resize to the underlying form.
func (m *CPUOptionsModel) SetSize(width, height int) { m.form.SetSize(width, height) }

// FileBrowserActive returns false (CPU options has no file browser).
func (m *CPUOptionsModel) FileBrowserActive() bool { return false }

// Form returns the underlying form model (for testing/internal access).
func (m *CPUOptionsModel) Form() *CPUOptionsFormModel {
	return m.form.Model().(*CPUOptionsFormModel)
}
