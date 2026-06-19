package styles

import (
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
