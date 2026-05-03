// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// cpuTopoFocusData carries per-core data through the form framework.
type cpuTopoFocusData struct {
	dieID   int
	coreID  int
	dieLabel string
	coreInfo *models.CPUCore
}

// CPUTopologyFormModel is a scrollable form for editing global CPU topology.
// It implements the form.FormModel interface for use with ScrollableForm.
type CPUTopologyFormModel struct {
	vmManager *vm.Manager

	// Host topology data
	hostTopo models.HostCPUTopology

	// Global configuration
	topology models.CPUTopology

	// Quick lookup: coreKey (dieID:coreID) -> selected for VM
	coreSelected map[string]bool

	// Focus state
	positions  []form.FocusPos
	focusIndex int

	// Per-field inline error messages
	errors map[string]string

	// Scan error
	scanErr error

	// Size (for viewport sync, used by framework's SetSize)
	contentW int
	contentH int
	vp       viewport.Model
	ready    bool

	// Rendering cache (for backward-compatible View)
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

	m.positions = m.BuildPositions()
	return m, nil
}

// --- Backward-compatible position access ---

// cpuTopoPos is a legacy position type for backward-compatible test access.
type cpuTopoPos struct {
	kind     int
	dieID    int
	coreID   int
	dieLabel string
	coreInfo *models.CPUCore
}

// Constants for backward compatibility with existing tests.
const (
	cpuTopoToggle = int(form.FocusToggle) // toggle position kind
	cpuTopoSave   = int(form.FocusButton) // save button position kind
)

// currentPos returns the focus position at the current focusIndex (for backward compatibility).
func (m *CPUTopologyFormModel) currentPos() cpuTopoPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return cpuTopoPos{kind: cpuTopoSave}
	}
	p := m.positions[m.focusIndex]
	d := p.Data.(cpuTopoFocusData)
	return cpuTopoPos{
		kind:     int(p.Kind),
		dieID:    d.dieID,
		coreID:   d.coreID,
		dieLabel: d.dieLabel,
		coreInfo: d.coreInfo,
	}
}

// --- FormModel Interface Implementation ---

// BuildPositions returns the current list of navigable positions.
func (m *CPUTopologyFormModel) BuildPositions() []form.FocusPos {
	var positions []form.FocusPos

	if m.scanErr != nil || len(m.hostTopo.Dies) == 0 {
		positions = append(positions, form.FocusPos{
			Kind: form.FocusButton, Label: "Save", Key: "save",
			Data: cpuTopoFocusData{},
		})
		return positions
	}

	for _, die := range m.hostTopo.Dies {
		dieLabel := fmt.Sprintf("Die %d", die.ID)
		if die.L3CacheKB > 0 {
			dieLabel += fmt.Sprintf(" — L3 Cache: %s", formatCacheSize(die.L3CacheKB))
		}

		for _, core := range die.CoreDetails {
			positions = append(positions, form.FocusPos{
				Kind:  form.FocusToggle,
				Label: fmt.Sprintf("Core %d", core.ID),
				Key:   coreKey(die.ID, core.ID),
				Data: cpuTopoFocusData{
					dieID:    die.ID,
					coreID:   core.ID,
					dieLabel: dieLabel,
					coreInfo: &core,
				},
			})
		}
	}

	// Save button
	positions = append(positions, form.FocusPos{
		Kind: form.FocusButton, Label: "Save", Key: "save",
		Data: cpuTopoFocusData{},
	})

	return positions
}

// CurrentIndex returns the index of the currently focused position.
func (m *CPUTopologyFormModel) CurrentIndex() int {
	return m.focusIndex
}

// SetFocusIndex sets the focused position index.
func (m *CPUTopologyFormModel) SetFocusIndex(i int) {
	m.focusIndex = i
}

// RenderHeader returns the form header.
func (m *CPUTopologyFormModel) RenderHeader() string {
	if m.scanErr != nil || len(m.hostTopo.Dies) == 0 {
		return cpuTopoFocusStyle.Render("CPU Topology")
	}
	return cpuTopoFocusStyle.Render("CPU Topology") + "\n" +
		cpuTopoLabelStyle.Render(fmt.Sprintf("Host: %d dies, %d cores, %d threads",
			len(m.hostTopo.Dies), m.hostTopo.TotalCores, m.hostTopo.TotalCPUs))
}

// RenderFooter returns the form footer.
func (m *CPUTopologyFormModel) RenderFooter() string {
	var parts []string

	// Summary line (for save position rendering context)
	// This is handled in RenderPosition for the save button

	if errMsg, ok := m.errors["save"]; ok {
		parts = append(parts, "")
		parts = append(parts, cpuTopoErrorStyle.Render("Error: "+errMsg))
	}

	parts = append(parts, "")
	parts = append(parts, cpuTopoMutedStyle.Render("↑/↓ Navigate  Space Toggle  ESC Cancel"))
	return strings.Join(parts, "\n")
}

