// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// Update implements tea.Model
func (m *StartStopScriptFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Delegate to active file browser first
	if m.fileBrowser != nil && m.fileBrowser.active {
		inner, cmd := m.fileBrowser.Update(msg)
		if fb, ok := inner.(*FileBrowserModel); ok {
			m.fileBrowser = fb
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", " ":
			// Enter or Space toggles the mode when focused on toggle
			// or opens file browser when focused on browse buttons
			return m.handleEnter()
		case "left":
			// Left selects "Use Builtin Script" when focused on toggle
			m.handleArrowKey(false)
		case "right":
			// Right selects "Use Custom Script" when focused on toggle
			m.handleArrowKey(true)
		case "tab":
			m.focusIndex++
			if m.focusIndex >= len(m.positions) {
				m.focusIndex = 0
			}
		case "shift+tab":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.positions) - 1
			}
		case "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.positions) - 1
			}
		case "down":
			m.focusIndex++
			if m.focusIndex >= len(m.positions) {
				m.focusIndex = 0
			}
		}
		m.syncViewport()
	case DirectoryLoadedMsg:
		// Directory loaded in file browser - trigger view refresh
		m.syncViewport()
	case FileSelectedMsg:
		// Handle file selection from browser
		m.handleFileSelected(msg)
	}
	return m, nil
}

// handleEnter handles Enter key in the form
func (m *StartStopScriptFormModel) handleEnter() (tea.Model, tea.Cmd) {
	if m.focusIndex < len(m.positions) {
		pos := m.positions[m.focusIndex]
		switch pos.kind {
		case startStopScriptToggle:
			m.config.UseBuiltin = !m.config.UseBuiltin
			m.rebuildPositions()

		case startStopScriptStartBrowse:
			// Open file browser for start script
			m.fileBrowser = NewFileBrowserModel(FileTypeAll)
			m.fileBrowser.SetDirectory("/media/dkvmdata")
			return m, m.fileBrowser.Init()

		case startStopScriptStopBrowse:
			// Open file browser for stop script
			m.fileBrowser = NewFileBrowserModel(FileTypeAll)
			m.fileBrowser.SetDirectory("/media/dkvmdata")
			return m, m.fileBrowser.Init()
		}
	}
	return m, nil
}

// handleFileSelected processes the result from the file browser
func (m *StartStopScriptFormModel) handleFileSelected(msg FileSelectedMsg) {
	m.fileBrowser = nil
	if !msg.Canceled && m.focusIndex < len(m.positions) {
		pos := m.positions[m.focusIndex]
		switch pos.kind {
		case startStopScriptStartBrowse:
			m.config.StartScript = msg.Path
		case startStopScriptStopBrowse:
			m.config.StopScript = msg.Path
		}
	}
	m.syncViewport()
}

// handleArrowKey handles arrow keys when focused on the toggle
func (m *StartStopScriptFormModel) handleArrowKey(moveRight bool) {
	if m.focusIndex < len(m.positions) {
		pos := m.positions[m.focusIndex]
		if pos.kind == startStopScriptToggle {
			if moveRight {
				m.config.UseBuiltin = false
			} else {
				m.config.UseBuiltin = true
			}
			m.rebuildPositions()
		}
	}
}

// startStopScriptFocusKind describes what a start/stop script form focus position represents
type startStopScriptFocusKind int

const (
	startStopScriptToggle startStopScriptFocusKind = iota
	startStopScriptStartPath
	startStopScriptStartBrowse
	startStopScriptStopPath
	startStopScriptStopBrowse
	startStopScriptSave
	startStopScriptCancel
)

// startStopScriptFocusPos is one navigable position in the CustomScript form
type startStopScriptFocusPos struct {
	kind      startStopScriptFocusKind
	fieldName string
}

// StartStopScriptFormModel is a form for editing start/stop scripts
type StartStopScriptFormModel struct {
	vmManager *vm.Manager
	config    models.StartStopScript
	pciConfig models.PCIPassthroughConfig // For displaying in builtin mode

	// Flat list of focusable positions
	positions  []startStopScriptFocusPos
	focusIndex int

	// File browser state (for script path selection via FileBrowserModel)
	fileBrowser *FileBrowserModel

	// Per-field text cursor (character offset within the value)
	cursorOffsets map[string]int

	// Per-field inline error messages
	errors map[string]string

	// Scrollable viewport
	vp       viewport.Model
	ready    bool
	contentW int
	contentH int

	// Rendering cache
	renderedLines []string
}

// NewStartStopScriptFormModel creates a new StartStopScript form model
func NewStartStopScriptFormModel(vmManager *vm.Manager) (*StartStopScriptFormModel, error) {
	cfg, _ := vmManager.GetStartStopScript()
	pciCfg, _ := vmManager.GetPCIPassthroughConfig()

	m := &StartStopScriptFormModel{
		vmManager:     vmManager,
		config:        cfg,
		pciConfig:     pciCfg,
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	m.rebuildPositions()
	return m, nil
}

// cursorOffset returns the cursor offset for the given position key
func (m *StartStopScriptFormModel) cursorOffset(key string) int {
	if off, ok := m.cursorOffsets[key]; ok {
		return off
	}
	return -1 // sentinel meaning "end"
}

// setCursorOffset sets cursor offset; -1 means end
func (m *StartStopScriptFormModel) setCursorOffset(key string, off int) {
	m.cursorOffsets[key] = off
}

// effectiveCursor returns the actual cursor position (0-based character index)
func (m *StartStopScriptFormModel) effectiveCursor(key string, val string) int {
	off := m.cursorOffset(key)
	if off < 0 {
		return len(val)
	}
	if off > len(val) {
		return len(val)
	}
	return off
}

// rebuildPositions rebuilds the list of focusable positions based on current config
func (m *StartStopScriptFormModel) rebuildPositions() {
	m.positions = []startStopScriptFocusPos{
		{kind: startStopScriptToggle, fieldName: "toggle"},
	}

	// Add path fields only in custom mode
	if !m.config.UseBuiltin {
		m.positions = append(m.positions, startStopScriptFocusPos{kind: startStopScriptStartPath, fieldName: "start_path"})
		m.positions = append(m.positions, startStopScriptFocusPos{kind: startStopScriptStartBrowse, fieldName: "start_browse"})
		m.positions = append(m.positions, startStopScriptFocusPos{kind: startStopScriptStopPath, fieldName: "stop_path"})
		m.positions = append(m.positions, startStopScriptFocusPos{kind: startStopScriptStopBrowse, fieldName: "stop_browse"})
	}

	m.positions = append(m.positions, startStopScriptFocusPos{kind: startStopScriptSave, fieldName: "save"})
	m.positions = append(m.positions, startStopScriptFocusPos{kind: startStopScriptCancel, fieldName: "cancel"})

	// Ensure focusIndex is valid
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = 0
	}
}

// Init implements tea.Model
func (m *StartStopScriptFormModel) Init() tea.Cmd {
	return nil
}

// currentPos returns the focus position at the current focusIndex
func (m *StartStopScriptFormModel) currentPos() startStopScriptFocusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return startStopScriptFocusPos{}
	}
	return m.positions[m.focusIndex]
}
