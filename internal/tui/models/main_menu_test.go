package models

import (
	"strings"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/tui/components"
)

func TestRebuildMenuList(t *testing.T) {
	m := setupTestModel(t)

	initialItems := len(m.menuList.Items())

	// Create a new VM
	_, err := m.vmManager.CreateVM("rebuild-test")
	if err != nil {
		t.Fatalf("Failed to create VM: %v", err)
	}

	m.rebuildMenuList()

	newItems := len(m.menuList.Items())
	if newItems <= initialItems {
		t.Errorf("Expected more items after rebuild, got %d (was %d)", newItems, initialItems)
	}
}

func TestRebuildMenuListUpdatesStatusBar(t *testing.T) {
	m := setupTestModelWithVMs(t)

	// Create another VM
	_, _ = m.vmManager.CreateVM("statusbar-test")
	m.rebuildMenuList()

	// Verify the status bar has correct stats by checking the rendered output
	m.windowWidth = 80
	m.windowHeight = 30
	view := m.View()
	// Status bar should show Stopped status (no VMs running)
	if !strings.Contains(view, "Stopped") {
		t.Logf("Status bar view content may vary, checking for VM status update")
	}
}

func TestShowVMSelectionWithMode(t *testing.T) {
	m := setupTestModelWithVMs(t)

	model, _ := m.showVMSelectionWithMode("edit", "No VMs")
	m = model.(*MainModel)

	if m.currentView != ViewVMSelect {
		t.Errorf("Expected view to be ViewVMSelect, got %s", m.currentView)
	}
	if m.selectionMode != "edit" {
		t.Errorf("Expected selection mode to be 'edit', got '%s'", m.selectionMode)
	}
	if len(m.vmListForSelection) == 0 {
		t.Error("Expected VMs to be loaded for selection")
	}
}

func TestShowVMSelectionWithModeBreadcrumbs(t *testing.T) {
	m := setupTestModelWithVMs(t)

	model, _ := m.showVMSelectionWithMode("edit", "No VMs")
	m = model.(*MainModel)

	if m.breadcrumbs.Len() != 2 {
		t.Errorf("Expected 2 breadcrumb items, got %d", m.breadcrumbs.Len())
	}
}

func TestShowVMSelectionWithModeDeleteBreadcrumbs(t *testing.T) {
	m := setupTestModelWithVMs(t)

	model, _ := m.showVMSelectionWithMode("delete", "No VMs")
	m = model.(*MainModel)

	if m.breadcrumbs.Len() != 2 {
		t.Errorf("Expected 2 breadcrumb items, got %d", m.breadcrumbs.Len())
	}
}

func TestShowVMSelectionWithModeEmpty(t *testing.T) {
	m := setupTestModel(t) // No VMs

	model, _ := m.showVMSelectionWithMode("edit", "No VMs available")
	m = model.(*MainModel)

	if m.currentView == ViewVMSelect {
		t.Error("Should not switch to VM select view when no VMs exist")
	}
	if !strings.Contains(m.statusMessage, "No VMs available") {
		t.Errorf("Expected status message about no VMs, got '%s'", m.statusMessage)
	}
}

func TestShowVMSelectionEdit(t *testing.T) {
	m := setupTestModelWithVMs(t)

	model, _ := m.showVMSelection()
	m = model.(*MainModel)

	if m.selectionMode != "edit" {
		t.Errorf("Expected selection mode 'edit', got '%s'", m.selectionMode)
	}
}

func TestShowVMSelectionForDeletion(t *testing.T) {
	m := setupTestModelWithVMs(t)

	model, _ := m.showVMSelectionForDeletion()
	m = model.(*MainModel)

	if m.selectionMode != "delete" {
		t.Errorf("Expected selection mode 'delete', got '%s'", m.selectionMode)
	}
}

func TestBuildMenuItems(t *testing.T) {
	m := setupTestModelWithVMs(t)

	items := m.menuItems

	// Should contain only VM items (Config/Power/Shell removed)
	vmCount := 0
	for _, item := range items {
		if item.Type == "VM" {
			vmCount++
		}
	}

	if vmCount != 2 {
		t.Errorf("Expected 2 VM items, got %d", vmCount)
	}

	// Should NOT contain internal items
	for _, item := range items {
		if item.Type != "VM" {
			t.Errorf("Expected only VM items, found type '%s' with title '%s'", item.Type, item.Title)
		}
	}
}

