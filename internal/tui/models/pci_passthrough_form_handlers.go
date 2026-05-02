// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model
func (m *PCIPassthroughFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.contentW = msg.Width
		m.contentH = msg.Height
		if !m.ready {
			m.vp = viewport.New(msg.Width, msg.Height)
			m.ready = true
		} else {
			m.vp.Width = msg.Width
			m.vp.Height = msg.Height
		}
		m.syncViewport()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case PCIVFIOKernelAppliedMsg:
		if msg.Success {
			m.kernelMsg = "vfio-pci.ids applied to grub.cfg successfully"
			m.kernelMsgErr = false
		} else {
			m.kernelMsg = msg.Error
			m.kernelMsgErr = true
		}
		m.syncViewport()
		return m, nil
	}
	return m, nil
}

// handleKey processes keyboard input
func (m *PCIPassthroughFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// ESC is handled by MainModel (returnFromSubView), never reach here

	switch key {
	case "tab":
		m.moveFocus(1)
		m.syncViewport()
	case "shift+tab":
		m.moveFocus(-1)
		m.syncViewport()
	case "up":
		m.moveFocus(-1)
		m.syncViewport()
	case "down":
		m.moveFocus(1)
		m.syncViewport()
	case "enter", " ":
		return m.handleEnterOrApply()
	default:
		// No text input fields remaining (ROM removed)
	}
	return m, nil
}

// handleEnterOrApply acts contextually: toggle, save, or apply kernel
func (m *PCIPassthroughFormModel) handleEnterOrApply() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	switch pos.kind {
	case pciToggle:
		m.toggleDevice(pos.deviceAddr)
		m.rebuildPositions()
		m.syncViewport()
		return m, nil
	case pciSave:
		return m.validateAndSave()
	case pciApplyKernel:
		return m.handleApplyKernel()
	default:
		// Skip non-focusable positions
		m.moveFocus(1)
		m.syncViewport()
		return m, nil
	}
}

// View implements tea.Model
func (m *PCIPassthroughFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// SetSize updates the form dimensions
func (m *PCIPassthroughFormModel) SetSize(w, h int) {
	m.contentW = w
	m.contentH = h
	if !m.ready {
		m.vp = viewport.New(w, h)
		m.ready = true
	} else {
		m.vp.Width = w
		m.vp.Height = h
	}
	m.syncViewport()
}
