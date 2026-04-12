package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewBreadcrumbs(t *testing.T) {
	bc := NewBreadcrumbs()

	if bc.Len() != 0 {
		t.Errorf("Expected length 0, got %d", bc.Len())
	}

	items := bc.GetItems()
	if len(items) != 0 {
		t.Errorf("Expected empty items slice, got %d items", len(items))
	}

	current := bc.GetCurrent()
	if current.Label != "" || current.View != "" || current.Tab != 0 {
		t.Errorf("Expected empty current item, got %+v", current)
	}
}

func TestAddItem(t *testing.T) {
	bc := NewBreadcrumbs()

	tests := []struct {
		name     string
		label    string
		view     string
		tab      int
		expected BreadcrumbItem
	}{
		{
			name:     "Add first item",
			label:    "Home",
			view:     "home",
			tab:      0,
			expected: BreadcrumbItem{Label: "Home", View: "home", Tab: 0},
		},
		{
			name:     "Add second item",
			label:    "Start VM",
			view:     "vms",
			tab:      1,
			expected: BreadcrumbItem{Label: "Start VM", View: "vms", Tab: 1},
		},
		{
			name:     "Add item with special characters",
			label:    "VM Details",
			view:     "vm_details",
			tab:      2,
			expected: BreadcrumbItem{Label: "VM Details", View: "vm_details", Tab: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc.AddItem(tt.label, tt.view, tt.tab)

			if bc.Len() != 1 {
				t.Errorf("Expected length 1, got %d", bc.Len())
			}

			current := bc.GetCurrent()
			if current.Label != tt.expected.Label {
				t.Errorf("Expected label '%s', got '%s'", tt.expected.Label, current.Label)
			}
			if current.View != tt.expected.View {
				t.Errorf("Expected view '%s', got '%s'", tt.expected.View, current.View)
			}
			if current.Tab != tt.expected.Tab {
				t.Errorf("Expected tab %d, got %d", tt.expected.Tab, current.Tab)
			}

			// Clear for next test
			bc.Clear()
		})
	}
}

func TestAddMultipleItems(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	if bc.Len() != 3 {
		t.Errorf("Expected length 3, got %d", bc.Len())
	}

	items := bc.GetItems()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Check all items
	expectedItems := []BreadcrumbItem{
		{Label: "Home", View: "home", Tab: 0},
		{Label: "Start VM", View: "vms", Tab: 1},
		{Label: "VM Details", View: "vm_details", Tab: 2},
	}

	for i, expected := range expectedItems {
		if items[i].Label != expected.Label {
			t.Errorf("Item %d: Expected label '%s', got '%s'", i, expected.Label, items[i].Label)
		}
		if items[i].View != expected.View {
			t.Errorf("Item %d: Expected view '%s', got '%s'", i, expected.View, items[i].View)
		}
		if items[i].Tab != expected.Tab {
			t.Errorf("Item %d: Expected tab %d, got %d", i, expected.Tab, items[i].Tab)
		}
	}

	// Current should be the last item
	current := bc.GetCurrent()
	if current.Label != "VM Details" {
		t.Errorf("Expected current label 'VM Details', got '%s'", current.Label)
	}
}

func TestRemoveItem(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	tests := []struct {
		name          string
		removeIndex   int
		expectedLen   int
		expectedLabel string
	}{
		{
			name:          "Remove middle item",
			removeIndex:   1,
			expectedLen:   2,
			expectedLabel: "VM Details",
		},
		{
			name:          "Remove first item",
			removeIndex:   0,
			expectedLen:   1,
			expectedLabel: "VM Details",
		},
		{
			name:          "Remove last item",
			removeIndex:   0,
			expectedLen:   0,
			expectedLabel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc.RemoveItem(tt.removeIndex)

			if bc.Len() != tt.expectedLen {
				t.Errorf("Expected length %d, got %d", tt.expectedLen, bc.Len())
			}

			if tt.expectedLen > 0 {
				current := bc.GetCurrent()
				if current.Label != tt.expectedLabel {
					t.Errorf("Expected current label '%s', got '%s'", tt.expectedLabel, current.Label)
				}
			}
		})
	}
}

func TestRemoveItemInvalidIndex(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)

	// Try to remove with invalid indices
	bc.RemoveItem(-1)
	if bc.Len() != 2 {
		t.Errorf("Expected length 2 after removing -1, got %d", bc.Len())
	}

	bc.RemoveItem(5)
	if bc.Len() != 2 {
		t.Errorf("Expected length 2 after removing 5, got %d", bc.Len())
	}
}

