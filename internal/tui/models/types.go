// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"charm.land/bubbles/v2/list"
	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/domain"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// View constants
const (
	ViewMainMenu        = "main"
	ViewConfigMenu      = "config"
	ViewVMMenu          = "vm"
	ViewLogViewer       = "log"
	ViewFirstRunSetup   = "setup"
	ViewPowerMenu       = "power"
	ViewVMCreate        = "vm_create"
	ViewVMEdit          = "vm_edit"
	ViewVMSelect        = "vm_select"
	ViewVMDelete        = "vm_delete"
	ViewCPUOptions      = "cpu_options"
	ViewVMRunning       = "vm_running"
	ViewPCIPassthrough  = "pci_passthrough"
	ViewUSBPassthrough  = "usb_passthrough"
	ViewCPUTopology     = "cpu_topology"
	ViewVCPUPinning   = "vcpu_pinning"
	ViewSSHPassword     = "ssh_password"
	ViewStartStopScript = "start_stop_script"
	ViewLVCreate        = "lv_create"
	ViewMountPointWarning = "mount_point_warning"
)

// MainModelConfig holds configuration for creating a MainModel
type MainModelConfig struct {
	Config              *config.Config
	DebugMode           bool
	DryRunMode          bool
	SkipMountPointCheck bool
}

// ViewChangeMsg is sent by sub-models to request a view transition
type ViewChangeMsg struct {
	View string
}

// VMDeletedMsg is sent when a VM is deleted
type VMDeletedMsg struct {
	VMName string
	VMID   string
}

// VMRunningViewExitMsg is sent when user requests to exit the VM running view
type VMRunningViewExitMsg struct{}

// LBUCommitMsg is sent when an lbu commit operation completes
type LBUCommitMsg struct {
	Output  string
	Success bool
}

// RebootMsg is sent when a reboot operation completes
type RebootMsg struct {
	Output  string
	Success bool
}

// PowerOffMsg is sent when a power off operation completes
type PowerOffMsg struct {
	Output  string
	Success bool
}

// MainModel is the main model for the application
type MainModel struct {
	// Current view
	currentView string

	// View registry for sub-view management
	viewRegistry *ViewRegistry

	// Configuration
	cfg *config.Config

	// VM manager
	vmManager *vm.Manager

	// Config repository for persistence operations (CPU options, PCI/USB passthrough, etc.)
	configRepo *vm.Repository

	// Host discovery for hardware scanning
	hostDiscovery vm.HostDiscovery

	// Menu state
	selectedIndex int
	menuItems     []MenuItem
	menuList      list.Model

	// Config menu state
	configSelectedIndex int
	configList          list.Model
	configListView      string

	// Power menu state
	powerList list.Model
	powerListView     string

	// Quitting flag
	quitting bool

	// VM running model (kept for Runner() convenience accessor)
	vmRunningModel *VMRunningModel

	// Running VM ID - tracks the currently running VM to persist across rebuildMenuList calls
	runningVMID string

	// Debug mode flag
	debugMode bool

	// Dry-run mode flag
	dryRunMode bool

	// Skip mount point check flag (for testing)
	skipMountPointCheck bool

	// Tab navigation
	tabModel *components.TabModel

	// Status bar
	statusBar *components.StatusBar

	// Breadcrumbs for sub-navigation
	breadcrumbs *components.Breadcrumbs

	// Terminal dimensions (cached from WindowSizeMsg)
	windowWidth  int
	windowHeight int
}

// MenuItem represents a menu item
type MenuItem struct {
	Title    string
	Type     string // "VM", "INT_CONFIG", "INT_POWEROFF", "INT_SHELL"
	VMID     string // VM ID if type is "VM"
	Disabled bool
	VMData   *domain.VM
}
