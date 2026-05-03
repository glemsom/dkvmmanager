// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
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
)

// Package-level debug mode flag
var debugMode bool

// Package-level dry-run mode flag
var dryRunMode bool

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

	// Configuration
	cfg *config.Config

	// VM manager
	vmManager *vm.Manager

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

	// Status message
	statusMessage string

	// Quitting flag
	quitting bool

	// VM creation model
	vmCreateModel *VMCreateModel

	// VM edit model
	vmEditModel *VMEditModel

	// VM delete model
	vmDeleteModel *VMDeleteModel

	// CPU options model
	cpuOptionsModel *CPUOptionsModel

	// VM running model
	vmRunningModel *VMRunningModel

	// PCI passthrough model
	pciPassthroughModel *PCIPassthroughModel

	// USB passthrough model
	usbPassthroughModel *USBPassthroughModel

	// CPU topology model
	cpuTopologyModel *CPUTopologyModel

	// vCPU pinning model
	vcpuPinningModel *VCPUPinningModel

	// SSH password model
	sshPasswordModel *SSHPasswordModel

	// Running VM ID - tracks the currently running VM to persist across rebuildMenuList calls
	runningVMID string

	// Start/Stop script form model
	startStopScriptModel *StartStopScriptModel

	// LVM LV create model
	lvCreateModel *LVCreateModel

	// VM list for selection
	vmListForSelection []models.VM
	vmSelectList       list.Model

	// Selection mode (edit or delete)
	selectionMode string

	// Debug mode flag
	debugMode bool

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
	VMData   *models.VM
}
