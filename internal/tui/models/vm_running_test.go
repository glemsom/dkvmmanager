package models

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	models "github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

func setupRunningModel(t *testing.T, status string) *VMRunningModel {
	t.Helper()
	vmObj := &models.VM{Name: "test-vm", ID: "1"}
	return &VMRunningModel{
		vm:          vmObj,
		maxLogLines: 500,
		vp:          viewport.New(80, 24),
		ready:       true,
		width:       80,
		height:      24,
		status:      status,
	}
}

func TestVMRunningModelInit(t *testing.T) {
	m := setupRunningModel(t, "running")
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() should return non-nil batch command")
	}
}

func TestVMRunningModelViewNotReady(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.ready = false
	view := m.View()
	if view != "Loading..." {
		t.Errorf("Expected 'Loading...', got '%s'", view)
	}
}

func TestVMRunningModelViewReady(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	view := m.View()
	if view == "Loading..." {
		t.Error("View should not return 'Loading...' when ready")
	}
	if view == "" {
		t.Error("View should not be empty when ready")
	}
}

func TestVMRunningModelLogAccumulation(t *testing.T) {
	m := setupRunningModel(t, "running")
	for i := 0; i < 3; i++ {
		updated, _ := m.Update(VMLogMsg{Line: fmt.Sprintf("line %d", i)})
		m = updated.(*VMRunningModel)
	}
	if len(m.logLines) != 3 {
		t.Errorf("Expected 3 log lines, got %d", len(m.logLines))
	}
	if m.logLines[0] != "line 0" {
		t.Errorf("Expected 'line 0', got '%s'", m.logLines[0])
	}
	if m.logLines[2] != "line 2" {
		t.Errorf("Expected 'line 2', got '%s'", m.logLines[2])
	}
}

func TestVMRunningModelMaxLogLines(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.maxLogLines = 3
	for i := 0; i < 5; i++ {
		updated, _ := m.Update(VMLogMsg{Line: fmt.Sprintf("line %d", i)})
		m = updated.(*VMRunningModel)
	}
	if len(m.logLines) != 3 {
		t.Errorf("Expected 3 log lines (max), got %d", len(m.logLines))
	}
	if m.logLines[0] != "line 2" {
		t.Errorf("Expected oldest line 'line 2', got '%s'", m.logLines[0])
	}
	if m.logLines[2] != "line 4" {
		t.Errorf("Expected newest line 'line 4', got '%s'", m.logLines[2])
	}
}

func TestVMRunningModelEmptyLogMsg(t *testing.T) {
	m := setupRunningModel(t, "running")
	updated, cmd := m.Update(VMLogMsg{Line: ""})
	m = updated.(*VMRunningModel)
	if len(m.logLines) != 0 {
		t.Errorf("Expected 0 log lines, got %d", len(m.logLines))
	}
	if cmd != nil {
		t.Error("Expected nil command for empty log line")
	}
}

func TestVMRunningModelStoppedMsg(t *testing.T) {
	m := setupRunningModel(t, "running")
	updated, cmd := m.Update(VMStoppedMsg{VMName: "test-vm", VMID: "1", Reason: "exited"})
	m = updated.(*VMRunningModel)
	if m.status != "stopped" {
		t.Errorf("Expected status 'stopped', got '%s'", m.status)
	}
	if cmd != nil {
		t.Error("Expected nil command after VMStoppedMsg")
	}
}

func TestVMRunningModelStatusUpdate(t *testing.T) {
	m := setupRunningModel(t, "starting")
	threads := []int{100, 101, 102}
	updated, cmd := m.Update(VMStatusUpdateMsg{Status: "running", Threads: threads})
	m = updated.(*VMRunningModel)
	if m.status != "running" {
		t.Errorf("Expected status 'running', got '%s'", m.status)
	}
	if len(m.threads) != 3 {
		t.Fatalf("Expected 3 threads, got %d", len(m.threads))
	}
	for i, tid := range threads {
		if m.threads[i] != tid {
			t.Errorf("Expected thread[%d]=%d, got %d", i, tid, m.threads[i])
		}
	}
	if cmd == nil {
		t.Error("Expected non-nil command (pollStatus) after status update")
	}
}

