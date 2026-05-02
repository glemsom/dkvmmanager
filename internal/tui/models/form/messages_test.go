package form

import "testing"

// TestIsFormSavedMsg_DetectsFormSaved verifies that messages implementing
// FormSavedMsg are correctly detected.
func TestIsFormSavedMsg_DetectsFormSaved(t *testing.T) {
	msg := testSavedMsg{name: "test-form"}
	if !IsFormSavedMsg(msg) {
		t.Error("expected IsFormSavedMsg to return true for FormSavedMsg implementation")
	}
}

// TestIsFormSavedMsg_RejectsNonFormSaved verifies that ordinary messages
// are not falsely detected as form saved messages.
func TestIsFormSavedMsg_RejectsNonFormSaved(t *testing.T) {
	msg := struct{}{}
	if IsFormSavedMsg(msg) {
		t.Error("expected IsFormSavedMsg to return false for non-FormSavedMsg")
	}
}

// TestFormSavedMsg_Name verifies the FormName method returns the form identifier.
func TestFormSavedMsg_Name(t *testing.T) {
	msg := testSavedMsg{name: "cpu-options"}
	if got := FormSavedName(msg); got != "cpu-options" {
		t.Errorf("expected form name %q, got %q", "cpu-options", got)
	}
}

// TestFormSavedMsg_StatusMessage verifies the default status message behavior.
func TestFormSavedMsg_StatusMessage(t *testing.T) {
	// With custom status message
	msg := testSavedMsg{name: "ssh-password", status: "Password changed"}
	if got := FormSavedStatus(msg); got != "Password changed" {
		t.Errorf("expected status %q, got %q", "Password changed", got)
	}

	// Without custom status — should default to "<name> saved successfully"
	msg2 := testSavedMsg{name: "cpu-options"}
	if got := FormSavedStatus(msg2); got != "cpu-options saved successfully" {
		t.Errorf("expected default status %q, got %q", "cpu-options saved successfully", got)
	}
}

// testSavedMsg is a test implementation of FormSavedMsg.
type testSavedMsg struct {
	name   string
	status string
}

func (m testSavedMsg) IsFormSaved()          {}
func (m testSavedMsg) FormName() string      { return m.name }
func (m testSavedMsg) FormStatus() string    { return m.status }
