// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// StartStopScriptFormModel is a form for editing start/stop scripts.
// It implements the form.FormModel interface for use with ScrollableForm.
type StartStopScriptFormModel struct {
	vmManager *vm.Manager
	config    models.StartStopScript
	pciConfig models.PCIPassthroughConfig // For displaying in builtin mode

	// Focus state
	positions  []form.FocusPos
	focusIndex int

	// File browser state (for script path selection via FileBrowserModel)
	fileBrowser *FileBrowserModel

	// Per-field text cursor (character offset within the value)
	cursorOffsets map[string]int

	// Per-field inline error messages
	errors map[string]string

	// Viewport fields (for backward-compatible View)
	vp       viewport.Model
	ready    bool
	contentW int
	contentH int

	// Rendering cache (for backward-compatible View)
	renderedLines []string
}

// NewStartStopScriptFormModel creates a new StartStopScript form model
func NewStartStopScriptFormModel(vmManager *vm.Manager) (*StartStopScriptFormModel, error) {
	cfg, _ := vmManager.GetStartStopScript()
	pciCfg, _ := vmManager.GetPCIPassthroughConfig()

	m := &StartStopScriptFormModel{
		vmManager:     vmManager,
		config:        cfg,
		pciConfig:     pciCfg,
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	m.positions = m.BuildPositions()
	return m, nil
}

// --- FormModel Interface Implementation ---

// BuildPositions returns the list of navigable positions based on current config.
func (m *StartStopScriptFormModel) BuildPositions() []form.FocusPos {
	positions := []form.FocusPos{
		{Kind: form.FocusToggle, Label: "Mode", Key: "toggle"},
	}

	// Add path fields only in custom mode
	if !m.config.UseBuiltin {
		positions = append(positions, form.FocusPos{Kind: form.FocusText, Label: "Start Script", Key: "start_path"})
		positions = append(positions, form.FocusPos{Kind: form.FocusButton, Label: "Browse", Key: "start_browse"})
		positions = append(positions, form.FocusPos{Kind: form.FocusText, Label: "Stop Script", Key: "stop_path"})
		positions = append(positions, form.FocusPos{Kind: form.FocusButton, Label: "Browse", Key: "stop_browse"})
	}

	positions = append(positions, form.FocusPos{Kind: form.FocusButton, Label: "Save", Key: "save"})
	positions = append(positions, form.FocusPos{Kind: form.FocusButton, Label: "Cancel", Key: "cancel"})

	return positions
}

// CurrentIndex returns the index of the currently focused position.
func (m *StartStopScriptFormModel) CurrentIndex() int {
	return m.focusIndex
}

// SetFocusIndex sets the focused position index.
func (m *StartStopScriptFormModel) SetFocusIndex(i int) {
	m.focusIndex = i
}

// RenderHeader returns the form header markup.
func (m *StartStopScriptFormModel) RenderHeader() string {
	return startStopScriptSectionStyle.Render("Custom Start/Stop Script")
}

// RenderFooter returns the form footer markup (help text).
func (m *StartStopScriptFormModel) RenderFooter() string {
	return startStopScriptMutedStyle.Render("Tab Navigate  Space/Enter Select  ESC Cancel")
}

// RenderPosition returns the markup for a single position.
func (m *StartStopScriptFormModel) RenderPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
	switch pos.Kind {
	case form.FocusToggle:
		return m.renderTogglePosition(focused)
	case form.FocusText:
		return m.renderTextPosition(pos, focused, cursorOffset)
	case form.FocusButton:
		return m.renderButtonPosition(pos, focused)
	default:
		return ""
	}
}

// HandleEnter is called when the user presses Enter on a position.
func (m *StartStopScriptFormModel) HandleEnter(pos form.FocusPos) (form.FormResult, tea.Cmd) {
	switch pos.Kind {
	case form.FocusToggle:
		m.config.UseBuiltin = !m.config.UseBuiltin
		m.positions = m.BuildPositions()
		// Clamp focusIndex to valid range
		if m.focusIndex >= len(m.positions) {
			m.focusIndex = 0
		}
		return form.ResultNone, nil

	case form.FocusText:
		// Text field: enter doesn't trigger action, just stays focused
		return form.ResultNone, nil

	case form.FocusButton:
		if pos.Key == "start_browse" || pos.Key == "stop_browse" {
			// Open file browser
			m.fileBrowser = NewFileBrowserModel(FileTypeAll)
			m.fileBrowser.SetDirectory("/media/dkvmdata")
			return form.ResultNone, m.fileBrowser.Init()
		}
		if pos.Key == "save" {
			// Save the configuration
			if err := m.vmManager.SaveStartStopScript(m.config); err != nil {
				m.errors["save"] = err.Error()
			}
			return form.ResultSave, nil
		}
		if pos.Key == "cancel" {
			return form.ResultCancel, nil
		}
	}
	return form.ResultNone, nil
}

