// Package tui provides the BubbleTea-based terminal user interface for DKVM Manager
package tui

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/models"
)

// Run starts the TUI application
// Parameters:
//   - debug: enable debug mode with verbose logging
//   - dryRun: show QEMU command without launching
//   - testRun: run a test scenario and exit (main_menu, vm_create)
func Run(debug bool, dryRun bool, testRun string) {
	// Validate terminal size before starting
	if !validateAndLogTerminalSize(debug) {
		fmt.Println("Warning: Terminal size below minimum 80x25. UI may not render correctly.")
		fmt.Println("Press Enter to continue anyway...")
		fmt.Scanln()
	}

	// Set debug mode in models package
	if debug {
		models.SetDebugMode(true)
		log.Printf("[DEBUG] Starting TUI with debug=%v, dryRun=%v, testRun=%q", debug, dryRun, testRun)
	}

	// Set dry-run mode in models package
	if dryRun {
		models.SetDryRunMode(true)
		log.Printf("[DRY-RUN] Dry-run mode enabled")
	}

	// Create the initial model
	m, err := models.NewMainModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize: %v\n", err)
		os.Exit(1)
	}

	// If test run is specified, run in test mode
	if testRun != "" {
		runTestMode(m, testRun, debug)
		return
	}

	// Create the program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(), // Use alternate screen buffer
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}

// runTestMode runs the application in non-interactive test mode
func runTestMode(m *models.MainModel, scenario string, debug bool) {
	if debug {
		log.Printf("[DEBUG] Running test scenario: %s", scenario)
	}

	// Execute the test scenario
	m.RunTestScenario(scenario)

	if debug {
		log.Printf("[DEBUG] Test scenario completed: %s", scenario)
	}
}

// validateAndLogTerminalSize checks terminal dimensions and logs warnings
func validateAndLogTerminalSize(debug bool) bool {
	width, height := GetTerminalSize()

	if width == 0 || height == 0 {
		log.Println("[WARN] Unable to determine terminal size, using defaults")
		return false
	}

	if debug {
		log.Printf("[DEBUG] Terminal size: %dx%d", width, height)
	}

	isValid := CheckMinimumSize(width, height)

	if !isValid {
		log.Printf("[WARN] Terminal size %dx%d is below minimum 80x25", width, height)
		log.Printf("[WARN] UI may not render correctly in small terminals")
	}

	return isValid
}