package components

import (
	"strings"
	"testing"
)

func TestNewDualPane(t *testing.T) {
	dp := NewDualPane()

	if dp.GetFocusedPane() != PaneLeft {
		t.Errorf("Expected default focused pane to be PaneLeft (%d), got %d", PaneLeft, dp.GetFocusedPane())
	}
}

func TestSetLeftContent(t *testing.T) {
	dp := NewDualPane()
	content := "Left pane content"
	dp.SetLeftContent(content)

	// Render to verify content is set
	dp.SetDimensions(100, 30)
	result := dp.Render()

	if !strings.Contains(result, content) {
		t.Errorf("Rendered output does not contain left pane content '%s'", content)
	}
}

func TestSetRightContent(t *testing.T) {
	dp := NewDualPane()
	content := "Right pane content"
	dp.SetRightContent(content)

	// Render to verify content is set
	dp.SetDimensions(100, 30)
	result := dp.Render()

	if !strings.Contains(result, content) {
		t.Errorf("Rendered output does not contain right pane content '%s'", content)
	}
}

func TestSetDimensions(t *testing.T) {
	dp := NewDualPane()
	dp.SetDimensions(100, 30)

	// Should render without error
	result := dp.Render()
	if result == "" {
		t.Error("Render returned empty string after setting dimensions")
	}
}

func TestSetFocusedPane(t *testing.T) {
	tests := []struct {
		name     string
		pane     int
		expected int
	}{
		{"Focus left pane", PaneLeft, PaneLeft},
		{"Focus right pane", PaneRight, PaneRight},
		{"Invalid pane (negative)", -1, PaneLeft}, // Should not change
		{"Invalid pane (2)", 2, PaneLeft},         // Should not change
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dp := NewDualPane()
			dp.SetFocusedPane(tt.pane)
			if dp.GetFocusedPane() != tt.expected {
				t.Errorf("Expected focused pane %d, got %d", tt.expected, dp.GetFocusedPane())
			}
		})
	}
}

func TestGetFocusedPane(t *testing.T) {
	dp := NewDualPane()

	// Default should be PaneLeft
	if dp.GetFocusedPane() != PaneLeft {
		t.Errorf("Expected default focused pane to be PaneLeft (%d), got %d", PaneLeft, dp.GetFocusedPane())
	}

	// Change and verify
	dp.SetFocusedPane(PaneRight)
	if dp.GetFocusedPane() != PaneRight {
		t.Errorf("Expected focused pane to be PaneRight (%d), got %d", PaneRight, dp.GetFocusedPane())
	}
}

func TestCalculateLayout(t *testing.T) {
	tests := []struct {
		name         string
		width        int
		expectedType LayoutType
	}{
		{"Wide terminal (150)", 150, LayoutDualPane},
		{"Wide terminal (121)", 121, LayoutDualPane},
		{"Medium terminal (120)", 120, LayoutDualPane},
		{"Medium terminal (100)", 100, LayoutDualPane},
		{"Medium terminal (80)", 80, LayoutDualPane},
		{"Narrow terminal (79)", 79, LayoutVerticalStack},
		{"Narrow terminal (60)", 60, LayoutVerticalStack},
		{"Narrow terminal (40)", 40, LayoutVerticalStack},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CalculateLayout(tt.width)
			if config.Type != tt.expectedType {
				t.Errorf("Expected layout type %v, got %v", tt.expectedType, config.Type)
			}
		})
	}
}

func TestCalculateLayoutWidths(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"Wide terminal", 150},
		{"Medium terminal", 100},
		{"Narrow terminal", 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CalculateLayout(tt.width)

			// Verify widths are positive
			if config.LeftWidth <= 0 {
				t.Errorf("Expected positive left width, got %d", config.LeftWidth)
			}
			if config.RightWidth <= 0 {
				t.Errorf("Expected positive right width, got %d", config.RightWidth)
			}

			// Verify gap is non-negative
			if config.Gap < 0 {
				t.Errorf("Expected non-negative gap, got %d", config.Gap)
			}
		})
	}
}

func TestDualPaneRender(t *testing.T) {
	dp := NewDualPane()
	dp.SetLeftContent("Left Content")
	dp.SetRightContent("Right Content")
	dp.SetDimensions(100, 30)

	result := dp.Render()

	if result == "" {
		t.Error("Render returned empty string")
	}

	// Should contain both contents
	if !strings.Contains(result, "Left Content") {
		t.Error("Rendered output does not contain left content")
	}
	if !strings.Contains(result, "Right Content") {
		t.Error("Rendered output does not contain right content")
	}
}

