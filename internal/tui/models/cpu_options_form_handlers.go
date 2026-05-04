// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// handleKey processes keyboard input (backward-compatible Update path).
func (m *CPUOptionsFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.saving {
		return m, nil
	}

	key := msg.String()

	switch key {
	case "tab", "down":
		m.moveFocus(1)
		m.syncViewport()
	case "shift+tab", "up":
		m.moveFocus(-1)
		m.syncViewport()
	case "pgup":
		m.vp.HalfPageUp()
		return m, nil
	case "pgdown":
		m.vp.HalfPageDown()
		return m, nil
	case "enter", " ":
		return m.handleEnterKey()
	case "backspace":
		m.handleBackspaceKey()
	case "delete":
		m.handleDeleteKey()
	default:
		if len(key) == 1 {
			m.handleCharInput(key)
		}
	}
	return m, nil
}

// handleEnterKey acts contextually: toggle or save (backward compat).
func (m *CPUOptionsFormModel) handleEnterKey() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	switch pos.kind {
	case cpuOptToggle:
		m.toggleValue(pos.fieldName)
		return m, nil
	case cpuOptSave:
		return m.validateAndSaveCmd()
	default:
		// On a text field: move to next field
		m.moveFocus(1)
		return m, nil
	}
}

// handleBackspaceKey deletes the character before cursor (backward compat).
func (m *CPUOptionsFormModel) handleBackspaceKey() {
	pos := m.currentPos()
	if pos.kind != cpuOptText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	cursor := m.effectiveCursor(pos.fieldName, val)

	if cursor > 0 {
		newVal := val[:cursor-1] + val[cursor:]
		m.setTextValue(pos.fieldName, newVal)
		m.setCursorOffset(pos.fieldName, cursor-1)
	}
}

// handleDeleteKey deletes the character at cursor (backward compat).
func (m *CPUOptionsFormModel) handleDeleteKey() {
	pos := m.currentPos()
	if pos.kind != cpuOptText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	cursor := m.effectiveCursor(pos.fieldName, val)

	if cursor < len(val) {
		newVal := val[:cursor] + val[cursor+1:]
		m.setTextValue(pos.fieldName, newVal)
	}
}

// handleCharInput inserts a character at cursor (backward compat).
func (m *CPUOptionsFormModel) handleCharInput(ch string) {
	pos := m.currentPos()
	if pos.kind != cpuOptText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	cursor := m.effectiveCursor(pos.fieldName, val)

	newVal := val[:cursor] + ch + val[cursor:]
	m.setTextValue(pos.fieldName, newVal)
	m.setCursorOffset(pos.fieldName, cursor+1)
}

// validateAndSaveCmd persists the CPU options and returns a tea.Cmd.
func (m *CPUOptionsFormModel) validateAndSaveCmd() (tea.Model, tea.Cmd) {
	m.errors = make(map[string]string)

	if err := m.vmManager.SaveCPUOptions(m.options); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		return m, nil
	}

	return m, func() tea.Msg { return CPUOptionsUpdatedMsg{} }
}

// --- FormModel Interface Implementation ---

// HandleEnter is called when the user presses Enter on a position.
func (m *CPUOptionsFormModel) HandleEnter(pos form.FocusPos) (form.FormResult, tea.Cmd) {
	switch pos.Kind {
	case form.FocusToggle:
		m.toggleValue(pos.Key)
		return form.ResultNone, nil
	case form.FocusButton:
		if pos.Key == "save" {
			m.errors = make(map[string]string)
			if err := m.vmManager.SaveCPUOptions(m.options); err != nil {
				m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
				return form.ResultNone, nil
			}
			return form.ResultNone, func() tea.Msg { return CPUOptionsUpdatedMsg{} }
		}
		return form.ResultNone, nil
	default:
		// On a text field: move to next position
		m.focusIndex++
		if m.focusIndex >= len(m.positions) {
			m.focusIndex = len(m.positions) - 1
		}
		return form.ResultNone, nil
	}
}

// HandleChar inserts a character at the cursor in the focused text field.
func (m *CPUOptionsFormModel) HandleChar(pos form.FocusPos, ch string) {
	if pos.Kind != form.FocusText {
		return
	}
	val := m.getTextValue(pos.Key)
	cursor := m.effectiveCursor(pos.Key, val)
	newVal := val[:cursor] + ch + val[cursor:]
	m.setTextValue(pos.Key, newVal)
	m.cursorOffsets[pos.Key] = cursor + 1
}

// HandleBackspace deletes the character before cursor.
func (m *CPUOptionsFormModel) HandleBackspace(pos form.FocusPos) {
	if pos.Kind != form.FocusText {
		return
	}
	val := m.getTextValue(pos.Key)
	cursor := m.effectiveCursor(pos.Key, val)
	if cursor > 0 {
		newVal := val[:cursor-1] + val[cursor:]
		m.setTextValue(pos.Key, newVal)
		m.cursorOffsets[pos.Key] = cursor - 1
	}
}

// HandleDelete deletes the character at cursor.
func (m *CPUOptionsFormModel) HandleDelete(pos form.FocusPos) {
	if pos.Kind != form.FocusText {
		return
	}
	val := m.getTextValue(pos.Key)
	cursor := m.effectiveCursor(pos.Key, val)
	if cursor < len(val) {
		newVal := val[:cursor] + val[cursor+1:]
		m.setTextValue(pos.Key, newVal)
		// Cursor stays at same position
	}
}

// HandleMessage handles custom messages (e.g., async command results).
func (m *CPUOptionsFormModel) HandleMessage(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case cpuOptionsErrorMsg:
		m.saving = false
		m.statusMessage = msg.err
		return nil
	}
	return nil
}

// cpuOptionsErrorMsg is sent when save fails.
type cpuOptionsErrorMsg struct {
	err string
}
