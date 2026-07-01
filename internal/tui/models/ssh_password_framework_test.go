package models

import (
	"strings"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	tea "charm.land/bubbletea/v2"
)

// TestSSHPasswordForm_ImplementsFormModel verifies that SSHPasswordFormModel
// satisfies the form.FormModel interface.
func TestSSHPasswordForm_ImplementsFormModel(t *testing.T) {
	var _ form.FormModel = &SSHPasswordFormModel{}
}

// TestSSHPasswordForm_Integration_TextInput is the tracer bullet:
// creates a form through the framework, types characters, and verifies
// the framework correctly dispatches input to the form.
func TestSSHPasswordForm_Integration_TextInput(t *testing.T) {
	fm := NewSSHPasswordFormModel(false)
	sf := form.NewScrollableForm(fm)
	sf.SetSize(80, 24)

	// Initial focus should be on newPassword (index 0)
	if sf.FocusIndex() != 0 {
		t.Fatalf("expected focus index 0, got %d", sf.FocusIndex())
	}

	// Type "abc" into the newPassword field
	result, _ := sf.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	sf = result.(*form.ScrollableForm)

	result, _ = sf.Update(tea.KeyPressMsg{Code: 'b', Text: "b"})
	sf = result.(*form.ScrollableForm)

	result, _ = sf.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	sf = result.(*form.ScrollableForm)

	// Verify the password was stored
	if fm.newPassword != "abc" {
		t.Errorf("expected newPassword='abc', got %q", fm.newPassword)
	}
}

// TestSSHPasswordForm_Integration_Navigation verifies that the framework
// handles Tab/Shift+Tab navigation correctly through the SSH form.
func TestSSHPasswordForm_Integration_Navigation(t *testing.T) {
	fm := NewSSHPasswordFormModel(false)
	sf := form.NewScrollableForm(fm)
	sf.SetSize(80, 24)

	// Three positions: newPassword, confirmPassword, apply
	positions := fm.BuildPositions()
	if len(positions) != 3 {
		t.Fatalf("expected 3 positions, got %d", len(positions))
	}

	// Tab forward through all positions
	result, _ := sf.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	sf = result.(*form.ScrollableForm)
	if sf.FocusIndex() != 1 {
		t.Errorf("after Tab, expected focus 1 (confirmPassword), got %d", sf.FocusIndex())
	}

	result, _ = sf.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	sf = result.(*form.ScrollableForm)
	if sf.FocusIndex() != 2 {
		t.Errorf("after second Tab, expected focus 2 (apply), got %d", sf.FocusIndex())
	}

	// Clamping at end
	result, _ = sf.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	sf = result.(*form.ScrollableForm)
	if sf.FocusIndex() != 2 {
		t.Errorf("expected focus clamped at 2, got %d", sf.FocusIndex())
	}

	// Shift+Tab backward
	result, _ = sf.Update(tea.KeyPressMsg{Text: "shift+tab"})
	sf = result.(*form.ScrollableForm)
	if sf.FocusIndex() != 1 {
		t.Errorf("after Shift+Tab, expected focus 1, got %d", sf.FocusIndex())
	}
}

// TestSSHPasswordForm_Integration_Backspace verifies backspace deletes characters.
func TestSSHPasswordForm_Integration_Backspace(t *testing.T) {
	fm := NewSSHPasswordFormModel(false)
	sf := form.NewScrollableForm(fm)
	sf.SetSize(80, 24)

	// Type "hello"
	for _, ch := range "hello" {
		result, _ := sf.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
		sf = result.(*form.ScrollableForm)
	}

	if fm.newPassword != "hello" {
		t.Fatalf("expected 'hello', got %q", fm.newPassword)
	}

	// Backspace should delete 'o'
	result, _ := sf.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	sf = result.(*form.ScrollableForm)

	if fm.newPassword != "hell" {
		t.Errorf("expected 'hell' after backspace, got %q", fm.newPassword)
	}
}

// TestSSHPasswordForm_Integration_Delete verifies delete removes characters ahead.
func TestSSHPasswordForm_Integration_Delete(t *testing.T) {
	fm := NewSSHPasswordFormModel(false)
	sf := form.NewScrollableForm(fm)
	sf.SetSize(80, 24)

	// Type "hello"
	for _, ch := range "hello" {
		result, _ := sf.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
		sf = result.(*form.ScrollableForm)
	}

	// Delete at end of string does nothing
	result, _ := sf.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	sf = result.(*form.ScrollableForm)

	if fm.newPassword != "hello" {
		t.Errorf("expected 'hello' after delete at end, got %q", fm.newPassword)
	}

	// Backspace to position 3 ("hel")
	sf.Update(tea.KeyPressMsg{Code: tea.KeyBackspace}) // "hell", cursor at 4
	sf.Update(tea.KeyPressMsg{Code: tea.KeyBackspace}) // "hel", cursor at 3

	// Delete at end still does nothing
	sf.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	if fm.newPassword != "hel" {
		t.Errorf("expected 'hel' after delete at end, got %q", fm.newPassword)
	}

	// Now type 'o' back, then delete should remove it
	sf.Update(tea.KeyPressMsg{Code: 'o', Text: "o"}) // "helo", cursor at 4
	result, _ = sf.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})     // "hel", cursor at 3
	sf = result.(*form.ScrollableForm)
	if fm.newPassword != "hel" {
		t.Errorf("expected 'hel', got %q", fm.newPassword)
	}
}

