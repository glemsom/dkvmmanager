// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VMCreatedMsg is a message indicating a VM was successfully created
type VMCreatedMsg struct {
	VMName string
}

// VMCreateModel wraps the VM form in the ScrollableForm framework.
type VMCreateModel struct {
	form *form.ScrollableForm
}

// NewVMCreateModel creates a new VM creation model.
func NewVMCreateModel(vmManager *vm.Manager) *VMCreateModel {
	fm := NewVMFormModel(vmManager)
	return &VMCreateModel{
		form: form.NewScrollableForm(fm),
	}
}

// Init initializes the model.
func (m *VMCreateModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages.
func (m *VMCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View returns the view for the model.
func (m *VMCreateModel) View() string {
	return m.form.View()
}

// FileBrowserActive returns true if the form's file browser is active.
func (m *VMCreateModel) FileBrowserActive() bool {
	if fm, ok := m.form.Model().(*VMFormModel); ok {
		return fm.FileBrowserActive()
	}
	return false
}