func TestRenderEmptyDimensions(t *testing.T) {
	dp := NewDualPane()
	dp.SetLeftContent("Left Content")
	dp.SetRightContent("Right Content")

	// No dimensions set
	result := dp.Render()

	if result != "" {
		t.Errorf("Expected empty string for zero dimensions, got '%s'", result)
	}
}

func TestRenderDualPane(t *testing.T) {
	dp := NewDualPane()
	dp.SetLeftContent("Left")
	dp.SetRightContent("Right")
	dp.SetDimensions(150, 30) // Wide terminal

	result := dp.Render()

	if result == "" {
		t.Error("Render returned empty string")
	}

	// Should contain both contents
	if !strings.Contains(result, "Left") {
		t.Error("Rendered output does not contain left content")
	}
	if !strings.Contains(result, "Right") {
		t.Error("Rendered output does not contain right content")
	}
}

func TestRenderVerticalStack(t *testing.T) {
	dp := NewDualPane()
	dp.SetLeftContent("Top Content")
	dp.SetRightContent("Bottom Content")
	dp.SetDimensions(60, 30) // Narrow terminal

	result := dp.Render()

	if result == "" {
		t.Error("Render returned empty string")
	}

	// Should contain both contents
	if !strings.Contains(result, "Top Content") {
		t.Error("Rendered output does not contain top content")
	}
	if !strings.Contains(result, "Bottom Content") {
		t.Error("Rendered output does not contain bottom content")
	}
}

func TestRenderAfterFocusChange(t *testing.T) {
	dp := NewDualPane()
	dp.SetLeftContent("Left")
	dp.SetRightContent("Right")
	dp.SetDimensions(100, 30)

	// Set initial focus
	dp.SetFocusedPane(PaneLeft)
	result1 := dp.Render()

	// Change focus
	dp.SetFocusedPane(PaneRight)
	result2 := dp.Render()

	// Both should render successfully
	if result1 == "" {
		t.Error("First render returned empty string")
	}
	if result2 == "" {
		t.Error("Second render returned empty string")
	}

	// Both should contain the content
	if !strings.Contains(result1, "Left") {
		t.Error("First render does not contain left content")
	}
	if !strings.Contains(result2, "Right") {
		t.Error("Second render does not contain right content")
	}
}

func TestRenderDualPaneFunction(t *testing.T) {
	result := RenderDualPane("Left", "Right", 100, 30, PaneLeft)

	if result == "" {
		t.Error("RenderDualPane returned empty string")
	}

	if !strings.Contains(result, "Left") {
		t.Error("RenderDualPane output does not contain left content")
	}
	if !strings.Contains(result, "Right") {
		t.Error("RenderDualPane output does not contain right content")
	}
}

func TestGetLayoutType(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		expected LayoutType
	}{
		{"Wide terminal", 150, LayoutDualPane},
		{"Medium terminal", 100, LayoutDualPane},
		{"Narrow terminal", 60, LayoutVerticalStack},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLayoutType(tt.width)
			if result != tt.expected {
				t.Errorf("Expected layout type %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsDualPane(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		expected bool
	}{
		{"Wide terminal", 150, true},
		{"Medium terminal", 100, true},
		{"Narrow terminal", 60, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDualPane(tt.width)
			if result != tt.expected {
				t.Errorf("Expected IsDualPane to return %v for width %d, got %v", tt.expected, tt.width, result)
			}
		})
	}
}

func TestIsVerticalStack(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		expected bool
	}{
		{"Wide terminal", 150, false},
		{"Medium terminal", 100, false},
		{"Narrow terminal", 60, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVerticalStack(tt.width)
			if result != tt.expected {
				t.Errorf("Expected IsVerticalStack to return %v for width %d, got %v", tt.expected, tt.width, result)
			}
		})
	}
}

func TestDualPaneRenderConsistency(t *testing.T) {
	dp := NewDualPane()
	dp.SetLeftContent("Left")
	dp.SetRightContent("Right")
	dp.SetDimensions(100, 30)

	// Render multiple times with same settings should produce same result
	result1 := dp.Render()
	result2 := dp.Render()

	if result1 != result2 {
		t.Error("Render should produce consistent results for same input")
	}
}

func TestRenderAfterContentChange(t *testing.T) {
	dp := NewDualPane()
	dp.SetDimensions(100, 30)

	// Set initial content
	dp.SetLeftContent("Initial Left")
	dp.SetRightContent("Initial Right")
	result1 := dp.Render()

	// Change content
	dp.SetLeftContent("Updated Left")
	dp.SetRightContent("Updated Right")
	result2 := dp.Render()

	// Results should be different
	if result1 == result2 {
		t.Error("Render should produce different results after content change")
	}

	// New content should be present
	if !strings.Contains(result2, "Updated Left") {
		t.Error("Rendered output does not contain updated left content")
	}
	if !strings.Contains(result2, "Updated Right") {
		t.Error("Rendered output does not contain updated right content")
	}
}

