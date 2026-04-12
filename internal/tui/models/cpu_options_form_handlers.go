// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model
func (m *CPUOptionsFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *CPUOptionsFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case "pgup":
		m.pageUp()
	case "pgdown":
		m.pageDown()
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
func (m *CPUOptionsFormModel) handleEnter() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	switch pos.kind {
	case cpuOptToggle:
		m.toggleValue(pos.fieldName)
		m.syncViewport()
		return m, nil
	case cpuOptSave:
		return m.validateAndSave()
	default:
		// On a text field: move to next field
		m.moveFocus(1)
		m.syncViewport()
		return m, nil
	}
}

// handleBackspace deletes the character before cursor in the focused text field
func (m *CPUOptionsFormModel) handleBackspace() {
	pos := m.currentPos()
	if pos.kind != cpuOptText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	key := cpuOptPosKey(pos)
	cursor := m.effectiveCursor(key, val)

	if cursor > 0 {
		newVal := val[:cursor-1] + val[cursor:]
		m.setTextValue(pos.fieldName, newVal)
		m.setCursorOffset(key, cursor-1)
	}
}

// handleDelete deletes the character ahead of cursor in the focused text field
func (m *CPUOptionsFormModel) handleDelete() {
	pos := m.currentPos()
	if pos.kind != cpuOptText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	key := cpuOptPosKey(pos)
	cursor := m.effectiveCursor(key, val)

	if cursor < len(val) {
		newVal := val[:cursor] + val[cursor+1:]
		m.setTextValue(pos.fieldName, newVal)
	}
}

// handleCharInput inserts a character at the cursor in the focused text field
func (m *CPUOptionsFormModel) handleCharInput(ch string) {
	pos := m.currentPos()
	if pos.kind != cpuOptText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	key := cpuOptPosKey(pos)
	cursor := m.effectiveCursor(key, val)

	newVal := val[:cursor] + ch + val[cursor:]
	m.setTextValue(pos.fieldName, newVal)
	m.setCursorOffset(key, cursor+1)
}

// validateAndSave persists the CPU options
func (m *CPUOptionsFormModel) validateAndSave() (tea.Model, tea.Cmd) {
	m.errors = make(map[string]string)

	if err := m.vmManager.SaveCPUOptions(m.options); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		m.syncViewport()
		return m, nil
	}

	return m, func() tea.Msg { return CPUOptionsUpdatedMsg{} }
}
