// DKVM Manager - Terminal-based virtual machine management application
// Built with Go and BubbleTea framework
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/glemsom/dkvmmanager/internal/tui"
)

var (
	debug          = flag.Bool("debug", false, "Enable debug mode with verbose logging to debug.log")
	dryRun         = flag.Bool("dry-run", false, "Dry-run mode: show QEMU command without launching")
	testRun        = flag.String("test", "", "Run test scenario and exit (main_menu, vm_create)")
	skipMountCheck = flag.Bool("skip-mount-check", false, "Skip mount point check (for testing without actual mount)")

	// debugLogFile holds the debug log file returned by tea.LogToFile when
	// debug mode is enabled. It is closed on shutdown via closeDebugLog.
	debugLogFile *os.File
)

func main() {
	flag.Parse()

	// Setup debug logging if enabled
	// Debug messages must go ONLY to the debug.log file, never to the terminal.
	// When -debug is set, BubbleTea disables AltScreen and renders on stderr
	// (see internal/tui/tui.go). If log.* output were to leak to stderr it would
	// appear on screen, corrupting the TUI display.
	if *debug {
		if err := setupDebugLog(); err != nil {
			// Fallback: discard all log output so it never reaches the terminal.
			log.SetOutput(io.Discard)
			log.SetPrefix("dkvmmanager ")
			fmt.Fprintf(os.Stderr, "Warning: could not create debug.log, discarding debug output: %v\n", err)
		} else {
			log.Println("[DEBUG] Debug mode enabled")
		}
		defer closeDebugLog()
	}

	// Run the TUI application with debug options
	tui.Run(*debug, *dryRun, *testRun, *skipMountCheck)
	os.Exit(0)
}

// setupDebugLog attempts to open a debug log file in a writable location.
// It tries several paths in order of preference so that log.* output is
// always redirected away from stderr (which the TUI shares in debug mode).
// The file is closed when main() returns (via defer in the caller).
func setupDebugLog() error {
	// Order of preference: CWD, /tmp, HOME
	candidates := []string{"debug.log"}

	home, err := os.UserHomeDir()
	if err == nil {
		candidates = append(candidates, filepath.Join(home, "debug.log"))
	}
	candidates = append(candidates, "/tmp/debug.log")

	for _, path := range candidates {
		f, err := tea.LogToFile(path, "dkvmmanager")
		if err == nil {
			debugLogFile = f
			return nil
		}
	}
	return fmt.Errorf("all log file paths failed")
}

// closeDebugLog closes the debug log file (if open) and clears the reference.
// Safe to call multiple times (idempotent).
func closeDebugLog() {
	if debugLogFile != nil {
		debugLogFile.Close()
		debugLogFile = nil
	}
}
