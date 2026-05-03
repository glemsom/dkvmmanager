// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// HandleEnter acts contextually: add item, save, or move to next field
func (m *VMFormModel) HandleEnter(pos form.FocusPos) (form.FormResult, tea.Cmd) {
	switch pos.Key {
	case "hardDisks_add", "cdroms_add":
		m.addItem(pos.Key)
		return form.ResultNone, nil
	case "save":
		return form.ResultSave, m.validateAndSaveCmd()
	case "vncEnabled", "tpmEnabled", "networkMode":
		m.toggleValue(pos.Key)
		return form.ResultNone, nil
	default:
		// For list items, open file picker
		if len(pos.Key) > 9 && (pos.Key[:9] == "hardDisks" || pos.Key[:6] == "cdroms") {
			// Check if this is a list item (not a label or add button)
			if pos.Kind == form.FocusList {
				return form.ResultNone, m.openFilePickerCmd(pos)
			}
		}
		// On a text field: move to next editable position
		m.moveFocusNextEditable()
		return form.ResultNone, nil
	}
}

// openFilePickerCmd creates and activates the file browser
func (m *VMFormModel) openFilePickerCmd(pos form.FocusPos) tea.Cmd {
	// Parse field name and index from key like "hardDisks_0" or "cdroms_1"
	// Find the last underscore to split field name from index
	fieldName := pos.Key
	idx := 0
	if pos.Kind == form.FocusList {
		lastUnderscore := -1
		for i := len(pos.Key) - 1; i >= 0; i-- {
			if pos.Key[i] == '_' {
				lastUnderscore = i
				break
			}
		}
		if lastUnderscore > 0 {
			fieldName = pos.Key[:lastUnderscore]
			// Parse index from digits after the underscore
			for i := lastUnderscore + 1; i < len(pos.Key) && pos.Key[i] >= '0' && pos.Key[i] <= '9'; i++ {
				idx = idx*10 + int(pos.Key[i]-'0')
			}
		}
	}

	m.browsingFieldName = fieldName
	m.browsingIndex = idx

	if m.browsingFieldName == "hardDisks" {
		m.addDiskModel = NewAddDiskModel(m.vmManager)
		return m.addDiskModel.Init()
	}

	m.fileBrowser = NewFileBrowserModel(FileTypeISO)
	return m.fileBrowser.Init()
}

// HandleChar inserts a character at the cursor in the focused text field
func (m *VMFormModel) HandleChar(pos form.FocusPos, ch string) {
	if pos.Kind != form.FocusText && pos.Kind != form.FocusList {
		return
	}

	val := m.getValue(pos)
	key := pos.Key
	cursor := m.effectiveCursor(key, val)

	newVal := val[:cursor] + ch + val[cursor:]
	m.setValue(pos, newVal)
	m.setCursorOffset(key, cursor+1)
}

// HandleBackspace deletes the character before cursor
func (m *VMFormModel) HandleBackspace(pos form.FocusPos) {
	if pos.Kind == form.FocusButton || pos.Kind == form.FocusHeader {
		return
	}

	val := m.getValue(pos)
	key := pos.Key
	cursor := m.effectiveCursor(key, val)

	if cursor > 0 {
		newVal := val[:cursor-1] + val[cursor:]
		m.setValue(pos, newVal)
		m.setCursorOffset(key, cursor-1)
	}
}

// HandleDelete acts on list items (remove) or deletes char ahead of cursor
func (m *VMFormModel) HandleDelete(pos form.FocusPos) {
	if pos.Kind == form.FocusList {
		m.removeListItemByPos(pos)
		return
	}
	if pos.Kind == form.FocusButton || pos.Kind == form.FocusHeader {
		return
	}

	val := m.getValue(pos)
	key := pos.Key
	cursor := m.effectiveCursor(key, val)

	if cursor < len(val) {
		newVal := val[:cursor] + val[cursor+1:]
		m.setValue(pos, newVal)
	}
}

// HandleMessage handles custom messages (e.g., async command results).
func (m *VMFormModel) HandleMessage(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case FileSelectedMsg:
		return m.handleFileSelectedCmd(msg)
	case DiskAddedMsg:
		return m.handleDiskAddedCmd(msg)
	}
	return nil
}

