package tui

import (
	"os"
	"testing"
)

// Since we can't easily mock unix.IoctlGetWinsize in tests, we test the fallback
// behavior through environment variables

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name      string
		envVar    string
		envValue  string
		defaultVal int
		expected  int
	}{
		{"empty env", "TEST_VAR", "", 10, 10},
		{"valid int", "TEST_VAR", "42", 10, 42},
		{"invalid int", "TEST_VAR", "abc", 10, 10},
		{"negative int", "TEST_VAR", "-5", 10, -5},
		{"zero", "TEST_VAR", "0", 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			} else {
				os.Unsetenv(tt.envVar)
			}

			result := getEnvInt(tt.envVar, tt.defaultVal)
			if result != tt.expected {
				t.Errorf("getEnvInt(%q, %d) = %d, want %d", tt.envVar, tt.defaultVal, result, tt.expected)
			}
		})
	}
}

func TestCheckMinimumSize(t *testing.T) {
	tests := []struct {
		width     int
		height    int
		expected  bool
	}{
		{80, 25, true},
		{100, 30, true},
		{80, 24, false},
		{79, 25, false},
		{0, 0, false},
		{79, 24, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := CheckMinimumSize(tt.width, tt.height)
			if result != tt.expected {
				t.Errorf("CheckMinimumSize(%d, %d) = %v, want %v", tt.width, tt.height, result, tt.expected)
			}
		})
	}
}

func TestTerminalSizeConstants(t *testing.T) {
	if MinWidth != 80 {
		t.Errorf("MinWidth = %d, want 80", MinWidth)
	}
	if MinHeight != 25 {
		t.Errorf("MinHeight = %d, want 25", MinHeight)
	}
}