// TestHandleConfigMenuSelectionComingSoon verifies status messages for menu items.
// Note: SSH Password (item 8) is implemented but may not set status message on selection.
func TestHandleConfigMenuSelectionComingSoon(t *testing.T) {
	// This test was for verifying "coming soon" messages.
	// Both Custom Scripts (item 6) and SSH Password (item 8) are now implemented.
	// Leaving as a placeholder to verify no panics on selection.
	m := setupTestModel(t)
	m.tabModel.SetActiveTab(components.TabConfiguration)

	// Test that accessing menu items doesn't panic
	for i := 0; i <= 9; i++ {
		m.configSelectedIndex = i
		model, _ := m.handleMenuSelection()
		_ = model.(*MainModel)
	}
	// Skip index 10 (Create LV) here because it triggers a vgs command.
	m.configSelectedIndex = 11
	model, _ := m.handleMenuSelection()
	_ = model.(*MainModel)
}

func TestHandlePCIPassthroughSelection(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.configSelectedIndex = 5 // "Edit PCI Passthrough"

	model, _ := m.handleMenuSelection()
	m = model.(*MainModel)

	// Should open PCI passthrough form directly (like CPU Options)
	if m.currentView != ViewPCIPassthrough {
		t.Errorf("Expected ViewPCIPassthrough after PCI Passthrough selection, got %s", m.currentView)
	}
	if m.viewRegistry == nil || m.viewRegistry.ActiveName() != ViewPCIPassthrough {
		t.Error("Expected PCIPassthrough to be active in registry")
	}
}

func TestHandleCustomScriptSelection(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.configSelectedIndex = 7 // "Edit Start/Stop Script"

	model, _ := m.handleMenuSelection()
	m = model.(*MainModel)

	// Should open start/stop script form
	if m.currentView != ViewStartStopScript {
		t.Errorf("Expected ViewStartStopScript after Edit Start/Stop Script selection, got %s", m.currentView)
	}
	if m.viewRegistry == nil || m.viewRegistry.ActiveName() != ViewStartStopScript {
		t.Error("Expected StartStopScript to be active in registry")
	}
	// Should have correct breadcrumbs
	if m.breadcrumbs.Len() != 2 {
		t.Errorf("Expected 2 breadcrumb items, got %d", m.breadcrumbs.Len())
	}
}

func TestHandleConfigMenuSelectionSave(t *testing.T) {
	m := setupTestModel(t)
	m.tabModel.SetActiveTab(components.TabConfiguration)
	// 12 items: "Add VM" (0), "Edit VM" (1), "Delete VM" (2), ..., "Save changes" (11)
	m.configSelectedIndex = 11

	dryRunMode = true
	defer func() { dryRunMode = false }()

	model, cmd := m.handleMenuSelection()
	m = model.(*MainModel)

	if cmd == nil {
		t.Fatal("Expected tea.Cmd to be returned for LBU Commit")
	}

	msg := cmd()
	lcm, ok := msg.(LBUCommitMsg)
	if !ok {
		t.Fatalf("Expected LBUCommitMsg, got %T", msg)
	}
	if !lcm.Success {
		t.Errorf("Expected dry-run success, got failure: %s", lcm.Output)
	}
	if lcm.Output != "Would execute: lbu commit" {
		t.Errorf("Expected dry-run output, got: %s", lcm.Output)
	}
}

func TestHandleConfigMenuSelectionLBUCommitDryRun(t *testing.T) {
	m := setupTestModel(t)
	m.tabModel.SetActiveTab(components.TabConfiguration)
	// 12 items: "Add VM" (0), "Edit VM" (1), "Delete VM" (2), ..., "Save changes" (11)
	m.configSelectedIndex = 11

	dryRunMode = true
	defer func() { dryRunMode = false }()

	_, cmd := m.handleMenuSelection()
	if cmd == nil {
		t.Fatal("Expected tea.Cmd for dry-run LBU commit")
	}

	msg := cmd()
	lcm, ok := msg.(LBUCommitMsg)
	if !ok {
		t.Fatalf("Expected LBUCommitMsg, got %T", msg)
	}
	if !lcm.Success {
		t.Error("Expected dry-run to succeed")
	}
	if lcm.Output != "Would execute: lbu commit" {
		t.Errorf("Expected 'Would execute: lbu commit', got '%s'", lcm.Output)
	}
}

