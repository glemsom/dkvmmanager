package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewTabModel(t *testing.T) {
	tm := NewTabModel()

	if tm.GetActiveTab() != TabVMs {
		t.Errorf("Expected active tab to be TabVMs, got %v", tm.GetActiveTab())
	}

	tabs := tm.GetTabs()
	if len(tabs) != 3 {
		t.Errorf("Expected 3 tabs, got %d", len(tabs))
	}
}

func TestSetActiveTab(t *testing.T) {
	tm := NewTabModel()

	tests := []struct {
		name     string
		tab      Tab
		expected Tab
	}{
		{"Set to VMs", TabVMs, TabVMs},
		{"Set to Configuration", TabConfiguration, TabConfiguration},
		{"Set to Power", TabPower, TabPower},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetActiveTab(tt.tab)
			if tm.GetActiveTab() != tt.expected {
				t.Errorf("Expected active tab %v, got %v", tt.expected, tm.GetActiveTab())
			}
		})
	}
}

func TestGetActiveTab(t *testing.T) {
	tm := NewTabModel()

	// Default should be TabVMs
	if tm.GetActiveTab() != TabVMs {
		t.Errorf("Expected default active tab to be TabVMs, got %v", tm.GetActiveTab())
	}

	// Change and verify
	tm.SetActiveTab(TabConfiguration)
	if tm.GetActiveTab() != TabConfiguration {
		t.Errorf("Expected active tab to be TabConfiguration, got %v", tm.GetActiveTab())
	}
}

func TestGetTabs(t *testing.T) {
	tm := NewTabModel()

	tabs := tm.GetTabs()
	if len(tabs) != 3 {
		t.Errorf("Expected 3 tabs, got %d", len(tabs))
	}

	expectedTabs := []Tab{TabVMs, TabConfiguration, TabPower}
	for i, expected := range expectedTabs {
		if tabs[i] != expected {
			t.Errorf("Expected tab %d to be %v, got %v", i, expected, tabs[i])
		}
	}
}

func TestSetTabContent(t *testing.T) {
	tm := NewTabModel()

	tests := []struct {
		name    string
		tab     Tab
		content string
	}{
		{"Set VMs content", TabVMs, "VM List Content"},
		{"Set Configuration content", TabConfiguration, "Configuration Content"},
		{"Set Power content", TabPower, "Power Management Content"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetTabContent(tt.tab, tt.content)
			if tm.GetTabContent(tt.tab) != tt.content {
				t.Errorf("Expected content '%s', got '%s'", tt.content, tm.GetTabContent(tt.tab))
			}
		})
	}
}

func TestGetTabContent(t *testing.T) {
	tm := NewTabModel()

	// Empty content initially
	if tm.GetTabContent(TabVMs) != "" {
		t.Errorf("Expected empty content initially, got '%s'", tm.GetTabContent(TabVMs))
	}

	// Set and verify
	tm.SetTabContent(TabVMs, "Test Content")
	if tm.GetTabContent(TabVMs) != "Test Content" {
		t.Errorf("Expected 'Test Content', got '%s'", tm.GetTabContent(TabVMs))
	}
}

