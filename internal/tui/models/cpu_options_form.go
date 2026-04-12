// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// cpuOptFocusKind describes what a CPU options focus position represents
type cpuOptFocusKind int

const (
	cpuOptToggle cpuOptFocusKind = iota
	cpuOptText
	cpuOptSave
)

// cpuOptFocusPos is one navigable position in the CPU options form
type cpuOptFocusPos struct {
	kind      cpuOptFocusKind
	fieldName string
}

// posKey returns a unique key for the position
func cpuOptPosKey(p cpuOptFocusPos) string {
	return p.fieldName
}

// CPUOptionsFormModel is a scrollable toggle form for editing global CPU options
type CPUOptionsFormModel struct {
	vmManager *vm.Manager
	options   models.CPUOptions

	// Flat list of focusable positions
	positions  []cpuOptFocusPos
	focusIndex int

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

// NewCPUOptionsFormModel creates a new CPU options form model
func NewCPUOptionsFormModel(vmManager *vm.Manager) *CPUOptionsFormModel {
	opts, _ := vmManager.GetCPUOptions()
	m := &CPUOptionsFormModel{
		vmManager:     vmManager,
		options:       opts,
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	m.rebuildPositions()
	return m
}

// cursorOffset returns the cursor offset for the given position key
func (m *CPUOptionsFormModel) cursorOffset(key string) int {
	if off, ok := m.cursorOffsets[key]; ok {
		return off
	}
	return -1 // sentinel meaning "end"
}

// setCursorOffset sets cursor offset; -1 means end
func (m *CPUOptionsFormModel) setCursorOffset(key string, off int) {
	m.cursorOffsets[key] = off
}

// effectiveCursor returns the actual cursor position (0-based character index)
func (m *CPUOptionsFormModel) effectiveCursor(key string, val string) int {
	off := m.cursorOffset(key)
	if off < 0 {
		return len(val)
	}
	if off > len(val) {
		return len(val)
	}
	return off
}

// Init implements tea.Model
func (m *CPUOptionsFormModel) Init() tea.Cmd {
	return nil
}
