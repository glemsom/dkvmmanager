// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// This file contains the VMFormModel Update/View handlers that tie together
// the model, validation, and UI components.

// openFilePicker creates and activates the appropriate file browser for the given position
func (m *VMFormModel) openFilePicker(pos focusPos) (tea.Model, tea.Cmd) {
	m.browsingFieldName = pos.fieldName
	m.browsingIndex = pos.listIndex

	if pos.fieldName == "hardDisks" {
		m.addDiskModel = NewAddDiskModel(m.vmManager)
		return m, m.addDiskModel.Init()
	}

	// cdroms
	m.fileBrowser = NewFileBrowserModel(FileTypeISO)
	return m, m.fileBrowser.Init()
}

// handleFileSelected processes the result from the ISO file browser
func (m *VMFormModel) handleFileSelected(msg FileSelectedMsg) (tea.Model, tea.Cmd) {
	m.fileBrowser = nil
	if !msg.Canceled && m.browsingFieldName == "cdroms" && m.browsingIndex < len(m.cdroms) {
		m.cdroms[m.browsingIndex] = msg.Path
		m.rebuildPositions()
		m.syncViewport()
	}
	return m, nil
}

// handleDiskAdded processes the result from the add disk model
func (m *VMFormModel) handleDiskAdded(msg DiskAddedMsg) (tea.Model, tea.Cmd) {
	m.addDiskModel = nil
	if !msg.Canceled && m.browsingFieldName == "hardDisks" && m.browsingIndex < len(m.hardDisks) {
		m.hardDisks[m.browsingIndex] = msg.Path
		m.rebuildPositions()
		m.syncViewport()
	}
	return m, nil
}

// Update implements tea.Model
func (m *VMFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case FileSelectedMsg:
		return m.handleFileSelected(msg)

	case DiskAddedMsg:
		return m.handleDiskAdded(msg)

	case tea.KeyMsg:
		// Delegate to active file browser / add disk model first
		if m.fileBrowser != nil && m.fileBrowser.active {
			inner, cmd := m.fileBrowser.Update(msg)
			if fb, ok := inner.(*FileBrowserModel); ok {
				m.fileBrowser = fb
			}
			return m, cmd
		}
		if m.addDiskModel != nil && m.addDiskModel.active {
			inner, cmd := m.addDiskModel.Update(msg)
			if adm, ok := inner.(*AddDiskModel); ok {
				m.addDiskModel = adm
			}
			return m, cmd
		}
		return m.handleKey(msg)
	}
	return m, nil
}

// handleKey processes keyboard input
func (m *VMFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// ESC is handled by MainModel (returnFromSubView), never reach here

	switch key {
	case "tab":
		m.moveFocusNextEditable()
		m.syncViewport()
	case "shift+tab":
		m.moveFocusPrevEditable()
		m.syncViewport()
	case "up":
		m.moveFocus(-1)
		m.syncViewport()
	case "down":
		m.moveFocus(1)
		m.syncViewport()
	case " ":
		pos := m.currentPos()
		if pos.kind == focusText && pos.fieldName != "hardDisks_label" && pos.fieldName != "cdroms_label" {
			// Insert space in text fields
			m.handleCharInput(" ")
			m.syncViewport()
		} else {
			// Act like enter on buttons/toggles/labels
			return m.handleEnter()
		}
	case "enter":
		return m.handleEnter()
	case "backspace":
		m.handleBackspace()
		m.syncViewport()
	case "delete":
		m.handleDelete()
	default:
		if len(key) == 1 {
			m.handleCharInput(key)
			m.syncViewport()
		}
	}
	return m, nil
}

// handleEnter acts contextually: add item, save, or move to next field
func (m *VMFormModel) handleEnter() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	switch pos.kind {
	case focusAddBtn:
		m.addItem(pos.fieldName)
		m.syncViewport()
		return m, nil
	case focusToggle:
		m.toggleValue(pos.fieldName)
		m.syncViewport()
		return m, nil
	case focusSaveBtn:
		return m.validateAndSave()
	case focusListItem:
		return m.openFilePicker(pos)
	default:
		// On a text field: move to next editable position
		m.moveFocusNextEditable()
		m.syncViewport()
		return m, nil
	}
}

// handleBackspace deletes the character before cursor in the focused text field
func (m *VMFormModel) handleBackspace() {
	pos := m.currentPos()
	if pos.kind == focusAddBtn || pos.kind == focusSaveBtn {
		return
	}
	val := m.getValue(pos)
	key := posKey(pos)
	cursor := m.effectiveCursor(key, val)

	if cursor > 0 {
		newVal := val[:cursor-1] + val[cursor:]
		m.setValue(pos, newVal)
		m.setCursorOffset(key, cursor-1)
	}
}

