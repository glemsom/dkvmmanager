package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming messages
func (m *BlockDeviceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case BlockDeviceLoadedMsg:
		// Devices already loaded via side effect in loadDevices command.
		// This message exists to trigger a view refresh.
		return m, nil
	}
	return m, nil
}

// handleKeyPress handles keyboard input
func (m *BlockDeviceModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errorMsg = ""

	switch msg.String() {
	case "ctrl+c", "esc":
		m.active = false
		return m, func() tea.Msg { return FileSelectedMsg{Path: "", Canceled: true} }

	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		if m.selectedIndex < len(m.devices)-1 {
			m.selectedIndex++
		}

	case "enter", " ":
		return m.handleEnter()
	}

	return m, nil
}

// handleEnter handles the enter key
func (m *BlockDeviceModel) handleEnter() (tea.Model, tea.Cmd) {
	if len(m.devices) == 0 {
		return m, nil
	}

	selected := m.devices[m.selectedIndex]
	m.active = false
	m.selectedPath = selected.Path

	return m, func() tea.Msg { return FileSelectedMsg{Path: selected.Path, Canceled: false} }
}

// Update handles incoming messages
func (m *AddDiskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle FileSelectedMsg even when inactive — the model may have been
	// deactivated during command execution in a prior Update call, but the
	// FileSelectedMsg arrives in a subsequent call. We must still convert it
	// to DiskAddedMsg so the caller gets the result.
	if fsm, ok := msg.(FileSelectedMsg); ok {
		return m.handleFileSelected(fsm)
	}

	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Delegate to internal models when in step 1, 2, or 3
		if m.step == 1 && m.fileBrowser != nil {
			inner, cmd := m.fileBrowser.Update(msg)
			if fb, ok := inner.(*FileBrowserModel); ok {
				m.fileBrowser = fb
			}
			return m, cmd
		}
		if m.step == 2 && m.blockDevice != nil {
			inner, cmd := m.blockDevice.Update(msg)
			if bd, ok := inner.(*BlockDeviceModel); ok {
				m.blockDevice = bd
			}
			return m, cmd
		}
		if m.step == 3 && m.lvmVolume != nil {
			inner, cmd := m.lvmVolume.Update(msg)
			if lv, ok := inner.(*LVMVolumeModel); ok {
				m.lvmVolume = lv
			}
			return m, cmd
		}
		return m.handleKeyPress(msg)
	}
	return m, nil
}

// handleKeyPress handles keyboard input
func (m *AddDiskModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errorMsg = ""

	switch msg.String() {
	case "ctrl+c", "esc":
		m.active = false
		return m, func() tea.Msg { return DiskAddedMsg{Path: "", Canceled: true} }

	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		if m.selectedIndex < 2 {
			m.selectedIndex++
		}

	case "enter", " ":
		return m.handleEnter()
	}

	return m, nil
}

// handleEnter handles enter key
func (m *AddDiskModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case 0:
		// Choose source type
		m.sourceType = DiskSourceType(m.selectedIndex)
		if m.selectedIndex == 0 {
			// File/image
			m.step = 1
			m.selectedIndex = 0
			return m, m.fileBrowser.Init()
		}
		if m.selectedIndex == 1 {
			// Block device
			m.step = 2
			m.selectedIndex = 0
			return m, m.blockDevice.Init()
		}
		// LVM Logical Volume
		m.step = 3
		m.selectedIndex = 0
		return m, m.lvmVolume.Init()
	}

	return m, nil
}

// handleFileSelected handles file selection from browser
func (m *AddDiskModel) handleFileSelected(msg FileSelectedMsg) (tea.Model, tea.Cmd) {
	if msg.Canceled {
		m.active = false
		return m, func() tea.Msg { return DiskAddedMsg{Path: "", Canceled: true} }
	}

	m.path = msg.Path
	m.active = false
	return m, func() tea.Msg { return DiskAddedMsg{Path: msg.Path, Canceled: false} }
}