func TestSetCurrent(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	tests := []struct {
		name          string
		index         int
		expectedLabel string
	}{
		{
			name:          "Set current to first item",
			index:         0,
			expectedLabel: "Home",
		},
		{
			name:          "Set current to middle item",
			index:         1,
			expectedLabel: "Start VM",
		},
		{
			name:          "Set current to last item",
			index:         2,
			expectedLabel: "VM Details",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc.SetCurrent(tt.index)

			current := bc.GetCurrent()
			if current.Label != tt.expectedLabel {
				t.Errorf("Expected current label '%s', got '%s'", tt.expectedLabel, current.Label)
			}
		})
	}
}

func TestSetCurrentInvalidIndex(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)

	// Try to set current with invalid indices
	bc.SetCurrent(-1)
	current := bc.GetCurrent()
	if current.Label != "Start VM" {
		t.Errorf("Expected current label 'VMs' after setting -1, got '%s'", current.Label)
	}

	bc.SetCurrent(5)
	current = bc.GetCurrent()
	if current.Label != "Start VM" {
		t.Errorf("Expected current label 'VMs' after setting 5, got '%s'", current.Label)
	}
}

func TestGetCurrent(t *testing.T) {
	bc := NewBreadcrumbs()

	// Empty breadcrumbs
	current := bc.GetCurrent()
	if current.Label != "" || current.View != "" || current.Tab != 0 {
		t.Errorf("Expected empty current item, got %+v", current)
	}

	// Add items
	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)

	current = bc.GetCurrent()
	if current.Label != "Start VM" {
		t.Errorf("Expected current label 'VMs', got '%s'", current.Label)
	}
}

func TestGetItems(t *testing.T) {
	bc := NewBreadcrumbs()

	// Empty breadcrumbs
	items := bc.GetItems()
	if len(items) != 0 {
		t.Errorf("Expected empty items slice, got %d items", len(items))
	}

	// Add items
	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	items = bc.GetItems()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	// Verify items
	expectedItems := []BreadcrumbItem{
		{Label: "Home", View: "home", Tab: 0},
		{Label: "Start VM", View: "vms", Tab: 1},
		{Label: "VM Details", View: "vm_details", Tab: 2},
	}

	for i, expected := range expectedItems {
		if items[i].Label != expected.Label {
			t.Errorf("Item %d: Expected label '%s', got '%s'", i, expected.Label, items[i].Label)
		}
		if items[i].View != expected.View {
			t.Errorf("Item %d: Expected view '%s', got '%s'", i, expected.View, items[i].View)
		}
		if items[i].Tab != expected.Tab {
			t.Errorf("Item %d: Expected tab %d, got %d", i, expected.Tab, items[i].Tab)
		}
	}
}

func TestClear(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	if bc.Len() != 3 {
		t.Errorf("Expected length 3 before clear, got %d", bc.Len())
	}

	bc.Clear()

	if bc.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", bc.Len())
	}

	items := bc.GetItems()
	if len(items) != 0 {
		t.Errorf("Expected empty items slice after clear, got %d items", len(items))
	}

	current := bc.GetCurrent()
	if current.Label != "" || current.View != "" || current.Tab != 0 {
		t.Errorf("Expected empty current item after clear, got %+v", current)
	}
}

func TestLen(t *testing.T) {
	bc := NewBreadcrumbs()

	if bc.Len() != 0 {
		t.Errorf("Expected length 0, got %d", bc.Len())
	}

	bc.AddItem("Home", "home", 0)
	if bc.Len() != 1 {
		t.Errorf("Expected length 1, got %d", bc.Len())
	}

	bc.AddItem("Start VM", "vms", 1)
	if bc.Len() != 2 {
		t.Errorf("Expected length 2, got %d", bc.Len())
	}

	bc.RemoveItem(0)
	if bc.Len() != 1 {
		t.Errorf("Expected length 1 after remove, got %d", bc.Len())
	}

	bc.Clear()
	if bc.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", bc.Len())
	}
}

func TestRenderEmpty(t *testing.T) {
	bc := NewBreadcrumbs()

	result := bc.Render()
	if result != "" {
		t.Errorf("Expected empty string for empty breadcrumbs, got '%s'", result)
	}
}

