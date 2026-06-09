// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "charm.land/bubbletea/v2"
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
func NewUSBPassthroughModel(repo *vm.Repository, hostDiscovery vm.HostDiscovery) (*USBPassthroughModel, error) {
	fm, err := NewUSBPassthroughFormModel(repo, hostDiscovery)
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
func (m *USBPassthroughModel) View() tea.View {
	return m.form.View()
}

// SetSize forwards window resize to the underlying form.
func (m *USBPassthroughModel) SetSize(width, height int) { m.form.SetSize(width, height) }

// FileBrowserActive returns false (USB passthrough has no file browser).
func (m *USBPassthroughModel) FileBrowserActive() bool { return false }

// Form returns the underlying form model (for testing/internal access).
func (m *USBPassthroughModel) Form() *USBPassthroughFormModel {
	return m.form.Model().(*USBPassthroughFormModel)
}
