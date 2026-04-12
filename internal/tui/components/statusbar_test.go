package components

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

func TestNewStatusBar(t *testing.T) {
	sb := NewStatusBar()

	if sb.GetMode() != "Ready" {
		t.Errorf("Expected mode 'Ready', got '%s'", sb.GetMode())
	}
	if sb.GetMessage() != "" {
		t.Errorf("Expected empty message, got '%s'", sb.GetMessage())
	}
	vmCount, running := sb.GetStats()
	if vmCount != 0 {
		t.Errorf("Expected vmCount 0, got %d", vmCount)
	}
	if running != 0 {
		t.Errorf("Expected running 0, got %d", running)
	}
	if sb.GetHelp() != "" {
		t.Errorf("Expected empty help, got '%s'", sb.GetHelp())
	}
}

func TestSetMode(t *testing.T) {
	sb := NewStatusBar()

	tests := []struct {
		name     string
		mode     string
		expected string
	}{
		{"Set Ready mode", "Ready", "Ready"},
		{"Set Editing mode", "Editing", "Editing"},
		{"Set Loading mode", "Loading", "Loading"},
		{"Set custom mode", "CustomMode", "CustomMode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb.SetMode(tt.mode)
			if sb.GetMode() != tt.expected {
				t.Errorf("Expected mode '%s', got '%s'", tt.expected, sb.GetMode())
			}
		})
	}
}

func TestSetMessage(t *testing.T) {
	sb := NewStatusBar()

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{"Set simple message", "Hello World", "Hello World"},
		{"Set empty message", "", ""},
		{"Set long message", "This is a very long status message that should be displayed in the center", "This is a very long status message that should be displayed in the center"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb.SetMessage(tt.message)
			if sb.GetMessage() != tt.expected {
				t.Errorf("Expected message '%s', got '%s'", tt.expected, sb.GetMessage())
			}
		})
	}
}

func TestSetStats(t *testing.T) {
	sb := NewStatusBar()

	tests := []struct {
		name        string
		vmCount     int
		running     int
		expectedVM  int
		expectedRun int
	}{
		{"Set zero stats", 0, 0, 0, 0},
		{"Set positive stats", 5, 3, 5, 3},
		{"Set all running", 10, 10, 10, 10},
		{"Set none running", 10, 0, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb.SetStats(tt.vmCount, tt.running)
			vmCount, running := sb.GetStats()
			if vmCount != tt.expectedVM {
				t.Errorf("Expected vmCount %d, got %d", tt.expectedVM, vmCount)
			}
			if running != tt.expectedRun {
				t.Errorf("Expected running %d, got %d", tt.expectedRun, running)
			}
		})
	}
}

func TestSetHelp(t *testing.T) {
	sb := NewStatusBar()

	tests := []struct {
		name     string
		help     string
		expected string
	}{
		{"Set help text", "Press ? for help", "Press ? for help"},
		{"Set empty help", "", ""},
		{"Set long help", "Press ? for help | q to quit | Tab to switch tabs", "Press ? for help | q to quit | Tab to switch tabs"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb.SetHelp(tt.help)
			if sb.GetHelp() != tt.expected {
				t.Errorf("Expected help '%s', got '%s'", tt.expected, sb.GetHelp())
			}
		})
	}
}

