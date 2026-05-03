package models

// VCPUPinningUpdatedMsg indicates vCPU pinning config was saved.
// It implements form.FormSavedMsg for the ScrollableForm framework.
type VCPUPinningUpdatedMsg struct{}

// IsFormSaved implements form.FormSavedMsg.
func (VCPUPinningUpdatedMsg) IsFormSaved() {}

// FormName implements form.FormSavedMsg.
func (VCPUPinningUpdatedMsg) FormName() string { return "vCPU Pinning" }

// FormStatus implements form.FormSavedMsg.
func (VCPUPinningUpdatedMsg) FormStatus() string { return "" }
