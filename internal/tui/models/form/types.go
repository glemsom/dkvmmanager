// Package form provides a reusable scrollable form framework for BubbleTea TUI applications.
//
// Forms implement the FormModel interface to integrate with the ScrollableForm framework,
// which handles common concerns: viewport management, focus navigation, key dispatching,
// and cursor tracking.
package form

import tea "github.com/charmbracelet/bubbletea"

// FocusKind defines the type of focusable element in a form.
type FocusKind int

const (
	// FocusText represents an editable text field.
	FocusText FocusKind = iota
	// FocusToggle represents a boolean toggle (on/off, yes/no).
	FocusToggle
	// FocusList represents a selectable list item (e.g., file path in a list).
	FocusList
	// FocusButton represents an action button (save, add, cancel).
	FocusButton
	// FocusHeader represents a non-interactive header/label (display only, focusable for navigation).
	FocusHeader
	// FocusCustom represents a custom rendering handled by the form itself.
	FocusCustom
)

// FocusPos represents one navigable position in a form.
type FocusPos struct {
	// Kind defines the type of element at this position.
	Kind FocusKind
	// Label is a human-readable label for display.
	Label string
	// Key is a unique identifier for cursor tracking and error mapping.
	Key string
	// Data is form-specific context (e.g., list index, field metadata).
	Data any
}

// FormResult indicates what action the form wants to take after user interaction.
type FormResult int

const (
	// ResultNone means no special action; the form remains active.
	ResultNone FormResult = iota
	// ResultSave means the form data should be saved and the form closed.
	ResultSave
	// ResultCancel means the form should be cancelled and closed.
	ResultCancel
	// ResultCustom means the form is requesting a custom action via a tea.Cmd.
	ResultCustom
)

// FormModel defines the contract for forms using the scrollable form framework.
//
// Each form implementation provides domain-specific behavior (fields, validation,
// rendering) while the ScrollableForm framework handles navigation, viewport sync,
// and key dispatching.
type FormModel interface {
	// Position Management

	// BuildPositions returns the current list of navigable positions.
	// Called after any structural change (e.g., adding/removing list items).
	BuildPositions() []FocusPos

	// CurrentIndex returns the index of the currently focused position.
	CurrentIndex() int

	// SetFocusIndex sets the focused position index.
	SetFocusIndex(int)

	// Rendering

	// RenderHeader returns the form header markup (title, description, etc.).
	RenderHeader() string

	// RenderPosition returns the markup for a single position.
	// The focused parameter indicates whether this position currently has focus.
	// cursorOffset is the character offset within the text field (for text cursor display).
	RenderPosition(pos FocusPos, focused bool, cursorOffset int) string

	// RenderFooter returns the form footer markup (help text, status, etc.).
	RenderFooter() string

	// Interaction

	// HandleEnter is called when the user presses Enter on a position.
	// Returns the desired result and an optional tea.Cmd.
	HandleEnter(pos FocusPos) (FormResult, tea.Cmd)

	// HandleChar is called when the user types a character into a text field.
	HandleChar(pos FocusPos, ch string)

	// HandleBackspace is called when the user presses Backspace.
	HandleBackspace(pos FocusPos)

	// HandleDelete is called when the user presses Delete.
	HandleDelete(pos FocusPos)

	// Lifecycle

	// OnEnter is called when the form becomes active.
	OnEnter()

	// OnExit is called when the form is dismissed.
	OnExit()

	// SetSize is called when the form dimensions change.
	SetSize(width, height int)

	// SetFocused sets whether the form has keyboard focus.
	SetFocused(bool)
}
