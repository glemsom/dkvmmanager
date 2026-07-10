// Package models — deep analysis of VMRunningModel for bugs.
package models

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	"github.com/glemsom/dkvmmanager/internal/domain"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

func TestBug_InfoHeightMatchesRendered(t *testing.T) {
	// Regression: calculateInfoHeight must match actual rendered lines
	// in renderInfoPanel. Test with various metric combinations.
	tests := []struct {
		name    string
		metrics vm.Metrics
	}{
		{"empty", vm.Metrics{Status: "running"}},
		{"with vcpus", vm.Metrics{Status: "running", VCPUs: []vm.VCPUStat{{ThreadID: 1, CPUTimeNs: 5000}}}},
		{"with host metrics", vm.Metrics{Status: "running", HostRSSBytes: 1024, HostCPUJiffies: 100}},
		{"with balloon", vm.Metrics{Status: "running", BalloonBytes: 1024 * 1024 * 1024}},
		{"with block devices", vm.Metrics{Status: "running", BlockDevices: []vm.BlockStat{{Device: "drive0"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &VMRunningModel{
				vm:          &domain.VM{Name: "test-vm", ID: "1"},
				maxLogLines: 500,
				vp:          viewport.New(viewport.WithWidth(80), viewport.WithHeight(24)),
				ready:       true,
				width:       80,
				height:      24,
				status:      "running",
				metrics:     tt.metrics,
			}
			calcH := m.calculateInfoHeight()
			renderedH := strings.Count(m.renderInfoPanel(), "\n")
			if calcH != renderedH {
				t.Errorf("infoHeight=%d but rendered lines=%d", calcH, renderedH)
			}
		})
	}
}

func TestBug_DuplicateSeparator(t *testing.T) {
	// BUG: "─── QEMU Output ───" renders TWICE:
	//  1. As last line of renderInfoPanel()
	//  2. As first line of renderLogContent() (inside viewport)
	// The viewport version was added so the separator is visible when
	// scrolling up, but the info panel version was never removed.
	m := &VMRunningModel{
		vm:          &domain.VM{Name: "test-vm", ID: "1"},
		maxLogLines: 500,
		vp:          viewport.New(viewport.WithWidth(80), viewport.WithHeight(24)),
		ready:       true,
		width:       80,
		height:      24,
		status:      "running",
	}
	m.updateViewport()
	content := m.View().Content

	count := strings.Count(content, "QEMU Output")
	if count == 1 {
		t.Log("OK: exactly one QEMU Output separator")
	} else {
		t.Errorf("BUG: found %d 'QEMU Output' occurrences (expected 1) — separator renders in both infoPanel AND viewport", count)
	}
}

func TestBug_ThreadsFieldUnused(t *testing.T) {
	// BUG: VMRunningModel.threads []int is populated on every status update
	// but never rendered in the view. Field written but never read.
	m := &VMRunningModel{
		vm:          &domain.VM{Name: "test-vm", ID: "1"},
		maxLogLines: 500,
		vp:          viewport.New(viewport.WithWidth(80), viewport.WithHeight(24)),
		ready:       true,
		width:       80,
		height:      24,
		status:      "running",
	}

	// Send a status update with threads
	m.Update(VMStatusUpdateMsg{Status: "running", Threads: []int{100, 101, 102}})

	content := m.View().Content
	if strings.Contains(content, "Thread") || strings.Contains(content, "thread") {
		t.Log("OK: thread info present in view")
	} else {
		t.Log("thread info not shown in view — field 'threads' is populated but never rendered")
	}
}

func TestBug_ForcedAutoScrollOnEveryLogLine(t *testing.T) {
	// BUG: Every new log line calls m.vp.GotoBottom(), forcing the user
	// to the bottom even if they scrolled up to read older output.
	// Auto-scroll should only trigger when the viewport was already at bottom.
	m := &VMRunningModel{
		vm:          &domain.VM{Name: "test-vm", ID: "1"},
		maxLogLines: 500,
		vp:          viewport.New(viewport.WithWidth(80), viewport.WithHeight(24)),
		ready:       true,
		width:       80,
		height:      24,
		status:      "running",
	}

	// Add initial log lines
	for i := 0; i < 10; i++ {
		m.Update(VMLogMsg{Line: fmt.Sprintf("line %d", i)})
	}

	// User scrolls up
	m.vp.SetYOffset(0)

	// New log line arrives
	m.Update(VMLogMsg{Line: "new line after scroll"})

	yOff := m.vp.YOffset()
	if yOff == 0 {
		t.Log("OK: viewport stayed at YOffset=0 after manual scroll-up and new log line")
	} else {
		t.Errorf("BUG: viewport forced to YOffset=%d instead of staying at 0 — auto-scroll on every log line ignores user scroll position", yOff)
	}
}

func TestBug_PollStatusOverwritesStopping(t *testing.T) {
	// BUG: When user presses 'q' to stop VM, handleKeyPress sets
	// status="stopping" but does NOT return pollStatus(). The old
	// pollStatus tick fires one more time, querying QMP, and
	// potentially overwriting "stopping" with "unknown", "running",
	// or whatever QMP returns as the VM shuts down.
	m := &VMRunningModel{
		vm:          &domain.VM{Name: "test-vm", ID: "1"},
		maxLogLines: 500,
		vp:          viewport.New(viewport.WithWidth(80), viewport.WithHeight(24)),
		ready:       true,
		width:       80,
		height:      24,
		status:      "running",
		// No runner — simulates what happens when QMP returns after Stop
	}

	m.status = "stopping"

	// Simulate a delayed status poll firing after stop
	m.Update(VMStatusUpdateMsg{Status: "unknown", Threads: nil})

	if m.status != "stopping" {
		t.Errorf("BUG: pollStatus overwrote 'stopping' with '%s' — status chain not broken on stop", m.status)
	} else {
		t.Log("OK: stopping preserved")
	}
}

func TestBug_StoppedStatusAfterStopThenQMP(t *testing.T) {
	// BUG: When user presses 'q', QMP client.Quit() is called.
	// As QEMU shuts down, QMP might return "shutdown" status which
	// falls into the default render case → shows [STARTING].
	m := &VMRunningModel{
		vm:          &domain.VM{Name: "test-vm", ID: "1"},
		maxLogLines: 500,
		vp:          viewport.New(viewport.WithWidth(80), viewport.WithHeight(24)),
		ready:       true,
		width:       80,
		height:      24,
		status:      "shutdown",
	}
	m.updateViewport()
	content := m.View().Content

	// "shutdown" should map to [STOPPED] or similar, not [STARTING]
	if strings.Contains(content, "[STARTING]") {
		t.Errorf("BUG: QMP status 'shutdown' renders as [STARTING] — missing case in status badge mapping")
	} else if strings.Contains(content, "[STOPPED]") {
		t.Log("OK: 'shutdown' maps to [STOPPED]")
	} else {
		t.Log("'shutdown' shows something else (check which)")
	}
}

func TestBug_StartedMsgNoRunnerDoesntPanic(t *testing.T) {
	// VMStartedMsg handler sets m.pollingSince = time.Now()
	// But if msg.Runner is nil (shouldn't happen, but defensive), would panic.
	// Ensure no panic, though code doesn't nil-check runner.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("BUG: VMStartedMsg handler panicked with nil runner: %v", r)
		}
	}()

	m := &VMRunningModel{
		vm:          &domain.VM{Name: "test-vm", ID: "1"},
		maxLogLines: 500,
		vp:          viewport.New(viewport.WithWidth(80), viewport.WithHeight(24)),
		ready:       true,
		width:       80,
		height:      24,
		status:      "starting",
	}

	// Passing nil Runner on purpose to test
	m.Update(VMStartedMsg{Runner: nil, VMName: "test-vm", VMID: "1"})
	// Should not panic
}
