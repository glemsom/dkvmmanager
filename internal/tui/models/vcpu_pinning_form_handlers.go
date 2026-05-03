// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// HandleEnter is called when the user presses Enter on a position.
func (m *VCPUPinningFormModel) HandleEnter(pos form.FocusPos) (form.FormResult, tea.Cmd) {
	switch pos.Kind {
	case form.FocusToggle:
		m.pinning.Enabled = !m.pinning.Enabled
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

	default:
		m.focusIndex++
		if m.focusIndex >= len(m.positions) {
			m.focusIndex = len(m.positions) - 1
		}
		return form.ResultNone, nil
	}
}

// HandleChar is a no-op (no text fields in vCPU pinning form).
func (m *VCPUPinningFormModel) HandleChar(pos form.FocusPos, ch string) {}

// HandleBackspace is a no-op (no text fields in vCPU pinning form).
func (m *VCPUPinningFormModel) HandleBackspace(pos form.FocusPos) {}

// HandleDelete is a no-op (no text fields in vCPU pinning form).
func (m *VCPUPinningFormModel) HandleDelete(pos form.FocusPos) {}

// HandleMessage handles async messages (e.g., kernel apply results).
func (m *VCPUPinningFormModel) HandleMessage(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case VCPUCPUKernelAppliedMsg:
		if msg.Success {
			if m.pinning.Enabled && len(m.pinning.Mappings) > 0 {
				m.kernelMsg = "Kernel CPU isolation parameters applied to grub.cfg"
			} else {
				m.kernelMsg = "Kernel CPU isolation parameters removed from grub.cfg"
			}
			m.kernelMsgErr = false
		} else {
			m.kernelMsg = msg.Error
			m.kernelMsgErr = true
		}
		return nil
	}
	return nil
}