func (m *VMFormModel) handleFileSelectedCmd(msg FileSelectedMsg) tea.Cmd {
	m.fileBrowser = nil
	if msg.Canceled {
		return nil
	}

	// Find the key for this index
	key := m.browsingFieldName + "_" + string(rune('0'+m.browsingIndex))
	if m.browsingFieldName == "cdroms" && m.browsingIndex < len(m.cdroms) {
		m.cdroms[m.browsingIndex] = msg.Path
		m.rebuildPositions()
	}
	_ = key // avoid unused variable error
	return nil
}

func (m *VMFormModel) handleDiskAddedCmd(msg DiskAddedMsg) tea.Cmd {
	m.addDiskModel = nil
	if msg.Canceled {
		return nil
	}

	if m.browsingFieldName == "hardDisks" && m.browsingIndex < len(m.hardDisks) {
		m.hardDisks[m.browsingIndex] = msg.Path
		m.rebuildPositions()
	}
	return nil
}

// moveFocus moves focus by delta in the flat positions list
func (m *VMFormModel) moveFocus(delta int) {
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// moveFocusNextEditable moves to the next position that is not a label or header
func (m *VMFormModel) moveFocusNextEditable() {
	for i := m.focusIndex + 1; i < len(m.positions); i++ {
		pos := m.positions[i]
		if pos.Kind != form.FocusHeader {
			m.focusIndex = i
			return
		}
	}
}

// addItem appends an empty item to a list field and rebuilds positions
func (m *VMFormModel) addItem(fieldName string) {
	switch fieldName {
	case "hardDisks_add":
		m.hardDisks = append(m.hardDisks, "")
		m.rebuildPositions()
		// Focus the new item
		for i, p := range m.positions {
			if p.Kind == form.FocusList && p.Key == fmt.Sprintf("hardDisks_%d", len(m.hardDisks)-1) {
				m.focusIndex = i
				return
			}
		}
	case "cdroms_add":
		m.cdroms = append(m.cdroms, "")
		m.rebuildPositions()
		for i, p := range m.positions {
			if p.Kind == form.FocusList && p.Key == fmt.Sprintf("cdroms_%d", len(m.cdroms)-1) {
				m.focusIndex = i
				return
			}
		}
	}
}

// removeListItemByPos removes a list item by position
func (m *VMFormModel) removeListItemByPos(pos form.FocusPos) {
	switch {
	case len(pos.Key) > 9 && pos.Key[:9] == "hardDisks":
		m.removeListAt("hardDisks", pos.Key)
	case len(pos.Key) > 6 && pos.Key[:6] == "cdroms":
		m.removeListAt("cdroms", pos.Key)
	}
}

func (m *VMFormModel) removeListAt(fieldName, key string) {
	var idx int
	for i := len(key) - 1; i >= 0 && key[i] >= '0' && key[i] <= '9'; i-- {
		idx = idx*10 + int(key[i]-'0')
	}

	switch fieldName {
	case "hardDisks":
		if len(m.hardDisks) <= 1 {
			m.hardDisks[0] = ""
			m.setCursorOffset(fmt.Sprintf("hardDisks_%d", idx), 0)
		} else {
			m.hardDisks = append(m.hardDisks[:idx], m.hardDisks[idx+1:]...)
		}
	case "cdroms":
		if idx < len(m.cdroms) {
			m.cdroms = append(m.cdroms[:idx], m.cdroms[idx+1:]...)
		}
	}
	m.rebuildPositions()
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// toggleValue toggles a boolean field
func (m *VMFormModel) toggleValue(fieldName string) {
	switch fieldName {
	case "vncEnabled":
		m.vncEnabled = !m.vncEnabled
	case "tpmEnabled":
		m.tpmEnabled = !m.tpmEnabled
	case "networkMode":
		if m.networkMode == "nat" {
			m.networkMode = "bridge"
		} else {
			m.networkMode = "nat"
		}
	}
}

// unused field to satisfy compiler (vmManager reference needed)
var _ = vm.Manager{}