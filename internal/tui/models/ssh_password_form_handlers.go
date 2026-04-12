// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// sshPasswordErrorMsg is sent when password change fails
type sshPasswordErrorMsg struct {
	err string
}

// Update implements tea.Model
func (m *SSHPasswordFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case sshPasswordErrorMsg:
		m.applying = false
		m.statusMessage = msg.err
		m.syncViewport()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey processes keyboard input
func (m *SSHPasswordFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.applying {
		return m, nil
	}

	key := msg.String()

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

// handleEnter acts contextually: apply or move to next field
func (m *SSHPasswordFormModel) handleEnter() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	switch pos.kind {
	case sshPwApply:
		if !m.validate() {
			m.syncViewport()
			return m, nil
		}
		m.applying = true
		m.statusMessage = ""
		m.syncViewport()
		return m, m.apply()
	default:
		m.moveFocus(1)
		m.syncViewport()
		return m, nil
	}
}

// handleBackspace deletes the character before cursor in the focused text field
func (m *SSHPasswordFormModel) handleBackspace() {
	pos := m.currentPos()
	if pos.kind != sshPwText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	key := pos.fieldName
	cursor := m.effectiveCursor(key, val)

	if cursor > 0 {
		newVal := val[:cursor-1] + val[cursor:]
		m.setTextValue(pos.fieldName, newVal)
		m.setCursorOffset(key, cursor-1)
	}
}

// handleDelete deletes the character ahead of cursor in the focused text field
func (m *SSHPasswordFormModel) handleDelete() {
	pos := m.currentPos()
	if pos.kind != sshPwText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	key := pos.fieldName
	cursor := m.effectiveCursor(key, val)

	if cursor < len(val) {
		newVal := val[:cursor] + val[cursor+1:]
		m.setTextValue(pos.fieldName, newVal)
	}
}

// handleCharInput inserts a character at the cursor in the focused text field
func (m *SSHPasswordFormModel) handleCharInput(ch string) {
	pos := m.currentPos()
	if pos.kind != sshPwText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	key := pos.fieldName
	cursor := m.effectiveCursor(key, val)

	newVal := val[:cursor] + ch + val[cursor:]
	m.setTextValue(pos.fieldName, newVal)
	m.setCursorOffset(key, cursor+1)
}
