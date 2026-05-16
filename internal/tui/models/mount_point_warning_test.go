package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMountPointWarningModelEnterKey(t *testing.T) {
	m := NewMountPointWarningModel()

	// Test Enter key - using Type: tea.KeyEnter
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	model, cmd := m.Update(keyMsg)

	if _, ok := model.(*MountPointWarningModel); !ok {
		t.Error("Expected model to remain MountPointWarningModel")
	}

	// Execute the command to get the message
	msg := cmd()
	if vcm, ok := msg.(ViewChangeMsg); ok {
		if vcm.View != ViewMainMenu {
			t.Errorf("Expected ViewMainMenu, got %s", vcm.View)
		}
	} else {
		t.Errorf("Expected ViewChangeMsg, got %T", msg)
	}
}

func TestMountPointWarningModelEscKey(t *testing.T) {
	m := NewMountPointWarningModel()

	keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
	model, cmd := m.Update(keyMsg)

	if _, ok := model.(*MountPointWarningModel); !ok {
		t.Error("Expected model to remain MountPointWarningModel")
	}

	msg := cmd()
	if vcm, ok := msg.(ViewChangeMsg); ok {
		if vcm.View != ViewMainMenu {
			t.Errorf("Expected ViewMainMenu, got %s", vcm.View)
		}
	} else {
		t.Errorf("Expected ViewChangeMsg, got %T", msg)
	}
}

func TestMountPointWarningModelSpaceKey(t *testing.T) {
	m := NewMountPointWarningModel()

	keyMsg := tea.KeyMsg{Type: tea.KeySpace}
	model, cmd := m.Update(keyMsg)

	if _, ok := model.(*MountPointWarningModel); !ok {
		t.Error("Expected model to remain MountPointWarningModel")
	}

	msg := cmd()
	if vcm, ok := msg.(ViewChangeMsg); ok {
		if vcm.View != ViewMainMenu {
			t.Errorf("Expected ViewMainMenu, got %s", vcm.View)
		}
	} else {
		t.Errorf("Expected ViewChangeMsg, got %T", msg)
	}
}

func TestMountPointWarningModelKeyString(t *testing.T) {
	// Test what msg.String() returns for various key types
	testCases := []struct {
		keyType tea.KeyType
	}{
		{tea.KeyEnter},
		{tea.KeyEsc},
		{tea.KeySpace},
	}

	for _, tc := range testCases {
		keyMsg := tea.KeyMsg{Type: tc.keyType}
		t.Logf("KeyType %v: String() = %q", tc.keyType, keyMsg.String())
	}
}

func TestMountPointWarningModelKeyStringWithRunes(t *testing.T) {
	// Test with KeyRunes - this is what some terminals send for Enter
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\n'}}
	t.Logf("KeyRunes with \\n: String() = %q", keyMsg.String())

	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\r'}}
	t.Logf("KeyRunes with \\r: String() = %q", keyMsg.String())
}