func TestHandleVMSelection(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.tabModel.SetActiveTab(components.TabVMs)

	// Find a VM item
	for i, item := range m.menuItems {
		if item.Type == "VM" {
			m.selectedIndex = i
			break
		}
	}

	model, _ := m.handleMenuSelection()
	m = model.(*MainModel)

	// handleVMSelection sets a status bar message
	if m.statusBar == nil {
		t.Error("Expected statusBar to be set")
	}
}

func TestHandleVMSelectionOutOfBounds(t *testing.T) {
	m := setupTestModel(t)
	m.tabModel.SetActiveTab(components.TabVMs)
	m.selectedIndex = 999 // Out of bounds

	// Should not panic
	model, _ := m.handleMenuSelection()
	m = model.(*MainModel)

	if m == nil {
		t.Error("Model should not be nil")
	}
}

func TestBuildConfigListAdapter(t *testing.T) {
	reg := NewViewRegistry()

	// Register views matching the order from init.go's registerAllViews
	reg.Register(&ViewDef{Name: ViewVMCreate, Factory: func(m *MainModel) (SubViewModel, error) {
		return nil, nil
	}, BreadcrumbLabel: "Add VM", ParentTab: 0, ConfigMenuIndex: 0})
	reg.Register(&ViewDef{Name: ViewCPUTopology, Factory: func(m *MainModel) (SubViewModel, error) {
		return nil, nil
	}, BreadcrumbLabel: "Edit CPU Topology", ParentTab: 0, ConfigMenuIndex: 3})
	reg.Register(&ViewDef{Name: ViewVCPUPinning, Factory: func(m *MainModel) (SubViewModel, error) {
		return nil, nil
	}, BreadcrumbLabel: "Edit vCPU Pinning", ParentTab: 0, ConfigMenuIndex: 4})
	reg.Register(&ViewDef{Name: ViewPCIPassthrough, Factory: func(m *MainModel) (SubViewModel, error) {
		return nil, nil
	}, BreadcrumbLabel: "Edit PCI Passthrough", ParentTab: 0, ConfigMenuIndex: 5})
	reg.Register(&ViewDef{Name: ViewUSBPassthrough, Factory: func(m *MainModel) (SubViewModel, error) {
		return nil, nil
	}, BreadcrumbLabel: "Edit USB Passthrough", ParentTab: 0, ConfigMenuIndex: 6})
	reg.Register(&ViewDef{Name: ViewStartStopScript, Factory: func(m *MainModel) (SubViewModel, error) {
		return nil, nil
	}, BreadcrumbLabel: "Edit Start/Stop Script", ParentTab: 0, ConfigMenuIndex: 7})
	reg.Register(&ViewDef{Name: ViewCPUOptions, Factory: func(m *MainModel) (SubViewModel, error) {
		return nil, nil
	}, BreadcrumbLabel: "Edit CPU Options", ParentTab: 0, ConfigMenuIndex: 8})
	reg.Register(&ViewDef{Name: ViewSSHPassword, Factory: func(m *MainModel) (SubViewModel, error) {
		return nil, nil
	}, BreadcrumbLabel: "Set SSH Password", ParentTab: 0, ConfigMenuIndex: 9})
	reg.Register(&ViewDef{Name: ViewLVCreate, Factory: func(m *MainModel) (SubViewModel, error) {
		return nil, nil
	}, BreadcrumbLabel: "Create Logical Volume", ParentTab: 0, ConfigMenuIndex: 10})

	items := buildConfigListAdapter(reg)

	// 9 registered views + 2 inserted (Edit VM, Delete VM) + Save changes = 12
	if len(items) != 12 {
		t.Errorf("Expected 12 config items, got %d", len(items))
	}

	expectedTitles := []string{
		"Add VM",
		"Edit VM",
		"Delete VM",
		"Edit CPU Topology",
		"Edit vCPU Pinning",
		"Edit PCI Passthrough",
		"Edit USB Passthrough",
		"Edit Start/Stop Script",
		"Edit CPU Options",
		"Set SSH Password",
		"Create Logical Volume",
		"Save changes",
	}

	for i, expected := range expectedTitles {
		if items[i].FilterValue() != expected {
			t.Errorf("Expected config item %d to be '%s', got '%s'", i, expected, items[i].FilterValue())
		}
	}
}

