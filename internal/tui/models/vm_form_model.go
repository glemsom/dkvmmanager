// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VMFormModel is the shared scrollable form for creating and editing VMs.
// It implements the form.FormModel interface for use with ScrollableForm.
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
	tpmEnabled  bool

	// Focus state - uses form.FocusPos
	positions     []form.FocusPos
	focusIndex    int
	cursorOffsets map[string]int // key = field identity e.g. "vmName", "hardDisks_0"
	errors        map[string]string

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

// --- FormModel Interface Implementation ---

// BuildPositions returns the current list of navigable positions.
func (m *VMFormModel) BuildPositions() []form.FocusPos {
	return m.positions
}

// CurrentIndex returns the index of the currently focused position.
func (m *VMFormModel) CurrentIndex() int {
	return m.focusIndex
}

// SetFocusIndex sets the focused position index.
func (m *VMFormModel) SetFocusIndex(i int) {
	m.focusIndex = i
}

// RenderHeader returns the form header.
func (m *VMFormModel) RenderHeader() string {
	return "Create/Edit VM"
}

// RenderFooter returns the form footer.
func (m *VMFormModel) RenderFooter() string {
	return "Tab Navigate  Space/Enter Browse  ESC Cancel"
}

// OnEnter is called when the form becomes active.
func (m *VMFormModel) OnEnter() {}

// OnExit is called when the form is dismissed.
func (m *VMFormModel) OnExit() {}

// SetSize is called when the form dimensions change.
func (m *VMFormModel) SetSize(width int, height int) {}

// SetFocused sets whether the form has keyboard focus.
func (m *VMFormModel) SetFocused(bool) {}

// --- Internal Helpers ---

// rebuildPositions reconstructs the flat focus list from current field values
func (m *VMFormModel) rebuildPositions() {
	m.positions = nil

	// VM Name
	m.positions = append(m.positions, form.FocusPos{Kind: form.FocusText, Label: "VM Name", Key: "vmName"})

	// Hard Disks label (header-style)
	m.positions = append(m.positions, form.FocusPos{Kind: form.FocusHeader, Label: "Hard Disks:", Key: "hardDisks_label"})
	// Hard Disk items
	for i := range m.hardDisks {
		m.positions = append(m.positions, form.FocusPos{
			Kind:  form.FocusList,
			Label: fmt.Sprintf("Disk %d", i+1),
			Key:   fmt.Sprintf("hardDisks_%d", i),
			Data:  i,
		})
	}
	// Add button
	m.positions = append(m.positions, form.FocusPos{Kind: form.FocusButton, Label: "[+] Add Disk", Key: "hardDisks_add"})

	// CDROMs label (header-style)
	m.positions = append(m.positions, form.FocusPos{Kind: form.FocusHeader, Label: "CD/DVD Drives (ISOs):", Key: "cdroms_label"})
	// CDROM items
	for i := range m.cdroms {
		m.positions = append(m.positions, form.FocusPos{
			Kind:  form.FocusList,
			Label: fmt.Sprintf("CDROM %d", i+1),
			Key:   fmt.Sprintf("cdroms_%d", i),
			Data:  i,
		})
	}
	// Add button
	m.positions = append(m.positions, form.FocusPos{Kind: form.FocusButton, Label: "[+] Add CDROM", Key: "cdroms_add"})

	// Text fields
	m.positions = append(m.positions, form.FocusPos{Kind: form.FocusText, Label: "MAC Address", Key: "macAddress"})

	// Toggle fields
	m.positions = append(m.positions, form.FocusPos{Kind: form.FocusToggle, Label: "VNC", Key: "vncEnabled"})
	m.positions = append(m.positions, form.FocusPos{Kind: form.FocusToggle, Label: "Network", Key: "networkMode"})
	m.positions = append(m.positions, form.FocusPos{Kind: form.FocusToggle, Label: "TPM", Key: "tpmEnabled"})

	// Save button
	m.positions = append(m.positions, form.FocusPos{Kind: form.FocusButton, Label: "Save", Key: "save"})
}

// posKey returns a unique key for the position (used for cursor offsets / errors)
func (m *VMFormModel) posKey(pos form.FocusPos) string {
	return pos.Key
}

// getValue returns the text value at a focus position
func (m *VMFormModel) getValue(pos form.FocusPos) string {
	switch pos.Key {
	case "vmName":
		return m.vmName
	case "macAddress":
		return m.macAddress
	}
	// Check for list items (hardDisks_N, cdroms_N)
	if pos.Kind == form.FocusList {
		var disks *[]string
		switch {
		case len(pos.Key) > 9 && pos.Key[:9] == "hardDisks":
			disks = &m.hardDisks
		case len(pos.Key) > 6 && pos.Key[:6] == "cdroms":
			disks = &m.cdroms
		}
		if disks != nil {
			var idx int
			_, _ = fmt.Sscanf(pos.Key, "%*[^0]%d", &idx)
			if idx < len(*disks) {
				return (*disks)[idx]
			}
		}
	}
	return ""
}

// setValue sets the text value at a focus position
func (m *VMFormModel) setValue(pos form.FocusPos, val string) {
	switch pos.Key {
	case "vmName":
		m.vmName = val
	case "macAddress":
		m.macAddress = val
	}
	// Check for list items (hardDisks_N, cdroms_N)
	if pos.Kind == form.FocusList {
		var disks *[]string
		switch {
		case len(pos.Key) > 9 && pos.Key[:9] == "hardDisks":
			disks = &m.hardDisks
		case len(pos.Key) > 6 && pos.Key[:6] == "cdroms":
			disks = &m.cdroms
		}
		if disks != nil {
			var idx int
			_, _ = fmt.Sscanf(pos.Key, "%*[^0]%d", &idx)
			if idx < len(*disks) {
				(*disks)[idx] = val
			}
		}
	}
}

// cursorOffset returns the cursor offset for the given position key
func (m *VMFormModel) cursorOffset(key string) int {
	if off, ok := m.cursorOffsets[key]; ok {
		return off
	}
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