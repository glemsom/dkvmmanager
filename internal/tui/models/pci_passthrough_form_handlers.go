// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

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
		return m.handleEnter()
	case "backspace":
		m.handleBackspace()
		m.syncViewport()
	case "delete":
		m.handleDelete()
		m.syncViewport()
	default:
		if len(key) == 1 {
			m.handleCharInput(key)
			m.syncViewport()
		}
	}
	return m, nil
}

// handleEnter acts contextually: toggle or save
func (m *PCIPassthroughFormModel) handleEnter() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	switch pos.kind {
	case pciToggle:
		m.toggleDevice(pos.deviceAddr)
		m.rebuildPositions()
		m.syncViewport()
		return m, nil
	case pciSave:
		return m.validateAndSave()
	default:
		// On a ROM path field: move to next field
		m.moveFocus(1)
		m.syncViewport()
		return m, nil
	}
}

// handleBackspace deletes the character before cursor in the focused ROM path field
func (m *PCIPassthroughFormModel) handleBackspace() {
	pos := m.currentPos()
	if pos.kind != pciROMPath {
		return
	}
	val := m.romPaths[pos.deviceAddr]
	key := fmt.Sprintf("rom_%s", pos.deviceAddr)
	cursor := m.effectiveCursor(key, val)

	if cursor > 0 {
		newVal := val[:cursor-1] + val[cursor:]
		m.romPaths[pos.deviceAddr] = newVal
		m.setCursorOffset(key, cursor-1)
	}
}

// handleDelete deletes the character ahead of cursor in the focused ROM path field
func (m *PCIPassthroughFormModel) handleDelete() {
	pos := m.currentPos()
	if pos.kind != pciROMPath {
		return
	}
	val := m.romPaths[pos.deviceAddr]
	key := fmt.Sprintf("rom_%s", pos.deviceAddr)
	cursor := m.effectiveCursor(key, val)

	if cursor < len(val) {
		newVal := val[:cursor] + val[cursor+1:]
		m.romPaths[pos.deviceAddr] = newVal
	}
}

// handleCharInput inserts a character at the cursor in the focused ROM path field
func (m *PCIPassthroughFormModel) handleCharInput(ch string) {
	pos := m.currentPos()
	if pos.kind != pciROMPath {
		return
	}
	val := m.romPaths[pos.deviceAddr]
	key := fmt.Sprintf("rom_%s", pos.deviceAddr)
	cursor := m.effectiveCursor(key, val)

	newVal := val[:cursor] + ch + val[cursor:]
	m.romPaths[pos.deviceAddr] = newVal
	m.setCursorOffset(key, cursor+1)
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
