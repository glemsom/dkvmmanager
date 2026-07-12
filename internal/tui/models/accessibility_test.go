// Package models provides BubbleTea models for DKVM Manager.
// This file adds accessibility smoke tests that detect invisible text
// (foreground color == background color in rendered output).
package models

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
)

// ---------------------------------------------------------------------------
// ANSI SGR parser — tracks foreground/background color state through rendered
// output and flags any character span where both colors are set to the same
// ANSI 0–15 value.
// ---------------------------------------------------------------------------

// ansiState tracks the current foreground and background color.
// nil means "not set" (terminal default — no invisibility risk).
type ansiState struct {
	fg *int // ANSI 0–15
	bg *int
}

// reset clears both fg and bg to nil.
func (s *ansiState) reset() {
	s.fg = nil
	s.bg = nil
}

func (s *ansiState) setFG(v int)    { s.fg = &v }
func (s *ansiState) setBG(v int)    { s.bg = &v }
func (s *ansiState) invisible() bool { return s.fg != nil && s.bg != nil && *s.fg == *s.bg }

// invisibleSpan records a region of invisible text in the rendered output.
type invisibleSpan struct {
	line    int    // 1-based
	col     int    // byte offset in line
	text    string // the invisible text
	fgColor int    // the offending ANSI color index
}

// String returns a user-readable report line.
func (s invisibleSpan) String() string {
	return fmt.Sprintf("line %d col %d: fg=bg=%d %q", s.line, s.col, s.fgColor, s.text)
}

// checkInvisibleText scans rendered output and reports any text spans where
// the foreground and background colors match (making the text invisible).
// It uses t.Error to report findings.
func checkInvisibleText(t *testing.T, label, rendered string) {
	t.Helper()

	if rendered == "" {
		t.Errorf("%s: empty rendered output — nothing to check", label)
		return
	}

	spans := findInvisibleTextSpans(rendered)
	for _, s := range spans {
		t.Errorf("%s %s", label, s.String())
	}
}

// findInvisibleTextSpans returns all invisible text spans in the rendered
// ANSI SGR output. It is exported so it can be unit-tested separately.
func findInvisibleTextSpans(rendered string) []invisibleSpan {
	var spans []invisibleSpan

	state := ansiState{}
	lines := strings.Split(rendered, "\n")

	for lineIdx, line := range lines {
		raw := []byte(line)
		var buf strings.Builder
		spanCol := 0
		col := 0

		resetSpan := func() {
			if buf.Len() > 0 && state.invisible() {
				spans = append(spans, invisibleSpan{
					line:    lineIdx + 1,
					col:     spanCol,
					text:    buf.String(),
					fgColor: *state.fg,
				})
			}
			buf.Reset()
		}

		i := 0
		for i < len(raw) {
			b := raw[i]

			// Detect CSI sequence ESC [
			if b == 0x1b && i+1 < len(raw) && raw[i+1] == '[' {
				resetSpan()
				i = parseSGR(raw, i+2, &state)
				continue
			}

			if buf.Len() == 0 {
				spanCol = col
			}
			buf.WriteByte(b)
			col++
			i++
		}

		resetSpan()
	}

	return spans
}

// parseSGR parses a CSI SGR sequence starting after "ESC[" and updates state.
// Returns the index after the terminating byte ('m' or unrecognized).
func parseSGR(raw []byte, start int, state *ansiState) int {
	i := start
	params := []int{}
	cur := 0
	hasDigit := false

	for i < len(raw) {
		c := raw[i]
		if c >= '0' && c <= '9' {
			cur = cur*10 + int(c-'0')
			hasDigit = true
		} else if c == ';' {
			if hasDigit {
				params = append(params, cur)
			}
			cur = 0
			hasDigit = false
		} else if c == 'm' {
			if hasDigit {
				params = append(params, cur)
			}
			i++
			applySGR(params, state)
			return i
		} else {
			// Unsupported CSI command — skip to terminator
			if c >= 0x40 && c <= 0x7e {
				i++
				return i
			}
		}
		i++
	}
	return i
}

