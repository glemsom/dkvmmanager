package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// sendRunes sends individual rune key presses through Update() and returns the updated model.
// Each rune is sent as a separate KeyMsg, exercising the full Update() dispatch pipeline.
func sendRunes(t *testing.T, m *MainModel, runes ...rune) *MainModel {
	t.Helper()
	for _, r := range runes {
		model, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = model.(*MainModel)
	}
	return m
}

// sendKeys sends special key presses (KeyEnter, KeyEsc, KeyTab, etc.) through Update()
// and returns the updated model.
func sendKeys(t *testing.T, m *MainModel, keys ...tea.KeyType) *MainModel {
	t.Helper()
	for _, k := range keys {
		model, _ := m.Update(tea.KeyMsg{Type: k})
		m = model.(*MainModel)
	}
	return m
}

// sendKeysWithCmd sends key presses through Update() and executes any returned tea.Cmd,
// feeding the resulting message back through Update(). Returns the final model.
// This exercises the full async command pipeline that real user interactions trigger.
func sendKeysWithCmd(t *testing.T, m *MainModel, keys ...tea.KeyType) *MainModel {
	t.Helper()
	for _, k := range keys {
		model, cmd := m.Update(tea.KeyMsg{Type: k})
		m = model.(*MainModel)
		if cmd != nil {
			msg := cmd()
			model, _ = m.Update(msg)
			m = model.(*MainModel)
		}
	}
	return m
}

// setupTestModelWithTwoVMs creates a MainModel with exactly 2 VMs and standard test dimensions.
// This is shared across scenario_test.go and golden_test.go.
func setupTestModelForScenarios(t *testing.T) *MainModel {
	t.Helper()
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30
	// Reset list cursor to first item for deterministic tests
	m.menuList.Select(0)
	return m
}