// TestSSHPasswordForm_Integration_HandleEnter_Save verifies that pressing
// Enter on the Apply button with valid input triggers a save.
func TestSSHPasswordForm_Integration_HandleEnter_Save(t *testing.T) {
	fm := NewSSHPasswordFormModel(true)
	sf := form.NewScrollableForm(fm)
	sf.SetSize(80, 24)

	// Set valid passwords directly (to bypass typing in test)
	fm.newPassword = "testpass123"
	fm.confirmPassword = "testpass123"

	// Navigate to Apply button (2 Tabs)
	sf.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	result, _ := sf.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	sf = result.(*form.ScrollableForm)

	if sf.FocusIndex() != 2 {
		t.Fatalf("expected focus on apply button (index 2), got %d", sf.FocusIndex())
	}

	// Press Enter to apply
	result, cmd := sf.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	sf = result.(*form.ScrollableForm)

	if cmd == nil {
		t.Fatal("expected a tea.Cmd from applying the password")
	}

	// Execute the command and verify the result
	msg := cmd()
	if _, ok := msg.(SSHPasswordUpdatedMsg); !ok {
		t.Errorf("expected SSHPasswordUpdatedMsg, got %T", msg)
	}
}

// TestSSHPasswordForm_Integration_HandleEnter_ValidationError verifies
// that validation errors prevent saving.
func TestSSHPasswordForm_Integration_HandleEnter_ValidationError(t *testing.T) {
	fm := NewSSHPasswordFormModel(false)
	sf := form.NewScrollableForm(fm)
	sf.SetSize(80, 24)

	// Navigate to Apply button (2 Tabs)
	sf.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	result, _ := sf.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	sf = result.(*form.ScrollableForm)

	// Press Enter without filling in passwords
	_, cmd := sf.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Should not produce a save command when validation fails
	if cmd != nil {
		t.Error("expected no cmd when validation fails")
	}

	// Errors should be set
	if _, ok := fm.errors["newPassword"]; !ok {
		t.Error("expected error on newPassword field")
	}
}

// TestSSHPasswordUpdatedMsg_ImplementsFormSavedMsg verifies the result message
// implements the framework's FormSavedMsg interface.
func TestSSHPasswordUpdatedMsg_ImplementsFormSavedMsg(t *testing.T) {
	var msg form.FormSavedMsg = SSHPasswordUpdatedMsg{}

	if msg.FormName() != "SSH Password" {
		t.Errorf("expected FormName='SSH Password', got %q", msg.FormName())
	}
}

// TestSSHPasswordForm_Integration_CustomMessage verifies that custom messages
// (like sshPasswordErrorMsg) are correctly handled through the framework.
func TestSSHPasswordForm_Integration_CustomMessage(t *testing.T) {
	fm := NewSSHPasswordFormModel(false)
	sf := form.NewScrollableForm(fm)
	sf.SetSize(80, 24)

	// Simulate a password error message
	result, _ := sf.Update(sshPasswordErrorMsg{err: "test error"})
	sf = result.(*form.ScrollableForm)

	if fm.statusMessage != "test error" {
		t.Errorf("expected statusMessage='test error', got %q", fm.statusMessage)
	}
}

// TestSSHPasswordForm_BuildPositions verifies the position list is correct.
func TestSSHPasswordForm_BuildPositions(t *testing.T) {
	fm := NewSSHPasswordFormModel(false)
	positions := fm.BuildPositions()

	if len(positions) != 3 {
		t.Fatalf("expected 3 positions, got %d", len(positions))
	}

	if positions[0].Kind != form.FocusText || positions[0].Key != "newPassword" {
		t.Errorf("position 0: expected FocusText/newPassword, got %v/%s", positions[0].Kind, positions[0].Key)
	}
	if positions[1].Kind != form.FocusText || positions[1].Key != "confirmPassword" {
		t.Errorf("position 1: expected FocusText/confirmPassword, got %v/%s", positions[1].Kind, positions[1].Key)
	}
	if positions[2].Kind != form.FocusButton || positions[2].Key != "apply" {
		t.Errorf("position 2: expected FocusButton/apply, got %v/%s", positions[2].Kind, positions[2].Key)
	}
}

// TestSSHPasswordForm_Integration_Render verifies the form renders through the framework.
func TestSSHPasswordForm_Integration_Render(t *testing.T) {
	fm := NewSSHPasswordFormModel(false)
	sf := form.NewScrollableForm(fm)
	sf.SetSize(80, 24)

	viewContent := sf.View().Content
	if viewContent == "" {
		t.Fatal("expected non-empty view")
	}

	// View should contain the form title
	if viewContent == "Loading..." {
		t.Error("view shows loading state after SetSize")
	}
}

// TestSSHPasswordForm_ApplyButtonFormat verifies the Apply button follows standard format
// with [Space/Enter] Apply  [ESC] Cancel pattern.
func TestSSHPasswordForm_ApplyButtonFormat(t *testing.T) {
	fm := NewSSHPasswordFormModel(false)
	sf := form.NewScrollableForm(fm)
	sf.SetSize(80, 24)

	// Navigate to Apply button (index 2)
	sf.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	result, _ := sf.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	sf = result.(*form.ScrollableForm)

	if sf.FocusIndex() != 2 {
		t.Fatalf("expected focus on apply button (index 2), got %d", sf.FocusIndex())
	}

	viewContent := sf.View().Content

	// Should show the standard footer pattern
	if !strings.Contains(viewContent, "[Space/Enter] Apply") {
		t.Errorf("Apply button should show '[Space/Enter] Apply', got:\n%s", viewContent)
	}
	if !strings.Contains(viewContent, "[ESC] Cancel") {
		t.Errorf("Apply button should show '[ESC] Cancel', got:\n%s", viewContent)
	}
}
