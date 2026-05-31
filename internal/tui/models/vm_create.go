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
	if fm, ok := m.form.Model().(*VMFormModel); ok {
		if fm.FileBrowserActive() {
			// Handle DiskAddedMsg - forward it to the form for processing
			if dam, ok := msg.(DiskAddedMsg); ok {
				_ = fm.HandleMessage(dam)
				return m, nil
			}
			// Handle FileSelectedMsg - forward it to the form for processing
			if fsm, ok := msg.(FileSelectedMsg); ok {
				_ = fm.HandleMessage(fsm)
				return m, nil
			}
			if fm.addDiskModel != nil {
				inner, cmd := fm.addDiskModel.Update(msg)
				if ad, ok := inner.(*AddDiskModel); ok {
					fm.addDiskModel = ad
				}
				if cmd != nil {
					nextMsg := cmd()
					// Feed the result back through AddDiskModel so it can convert
					// FileSelectedMsg into DiskAddedMsg before reaching the form.
					inner2, cmd2 := fm.addDiskModel.Update(nextMsg)
					if ad, ok := inner2.(*AddDiskModel); ok {
						fm.addDiskModel = ad
					}
					if cmd2 != nil {
						return m.forwardToForm(cmd2())
					}
					return m, nil
				}
				return m, nil
			}
			if fm.fileBrowser != nil {
				inner, cmd := fm.fileBrowser.Update(msg)
				if fb, ok := inner.(*FileBrowserModel); ok {
					fm.fileBrowser = fb
				}
				if cmd != nil {
					return m.forwardToForm(cmd())
				}
				return m, nil
			}
		}
	}

	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// forwardToForm passes a message into the wrapped ScrollableForm so the
// VMFormModel can handle it through HandleMessage.
func (m *VMCreateModel) forwardToForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View returns the view for the model.
func (m *VMCreateModel) View() string {
	if fm, ok := m.form.Model().(*VMFormModel); ok {
		if fm.addDiskModel != nil && fm.addDiskModel.active {
			return fm.addDiskModel.View()
		}
		if fm.fileBrowser != nil && fm.fileBrowser.active {
			return fm.fileBrowser.View()
		}
	}
	return m.form.View()
}

// SetSize forwards window resize to the underlying form.
func (m *VMCreateModel) SetSize(width, height int) { m.form.SetSize(width, height) }

// FileBrowserActive returns true if the form's file browser is active.
func (m *VMCreateModel) FileBrowserActive() bool {
	if fm, ok := m.form.Model().(*VMFormModel); ok {
		return fm.FileBrowserActive()
	}
	return false
}