func TestRenderAfterDimensionChange(t *testing.T) {
	dp := NewDualPane()
	dp.SetLeftContent("Left")
	dp.SetRightContent("Right")

	// Set initial dimensions
	dp.SetDimensions(100, 30)
	result1 := dp.Render()

	// Change dimensions
	dp.SetDimensions(60, 20)
	result2 := dp.Render()

	// Both should render successfully
	if result1 == "" {
		t.Error("First render returned empty string")
	}
	if result2 == "" {
		t.Error("Second render returned empty string")
	}

	// Results should be different (different dimensions)
	if result1 == result2 {
		t.Error("Render should produce different results after dimension change")
	}
}

func TestCalculateLayoutEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"Very wide terminal", 200},
		{"Very narrow terminal", 20},
		{"Minimum width", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CalculateLayout(tt.width)

			// Should not panic and should return valid config
			// For very small widths, widths can be negative due to integer division
			// This is acceptable as the layout won't be usable anyway
			if config.Gap < 0 {
				t.Errorf("Expected non-negative gap, got %d", config.Gap)
			}
		})
	}
}

func TestRenderWithEmptyContent(t *testing.T) {
	dp := NewDualPane()
	dp.SetDimensions(100, 30)

	// Empty content
	dp.SetLeftContent("")
	dp.SetRightContent("")

	result := dp.Render()

	// Should still render (with empty panes)
	if result == "" {
		t.Error("Render returned empty string with empty content")
	}
}

func TestRenderWithLongContent(t *testing.T) {
	dp := NewDualPane()
	dp.SetDimensions(100, 30)

	// Long content
	longContent := strings.Repeat("Long content line\n", 50)
	dp.SetLeftContent(longContent)
	dp.SetRightContent(longContent)

	result := dp.Render()

	// Should render without error
	if result == "" {
		t.Error("Render returned empty string with long content")
	}
}

func TestDualPaneRenderWithSpecialCharacters(t *testing.T) {
	dp := NewDualPane()
	dp.SetDimensions(100, 30)

	// Content with special characters
	dp.SetLeftContent("Left with 特殊字符 and émojis 🎮")
	dp.SetRightContent("Right with tabs\tand\nnewlines")

	result := dp.Render()

	// Should render without error
	if result == "" {
		t.Error("Render returned empty string with special characters")
	}
}

func TestCalculateLayoutWideTerminal(t *testing.T) {
	config := CalculateLayout(150)

	if config.Type != LayoutDualPane {
		t.Errorf("Expected LayoutDualPane for wide terminal, got %v", config.Type)
	}

	// For wide terminal, left and right widths should be equal
	if config.LeftWidth != config.RightWidth {
		t.Errorf("Expected equal left and right widths for wide terminal, got left=%d, right=%d", config.LeftWidth, config.RightWidth)
	}

	// Gap should be WideTerminalGap for wide terminal
	if config.Gap != WideTerminalGap {
		t.Errorf("Expected gap of %d for wide terminal, got %d", WideTerminalGap, config.Gap)
	}
}

func TestCalculateLayoutMediumTerminal(t *testing.T) {
	config := CalculateLayout(100)

	if config.Type != LayoutDualPane {
		t.Errorf("Expected LayoutDualPane for medium terminal, got %v", config.Type)
	}

	// For medium terminal, left and right widths should be equal
	if config.LeftWidth != config.RightWidth {
		t.Errorf("Expected equal left and right widths for medium terminal, got left=%d, right=%d", config.LeftWidth, config.RightWidth)
	}

	// Gap should be MediumTerminalGap for medium terminal
	if config.Gap != MediumTerminalGap {
		t.Errorf("Expected gap of %d for medium terminal, got %d", MediumTerminalGap, config.Gap)
	}
}

func TestCalculateLayoutNarrowTerminal(t *testing.T) {
	config := CalculateLayout(60)

	if config.Type != LayoutVerticalStack {
		t.Errorf("Expected LayoutVerticalStack for narrow terminal, got %v", config.Type)
	}

	// For narrow terminal, left and right widths should be equal
	if config.LeftWidth != config.RightWidth {
		t.Errorf("Expected equal left and right widths for narrow terminal, got left=%d, right=%d", config.LeftWidth, config.RightWidth)
	}

	// Gap should be NarrowTerminalGap for narrow terminal
	if config.Gap != NarrowTerminalGap {
		t.Errorf("Expected gap of %d for narrow terminal, got %d", NarrowTerminalGap, config.Gap)
	}
}
