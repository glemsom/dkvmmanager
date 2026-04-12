// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// vcpuPinningFocusKind describes what a vCPU pinning focus position represents
type vcpuPinningFocusKind int

const (
	vcpuPinningToggle vcpuPinningFocusKind = iota // Toggle pinning enabled
	vcpuPinningSave                              // Save button
)

// VCPUPinningFormModel is a scrollable form for editing global vCPU pinning
type VCPUPinningFormModel struct {
	vmManager *vm.Manager

	// Host topology data (read-only, for display)
	hostTopo models.HostCPUTopology

	// Global CPU topology (read-only, for computing pinning)
	topology models.CPUTopology

	// Global configuration
	pinning models.VCPUPinningGlobal

	// Focus position
	focusPos vcpuPinningFocusKind

	// Per-field inline error messages
	errors map[string]string

	// Scan error
	scanErr error

	// Scrollable viewport
	vp       viewport.Model
	ready    bool
	contentW int
	contentH int

	// Rendering cache
	renderedLines []string
}

// NewVCPUPinningFormModel creates a new vCPU pinning form model
func NewVCPUPinningFormModel(vmManager *vm.Manager) (*VCPUPinningFormModel, error) {
	// Scan host topology
	scanner := vm.NewCPUScanner()
	hostTopo, scanErr := scanner.ScanTopology()

	// Load global CPU topology config (needed to compute pinning)
	topology, err := vmManager.GetCPUTopology()
	if err != nil {
		return nil, fmt.Errorf("failed to load CPU topology: %w", err)
	}

	// Load global vCPU pinning config
	pinning, err := vmManager.GetVCPUPinningGlobal()
	if err != nil {
		return nil, fmt.Errorf("failed to load vcpu pinning: %w", err)
	}

	// If pinning is enabled but no mappings, recompute from topology
	if pinning.Enabled && len(pinning.Mappings) == 0 && scanErr == nil {
		pinning, err = vm.ComputePinningFromTopology(topology, hostTopo)
		if err != nil {
			// Fall back to saved config, will show warning
			pinning, _ = vmManager.GetVCPUPinningGlobal()
		}
	}

	m := &VCPUPinningFormModel{
		vmManager:  vmManager,
		hostTopo:   hostTopo,
		topology:   topology,
		pinning:   pinning,
		focusPos:  vcpuPinningToggle,
		errors:    make(map[string]string),
		scanErr:   scanErr,
	}

	m.syncViewport()
	return m, nil
}

// Init implements tea.Model
func (m *VCPUPinningFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *VCPUPinningFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.contentW = msg.Width
		m.contentH = msg.Height
		if !m.ready {
			m.vp = viewport.New(msg.Width, msg.Height)
			m.ready = true
		} else {
			m.vp.Width = msg.Width
			m.vp.Height = msg.Height
		}
		m.syncViewport()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// View implements tea.Model
func (m *VCPUPinningFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// SetSize updates the form dimensions
func (m *VCPUPinningFormModel) SetSize(w, h int) {
	m.contentW = w
	m.contentH = h
	if !m.ready {
		m.vp = viewport.New(w, h)
		m.ready = true
	} else {
		m.vp.Width = w
		m.vp.Height = h
	}
	m.syncViewport()
}

// handleKey processes keyboard input
func (m *VCPUPinningFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "tab", "down":
		m.moveFocus(1)
		m.syncViewport()
	case "shift+tab", "up":
		m.moveFocus(-1)
		m.syncViewport()
	case "p":
		m.pinning.Enabled = !m.pinning.Enabled
		m.syncViewport()
	case "enter", " ":
		return m.handleEnter()
	}
	return m, nil
}

// handleEnter acts contextually: toggle or save
func (m *VCPUPinningFormModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.focusPos {
	case vcpuPinningToggle:
		m.pinning.Enabled = !m.pinning.Enabled
		m.syncViewport()
		return m, nil
	case vcpuPinningSave:
		return m.validateAndSave()
	}
	return m, nil
}

// moveFocus moves the focus by delta positions
func (m *VCPUPinningFormModel) moveFocus(delta int) {
	if m.focusPos == vcpuPinningToggle && delta > 0 {
		m.focusPos = vcpuPinningSave
	} else if m.focusPos == vcpuPinningSave && delta < 0 {
		m.focusPos = vcpuPinningToggle
	}
}