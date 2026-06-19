//go:build ignore

package main

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

func main() {
	fmt.Println("--- Simulating different TERM values (env override) ---")
	
	for _, t := range []string{"linux", "xterm-256color", "dumb"} {
		env := append(os.Environ(), "TERM="+t)
		p := colorprofile.Detect(os.Stdout, env)
		
		// Check what the actual escape codes are
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Background(lipgloss.Color("0")).
			Render("COLOR8_ON_BLACK")
			
		fmt.Printf("  TERM=%-16s -> profile=%s (%d)\n", t, p, p)
		fmt.Printf("    Color8 on Black: %q\n", style)
		
		// Also check color 7
		style7 := lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			Background(lipgloss.Color("0")).
			Render("COLOR7_ON_BLACK")
		fmt.Printf("    Color7 on Black: %q\n", style7)

		// Color 6 (cyan)
		style6 := lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Background(lipgloss.Color("0")).
			Render("COLOR6_ON_BLACK")
		fmt.Printf("    Color6 on Black: %q\n", style6)
	}
}
