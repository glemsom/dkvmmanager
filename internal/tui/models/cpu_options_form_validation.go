// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import "fmt"

// Validation logic is embedded in handlers via validateAndSave
// This file exists for potential future extraction of complex validation rules

// validateField performs field-specific validation
func (m *CPUOptionsFormModel) validateField(fieldName string, value string) string {
	switch fieldName {
	case "VendorID":
		// VendorID should be 12 characters or empty
		if len(value) > 0 && len(value) != 12 {
			return "must be exactly 12 characters"
		}
	case "HVSpinlocks":
		// Should be a valid number or empty
		if value != "" {
			// Accept hex format like "0x10000 or decimal
			// Simple validation: must be numeric or hex
			// This is handled by the model in practice
		}
	}
	return ""
}

// clearErrors clears all validation errors
func (m *CPUOptionsFormModel) clearErrors() {
	m.errors = make(map[string]string)
}

// hasErrors returns true if there are any validation errors
func (m *CPUOptionsFormModel) hasErrors() bool {
	return len(m.errors) > 0
}

// errorForField returns the error message for a specific field
func (m *CPUOptionsFormModel) errorForField(fieldName string) string {
	if err, ok := m.errors[fieldName]; ok {
		return err
	}
	return ""
}

// setError sets an error for a specific field
func (m *CPUOptionsFormModel) setError(fieldName string, errMsg string) {
	m.errors[fieldName] = errMsg
}

// saveWithValidation performs validation before saving
// Returns error if validation fails
func (m *CPUOptionsFormModel) saveWithValidation() error {
	// Clear previous errors
	m.clearErrors()

	// Validate VendorID
	if err := m.validateField("VendorID", m.getTextValue("VendorID")); err != "" {
		m.setError("VendorID", err)
		return fmt.Errorf("VendorID: %s", err)
	}

	// Save to VM manager (dereference pointer)
	if err := m.vmManager.SaveCPUOptions(*m.options); err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	return nil
}