func TestVMRunningModelWindowSize(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.ready = false
	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(*VMRunningModel)
	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("Expected height 40, got %d", m.height)
	}
	if !m.ready {
		t.Error("Expected ready=true after first WindowSizeMsg")
	}
	if cmd != nil {
		t.Error("Expected nil command after WindowSizeMsg")
	}
}

func TestVMRunningModelWindowSizeUpdate(t *testing.T) {
	m := setupRunningModel(t, "running")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(*VMRunningModel)
	updated, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = updated.(*VMRunningModel)
	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}
	if m.height != 30 {
		t.Errorf("Expected height 30, got %d", m.height)
	}
	if m.vp.Width != 100 {
		t.Errorf("Expected viewport width 100, got %d", m.vp.Width)
	}
	if m.vp.Height != 23 { // 30 - infoHeight(4) - 3 = 23
		t.Errorf("Expected viewport height 23, got %d", m.vp.Height)
	}
}

func TestVMRunningModelSetSize(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.ready = false
	m.SetSize(100, 30)
	if m.width != 100 {
		t.Errorf("Expected width 100, got %d", m.width)
	}
	if m.height != 30 {
		t.Errorf("Expected height 30, got %d", m.height)
	}
	if !m.ready {
		t.Error("Expected ready=true after SetSize")
	}
}

func TestVMRunningModelSetSizeTwice(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.SetSize(100, 30)
	m.SetSize(120, 40)
	if m.width != 120 {
		t.Errorf("Expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("Expected height 40, got %d", m.height)
	}
	if m.vp.Width != 120 {
		t.Errorf("Expected viewport width 120, got %d", m.vp.Width)
	}
	if m.vp.Height != 33 { // 40 - infoHeight(4) - 3 = 33
		t.Errorf("Expected viewport height 33, got %d", m.vp.Height)
	}
}

func TestVMRunningModelViewContentHeader(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	view := m.View()
	if !strings.Contains(view, "[RUNNING]") {
		t.Error("View should contain status '[RUNNING]'")
	}
}

func TestVMRunningModelViewContentStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		expect string
	}{
		{"running", "running", "[RUNNING]"},
		{"stopped", "stopped", "[STOPPED]"},
		{"stopping", "stopping", "[STOPPING]"},
		{"starting", "starting", "[STARTING]"},
		{"exited", "exited", "[STOPPED]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setupRunningModel(t, tt.status)
			m.updateViewport()
			view := m.View()
			if !strings.Contains(view, tt.expect) {
				t.Errorf("Expected view to contain '%s', got:\n%s", tt.expect, view)
			}
		})
	}
}

func TestVMRunningModelViewContentLogs(t *testing.T) {
	m := setupRunningModel(t, "running")
	updated, _ := m.Update(VMLogMsg{Line: "qemu: starting"})
	m = updated.(*VMRunningModel)
	updated, _ = m.Update(VMLogMsg{Line: "qemu: boot complete"})
	m = updated.(*VMRunningModel)
	view := m.View()
	if !strings.Contains(view, "qemu: starting") {
		t.Error("View should contain first log line")
	}
	if !strings.Contains(view, "qemu: boot complete") {
		t.Error("View should contain second log line")
	}
}

func TestVMRunningModelViewEmptyLogs(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	view := m.View()
	if !strings.Contains(view, "Waiting for output...") {
		t.Error("View should show 'Waiting for output...' with no logs")
	}
}

func TestVMRunningModelViewFooterNilRunner(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	view := m.View()
	if !strings.Contains(view, "q: Exit view") {
		t.Error("View should show 'q: Exit view' when runner is nil")
	}
}

func TestVMRunningModelViewFooterStopped(t *testing.T) {
	m := setupRunningModel(t, "stopped")
	m.updateViewport()
	view := m.View()
	if !strings.Contains(view, "q: Exit view") {
		t.Error("View should show 'q: Exit view' when stopped")
	}
}

func TestVMRunningModelKeyQWhenRunning(t *testing.T) {
	m := setupRunningModel(t, "running")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = updated.(*VMRunningModel)
	if m.status != "running" {
		t.Errorf("Expected status 'running' (no runner to stop), got '%s'", m.status)
	}
	if cmd != nil {
		t.Error("Expected nil command when runner is nil")
	}
}

