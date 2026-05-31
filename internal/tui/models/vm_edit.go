// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VMUpdatedMsg is a message indicating a VM was successfully updated
type VMUpdatedMsg struct {
	VMName string
}

// VMEditModel wraps the VM form in the ScrollableForm framework for editing existing VMs.
type VMEditModel struct {
	form *form.ScrollableForm
}

// NewVMEditModel creates a new VM edit model.
func NewVMEditModel(vmManager *vm.Manager, vmID string) (*VMEditModel, error) {
	// Load the VM
	vms, err := vmManager.ListVMs()
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	var targetVM *models.VM
	for i := range vms {
		if vms[i].ID == vmID {
			targetVM = &vms[i]
			break
		}
	}

	if targetVM == nil {
		return nil, fmt.Errorf("VM not found: %s", vmID)
	}

	return &VMEditModel{
		form: form.NewScrollableForm(NewVMFormModelEdit(vmManager, targetVM)),
	}, nil
}

// Init initializes the model.
func (m *VMEditModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages.
func (m *VMEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if fm, ok := m.form.Model().(*VMFormModel); ok {
		if fm.FileBrowserActive() {
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
// VMFormModel can handle it through HandleMessage. It also returns any
// command produced by the form update.
func (m *VMEditModel) forwardToForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View returns the view for the model.
func (m *VMEditModel) View() string {
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
func (m *VMEditModel) SetSize(width, height int) { m.form.SetSize(width, height) }

// FileBrowserActive returns true if the form's file browser is active.
func (m *VMEditModel) FileBrowserActive() bool {
	if fm, ok := m.form.Model().(*VMFormModel); ok {
		return fm.FileBrowserActive()
	}
	return false
}

// Form returns the underlying VMFormModel (for testing/internal access).
func (m *VMEditModel) Form() *VMFormModel {
	return m.form.Model().(*VMFormModel)
}