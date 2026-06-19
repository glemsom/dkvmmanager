//go:build ignore

package main

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

func main() {
	term := os.Getenv("TERM")
	fmt.Printf("TERM=%s\n", term)

	// Detect color profile
	profile := colorprofile.Detect(os.Stdout, os.Environ())
	fmt.Printf("Detected color profile: %v\n", profile)
	fmt.Println()

	mutedOnBlack := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("0")).
		Render("MUTED_TEXT")

	fmt.Printf("Muted on Black: %q\n", mutedOnBlack)
}