func TestVMRunningModelKeyCtrlCWhenRunning(t *testing.T) {
	m := setupRunningModel(t, "running")
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(*VMRunningModel)
	if m.status != "running" {
		t.Errorf("Expected status 'running' (no runner to force stop), got '%s'", m.status)
	}
	if cmd != nil {
		t.Error("Expected nil command when runner is nil")
	}
}

func TestVMRunningModelViewportScroll(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	initialOffset := m.vp.YOffset
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(*VMRunningModel)
	_ = initialOffset
}

func TestVMRunningModelLogMsgReturnsWaitCmd(t *testing.T) {
	m := setupRunningModel(t, "running")
	_, cmd := m.Update(VMLogMsg{Line: "some log"})
	if cmd == nil {
		t.Error("Expected non-nil command (waitForLog) after non-empty log line")
	}
}

func TestVMRunningModelStoppedMsgUpdatesViewport(t *testing.T) {
	m := setupRunningModel(t, "running")
	updated, _ := m.Update(VMStoppedMsg{VMName: "test-vm", VMID: "1", Reason: "exited"})
	m = updated.(*VMRunningModel)
	view := m.View()
	if !strings.Contains(view, "[STOPPED]") {
		t.Error("View should show [STOPPED] after VMStoppedMsg")
	}
}

func TestVMRunningModelMultipleStatusUpdates(t *testing.T) {
	m := setupRunningModel(t, "starting")
	updated, _ := m.Update(VMStatusUpdateMsg{Status: "running", Threads: []int{100}})
	m = updated.(*VMRunningModel)
	updated, _ = m.Update(VMStatusUpdateMsg{Status: "running", Threads: []int{100, 101}})
	m = updated.(*VMRunningModel)
	if m.status != "running" {
		t.Errorf("Expected status 'running', got '%s'", m.status)
	}
	if len(m.threads) != 2 {
		t.Errorf("Expected 2 threads, got %d", len(m.threads))
	}
}

func TestVMRunningModelStatusUpdateNoThreads(t *testing.T) {
	m := setupRunningModel(t, "starting")
	updated, _ := m.Update(VMStatusUpdateMsg{Status: "running", Threads: nil})
	m = updated.(*VMRunningModel)
	if m.status != "running" {
		t.Errorf("Expected status 'running', got '%s'", m.status)
	}
	if len(m.threads) != 0 {
		t.Errorf("Expected 0 threads, got %d", len(m.threads))
	}
}

func TestVMRunningModelStartTimeZero(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.startTime = time.Time{}
	m.updateViewport()
	view := m.View()
	if !strings.Contains(view, "[RUNNING]") {
		t.Error("View should show [RUNNING]")
	}
}

func TestVMRunningModelViewContainsStatusIndicator(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	view := m.View()
	if !strings.Contains(view, "Status:") {
		t.Error("View should contain 'Status:' label")
	}
}

func TestVMRunningModelViewContainsQEMUOutput(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	view := m.View()
	if !strings.Contains(view, "QEMU Output") {
		t.Error("View should contain 'QEMU Output' separator")
	}
}

func TestVMRunningModelRunnerNilSafe(t *testing.T) {
	m := setupRunningModel(t, "running")
	if m.Runner() != nil {
		t.Error("Expected Runner() to return nil")
	}
}

func TestVMRunningModelSetSizeUpdatesViewport(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.SetSize(100, 30)
	m.updateViewport()
	view := m.View()
	if !strings.Contains(view, "[RUNNING]") {
		t.Error("View should contain status after SetSize and updateViewport")
	}
}

func TestVMRunningModelLogOrderPreserved(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.maxLogLines = 10
	lines := []string{"first", "second", "third", "fourth", "fifth"}
	for _, line := range lines {
		updated, _ := m.Update(VMLogMsg{Line: line})
		m = updated.(*VMRunningModel)
	}
	for i, expected := range lines {
		if m.logLines[i] != expected {
			t.Errorf("Expected logLines[%d]='%s', got '%s'", i, expected, m.logLines[i])
		}
	}
}

func TestVMRunningModelMaxLogLinesExact(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.maxLogLines = 3
	for i := 0; i < 3; i++ {
		updated, _ := m.Update(VMLogMsg{Line: fmt.Sprintf("line %d", i)})
		m = updated.(*VMRunningModel)
	}
	if len(m.logLines) != 3 {
		t.Errorf("Expected exactly 3 log lines, got %d", len(m.logLines))
	}
	if m.logLines[0] != "line 0" {
		t.Errorf("Expected 'line 0', got '%s'", m.logLines[0])
	}
}

