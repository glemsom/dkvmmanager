package models

import (
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

// waitForString polls tm.Output() until it contains the expected string or times out.
// NOTE: teatest.WaitFor drains the output buffer, so use waitForStringAfter for
// subsequent checks in the same test.
func waitForString(t *testing.T, tm *teatest.TestModel, s string) {
	t.Helper()
	teatest.WaitFor(
		t,
		tm.Output(),
		func(b []byte) bool {
			return strings.Contains(string(b), s)
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*10),
	)
}

// waitForStringAfter drains the output buffer, then waits for new output containing s.
// Use this when you need multiple string checks with messages in between.
func waitForStringAfter(t *testing.T, tm *teatest.TestModel, s string) {
	t.Helper()
	// Drain previous output
	io.ReadAll(tm.Output())
	// Wait for new output containing s
	teatest.WaitFor(
		t,
		tm.Output(),
		func(b []byte) bool {
			return strings.Contains(string(b), s)
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*10),
	)
}

// TestTeatestQuitFlow verifies the full quit flow via Ctrl+C through a running tea.Program.
func TestTeatestQuitFlow(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	waitForString(t, tm, "DKVM Manager")

	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})

	waitForStringAfter(t, tm, "Goodbye")
}

// TestTeatestQuitViaQ verifies quit via 'q' key through a running tea.Program.
func TestTeatestQuitViaQ(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	waitForString(t, tm, "DKVM Manager")

	tm.Type("q")

	waitForStringAfter(t, tm, "Goodbye")
}

// TestTeatestInitialRender verifies the initial render contains expected elements.
func TestTeatestInitialRender(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	waitForString(t, tm, "Start VM")
}

// TestTeatestInitialRenderWithVMs verifies initial render with VMs configured.
func TestTeatestInitialRenderWithVMs(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	waitForString(t, tm, "test-vm-1")
}

// TestTeatestTabSwitching verifies tab switching through a running tea.Program.
func TestTeatestTabSwitching(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	waitForString(t, tm, "Start VM")

	// Press Tab to switch to Configuration
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

	waitForStringAfter(t, tm, "Add new VM")
}

// TestTeatestConfigToVMCreate verifies navigating to VM create through the running program.
func TestTeatestConfigToVMCreate(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	waitForString(t, tm, "Start VM")

	// Switch to Configuration tab
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	// Select Add new VM
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	waitForStringAfter(t, tm, "Add VM")
}

// TestTeatestESCFromSubView verifies ESC returns to main menu through the running program.
func TestTeatestESCFromSubView(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	waitForString(t, tm, "Start VM")

	// Navigate to VM create via Config tab -> Enter (Add new VM)
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	waitForStringAfter(t, tm, "Add new VM")
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	waitForStringAfter(t, tm, "Add VM")

	// Press ESC to return
	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})

	// Should see config tab content (ESC returns to config tab)
	waitForStringAfter(t, tm, "Add new VM")
}

// TestTeatestStatusbar verifies the status bar renders in the running program.
func TestTeatestStatusbar(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	waitForString(t, tm, "Ready")
}

// TestTeatestWindowSize verifies window size handling in a running program.
func TestTeatestWindowSize(t *testing.T) {
	m := setupTestModel(t)

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(120, 40),
	)

	waitForString(t, tm, "Start VM")
}

// TestTeatestGoldenMainView uses teatest's built-in golden file comparison for the main view.
func TestTeatestGoldenMainView(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	teatest.WaitFor(
		t,
		tm.Output(),
		func(b []byte) bool {
			return strings.Contains(string(b), "DKVM Manager")
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*5),
	)

	// Quit the program so FinalOutput can return
	tm.Quit()

	out, err := io.ReadAll(tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second*3)))
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	teatest.RequireEqualOutput(t, out)
}

// TestTeatestGoldenMainViewWithVMs uses teatest's golden file for main view with VMs.
func TestTeatestGoldenMainViewWithVMs(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	teatest.WaitFor(
		t,
		tm.Output(),
		func(b []byte) bool {
			return strings.Contains(string(b), "test-vm-1")
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*5),
	)

	// Quit the program so FinalOutput can return
	tm.Quit()

	out, err := io.ReadAll(tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second*3)))
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	teatest.RequireEqualOutput(t, out)
}

// TestTeatestConfigTabSwitch verifies switching to config tab in running program.
func TestTeatestConfigTabSwitch(t *testing.T) {
	m := setupTestModelWithVMs(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	waitForString(t, tm, "test-vm-1")

	// Press Tab to switch to Configuration
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

	waitForStringAfter(t, tm, "Edit VM")

	// Press Shift+Tab to return to VMs
	tm.Send(tea.KeyMsg{Type: tea.KeyShiftTab})

	waitForStringAfter(t, tm, "test-vm-1")
}

// TestTeatestPowerTabSwitch verifies switching to power tab in running program.
func TestTeatestPowerTabSwitch(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	tm := teatest.NewTestModel(t, m,
		teatest.WithInitialTermSize(80, 30),
	)

	waitForString(t, tm, "Start VM")

	// Press Tab twice to get to Power tab
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})
	tm.Send(tea.KeyMsg{Type: tea.KeyTab})

	waitForStringAfter(t, tm, "Reboot system")
}
