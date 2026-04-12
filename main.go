// DKVM Manager - Terminal-based virtual machine management application
// Built with Go and BubbleTea framework
package main

import (
	"flag"
	"log"
	"os"

	"github.com/glemsom/dkvmmanager/internal/tui"
)

var (
	debug   = flag.Bool("debug", false, "Enable debug mode with verbose logging to debug.log")
	dryRun  = flag.Bool("dry-run", false, "Dry-run mode: show QEMU command without launching")
	testRun = flag.String("test", "", "Run test scenario and exit (main_menu, vm_create)")
)

func main() {
	flag.Parse()

	// Setup debug logging if enabled
	if *debug {
		f, err := os.Create("debug.log")
		if err != nil {
			log.Printf("Warning: could not create debug.log: %v", err)
		} else {
			log.SetOutput(f)
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Println("[DEBUG] Debug mode enabled")
		}
	}

	// Run the TUI application with debug options
	tui.Run(*debug, *dryRun, *testRun)
	os.Exit(0)
}
