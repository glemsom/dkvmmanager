// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// HandleEnter is called when the user presses Enter on a position.
func (m *PCIPassthroughFormModel) HandleEnter(pos form.FocusPos) (form.FormResult, tea.Cmd) {
	switch pos.Kind {
	case form.FocusToggle:
		d := pos.Data.(pciFocusData)
		m.toggleDevice(d.Address)
		return form.ResultNone, nil

	case form.FocusButton:
		if pos.Key == "save" {
			return m.validateAndSaveCmd()
		}
		if pos.Key == "apply_kernel" {
			return form.ResultNone, m.handleApplyKernelCmd()
		}
		// Unknown button, move focus forward
		m.focusIndex++
		if m.focusIndex >= len(m.positions) {
			m.focusIndex = len(m.positions) - 1
		}
		return form.ResultNone, nil

	case form.FocusHeader:
		// Headers are non-interactive; move focus forward
		m.focusIndex++
		if m.focusIndex >= len(m.positions) {
			m.focusIndex = len(m.positions) - 1
		}
		return form.ResultNone, nil

	default:
		m.focusIndex++
		if m.focusIndex >= len(m.positions) {
			m.focusIndex = len(m.positions) - 1
		}
		return form.ResultNone, nil
	}
}

// HandleChar is a no-op (no text fields in PCI passthrough form).
func (m *PCIPassthroughFormModel) HandleChar(pos form.FocusPos, ch string) {}

// HandleBackspace is a no-op (no text fields in PCI passthrough form).
func (m *PCIPassthroughFormModel) HandleBackspace(pos form.FocusPos) {}

// HandleDelete is a no-op (no text fields in PCI passthrough form).
func (m *PCIPassthroughFormModel) HandleDelete(pos form.FocusPos) {}

// HandleMessage handles async messages (e.g., kernel apply results).
func (m *PCIPassthroughFormModel) HandleMessage(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case PCIVFIOKernelAppliedMsg:
		if msg.Success {
			m.kernelMsg = "vfio-pci.ids applied to grub.cfg successfully"
			m.kernelMsgErr = false
		} else {
			m.kernelMsg = msg.Error
			m.kernelMsgErr = true
		}
		return nil
	}
	return nil
}
