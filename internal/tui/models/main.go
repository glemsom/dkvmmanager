package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *MainModel) Init() tea.Cmd {
	return m.init()
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.update(msg)
}

func (m *MainModel) View() string {
	return m.view()
}
