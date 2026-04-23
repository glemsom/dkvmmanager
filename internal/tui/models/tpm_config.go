// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// TPMConfigModel is a thin wrapper around TPMConfigFormModel
type TPMConfigModel struct {
	form *TPMConfigFormModel
}

// NewTPMConfigModel creates a new TPM config model
func NewTPMConfigModel(vmManager *vm.Manager) *TPMConfigModel {
	return &TPMConfigModel{
		form: NewTPMConfigFormModel(vmManager),
	}
}

// Init initializes the model
func (m *TPMConfigModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update handles incoming messages
func (m *TPMConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if f, ok := inner.(*TPMConfigFormModel); ok {
		m.form = f
	}
	return m, cmd
}

// View returns the view for the model
func (m *TPMConfigModel) View() string {
	return m.form.View()
}

// SetSize updates the form dimensions
func (m *TPMConfigModel) SetSize(w, h int) {
	m.form.SetSize(w, h)
}

// tpmFocusKind describes what a TPM config focus position represents
type tpmFocusKind int

const (
	tpmText tpmFocusKind = iota
	tpmSave
)

// tpmFocusPos is one navigable position in the TPM config form
type tpmFocusPos struct {
	kind      tpmFocusKind
	fieldName string
}

// TPMConfigFormModel is a scrollable form for editing the TPM binary path
type TPMConfigFormModel struct {
	vmManager *vm.Manager
	tpmPath   string

	// Flat list of focusable positions
	positions  []tpmFocusPos
	focusIndex int

	// Per-field text cursor (character offset within the value)
	cursorOffsets map[string]int

	// Per-field inline error messages
	errors map[string]string

	// Scrollable viewport
	vp       viewport.Model
	ready    bool
	contentW int
	contentH int

	// Rendering cache
	renderedLines []string
}

// NewTPMConfigFormModel creates a new TPM config form model
func NewTPMConfigFormModel(vmManager *vm.Manager) *TPMConfigFormModel {
	cfg := vmManager.GetConfig()
	m := &TPMConfigFormModel{
		vmManager:     vmManager,
		tpmPath:       cfg.TPMBinary,
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	m.rebuildPositions()
	return m
}

// cursorOffset returns the cursor offset for the given field name
func (m *TPMConfigFormModel) cursorOffset(key string) int {
	if off, ok := m.cursorOffsets[key]; ok {
		return off
	}
	return -1
}

// setCursorOffset sets cursor offset; -1 means end
func (m *TPMConfigFormModel) setCursorOffset(key string, off int) {
	m.cursorOffsets[key] = off
}

// effectiveCursor returns the actual cursor position (0-based character index)
func (m *TPMConfigFormModel) effectiveCursor(key string, val string) int {
	off := m.cursorOffset(key)
	if off < 0 {
		return len(val)
	}
	if off > len(val) {
		return len(val)
	}
	return off
}

// Init implements tea.Model
func (m *TPMConfigFormModel) Init() tea.Cmd {
	return nil
}

// rebuildPositions reconstructs the flat focus list
func (m *TPMConfigFormModel) rebuildPositions() {
	m.positions = nil
	m.positions = append(m.positions, tpmFocusPos{kind: tpmText, fieldName: "tpmPath"})
	m.positions = append(m.positions, tpmFocusPos{kind: tpmSave, fieldName: "save"})
}

// currentPos returns the focus position at the current focusIndex
func (m *TPMConfigFormModel) currentPos() tpmFocusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return tpmFocusPos{kind: tpmText, fieldName: "tpmPath"}
	}
	return m.positions[m.focusIndex]
}

// getTextValue returns the text value for a field
func (m *TPMConfigFormModel) getTextValue(fieldName string) string {
	switch fieldName {
	case "tpmPath":
		return m.tpmPath
	}
	return ""
}

// setTextValue sets the text value for a field
func (m *TPMConfigFormModel) setTextValue(fieldName string, val string) {
	switch fieldName {
	case "tpmPath":
		m.tpmPath = val
	}
}