// applySGR applies SGR parameters to the color state.
func applySGR(params []int, state *ansiState) {
	if len(params) == 0 {
		// ESC[m is equivalent to ESC[0m (full reset)
		state.reset()
		return
	}

	j := 0
	for j < len(params) {
		p := params[j]
		switch {
		case p == 0:
			state.reset()
		case p == 1:
			// Bold — no color impact, skip
		case p == 2:
			// Faint — no color impact
		case p == 7:
			// Reverse video — would swap fg/bg, but lipgloss doesn't
			// use this in the current codebase; skip.
		case p == 38:
			// Extended foreground: 38 ; 5 ; N  OR  38 ; 2 ; R ; G ; B
			if j+2 < len(params) && params[j+1] == 5 {
				if n := normalizeIndexedToANSI(params[j+2]); n >= 0 {
					state.setFG(n)
				}
				j += 3
				continue
			} else if j+5 < len(params) && params[j+1] == 2 {
				// True color — skip (our theme doesn't use it for text)
				j += 6
				continue
			}
			j++
		case p == 48:
			// Extended background
			if j+2 < len(params) && params[j+1] == 5 {
				if n := normalizeIndexedToANSI(params[j+2]); n >= 0 {
					state.setBG(n)
				}
				j += 3
				continue
			} else if j+5 < len(params) && params[j+1] == 2 {
				j += 6
				continue
			}
			j++
		case p >= 30 && p <= 37:
			state.setFG(p - 30)
		case p >= 40 && p <= 47:
			state.setBG(p - 40)
		case p >= 90 && p <= 97:
			state.setFG(p - 90 + 8)
		case p >= 100 && p <= 107:
			state.setBG(p - 100 + 8)
		}
		j++
	}
}

// normalizeIndexedToANSI maps a 256-color palette index to ANSI 0-15
// when there is a direct mapping. Returns -1 for non-ANSI indices.
func normalizeIndexedToANSI(idx int) int {
	if idx >= 0 && idx <= 15 {
		return idx
	}
	return -1
}

// ---------------------------------------------------------------------------
// Test helpers — render common view models and check for invisible text
// ---------------------------------------------------------------------------

