package form

import (
	"testing"
)

func TestDefaultKeyBindings_Tab(t *testing.T) {
	kb := DefaultKeyBindings()
	if !kb.isTab("tab") {
		t.Error("expected 'tab' to match Tab binding")
	}
	if kb.isTab("enter") {
		t.Error("expected 'enter' to NOT match Tab binding")
	}
}

func TestDefaultKeyBindings_ShiftTab(t *testing.T) {
	kb := DefaultKeyBindings()
	if !kb.isShiftTab("shift+tab") {
		t.Error("expected 'shift+tab' to match ShiftTab binding")
	}
}

func TestDefaultKeyBindings_Navigation(t *testing.T) {
	kb := DefaultKeyBindings()

	if !kb.isNavUp("up") {
		t.Error("expected 'up' to match Up binding")
	}
	if !kb.isNavDown("down") {
		t.Error("expected 'down' to match Down binding")
	}
	if kb.isNavUp("down") {
		t.Error("expected 'down' to NOT match Up binding")
	}
}

func TestDefaultKeyBindings_Actions(t *testing.T) {
	kb := DefaultKeyBindings()

	if !kb.isEnter("enter") {
		t.Error("expected 'enter' to match Enter binding")
	}
	if !kb.isBackspace("backspace") {
		t.Error("expected 'backspace' to match Backspace binding")
	}
	if !kb.isDelete("delete") {
		t.Error("expected 'delete' to match Delete binding")
	}
	if !kb.isSpace(" ") {
		t.Error("expected ' ' to match Space binding")
	}
}

func TestKeyBindings_CustomBindings(t *testing.T) {
	kb := KeyBindings{
		Tab:      []string{"ctrl+n"},
		ShiftTab: []string{"ctrl+p"},
		Up:       []string{"k"},
		Down:     []string{"j"},
		Enter:    []string{"enter", "return"},
	}

	if !kb.isTab("ctrl+n") {
		t.Error("expected custom tab binding 'ctrl+n' to match")
	}
	if !kb.isNavUp("k") {
		t.Error("expected custom up binding 'k' to match")
	}
	if !kb.isEnter("return") {
		t.Error("expected 'return' to match custom Enter binding")
	}
}
