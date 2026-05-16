// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
	"github.com/glemsom/dkvmmanager/internal/vm"

	"golang.org/x/sys/unix"
)
func NewMainModel() (*MainModel, error) {
	return NewMainModelWithConfig(config.Load())
}

// NewMainModelWithConfig creates a new main model with the given configuration
func NewMainModelWithConfig(cfg *config.Config) (*MainModel, error) {

	// Create config repository
	repo, err := vm.NewRepository(cfg.VMsConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM repository: %w", err)
	}

	// Create VM manager
	vmMgr, err := vm.NewManagerWithRepository(cfg, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM manager: %w", err)
	}

	// Build menu items
	menuItems := buildMenuItems(vmMgr)

	if debugMode {
		log.Printf("[DEBUG] MainModel created with %d menu items", len(menuItems))
	}

	// Create view registry and register all views
	viewReg := NewViewRegistry()
	registerAllViews(viewReg, vmMgr, repo)

	// Check if /media/dkvmdata is a mount point
	initialView := ViewMainMenu
	if mounted, err := isMountPoint(dkvmDataMountPath); err == nil && !mounted {
		if debugMode {
			log.Printf("[DEBUG] %s is not a mount point, showing warning", dkvmDataMountPath)
		}
		initialView = ViewMountPointWarning
	} else if err != nil {
		if debugMode {
			log.Printf("[DEBUG] Error checking mount point %s: %v", dkvmDataMountPath, err)
		}
	}

	// Initialize menu list
	menuListAdapter := buildMenuListAdapter(menuItems)
	delegate := MenuItemDelegate{}
	menuList := list.New(menuListAdapter, delegate, 80, 20)
	menuList.SetShowTitle(false)
	menuList.SetShowStatusBar(false)
	menuList.SetFilteringEnabled(false)
	menuList.SetShowHelp(false)

	// Initialize config list from view registry
	configListAdapter := buildConfigListAdapter(viewReg)
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

	m := &MainModel{
		currentView:   initialView,
		viewRegistry:  viewReg,
		cfg:           cfg,
		vmManager:     vmMgr,
		configRepo:    repo,
		hostDiscovery: &vm.DefaultHostDiscovery{},
		selectedIndex: 0,
		menuItems:     menuItems,
		menuList:      menuList,
		configList:    configList,
		powerList:     powerList,
		debugMode:     debugMode,
		tabModel:      tabModel,
		statusBar:     statusBar,
		breadcrumbs:   breadcrumbs,
	}

	// Activate mount point warning through the registry if needed
	if initialView == ViewMountPointWarning {
		if _, err := viewReg.Activate(ViewMountPointWarning, m); err != nil {
			if debugMode {
				log.Printf("[DEBUG] Failed to activate mount point warning: %v", err)
			}
			m.currentView = ViewMainMenu
		} else {
			// Set initial size using fallback terminal size since WindowSizeMsg
			// hasn't been received yet
			width, height := getInitialTerminalSize()
			m.viewRegistry.ActiveModel().SetSize(width-4, height-2)
		}
	}

	return m, nil
}

