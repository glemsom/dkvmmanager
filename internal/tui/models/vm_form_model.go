// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VMFormModel is the shared scrollable form for creating and editing VMs
type VMFormModel struct {
	mode      FormMode
	vmManager *vm.Manager
	vm        *models.VM // non-nil only in edit mode

	// Field values
	vmName      string
	hardDisks   []string
	cdroms      []string
	macAddress  string
	vncEnabled  bool
	networkMode string // "bridge" or "nat"
	tpmEnabled    bool

	// Flat list of focusable positions
	positions  []focusPos
	focusIndex int

	// Per-field text cursor (character offset within the value)
	cursorOffsets map[string]int // key = field identity e. g. "vmName", "hardDisks_0"

	// Per-field inline error messages
	errors map[string]string

	// Scrollable viewport
	vp       viewport.Model
	ready    bool
	contentW int
	contentH int

	// Rendering cache
	renderedLines []string

	// File browser state (for ISO selection via FileBrowserModel)
	fileBrowser *FileBrowserModel

	// Add disk model state (for harddisk selection via AddDiskModel)
	addDiskModel *AddDiskModel

	// Which list field is being browsed ("hardDisks" or "cdroms")
	browsingFieldName string

	// Which list slot index is being edited
	browsingIndex int
}

// NewVMFormModel creates a form in Create mode with sensible defaults
func NewVMFormModel(vmManager *vm.Manager) *VMFormModel {
	m := &VMFormModel{
		mode:          FormCreate,
		vmManager:     vmManager,
		macAddress:    vmManager.GenerateMAC(),
		vncEnabled:    true,
		networkMode:   "nat",
		tpmEnabled:    false,
		hardDisks:     []string{""}, // start with one empty disk slot
		cdroms:        []string{},
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	m.rebuildPositions()
	return m
}

// NewVMFormModelEdit creates a form in Edit mode pre-filled from an existing VM
func NewVMFormModelEdit(vmManager *vm.Manager, vmObj *models.VM) *VMFormModel {
	hd := vmObj.HardDisks
	if len(hd) == 0 {
		hd = []string{""}
	}
	cd := vmObj.CDROMs
	if len(cd) == 0 {
		cd = []string{}
	}
	netMode := vmObj.NetworkMode
	if netMode == "" {
		netMode = "nat"
	}
	m := &VMFormModel{
		mode:          FormEdit,
		vmManager:     vmManager,
		vm:            vmObj,
		vmName:        vmObj.Name,
		hardDisks:     hd,
		cdroms:        cd,
		macAddress:    vmObj.MAC,
		vncEnabled:    vmObj.VNCListen != "",
		networkMode:   netMode,
		tpmEnabled:    vmObj.TPMEnabled,
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
	}
	m.rebuildPositions()
	return m
}

// rebuildPositions reconstructs the flat focus list from current field values
func (m *VMFormModel) rebuildPositions() {
	m.positions = nil

	// VM Name
	m.positions = append(m.positions, focusPos{kind: focusText, fieldName: "vmName"})

	// Hard Disks label + N items + Add button
	m.positions = append(m.positions, focusPos{kind: focusText, fieldName: "hardDisks_label"})
	for i := range m.hardDisks {
		m.positions = append(m.positions, focusPos{kind: focusListItem, fieldName: "hardDisks", listIndex: i})
	}
	m.positions = append(m.positions, focusPos{kind: focusAddBtn, fieldName: "hardDisks"})

	// CDROMs label + N items + Add button
	m.positions = append(m.positions, focusPos{kind: focusText, fieldName: "cdroms_label"})
	for i := range m.cdroms {
		m.positions = append(m.positions, focusPos{kind: focusListItem, fieldName: "cdroms", listIndex: i})
	}
	m.positions = append(m.positions, focusPos{kind: focusAddBtn, fieldName: "cdroms"})

	// Text fields
	m.positions = append(m.positions, focusPos{kind: focusText, fieldName: "macAddress"})
	m.positions = append(m.positions, focusPos{kind: focusToggle, fieldName: "vncEnabled"})
	m.positions = append(m.positions, focusPos{kind: focusToggle, fieldName: "networkMode"})
	m.positions = append(m.positions, focusPos{kind: focusToggle, fieldName: "tpmEnabled"})

	// Save button
	m.positions = append(m.positions, focusPos{kind: focusSaveBtn, fieldName: "save"})
}

// currentPos returns the focus position at the current focusIndex
func (m *VMFormModel) currentPos() focusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return focusPos{kind: focusText, fieldName: "vmName"}
	}
	return m.positions[m.focusIndex]
}

// posKey returns a unique key for the position (used for cursor offsets / errors)
func posKey(p focusPos) string {
	if p.kind == focusListItem {
		return fmt.Sprintf("%s_%d", p.fieldName, p.listIndex)
	}
	return p.fieldName
}

// getValue returns the text value at a focus position
func (m *VMFormModel) getValue(p focusPos) string {
	switch p.fieldName {
	case "vmName":
		return m.vmName
	case "hardDisks":
		if p.kind == focusListItem && p.listIndex < len(m.hardDisks) {
			return m.hardDisks[p.listIndex]
		}
	case "cdroms":
		if p.kind == focusListItem && p.listIndex < len(m.cdroms) {
			return m.cdroms[p.listIndex]
		}
	case "macAddress":
		return m.macAddress
	}
	return ""
}

// setValue sets the text value at a focus position
func (m *VMFormModel) setValue(p focusPos, val string) {
	switch p.fieldName {
	case "vmName":
		m.vmName = val
	case "hardDisks":
		if p.kind == focusListItem && p.listIndex < len(m.hardDisks) {
			m.hardDisks[p.listIndex] = val
		}
	case "cdroms":
		if p.kind == focusListItem && p.listIndex < len(m.cdroms) {
			m.cdroms[p.listIndex] = val
		}
	case "macAddress":
		m.macAddress = val
	}
}

// cursorOffset returns the cursor offset for the given position key
func (m *VMFormModel) cursorOffset(key string) int {
	if off, ok := m.cursorOffsets[key]; ok {
		return off
	}
	// Default: cursor at end of current value
	return -1 // sentinel meaning "end"
}

// setCursorOffset sets cursor offset; -1 means end
func (m *VMFormModel) setCursorOffset(key string, off int) {
	m.cursorOffsets[key] = off
}

// effectiveCursor returns the actual cursor position (0-based character index)
func (m *VMFormModel) effectiveCursor(key string, val string) int {
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
func (m *VMFormModel) Init() tea.Cmd {
	return nil
}

// FileBrowserActive returns true if the file browser or add disk model is currently active
func (m *VMFormModel) FileBrowserActive() bool {
	if m.fileBrowser != nil && m.fileBrowser.active {
		return true
	}
	if m.addDiskModel != nil && m.addDiskModel.active {
		return true
	}
	return false
}
