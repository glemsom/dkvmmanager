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

// Constants for backward compatibility with existing tests.
// Note: these map to form.FocusKind values used in positions.
const (
	sshPwText  = int(form.FocusText)   // text field position kind
	sshPwApply = int(form.FocusButton) // apply button position kind
)

// sshPasswordErrorMsg is sent when password change fails.
type sshPasswordErrorMsg struct {
	err string
}

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

// renderPasswordInput renders a masked password input field.
func (m *SSHPasswordFormModel) renderPasswordInput(label, value string, cursor int, focused bool) string {
	prefix := "  "
	if focused {
		prefix = sshPwFocusStyle.Render("> ")
	}

	labelPart := sshPwLabelStyle.Render(label + ": ")
	masked := strings.Repeat("*", len(value))

	var valPart string
	if focused {
		if cursor < len(masked) {
			before := masked[:cursor]
			at := string(masked[cursor])
			after := ""
			if cursor+1 < len(masked) {
				after = masked[cursor+1:]
			}
			valPart = sshPwInputStyle.Render(before) +
				lipgloss.NewStyle().Reverse(true).Render(at) +
				sshPwInputStyle.Render(after)
		} else {
			valPart = sshPwInputStyle.Render(masked) + sshPwFocusStyle.Render("_")
		}
	} else {
		if value == "" {
			valPart = sshPwMutedStyle.Render("(empty)")
		} else {
			valPart = sshPwInputStyle.Render(masked)
		}
	}

	return prefix + labelPart + valPart
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

// renderAllLines generates the complete content lines for the viewport.
func (m *SSHPasswordFormModel) renderAllLines() []string {
	var lines []string
	lines = append(lines, sshPwFocusStyle.Render("Set SSH Password"), "")

	for i, pos := range m.positions {
		focused := i == m.focusIndex
		cursor := m.effectiveCursor(pos.Key, m.getTextValue(pos.Key))
		if focused {
			cursor = m.cursorOffset(pos.Key)
			if cursor < 0 {
				cursor = m.effectiveCursor(pos.Key, m.getTextValue(pos.Key))
			}
		}
		lines = append(lines, m.RenderPosition(pos, focused, cursor))

		// Add error line if applicable
		if pos.Kind == form.FocusText {
			if err, hasErr := m.errors[pos.Key]; hasErr {
				lines = append(lines, sshPwErrorStyle.Render(err))
			}
		}
		if pos.Kind == form.FocusButton {
			lines = append(lines, "") // blank separator before button
		}
	}

	// Footer
	lines = append(lines, "")
	if m.statusMessage != "" {
		lines = append(lines, sshPwErrorStyle.Render(m.statusMessage), "")
	}
	lines = append(lines, sshPwMutedStyle.Render("Tab/Shift+Tab Navigate  Space/Enter Select  ESC Cancel"))
	return lines
}

// Init implements tea.Model (for backward compatibility).
func (m *SSHPasswordFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model (for backward compatibility).
func (m *SSHPasswordFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case sshPasswordErrorMsg:
		m.applying = false
		m.statusMessage = msg.err
		return m, nil
	}
	return m, nil
}

// handleKey processes keyboard input for backward-compatible Update.
func (m *SSHPasswordFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.applying {
		return m, nil
	}

	key := msg.String()

	switch key {
	case "tab":
		m.moveFocus(1)
	case "shift+tab":
		m.moveFocus(-1)
	case "up":
		m.moveFocus(-1)
	case "down":
		m.moveFocus(1)
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

// moveFocus moves focus by delta in the flat positions list.
func (m *SSHPasswordFormModel) moveFocus(delta int) {
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// handleEnterKey acts contextually: apply or move to next field.
func (m *SSHPasswordFormModel) handleEnterKey() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	if pos.kind == sshPwApply {
		if !m.validate() {
			return m, nil
		}
		m.applying = true
		m.statusMessage = ""
		return m, m.apply()
	}
	m.moveFocus(1)
	return m, nil
}

// handleBackspaceKey deletes the character before cursor.
func (m *SSHPasswordFormModel) handleBackspaceKey() {
	pos := m.currentPos()
	if pos.kind != sshPwText {
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

// handleDeleteKey deletes the character at cursor.
func (m *SSHPasswordFormModel) handleDeleteKey() {
	pos := m.currentPos()
	if pos.kind != sshPwText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	cursor := m.effectiveCursor(pos.fieldName, val)
	if cursor < len(val) {
		newVal := val[:cursor] + val[cursor+1:]
		m.setTextValue(pos.fieldName, newVal)
	}
}

// handleCharInput inserts a character at cursor.
func (m *SSHPasswordFormModel) handleCharInput(ch string) {
	pos := m.currentPos()
	if pos.kind != sshPwText {
		return
	}
	val := m.getTextValue(pos.fieldName)
	cursor := m.effectiveCursor(pos.fieldName, val)
	newVal := val[:cursor] + ch + val[cursor:]
	m.setTextValue(pos.fieldName, newVal)
	m.setCursorOffset(pos.fieldName, cursor+1)
}

// View implements tea.Model (for backward compatibility).
func (m *SSHPasswordFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	m.renderedLines = m.renderAllLines()
	totalContent := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(totalContent)
	return m.vp.View()
}

// sshPos is a legacy position type for backward-compatible test access.
type sshPos struct {
	kind      int
	fieldName string
}

// currentPos returns the current focused position (for backward compatibility).
func (m *SSHPasswordFormModel) currentPos() sshPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		if len(m.positions) > 0 {
			p := m.positions[0]
			return sshPos{kind: int(p.Kind), fieldName: p.Key}
		}
		return sshPos{}
	}
	p := m.positions[m.focusIndex]
	return sshPos{kind: int(p.Kind), fieldName: p.Key}
}
