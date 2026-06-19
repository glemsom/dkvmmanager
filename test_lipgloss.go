//go:build ignore

package main

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
)

func main() {
	term := os.Getenv("TERM")
	fmt.Printf("TERM=%s\n\n", term)

	// Test Muted (color 8, dark gray) on Background (color 0, black)
	mutedOnBlack := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("0")).
		Render("MUTED_TEXT")

	// Test Primary (color 6, dark cyan) on Background, bold
	primaryOnBlack := lipgloss.NewStyle().
		Foreground(lipgloss.Color("6")).
		Background(lipgloss.Color("0")).
		Bold(true).
		Render("PRIMARY_TEXT")

	// Test Foreground (color 7, light gray) on Background
	fgOnBlack := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("0")).
		Render("FOREGROUND_TEXT")

	fmt.Printf("Muted on Black: %q\n", mutedOnBlack)
	fmt.Printf("Primary on Black: %q\n", primaryOnBlack)
	fmt.Printf("Foreground on Black: %q\n", fgOnBlack)
	fmt.Println()

	fmt.Println("--- Visual output below ---")
	fmt.Println(mutedOnBlack)
	fmt.Println(primaryOnBlack)
	fmt.Println(fgOnBlack)
}
