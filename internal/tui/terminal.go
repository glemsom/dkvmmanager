// Package tui provides the BubbleTea-based terminal user interface for DKVM Manager
package tui

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/sys/unix"
)

// Minimum dimensions for DKVM Manager (80x25 for Alpine console)
const (
	MinWidth  = 80
	MinHeight = 25
)

// TerminalSize represents terminal dimensions
type TerminalSize struct {
	Width  int
	Height int
}

// GetTerminalSize returns the current terminal size
// Returns (0, 0) if unable to determine size
func GetTerminalSize() (width, height int) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		// Try fallback: check COLUMNS/LINES environment variables
		width = getEnvInt("COLUMNS", 0)
		height = getEnvInt("LINES", 0)
		if width == 0 || height == 0 {
			return 0, 0
		}
		return width, height
	}
	width = int(ws.Col)
	height = int(ws.Row)
	// Validate the values - IoctlGetWinsize can return garbage in some environments
	// Typical terminal sizes are between 10x10 and 1000x1000, but we should be lenient
	// A value like 200x56 without reasonable bounds is suspicious
	if width < 10 || width > 1000 || height < 10 || height > 1000 {
		// Values seem unreasonable, try environment variables instead
		envWidth := getEnvInt("COLUMNS", 0)
		envHeight := getEnvInt("LINES", 0)
		if envWidth > 0 && envHeight > 0 && envWidth < 1000 && envHeight < 1000 {
			return envWidth, envHeight
		}
		// Return 0 to indicate failure, let caller use defaults
		return 0, 0
	}
	return width, height
}

// GetTerminalSizeWithFallback returns terminal size with fallback defaults
// If unable to detect, returns default 80x25
func GetTerminalSizeWithFallback() TerminalSize {
	width, height := GetTerminalSize()
	if width == 0 || height == 0 {
		return TerminalSize{Width: 80, Height: 25}
	}
	return TerminalSize{Width: width, Height: height}
}

// CheckMinimumSize checks if terminal meets minimum size requirements
// Returns true if terminal is large enough
func CheckMinimumSize(width, height int) bool {
	return width >= MinWidth && height >= MinHeight
}

// ValidateTerminalSize checks terminal dimensions and returns warnings if too small
func ValidateTerminalSize() (isValid bool, warnings []string) {
	width, height := GetTerminalSize()

	if width == 0 || height == 0 {
		warnings = append(warnings, "Unable to determine terminal size, assuming defaults")
		return false, warnings
	}

	if width < MinWidth {
		warnings = append(warnings, fmt.Sprintf("Terminal width %d is below minimum %d", width, MinWidth))
	}

	if height < MinHeight {
		warnings = append(warnings, fmt.Sprintf("Terminal height %d is below minimum %d", height, MinHeight))
	}

	isValid = len(warnings) == 0
	return
}

// getEnvInt retrieves an integer from an environment variable
func getEnvInt(name string, defaultValue int) int {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intVal
}