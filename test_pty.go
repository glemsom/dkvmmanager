//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

func main() {
	mutedOnBlack := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("0")).
		Render("MUTED_TEXT")

	primaryOnBlack := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Background(lipgloss.Color("0")).
		Bold(true).
		Render("PRIMARY_TEXT")

	fmt.Printf("Direct stdout output:\n")
	fmt.Printf("  Muted: %q\n", mutedOnBlack)
	fmt.Printf("  Primary: %q\n", primaryOnBlack)
	
	// Check the detected profile
	p := colorprofile.Detect(os.Stdout, os.Environ())
	fmt.Printf("  Detected profile: %s (%d)\n", p, p)
	
	// Can we get terminfo info?
	fmt.Printf("  TERM env: %s\n", os.Getenv("TERM"))

	// Check if we can open /dev/tty
	if _, err := os.Stat("/dev/tty"); err == nil {
		fmt.Printf("  /dev/tty exists\n")
	} else {
		fmt.Printf("  /dev/tty: %v\n", err)
	}

	// Try opening /dev/pts/* to simulate a terminal
	ptsDir, _ := os.ReadDir("/dev/pts/")
	fmt.Printf("  PTS entries: ")
	for _, e := range ptsDir {
		fmt.Printf("%s ", e.Name())
	}
	fmt.Println()
	
	// Test what happens when stdout IS a terminal
	// Spawn ourselves in a PTY
	cmd := exec.Command("sh", "-c", `echo "child TERM=$TERM"`)
	cmd.Env = append(os.Environ(), "TERM=linux")
	out, _ := cmd.CombinedOutput()
	fmt.Printf("  Subprocess: %s", strings.TrimSpace(string(out)))
}
