// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// StartStopScriptFormUpdate handles update messages for the start/stop script form
// Deprecated: delegation is now through StartStopScriptModel.Update() in key_handlers.go.
func (m *MainModel) StartStopScriptFormUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	model := m.startStopScriptModel
	if model == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleStartStopScriptFormKeyPress(msg)
	case FileSelectedMsg:
		// Forward to the form model to handle file selection
		inner, cmd := model.Update(msg)
		if ssm, ok := inner.(*StartStopScriptModel); ok {
			m.startStopScriptModel = ssm
		}
		return m, cmd
	}

	return m, nil
}

// handleStartStopScriptFormKeyPress handles key presses in the start/stop script form
// Deprecated: delegation is now through StartStopScriptModel.Update() in key_handlers.go.
func (m *MainModel) handleStartStopScriptFormKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	model := m.startStopScriptModel
	if model == nil {
		return m, nil
	}
	sf := model.Form()

	// Delegate to active file browser first
	if sf.fileBrowser != nil && sf.fileBrowser.active {
		inner, cmd := sf.fileBrowser.Update(msg)
		if fb, ok := inner.(*FileBrowserModel); ok {
			sf.fileBrowser = fb
			m.startStopScriptModel = model
		}
		return m, cmd
	}

	key := msg.String()

	// For enter/space, delegate to form model to handle browse buttons properly
	// The form model will create a file browser when focused on browse buttons
	// For toggle/save/cancel, use handleStartStopScriptFormEnter directly
	if key == "enter" || key == " " {
		if sf.focusIndex < len(sf.positions) {
			pos := sf.positions[sf.focusIndex]
			if pos.Key == "start_browse" || pos.Key == "stop_browse" {
				// Delegate to form model to open file browser
				inner, cmd := model.Update(msg)
				if ssm, ok := inner.(*StartStopScriptModel); ok {
					m.startStopScriptModel = ssm
				}
				return m, cmd
			}
		}
		// For toggle/save/cancel, handle directly
		return m.handleStartStopScriptFormEnter()
	}

	switch key {
	case "tab":
		sf.focusIndex++
		if sf.focusIndex >= len(sf.positions) {
			sf.focusIndex = 0
		}
		sf.rebuildPositions()

	case "shift+tab":
		sf.focusIndex--
		if sf.focusIndex < 0 {
			sf.focusIndex = len(sf.positions) - 1
		}
		sf.rebuildPositions()

	case "up":
		sf.focusIndex--
		if sf.focusIndex < 0 {
			sf.focusIndex = len(sf.positions) - 1
		}
		sf.rebuildPositions()

	case "down":
		sf.focusIndex++
		if sf.focusIndex >= len(sf.positions) {
			sf.focusIndex = 0
		}
		sf.rebuildPositions()

	case "esc":
		m.currentView = ViewConfigMenu
		m.breadcrumbs.Clear()
		m.breadcrumbs.AddItem("Configuration", "config", 1)
	}

	// Trigger re-render
	sf.syncViewport()
	m.startStopScriptModel = model
	return m, nil
}

// handleStartStopScriptFormEnter handles Enter key in the start/stop script form
// Deprecated: delegation is now through StartStopScriptModel.Update() in key_handlers.go.
func (m *MainModel) handleStartStopScriptFormEnter() (tea.Model, tea.Cmd) {
	model := m.startStopScriptModel
	if model == nil {
		return m, nil
	}
	sf := model.Form()

	// Check which position is focused
	if sf.focusIndex < len(sf.positions) {
		pos := sf.positions[sf.focusIndex]

		switch pos.Kind {
		case form.FocusToggle:
			// Toggle between builtin and custom
			sf.config.UseBuiltin = !sf.config.UseBuiltin
			sf.rebuildPositions()

		case form.FocusButton:
			switch pos.Key {
			case "save":
				// Save the configuration
				if err := m.vmManager.SaveStartStopScript(sf.config); err != nil {
					m.statusMessage = "Error saving start/stop script config: " + err.Error()
				} else {
					m.statusMessage = "Start/Stop script configuration saved"
				}
				m.currentView = ViewConfigMenu
				m.breadcrumbs.Clear()
				m.breadcrumbs.AddItem("Configuration", "config", 1)

			case "cancel":
				// Cancel and go back
				m.currentView = ViewConfigMenu
				m.breadcrumbs.Clear()
				m.breadcrumbs.AddItem("Configuration", "config", 1)
			}
		}
	}

	sf.syncViewport()
	m.startStopScriptModel = model
	return m, nil
}
