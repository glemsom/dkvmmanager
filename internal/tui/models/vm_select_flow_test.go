// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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

	// Step 3: Press Enter to select first VM - uses sendKeysWithCmd to
	// exercise the full async command pipeline (VMSelectModel -> VMSelectedMsg -> handleVMSelected)
	t.Logf("Before Enter: currentView=%s, registryActive=%v", m.currentView, m.viewRegistry != nil && m.viewRegistry.IsActive())
	m = sendKeysWithCmd(t, m, tea.KeyEnter)
	t.Logf("After Enter: currentView=%s, vmEditModel=%v", m.currentView, m.viewRegistry != nil && m.viewRegistry.ActiveName() == ViewVMEdit)

	if m.currentView != ViewVMEdit {
		t.Errorf("Expected ViewVMEdit after pressing Enter in VMSelect, got %s", m.currentView)
	}
	if m.viewRegistry == nil || m.viewRegistry.ActiveName() != ViewVMEdit {
		t.Error("Expected VMEdit to be active in registry")
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

	// Press Down arrow - Verify no panic, cursor moves
	m = sendKeys(t, m, tea.KeyDown)

	// Press Enter to select (uses cmd chain to handle VMSelectedMsg)
	m = sendKeysWithCmd(t, m, tea.KeyEnter)

	if m.currentView != ViewVMEdit {
		t.Errorf("Expected ViewVMEdit after selecting VM, got %s", m.currentView)
	}
	if m.viewRegistry == nil || m.viewRegistry.ActiveName() != ViewVMEdit {
		t.Error("Expected VMEdit to be active in registry")
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
	viewContent := m.View().Content
	t.Logf("Full view:\n%s", viewContent)

	// Verify the view contains expected elements
	if !strings.Contains(viewContent, "DKVM Manager") {
		t.Error("View should contain header 'DKVM Manager'")
	}
	if !strings.Contains(viewContent, "alpha-vm") {
		t.Error("View should contain VM name 'alpha-vm'")
	}

	// The content view should show the VM list with help text
	if m.viewRegistry != nil && m.viewRegistry.ActiveName() == ViewVMSelect {
		contentView := m.viewRegistry.ActiveModel().View().Content
		t.Logf("Content view:\n%s", contentView)

		if !strings.Contains(contentView, "Navigate") {
			t.Error("Content view should contain help text 'Navigate'")
		}
		if !strings.Contains(contentView, "Enter Select") {
			t.Error("Content view should contain help text 'Enter Select'")
		}
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

	// Step 3: Press Enter to select first VM (use cmd chain)
	m = sendKeysWithCmd(t, m, tea.KeyEnter)

	if m.currentView != ViewVMDelete {
		t.Errorf("Expected ViewVMDelete after pressing Enter in VMSelect, got %s", m.currentView)
	}
}