func checkMainViewInvisibleText(t *testing.T, m *MainModel) {
	t.Helper()
	rendered := m.View().Content
	checkInvisibleText(t, "MainModel.View()", rendered)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestMainMenuNoInvisibleText verifies the main menu view has no invisible text.
func TestMainMenuNoInvisibleText(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30
	checkMainViewInvisibleText(t, m)
}

// TestVMsTabNoInvisibleText verifies the VMs tab with VMs has no invisible text.
func TestVMsTabNoInvisibleText(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30
	checkMainViewInvisibleText(t, m)
}

// TestConfigTabNoInvisibleText verifies the Configuration tab has no invisible text.
func TestConfigTabNoInvisibleText(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30
	m.tabModel.SetActiveTab(components.TabConfiguration)
	checkMainViewInvisibleText(t, m)
}

// TestPowerTabNoInvisibleText verifies the Power tab has no invisible text.
func TestPowerTabNoInvisibleText(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30
	m.tabModel.SetActiveTab(components.TabPower)
	checkMainViewInvisibleText(t, m)
}

// TestMountPointWarningNoInvisibleText verifies the mount point warning view.
func TestMountPointWarningNoInvisibleText(t *testing.T) {
	m := NewMountPointWarningModel()
	m.SetSize(80, 30)
	rendered := m.View().Content
	checkInvisibleText(t, "MountPointWarningModel.View()", rendered)
}

// TestVMSelectNoInvisibleText verifies the VM selection view.
func TestVMSelectNoInvisibleText(t *testing.T) {
	m := setupTestModelForScenarios(t)
	model, _ := m.showVMSelection()
	m2 := model.(*MainModel)
	rendered := m2.View().Content
	checkInvisibleText(t, "VMSelectModel.View()", rendered)
}

// TestQuitNoInvisibleText verifies the quit/goodbye view.
func TestQuitNoInvisibleText(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30
	m.quitting = true
	checkMainViewInvisibleText(t, m)
}

// TestEmptyVMsTabNoInvisibleText verifies the VMs tab with no VMs.
func TestEmptyVMsTabNoInvisibleText(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30
	// Stay on VMs tab (default) with no VMs
	checkMainViewInvisibleText(t, m)
}

// TestVMDetailNoInvisibleText verifies the VM detail view (right panel)
// rendered in the VMs tab when a VM is selected.
func TestVMDetailNoInvisibleText(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30
	// Select first VM to show detail panel
	m.menuList.Select(0)
	m.selectedIndex = 0
	checkMainViewInvisibleText(t, m)
}

// TestInvisibleTextDetector verifies the detector itself works correctly
// by feeding known ANSI sequences.
func TestInvisibleTextDetector(t *testing.T) {
	tests := []struct {
		name     string
		input    string // ANSI-rendered string
		want     int    // expected invisible spans
	}{
		{
			name:  "plain text — no ANSI",
			input: "Hello World",
			want:  0,
		},
		{
			name:  "fg different from bg — safe",
			input: "\x1b[36;40mHello\x1b[m", // fg=6 cyan, bg=0 black
			want:  0,
		},
		{
			name:  "fg equals bg — invisible",
			input: "\x1b[30;40minvisible\x1b[m", // fg=0 black, bg=0 black
			want:  1,
		},
		{
			name:  "fg=bg bright — invisible",
			input: "\x1b[91;101mbright\x1b[m", // fg=9 bright red, bg=101 bright red
			want:  1,
		},
		{
			name:  "mixed — some visible, some invisible",
			input: "\x1b[36mvisible\x1b[m \x1b[30;40minvisible\x1b[m",
			want:  1,
		},
		{
			name:  "nested invisible — outer bg + inner fg match",
			input: "\x1b[40m\x1b[30mhidden\x1b[m\x1b[m",
			want:  1,
		},
		{
			name:  "no fg set — not invisible regardless of bg",
			input: "\x1b[40mtext with bg only\x1b[m",
			want:  0,
		},
		{
			name:  "no bg set — not invisible regardless of fg",
			input: "\x1b[30mtext with fg only\x1b[m",
			want:  0,
		},
		{
			name:  "256-color invisible",
			input: "\x1b[48;5;0m\x1b[38;5;0mhidden\x1b[m\x1b[m", // bg=0, fg=0 via 256-color
			want:  1,
		},
		{
			name:  "reset between — no false positive",
			input: "\x1b[30;40mhidden\x1b[mvisible\x1b[30mstill visible",
			want:  1, // only "hidden" is invisible, rest is fine
		},
		{
			name:  "border rendering — no false positive",
			input: "\x1b[90m┌────┐\x1b[m\n\x1b[90m│\x1b[mtext\x1b[90m│\x1b[m\n\x1b[90m└────┘\x1b[m",
			want:  0, // borders are fg=8, text is default — no bg collision
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findInvisibleTextSpans(tt.input)
			if len(got) != tt.want {
				t.Errorf("got %d invisible spans, want %d", len(got), tt.want)
				for _, s := range got {
					t.Logf("  %s", s.String())
				}
			}
		})
	}
}

// TestVMEditViewNoInvisibleText verifies the VM edit form view.
func TestVMEditViewNoInvisibleText(t *testing.T) {
	m := setupTestModelForScenarios(t)
	m.windowWidth = 80
	m.windowHeight = 30

	// Navigate: Config tab -> Edit VM (Enter) -> VMSelect -> Enter to select VM
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.configSelectedIndex = 1 // "Edit VM"
	model, cmd := m.handleMenuSelection()
	m2 := model.(*MainModel)
	if cmd != nil {
		msg := cmd()
		var updated tea.Model
		updated, _ = m2.Update(msg)
		m2 = updated.(*MainModel)
	}

	// Now in VMSelect, press Enter to select a VM
	var m3 tea.Model
	var cmd2 tea.Cmd
	m3, cmd2 = m2.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m4 := m3.(*MainModel)
	if cmd2 != nil {
		msg := cmd2()
		var updated tea.Model
		updated, _ = m4.Update(msg)
		m4 = updated.(*MainModel)
	}

	rendered := m4.View().Content
	checkInvisibleText(t, "VMEdit.View()", rendered)
}
