package form

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// ScrollableForm is the framework's core model. It owns a viewport and delegates
// rendering and interaction to a FormModel implementation.
//
// ScrollableForm handles:
//   - Viewport initialization and synchronization
//   - Focus navigation (tab, shift+tab, up, down)
//   - Key dispatching (char input, backspace, delete, enter)
//   - Cursor position tracking per field
type ScrollableForm struct {
	model FormModel

	// Viewport state
	vp    viewport.Model
	ready bool

	// Content dimensions
	contentW int
	contentH int

	// Focus state (mirrored from model for framework-level access)
	focusIndex int

	// Per-field text cursor offsets (character position within field value)
	cursorOffsets map[string]int

	// Whether the form has focus
	focused bool
}

// NewScrollableForm creates a new ScrollableForm wrapping the given FormModel.
func NewScrollableForm(model FormModel) *ScrollableForm {
	return &ScrollableForm{
		model:         model,
		cursorOffsets: make(map[string]int),
		focused:       true,
		focusIndex:    model.CurrentIndex(),
	}
}

// Init implements tea.Model.
func (sf *ScrollableForm) Init() tea.Cmd {
	sf.model.OnEnter()
	return nil
}

// Ready returns true if the viewport has been initialized.
func (sf *ScrollableForm) Ready() bool {
	return sf.ready
}

// FocusIndex returns the current focus index.
func (sf *ScrollableForm) FocusIndex() int {
	return sf.focusIndex
}

// Model returns the underlying FormModel (for testing/internal access).
func (sf *ScrollableForm) Model() FormModel {
	return sf.model
}

// SetFocusIndex sets the focus index directly.
func (sf *ScrollableForm) SetFocusIndex(i int) {
	sf.focusIndex = i
	sf.model.SetFocusIndex(i)
}

// MoveFocus moves focus by delta (positive = down, negative = up).
func (sf *ScrollableForm) MoveFocus(delta int) {
	positions := sf.model.BuildPositions()
	sf.focusIndex = moveFocus(positions, sf.focusIndex, delta)
	sf.model.SetFocusIndex(sf.focusIndex)
}

// SetFocused sets whether the form has keyboard focus.
func (sf *ScrollableForm) SetFocused(f bool) {
	sf.focused = f
	sf.model.SetFocused(f)
}

// SetSize updates the form dimensions and initializes the viewport if needed.
func (sf *ScrollableForm) SetSize(w, h int) {
	sf.contentW = w
	sf.contentH = h
	if !sf.ready {
		sf.vp = viewport.New(w, h)
		sf.ready = true
	} else {
		sf.vp.Width = w
		sf.vp.Height = h
	}
	sf.model.SetSize(w, h)
	sf.syncViewport()
}

// handleMessage is an optional interface for forms that need to handle
// custom messages (e.g., async command results). If a FormModel implements
// this, ScrollableForm will delegate unknown messages to it.
type handleMessage interface {
	HandleMessage(msg tea.Msg) tea.Cmd
}

// Update implements tea.Model.
func (sf *ScrollableForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sf.contentW = msg.Width
		sf.contentH = msg.Height
		if !sf.ready {
			sf.vp = viewport.New(msg.Width, msg.Height)
			sf.ready = true
		} else {
			sf.vp.Width = msg.Width
			sf.vp.Height = msg.Height
		}
		sf.model.SetSize(msg.Width, msg.Height)
		sf.syncViewport()
		return sf, nil

	case tea.KeyMsg:
		return sf.handleKey(msg)

	default:
		// Delegate custom messages (e.g., async results) to the form
		if hm, ok := sf.model.(handleMessage); ok {
			if cmd := hm.HandleMessage(msg); cmd != nil {
				return sf, cmd
			}
		}
	}
	return sf, nil
}

// handleKey dispatches keyboard input to the appropriate form method.
func (sf *ScrollableForm) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "tab":
		sf.MoveFocus(1)
		sf.syncViewport()
		return sf, nil

	case "shift+tab":
		sf.MoveFocus(-1)
		sf.syncViewport()
		return sf, nil

	case "up":
		sf.MoveFocus(-1)
		sf.syncViewport()
		return sf, nil

	case "down":
		sf.MoveFocus(1)
		sf.syncViewport()
		return sf, nil

	case "enter":
		return sf.handleEnter()

	case "backspace":
		pos := sf.currentPos()
		sf.model.HandleBackspace(pos)
		sf.syncViewport()
		return sf, nil

	case "delete":
		pos := sf.currentPos()
		sf.model.HandleDelete(pos)
		sf.syncViewport()
		return sf, nil

	default:
		if len(key) == 1 {
			pos := sf.currentPos()
			sf.model.HandleChar(pos, key)
			sf.syncViewport()
			return sf, nil
		}
	}
	return sf, nil
}

// handleEnter dispatches Enter key to the form model and handles the result.
func (sf *ScrollableForm) handleEnter() (tea.Model, tea.Cmd) {
	pos := sf.currentPos()
	result, cmd := sf.model.HandleEnter(pos)

	switch result {
	case ResultSave, ResultCancel:
		// These results signal the form wants to exit;
		// the parent model handles the actual dismissal.
		// For the framework, we just return the result.
		return sf, cmd

	case ResultCustom:
		return sf, cmd

	case ResultNone:
		// Stay on the form, sync viewport. Preserve any cmd from HandleEnter.
		sf.syncViewport()
		return sf, cmd

	default:
		sf.syncViewport()
		return sf, cmd
	}
}

// View implements tea.Model.
func (sf *ScrollableForm) View() string {
	if !sf.ready {
		return "Loading..."
	}
	return sf.vp.View()
}

// syncViewport rebuilds the content and syncs the viewport to the focused position.
func (sf *ScrollableForm) syncViewport() {
	positions := sf.model.BuildPositions()

	var content strings.Builder
	content.WriteString(sf.model.RenderHeader())
	content.WriteString("\n")

	for i, pos := range positions {
		focused := i == sf.focusIndex
		cursorOff := sf.cursorOffset(pos.Key)
		content.WriteString(sf.model.RenderPosition(pos, focused, cursorOff))
		content.WriteString("\n")
	}

	content.WriteString(sf.model.RenderFooter())

	// Update viewport content
	sf.vp.SetContent(content.String())

	// Scroll to keep focused line visible
	focusedLine := focusedLineIndex(positions, sf.focusIndex, 1) // +1 for header
	sf.vp.YOffset = clampOffset(sf.vp.YOffset, focusedLine, sf.vp.Height)
}

// cursorOffset returns the cursor offset for a field key.
// -1 means cursor at end (default).
func (sf *ScrollableForm) cursorOffset(key string) int {
	if off, ok := sf.cursorOffsets[key]; ok {
		return off
	}
	return -1
}

// setCursorOffset sets the cursor offset for a field key.
func (sf *ScrollableForm) setCursorOffset(key string, off int) {
	sf.cursorOffsets[key] = off
}

// currentPos returns the FocusPos at the current focus index.
func (sf *ScrollableForm) currentPos() FocusPos {
	positions := sf.model.BuildPositions()
	if sf.focusIndex < 0 || sf.focusIndex >= len(positions) {
		if len(positions) > 0 {
			return positions[0]
		}
		return FocusPos{}
	}
	return positions[sf.focusIndex]
}
