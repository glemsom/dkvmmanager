package styles

import (
	"strings"
	"testing"
)

func TestGetModeIcon_Unicode(t *testing.T) {
	tests := []struct {
		mode     string
		expected string
	}{
		{"Ready", "◌"},
		{"Editing", "⚙"},
		{"Loading", "◌"},
		{"Running", "▶"},
		{"Stopped", "■"},
		{"Error", "⚠"},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			t.Setenv("TERM", "xterm-256color")
			got := GetModeIcon(tt.mode)
			if got != tt.expected {
				t.Errorf("GetModeIcon(%q) = %q, want %q (TERM=xterm-256color)", tt.mode, got, tt.expected)
			}
		})
	}
}

func TestGetModeIcon_Ascii(t *testing.T) {
	tests := []struct {
		mode     string
		expected string
	}{
		{"Ready", "-"},
		{"Editing", "*"},
		{"Loading", "-"},
		{"Running", ">"},
		{"Stopped", "#"},
		{"Error", "!"},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			t.Setenv("TERM", "linux")
			got := GetModeIcon(tt.mode)
			if got != tt.expected {
				t.Errorf("GetModeIcon(%q) = %q, want %q (TERM=linux)", tt.mode, got, tt.expected)
			}
		})
	}
}

func TestGetModeIcon_UnknownMode(t *testing.T) {
	t.Setenv("TERM", "linux")
	got := GetModeIcon("UnknownMode")
	if got == "" {
		t.Error("GetModeIcon('UnknownMode') should return a fallback, not empty")
	}
}

func TestGetStatusIndicator_Unicode(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"running", "●"},
		{"stopped", "○"},
		{"error", "●"},
		{"unknown", "○"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			t.Setenv("TERM", "xterm-256color")
			got := StatusIndicator(tt.status)
			if got == "" {
				t.Errorf("StatusIndicator(%q) should not be empty", tt.status)
			}
		})
	}
}

func TestGetStatusIndicator_Ascii(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"running", "*"},
		{"stopped", "o"},
		{"error", "*"},
		{"unknown", "o"},
	}
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			t.Setenv("TERM", "linux")
			got := StatusIndicator(tt.status)
			if got == "" {
				t.Errorf("StatusIndicator(%q) should not be empty", tt.status)
			}
		})
	}
}

func TestBorderStyle_Ascii(t *testing.T) {
	t.Setenv("TERM", "linux")
	style := BorderStyle()
	rendered := style.Render("Test")
	if !strings.Contains(rendered, "+") {
		t.Errorf("BorderStyle() with TERM=linux should use ASCII border with '+', got: %q", rendered)
	}
	if strings.Contains(rendered, "╭") || strings.Contains(rendered, "╮") {
		t.Errorf("BorderStyle() with TERM=linux should NOT use Unicode border chars, got: %q", rendered)
	}
}

func TestBorderStyle_Unicode(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	style := BorderStyle()
	rendered := style.Render("Test")
	if !strings.Contains(rendered, "╭") {
		t.Errorf("BorderStyle() with TERM=xterm-256color should use Unicode border, got: %q", rendered)
	}
}

func TestActiveBorderStyle_Ascii(t *testing.T) {
	t.Setenv("TERM", "linux")
	style := ActiveBorderStyle()
	rendered := style.Render("Test")
	if !strings.Contains(rendered, "+") {
		t.Errorf("ActiveBorderStyle() with TERM=linux should use ASCII border with '+', got: %q", rendered)
	}
	if strings.Contains(rendered, "╭") || strings.Contains(rendered, "╮") {
		t.Errorf("ActiveBorderStyle() with TERM=linux should NOT use Unicode border chars, got: %q", rendered)
	}
}

func TestActiveBorderStyle_Unicode(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	style := ActiveBorderStyle()
	rendered := style.Render("Test")
	if !strings.Contains(rendered, "╭") {
		t.Errorf("ActiveBorderStyle() with TERM=xterm-256color should use Unicode border, got: %q", rendered)
	}
}

func TestInputStyle_Ascii(t *testing.T) {
	t.Setenv("TERM", "linux")
	style := InputStyle()
	rendered := style.Render("Test")
	if !strings.Contains(rendered, "+") {
		t.Errorf("InputStyle() with TERM=linux should use ASCII border with '+', got: %q", rendered)
	}
}

func TestInputStyle_Unicode(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	style := InputStyle()
	rendered := style.Render("Test")
	if !strings.Contains(rendered, "┌") {
		t.Errorf("InputStyle() with TERM=xterm-256color should use NormalBorder, got: %q", rendered)
	}
}

func TestTooltipStyle_Ascii(t *testing.T) {
	t.Setenv("TERM", "linux")
	style := TooltipStyle()
	rendered := style.Render("Test")
	if !strings.Contains(rendered, "+") {
		t.Errorf("TooltipStyle() with TERM=linux should use ASCII border with '+', got: %q", rendered)
	}
}

func TestTooltipStyle_Unicode(t *testing.T) {
	t.Setenv("TERM", "xterm-256color")
	style := TooltipStyle()
	rendered := style.Render("Test")
	if !strings.Contains(rendered, "╭") {
		t.Errorf("TooltipStyle() with TERM=xterm-256color should use RoundedBorder, got: %q", rendered)
	}
}

func TestModalStyle_Ascii(t *testing.T) {
	t.Setenv("TERM", "linux")
	style := ModalStyle()
	rendered := style.Render("Test")
	if !strings.Contains(rendered, "+") {
		t.Errorf("ModalStyle() with TERM=linux should use ASCII border with '+', got: %q", rendered)
	}
}
