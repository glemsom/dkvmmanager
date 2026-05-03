// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// PCIPassthroughUpdatedMsg is a message indicating PCI passthrough config was saved
type PCIPassthroughUpdatedMsg struct{}

// IsFormSaved implements form.FormSavedMsg.
func (PCIPassthroughUpdatedMsg) IsFormSaved() {}

// FormName implements form.FormSavedMsg.
func (PCIPassthroughUpdatedMsg) FormName() string { return "PCI Passthrough" }

// FormStatus implements form.FormSavedMsg.
func (PCIPassthroughUpdatedMsg) FormStatus() string { return "" }

// PCIVFIOKernelAppliedMsg is sent when vfio-pci.ids has been applied to grub.cfg
type PCIVFIOKernelAppliedMsg struct {
	Success bool
	Error   string
}

// PCIPassthroughModel is a thin wrapper around PCIPassthroughFormModel using the ScrollableForm framework.
type PCIPassthroughModel struct {
	form *form.ScrollableForm
}

// NewPCIPassthroughModel creates a new PCI passthrough model
func NewPCIPassthroughModel(vmManager *vm.Manager) (*PCIPassthroughModel, error) {
	fm, err := NewPCIPassthroughFormModel(vmManager)
	if err != nil {
		return nil, err
	}
	return &PCIPassthroughModel{
		form: form.NewScrollableForm(fm),
	}, nil
}

// Init initializes the model
func (m *PCIPassthroughModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *PCIPassthroughModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View returns the view for the model
func (m *PCIPassthroughModel) View() string {
	return m.form.View()
}

// Form returns the underlying form model (for testing/internal access).
func (m *PCIPassthroughModel) Form() *PCIPassthroughFormModel {
	return m.form.Model().(*PCIPassthroughFormModel)
}