// HandleChar is called when the user types a character into a text field.
func (m *StartStopScriptFormModel) HandleChar(pos form.FocusPos, ch string) {
	if pos.Kind != form.FocusText {
		return
	}

	val := m.getScriptPath(pos.Key)
	cur := m.effectiveCursor(pos.Key, val)
	newVal := val[:cur] + ch + val[cur:]

	switch pos.Key {
	case "start_path":
		m.config.StartScript = newVal
	case "stop_path":
		m.config.StopScript = newVal
	}

	m.setCursorOffset(pos.Key, cur+1)
}

// HandleBackspace is called when the user presses Backspace.
func (m *StartStopScriptFormModel) HandleBackspace(pos form.FocusPos) {
	if pos.Kind != form.FocusText {
		return
	}

	val := m.getScriptPath(pos.Key)
	cur := m.effectiveCursor(pos.Key, val)
	if cur > 0 {
		newVal := val[:cur-1] + val[cur:]
		switch pos.Key {
		case "start_path":
			m.config.StartScript = newVal
		case "stop_path":
			m.config.StopScript = newVal
		}
		m.setCursorOffset(pos.Key, cur-1)
	}
}

// HandleDelete is called when the user presses Delete.
func (m *StartStopScriptFormModel) HandleDelete(pos form.FocusPos) {
	if pos.Kind != form.FocusText {
		return
	}

	val := m.getScriptPath(pos.Key)
	cur := m.effectiveCursor(pos.Key, val)
	if cur < len(val) {
		newVal := val[:cur] + val[cur+1:]
		switch pos.Key {
		case "start_path":
			m.config.StartScript = newVal
		case "stop_path":
			m.config.StopScript = newVal
		}
		m.setCursorOffset(pos.Key, cur)
	}
}

// OnEnter is called when the form becomes active.
func (m *StartStopScriptFormModel) OnEnter() {}

// OnExit is called when the form is dismissed.
func (m *StartStopScriptFormModel) OnExit() {}

// SetSize updates the form dimensions.
func (m *StartStopScriptFormModel) SetSize(w, h int) {
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
func (m *StartStopScriptFormModel) SetFocused(bool) {}

// --- Helper methods ---

// getScriptPath returns the script path for a given position key.
func (m *StartStopScriptFormModel) getScriptPath(key string) string {
	switch key {
	case "start_path":
		if m.config.StartScript == "" {
			return "/media/dkvmdata/start.sh"
		}
		return m.config.StartScript
	case "stop_path":
		if m.config.StopScript == "" {
			return "/media/dkvmdata/stop.sh"
		}
		return m.config.StopScript
	}
	return ""
}

// cursorOffset returns the cursor offset for the given position key
func (m *StartStopScriptFormModel) cursorOffset(key string) int {
	if off, ok := m.cursorOffsets[key]; ok {
		return off
	}
	return -1 // sentinel meaning "end"
}

// setCursorOffset sets cursor offset; -1 means end
func (m *StartStopScriptFormModel) setCursorOffset(key string, off int) {
	m.cursorOffsets[key] = off
}

// effectiveCursor returns the actual cursor position (0-based character index)
func (m *StartStopScriptFormModel) effectiveCursor(key string, val string) int {
	off := m.cursorOffset(key)
	if off < 0 {
		return len(val)
	}
	if off > len(val) {
		return len(val)
	}
	return off
}

// --- Backward-compatible Init/Update/View ---

// Init implements tea.Model (for backward compatibility).
func (m *StartStopScriptFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model (for backward compatibility).
func (m *StartStopScriptFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Delegate to active file browser first
	if m.fileBrowser != nil && m.fileBrowser.active {
		inner, cmd := m.fileBrowser.Update(msg)
		if fb, ok := inner.(*FileBrowserModel); ok {
			m.fileBrowser = fb
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			// Enter or Space toggles the mode when focused on toggle
			// or opens file browser when focused on browse buttons
			return m.handleEnter()
		case "left":
			// Left selects "Use Builtin Script" when focused on toggle
			m.handleArrowKey(false)
		case "right":
			// Right selects "Use Custom Script" when focused on toggle
			m.handleArrowKey(true)
		case "tab":
			m.focusIndex++
			if m.focusIndex >= len(m.positions) {
				m.focusIndex = 0
			}
		case "shift+tab":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.positions) - 1
			}
		case "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.positions) - 1
			}
		case "down":
			m.focusIndex++
			if m.focusIndex >= len(m.positions) {
				m.focusIndex = 0
			}
		}
		m.syncViewport()
	case DirectoryLoadedMsg:
		// Directory loaded in file browser - trigger view refresh
		m.syncViewport()
	case FileSelectedMsg:
		// Handle file selection from browser
		m.handleFileSelected(msg)
	}
	return m, nil
}

