// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// usbFocusKind describes what a USB passthrough focus position represents
type usbFocusKind int

const (
	usbToggle usbFocusKind = iota // Toggle device selection
	usbSave                       // Save button
)

// usbFocusPos is one navigable position in the USB passthrough form
type usbFocusPos struct {
	kind     usbFocusKind
	deviceID string // Vendor:Product key (e.g., "046d:c52b")
}

// deviceKey returns a unique key for a USB device based on vendor:product
func usbDeviceKey(vendor, product string) string {
	return vendor + ":" + product
}

// USBPassthroughFormModel is a scrollable form for editing USB passthrough config
type USBPassthroughFormModel struct {
	vmManager *vm.Manager
	devices   []models.USBDevice          // All scanned devices
	config    models.USBPassthroughConfig // Current config (selected devices)
	selected  map[string]bool             // Quick lookup: vendor:product -> selected

	// Flat list of focusable positions
	positions  []usbFocusPos
	focusIndex int

	// Per-field inline error messages
	errors map[string]string

	// Validation warnings (non-fatal, displayed after save)
	warnings []string

	// Scrollable viewport
	vp       viewport.Model
	ready    bool
	contentW int
	contentH int

	// Rendering cache
	renderedLines []string

	// Scan state
	scanErr error
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

	m.rebuildPositions()
	return m, nil
}

// rebuildPositions reconstructs the flat focus list from scanned devices
func (m *USBPassthroughFormModel) rebuildPositions() {
	m.positions = nil

	for _, dev := range m.devices {
		m.positions = append(m.positions, usbFocusPos{
			kind:     usbToggle,
			deviceID: usbDeviceKey(dev.Vendor, dev.Product),
		})
	}

	// Save button
	m.positions = append(m.positions, usbFocusPos{
		kind:     usbSave,
		deviceID: "",
	})
}

// currentPos returns the focus position at the current focusIndex
func (m *USBPassthroughFormModel) currentPos() usbFocusPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		return usbFocusPos{kind: usbSave}
	}
	return m.positions[m.focusIndex]
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

// Init implements tea.Model
func (m *USBPassthroughFormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *USBPassthroughFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

// handleKey processes keyboard input
func (m *USBPassthroughFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "tab":
		m.moveFocus(1)
		m.syncViewport()
	case "shift+tab":
		m.moveFocus(-1)
		m.syncViewport()
	case "up":
		m.moveFocus(-1)
		m.syncViewport()
	case "down":
		m.moveFocus(1)
		m.syncViewport()
	case "enter", " ":
		return m.handleEnter()
	}
	return m, nil
}

// handleEnter acts contextually: toggle or save
func (m *USBPassthroughFormModel) handleEnter() (tea.Model, tea.Cmd) {
	pos := m.currentPos()
	switch pos.kind {
	case usbToggle:
		m.toggleDevice(pos.deviceID)
		m.rebuildPositions()
		m.syncViewport()
		return m, nil
	case usbSave:
		return m.validateAndSave()
	}
	return m, nil
}

// toggleDevice toggles selection of a USB device
func (m *USBPassthroughFormModel) toggleDevice(id string) {
	if m.selected[id] {
		delete(m.selected, id)
	} else {
		m.selected[id] = true
	}
}

// moveFocus moves focus by delta in the flat positions list
func (m *USBPassthroughFormModel) moveFocus(delta int) {
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}

// validateAndSave persists the USB passthrough config
func (m *USBPassthroughFormModel) validateAndSave() (tea.Model, tea.Cmd) {
	m.errors = make(map[string]string)
	m.warnings = nil

	// Build config from selected devices
	var devices []models.USBPassthroughDevice
	for _, dev := range m.devices {
		key := usbDeviceKey(dev.Vendor, dev.Product)
		if !m.selected[key] {
			continue
		}
		devices = append(devices, models.USBPassthroughDevice{
			Vendor:  dev.Vendor,
			Product: dev.Product,
			Name:    dev.Name,
			BusID:   dev.ID,
		})
	}

	// Validate before saving
	warnings, valErrors := vm.ValidateUSBDevices(devices)
	if len(valErrors) > 0 {
		m.errors["save"] = strings.Join(valErrors, "; ")
		m.syncViewport()
		return m, nil
	}

	cfg := models.USBPassthroughConfig{
		Devices: devices,
	}

	if err := m.vmManager.SaveUSBPassthroughConfig(cfg); err != nil {
		m.errors["save"] = fmt.Sprintf("Failed to save: %v", err)
		m.syncViewport()
		return m, nil
	}

	// Store warnings to display after successful save
	m.warnings = warnings

	return m, func() tea.Msg {
		return USBPassthroughUpdatedMsg{}
	}
}

// syncViewport regenerates the rendered lines and syncs the viewport
func (m *USBPassthroughFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	totalContent := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(totalContent)
	if m.focusedLineIndex() >= 0 {
		m.vp.SetYOffset(clampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height))
	}
}

// focusedLineIndex maps focusIndex to a rendered line index
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

		switch p.kind {
		case usbToggle:
			line++
		case usbSave:
			line++ // blank before button
			line++ // button
		}
	}

	return line
}

// View implements tea.Model
func (m *USBPassthroughFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

// SetSize updates the form dimensions
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
	m.syncViewport()
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

// renderAllLines produces the full list of output lines for the form
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
		if m.currentPos().kind == usbSave && m.focusIndex == len(m.positions)-1 {
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

// renderPosition appends lines for one focus position
func (m *USBPassthroughFormModel) renderPosition(lines []string, pos usbFocusPos, focused bool) []string {
	switch pos.kind {
	case usbToggle:
		dev := m.getDeviceByID(pos.deviceID)
		if dev == nil {
			lines = append(lines, "  ???")
			return lines
		}
		selected := m.selected[pos.deviceID]
		lines = append(lines, m.renderDeviceToggle(dev, selected, focused))
		return lines

	case usbSave:
		lines = append(lines, "")
		saveText := usbMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
		if focused {
			saveText = usbSaveStyle.Render("[Space/Enter] Save") + "    " + usbMutedStyle.Render("[ESC] Cancel")
		}
		lines = append(lines, saveText)
		return lines
	}

	return lines
}

// renderDeviceToggle renders a USB device as a toggle line
func (m *USBPassthroughFormModel) renderDeviceToggle(dev *models.USBDevice, selected, focused bool) string {
	prefix := "  "
	if focused {
		prefix = usbFocusStyle.Render("> ")
	}

	// Toggle indicator
	var togglePart string
	if selected {
		if focused {
			togglePart = usbFocusStyle.Render("[X]")
		} else {
			togglePart = usbInputStyle.Render("[X]")
		}
	} else {
		if focused {
			togglePart = usbFocusStyle.Render("[ ]")
		} else {
			togglePart = usbMutedStyle.Render("[ ]")
		}
	}

	// Device name and IDs
	nameStr := usbLabelStyle.Render(dev.Name)
	idStr := usbMutedStyle.Render(fmt.Sprintf(" [%s:%s]", dev.Vendor, dev.Product))
	busStr := ""
	if dev.ID != "" {
		busStr = usbMutedStyle.Render(fmt.Sprintf(" (Bus %s)", dev.ID))
	}

	return prefix + togglePart + " " + nameStr + idStr + busStr
}
