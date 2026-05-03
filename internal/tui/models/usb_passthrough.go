// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// USBPassthroughUpdatedMsg is a message indicating USB passthrough config was saved
type USBPassthroughUpdatedMsg struct{}

// IsFormSaved implements form.FormSavedMsg.
func (USBPassthroughUpdatedMsg) IsFormSaved() {}

// FormName implements form.FormSavedMsg.
func (USBPassthroughUpdatedMsg) FormName() string { return "USB Passthrough" }

// FormStatus implements form.FormSavedMsg.
func (USBPassthroughUpdatedMsg) FormStatus() string { return "" }

// USBPassthroughModel is a thin wrapper around USBPassthroughFormModel using the ScrollableForm framework.
type USBPassthroughModel struct {
	form *form.ScrollableForm
}

// NewUSBPassthroughModel creates a new USB passthrough model
func NewUSBPassthroughModel(vmManager *vm.Manager) (*USBPassthroughModel, error) {
	fm, err := NewUSBPassthroughFormModel(vmManager)
	if err != nil {
		return nil, err
	}
	return &USBPassthroughModel{
		form: form.NewScrollableForm(fm),
	}, nil
}

// Init initializes the model
func (m *USBPassthroughModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *USBPassthroughModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View returns the view for the model
func (m *USBPassthroughModel) View() string {
	return m.form.View()
}

// Form returns the underlying form model (for testing/internal access).
func (m *USBPassthroughModel) Form() *USBPassthroughFormModel {
	return m.form.Model().(*USBPassthroughFormModel)
}
