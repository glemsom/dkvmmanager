package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

type VMListView struct {
	vms     []models.VM
	cursor  int
	width   int
	height  int
	focused bool
}

func NewVMListView(vms []models.VM, width, height int) *VMListView {
	return &VMListView{
		vms:     vms,
		cursor:  0,
		width:   width,
		height:  height,
		focused: true,
	}
}

func (v *VMListView) SetVMs(vms []models.VM) {
	v.vms = vms
	if v.cursor >= len(vms) && len(vms) > 0 {
		v.cursor = len(vms) - 1
	}
}

func (v *VMListView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *VMListView) SetCursor(cursor int) {
	if cursor >= 0 && cursor < len(v.vms) {
		v.cursor = cursor
	}
}

func (v *VMListView) Cursor() int {
	return v.cursor
}

func (v *VMListView) SetFocused(focused bool) {
	v.focused = focused
}

func (v *VMListView) IsFocused() bool {
	return v.focused
}

func (v *VMListView) MoveUp() bool {
	if v.cursor > 0 {
		v.cursor--
		return true
	}
	return false
}

func (v *VMListView) MoveDown() bool {
	if v.cursor < len(v.vms)-1 {
		v.cursor++
		return true
	}
	return false
}

func (v *VMListView) SelectedVM() *models.VM {
	if v.cursor >= 0 && v.cursor < len(v.vms) {
		return &v.vms[v.cursor]
	}
	return nil
}

func (v *VMListView) View() string {
	if len(v.vms) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Colors.Muted).
			Italic(true)
		return emptyStyle.Render("No virtual machines configured.\n\nPress 'n' to create a new VM.")
	}

	var lines []string

	normalCursor := " "
	selectedCursor := ">"
	if v.focused {
		selectedCursor = lipgloss.NewStyle().
			Foreground(styles.Colors.Primary).
			Bold(true).
			Render(">")
	} else {
		selectedCursor = lipgloss.NewStyle().
			Foreground(styles.Colors.Muted).
			Render(">")
	}

	for i, vm := range v.vms {
		status := "stopped"
		if vm.MAC != "" {
			status = "running"
		}
		statusIcon := styles.StatusIndicator(status)

		var nameStyle lipgloss.Style
		if i == v.cursor {
			// Enhanced selected item styling with focus awareness
			if v.focused {
				nameStyle = lipgloss.NewStyle().
					Foreground(styles.Colors.Primary).
					Bold(true).
					Background(styles.Colors.FocusedBackground)
			} else {
				nameStyle = lipgloss.NewStyle().
					Foreground(styles.Colors.Primary).
					Bold(true)
			}
		} else {
			nameStyle = styles.ListItemNormalStyle()
		}

		cursor := normalCursor
		if i == v.cursor {
			cursor = selectedCursor
		}

		line := cursor + " " + statusIcon + " " + nameStyle.Render(vm.Name)
		lines = append(lines, line)
	}

	// Apply panel styling for elevation and visual grouping
	var content string
	if v.focused {
		content = styles.ActiveLayeredPanelStyle().
			Width(v.width).
			Height(v.height).
			Render(strings.Join(lines, "\n"))
	} else {
		content = styles.LayeredPanelStyle().
			Width(v.width).
			Height(v.height).
			Render(strings.Join(lines, "\n"))
	}

	return content
}