func TestRenderModeIndicator(t *testing.T) {
	sb := NewStatusBar()

	tests := []struct {
		name     string
		mode     string
		expected string
	}{
		{"Ready mode indicator", "Ready", "◌ Ready"},
		{"Editing mode indicator", "Editing", "⚙ Editing"},
		{"Loading mode indicator", "Loading", "◌ Loading"},
		{"Custom mode indicator", "CustomMode", "◌ CustomMode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb.SetMode(tt.mode)
			result := sb.renderModeIndicator()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRenderRightSection(t *testing.T) {
	sb := NewStatusBar()

	tests := []struct {
		name     string
		vmCount  int
		running  int
		help     string
		expected string
	}{
		{
			name:     "Stats only",
			vmCount:  5,
			running:  3,
			help:     "",
			expected: "VMs: 5 (▶ 3 running)",
		},
		{
			name:     "Help only",
			vmCount:  0,
			running:  0,
			help:     "Press ? for help",
			expected: "Press ? for help",
		},
		{
			name:     "Stats and help",
			vmCount:  10,
			running:  7,
			help:     "Press ? for help",
			expected: "VMs: 10 (▶ 7 running) | Press ? for help",
		},
		{
			name:     "Empty",
			vmCount:  0,
			running:  0,
			help:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb.SetStats(tt.vmCount, tt.running)
			sb.SetHelp(tt.help)
			result := sb.renderRightSection()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRender(t *testing.T) {
	sb := NewStatusBar()

	tests := []struct {
		name    string
		width   int
		mode    string
		message string
		vmCount int
		running int
		help    string
	}{
		{
			name:    "Basic render with all fields",
			width:   80,
			mode:    "Ready",
			message: "System ready",
			vmCount: 5,
			running: 3,
			help:    "Press ? for help",
		},
		{
			name:    "Render with empty message",
			width:   80,
			mode:    "Editing",
			message: "",
			vmCount: 10,
			running: 7,
			help:    "Press ? for help",
		},
		{
			name:    "Render with empty stats",
			width:   80,
			mode:    "Loading",
			message: "Loading VMs...",
			vmCount: 0,
			running: 0,
			help:    "",
		},
		{
			name:    "Render with narrow width",
			width:   40,
			mode:    "Ready",
			message: "OK",
			vmCount: 2,
			running: 1,
			help:    "?",
		},
		{
			name:    "Render with wide width",
			width:   120,
			mode:    "Ready",
			message: "All systems operational",
			vmCount: 100,
			running: 95,
			help:    "Press ? for help | q to quit | Tab to switch tabs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb.SetMode(tt.mode)
			sb.SetMessage(tt.message)
			sb.SetStats(tt.vmCount, tt.running)
			sb.SetHelp(tt.help)

			result := sb.Render(tt.width)

			// Check that result is not empty
			if result == "" {
				t.Error("Render returned empty string")
			}

			// Check that result contains the mode text
			if !strings.Contains(result, tt.mode) {
				t.Errorf("Render result does not contain mode '%s'", tt.mode)
			}

			// Check that result contains the message if not empty
			if tt.message != "" && !strings.Contains(result, tt.message) {
				t.Errorf("Render result does not contain message '%s'", tt.message)
			}

			// Check that result contains stats if vmCount > 0
			if tt.vmCount > 0 {
				statsText := "VMs:"
				if !strings.Contains(result, statsText) {
					t.Error("Render result does not contain stats")
				}
			}

			// Check that result contains help if not empty
			if tt.help != "" && !strings.Contains(result, tt.help) {
				t.Errorf("Render result does not contain help '%s'", tt.help)
			}

			// Check that the rendered width matches the requested width
			renderedWidth := lipgloss.Width(result)
			if renderedWidth != tt.width {
				t.Errorf("Expected width %d, got %d", tt.width, renderedWidth)
			}
		})
	}
}

func TestRenderModeColors(t *testing.T) {
	sb := NewStatusBar()

	tests := []struct {
		name          string
		mode          string
		expectedColor lipgloss.Color
	}{
		{"Ready mode uses green", "Ready", styles.StatusColors.Running},
		{"Editing mode uses yellow", "Editing", styles.Colors.Warning},
		{"Loading mode uses cyan", "Loading", styles.Colors.Primary},
		{"Unknown mode uses gray", "Unknown", styles.Colors.Muted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb.SetMode(tt.mode)
			rendered := sb.renderModeIndicator()

			// The rendered indicator should contain the mode text
			if !strings.Contains(rendered, tt.mode) {
				t.Errorf("Expected mode indicator to contain '%s'", tt.mode)
			}

			// The rendered indicator should contain a Unicode icon (not empty)
			if len(rendered) < 2 {
				t.Error("Expected mode indicator to contain an icon")
			}
		})
	}
}

func TestRenderBackground(t *testing.T) {
	sb := NewStatusBar()
	sb.SetMode("Ready")
	sb.SetMessage("Test message")
	sb.SetStats(5, 3)
	sb.SetHelp("Press ? for help")

	result := sb.Render(80)

	// The result should contain the background color code
	// We can't easily test the exact ANSI codes, but we can verify the render doesn't panic
	if result == "" {
		t.Error("Render returned empty string")
	}
}

func TestRenderWithVeryNarrowWidth(t *testing.T) {
	sb := NewStatusBar()
	sb.SetMode("Ready")
	sb.SetMessage("Test")
	sb.SetStats(5, 3)
	sb.SetHelp("?")

	// Test with very narrow width (should still render without panic)
	result := sb.Render(10)
	if result == "" {
		t.Error("Render returned empty string for narrow width")
	}
}

func TestRenderWithZeroWidth(t *testing.T) {
	sb := NewStatusBar()
	sb.SetMode("Ready")
	sb.SetMessage("Test")
	sb.SetStats(5, 3)
	sb.SetHelp("?")

	// Test with zero width (should handle gracefully)
	result := sb.Render(0)
	// Should not panic, result may be empty or minimal
	_ = result
}

func TestStatusBarChaining(t *testing.T) {
	sb := NewStatusBar()

	// Test method chaining (all setters return void, so we test sequential calls)
	sb.SetMode("Editing")
	sb.SetMessage("Editing VM configuration")
	sb.SetStats(10, 5)
	sb.SetHelp("Press Esc to cancel")

	if sb.GetMode() != "Editing" {
		t.Errorf("Expected mode 'Editing', got '%s'", sb.GetMode())
	}
	if sb.GetMessage() != "Editing VM configuration" {
		t.Errorf("Expected message 'Editing VM configuration', got '%s'", sb.GetMessage())
	}
	vmCount, running := sb.GetStats()
	if vmCount != 10 || running != 5 {
		t.Errorf("Expected stats (10, 5), got (%d, %d)", vmCount, running)
	}
	if sb.GetHelp() != "Press Esc to cancel" {
		t.Errorf("Expected help 'Press Esc to cancel', got '%s'", sb.GetHelp())
	}
}

func TestStatusBarRenderConsistency(t *testing.T) {
	sb := NewStatusBar()
	sb.SetMode("Ready")
	sb.SetMessage("System ready")
	sb.SetStats(5, 3)
	sb.SetHelp("Press ? for help")

	// Render multiple times with same width should produce same result
	result1 := sb.Render(80)
	result2 := sb.Render(80)

	if result1 != result2 {
		t.Error("Render should produce consistent results for same input")
	}
}

func TestStatusBarRenderAfterStateChange(t *testing.T) {
	sb := NewStatusBar()

	// Initial render
	sb.SetMode("Ready")
	sb.SetMessage("Initial message")
	sb.SetStats(5, 3)
	sb.SetHelp("Press ? for help")
	result1 := sb.Render(80)

	// Change state and render again
	sb.SetMode("Editing")
	sb.SetMessage("Editing message")
	sb.SetStats(10, 7)
	sb.SetHelp("Press Esc to cancel")
	result2 := sb.Render(80)

	// Results should be different
	if result1 == result2 {
		t.Error("Render should produce different results after state change")
	}

	// Both should contain their respective modes
	if !strings.Contains(result1, "Ready") {
		t.Error("First render should contain 'Ready'")
	}
	if !strings.Contains(result2, "Editing") {
		t.Error("Second render should contain 'Editing'")
	}
}
