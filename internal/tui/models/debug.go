// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"log"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// runLBUCommit executes lbu commit asynchronously and returns the result
func runLBUCommit() tea.Cmd {
	return func() tea.Msg {
		if dryRunMode {
			return LBUCommitMsg{
				Output:  "Would execute: lbu commit",
				Success: true,
			}
		}
		cmd := exec.Command("lbu", "commit")
		output, err := cmd.CombinedOutput()
		return LBUCommitMsg{
			Output:  string(output),
			Success: err == nil,
		}
	}
}

// RunTestScenario runs a test scenario in non-interactive mode
// This is used by the debug agent to test specific application states
func (m *MainModel) RunTestScenario(scenario string) {
	if m.debugMode {
		log.Printf("[DEBUG] Running test scenario: %s", scenario)
	}

	switch scenario {
	case "main_menu":
		// Just stay in main menu - useful for testing initial state
		fmt.Println("Test scenario: main_menu")
		fmt.Printf("Current view: %s\n", m.currentView)
		fmt.Printf("Menu items: %d\n", len(m.menuItems))
		for i, item := range m.menuItems {
			fmt.Printf("  [%d] %s (%s)\n", i, item.Title, item.Type)
		}

	case "vm_create":
		// Navigate to VM creation view
		m.currentView = ViewConfigMenu
		m.configSelectedIndex = 0 // "Add new VM"
		// Simulate pressing enter to go to VM create
		m.vmCreateModel = NewVMCreateModel(m.vmManager)
		m.vmCreateModel.form.SetSize(m.windowWidth-4, m.contentHeight()-2)
		m.currentView = ViewVMCreate
		fmt.Println("Test scenario: vm_create")
		fmt.Printf("Current view: %s\n", m.currentView)

	default:
		fmt.Printf("Unknown test scenario: %s\n", scenario)
		fmt.Println("Available scenarios: main_menu, vm_create")
	}

	if m.debugMode {
		log.Printf("[DEBUG] Test scenario completed: %s", scenario)
	}
}
