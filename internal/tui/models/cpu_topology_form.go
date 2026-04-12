// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// cpuTopoFocusKind describes what a CPU topology focus position represents
type cpuTopoFocusKind int

const (
	cpuTopoToggle cpuTopoFocusKind = iota // Toggle physical core assignment
	cpuTopoSave                           // Save button
)

// cpuTopoFocusPos is one navigable position in the CPU topology form
type cpuTopoFocusPos struct {
	kind     cpuTopoFocusKind
	dieID    int
	coreID   int    // Physical core ID (for toggles)
	dieLabel string // Die label for rendering
	coreInfo *models.CPUCore
}

// CPUTopologyFormModel is a scrollable form for editing global CPU topology
type CPUTopologyFormModel struct {
	vmManager *vm.Manager

	// Host topology data
	hostTopo models.HostCPUTopology

	// Global configuration
	topology models.CPUTopology

	// Quick lookup: coreKey (dieID:coreID) -> selected for VM
	coreSelected map[string]bool

	// Flat list of focusable positions
	positions  []cpuTopoFocusPos
	focusIndex int

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

// coreKey returns a unique key for a physical core
func coreKey(dieID, coreID int) string {
	return fmt.Sprintf("%d:%d", dieID, coreID)
}

// NewCPUTopologyFormModel creates a new CPU topology form model
func NewCPUTopologyFormModel(vmManager *vm.Manager) (*CPUTopologyFormModel, error) {
	// Scan host topology
	scanner := vm.NewCPUScanner()
	hostTopo, scanErr := scanner.ScanTopology()

	// Load global CPU topology config
	topology, err := vmManager.GetCPUTopology()
	if err != nil {
		return nil, fmt.Errorf("failed to load CPU topology: %w", err)
	}
	// Build core selected lookup from saved config
	coreSelected := make(map[string]bool)
	if topology.Enabled && len(topology.SelectedCPUs) > 0 {
		selectedSet := make(map[int]bool)
		for _, cpuID := range topology.SelectedCPUs {
			selectedSet[cpuID] = true
		}
		// Map from logical CPU IDs back to core selections
		if scanErr == nil {
			for _, die := range hostTopo.Dies {
				for _, core := range die.CoreDetails {
					allThreadsSelected := true
					for _, t := range core.Threads {
						if !selectedSet[t] {
							allThreadsSelected = false
							break
						}
					}
					if allThreadsSelected && len(core.Threads) > 0 {
						coreSelected[coreKey(die.ID, core.ID)] = true
					}
				}
			}
		}
	} else {
		// Default: first core as HOST, rest as VM
		if scanErr == nil && hostTopo.TotalCores > 1 {
			first := true
			for _, die := range hostTopo.Dies {
				for _, core := range die.CoreDetails {
					if first {
						first = false
						continue
					}
					coreSelected[coreKey(die.ID, core.ID)] = true
				}
			}
		}
	}

	m := &CPUTopologyFormModel{
		vmManager:    vmManager,
		hostTopo:     hostTopo,
		topology:     topology,
		coreSelected: coreSelected,
		errors:       make(map[string]string),
		scanErr:      scanErr,
	}

	m.rebuildPositions()
	return m, nil
}

// rebuildPositions reconstructs the flat focus list from host topology
func (m *CPUTopologyFormModel) rebuildPositions() {
	m.positions = nil

	if m.scanErr != nil || len(m.hostTopo.Dies) == 0 {
		m.positions = append(m.positions, cpuTopoFocusPos{
			kind: cpuTopoSave,
		})
		return
	}

	for _, die := range m.hostTopo.Dies {
		dieLabel := fmt.Sprintf("Die %d", die.ID)
		if die.L3CacheKB > 0 {
			dieLabel += fmt.Sprintf(" — L3 Cache: %s", formatCacheSize(die.L3CacheKB))
		}

		for _, core := range die.CoreDetails {
			m.positions = append(m.positions, cpuTopoFocusPos{
				kind:     cpuTopoToggle,
				dieID:    die.ID,
				coreID:   core.ID,
				dieLabel: dieLabel,
				coreInfo: &core,
			})
		}
	}

	// Save button
	m.positions = append(m.positions, cpuTopoFocusPos{
		kind: cpuTopoSave,
	})
}

// currentPos returns the focus position at the current focusIndex
func (m *CPUTopologyFormModel) currentPos() cpuTopoFocusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return cpuTopoFocusPos{kind: cpuTopoSave}
	}
	return m.positions[m.focusIndex]
}

// Init implements tea.Model
func (m *CPUTopologyFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *CPUTopologyFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (m *CPUTopologyFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// SetSize updates the form dimensions
func (m *CPUTopologyFormModel) SetSize(w, h int) {
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
func (m *CPUTopologyFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "tab", "down":
		m.moveFocus(1)
		m.syncViewport()
	case "shift+tab", "up":
		m.moveFocus(-1)
		m.syncViewport()
	case "enter", " ":
		return m.handleEnter()
	}
	return m, nil
}

// handleEnter acts contextually: toggle or save
func (m *CPUTopologyFormModel) handleEnter() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	switch pos.kind {
	case cpuTopoToggle:
		m.toggleCore(pos.dieID, pos.coreID)
		m.syncViewport()
		return m, nil
	case cpuTopoSave:
		return m.validateAndSave()
	default:
		m.moveFocus(1)
		m.syncViewport()
		return m, nil
	}
}

// handleSpace toggles the focused core
func (m *CPUTopologyFormModel) handleSpace() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	if pos.kind == cpuTopoToggle {
		m.toggleCore(pos.dieID, pos.coreID)
		m.syncViewport()
	}
	return m, nil
}

// toggleCore toggles selection of a physical core
func (m *CPUTopologyFormModel) toggleCore(dieID, coreID int) {
	key := coreKey(dieID, coreID)
	if m.coreSelected[key] {
		delete(m.coreSelected, key)
	} else {
		m.coreSelected[key] = true
	}
}

// formatCacheSize formats cache size in KB to a human-readable string
func formatCacheSize(kb int) string {
	if kb >= 1024*1024 {
		return fmt.Sprintf("%.0fG", float64(kb)/(1024*1024))
	}
	if kb >= 1024 {
		return fmt.Sprintf("%dM", kb/1024)
	}
	return fmt.Sprintf("%dK", kb)
}
