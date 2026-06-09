// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "charm.land/bubbletea/v2"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// CPUTopologyUpdatedMsg is a message indicating CPU topology config was saved
type CPUTopologyUpdatedMsg struct{}

// IsFormSaved implements form.FormSavedMsg.
func (CPUTopologyUpdatedMsg) IsFormSaved() {}

// FormName implements form.FormSavedMsg.
func (CPUTopologyUpdatedMsg) FormName() string { return "CPU Topology" }

// FormStatus implements form.FormSavedMsg.
func (CPUTopologyUpdatedMsg) FormStatus() string { return "" }

// CPUTopologyModel is a thin wrapper around CPUTopologyFormModel using the ScrollableForm framework.
type CPUTopologyModel struct {
	form *form.ScrollableForm
}

// NewCPUTopologyModel creates a new CPU topology model
func NewCPUTopologyModel(repo *vm.Repository) (*CPUTopologyModel, error) {
	formModel, err := NewCPUTopologyFormModel(repo)
	if err != nil {
		return nil, err
	}
	return &CPUTopologyModel{form: form.NewScrollableForm(formModel)}, nil
}

// Init initializes the model
func (m *CPUTopologyModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *CPUTopologyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View returns the view for the model
func (m *CPUTopologyModel) View() tea.View {
	return m.form.View()
}

// SetSize forwards window resize to the underlying form.
func (m *CPUTopologyModel) SetSize(width, height int) { m.form.SetSize(width, height) }

// FileBrowserActive returns false (CPU topology has no file browser).
func (m *CPUTopologyModel) FileBrowserActive() bool { return false }

// Form returns the underlying form model (for testing/internal access).
func (m *CPUTopologyModel) Form() *CPUTopologyFormModel {
	return m.form.Model().(*CPUTopologyFormModel)
}
