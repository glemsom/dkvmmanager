package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// BreadcrumbItem represents a single breadcrumb navigation item
type BreadcrumbItem struct {
	Label string // Display text
	View  string // View identifier
	Tab   int    // Tab index
}

// Breadcrumbs represents a breadcrumb navigation component
type Breadcrumbs struct {
	items        []BreadcrumbItem
	currentIndex int
}

// NewBreadcrumbs creates a new Breadcrumbs instance
func NewBreadcrumbs() *Breadcrumbs {
	return &Breadcrumbs{
		items:        []BreadcrumbItem{},
		currentIndex: -1,
	}
}

// AddItem adds a new breadcrumb item to the navigation
func (b *Breadcrumbs) AddItem(label, view string, tabIndex int) {
	item := BreadcrumbItem{
		Label: label,
		View:  view,
		Tab:   tabIndex,
	}
	b.items = append(b.items, item)
	b.currentIndex = len(b.items) - 1
}

// RemoveItem removes a breadcrumb item at the specified index
func (b *Breadcrumbs) RemoveItem(index int) {
	if index < 0 || index >= len(b.items) {
		return
	}

	b.items = append(b.items[:index], b.items[index+1:]...)

	// Adjust currentIndex if necessary
	if b.currentIndex >= len(b.items) {
		b.currentIndex = len(b.items) - 1
	}
}

// SetCurrent sets the current position in the breadcrumb navigation
func (b *Breadcrumbs) SetCurrent(index int) {
	if index < 0 || index >= len(b.items) {
		return
	}
	b.currentIndex = index
}

// GetCurrent returns the current breadcrumb item
func (b *Breadcrumbs) GetCurrent() BreadcrumbItem {
	if b.currentIndex < 0 || b.currentIndex >= len(b.items) {
		return BreadcrumbItem{}
	}
	return b.items[b.currentIndex]
}

// GetItems returns all breadcrumb items
func (b *Breadcrumbs) GetItems() []BreadcrumbItem {
	return b.items
}

// Clear removes all breadcrumb items
func (b *Breadcrumbs) Clear() {
	b.items = []BreadcrumbItem{}
	b.currentIndex = -1
}

// Len returns the number of breadcrumb items
func (b *Breadcrumbs) Len() int {
	return len(b.items)
}

// Render renders the breadcrumb navigation as a styled string
func (b *Breadcrumbs) Render() string {
	if len(b.items) == 0 {
		return ""
	}

	var parts []string

	for i, item := range b.items {
		if i < len(b.items)-1 {
			// Clickable breadcrumb (underlined, accent color)
			style := lipgloss.NewStyle().
				Foreground(styles.Colors.Primary).
				Underline(true)

			parts = append(parts, style.Render(item.Label))
			parts = append(parts, lipgloss.NewStyle().
				Foreground(styles.Colors.Muted).
				Render(" > "))
		} else {
			// Current location (bold, secondary color)
			style := lipgloss.NewStyle().
				Foreground(styles.Colors.Secondary).
				Bold(true)

			parts = append(parts, style.Render(item.Label))
		}
	}

	return strings.Join(parts, "")
}