func TestRenderSingleItem(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)

	result := bc.Render()

	// Should contain the label
	if !strings.Contains(result, "Home") {
		t.Error("Render result does not contain 'Home'")
	}

	// Should not contain separator for single item
	if strings.Contains(result, " > ") {
		t.Error("Render result should not contain separator for single item")
	}
}

func TestRenderMultipleItems(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	result := bc.Render()

	// Should contain all labels
	if !strings.Contains(result, "Home") {
		t.Error("Render result does not contain 'Home'")
	}
	if !strings.Contains(result, "Start VM") {
		t.Error("Render result does not contain 'VMs'")
	}
	if !strings.Contains(result, "VM Details") {
		t.Error("Render result does not contain 'VM Details'")
	}

	// Should contain separators
	separatorCount := strings.Count(result, " > ")
	if separatorCount != 2 {
		t.Errorf("Expected 2 separators, got %d", separatorCount)
	}
}

func TestRenderStyling(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)

	result := bc.Render()

	// The result should contain the labels
	if !strings.Contains(result, "Home") {
		t.Error("Render result does not contain 'Home'")
	}

	if !strings.Contains(result, "Start VM") {
		t.Error("Render result does not contain 'VMs'")
	}

	// Should contain the separator
	if !strings.Contains(result, " > ") {
		t.Error("Render result does not contain separator")
	}
}

func TestRenderCurrentItemStyling(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)

	result := bc.Render()

	// The current item (last) should be present
	if !strings.Contains(result, "Start VM") {
		t.Error("Render result does not contain current item 'VMs'")
	}
}

func TestRenderClickableItemsStyling(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	result := bc.Render()

	// Clickable items (before current) should be present
	if !strings.Contains(result, "Home") {
		t.Error("Render result does not contain clickable item 'Home'")
	}

	if !strings.Contains(result, "Start VM") {
		t.Error("Render result does not contain clickable item 'VMs'")
	}
}

func TestRenderAfterStateChange(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)

	result1 := bc.Render()

	// Add another item
	bc.AddItem("VM Details", "vm_details", 2)

	result2 := bc.Render()

	// Results should be different
	if result1 == result2 {
		t.Error("Render should produce different results after state change")
	}

	// First result should not contain "VM Details"
	if strings.Contains(result1, "VM Details") {
		t.Error("First render should not contain 'VM Details'")
	}

	// Second result should contain "VM Details"
	if !strings.Contains(result2, "VM Details") {
		t.Error("Second render should contain 'VM Details'")
	}
}

func TestRenderConsistency(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	// Render multiple times should produce same result
	result1 := bc.Render()
	result2 := bc.Render()

	if result1 != result2 {
		t.Error("Render should produce consistent results for same input")
	}
}

func TestRenderAfterClear(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)

	result1 := bc.Render()
	if result1 == "" {
		t.Error("Render should not be empty before clear")
	}

	bc.Clear()

	result2 := bc.Render()
	if result2 != "" {
		t.Error("Render should be empty after clear")
	}
}

func TestRenderAfterRemoveItem(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	result1 := bc.Render()

	// Remove middle item
	bc.RemoveItem(1)

	result2 := bc.Render()

	// Results should be different
	if result1 == result2 {
		t.Error("Render should produce different results after removing item")
	}

	// First result should contain "Start VM"
	if !strings.Contains(result1, "Start VM") {
		t.Error("First render should contain 'VMs'")
	}

	// Second result should not contain "Start VM"
	if strings.Contains(result2, "Start VM") {
		t.Error("Second render should not contain 'VMs'")
	}
}

func TestRenderAfterSetCurrent(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	// Current is last item by default
	result1 := bc.Render()

	// Set current to first item
	bc.SetCurrent(0)

	result2 := bc.Render()

	// Both results should contain all labels
	if !strings.Contains(result1, "Home") {
		t.Error("First render should contain 'Home'")
	}
	if !strings.Contains(result1, "Start VM") {
		t.Error("First render should contain 'VMs'")
	}
	if !strings.Contains(result1, "VM Details") {
		t.Error("First render should contain 'VM Details'")
	}

	if !strings.Contains(result2, "Home") {
		t.Error("Second render should contain 'Home'")
	}
	if !strings.Contains(result2, "Start VM") {
		t.Error("Second render should contain 'VMs'")
	}
	if !strings.Contains(result2, "VM Details") {
		t.Error("Second render should contain 'VM Details'")
	}
}

