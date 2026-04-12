// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// sshPwFocusKind describes what a focus position represents
type sshPwFocusKind int

const (
	sshPwText sshPwFocusKind = iota
	sshPwApply
)

// sshPwFocusPos is one navigable position in the SSH password form
type sshPwFocusPos struct {
	kind      sshPwFocusKind
	fieldName string
}

// SSHPasswordFormModel is a scrollable form for setting the SSH password
type SSHPasswordFormModel struct {
	newPassword     string
	confirmPassword string

	// Flat list of focusable positions
	positions  []sshPwFocusPos
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

	// Status message (for apply errors)
	statusMessage string

	// Whether an apply operation is in progress
	applying bool
}

// NewSSHPasswordFormModel creates a new SSH password form model
func NewSSHPasswordFormModel() *SSHPasswordFormModel {
	m := &SSHPasswordFormModel{
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	m.rebuildPositions()
	return m
}

// cursorOffset returns the cursor offset for the given field name
func (m *SSHPasswordFormModel) cursorOffset(key string) int {
	if off, ok := m.cursorOffsets[key]; ok {
		return off
	}
	return -1
}

// setCursorOffset sets cursor offset; -1 means end
func (m *SSHPasswordFormModel) setCursorOffset(key string, off int) {
	m.cursorOffsets[key] = off
}

// effectiveCursor returns the actual cursor position (0-based character index)
func (m *SSHPasswordFormModel) effectiveCursor(key string, val string) int {
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
func (m *SSHPasswordFormModel) Init() tea.Cmd {
	return nil
}

// rebuildPositions reconstructs the flat focus list
func (m *SSHPasswordFormModel) rebuildPositions() {
	m.positions = nil
	m.positions = append(m.positions, sshPwFocusPos{kind: sshPwText, fieldName: "newPassword"})
	m.positions = append(m.positions, sshPwFocusPos{kind: sshPwText, fieldName: "confirmPassword"})
	m.positions = append(m.positions, sshPwFocusPos{kind: sshPwApply, fieldName: "apply"})
}

// currentPos returns the focus position at the current focusIndex
func (m *SSHPasswordFormModel) currentPos() sshPwFocusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return sshPwFocusPos{kind: sshPwText, fieldName: "newPassword"}
	}
	return m.positions[m.focusIndex]
}

// getTextValue returns the text value for a field
func (m *SSHPasswordFormModel) getTextValue(fieldName string) string {
	switch fieldName {
	case "newPassword":
		return m.newPassword
	case "confirmPassword":
		return m.confirmPassword
	}
	return ""
}

// setTextValue sets the text value for a field
func (m *SSHPasswordFormModel) setTextValue(fieldName string, val string) {
	switch fieldName {
	case "newPassword":
		m.newPassword = val
	case "confirmPassword":
		m.confirmPassword = val
	}
}

// View implements tea.Model
func (m *SSHPasswordFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// SetSize updates the form dimensions
func (m *SSHPasswordFormModel) SetSize(w, h int) {
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
