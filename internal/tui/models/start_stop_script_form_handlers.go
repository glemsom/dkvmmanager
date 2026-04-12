// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

// StartStopScriptFormUpdate handles update messages for the start/stop script form
func (m *MainModel) StartStopScriptFormUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	form := m.startStopScriptFormModel
	if form == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleStartStopScriptFormKeyPress(msg)
	case FileSelectedMsg:
		// Forward to the form model to handle file selection
		inner, cmd := form.Update(msg)
		if f, ok := inner.(*StartStopScriptFormModel); ok {
			m.startStopScriptFormModel = f
		}
		return m, cmd
	}

	return m, nil
}

// handleStartStopScriptFormKeyPress handles key presses in the start/stop script form
func (m *MainModel) handleStartStopScriptFormKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	form := m.startStopScriptFormModel
	if form == nil {
		return m, nil
	}

	// Delegate to active file browser first
	if form.fileBrowser != nil && form.fileBrowser.active {
		inner, cmd := form.fileBrowser.Update(msg)
		if fb, ok := inner.(*FileBrowserModel); ok {
			form.fileBrowser = fb
			m.startStopScriptFormModel = form
		}
		return m, cmd
	}

	key := msg.String()


	// For enter/space, delegate to form model to handle browse buttons properly
	// The form model will create a file browser when focused on browse buttons
	// For toggle/save/cancel, use handleStartStopScriptFormEnter directly
	if key == "enter" || key == " " {
		if form.focusIndex < len(form.positions) {
			pos := form.positions[form.focusIndex]
			if pos.kind == startStopScriptStartBrowse || pos.kind == startStopScriptStopBrowse {
				// Delegate to form model to open file browser
				inner, cmd := form.Update(msg)
				if f, ok := inner.(*StartStopScriptFormModel); ok {
					m.startStopScriptFormModel = f
				}
				return m, cmd
			}
		}
		// For toggle/save/cancel, handle directly
		return m.handleStartStopScriptFormEnter()
	}

	switch key {
	case "tab":
		form.focusIndex++
		if form.focusIndex >= len(form.positions) {
			form.focusIndex = 0
		}
		form.rebuildPositions()

	case "shift+tab":
		form.focusIndex--
		if form.focusIndex < 0 {
			form.focusIndex = len(form.positions) - 1
		}
		form.rebuildPositions()

	case "up":
		form.focusIndex--
		if form.focusIndex < 0 {
			form.focusIndex = len(form.positions) - 1
		}
		form.rebuildPositions()

	case "down":
		form.focusIndex++
		if form.focusIndex >= len(form.positions) {
			form.focusIndex = 0
		}
		form.rebuildPositions()

	case "esc":
		m.currentView = ViewConfigMenu
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Configuration", "config", 1)
	}

	// Trigger re-render
	form.syncViewport()
	m.startStopScriptFormModel = form
	return m, nil
}

// handleStartStopScriptFormEnter handles Enter key in the start/stop script form
func (m *MainModel) handleStartStopScriptFormEnter() (tea.Model, tea.Cmd) {
	form := m.startStopScriptFormModel
	if form == nil {
		return m, nil
	}

	// Check which position is focused
	if form.focusIndex < len(form.positions) {
		pos := form.positions[form.focusIndex]

		switch pos.kind {
		case startStopScriptToggle:
			// Toggle between builtin and custom
			form.config.UseBuiltin = !form.config.UseBuiltin
			form.rebuildPositions()

		case startStopScriptSave:
			// Save the configuration
			if err := m.vmManager.SaveStartStopScript(form.config); err != nil {
				m.statusMessage = "Error saving start/stop script config: " + err.Error()
			} else {
				m.statusMessage = "Start/Stop script configuration saved"
			}
			m.currentView = ViewConfigMenu
			m.breadcrumbs.Clear()
			m.breadcrumbs.AddItem("Configuration", "config", 1)

		case startStopScriptCancel:
			// Cancel and go back
			m.currentView = ViewConfigMenu
			m.breadcrumbs.Clear()
			m.breadcrumbs.AddItem("Configuration", "config", 1)
		}
	}

	form.syncViewport()
	m.startStopScriptFormModel = form
	return m, nil
}
