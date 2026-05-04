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

// PCIPassthroughFormModel is a scrollable form for editing PCI passthrough config.
// It implements the form.FormModel interface for use with ScrollableForm.
type PCIPassthroughFormModel struct {
	vmManager *vm.Manager
	devices   []models.PCIDevice          // All scanned devices
	config    models.PCIPassthroughConfig // Current config (selected devices)
	selected  map[string]bool             // Quick lookup: address -> selected

	// IOMMU group index: group number -> list of device pointers
	// Group -1 represents ungrouped devices
	iommuGroups map[int][]*models.PCIDevice

	// Focus state
	positions  []form.FocusPos
	focusIndex int

	// Per-field inline error messages
	errors map[string]string

	// Validation warnings (non-fatal, displayed after save)
	warnings []string

	// Scan state
	scanErr error

	// Kernel apply status message (shown after applying)
	kernelMsg     string
	kernelMsgErr  bool

	// Size (for viewport sync, used by framework's SetSize)
	contentW int
	contentH int
	vp       viewport.Model
	ready    bool

	// Rendering cache (for backward-compatible View)
	renderedLines []string
}

// NewPCIPassthroughFormModel creates a new PCI passthrough form model
func NewPCIPassthroughFormModel(vmManager *vm.Manager) (*PCIPassthroughFormModel, error) {
	// Scan devices
	scanner := vm.NewPCIScanner()
	allDevices, scanErr := scanner.ScanDevices()

	// Load existing config
	cfg, _ := vmManager.GetPCIPassthroughConfig()

	// Build lookup maps
	selected := make(map[string]bool)
	for _, dev := range cfg.Devices {
		selected[dev.Address] = true
	}

	m := &PCIPassthroughFormModel{
		vmManager:   vmManager,
		devices:     allDevices,
		config:      cfg,
		selected:    selected,
		errors:      make(map[string]string),
		scanErr:     scanErr,
	}

	m.buildIOMMUGroups()
	m.positions = m.BuildPositions()
	return m, nil
}

// buildIOMMUGroups indexes devices by IOMMU group number.
// Devices with IOMMUGroup < 0 are placed in the -1 (ungrouped) bucket.
func (m *PCIPassthroughFormModel) buildIOMMUGroups() {
	m.iommuGroups = make(map[int][]*models.PCIDevice)
	for i := range m.devices {
		dev := &m.devices[i]
		group := dev.IOMMUGroup
		if group < 0 {
			group = -1
		}
		m.iommuGroups[group] = append(m.iommuGroups[group], dev)
	}
}

// --- FormModel Interface Implementation ---

// CurrentIndex returns the index of the currently focused position.
func (m *PCIPassthroughFormModel) CurrentIndex() int {
	return m.focusIndex
}

// SetFocusIndex sets the focused position index.
func (m *PCIPassthroughFormModel) SetFocusIndex(i int) {
	m.focusIndex = i
}

// OnEnter is called when the form becomes active.
func (m *PCIPassthroughFormModel) OnEnter() {}

// OnExit is called when the form is dismissed.
func (m *PCIPassthroughFormModel) OnExit() {}

