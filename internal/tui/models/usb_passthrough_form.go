// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// deviceKey returns a unique key for a USB device based on vendor:product
func usbDeviceKey(vendor, product string) string {
	return vendor + ":" + product
}

// USBPassthroughFormModel is a scrollable form for editing USB passthrough config.
// It implements the form.FormModel interface for use with ScrollableForm.
type USBPassthroughFormModel struct {
	vmManager *vm.Manager
	devices   []models.USBDevice          // All scanned devices
	config    models.USBPassthroughConfig // Current config (selected devices)
	selected  map[string]bool             // Quick lookup: vendor:product -> selected

	// Focus state
	positions  []form.FocusPos
	focusIndex int

	// Per-field inline error messages
	errors map[string]string

	// Validation warnings (non-fatal, displayed after save)
	warnings []string

	// Scan state
	scanErr error

	// Size (for viewport sync, used by framework's SetSize)
	contentW int
	contentH int
	vp       viewport.Model
	ready    bool

	// Rendering cache (for backward-compatible View)
	renderedLines []string
}

// NewUSBPassthroughFormModel creates a new USB passthrough form model
func NewUSBPassthroughFormModel(vmManager *vm.Manager) (*USBPassthroughFormModel, error) {
	// Scan devices
	scanner := vm.NewUSBScanner()
	allDevices, scanErr := scanner.ScanDevices()

	// Load existing config
	cfg, _ := vmManager.GetUSBPassthroughConfig()

	// Build lookup map
	selected := make(map[string]bool)
	for _, dev := range cfg.Devices {
		selected[usbDeviceKey(dev.Vendor, dev.Product)] = true
	}

	m := &USBPassthroughFormModel{
		vmManager: vmManager,
		devices:   allDevices,
		config:    cfg,
		selected:  selected,
		errors:    make(map[string]string),
		scanErr:   scanErr,
	}

	m.positions = m.BuildPositions()
	return m, nil
}

// getDeviceByID finds a device by vendor:product key
func (m *USBPassthroughFormModel) getDeviceByID(id string) *models.USBDevice {
	for i := range m.devices {
		if usbDeviceKey(m.devices[i].Vendor, m.devices[i].Product) == id {
			return &m.devices[i]
		}
	}
	return nil
}

// toggleDevice toggles selection of a USB device
func (m *USBPassthroughFormModel) toggleDevice(id string) {
	if m.selected[id] {
		delete(m.selected, id)
	} else {
		m.selected[id] = true
	}
}

// --- FormModel Interface Implementation ---

// CurrentIndex returns the index of the currently focused position.
func (m *USBPassthroughFormModel) CurrentIndex() int {
	return m.focusIndex
}

// SetFocusIndex sets the focused position index.
func (m *USBPassthroughFormModel) SetFocusIndex(i int) {
	m.focusIndex = i
}

// OnEnter is called when the form becomes active.
func (m *USBPassthroughFormModel) OnEnter() {}

// OnExit is called when the form is dismissed.
func (m *USBPassthroughFormModel) OnExit() {}

