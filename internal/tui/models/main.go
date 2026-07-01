package models

import (
	tea "charm.land/bubbletea/v2"
)

func (m *MainModel) Init() tea.Cmd {
	return m.init()
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.update(msg)
}

func (m *MainModel) View() tea.View {
	v := tea.NewView(m.view())
	// AltScreen is disabled in debug mode so log output remains visible.
	v.AltScreen = !m.debugMode
	return v
}