// handleDelete acts on list items (remove) or deletes char ahead of cursor
func (m *VMFormModel) handleDelete() {
	pos := m.currentPos()
	if pos.kind == focusListItem {
		m.removeListItem(pos.fieldName, pos.listIndex)
		m.syncViewport()
		return
	}
	if pos.kind == focusAddBtn || pos.kind == focusSaveBtn {
		return
	}
	val := m.getValue(pos)
	key := posKey(pos)
	cursor := m.effectiveCursor(key, val)

	if cursor < len(val) {
		newVal := val[:cursor] + val[cursor+1:]
		m.setValue(pos, newVal)
		// cursor stays at same position
	}
}

// handleCharInput inserts a character at the cursor in the focused field
func (m *VMFormModel) handleCharInput(ch string) {
	pos := m.currentPos()
	if pos.kind == focusAddBtn || pos.kind == focusSaveBtn {
		return
	}
	val := m.getValue(pos)
	key := posKey(pos)
	cursor := m.effectiveCursor(key, val)

	newVal := val[:cursor] + ch + val[cursor:]
	m.setValue(pos, newVal)
	m.setCursorOffset(key, cursor+1)
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

// moveFocusNextEditable moves to the next position that is not a label
func (m *VMFormModel) moveFocusNextEditable() {
	for i := m.focusIndex + 1; i < len(m.positions); i++ {
		if m.positions[i].fieldName != "hardDisks_label" && m.positions[i].fieldName != "cdroms_label" {
			m.focusIndex = i
			return
		}
	}
}

// moveFocusPrevEditable moves to the previous position that is not a label
func (m *VMFormModel) moveFocusPrevEditable() {
	for i := m.focusIndex - 1; i >= 0; i-- {
		if m.positions[i].fieldName != "hardDisks_label" && m.positions[i].fieldName != "cdroms_label" {
			m.focusIndex = i
			return
		}
	}
}

// addItem appends an empty item to a list field and rebuilds positions
func (m *VMFormModel) addItem(fieldName string) {
	switch fieldName {
	case "hardDisks":
		m.hardDisks = append(m.hardDisks, "")
		m.rebuildPositions()
		// Focus the new item (the add button shifts down by 1)
		for i, p := range m.positions {
			if p.kind == focusListItem && p.fieldName == "hardDisks" && p.listIndex == len(m.hardDisks)-1 {
				m.focusIndex = i
				return
			}
		}
	case "cdroms":
		m.cdroms = append(m.cdroms, "")
		m.rebuildPositions()
		for i, p := range m.positions {
			if p.kind == focusListItem && p.fieldName == "cdroms" && p.listIndex == len(m.cdroms)-1 {
				m.focusIndex = i
				return
			}
		}
	}
}

// removeListItem removes an item from a list field and rebuilds positions
func (m *VMFormModel) removeListItem(fieldName string, index int) {
	switch fieldName {
	case "hardDisks":
		if len(m.hardDisks) <= 1 {
			// Keep at least one slot, just clear it
			m.hardDisks[0] = ""
			key := posKey(focusPos{kind: focusListItem, fieldName: "hardDisks", listIndex: 0})
			m.setCursorOffset(key, 0)
		} else {
			m.hardDisks = append(m.hardDisks[:index], m.hardDisks[index+1:]...)
		}
	case "cdroms":
		if index < len(m.cdroms) {
			m.cdroms = append(m.cdroms[:index], m.cdroms[index+1:]...)
		}
	}
	m.rebuildPositions()
	// Clamp focus
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// toggleValue toggles a boolean field
func (m *VMFormModel) toggleValue(fieldName string) {
	switch fieldName {
	case "vncEnabled":
		m.vncEnabled = !m.vncEnabled
	case "networkMode":
		if m.networkMode == "nat" {
			m.networkMode = "bridge"
		} else {
			m.networkMode = "nat"
		}
	}
}

// View implements tea.Model
func (m *VMFormModel) View() string {
	if m.fileBrowser != nil && m.fileBrowser.active {
		return m.fileBrowser.View()
	}
	if m.addDiskModel != nil && m.addDiskModel.active {
		return m.addDiskModel.View()
	}
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// SetSize updates the form dimensions
func (m *VMFormModel) SetSize(w, h int) {
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

// unused field to satisfy compiler (vmManager reference needed)
var _ = vm.Manager{}

var (
	formLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	formFocusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
	formInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	formErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	formMutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	formSaveStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
)