func TestBreadcrumbItemStruct(t *testing.T) {
	item := BreadcrumbItem{
		Label: "Test Label",
		View:  "test_view",
		Tab:   5,
	}

	if item.Label != "Test Label" {
		t.Errorf("Expected label 'Test Label', got '%s'", item.Label)
	}
	if item.View != "test_view" {
		t.Errorf("Expected view 'test_view', got '%s'", item.View)
	}
	if item.Tab != 5 {
		t.Errorf("Expected tab 5, got %d", item.Tab)
	}
}

func TestRenderWithSpecialCharacters(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("VM's Details", "vm_details", 1)
	bc.AddItem("Settings & Config", "settings", 2)

	result := bc.Render()

	// Should contain all labels with special characters
	if !strings.Contains(result, "Home") {
		t.Error("Render result does not contain 'Home'")
	}
	if !strings.Contains(result, "VM's Details") {
		t.Error("Render result does not contain 'VM's Details'")
	}
	if !strings.Contains(result, "Settings & Config") {
		t.Error("Render result does not contain 'Settings & Config'")
	}
}

func TestRenderWithEmptyLabel(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("", "home", 0)
	bc.AddItem("Start VM", "vms", 1)

	result := bc.Render()

	// Should still render with empty label
	if !strings.Contains(result, "Start VM") {
		t.Error("Render result does not contain 'VMs'")
	}
}

func TestRenderWithLongLabel(t *testing.T) {
	bc := NewBreadcrumbs()

	longLabel := "This is a very long breadcrumb label that might be used in the UI"
	bc.AddItem("Home", "home", 0)
	bc.AddItem(longLabel, "long_view", 1)

	result := bc.Render()

	// Should contain the long label
	if !strings.Contains(result, longLabel) {
		t.Error("Render result does not contain long label")
	}
}

func TestRenderColorCodes(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)

	result := bc.Render()

	// Verify that the labels are present
	if !strings.Contains(result, "Home") {
		t.Error("Render result does not contain 'Home'")
	}

	if !strings.Contains(result, "Start VM") {
		t.Error("Render result does not contain 'VMs'")
	}

	// Verify that the separator is present
	if !strings.Contains(result, " > ") {
		t.Error("Render result does not contain separator")
	}
}

func TestRenderSeparatorCount(t *testing.T) {
	tests := []struct {
		name              string
		itemCount         int
		expectedSeparator int
	}{
		{"No items", 0, 0},
		{"One item", 1, 0},
		{"Two items", 2, 1},
		{"Three items", 3, 2},
		{"Four items", 4, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBreadcrumbs()

			for i := 0; i < tt.itemCount; i++ {
				bc.AddItem("Item", "view", i)
			}

			result := bc.Render()
			separatorCount := strings.Count(result, " > ")

			if separatorCount != tt.expectedSeparator {
				t.Errorf("Expected %d separators, got %d", tt.expectedSeparator, separatorCount)
			}
		})
	}
}

func TestRenderWidth(t *testing.T) {
	bc := NewBreadcrumbs()

	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)

	result := bc.Render()

	// The rendered width should be greater than 0
	renderedWidth := lipgloss.Width(result)
	if renderedWidth == 0 {
		t.Error("Rendered width should be greater than 0")
	}
}

func TestRenderAfterMultipleOperations(t *testing.T) {
	bc := NewBreadcrumbs()

	// Add items
	bc.AddItem("Home", "home", 0)
	bc.AddItem("Start VM", "vms", 1)
	bc.AddItem("VM Details", "vm_details", 2)

	// Remove middle item
	bc.RemoveItem(1)

	// Add another item
	bc.AddItem("Settings", "settings", 3)

	// Set current to first item
	bc.SetCurrent(0)

	result := bc.Render()

	// Should contain expected items
	if !strings.Contains(result, "Home") {
		t.Error("Render result does not contain 'Home'")
	}
	if !strings.Contains(result, "VM Details") {
		t.Error("Render result does not contain 'VM Details'")
	}
	if !strings.Contains(result, "Settings") {
		t.Error("Render result does not contain 'Settings'")
	}

	// Should not contain removed item
	if strings.Contains(result, "Start VM") {
		t.Error("Render result should not contain 'VMs' after removal")
	}
}
