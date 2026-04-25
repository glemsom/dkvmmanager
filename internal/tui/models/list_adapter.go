package models

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// MenuItemAdapter wraps MenuItem to implement list.Item interface
type MenuItemAdapter struct {
	MenuItem
}

func (a MenuItemAdapter) FilterValue() string { return a.MenuItem.Title }
func (a MenuItemAdapter) Title() string       { return a.MenuItem.Title }
func (a MenuItemAdapter) Description() string { return menuItemDescription(a.MenuItem.Type) }

// menuItemDescription returns a human-readable description for a menu item type
func menuItemDescription(itemType string) string {
	switch itemType {
	case "VM":
		return "Virtual Machine"
	case "INT_CONFIG":
		return "Configuration"
	case "INT_POWEROFF":
		return "Power management"
	case "INT_POWER_REBOOT":
		return "Restart the host system"
	case "INT_POWER_OFF":
		return "Shut down the host system"
	case "INT_SHELL":
		return "Shell access"
	default:
		return itemType
	}
}

// MenuItemDelegate renders menu items with custom styling
type MenuItemDelegate struct{}

func (d MenuItemDelegate) Height() int  { return 1 }
func (d MenuItemDelegate) Spacing() int { return 0 }
func (d MenuItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}
func (d MenuItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(MenuItemAdapter)
	if !ok {
		return
	}

	str := "  " + item.MenuItem.Title
	var style lipgloss.Style

	if item.MenuItem.Disabled {
		style = styles.ListItemDisabledStyle().
			Background(styles.Colors.Background).
			Width(m.Width())
	} else if index == m.Index() {
		// ListItemSelectedStyle has PaddingLeft(1), we add ">  " prefix for total 4-space equivalent (matching unselected)
		style = styles.ListItemSelectedStyle().
			Width(m.Width())
		str = ">  " + item.MenuItem.Title
	} else {
		// Use muted color for non-selected items to create visual hierarchy
		style = lipgloss.NewStyle().
			Foreground(styles.Colors.Muted).
			Background(styles.Colors.Background).
			PaddingLeft(2).
			Width(m.Width())
	}

	fmt.Fprint(w, style.Render(str))
}

// VMListAdapter wraps models.VM to implement list.Item interface
type VMListAdapter struct {
	models.VM
}

func (a VMListAdapter) FilterValue() string { return a.VM.Name }
func (a VMListAdapter) Title() string       { return a.VM.Name }
func (a VMListAdapter) Description() string {
	return fmt.Sprintf("ID: %s  Disks: %d", a.VM.ID, len(a.VM.HardDisks)+len(a.VM.CDROMs))
}

// VMListItemDelegate renders VM list items with custom styling
type VMListItemDelegate struct{}

func (d VMListItemDelegate) Height() int  { return 1 }
func (d VMListItemDelegate) Spacing() int { return 0 }
func (d VMListItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}
func (d VMListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(VMListAdapter)
	if !ok {
		return
	}

	var str string
	var style lipgloss.Style

	if index == m.Index() {
		style = styles.ListItemSelectedStyle().
			Width(m.Width())
		str = ">  " + item.VM.Name
	} else {
		// Use muted color for non-selected items to create visual hierarchy
		style = lipgloss.NewStyle().
			Foreground(styles.Colors.Muted).
			Background(styles.Colors.Background).
			PaddingLeft(2).
			Width(m.Width())
		str = "  " + item.VM.Name
	}

	fmt.Fprint(w, style.Render(str))
}

// buildVMListAdapter converts VMs to list.Item slice
func buildVMListAdapter(vms []models.VM) []list.Item {
	listItems := make([]list.Item, len(vms))
	for i, vm := range vms {
		listItems[i] = VMListAdapter{VM: vm}
	}
	return listItems
}
