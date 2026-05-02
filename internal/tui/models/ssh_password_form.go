// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// Style variables for the SSH password form.
var (
	sshPwLabelStyle   = lipgloss.NewStyle().Foreground(styles.Colors.ForegroundDim)
	sshPwFocusStyle   = styles.FormFocusStyle()
	sshPwInputStyle   = styles.FormInputStyle()
	sshPwErrorStyle   = styles.ErrorTextStyle()
	sshPwMutedStyle   = styles.FormMutedStyle()
	sshPwSaveStyle    = styles.FormSaveStyle()
	sshPwStrengthStyle = styles.FormInputStyle()
)

// SSHPasswordFormModel is a scrollable form for setting the SSH password.
// It implements the form.FormModel interface for use with ScrollableForm.
type SSHPasswordFormModel struct {
	newPassword     string
	confirmPassword string

	// Focus state
	positions  []form.FocusPos
	focusIndex int

	// Per-field text cursor (character offset within the value)
	cursorOffsets map[string]int

	// Per-field inline error messages
	errors map[string]string

	// Status message (for apply errors)
	statusMessage string

	// Whether an apply operation is in progress
	applying bool

	// Size (for viewport sync, used by framework's SetSize)
	contentW int
	contentH int
	vp       viewport.Model
	ready    bool

	// Rendering cache
	renderedLines []string
}

// NewSSHPasswordFormModel creates a new SSH password form model.
func NewSSHPasswordFormModel() *SSHPasswordFormModel {
	m := &SSHPasswordFormModel{
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	m.positions = m.buildPositions()
	return m
}

// --- FormModel Interface Implementation ---

// BuildPositions returns the current list of navigable positions.
func (m *SSHPasswordFormModel) BuildPositions() []form.FocusPos {
	return m.buildPositions()
}

func (m *SSHPasswordFormModel) buildPositions() []form.FocusPos {
	return []form.FocusPos{
		{Kind: form.FocusText, Label: "New Password", Key: "newPassword"},
		{Kind: form.FocusText, Label: "Confirm Password", Key: "confirmPassword"},
		{Kind: form.FocusButton, Label: "Apply", Key: "apply"},
	}
}

// CurrentIndex returns the index of the currently focused position.
func (m *SSHPasswordFormModel) CurrentIndex() int {
	return m.focusIndex
}

// SetFocusIndex sets the focused position index.
func (m *SSHPasswordFormModel) SetFocusIndex(i int) {
	m.focusIndex = i
}

// RenderHeader returns the form header.
func (m *SSHPasswordFormModel) RenderHeader() string {
	return sshPwFocusStyle.Render("Set SSH Password")
}

// RenderPosition returns the markup for a single position.
func (m *SSHPasswordFormModel) RenderPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
	switch pos.Key {
	case "newPassword", "confirmPassword":
		val := m.getTextValue(pos.Key)
		cursor := m.effectiveCursor(pos.Key, val)
		if focused && cursorOffset >= 0 {
			cursor = cursorOffset
		}
		return m.renderPasswordInput(pos.Label, val, cursor, focused)
	case "apply":
		// Strength indicator + button
		strength := passwordStrength(m.newPassword)
		strengthText, strengthColor := strengthLabel(strength)
		barWidth := 10
		filled := int(float64(strength) / 5.0 * float64(barWidth))
		bar := sshPwStrengthStyle.Foreground(strengthColor).Render(strings.Repeat("█", filled)) +
			sshPwMutedStyle.Render(strings.Repeat("░", barWidth-filled))
		lines := []string{
			sshPwLabelStyle.Render("Strength: ") + bar + " " + lipgloss.NewStyle().Foreground(strengthColor).Render(strengthText),
			"",
		}
		applyText := sshPwMutedStyle.Render("[Apply]")
		if focused {
			applyText = sshPwSaveStyle.Render("[Apply]")
		}
		lines = append(lines, applyText)
		return strings.Join(lines, "\n")
	}
	return ""
}

// RenderFooter returns the form footer.
func (m *SSHPasswordFormModel) RenderFooter() string {
	var parts []string
	if m.statusMessage != "" {
		parts = append(parts, "")
		parts = append(parts, sshPwErrorStyle.Render(m.statusMessage))
	}
	parts = append(parts, "")
	parts = append(parts, sshPwMutedStyle.Render("Tab/Shift+Tab Navigate  Space/Enter Select  ESC Cancel"))
	return strings.Join(parts, "\n")
}

// HandleEnter is called when the user presses Enter on a position.
func (m *SSHPasswordFormModel) HandleEnter(pos form.FocusPos) (form.FormResult, tea.Cmd) {
	switch pos.Key {
	case "apply":
		if !m.validate() {
			return form.ResultNone, nil
		}
		m.applying = true
		m.statusMessage = ""
		return form.ResultNone, m.apply()
	default:
		// Move to next field
		m.SetFocusIndex(m.focusIndex + 1)
		if m.focusIndex >= len(m.positions) {
			m.focusIndex = len(m.positions) - 1
		}
		return form.ResultNone, nil
	}
}

// HandleChar inserts a character at the cursor in the focused text field.
func (m *SSHPasswordFormModel) HandleChar(pos form.FocusPos, ch string) {
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
func (m *SSHPasswordFormModel) HandleBackspace(pos form.FocusPos) {
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
func (m *SSHPasswordFormModel) HandleDelete(pos form.FocusPos) {
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

// OnEnter is called when the form becomes active.
func (m *SSHPasswordFormModel) OnEnter() {}

// OnExit is called when the form is dismissed.
func (m *SSHPasswordFormModel) OnExit() {}

// SetSize updates the form dimensions.
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
}

// SetFocused sets whether the form has keyboard focus.
func (m *SSHPasswordFormModel) SetFocused(bool) {}

// HandleMessage handles custom messages (e.g., async command results).
func (m *SSHPasswordFormModel) HandleMessage(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case sshPasswordErrorMsg:
		m.applying = false
		m.statusMessage = msg.err
		return nil
	}
	return nil
}

// --- Internal helpers ---

// getTextValue returns the text value for a field.
func (m *SSHPasswordFormModel) getTextValue(fieldName string) string {
	switch fieldName {
	case "newPassword":
		return m.newPassword
	case "confirmPassword":
		return m.confirmPassword
	}
	return ""
}

// setTextValue sets the text value for a field.
func (m *SSHPasswordFormModel) setTextValue(fieldName string, val string) {
	switch fieldName {
	case "newPassword":
		m.newPassword = val
	case "confirmPassword":
		m.confirmPassword = val
	}
}

// effectiveCursor returns the actual cursor position (0-based character index).
// -1 means cursor at end (default).
func (m *SSHPasswordFormModel) effectiveCursor(key string, val string) int {
	off, ok := m.cursorOffsets[key]
	if !ok || off < 0 {
		return len(val)
	}
	if off > len(val) {
		return len(val)
	}
	return off
}

// cursorOffset returns the cursor offset for the given field name.
func (m *SSHPasswordFormModel) cursorOffset(key string) int {
	if off, ok := m.cursorOffsets[key]; ok {
		return off
	}
	return -1
}

// setCursorOffset sets cursor offset; -1 means end.
func (m *SSHPasswordFormModel) setCursorOffset(key string, off int) {
	m.cursorOffsets[key] = off
}
