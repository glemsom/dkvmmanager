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
	pciGroupHeader pciFocusKind = iota // IOMMU group header (non-selectable)
	pciToggle                          // Toggle device selection
	pciSave                            // Save button
)

// pciFocusPos is one navigable position in the PCI passthrough form
type pciFocusPos struct {
	kind       pciFocusKind
	deviceAddr string // PCI address (e.g., "0000:01:00.0")
	groupNum   int    // IOMMU group number (used for pciGroupHeader positions)
}

// PCIPassthroughFormModel is a scrollable form for editing PCI passthrough config
type PCIPassthroughFormModel struct {
	vmManager *vm.Manager
	devices   []models.PCIDevice          // All scanned devices
	config    models.PCIPassthroughConfig // Current config (selected devices)
	selected  map[string]bool             // Quick lookup: address -> selected

	// IOMMU group index: group number -> list of device pointers
	// Group -1 represents ungrouped devices
	iommuGroups map[int][]*models.PCIDevice

	// Flat list of focusable positions (excludes group headers)
	positions  []pciFocusPos
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
	m.rebuildPositions()
	return m, nil
}

// Init implements tea.Model
func (m *PCIPassthroughFormModel) Init() tea.Cmd {
	return nil
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