// SetSize updates the form dimensions.
func (m *USBPassthroughFormModel) SetSize(w, h int) {
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
func (m *USBPassthroughFormModel) SetFocused(bool) {}

// --- Backward-compatible Init/Update/View ---

// Init implements tea.Model (for backward compatibility).
func (m *USBPassthroughFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model (for backward compatibility).
func (m *USBPassthroughFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		m.syncViewport()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// View implements tea.Model (for backward compatibility).
func (m *USBPassthroughFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// handleKey processes keyboard input (backward-compatible Update path).
func (m *USBPassthroughFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "tab", "down":
		m.moveFocus(1)
		m.syncViewport()
	case "shift+tab", "up":
		m.moveFocus(-1)
		m.syncViewport()
	case "enter", " ":
		return m.handleEnterOrSave()
	}
	return m, nil
}

// handleEnterOrSave acts contextually: toggle or save (backward compat).
func (m *USBPassthroughFormModel) handleEnterOrSave() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	switch pos.Kind {
	case form.FocusToggle:
		m.toggleDevice(pos.Key)
		m.positions = m.BuildPositions()
		m.syncViewport()
		return m, nil
	case form.FocusButton:
		if pos.Key == "save" {
			return m.validateAndSave()
		}
	}
	// Move focus forward for unknown positions
	m.moveFocus(1)
	m.syncViewport()
	return m, nil
}

// validateAndSave persists the USB passthrough config (backward compat, returns tea.Model).
func (m *USBPassthroughFormModel) validateAndSave() (tea.Model, tea.Cmd) {
	result, cmd := m.validateAndSaveCmd()
	if result == form.ResultSave {
		return m, cmd
	}
	return m, nil
}

// syncViewport regenerates the rendered lines and syncs the viewport.
func (m *USBPassthroughFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	totalContent := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(totalContent)
	if m.focusedLineIndex() >= 0 {
		m.vp.YOffset = form.ClampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height)
	}
}

// focusedLineIndex maps focusIndex to a rendered line index.
func (m *USBPassthroughFormModel) focusedLineIndex() int {
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
		case form.FocusToggle:
			line++
		case form.FocusButton:
			line++ // blank before button
			line++ // button
		}
	}

	return line
}

// renderAllLines produces the full list of output lines for the form (backward compat).
func (m *USBPassthroughFormModel) renderAllLines() []string {
	var lines []string

	if m.scanErr != nil {
		lines = append(lines, usbErrorStyle.Render(fmt.Sprintf("Warning: USB scan failed: %s", m.scanErr)))
		lines = append(lines, usbWarnStyle.Render("Devices cannot be configured without a scan."))
		lines = append(lines, "")
	}

	if len(m.devices) == 0 {
		lines = append(lines, usbMutedStyle.Render("No USB devices found on this system."))
		lines = append(lines, "")
		saveText := usbMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
		if m.currentPos().Kind == form.FocusButton && m.focusIndex == len(m.positions)-1 {
			saveText = usbSaveStyle.Render("[Space/Enter] Save") + "    " + usbMutedStyle.Render("[ESC] Cancel")
		}
		lines = append(lines, saveText)
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
		lines = append(lines, usbErrorStyle.Render("Error: "+errMsg))
	}

	// Validation warnings
	for _, w := range m.warnings {
		lines = append(lines, "")
		lines = append(lines, usbWarnStyle.Render("Warning: "+w))
	}

	// Footer
	lines = append(lines, "")
	lines = append(lines, usbMutedStyle.Render("Tab Navigate  Space/Enter Toggle  ESC Cancel"))

	return lines
}

// renderPosition appends lines for one focus position (backward compat).
func (m *USBPassthroughFormModel) renderPosition(lines []string, pos form.FocusPos, focused bool) []string {
	switch pos.Kind {
	case form.FocusToggle:
		dev := m.getDeviceByID(pos.Key)
		if dev == nil {
			lines = append(lines, "  ???")
			return lines
		}
		selected := m.selected[pos.Key]
		lines = append(lines, m.renderDeviceToggle(dev, selected, focused))
		return lines

	case form.FocusButton:
		if pos.Key == "save" {
			lines = append(lines, "")
			saveText := usbMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
			if focused {
				saveText = usbSaveStyle.Render("[Space/Enter] Save") + "    " + usbMutedStyle.Render("[ESC] Cancel")
			}
			lines = append(lines, saveText)
			return lines
		}
	}

	return lines
}

// --- Styles ---

var (
	usbLabelStyle = styles.FormLabelStyle()
	usbFocusStyle = styles.FormFocusStyle()
	usbInputStyle = styles.FormInputStyle()
	usbErrorStyle = styles.ErrorTextStyle()
	usbMutedStyle = styles.FormMutedStyle()
	usbSaveStyle  = styles.FormSaveStyle()
	usbWarnStyle  = styles.WarningTextStyle()
)
