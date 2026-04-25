// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// BlockDeviceModel is a model for listing and selecting block devices
type BlockDeviceModel struct {
	// Block devices
	devices []BlockDevice

	// Selected device path
	selectedPath string

	// Selection state
	selectedIndex int

	// Error message
	errorMsg string

	// Whether model is active
	active bool
}

// BlockDevice represents a block device
type BlockDevice struct {
	Path     string
	Name     string
	Size     string
	Type     string
	ReadOnly bool
}

// NewBlockDeviceModel creates a new block device lister model
func NewBlockDeviceModel() *BlockDeviceModel {
	return &BlockDeviceModel{
		selectedIndex: 0,
		active:        true,
	}
}

// Init initializes the model
func (m *BlockDeviceModel) Init() tea.Cmd {
	return m.loadDevices
}

// BlockDeviceLoadedMsg is sent after loadDevices completes to trigger a view refresh
type BlockDeviceLoadedMsg struct{}

// View returns the rendered view for BlockDeviceModel
func (m *BlockDeviceModel) View() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Foreground).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Primary).
		Bold(true)

	deviceStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Foreground)

	sizeStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.ForegroundDim)

	typeStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.ForegroundDim)

	roStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Warning)

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted)

	var output string

	output += headerStyle.Render("Select Block Device") + "\n"
	output += "\n"
	output += "Available block devices:\n"
	output += "\n"

	if len(m.devices) == 0 {
		output += "  (no block devices found)\n"
	} else {
		for i, dev := range m.devices {
			var name string

			if dev.ReadOnly {
				name = deviceStyle.Render(dev.Name) + roStyle.Render(" [RO]")
			} else {
				name = deviceStyle.Render(dev.Name)
			}

			if i == m.selectedIndex {
				output += "> " + selectedStyle.Render(name)
			} else {
				output += "  " + name
			}

			output += "  " + sizeStyle.Render(dev.Size)
			output += "  " + typeStyle.Render(dev.Type)
			output += "\n"
		}
	}

	if m.errorMsg != "" {
		output += "\n" + lipgloss.NewStyle().Foreground(styles.Colors.Error).Render("Error: "+m.errorMsg)
	}

	output += "\n\n"
	output += helpStyle.Render("↑/↓ Navigate  Space/Enter Select  ESC Cancel") + "\n"

	return output
}

// GetSelectedPath returns the selected device path
func (m *BlockDeviceModel) GetSelectedPath() string {
	return m.selectedPath
}

// DiskSourceType represents the type of disk source
type DiskSourceType int

const (
	DiskSourceFile DiskSourceType = iota
	DiskSourceDevice
	DiskSourceLVM
)

// AddDiskModel is the model for adding a disk (file or block device)
type AddDiskModel struct {
	// VM Manager (for generating disk paths)
	vmManager *vm.Manager

	// Current step
	step int

	// Selected disk source type
	sourceType DiskSourceType

	// The selected/entered path
	path string

	// Selected index in menus
	selectedIndex int

	// File browser
	fileBrowser *FileBrowserModel

	// Block device lister
	blockDevice *BlockDeviceModel

	// LVM volume lister
	lvmVolume *LVMVolumeModel

	// For returning
	returnMsg tea.Msg

	// Error message
	errorMsg string

	// Whether active
	active bool
}

// NewAddDiskModel creates a new add disk model
func NewAddDiskModel(vmManager *vm.Manager) *AddDiskModel {
	return &AddDiskModel{
		vmManager:   vmManager,
		step:        0,
		active:      true,
		fileBrowser: NewFileBrowserModel(FileTypeDiskImage),
		blockDevice: NewBlockDeviceModel(),
		lvmVolume:   NewLVMVolumeModel(),
	}
}

// Init initializes the model
func (m *AddDiskModel) Init() tea.Cmd {
	return nil
}

// View returns the rendered view for AddDiskModel
func (m *AddDiskModel) View() string {
	switch m.step {
	case 0:
		return m.renderSourceSelect()
	case 1:
		if m.fileBrowser != nil {
			return m.fileBrowser.View()
		}
	case 2:
		if m.blockDevice != nil {
			return m.blockDevice.View()
		}
	case 3:
		if m.lvmVolume != nil {
			return m.lvmVolume.View()
		}
	}

	// Default
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Foreground).
		Bold(true)
	return headerStyle.Render("Add Disk")
}

func (m *AddDiskModel) renderSourceSelect() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Foreground).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Primary).
		Bold(true)

	var output string
	output += headerStyle.Render("Add Hard Disk") + "\n"
	output += "\n"
	output += "Select source type:\n"
	output += "\n"

	items := []string{"Disk image file", "Block device", "LVM Logical Volume"}
	for i, item := range items {
		if i == m.selectedIndex {
			output += "> " + selectedStyle.Render(item) + "\n"
		} else {
			output += "  " + item + "\n"
		}
	}

	output += "\n"
	output += "Space/Enter Select  ESC Cancel\n"

	return output
}

// DiskAddedMsg is a message indicating a disk was added
type DiskAddedMsg struct {
	Path     string
	Canceled bool
}