// RenderPosition returns the markup for a single position.
func (m *CPUTopologyFormModel) RenderPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
	switch pos.Kind {
	case form.FocusToggle:
		d := pos.Data.(cpuTopoFocusData)
		key := coreKey(d.dieID, d.coreID)
		selected := m.coreSelected[key]
		return m.renderCoreLine(d.coreInfo, selected, focused)

	case form.FocusButton:
		if pos.Key == "save" {
			// Summary + save button
			allocatedCores := 0
			for _, p := range m.positions {
				if p.Kind == form.FocusToggle {
					d := p.Data.(cpuTopoFocusData)
					key := coreKey(d.dieID, d.coreID)
					if m.coreSelected[key] {
						allocatedCores++
					}
				}
			}
			hostCores := m.hostTopo.TotalCores - allocatedCores

			var lines []string
			lines = append(lines, "")
			lines = append(lines, cpuTopoLabelStyle.Render(fmt.Sprintf(
				"Summary: %d cores for VMs, %d for host",
				allocatedCores, hostCores)))
			if hostCores == 0 {
				lines = append(lines, cpuTopoErrorStyle.Render("Warning: No cores reserved for host — system may become unresponsive"))
			}
			lines = append(lines, "")
			saveText := cpuTopoMutedStyle.Render("[Enter] Save    [ESC] Cancel")
			if focused {
				saveText = cpuTopoSaveStyle.Render("[Space/Enter] Save") + "    " + cpuTopoMutedStyle.Render("[ESC] Cancel")
			}
			lines = append(lines, saveText)
			return strings.Join(lines, "\n")
		}
	}
	return ""
}

// HandleEnter is called when the user presses Enter on a position.
func (m *CPUTopologyFormModel) HandleEnter(pos form.FocusPos) (form.FormResult, tea.Cmd) {
	switch pos.Kind {
	case form.FocusToggle:
		d := pos.Data.(cpuTopoFocusData)
		m.toggleCore(d.dieID, d.coreID)
		return form.ResultNone, nil
	case form.FocusButton:
		if pos.Key == "save" {
			return m.validateAndSaveCmd()
		}
		return form.ResultNone, nil
	default:
		m.focusIndex++
		if m.focusIndex >= len(m.positions) {
			m.focusIndex = len(m.positions) - 1
		}
		return form.ResultNone, nil
	}
}

// HandleChar is a no-op (no text fields in CPU topology form).
func (m *CPUTopologyFormModel) HandleChar(pos form.FocusPos, ch string) {}

// HandleBackspace is a no-op (no text fields in CPU topology form).
func (m *CPUTopologyFormModel) HandleBackspace(pos form.FocusPos) {}

// HandleDelete is a no-op (no text fields in CPU topology form).
func (m *CPUTopologyFormModel) HandleDelete(pos form.FocusPos) {}

// OnEnter is called when the form becomes active.
func (m *CPUTopologyFormModel) OnEnter() {}

// OnExit is called when the form is dismissed.
func (m *CPUTopologyFormModel) OnExit() {}

// SetSize updates the form dimensions.
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
}

// SetFocused sets whether the form has keyboard focus.
func (m *CPUTopologyFormModel) SetFocused(bool) {}

// --- Backward-compatible Init/Update/View ---

// Init implements tea.Model (for backward compatibility).
func (m *CPUTopologyFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model (for backward compatibility).
func (m *CPUTopologyFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// View implements tea.Model (for backward compatibility).
func (m *CPUTopologyFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	m.renderedLines = m.renderAllLines()
	totalContent := ""
	for i, line := range m.renderedLines {
		if i > 0 {
			totalContent += "\n"
		}
		totalContent += line
	}
	m.vp.SetContent(totalContent)
	return m.vp.View()
}

// handleKey processes keyboard input (backward-compatible Update path).
func (m *CPUTopologyFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "tab", "down":
		m.moveFocus(1)
	case "shift+tab", "up":
		m.moveFocus(-1)
	case "enter", " ":
		return m.handleEnterKey()
	}
	return m, nil
}

// handleEnterKey acts contextually: toggle or save (backward compat).
func (m *CPUTopologyFormModel) handleEnterKey() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	switch pos.kind {
	case cpuTopoToggle:
		m.toggleCore(pos.dieID, pos.coreID)
		return m, nil
	case cpuTopoSave:
		_, cmd := m.validateAndSaveCmd()
		return m, cmd
	default:
		m.moveFocus(1)
		return m, nil
	}
}

// handleSpace toggles the focused core (backward compat).
func (m *CPUTopologyFormModel) handleSpace() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	if pos.kind == cpuTopoToggle {
		m.toggleCore(pos.dieID, pos.coreID)
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