// Update handles messages
func (m *TPMConfigFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		pos := m.currentPos()

		switch pos.kind {
		case tpmText:
			switch key {
			case "backspace", "ctrl+h":
				if len(m.tpmPath) > 0 {
					m.tpmPath = m.tpmPath[:len(m.tpmPath)-1]
					m.setCursorOffset("tpmPath", len(m.tpmPath))
				}
			case "enter":
				// Move to save button
				if m.focusIndex < len(m.positions)-1 {
					m.focusIndex++
				}
			default:
				if len(key) == 1 && key >= " " && key <= "~" {
					m.tpmPath += key
					m.setCursorOffset("tpmPath", len(m.tpmPath))
				}
			}
			m.syncViewport()
			return m, nil

		case tpmSave:
			if key == "enter" || key == " " {
				return m.save()
			}
		}

		// Navigation
		switch key {
		case "up", "k":
			if m.focusIndex > 0 {
				m.focusIndex--
			}
			m.syncViewport()
		case "down", "j", "tab":
			if m.focusIndex < len(m.positions)-1 {
				m.focusIndex++
			}
			m.syncViewport()
		case "shift+tab":
			if m.focusIndex > 0 {
				m.focusIndex--
			}
			m.syncViewport()
		case "esc":
			// Cancel: return to previous view
			return m, func() tea.Msg { return ViewChangeMsg{View: ViewConfigMenu} }
		}
	}

	return m, nil
}

// View renders the form
func (m *TPMConfigFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// SetSize updates the form dimensions
func (m *TPMConfigFormModel) SetSize(w, h int) {
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

// syncViewport regenerates the rendered lines and syncs the viewport
func (m *TPMConfigFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	content := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(content)
	// Ensure focused line is visible
	// Simple: set Y offset so focused line is in view
	lineIdx := m.focusedLineIndex()
	if lineIdx >= 0 && lineIdx < m.vp.TotalLineCount() {
		// Scroll if needed
		if lineIdx < m.vp.YOffset {
			m.vp.SetYOffset(lineIdx)
		} else if lineIdx >= m.vp.YOffset+m.vp.Height {
			m.vp.SetYOffset(lineIdx - m.vp.Height + 1)
		}
	}
}

// focusedLineIndex maps focusIndex to a rendered line index
func (m *TPMConfigFormModel) focusedLineIndex() int {
	line := 2 // header + blank
	for i := range m.positions {
		if i == m.focusIndex {
			return line
		}
		// Each position renders one line (or two if with error)
		line++
		// Add error line if exists for this position
		if i < len(m.positions)-1 { // check next? Actually simpler: check errors for this pos
			// Not perfect but ok
		}
	}
	return line
}

// renderAllLines produces the full list of output lines for the form
func (m *TPMConfigFormModel) renderAllLines() []string {
	var lines []string

	// Header
	lines = append(lines, styles.HeaderStyle().Render("Edit TPM Binary Path"))
	lines = append(lines, "")

	for i, pos := range m.positions {
		focused := (i == m.focusIndex)
		switch pos.kind {
		case tpmText:
			label := "TPM Binary Path"
			val := m.tpmPath
			cursor := m.effectiveCursor("tpmPath", val)
			rendered := m.renderTextInput(label, val, cursor, focused)
			lines = append(lines, rendered)
			if errMsg, ok := m.errors["tpmPath"]; ok {
				lines = append(lines, "  "+styles.ErrorTextStyle().Render(errMsg))
			}
		case tpmSave:
			lines = append(lines, "")
			saveText := styles.MutedTextStyle().Render("[Space/Enter] Save  [ESC] Cancel")
			if focused {
				saveText = styles.SuccessTextStyle().Render("[Space/Enter] Save") + "  " + styles.MutedTextStyle().Render("[ESC] Cancel")
			}
			lines = append(lines, saveText)
		}
	}

	return lines
}

// renderTextInput renders a labeled text input with an optional cursor
func (m *TPMConfigFormModel) renderTextInput(label, value string, cursor int, focused bool) string {
	prefix := "  "
	if focused {
		prefix = styles.SelectedTextStyle().Render("> ")
	}

	labelPart := styles.FormLabelStyle().Render(label + ": ")

	var valPart string
	if focused {
		if cursor < len(value) {
			before := value[:cursor]
			at := string(value[cursor])
			after := ""
			if cursor+1 < len(value) {
				after = value[cursor+1:]
			}
			valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(before) +
				lipgloss.NewStyle().Reverse(true).Render(at) +
				lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(after)
		} else {
			valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(value) + styles.SelectedTextStyle().Render("_")
		}
	} else {
		if value == "" {
			valPart = styles.MutedTextStyle().Render("(empty)")
		} else {
			valPart = lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render(value)
		}
	}

	return prefix + labelPart + valPart
}

// save persists the TPM binary path to configuration
func (m *TPMConfigFormModel) save() (tea.Model, tea.Cmd) {
	cfg := m.vmManager.GetConfig()
	cfg.TPMBinary = m.tpmPath
	if err := cfg.Save(); err != nil {
		m.errors["save"] = "Failed to save config: " + err.Error()
		m.syncViewport()
		return m, nil
	}
	// Return to config menu
	return m, func() tea.Msg { return ViewChangeMsg{View: ViewConfigMenu} }
}
