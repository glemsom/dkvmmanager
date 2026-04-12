// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import tea "github.com/charmbracelet/bubbletea"

// SSHPasswordUpdatedMsg is a message indicating the SSH password was successfully changed
type SSHPasswordUpdatedMsg struct{}

// SSHPasswordModel is a thin wrapper around SSHPasswordFormModel
type SSHPasswordModel struct {
	form *SSHPasswordFormModel
}

// NewSSHPasswordModel creates a new SSH password model
func NewSSHPasswordModel() *SSHPasswordModel {
	return &SSHPasswordModel{
		form: NewSSHPasswordFormModel(),
	}
}

// Init initializes the model
func (m *SSHPasswordModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *SSHPasswordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if form, ok := inner.(*SSHPasswordFormModel); ok {
		m.form = form
	}
	return m, cmd
}

// View returns the view for the model
func (m *SSHPasswordModel) View() string {
	return m.form.View()
}
