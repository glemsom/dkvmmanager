package form

// FormSavedMsg is a marker interface implemented by all form save result messages.
// It enables uniform handling in main.go without importing every concrete form type.
//
// Each form defines its own concrete message type (e.g. SSHPasswordUpdatedMsg)
// and implements this interface to participate in the framework's result handling.
type FormSavedMsg interface {
	// IsFormSaved is the marker method that identifies this as a form save result.
	IsFormSaved()

	// FormName returns a human-readable name for the form (used in status messages).
	FormName() string

	// FormStatus returns a custom status message, or empty string to use a default.
	FormStatus() string
}

// IsFormSavedMsg reports whether msg implements the FormSavedMsg interface.
func IsFormSavedMsg(msg any) bool {
	_, ok := msg.(FormSavedMsg)
	return ok
}

// FormSavedName extracts the form name from a FormSavedMsg.
// Returns empty string if msg does not implement FormSavedMsg.
func FormSavedName(msg any) string {
	if m, ok := msg.(FormSavedMsg); ok {
		return m.FormName()
	}
	return ""
}

// FormSavedStatus extracts the status message from a FormSavedMsg.
// Returns a default "<name> saved successfully" if FormStatus() is empty.
// Returns empty string if msg does not implement FormSavedMsg.
func FormSavedStatus(msg any) string {
	if m, ok := msg.(FormSavedMsg); ok {
		if status := m.FormStatus(); status != "" {
			return status
		}
		return m.FormName() + " saved successfully"
	}
	return ""
}
