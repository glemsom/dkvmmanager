// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// StartStopScriptModel is a thin wrapper around StartStopScriptFormModel using the ScrollableForm framework.
type StartStopScriptModel struct {
	form *form.ScrollableForm
}

// NewStartStopScriptModel creates a new start/stop script model
func NewStartStopScriptModel(vmManager *vm.Manager) (*StartStopScriptModel, error) {
	fm, err := NewStartStopScriptFormModel(vmManager)
	if err != nil {
		return nil, err
	}
	return &StartStopScriptModel{
		form: form.NewScrollableForm(fm),
	}, nil
}

// Init initializes the model
func (m *StartStopScriptModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *StartStopScriptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View returns the view for the model
func (m *StartStopScriptModel) View() string {
	return m.form.View()
}

// Form returns the underlying form model (for testing/internal access).
func (m *StartStopScriptModel) Form() *StartStopScriptFormModel {
	return m.form.Model().(*StartStopScriptFormModel)
}

// FileBrowserActive returns true if the form's file browser is active.
func (m *StartStopScriptModel) FileBrowserActive() bool {
	fm := m.Form()
	return fm.fileBrowser != nil && fm.fileBrowser.active
}
