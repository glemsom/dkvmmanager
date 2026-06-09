package models

import (
	"strings"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/tui/components"
	"github.com/glemsom/dkvmmanager/internal/version"
)

func TestMainModelViewMainMenu(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	viewContent := m.View().Content
	if viewContent == "" {
		t.Error("View() returned empty string")
	}

	// Should contain version info (new format: "v"+version.Version)
	if !strings.Contains(viewContent, "v"+version.Version) {
		t.Error("Main menu should contain version info")
	}

	// Should contain tab names
	if !strings.Contains(viewContent, "Start VM") {
		t.Error("View should contain Start VM tab")
	}
}

func TestMainModelViewQuitting(t *testing.T) {
	m := setupTestModel(t)
	m.quitting = true

	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Goodbye") {
		t.Error("View() should show goodbye message when quitting")
	}
}

func TestMainModelViewVMSelect(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.currentView = ViewVMSelect
	m.selectionMode = "edit"
	m.vmListForSelection, _ = m.vmManager.ListVMs()

	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Navigate") {
		t.Error("VM select view should contain navigation help")
	}
}

func TestMainModelViewVMSelectDelete(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.currentView = ViewVMSelect
	m.selectionMode = "delete"
	m.vmListForSelection, _ = m.vmManager.ListVMs()

	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Navigate") {
		t.Error("VM select delete view should contain navigation help")
	}
}

func TestMainModelViewVMCreateLoading(t *testing.T) {
	m := setupTestModel(t)
	reg := NewViewRegistry()
	reg.Register(&ViewDef{Name: ViewVMCreate, Factory: nil, BreadcrumbLabel: "Add VM", ParentTab: components.TabConfiguration, ConfigMenuIndex: 0})
	m.currentView = ViewVMCreate
	m.viewRegistry = reg
	m.windowWidth = 80
	m.windowHeight = 30

	viewContent := m.View().Content
	// Active view with nil model should show empty view (registry dispatches to nil model)
	if viewContent == "" {
		t.Error("Expected some output even with nil model in registry")
	}
}

func TestMainModelViewVMEditLoading(t *testing.T) {
	m := setupTestModel(t)
	reg := NewViewRegistry()
	reg.Register(&ViewDef{Name: ViewVMEdit, Factory: nil, BreadcrumbLabel: "Edit VM", ParentTab: components.TabConfiguration, ConfigMenuIndex: 1})
	m.currentView = ViewVMEdit
	m.viewRegistry = reg
	m.windowWidth = 80
	m.windowHeight = 30

	viewContent := m.View().Content
	// Active view with nil model should show empty view
	if viewContent == "" {
		t.Error("Expected some output even with nil model in registry")
	}
}

func TestMainModelViewVMDeleteLoading(t *testing.T) {
	m := setupTestModel(t)
	m.currentView = ViewVMDelete
	m.vmDeleteModel = nil
	m.windowWidth = 80
	m.windowHeight = 30

	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Loading...") {
		t.Errorf("Expected 'Loading...' for nil vmDeleteModel, got '%s'", viewContent)
	}
}

func TestRenderVMSelectViewWithTable(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.currentView = ViewVMSelect
	m.selectionMode = "edit"
	m.vmListForSelection, _ = m.vmManager.ListVMs()

	view := m.renderVMSelectView()

	if !strings.Contains(view, "Navigate") {
		t.Error("VM select view should contain navigation help")
	}
}

func TestStatusBarRendered(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Ready") {
		t.Error("View should contain status bar with 'Ready' mode")
	}
}

func TestTabBarRendered(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Start VM") {
		t.Error("View should contain Start VM tab")
	}
	if !strings.Contains(viewContent, "Configuration") {
		t.Error("View should contain Configuration tab")
	}
	if !strings.Contains(viewContent, "Power") {
		t.Error("View should contain Power tab")
	}
}

func TestBreadcrumbsOnSubView(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	// Enter config tab and then VM create
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.currentView = ViewVMCreate
	reg := NewViewRegistry()
	reg.Register(&ViewDef{Name: ViewVMCreate, Factory: nil, BreadcrumbLabel: "Add VM", ParentTab: components.TabConfiguration, ConfigMenuIndex: 0})
	reg.SetActiveModel(reg.GetDef(ViewVMCreate), NewVMCreateModel(m.vmManager))
	m.viewRegistry = reg
	m.breadcrumbs.AddItem("Configuration", "config", 1)
	m.breadcrumbs.AddItem("Add new VM", "vm_create", 1)

	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Configuration") {
		t.Error("Breadcrumbs should show 'Configuration'")
	}
}

func TestContentHeight(t *testing.T) {
	m := setupTestModel(t)
	m.windowHeight = 30
	m.windowWidth = 80

	// Without breadcrumbs
	m.breadcrumbs.Clear()
	height := m.contentHeight()
	if height < 3 {
		t.Errorf("Expected content height >= 3, got %d", height)
	}

	// With breadcrumbs
	m.breadcrumbs.AddItem("Test", "test", 0)
	heightWithBreadcrumbs := m.contentHeight()
	if heightWithBreadcrumbs >= height {
		t.Errorf("Expected content height with breadcrumbs (%d) to be less than without (%d)", heightWithBreadcrumbs, height)
	}
}
