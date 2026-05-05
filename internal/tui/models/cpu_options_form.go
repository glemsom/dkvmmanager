// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/fields"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// CPUOptionsFormModel is a scrollable toggle form for editing global CPU options.
// It implements the form.FormModel interface for use with ScrollableForm.
type CPUOptionsFormModel struct {
	vmManager *vm.Manager
	options   *models.CPUOptions

	// Focus state
	positions  []form.FocusPos
	focusIndex int

	// Per-field text cursor (character offset within the value)
	cursorOffsets map[string]int

	// Per-field inline error messages
	errors map[string]string

	// Save status message (for save errors)
	statusMessage string

	// Whether a save operation is in progress
	saving bool

	// Size (for viewport sync, used by framework's SetSize)
	contentW int
	contentH int
	vp       viewport.Model
	ready    bool

	// Rendering cache (for backward-compatible View)
	renderedLines []string
}

// NewCPUOptionsFormModel creates a new CPU options form model
func NewCPUOptionsFormModel(vmManager *vm.Manager) *CPUOptionsFormModel {
	opts, _ := vmManager.GetCPUOptions()
	m := &CPUOptionsFormModel{
		vmManager:     vmManager,
		options:       &opts,
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	m.positions = m.BuildPositions()
	// Start focus at first interactive element (skip section header)
	m.focusIndex = 1
	return m
}

// getBoolField returns a boolean field value by name using reflection.
func (m *CPUOptionsFormModel) getBoolField(name string) bool {
	v := reflect.ValueOf(m.options).Elem().FieldByName(name)
	if !v.IsValid() {
		panic("field not found: " + name)
	}
	return v.Bool()
}

// setBoolField sets a boolean field value by name using reflection.
func (m *CPUOptionsFormModel) setBoolField(name string, val bool) {
	v := reflect.ValueOf(m.options).Elem().FieldByName(name)
	if !v.IsValid() {
		panic("field not found: " + name)
	}
	v.SetBool(val)
}

// getStringField returns a string field value by name using reflection.
func (m *CPUOptionsFormModel) getStringField(name string) string {
	v := reflect.ValueOf(m.options).Elem().FieldByName(name)
	if !v.IsValid() {
		panic("field not found: " + name)
	}
	return v.String()
}

// setStringField sets a string field value by name using reflection.
func (m *CPUOptionsFormModel) setStringField(name string, val string) {
	v := reflect.ValueOf(m.options).Elem().FieldByName(name)
	if !v.IsValid() {
		panic("field not found: " + name)
	}
	v.SetString(val)
}

// cursorOffset returns the cursor offset for the given position key.
func (m *CPUOptionsFormModel) cursorOffset(key string) int {
	if off, ok := m.cursorOffsets[key]; ok {
		return off
	}
	return -1 // sentinel meaning "end"
}

// setCursorOffset sets cursor offset; -1 means end.
func (m *CPUOptionsFormModel) setCursorOffset(key string, off int) {
	m.cursorOffsets[key] = off
}

// effectiveCursor returns the actual cursor position (0-based character index).
// -1 means cursor at end (default).
func (m *CPUOptionsFormModel) effectiveCursor(key string, val string) int {
	off := m.cursorOffset(key)
	if off < 0 {
		return len(val)
	}
	if off > len(val) {
		return len(val)
	}
	return off
}

// CurrentIndex returns the index of the currently focused position.
func (m *CPUOptionsFormModel) CurrentIndex() int {
	return m.focusIndex
}

// SetFocusIndex sets the focused position index.
func (m *CPUOptionsFormModel) SetFocusIndex(i int) {
	m.focusIndex = i
}

// OnEnter is called when the form becomes active.
func (m *CPUOptionsFormModel) OnEnter() {}

// OnExit is called when the form is dismissed.
func (m *CPUOptionsFormModel) OnExit() {}

// SetSize updates the form dimensions.
func (m *CPUOptionsFormModel) SetSize(w, h int) {
	m.contentW = w
	m.contentH = h
	if !m.ready {
		m.vp = viewport.New(w, h)
		m.ready = true
	} else {
		m.vp.Width = w
		m.vp.Height = h
	}
}

// SetFocused sets whether the form has keyboard focus.
func (m *CPUOptionsFormModel) SetFocused(bool) {}

// syncViewport regenerates the rendered lines and syncs the viewport.
func (m *CPUOptionsFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	totalContent := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(totalContent)
	if m.focusedLineIndex() >= 0 {
		m.vp.YOffset = form.ClampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height)
	}
}

// focusedLineIndex maps focusIndex to a rendered line index.
func (m *CPUOptionsFormModel) focusedLineIndex() int {
	line := 0
	// Header + blank = 2 lines
	line += 2

	for i, p := range m.positions {
		if i == m.focusIndex {
			return line
		}

		switch p.Kind {
		case form.FocusHeader:
			line += 2 // blank + header
		case form.FocusToggle:
			line++
		case form.FocusText:
			line++
		case form.FocusButton:
			line++
		}
	}

	return line
}

// --- Backward-compatible Init/Update/View ---

// Init implements tea.Model (for backward compatibility).
func (m *CPUOptionsFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model (for backward compatibility).
func (m *CPUOptionsFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		vp, _ := m.vp.Update(msg)
		m.vp = vp
		return m, nil

	case cpuOptionsErrorMsg:
		m.saving = false
		m.statusMessage = msg.err
		return m, nil
	}
	return m, nil
}

// View implements tea.Model (for backward compatibility).
func (m *CPUOptionsFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	m.renderedLines = m.renderAllLines()
	totalContent := ""
	for i, line := range m.renderedLines {
		if i > 0 {
			totalContent += "\n"
		}
		totalContent += line
	}
	m.vp.SetContent(totalContent)
	return m.vp.View()
}

// getToggleValue returns the boolean value for a toggle field (uses reflection).
func (m *CPUOptionsFormModel) getToggleValue(fieldName string) bool {
	return m.getBoolField(fieldName)
}

// toggleValue toggles a boolean field (uses reflection).
func (m *CPUOptionsFormModel) toggleValue(fieldName string) {
	m.setBoolField(fieldName, !m.getBoolField(fieldName))
}

// getTextValue returns the text value for a field (uses reflection).
func (m *CPUOptionsFormModel) getTextValue(fieldName string) string {
	return m.getStringField(fieldName)
}

// setTextValue sets the text value for a field (uses reflection).
func (m *CPUOptionsFormModel) setTextValue(fieldName string, val string) {
	m.setStringField(fieldName, val)
}

// getFieldKind returns the field kind for a given field name from the registry.
func getFieldKind(name string) fields.FieldKind {
	for _, f := range fields.CPUOptionsFields {
		if f.Name == name {
			return f.Kind
		}
	}
	return fields.FieldToggle // default
}