func TestVMRunningModelViewShowsThreads(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.threads = []int{100, 101}
	m.updateViewport()
	view := m.View()
	if !strings.Contains(view, "vCPU Threads:") {
		t.Error("View should show 'vCPU Threads:' when threads present")
	}
	if !strings.Contains(view, "100") {
		t.Error("View should show thread ID 100")
	}
	if !strings.Contains(view, "101") {
		t.Error("View should show thread ID 101")
	}
}

func TestVMRunningModelViewNoThreadsWhenEmpty(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.threads = nil
	m.updateViewport()
	view := m.View()
	if strings.Contains(view, "vCPU Threads:") {
		t.Error("View should not show 'vCPU Threads:' when threads empty")
	}
}

// --- Async VM Start Tests ---

func TestVMStartedMsgHandlerSetsRunner(t *testing.T) {
	m := setupRunningModel(t, "starting")

	// Create a real runner (with nil config, not used in this test)
	runner := vm.NewVMRunner(&models.VM{Name: "test-vm", ID: "1"}, nil)

	// Simulate receiving VMStartedMsg with a runner
	msg := VMStartedMsg{
		Runner: runner,
		VMName: "test-vm",
		VMID:   "1",
	}
	updated, cmd := m.Update(msg)
	m = updated.(*VMRunningModel)

	if m.runner == nil {
		t.Error("Expected runner to be set after VMStartedMsg")
	}
	if m.status != "starting" {
		t.Errorf("Expected status 'starting', got '%s'", m.status)
	}
	if cmd == nil {
		t.Error("Expected non-nil command after VMStartedMsg")
	}
}

func TestVMStartErrorMsgHandler(t *testing.T) {
	// VMStartErrorMsg is handled by the parent MainModel, not VMRunningModel.
	// This test verifies the message type exists and carries the expected data.
	expectedErr := fmt.Errorf("QEMU not found")
	errMsg := VMStartErrorMsg{
		VMName: "test-vm",
		Err:    expectedErr,
	}

	if errMsg.VMName != "test-vm" {
		t.Errorf("Expected VMName 'test-vm', got '%s'", errMsg.VMName)
	}
	if errMsg.Err.Error() != expectedErr.Error() {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, errMsg.Err)
	}
}

func TestStartVMCommandSuccess(t *testing.T) {
	// startVMCommand takes a *vm.VMRunner and calls .Start() which launches
	// real processes (QEMU, swtpm, etc.), making it unsuitable for unit tests.
	// The logic is verified indirectly via integration tests and the
	// VMStartedMsg handler test above.
	// This test confirms the function signature is correct and compiles.
	var _ func(runner *vm.VMRunner, vmName, vmID string) tea.Cmd = startVMCommand
}

func TestNilRunnerWaitForLogReturnsNil(t *testing.T) {
	// When runner is nil, waitForLog should return nil (no-op),
	// not an empty VMLogMsg.
	m := setupRunningModel(t, "starting")
	// Ensure runner is nil
	if m.runner != nil {
		t.Fatal("Expected nil runner in setupRunningModel")
	}

	// Execute the waitForLog command
	cmd := m.waitForLog()
	if cmd == nil {
		t.Fatal("waitForLog should return a non-nil tea.Cmd")
	}

	msg := cmd()
	if msg != nil {
		t.Errorf("Expected nil message when runner is nil, got %T: %v", msg, msg)
	}
}

func TestNilRunnerWaitForVMExitReturnsNil(t *testing.T) {
	// When runner is nil, waitForVMExit should return nil (no-op),
	// not a VMStoppedMsg.
	m := setupRunningModel(t, "starting")
	// Ensure runner is nil
	if m.runner != nil {
		t.Fatal("Expected nil runner in setupRunningModel")
	}

	// Execute the waitForVMExit command
	cmd := m.waitForVMExit()
	if cmd == nil {
		t.Fatal("waitForVMExit should return a non-nil tea.Cmd")
	}

	msg := cmd()
	if msg != nil {
		t.Errorf("Expected nil message when runner is nil, got %T: %v", msg, msg)
	}
}
