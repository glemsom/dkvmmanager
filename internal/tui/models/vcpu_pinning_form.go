// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VCPUCPUKernelAppliedMsg is sent when CPU kernel parameters have been applied to grub.cfg
type VCPUCPUKernelAppliedMsg struct {
	Success bool
	Error   string
}

// VCPUPinningFormModel is a scrollable form for editing global vCPU pinning.
// It implements the form.FormModel interface for use with ScrollableForm.
type VCPUPinningFormModel struct {
	vmManager *vm.Manager

	// Host topology data (read-only, for display)
	hostTopo models.HostCPUTopology

	// Global CPU topology (read-only, for computing pinning)
	topology models.CPUTopology

	// Global configuration
	pinning models.VCPUPinningGlobal

	// Focus state
	positions  []form.FocusPos
	focusIndex int

	// Per-field inline error messages
	errors map[string]string

	// Scan error
	scanErr error

	// Kernel apply status message (shown after applying)
	kernelMsg    string
	kernelMsgErr bool

	// Viewport fields (for backward-compatible View)
	contentW int
	contentH int
	vp       viewport.Model
	ready    bool

	// Rendering cache (for backward-compatible View)
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
		vmManager: vmManager,
		hostTopo:  hostTopo,
		topology:  topology,
		pinning:   pinning,
		errors:    make(map[string]string),
		scanErr:   scanErr,
	}

	m.positions = m.BuildPositions()
	return m, nil
}

// --- FormModel Interface Implementation ---

// BuildPositions returns the list of navigable positions.
func (m *VCPUPinningFormModel) BuildPositions() []form.FocusPos {
	return []form.FocusPos{
		{Kind: form.FocusToggle, Label: "Pinning Enabled", Key: "enabled"},
		{Kind: form.FocusButton, Label: "Save", Key: "save"},
		{Kind: form.FocusButton, Label: "Apply to Kernel", Key: "apply_kernel"},
	}
}

// CurrentIndex returns the index of the currently focused position.
func (m *VCPUPinningFormModel) CurrentIndex() int {
	return m.focusIndex
}

// SetFocusIndex sets the focused position index.
func (m *VCPUPinningFormModel) SetFocusIndex(i int) {
	m.focusIndex = i
}

// OnEnter is called when the form becomes active.
func (m *VCPUPinningFormModel) OnEnter() {}

// OnExit is called when the form is dismissed.
func (m *VCPUPinningFormModel) OnExit() {}

// SetSize updates the form dimensions.
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
}

// SetFocused sets whether the form has keyboard focus.
func (m *VCPUPinningFormModel) SetFocused(bool) {}

// --- Backward-compatible Init/Update/View ---

// Init implements tea.Model (for backward compatibility).
func (m *VCPUPinningFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model (for backward compatibility).
func (m *VCPUPinningFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		m.syncViewport()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		vp, _ := m.vp.Update(msg)
		m.vp = vp
		return m, nil

	case VCPUCPUKernelAppliedMsg:
		if msg.Success {
			if m.pinning.Enabled && len(m.pinning.Mappings) > 0 {
				m.kernelMsg = "Kernel CPU isolation parameters applied to grub.cfg"
			} else {
				m.kernelMsg = "Kernel CPU isolation parameters removed from grub.cfg"
			}
			m.kernelMsgErr = false
		} else {
			m.kernelMsg = msg.Error
			m.kernelMsgErr = true
		}
		m.syncViewport()
		return m, nil
	}
	return m, nil
}

// View implements tea.Model (for backward compatibility).
func (m *VCPUPinningFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// handleKey processes keyboard input (backward-compatible Update path).
func (m *VCPUPinningFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "tab", "down":
		m.moveFocus(1)
		m.syncViewport()
	case "shift+tab", "up":
		m.moveFocus(-1)
		m.syncViewport()
	case "pgup":
		m.vp.HalfPageUp()
		return m, nil
	case "pgdown":
		m.vp.HalfPageDown()
		return m, nil
	case "enter", " ":
		return m.handleEnterOrApply()
	}
	return m, nil
}

// handleEnterOrApply acts contextually: toggle, save, or apply kernel (backward compat).
func (m *VCPUPinningFormModel) handleEnterOrApply() (tea.Model, tea.Cmd) {
	pos := m.positions[m.focusIndex]
	switch pos.Kind {
	case form.FocusToggle:
		m.pinning.Enabled = !m.pinning.Enabled
		m.syncViewport()
		return m, nil
	case form.FocusButton:
		if pos.Key == "save" {
			return m.validateAndSave()
		}
		if pos.Key == "apply_kernel" {
			return m.handleApplyKernel()
		}
	}
	// Skip non-focusable positions
	m.moveFocus(1)
	m.syncViewport()
	return m, nil
}

// moveFocus moves focus by delta in the flat positions list.
func (m *VCPUPinningFormModel) moveFocus(delta int) {
	if delta == 0 {
		return
	}
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// validateAndSave persists the vCPU pinning config (backward compat, returns tea.Model).
func (m *VCPUPinningFormModel) validateAndSave() (tea.Model, tea.Cmd) {
	result, cmd := m.validateAndSaveCmd()
	if result == form.ResultSave {
		return m, cmd
	}
	return m, nil
}

// handleApplyKernel applies the current vCPU pinning config to grub.cfg (backward compat).
func (m *VCPUPinningFormModel) handleApplyKernel() (tea.Model, tea.Cmd) {
	cmd := m.handleApplyKernelCmd()
	return m, cmd
}