// View implements tea.Model (for backward compatibility).
func (m *StartStopScriptFormModel) View() string {
	// If file browser is active, show it instead of the form
	if m.fileBrowser != nil && m.fileBrowser.active {
		return m.fileBrowser.View()
	}

	if !m.ready {
		return "Loading form..."
	}
	// Ensure viewport content is rendered
	m.syncViewport()
	return m.vp.View()
}

// handleEnter handles Enter key in the form (backward compat).
func (m *StartStopScriptFormModel) handleEnter() (tea.Model, tea.Cmd) {
	if m.focusIndex < len(m.positions) {
		pos := m.positions[m.focusIndex]
		switch pos.Kind {
		case form.FocusToggle:
			m.config.UseBuiltin = !m.config.UseBuiltin
			m.positions = m.BuildPositions()

		case form.FocusButton:
			if pos.Key == "start_browse" {
				// Open file browser for start script
				m.fileBrowser = NewFileBrowserModel(FileTypeAll)
				m.fileBrowser.SetDirectory("/media/dkvmdata")
				return m, m.fileBrowser.Init()
			}
			if pos.Key == "stop_browse" {
				// Open file browser for stop script
				m.fileBrowser = NewFileBrowserModel(FileTypeAll)
				m.fileBrowser.SetDirectory("/media/dkvmdata")
				return m, m.fileBrowser.Init()
			}
		}
	}
	return m, nil
}

// handleFileSelected processes the result from the file browser (backward compat).
func (m *StartStopScriptFormModel) handleFileSelected(msg FileSelectedMsg) {
	m.fileBrowser = nil
	if !msg.Canceled && m.focusIndex < len(m.positions) {
		pos := m.positions[m.focusIndex]
		switch pos.Kind {
		case form.FocusButton:
			if pos.Key == "start_browse" {
				m.config.StartScript = msg.Path
			}
			if pos.Key == "stop_browse" {
				m.config.StopScript = msg.Path
			}
		}
	}
	m.syncViewport()
}

// handleArrowKey handles arrow keys when focused on the toggle (backward compat).
func (m *StartStopScriptFormModel) handleArrowKey(moveRight bool) {
	if m.focusIndex < len(m.positions) {
		pos := m.positions[m.focusIndex]
		if pos.Kind == form.FocusToggle {
			if moveRight {
				m.config.UseBuiltin = false
			} else {
				m.config.UseBuiltin = true
			}
			m.positions = m.BuildPositions()
		}
	}
}

// rebuildPositions rebuilds the list of focusable positions based on current config (backward compat).
func (m *StartStopScriptFormModel) rebuildPositions() {
	m.positions = m.BuildPositions()

	// Ensure focusIndex is valid
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = 0
	}
}

// currentPos returns the focus position at the current focusIndex (backward compat).
func (m *StartStopScriptFormModel) currentPos() form.FocusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return form.FocusPos{}
	}
	return m.positions[m.focusIndex]
}

// syncViewport updates the viewport content (backward compat).
func (m *StartStopScriptFormModel) syncViewport() {
	lines := m.renderAllLines()
	m.renderedLines = lines

	content := ""
	for i, line := range lines {
		if i > 0 {
			content += "\n"
		}
		content += line
	}

	m.vp.SetContent(content)
}
