// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"log"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// NewMainModel creates a new main model
func NewMainModel() (*MainModel, error) {
	return NewMainModelWithConfig(config.Load())
}

// NewMainModelWithConfig creates a new main model with the given configuration
func NewMainModelWithConfig(cfg *config.Config) (*MainModel, error) {

	// Create VM manager
	vmMgr, err := vm.NewManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM manager: %w", err)
	}

	// Build menu items
	menuItems := buildMenuItems(vmMgr)

	if debugMode {
		log.Printf("[DEBUG] MainModel created with %d menu items", len(menuItems))
	}

	// Initialize menu list
	menuListAdapter := buildMenuListAdapter(menuItems)
	delegate := MenuItemDelegate{}
	menuList := list.New(menuListAdapter, delegate, 80, 20)
	menuList.SetShowTitle(false)
	menuList.SetShowStatusBar(false)
	menuList.SetFilteringEnabled(false)
	menuList.SetShowHelp(false)

	// Initialize config list
	configListAdapter := buildConfigListAdapter()
	configDelegate := MenuItemDelegate{}
	configList := list.New(configListAdapter, configDelegate, 80, 20)
	configList.SetShowTitle(false)
	configList.SetShowStatusBar(false)
	configList.SetFilteringEnabled(false)
	configList.SetShowHelp(false)

	// Initialize power list
	powerListAdapter := buildPowerListAdapter()
	powerDelegate := MenuItemDelegate{}
	powerList := list.New(powerListAdapter, powerDelegate, 80, 20)
	powerList.SetShowTitle(false)
	powerList.SetShowStatusBar(false)
	powerList.SetFilteringEnabled(false)
	powerList.SetShowHelp(false)

	// Initialize tab model
	tabModel := components.NewTabModel()

	// Initialize status bar
	statusBar := components.NewStatusBar()
	vms, _ := vmMgr.ListVMs()
	statusBar.SetStats(len(vms), 0)

	// Initialize breadcrumbs
	breadcrumbs := components.NewBreadcrumbs()

	return &MainModel{
		currentView:   ViewMainMenu,
		cfg:           cfg,
		vmManager:     vmMgr,
		selectedIndex: 0,
		menuItems:     menuItems,
		menuList:      menuList,
		configList:    configList,
		powerList:     powerList,
		debugMode:     debugMode,
		tabModel:      tabModel,
		statusBar:     statusBar,
		breadcrumbs:   breadcrumbs,
	}, nil
}

// buildMenuListAdapter converts menu items to list.Item slice
func buildMenuListAdapter(items []MenuItem) []list.Item {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = MenuItemAdapter{MenuItem: item}
	}
	return listItems
}

// buildConfigListAdapter creates the config menu list items
func buildConfigListAdapter() []list.Item {
	items := []MenuItem{
		{Title: "Add new VM", Type: "INT_CONFIG"},
		{Title: "Edit VM", Type: "INT_CONFIG"},
		{Title: "Delete VM", Type: "INT_CONFIG"},
		{Title: "Edit CPU Topology", Type: "INT_CONFIG"},
		{Title: "Edit vCPU Pinning", Type: "INT_CONFIG"},
		{Title: "Edit PCI Passthrough", Type: "INT_CONFIG"},
		{Title: "Edit USB Passthrough", Type: "INT_CONFIG"},
		{Title: "Edit Start/Stop Script", Type: "INT_CONFIG"},
		{Title: "Edit CPU Options", Type: "INT_CONFIG"},
		{Title: "Set SSH Password", Type: "INT_CONFIG"},
		{Title: "Create Logical Volume", Type: "INT_CONFIG"},
		{Title: "Save changes", Type: "INT_CONFIG"},
	}
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = MenuItemAdapter{MenuItem: item}
	}
	return listItems
}

// buildPowerListAdapter creates the power menu list items
func buildPowerListAdapter() []list.Item {
	items := []MenuItem{
		{Title: "Reboot system", Type: "INT_POWER_REBOOT"},
		{Title: "Power off system", Type: "INT_POWER_OFF"},
	}
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = MenuItemAdapter{MenuItem: item}
	}
	return listItems
}

// SetDebugMode enables or disables debug mode for the models and vm packages
func SetDebugMode(enabled bool) {
	debugMode = enabled
	vm.SetDebugMode(enabled)
	if debugMode {
		log.Println("[DEBUG] Debug mode enabled for models package")
	}
}

// SetDryRunMode enables or disables dry-run mode for the models and vm packages
func SetDryRunMode(enabled bool) {
	dryRunMode = enabled
	vm.SetDryRunMode(enabled)
	if dryRunMode {
		log.Println("[DRY-RUN] Dry-run mode enabled for models package")
	}
}

// buildMenuItems builds the menu items from VM configurations
func buildMenuItems(mgr *vm.Manager) []MenuItem {
	items := []MenuItem{}

	// Get all VMs
	vms, err := mgr.ListVMs()
	if err != nil {
		if debugMode {
			log.Printf("[DEBUG] Error listing VMs: %v", err)
		}
	}

	// Sort VMs by ID to ensure deterministic ordering
	sort.Slice(vms, func(i, j int) bool {
		return vms[i].ID < vms[j].ID
	})

	for _, v := range vms {
		vmCopy := v
		items = append(items, MenuItem{
			Title:  v.Name,
			Type:   "VM",
			VMID:   v.ID,
			VMData: &vmCopy,
		})
	}

	return items
}

// rebuildMenuList refreshes menu items and syncs the list component
func (m *MainModel) rebuildMenuList() {
	m.menuItems = buildMenuItems(m.vmManager)
	m.menuList.SetItems(buildMenuListAdapter(m.menuItems))
	
	// Check if the VM that was running is still running
	runningCount := 0
	if m.runningVMID != "" && m.vmRunningModel != nil && m.vmRunningModel.Runner() != nil && m.vmRunningModel.Runner().IsRunning() {
		runningCount = 1
	}
	m.statusBar.SetStats(len(m.menuItems), runningCount)
}
