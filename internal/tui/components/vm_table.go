package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// VMTable represents a table for displaying VM information
type VMTable struct {
	table table.Model
	vms   []models.VM
}

// NewVMTable creates a new VM table
func NewVMTable(vms []models.VM, width, height int) *VMTable {
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "ID", Width: 20},
		{Title: "Disks", Width: 8},
		{Title: "MAC", Width: 18},
		{Title: "TPM", Width: 6},
	}

	rows := vmToRows(vms)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	// Custom styles
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(true).
		Foreground(styles.Colors.Primary)
	s.Selected = s.Selected.
		Foreground(styles.Colors.Primary).
		Bold(true).
		Background(lipgloss.Color("236"))
	s.Cell = s.Cell.
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("235"))
	t.SetStyles(s)

	return &VMTable{
		table: t,
		vms:   vms,
	}
}

// vmToRows converts VMs to table rows
func vmToRows(vms []models.VM) []table.Row {
	rows := make([]table.Row, len(vms))
	for i, vm := range vms {
		mac := vm.MAC
		if mac == "" {
			mac = "-"
		}

		tpm := "No"
		if vm.TPMEnabled {
			tpm = "Yes"
		}

		rows[i] = table.Row{
			vm.Name,
			vm.ID,
			fmt.Sprintf("%d", len(vm.HardDisks)),
			mac,
			tpm,
		}
	}
	return rows
}

// SetSize updates the table dimensions
func (v *VMTable) SetSize(width, height int) {
	v.table.SetWidth(width)
	v.table.SetHeight(height)
}

// Update handles table messages
func (v *VMTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	v.table, cmd = v.table.Update(msg)
	return v, cmd
}

// Init initializes the VM table
func (v *VMTable) Init() tea.Cmd {
	return nil
}

// View renders the table with a ">" cursor indicator on the selected row,
// matching the visual convention used by the rest of the TUI (main menu, config menu).
func (v *VMTable) View() string {
	cursor := v.table.Cursor()
	rows := v.table.Rows()

	if cursor >= 0 && cursor < len(rows) {
		// Deep copy rows to avoid mutating internal state
		copied := make([]table.Row, len(rows))
		for i, row := range rows {
			copied[i] = make(table.Row, len(row))
			copy(copied[i], row)
		}
		copied[cursor][0] = "> " + copied[cursor][0]
		v.table.SetRows(copied)
		defer v.table.SetRows(rows) // restore original after render
	}

	return v.table.View()
}

// SelectedVM returns the currently selected VM
func (v *VMTable) SelectedVM() *models.VM {
	if len(v.vms) == 0 {
		return nil
	}
	index := v.table.Cursor()
	if index < 0 || index >= len(v.vms) {
		return nil
	}
	return &v.vms[index]
}

// SetVMs updates the VM list
func (v *VMTable) SetVMs(vms []models.VM) {
	v.vms = vms
	rows := vmToRows(vms)
	v.table.SetRows(rows)
}

// Cursor returns the current cursor position
func (v *VMTable) Cursor() int {
	return v.table.Cursor()
}
