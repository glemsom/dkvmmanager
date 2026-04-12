// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// pciFocusKind describes what a PCI passthrough focus position represents
type pciFocusKind int

const (
	pciToggle  pciFocusKind = iota // Toggle device selection
	pciROMPath                     // ROM path text input (only when selected)
	pciSave                        // Save button
)

// pciFocusPos is one navigable position in the PCI passthrough form
type pciFocusPos struct {
	kind       pciFocusKind
	deviceAddr string // PCI address (e.g., "0000:01:00.0")
}

// PCIPassthroughFormModel is a scrollable form for editing PCI passthrough config
type PCIPassthroughFormModel struct {
	vmManager *vm.Manager
	devices   []models.PCIDevice          // All scanned devices
	config    models.PCIPassthroughConfig // Current config (selected devices)
	selected  map[string]bool             // Quick lookup: address -> selected
	romPaths  map[string]string           // Per-device ROM path

	// Flat list of focusable positions
	positions  []pciFocusPos
	focusIndex int

	// Per-field text cursor
	cursorOffsets map[string]int

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

// NewPCIPassthroughFormModel creates a new PCI passthrough form model
func NewPCIPassthroughFormModel(vmManager *vm.Manager) (*PCIPassthroughFormModel, error) {
	// Scan devices
	scanner := vm.NewPCIScanner()
	allDevices, scanErr := scanner.ScanDevices()

	// Load existing config
	cfg, _ := vmManager.GetPCIPassthroughConfig()

	// Build lookup maps
	selected := make(map[string]bool)
	romPaths := make(map[string]string)
	for _, dev := range cfg.Devices {
		selected[dev.Address] = true
		if dev.ROMPath != "" {
			romPaths[dev.Address] = dev.ROMPath
		}
	}

	m := &PCIPassthroughFormModel{
		vmManager:     vmManager,
		devices:       allDevices,
		config:        cfg,
		selected:      selected,
		romPaths:      romPaths,
		cursorOffsets: make(map[string]int),
		errors:        make(map[string]string),
		scanErr:       scanErr,
	}

	m.rebuildPositions()
	return m, nil
}

// Init implements tea.Model
func (m *PCIPassthroughFormModel) Init() tea.Cmd {
	return nil
}