func TestTabName(t *testing.T) {
	tests := []struct {
		name     string
		tab      Tab
		expected string
	}{
		{"VMs tab", TabVMs, "Start VM"},
		{"Configuration tab", TabConfiguration, "Configuration"},
		{"Power tab", TabPower, "Power"},
		{"Unknown tab", Tab(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TabName(tt.tab)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTabIndex(t *testing.T) {
	tests := []struct {
		name     string
		tab      Tab
		expected int
	}{
		{"VMs tab index", TabVMs, 0},
		{"Configuration tab index", TabConfiguration, 1},
		{"Power tab index", TabPower, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TabIndex(tt.tab)
			if result != tt.expected {
				t.Errorf("Expected index %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestTabFromIndex(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		expected Tab
	}{
		{"Index 0", 0, TabVMs},
		{"Index 1", 1, TabConfiguration},
		{"Index 2", 2, TabPower},
		{"Negative index", -1, TabVMs},
		{"Large index", 99, TabVMs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TabFromIndex(tt.index)
			if result != tt.expected {
				t.Errorf("Expected tab %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNextTab(t *testing.T) {
	tm := NewTabModel()

	tests := []struct {
		name     string
		current  Tab
		expected Tab
	}{
		{"From VMs", TabVMs, TabConfiguration},
		{"From Configuration", TabConfiguration, TabPower},
		{"From Power (wraps)", TabPower, TabVMs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetActiveTab(tt.current)
			result := tm.NextTab()
			if result != tt.expected {
				t.Errorf("Expected next tab %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestPrevTab(t *testing.T) {
	tm := NewTabModel()

	tests := []struct {
		name     string
		current  Tab
		expected Tab
	}{
		{"From VMs (wraps)", TabVMs, TabPower},
		{"From Configuration", TabConfiguration, TabVMs},
		{"From Power", TabPower, TabConfiguration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetActiveTab(tt.current)
			result := tm.PrevTab()
			if result != tt.expected {
				t.Errorf("Expected prev tab %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestRenderTabs(t *testing.T) {
	tm := NewTabModel()

	tests := []struct {
		name  string
		width int
	}{
		{"Narrow width", 40},
		{"Medium width", 80},
		{"Wide width", 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.RenderTabs(tt.width)

			// Should not be empty
			if result == "" {
				t.Error("RenderTabs returned empty string")
			}

			// Should contain all tab names
			if !strings.Contains(result, "Start VM") {
				t.Error("RenderTabs does not contain 'Start VM'")
			}
			if !strings.Contains(result, "Configuration") {
				t.Error("RenderTabs does not contain 'Configuration'")
			}
			if !strings.Contains(result, "Power") {
				t.Error("RenderTabs does not contain 'Power'")
			}
		})
	}
}

func TestRenderTabsActiveTab(t *testing.T) {
	tm := NewTabModel()

	// Test that active tab renders correctly with borders
	tm.SetActiveTab(TabVMs)
	resultVMs := tm.RenderTabs(80)

	tm.SetActiveTab(TabConfiguration)
	resultConfig := tm.RenderTabs(80)

	// Both should contain all tab names
	if !strings.Contains(resultVMs, "Start VM") {
		t.Error("RenderTabs does not contain 'Start VM'")
	}
	if !strings.Contains(resultVMs, "Configuration") {
		t.Error("RenderTabs does not contain 'Configuration'")
	}
	if !strings.Contains(resultConfig, "Configuration") {
		t.Error("RenderTabs does not contain 'Configuration'")
	}
	if !strings.Contains(resultConfig, "Start VM") {
		t.Error("RenderTabs does not contain 'Start VM'")
	}

	// Verify consistent line count (1 content + 1 separator = 2 lines)
	lines := strings.Split(resultVMs, "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in tab bar (content + separator), got %d", len(lines))
	}
}

func TestRenderTabsInactiveTab(t *testing.T) {
	tm := NewTabModel()

	// Set active tab to VMs
	tm.SetActiveTab(TabVMs)
	result := tm.RenderTabs(80)

	// Should contain all tab names
	if !strings.Contains(result, "Start VM") {
		t.Error("RenderTabs does not contain 'Start VM'")
	}
	if !strings.Contains(result, "Configuration") {
		t.Error("RenderTabs does not contain 'Configuration'")
	}
	if !strings.Contains(result, "Power") {
		t.Error("RenderTabs does not contain 'Power'")
	}
}

func TestRenderContent(t *testing.T) {
	tm := NewTabModel()

	// Set content for each tab
	tm.SetTabContent(TabVMs, "VM List Content")
	tm.SetTabContent(TabConfiguration, "Configuration Content")
	tm.SetTabContent(TabPower, "Power Content")

	tests := []struct {
		name     string
		tab      Tab
		width    int
		height   int
		expected string
	}{
		{"VMs content", TabVMs, 80, 20, "VM List Content"},
		{"Configuration content", TabConfiguration, 80, 20, "Configuration Content"},
		{"Power content", TabPower, 80, 20, "Power Content"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetActiveTab(tt.tab)
			result := tm.RenderContent(tt.width, tt.height)

			if result == "" {
				t.Error("RenderContent returned empty string")
			}

			if !strings.Contains(result, tt.expected) {
				t.Errorf("RenderContent does not contain expected content '%s'", tt.expected)
			}
		})
	}
}

func TestRenderContentEmpty(t *testing.T) {
	tm := NewTabModel()

	// No content set
	result := tm.RenderContent(80, 20)

	if result == "" {
		t.Error("RenderContent returned empty string")
	}

	// Should contain default message
	if !strings.Contains(result, "No content available") {
		t.Error("RenderContent does not contain default message for empty content")
	}
}

func TestRenderContentDimensions(t *testing.T) {
	tm := NewTabModel()
	tm.SetTabContent(TabVMs, "Test Content")

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"Small dimensions", 40, 10},
		{"Medium dimensions", 80, 20},
		{"Large dimensions", 120, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.RenderContent(tt.width, tt.height)

			if result == "" {
				t.Error("RenderContent returned empty string")
			}

			// Should contain the content
			if !strings.Contains(result, "Test Content") {
				t.Error("RenderContent does not contain expected content")
			}
		})
	}
}

func TestHandleKeyInput(t *testing.T) {
	tm := NewTabModel()

	tests := []struct {
		name           string
		key            string
		expectedTab    Tab
		expectedResult bool
	}{
		{"Tab key", "tab", TabConfiguration, true},
		{"Right arrow", "right", TabConfiguration, true},
		{"Shift+Tab", "shift+tab", TabPower, true},
		{"Left arrow", "left", TabPower, true},
		{"Unknown key", "x", TabVMs, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm.SetActiveTab(TabVMs)
			result := tm.HandleKeyInput(tt.key)

			if result != tt.expectedResult {
				t.Errorf("Expected result %v, got %v", tt.expectedResult, result)
			}

			if result && tm.GetActiveTab() != tt.expectedTab {
				t.Errorf("Expected active tab %v, got %v", tt.expectedTab, tm.GetActiveTab())
			}
		})
	}
}

func TestHandleKeyInputNavigation(t *testing.T) {
	tm := NewTabModel()

	// Test tab cycling
	tm.SetActiveTab(TabVMs)
	tm.HandleKeyInput("tab")
	if tm.GetActiveTab() != TabConfiguration {
		t.Errorf("Expected TabConfiguration after tab, got %v", tm.GetActiveTab())
	}

	tm.HandleKeyInput("tab")
	if tm.GetActiveTab() != TabPower {
		t.Errorf("Expected TabPower after second tab, got %v", tm.GetActiveTab())
	}

	tm.HandleKeyInput("tab")
	if tm.GetActiveTab() != TabVMs {
		t.Errorf("Expected TabVMs after third tab (wrap), got %v", tm.GetActiveTab())
	}
}

func TestHandleKeyInputReverseNavigation(t *testing.T) {
	tm := NewTabModel()

	// Test reverse tab cycling
	tm.SetActiveTab(TabVMs)
	tm.HandleKeyInput("shift+tab")
	if tm.GetActiveTab() != TabPower {
		t.Errorf("Expected TabPower after shift+tab, got %v", tm.GetActiveTab())
	}

	tm.HandleKeyInput("shift+tab")
	if tm.GetActiveTab() != TabConfiguration {
		t.Errorf("Expected TabConfiguration after second shift+tab, got %v", tm.GetActiveTab())
	}

	tm.HandleKeyInput("shift+tab")
	if tm.GetActiveTab() != TabVMs {
		t.Errorf("Expected TabVMs after third shift+tab (wrap), got %v", tm.GetActiveTab())
	}
}

func TestRenderTabsConsistency(t *testing.T) {
	tm := NewTabModel()
	tm.SetActiveTab(TabVMs)

	// Render multiple times with same width should produce same result
	result1 := tm.RenderTabs(80)
	result2 := tm.RenderTabs(80)

	if result1 != result2 {
		t.Error("RenderTabs should produce consistent results for same input")
	}
}

func TestRenderTabsAfterStateChange(t *testing.T) {
	tm := NewTabModel()

	// Initial render
	tm.SetActiveTab(TabVMs)
	result1 := tm.RenderTabs(80)

	// Change state and render again
	tm.SetActiveTab(TabConfiguration)
	result2 := tm.RenderTabs(80)

	// Both should contain all tab names (all tabs are always rendered)
	if !strings.Contains(result1, "Start VM") {
		t.Error("First render should contain 'Start VM'")
	}
	if !strings.Contains(result1, "Configuration") {
		t.Error("First render should contain 'Configuration'")
	}
	if !strings.Contains(result2, "Start VM") {
		t.Error("Second render should contain 'Start VM'")
	}
	if !strings.Contains(result2, "Configuration") {
		t.Error("Second render should contain 'Configuration'")
	}

	// Both should have consistent line count
	lines1 := strings.Split(result1, "\n")
	lines2 := strings.Split(result2, "\n")
	if len(lines1) != len(lines2) {
		t.Errorf("Tab bar height should be consistent: got %d vs %d lines", len(lines1), len(lines2))
	}
	if len(lines1) != 2 {
		t.Errorf("Expected 2 lines in tab bar (content + separator), got %d", len(lines1))
	}
}

func TestRenderTabsStyling(t *testing.T) {
	tm := NewTabModel()
	result := tm.RenderTabs(80)

	// Should contain all tab names
	if !strings.Contains(result, "Start VM") {
		t.Error("RenderTabs does not contain 'Start VM'")
	}
	if !strings.Contains(result, "Configuration") {
		t.Error("RenderTabs does not contain 'Configuration'")
	}
	if !strings.Contains(result, "Power") {
		t.Error("RenderTabs does not contain 'Power'")
	}

	// Should contain pipe separators
	if !strings.Contains(result, "|") {
		t.Error("RenderTabs should contain pipe separator '|'")
	}
}

func TestRenderContentStyling(t *testing.T) {
	tm := NewTabModel()
	tm.SetTabContent(TabVMs, "Test Content")
	result := tm.RenderContent(80, 20)

	// Should contain the content
	if !strings.Contains(result, "Test Content") {
		t.Error("RenderContent does not contain expected content")
	}

	// Should contain padding (spaces)
	if !strings.Contains(result, "  ") {
		t.Error("RenderContent does not contain padding")
	}
}

func TestTabNameWithBadge(t *testing.T) {
	tests := []struct {
		name     string
		tab      Tab
		count    int
		expected string
	}{
		{"VMs with count", TabVMs, 5, "Start VM (5)"},
		{"VMs with zero count", TabVMs, 0, "Start VM"},
		{"Config with count", TabConfiguration, 3, "Configuration (3)"},
		{"Power with zero count", TabPower, 0, "Power"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TabNameWithBadge(tt.tab, tt.count)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRenderTabsSeparators(t *testing.T) {
	tm := NewTabModel()
	result := tm.RenderTabs(80)

	// Should contain pipe separator
	if !strings.Contains(result, "|") {
		t.Error("RenderTabs does not contain pipe separator '|'")
	}

	// Should contain the positioned underline bar
	if !strings.Contains(result, "─") {
		t.Error("RenderTabs does not contain underline bar '─'")
	}

	// Should contain all tab names
	if !strings.Contains(result, "Start VM") {
		t.Error("RenderTabs does not contain 'Start VM'")
	}
	if !strings.Contains(result, "Configuration") {
		t.Error("RenderTabs does not contain 'Configuration'")
	}
	if !strings.Contains(result, "Power") {
		t.Error("RenderTabs does not contain 'Power'")
	}
}

func TestRenderTabsCentered(t *testing.T) {
	tm := NewTabModel()

	tests := []struct {
		name  string
		width int
	}{
		{"Width 40", 40},
		{"Width 80", 80},
		{"Width 120", 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.RenderTabs(tt.width)
			lines := strings.Split(result, "\n")
			if len(lines) != 2 {
				t.Fatalf("Expected 2 lines, got %d", len(lines))
			}

			contentLine := lines[0]

			// Visual width should equal requested width
			vw := lipgloss.Width(contentLine)
			if vw != tt.width {
				t.Errorf("Content line visual width = %d, want %d", vw, tt.width)
			}

			// Content should start with leading spaces (centered)
			leadingSpaces := len(contentLine) - len(strings.TrimLeft(contentLine, " "))
			if leadingSpaces == 0 {
				t.Error("Expected leading spaces for centered layout, got none")
			}

			// Tab names should still be present
			if !strings.Contains(contentLine, "Start VM") {
				t.Error("Centered content missing 'Start VM'")
			}
			if !strings.Contains(contentLine, "Configuration") {
				t.Error("Centered content missing 'Configuration'")
			}
			if !strings.Contains(contentLine, "Power") {
				t.Error("Centered content missing 'Power'")
			}
		})
	}
}

func TestRenderTabsUnderlineCentered(t *testing.T) {
	tm := NewTabModel()

	tests := []struct {
		name  string
		width int
	}{
		{"Width 40", 40},
		{"Width 80", 80},
		{"Width 120", 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.RenderTabs(tt.width)
			lines := strings.Split(result, "\n")
			if len(lines) != 2 {
				t.Fatalf("Expected 2 lines, got %d", len(lines))
			}

			contentLine := lines[0]
			underlineLine := lines[1]

			// Visual width should equal requested width for both lines
			vw := lipgloss.Width(underlineLine)
			if vw != tt.width {
				t.Errorf("Underline line visual width = %d, want %d", vw, tt.width)
			}

			// The underline bar should appear at the same horizontal offset as the
			// active tab name in the content line. Find where "─" starts in the
			// underline and where "VMs" (the default active tab) starts in content.
			ulIdx := strings.Index(underlineLine, "─")
			tabIdx := strings.Index(contentLine, "Start VM")
			if ulIdx == -1 {
				t.Fatal("Underline line has no '─' character")
			}
			if tabIdx == -1 {
				t.Fatal("Content line has no 'Start VM' for active tab")
			}
			if ulIdx != tabIdx {
				t.Errorf("Underline offset = %d, want %d (aligned with active tab)", ulIdx, tabIdx)
			}
		})
	}
}
