// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "charm.land/bubbletea/v2"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// LVCreateModel wraps the LVCreateFormModel in a form.ScrollableForm.
type LVCreateModel struct {
	form *form.ScrollableForm
}

// NewLVCreateModel creates a new LVCreateModel.
func NewLVCreateModel(dryRunMode, debugMode bool) *LVCreateModel {
	fm := NewLVCreateFormModel(dryRunMode, debugMode)
	return &LVCreateModel{form: form.NewScrollableForm(fm)}
}

// Init implements tea.Model.
func (m *LVCreateModel) Init() tea.Cmd {
	return m.form.Init()
}

// Update implements tea.Model.
func (m *LVCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View implements tea.Model.
func (m *LVCreateModel) View() tea.View {
	return m.form.View()
}

// SetSize forwards window resize to the underlying form.
func (m *LVCreateModel) SetSize(width, height int) { m.form.SetSize(width, height) }

// FileBrowserActive returns false (LV create has no file browser).
func (m *LVCreateModel) FileBrowserActive() bool { return false }

// Form returns the underlying ScrollableForm.
func (m *LVCreateModel) Form() *form.ScrollableForm {
	return m.form
}