// registerAllViews registers all config form views and special views in the registry.
func registerAllViews(reg *ViewRegistry, vmManager *vm.Manager, configRepo *vm.Repository) {
	// 0: Add new VM
	reg.Register(&ViewDef{Name: ViewVMCreate, Factory: func(m *MainModel) (SubViewModel, error) {
		return NewVMCreateModel(m.vmManager), nil
	}, BreadcrumbLabel: "Add VM", ParentTab: components.TabConfiguration, ConfigMenuIndex: 0})

	// 1: Edit VM — handled specially (needs VM selection first; activated via SetActiveModel)
	reg.Register(&ViewDef{Name: ViewVMEdit, Factory: nil, BreadcrumbLabel: "Edit VM", ParentTab: components.TabConfiguration, ConfigMenuIndex: -1})

	// 2: Delete VM — factory returns error; activated via SetActiveModel after VM selection
	reg.Register(&ViewDef{Name: ViewVMDelete, Factory: func(m *MainModel) (SubViewModel, error) {
		return nil, fmt.Errorf("use SetActiveModel for VMDelete")
	}, BreadcrumbLabel: "Delete VM", ParentTab: components.TabConfiguration, ConfigMenuIndex: -1})

	// 3: CPU Topology
	reg.Register(&ViewDef{Name: ViewCPUTopology, Factory: func(m *MainModel) (SubViewModel, error) {
		return NewCPUTopologyModel(m.configRepo)
	}, BreadcrumbLabel: "Edit CPU Topology", ParentTab: components.TabConfiguration, ConfigMenuIndex: 3})

	// 4: vCPU Pinning
	reg.Register(&ViewDef{Name: ViewVCPUPinning, Factory: func(m *MainModel) (SubViewModel, error) {
		return NewVCPUPinningModel(m.vmManager, m.configRepo)
	}, BreadcrumbLabel: "Edit vCPU Pinning", ParentTab: components.TabConfiguration, ConfigMenuIndex: 4})

	// 5: PCI Passthrough
	reg.Register(&ViewDef{Name: ViewPCIPassthrough, Factory: func(m *MainModel) (SubViewModel, error) {
		return NewPCIPassthroughModel(m.configRepo, m.vmManager, m.hostDiscovery)
	}, BreadcrumbLabel: "Edit PCI Passthrough", ParentTab: components.TabConfiguration, ConfigMenuIndex: 5})

	// 6: USB Passthrough
	reg.Register(&ViewDef{Name: ViewUSBPassthrough, Factory: func(m *MainModel) (SubViewModel, error) {
		return NewUSBPassthroughModel(m.configRepo, m.hostDiscovery)
	}, BreadcrumbLabel: "Edit USB Passthrough", ParentTab: components.TabConfiguration, ConfigMenuIndex: 6})

	// 7: Start/Stop Script
	reg.Register(&ViewDef{Name: ViewStartStopScript, Factory: func(m *MainModel) (SubViewModel, error) {
		return NewStartStopScriptModel(m.configRepo)
	}, BreadcrumbLabel: "Edit Start/Stop Script", ParentTab: components.TabConfiguration, ConfigMenuIndex: 7})

	// 8: CPU Options
	reg.Register(&ViewDef{Name: ViewCPUOptions, Factory: func(m *MainModel) (SubViewModel, error) {
		return NewCPUOptionsModel(m.configRepo), nil
	}, BreadcrumbLabel: "Edit CPU Options", ParentTab: components.TabConfiguration, ConfigMenuIndex: 8})

	// 9: SSH Password
	reg.Register(&ViewDef{Name: ViewSSHPassword, Factory: func(m *MainModel) (SubViewModel, error) {
		return NewSSHPasswordModel(), nil
	}, BreadcrumbLabel: "Set SSH Password", ParentTab: components.TabConfiguration, ConfigMenuIndex: 9})

	// 10: Create Logical Volume
	reg.Register(&ViewDef{Name: ViewLVCreate, Factory: func(m *MainModel) (SubViewModel, error) {
		return NewLVCreateModel(), nil
	}, BreadcrumbLabel: "Create Logical Volume", ParentTab: components.TabConfiguration, ConfigMenuIndex: 10})

	// Mount point warning (not in config menu)
	reg.Register(&ViewDef{Name: ViewMountPointWarning, Factory: func(m *MainModel) (SubViewModel, error) {
		return NewMountPointWarningModel(), nil
	}, BreadcrumbLabel: "Mount Point Warning", ParentTab: components.TabVMs, ConfigMenuIndex: -1})

	// VMRunning — persistent model; real model is set via SetActiveModel in handleVMSelection
	reg.Register(&ViewDef{Name: ViewVMRunning, Factory: func(m *MainModel) (SubViewModel, error) {
		// Factory only used as placeholder; real model created externally with runner info
		return NewVMRunningModel(nil, nil), nil
	}, BreadcrumbLabel: "VM Running", ParentTab: components.TabVMs, ConfigMenuIndex: -1})

	// VMSelect — inline state on MainModel, never factory-constructed
	reg.Register(&ViewDef{Name: ViewVMSelect, Factory: nil,
		BreadcrumbLabel: "", ParentTab: components.TabConfiguration, ConfigMenuIndex: -1})
}

// buildMenuListAdapter converts menu items to list.Item slice
func buildMenuListAdapter(items []MenuItem) []list.Item {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = MenuItemAdapter{MenuItem: item}
	}
	return listItems
}

// buildConfigListAdapter creates the config menu list items from the registry.
// Edit VM and Delete VM are inserted manually since they require VM selection first.
func buildConfigListAdapter(reg *ViewRegistry) []list.Item {
	items := reg.BuildConfigMenuItems()
	// Insert "Edit VM" and "Delete VM" after "Add VM" (position 1 and 2)
	// matching the original menu layout. The last item is "Save changes".
	editVM := MenuItem{Title: "Edit VM", Type: "INT_CONFIG"}
	deleteVM := MenuItem{Title: "Delete VM", Type: "INT_CONFIG"}

	// Items from registry: [Add VM, CPU Topology, vCPU Pinning, ..., LV Create, Save changes]
	// We want: [Add VM, Edit VM, Delete VM, CPU Topology, ..., LV Create, Save changes]
	listItems := make([]list.Item, 0, len(items)+2)
	// First item is always Add VM
	listItems = append(listItems, MenuItemAdapter{MenuItem: items[0]})
	// Insert Edit VM and Delete VM
	listItems = append(listItems, MenuItemAdapter{MenuItem: editVM})
	listItems = append(listItems, MenuItemAdapter{MenuItem: deleteVM})
	// Remaining registry items (excluding Save changes which is last)
	for i := 1; i < len(items)-1; i++ {
		listItems = append(listItems, MenuItemAdapter{MenuItem: items[i]})
	}
	// Save changes at the end
	listItems = append(listItems, MenuItemAdapter{MenuItem: items[len(items)-1]})
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

// getInitialTerminalSize gets terminal size with fallback defaults for initial view sizing.
// This avoids circular imports by duplicating terminal size logic here.
func getInitialTerminalSize() (width, height int) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		// Try environment variables
		width = getEnvInt("COLUMNS", 80)
		height = getEnvInt("LINES", 25)
		if width == 0 {
			width = 80
		}
		if height == 0 {
			height = 25
		}
		return width, height
	}
	return int(ws.Col), int(ws.Row)
}

// getEnvInt retrieves an integer from an environment variable
func getEnvInt(name string, defaultValue int) int {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intVal
}
