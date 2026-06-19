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
	fmt.Printf("TERM=%s\n\n", term)

	// Check what profile lipgloss detects
	profile := colorprofile.Detect(os.Stdout, os.Environ())
	fmt.Printf("ColorProfile: %s (%d)\n", profile, profile)

	// Test rendering with explicit different styles
	muted8 := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("0")).
		Render("COLOR_8")

	muted7 := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("0")).
		Render("COLOR_7")

	muted6 := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Background(lipgloss.Color("0")).
		Render("COLOR_6")

	fmt.Printf("Color 8 (dark gray) on 0 (black): %q\n", muted8)
	fmt.Printf("Color 7 (white) on 0 (black):     %q\n", muted7)
	fmt.Printf("Color 6 (cyan) on 0 (black):      %q\n", muted6)
}
