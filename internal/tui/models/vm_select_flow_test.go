package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
)

// TestVMSelectFullFlowViaUpdate simulates the exact user flow through Update():
// Config tab -> Edit VM (Enter) -> VMSelect shown -> Enter to select VM -> ViewVMEdit
func TestVMSelectFullFlowViaUpdate(t *testing.T) {
	m := setupTestModelForScenarios(t)

	// Step 1: Switch to Configuration tab
	m = sendKeys(t, m, tea.KeyTab)
	if m.tabModel.GetActiveTab() != components.TabConfiguration {
		t.Fatal("Expected Configuration tab")
	}

	// Step 2: Press Enter on "Edit VM" (configSelectedIndex=1)
	m.configSelectedIndex = 1
	m = sendKeys(t, m, tea.KeyEnter)

	if m.currentView != ViewVMSelect {
		t.Fatalf("Expected ViewVMSelect after selecting Edit VM, got %s", m.currentView)
	}

	// Step 3: Press Enter to select first VM - THIS IS THE KEY TEST
	t.Logf("Before Enter: currentView=%s, vmSelectList items=%d", m.currentView, len(m.vmListForSelection))
	m = sendKeys(t, m, tea.KeyEnter)
	t.Logf("After Enter: currentView=%s, vmEditModel=%v", m.currentView, m.vmEditModel != nil)

	if m.currentView != ViewVMEdit {
		t.Errorf("Expected ViewVMEdit after pressing Enter in VMSelect, got %s", m.currentView)
	}
	if m.vmEditModel == nil {
		t.Error("Expected vmEditModel to be initialized")
	}
}

// TestVMSelectArrowNavigationViaUpdate tests that arrow keys move the cursor
// in the VM selection table through the full Update() flow.
func TestVMSelectArrowNavigationViaUpdate(t *testing.T) {
	m := setupTestModelForScenarios(t)
	m.windowWidth = 80
	m.windowHeight = 30

	// Ensure we have 2 VMs
	vms, _ := m.vmManager.ListVMs()
	if len(vms) < 2 {
		t.Fatalf("Need at least 2 VMs, got %d", len(vms))
	}

	// Navigate to VMSelect
	m = sendKeys(t, m, tea.KeyTab) // Config tab
	m.configSelectedIndex = 1      // Edit VM
	m = sendKeys(t, m, tea.KeyEnter)

	if m.currentView != ViewVMSelect {
		t.Fatalf("Expected ViewVMSelect, got %s", m.currentView)
	}

	// Initial cursor should be 0
	initialCursor := m.vmSelectList.Index()
	t.Logf("Initial cursor: %d", initialCursor)

	// Press Down arrow
	m = sendKeys(t, m, tea.KeyDown)
	afterDownCursor := m.vmSelectList.Index()
	t.Logf("After Down cursor: %d", afterDownCursor)

	if afterDownCursor <= initialCursor {
		t.Errorf("Expected cursor to move down after KeyDown, got %d (was %d)", afterDownCursor, initialCursor)
	}

	// Press Enter to select second VM
	m = sendKeys(t, m, tea.KeyEnter)

	if m.currentView != ViewVMEdit {
		t.Errorf("Expected ViewVMEdit after selecting second VM, got %s", m.currentView)
	}
	if m.vmEditModel == nil {
		t.Error("Expected vmEditModel to be initialized")
	}

	// Verify the correct VM was selected (second VM in the table)
	tableVMs := m.vmListForSelection
	cursor := m.vmSelectList.Index()
	if cursor < 0 || cursor >= len(tableVMs) {
		t.Fatalf("Invalid cursor position: %d (have %d VMs)", cursor, len(tableVMs))
	}
	expectedVM := tableVMs[cursor]
	fm := getVMForm(m.vmEditModel.form)
	if fm == nil {
		t.Fatal("Could not get VMFormModel from edit model")
	}
	if fm.vmName != expectedVM.Name {
		t.Errorf("Expected VM name '%s' (table row %d), got '%s'", expectedVM.Name, cursor, fm.vmName)
	}
}

// TestVMSelectViewRendering verifies that the VMSelect view renders correctly
func TestVMSelectViewRendering(t *testing.T) {
	m := setupTestModelForScenarios(t)

	// Create 3 VMs to have a reasonable list
	m.vmManager.CreateVM("alpha-vm")
	m.vmManager.CreateVM("beta-vm")
	m.vmManager.CreateVM("gamma-vm")
	m.rebuildMenuList()

	// Navigate to VMSelect
	m = sendKeys(t, m, tea.KeyTab) // Config tab
	m.configSelectedIndex = 1      // Edit VM
	m = sendKeys(t, m, tea.KeyEnter)

	if m.currentView != ViewVMSelect {
		t.Fatalf("Expected ViewVMSelect, got %s", m.currentView)
	}

	// Render the full view
	view := m.View()
	t.Logf("Full view:\n%s", view)

	// Verify the view contains expected elements
	if !strings.Contains(view, "DKVM Manager") {
		t.Error("View should contain header 'DKVM Manager'")
	}
	if !strings.Contains(view, "alpha-vm") {
		t.Error("View should contain VM name 'alpha-vm'")
	}

	// The renderVMSelectView should show the table with help text
	contentView := m.renderVMSelectView()
	t.Logf("Content view:\n%s", contentView)

	if !strings.Contains(contentView, "Navigate") {
		t.Error("Content view should contain help text 'Navigate'")
	}
	if !strings.Contains(contentView, "Enter Select") {
		t.Error("Content view should contain help text 'Enter Select'")
	}
}

// TestVMSelectDeleteFullFlowViaUpdate simulates the delete flow through Update():
// Config tab -> Delete VM (Enter) -> VMSelect shown -> Enter to select VM -> ViewVMDelete
func TestVMSelectDeleteFullFlowViaUpdate(t *testing.T) {
	m := setupTestModelForScenarios(t)

	// Step 1: Switch to Configuration tab
	m = sendKeys(t, m, tea.KeyTab)

	// Step 2: Press Enter on "Delete VM" (configSelectedIndex=2)
	m.configSelectedIndex = 2
	m = sendKeys(t, m, tea.KeyEnter)

	if m.currentView != ViewVMSelect {
		t.Fatalf("Expected ViewVMSelect after selecting Delete VM, got %s", m.currentView)
	}

	// Step 3: Press Enter to select first VM
	m = sendKeys(t, m, tea.KeyEnter)

	if m.currentView != ViewVMDelete {
		t.Errorf("Expected ViewVMDelete after pressing Enter in VMSelect, got %s", m.currentView)
	}
	if m.vmDeleteModel == nil {
		t.Error("Expected vmDeleteModel to be initialized")
	}
}
