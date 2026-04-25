package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// Tab represents a tab identifier
type Tab int

// Tab constants for the three main tabs
const (
	TabVMs Tab = iota
	TabConfiguration
	TabPower
)

// TabModel represents the tab system component
type TabModel struct {
	tabs       []Tab
	activeTab  Tab
	tabContent map[Tab]string
	focused    bool   // Whether the tab bar is focused
}

// NewTabModel creates a new TabModel instance
func NewTabModel() *TabModel {
	return &TabModel{
		tabs:       []Tab{TabVMs, TabConfiguration, TabPower},
		activeTab:  TabVMs,
		tabContent: make(map[Tab]string),
	}
}

// SetActiveTab sets the active tab
func (t *TabModel) SetActiveTab(tab Tab) {
	t.activeTab = tab
}

// SetFocused sets whether the tab bar is focused
func (t *TabModel) SetFocused(focused bool) {
	t.focused = focused
}

// GetActiveTab returns the currently active tab
func (t *TabModel) GetActiveTab() Tab {
	return t.activeTab
}

// GetTabs returns all available tabs
func (t *TabModel) GetTabs() []Tab {
	return t.tabs
}

// SetTabContent sets the content for a specific tab
func (t *TabModel) SetTabContent(tab Tab, content string) {
	t.tabContent[tab] = content
}

// GetTabContent returns the content for a specific tab
func (t *TabModel) GetTabContent(tab Tab) string {
	return t.tabContent[tab]
}

// TabName returns the display name for a tab
func TabName(tab Tab) string {
	switch tab {
	case TabVMs:
		return "Start VM"
	case TabConfiguration:
		return "Configuration"
	case TabPower:
		return "Power"
	default:
		return "Unknown"
	}
}

// TabNameWithBadge returns the tab name with an optional count badge
func TabNameWithBadge(tab Tab, count int) string {
	name := TabName(tab)
	if count > 0 {
		return fmt.Sprintf("%s (%d)", name, count)
	}
	return name
}

// TabIndex returns the index of a tab
func TabIndex(tab Tab) int {
	return int(tab)
}

// TabFromIndex returns a tab from an index
func TabFromIndex(index int) Tab {
	if index < 0 || index > int(TabPower) {
		return TabVMs
	}
	return Tab(index)
}

// NextTab returns the next tab in the cycle
func (t *TabModel) NextTab() Tab {
	next := (t.activeTab + 1) % Tab(len(t.tabs))
	return next
}

// PrevTab returns the previous tab in the cycle
func (t *TabModel) PrevTab() Tab {
	prev := t.activeTab - 1
	if prev < 0 {
		prev = Tab(len(t.tabs) - 1)
	}
	return prev
}

// RenderTabs renders the tab bar with pipe separators.
// Active tab is bold and colored, inactive tabs are muted.
// A positioned underline bar highlights the active tab.
func (t *TabModel) RenderTabs(width int) string {
	// Ensure minimum width for proper rendering
	if width < 20 {
		width = 20
	}

	separator := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted).
		Background(styles.Colors.Background).
		Render(" | ")

	var contentRow string
	var underlineOffset int
	var activeTabWidth int
	offset := 0

	for i, tab := range t.tabs {
		tabName := TabName(tab)
		nameLen := len(tabName)

		if tab == t.activeTab {
			var tabStyle lipgloss.Style
			if t.focused {
				tabStyle = lipgloss.NewStyle().
					Foreground(styles.Colors.Primary).
					Background(styles.Colors.FocusedBackground).
					Bold(true)
			} else {
				// Enhanced information hierarchy: active tab gets subtle background even when unfocused
				tabStyle = lipgloss.NewStyle().
					Foreground(styles.Colors.Primary).
					Background(styles.Colors.Background).
					Bold(true)
			}
			contentRow += tabStyle.Render(tabName)
			underlineOffset = offset
			activeTabWidth = nameLen
		} else {
			contentRow += lipgloss.NewStyle().
				Foreground(styles.Colors.Muted).
				Background(styles.Colors.Background).
				Render(tabName)
		}

		offset += nameLen

		if i < len(t.tabs)-1 {
			contentRow += separator
			offset += 3 // " | " = 3 chars
		}
	}

	// Center the tab bar within the requested width and apply full-width background
	rowWidth := lipgloss.Width(contentRow)
	if rowWidth == 0 {
		rowWidth = offset
	}
	padding := 0
	if rowWidth < width {
		padding = (width - rowWidth) / 2
	}
	contentRow = strings.Repeat(" ", padding) + contentRow
	if paddedWidth := lipgloss.Width(contentRow); paddedWidth < width {
		contentRow += strings.Repeat(" ", width-paddedWidth)
	}

	// Apply background to the entire tab bar row
	contentRow = lipgloss.NewStyle().
		Background(styles.Colors.Background).
		Width(width).
		Render(contentRow)

	underlineOffset += padding
	underlineBar := lipgloss.NewStyle().
		Foreground(styles.Colors.Primary).
		Render(strings.Repeat("─", activeTabWidth))
	underline := strings.Repeat(" ", underlineOffset) + underlineBar
	if ulWidth := lipgloss.Width(underline); ulWidth < width {
		underline += strings.Repeat(" ", width-ulWidth)
	}

	return contentRow + "\n" + underline
}

// RenderContent renders the content for the active tab
func (t *TabModel) RenderContent(width, height int) string {
	content := t.tabContent[t.activeTab]
	if content == "" {
		content = "No content available for this tab"
	}

	// Apply panel styling for elevation and background
	return styles.LayeredPanelStyle().
		Width(width).
		Height(height).
		Render(content)
}

// HandleKeyInput handles keyboard input for tab navigation
func (t *TabModel) HandleKeyInput(key string) bool {
	switch key {
	case "tab", "right":
		t.SetActiveTab(t.NextTab())
		return true
	case "shift+tab", "left":
		t.SetActiveTab(t.PrevTab())
		return true
	}
	return false
}