// SetSize updates the form dimensions.
func (m *PCIPassthroughFormModel) SetSize(w, h int) {
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
func (m *PCIPassthroughFormModel) SetFocused(bool) {}

// --- Backward-compatible Init/Update/View ---

// Init implements tea.Model (for backward compatibility).
func (m *PCIPassthroughFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model (for backward compatibility).
func (m *PCIPassthroughFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case PCIVFIOKernelAppliedMsg:
		if msg.Success {
			m.kernelMsg = "vfio-pci.ids applied to grub.cfg successfully"
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
func (m *PCIPassthroughFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// handleKey processes keyboard input (backward-compatible Update path).
func (m *PCIPassthroughFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
func (m *PCIPassthroughFormModel) handleEnterOrApply() (tea.Model, tea.Cmd) {
	pos := m.positions[m.focusIndex]
	switch pos.Kind {
	case form.FocusToggle:
		d := pos.Data.(pciFocusData)
		m.toggleDevice(d.Address)
		m.positions = m.BuildPositions()
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

// syncViewport regenerates the rendered lines and syncs the viewport.
func (m *PCIPassthroughFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	totalContent := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(totalContent)
	if m.focusedLineIndex() >= 0 {
		m.vp.YOffset = form.ClampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height)
	}
}

// focusedLineIndex maps focusIndex to a rendered line index.
func (m *PCIPassthroughFormModel) focusedLineIndex() int {
	line := 0

	// Header + blank = 2 lines
	line += 2

	// Scan error warning if any
	if m.scanErr != nil {
		line += 2
	}

	for i, p := range m.positions {
		if i == m.focusIndex {
			return line
		}

		switch p.Kind {
		case form.FocusHeader:
			line++
		case form.FocusToggle:
			line++
		case form.FocusButton:
			if p.Key == "save" {
				line++ // blank before button
				line++ // button
			} else {
				line++ // button
			}
		}
	}

	return line
}

// renderAllLines produces the full list of output lines for the form (backward compat).
func (m *PCIPassthroughFormModel) renderAllLines() []string {
	var lines []string

	if m.scanErr != nil {
		lines = append(lines, pciErrorStyle.Render(fmt.Sprintf("Warning: PCI scan failed: %s", m.scanErr)))
		lines = append(lines, pciWarnStyle.Render("Devices cannot be configured without a scan."))
		lines = append(lines, "")
	}

	if len(m.devices) == 0 {
		lines = append(lines, pciMutedStyle.Render("No PCI devices found on this system."))
		lines = append(lines, "")
		saveText := pciMutedStyle.Render("[Space/Enter] Save") + "    " + pciMutedStyle.Render("[ESC] Cancel")
		lines = append(lines, saveText)
		lines = append(lines, "")
		applyText := pciMutedStyle.Render("[Space/Enter] Apply to Kernel")
		lines = append(lines, applyText)
		return lines
	}

	// Render each position
	for i, pos := range m.positions {
		focused := (i == m.focusIndex)
		lines = m.renderPosition(lines, pos, focused)
	}

	// Save error at the bottom
	if errMsg, ok := m.errors["save"]; ok {
		lines = append(lines, "")
		lines = append(lines, pciErrorStyle.Render("Error: "+errMsg))
	}

	// Kernel apply status message
	if m.kernelMsg != "" {
		lines = append(lines, "")
		if m.kernelMsgErr {
			lines = append(lines, pciErrorStyle.Render("Error: "+m.kernelMsg))
		} else {
			lines = append(lines, pciSaveStyle.Render(m.kernelMsg))
		}
	}

	// Validation warnings
	for _, w := range m.warnings {
		lines = append(lines, "")
		lines = append(lines, pciWarnStyle.Render("Warning: "+w))
	}

	// Footer
	lines = append(lines, "")
	lines = append(lines, pciMutedStyle.Render("Tab Navigate  PgUp/PgDown Scroll  Space/Enter Toggle/Action  ESC Cancel"))

	return lines
}

// renderPosition appends lines for one focus position (backward compat).
func (m *PCIPassthroughFormModel) renderPosition(lines []string, pos form.FocusPos, focused bool) []string {
	switch pos.Kind {
	case form.FocusHeader:
		lines = append(lines, m.renderGroupHeader(pos))
		return lines

	case form.FocusToggle:
		d := pos.Data.(pciFocusData)
		dev := m.getDeviceByAddr(d.Address)
		if dev == nil {
			lines = append(lines, "  ???")
			return lines
		}
		selected := m.selected[d.Address]
		lines = append(lines, m.renderDeviceToggle(dev, selected, focused))
		return lines

	case form.FocusButton:
		if pos.Key == "save" {
			lines = append(lines, "")
			saveText := pciMutedStyle.Render("[Space/Enter] Save") + "    " + pciMutedStyle.Render("[ESC] Cancel")
			if focused {
				saveText = pciSaveStyle.Render("[Space/Enter] Save") + "    " + pciMutedStyle.Render("[ESC] Cancel")
			}
			lines = append(lines, saveText)
			return lines
		}
		if pos.Key == "apply_kernel" {
			lines = append(lines, "")
			applyText := pciMutedStyle.Render("[Space/Enter] Apply to Kernel") + "    " + pciMutedStyle.Render("[ESC] Cancel")
			if focused {
				applyText = pciApplyStyle.Render("[Space/Enter] Apply to Kernel") + "    " + pciMutedStyle.Render("[ESC] Cancel")
			}
			lines = append(lines, applyText)
			return lines
		}
	}

	return lines
}

// validateAndSave persists the PCI passthrough config (backward compat, returns tea.Model).
func (m *PCIPassthroughFormModel) validateAndSave() (tea.Model, tea.Cmd) {
	result, cmd := m.validateAndSaveCmd()
	if result == form.ResultSave {
		return m, cmd
	}
	return m, nil
}

// handleApplyKernel applies the current PCI passthrough config to grub.cfg (backward compat).
func (m *PCIPassthroughFormModel) handleApplyKernel() (tea.Model, tea.Cmd) {
	cmd := m.handleApplyKernelCmd()
	return m, cmd
}
