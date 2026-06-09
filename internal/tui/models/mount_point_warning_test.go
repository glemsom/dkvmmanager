package models

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestMountPointWarningModelEnterKey(t *testing.T) {
	m := NewMountPointWarningModel()

	// Test Enter key - using Type: tea.KeyEnter
	keyMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
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

	keyMsg := tea.KeyPressMsg{Code: tea.KeyEsc}
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

	keyMsg := tea.KeyPressMsg{Code: tea.KeySpace}
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
		keyCode rune
	}{
		{tea.KeyEnter},
		{tea.KeyEsc},
		{tea.KeySpace},
	}

	for _, tc := range testCases {
		keyMsg := tea.KeyPressMsg{Code: tc.keyCode}
		t.Logf("KeyCode %v: String() = %q", tc.keyCode, keyMsg.String())
	}
}

func TestMountPointWarningModelKeyStringWithRunes(t *testing.T) {
	// Test with printable runes - this is what some terminals send for Enter
	keyMsg := tea.KeyPressMsg{Code: '\n', Text: "\n"}
	t.Logf("KeyPressMsg with \\n: String() = %q", keyMsg.String())

	keyMsg = tea.KeyPressMsg{Code: '\r', Text: "\r"}
	t.Logf("KeyPressMsg with \\r: String() = %q", keyMsg.String())
}

func TestMountPointWarningModelView(t *testing.T) {
	m := NewMountPointWarningModel()

	// Test View without SetSize - should still render
	viewContent := m.View().Content
	if viewContent == "" {
		t.Error("View() returned empty string")
	}
	if len(viewContent) < 10 {
		t.Error("View() returned unexpectedly short string")
	}
}

func TestMountPointWarningModelSetSize(t *testing.T) {
	m := NewMountPointWarningModel()

	// Test SetSize stores dimensions
	m.SetSize(100, 30)

	// View should still render
	viewContent := m.View().Content
	if viewContent == "" {
		t.Error("View() returned empty string after SetSize")
	}
}

func TestMountPointWarningModelViewWithSize(t *testing.T) {
	m := NewMountPointWarningModel()
	m.SetSize(80, 24)

	viewContent := m.View().Content
	if viewContent == "" {
		t.Error("View() returned empty string after SetSize")
	}

	// The view should contain the warning text
	if !strings.Contains(viewContent, "Mount Point Warning") {
		t.Error("View() missing 'Mount Point Warning' title")
	}
	if !strings.Contains(viewContent, "dkvmdata") {
		t.Error("View() missing 'dkvmdata' text")
	}
}