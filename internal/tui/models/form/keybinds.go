package form

// KeyBindings holds the key sequences for form navigation and actions.
// Forms can customize these by providing their own KeyBindings.
type KeyBindings struct {
	// Navigation
	Tab      []string
	ShiftTab []string
	Up       []string
	Down     []string
	PageUp   []string
	PageDown []string

	// Actions
	Enter     []string
	Space     []string
	Backspace []string
	Delete    []string
}

// DefaultKeyBindings returns the standard key bindings used by ScrollableForm.
func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		// Navigation
		Tab:      []string{"tab"},
		ShiftTab: []string{"shift+tab"},
		Up:       []string{"up"},
		Down:     []string{"down"},
		PageUp:   []string{"pgup"},
		PageDown: []string{"pgdown"},

		// Actions
		Enter:     []string{"enter"},
		Space:     []string{" "},
		Backspace: []string{"backspace"},
		Delete:    []string{"delete"},
	}
}

// matchesKey returns true if the given key string matches any of the bindings.
func (kb KeyBindings) matchesKey(key string, bindings []string) bool {
	for _, b := range bindings {
		if key == b {
			return true
		}
	}
	return false
}

// isNavUp returns true if the key is an upward navigation key.
func (kb KeyBindings) isNavUp(key string) bool {
	return kb.matchesKey(key, kb.Up)
}

// isNavDown returns true if the key is a downward navigation key.
func (kb KeyBindings) isNavDown(key string) bool {
	return kb.matchesKey(key, kb.Down)
}

// isTab returns true if the key is the tab key.
func (kb KeyBindings) isTab(key string) bool {
	return kb.matchesKey(key, kb.Tab)
}

// isShiftTab returns true if the key is shift+tab.
func (kb KeyBindings) isShiftTab(key string) bool {
	return kb.matchesKey(key, kb.ShiftTab)
}

// isEnter returns true if the key is the enter key.
func (kb KeyBindings) isEnter(key string) bool {
	return kb.matchesKey(key, kb.Enter)
}

// isBackspace returns true if the key is the backspace key.
func (kb KeyBindings) isBackspace(key string) bool {
	return kb.matchesKey(key, kb.Backspace)
}

// isDelete returns true if the key is the delete key.
func (kb KeyBindings) isDelete(key string) bool {
	return kb.matchesKey(key, kb.Delete)
}

// isSpace returns true if the key is the space bar.
func (kb KeyBindings) isSpace(key string) bool {
	return kb.matchesKey(key, kb.Space)
}