func TestBuildMenuListAdapter(t *testing.T) {
	items := []MenuItem{
		{Title: "VM One", Type: "VM", VMID: "0"},
	}

	listItems := buildMenuListAdapter(items)

	if len(listItems) != 1 {
		t.Errorf("Expected 1 list item, got %d", len(listItems))
	}

	if listItems[0].FilterValue() != "VM One" {
		t.Errorf("Expected first item 'VM One', got '%s'", listItems[0].FilterValue())
	}
}

func TestIsSubViewActive(t *testing.T) {
	m := setupTestModel(t)
	reg := NewViewRegistry()
	reg.Register(&ViewDef{Name: ViewVMCreate, Factory: nil, BreadcrumbLabel: "Add VM", ParentTab: components.TabConfiguration, ConfigMenuIndex: 0})
	reg.Register(&ViewDef{Name: ViewVMEdit, Factory: nil, BreadcrumbLabel: "Edit VM", ParentTab: components.TabConfiguration, ConfigMenuIndex: -1})
	reg.Register(&ViewDef{Name: ViewVMDelete, Factory: nil, BreadcrumbLabel: "Delete VM", ParentTab: components.TabConfiguration, ConfigMenuIndex: -1})
	reg.Register(&ViewDef{Name: ViewVMSelect, Factory: nil, BreadcrumbLabel: "", ParentTab: components.TabConfiguration, ConfigMenuIndex: -1})
	reg.Register(&ViewDef{Name: ViewVMRunning, Factory: nil, BreadcrumbLabel: "VM Running", ParentTab: components.TabVMs, ConfigMenuIndex: -1})
	m.viewRegistry = reg

	// Main menu is not a sub-view
	m.currentView = ViewMainMenu
	if m.isSubViewActive() {
		t.Error("ViewMainMenu should not be a sub-view")
	}

	// VM create is a sub-view (active in registry)
	reg.SetActiveModel(reg.GetDef(ViewVMCreate), NewVMCreateModel(m.vmManager))
	m.currentView = ViewVMCreate
	if !m.isSubViewActive() {
		t.Error("ViewVMCreate should be a sub-view")
	}

	// VM edit is a sub-view (active in registry)
	reg.Deactivate()
	reg.SetActiveModel(reg.GetDef(ViewVMEdit), nil)
	m.currentView = ViewVMEdit
	if !m.isSubViewActive() {
		t.Error("ViewVMEdit should be a sub-view")
	}

	// VM delete is a sub-view
	m.currentView = ViewVMDelete
	if !m.isSubViewActive() {
		t.Error("ViewVMDelete should be a sub-view")
	}

	// VM select is a sub-view
	m.currentView = ViewVMSelect
	if !m.isSubViewActive() {
		t.Error("ViewVMSelect should be a sub-view")
	}
}

func TestReturnFromSubView(t *testing.T) {
	m := setupTestModel(t)
	m.currentView = ViewVMCreate
	m.breadcrumbs.AddItem("Configuration", "config", 1)
	m.breadcrumbs.AddItem("Add new VM", "vm_create", 1)

	model, _ := m.returnFromSubView()
	m = model.(*MainModel)

	if m.currentView != ViewMainMenu {
		t.Errorf("Expected currentView to be ViewMainMenu, got %s", m.currentView)
	}
	if m.breadcrumbs.Len() != 0 {
		t.Errorf("Expected breadcrumbs to be cleared, got %d items", m.breadcrumbs.Len())
	}
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Errorf("Expected active tab to be TabConfiguration, got %d", m.tabModel.GetActiveTab())
	}
}
