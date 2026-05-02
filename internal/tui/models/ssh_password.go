// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// SSHPasswordUpdatedMsg is a message indicating the SSH password was successfully changed
type SSHPasswordUpdatedMsg struct{}

// IsFormSaved implements form.FormSavedMsg.
func (SSHPasswordUpdatedMsg) IsFormSaved() {}

// FormName implements form.FormSavedMsg.
func (SSHPasswordUpdatedMsg) FormName() string { return "SSH Password" }

// FormStatus implements form.FormSavedMsg.
func (SSHPasswordUpdatedMsg) FormStatus() string { return "" }

// SSHPasswordModel wraps the SSH password form in the ScrollableForm framework.
type SSHPasswordModel struct {
	form *form.ScrollableForm
}

// NewSSHPasswordModel creates a new SSH password model.
func NewSSHPasswordModel() *SSHPasswordModel {
	fm := NewSSHPasswordFormModel()
	return &SSHPasswordModel{
		form: form.NewScrollableForm(fm),
	}
}

// Init initializes the model.
func (m *SSHPasswordModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages.
func (m *SSHPasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View returns the view for the model.
func (m *SSHPasswordModel) View() string {
	return m.form.View()
}

// Form returns the underlying form model (for testing/internal access).
func (m *SSHPasswordModel) Form() *SSHPasswordFormModel {
	return m.form.Model().(*SSHPasswordFormModel)
}
