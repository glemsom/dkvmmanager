// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VMDeleteModel handles VM deletion confirmation
type VMDeleteModel struct {
	// VM manager
	vmManager *vm.Manager

	// VM to delete
	vm *models.VM

	// Selected option (0 = No, 1 = Yes)
	selectedIndex int

	// Status message
	statusMessage string

	// Debug mode flag
	debugMode bool
}

// NewVMDeleteModel creates a new VM delete model
func NewVMDeleteModel(vmManager *vm.Manager, vmID string) (*VMDeleteModel, error) {
	// Get VM details
	vm, err := vmManager.GetVM(vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}

	if debugMode {
		log.Printf("[DEBUG] NewVMDeleteModel created for VM: %s (ID: %s)", vm.Name, vm.ID)
	}

	return &VMDeleteModel{
		vmManager:     vmManager,
		vm:            vm,
		selectedIndex: 0,
		debugMode:     debugMode,
	}, nil
}

// Init initializes the model
func (m *VMDeleteModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages
func (m *VMDeleteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.debugMode {
		log.Printf("[DEBUG] VMDeleteModel.Update received: %T", msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *VMDeleteModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		if m.selectedIndex < 1 {
			m.selectedIndex++
		}

	case "enter", " ":
		if m.selectedIndex == 0 {
			// No - cancel deletion
			if m.debugMode {
				log.Printf("[DEBUG] VMDeleteModel: Deletion cancelled (No selected)")
			}
			return m, func() tea.Msg {
				return ViewChangeMsg{View: ViewConfigMenu}
			}
		} else {
			// Yes - confirm deletion
			if m.debugMode {
				log.Printf("[DEBUG] VMDeleteModel: Confirming deletion of VM: %s (ID: %s)", m.vm.Name, m.vm.ID)
			}
			err := m.vmManager.DeleteVM(m.vm.ID)
			if err != nil {
				m.statusMessage = "Error deleting VM: " + err.Error()
				if m.debugMode {
					log.Printf("[DEBUG] VMDeleteModel: Deletion failed: %v", err)
				}
				return m, nil
			}
			if m.debugMode {
				log.Printf("[DEBUG] VMDeleteModel: VM deleted successfully")
			}
			return m, func() tea.Msg {
				return VMDeletedMsg{VMName: m.vm.Name, VMID: m.vm.ID}
			}
		}
	}

	return m, nil
}

// View returns the view for the model
func (m *VMDeleteModel) View() string {
	warningStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Error).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Primary).
		Bold(true)

	itemStyle := lipgloss.NewStyle()

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted)

	var output string
	output += warningStyle.Render("WARNING: This action cannot be undone!") + "\n"
	output += "\n"
	output += fmt.Sprintf("Are you sure you want to delete VM '%s' (ID: %s)?\n", m.vm.Name, m.vm.ID)
	output += "\n"

	options := []string{"No", "Yes"}
	for i, option := range options {
		prefix := "  "
		if i == m.selectedIndex {
			prefix = "> "
			output += selectedStyle.Render(prefix+option) + "\n"
		} else {
			output += itemStyle.Render(prefix+option) + "\n"
		}
	}

	output += "\n"
	output += helpStyle.Render("↑/↓ Navigate  Space/Enter Select  ESC Cancel")

	if m.statusMessage != "" {
		output += "\n\n" + m.statusMessage
	}

	return output
}
