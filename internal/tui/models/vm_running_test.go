package models

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/glemsom/dkvmmanager/internal/domain"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

func setupRunningModel(t *testing.T, status string) *VMRunningModel {
	t.Helper()
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	return &VMRunningModel{
		vm:          vmObj,
		maxLogLines: 500,
		vp:          viewport.New(viewport.WithWidth(80), viewport.WithHeight(24)),
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
	viewContent := m.View().Content
	if viewContent != "Loading..." {
		t.Errorf("Expected 'Loading...', got '%s'", viewContent)
	}
}

func TestVMRunningModelViewReady(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	viewContent := m.View().Content
	if viewContent == "Loading..." {
		t.Error("View should not return 'Loading...' when ready")
	}
	if viewContent == "" {
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
	updated, cmd := m.Update(VMStatusUpdateMsg{Status: "running"})
	m = updated.(*VMRunningModel)
	if m.status != "running" {
		t.Errorf("Expected status 'running', got '%s'", m.status)
	}
	if cmd == nil {
		t.Error("Expected non-nil command (pollStatus) after status update")
	}
}

func TestVMRunningModelStatusUpdateStopped(t *testing.T) {
	// When status is terminal ("stopped"), VMStatusUpdateMsg should NOT re-arm the poll tick.
	m := setupRunningModel(t, "starting")
	m.status = "stopped"
	updated, cmd := m.Update(VMStatusUpdateMsg{Status: "running"})
	m = updated.(*VMRunningModel)
	if m.status != "stopped" {
		t.Errorf("Expected status to remain 'stopped', got '%s'", m.status)
	}
	if cmd != nil {
		t.Error("Expected nil command (no re-arm) when status is 'stopped' after VMStatusUpdateMsg")
	}
}

func TestVMRunningModelStatusUpdateStopping(t *testing.T) {
	// When status is terminal ("stopping"), VMStatusUpdateMsg should NOT re-arm the poll tick.
	m := setupRunningModel(t, "starting")
	m.status = "stopping"
	updated, cmd := m.Update(VMStatusUpdateMsg{Status: "running"})
	m = updated.(*VMRunningModel)
	if m.status != "stopping" {
		t.Errorf("Expected status to remain 'stopping', got '%s'", m.status)
	}
	if cmd != nil {
		t.Error("Expected nil command (no re-arm) when status is 'stopping' after VMStatusUpdateMsg")
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
	if m.vp.Width() != 100 {
		t.Errorf("Expected viewport width 100, got %d", m.vp.Width())
	}
	if m.vp.Height() != 25 { // 30 - infoHeight(2) - 3 = 25
		t.Errorf("Expected viewport height 25, got %d", m.vp.Height())
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
	if m.vp.Width() != 120 {
		t.Errorf("Expected viewport width 120, got %d", m.vp.Width())
	}
	if m.vp.Height() != 35 { // 40 - infoHeight(2) - 3 = 35
		t.Errorf("Expected viewport height 35, got %d", m.vp.Height())
	}
}

func TestVMRunningModelViewContentHeader(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	viewContent := m.View().Content
	if !strings.Contains(viewContent, "[RUNNING]") {
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
		{"paused", "paused", "[RUNNING]"},
		{"postmigrate", "postmigrate", "[RUNNING]"},
		{"prelaunch", "prelaunch", "[RUNNING]"},
		{"unknown", "unknown", "[STARTING]"},
		{"stopped", "stopped", "[STOPPED]"},
		{"stopping", "stopping", "[STOPPING]"},
		{"starting", "starting", "[STARTING]"},
		{"exited", "exited", "[STOPPED]"},
		{"finish", "finish", "[STOPPING]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := setupRunningModel(t, tt.status)
			m.updateViewport()
			viewContent := m.View().Content
			if !strings.Contains(viewContent, tt.expect) {
				t.Errorf("Expected view to contain '%s', got:\n%s", tt.expect, viewContent)
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
	viewContent := m.View().Content
	if !strings.Contains(viewContent, "qemu: starting") {
		t.Error("View should contain first log line")
	}
	if !strings.Contains(viewContent, "qemu: boot complete") {
		t.Error("View should contain second log line")
	}
}

func TestVMRunningModelViewEmptyLogs(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Waiting for output...") {
		t.Error("View should show 'Waiting for output...' with no logs")
	}
}

func TestVMRunningModelViewFooterNilRunner(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	viewContent := m.View().Content
	if !strings.Contains(viewContent, "q: Exit view") {
		t.Error("View should show 'q: Exit view' when runner is nil")
	}
}

func TestVMRunningModelViewFooterStopped(t *testing.T) {
	m := setupRunningModel(t, "stopped")
	m.updateViewport()
	viewContent := m.View().Content
	if !strings.Contains(viewContent, "q: Exit view") {
		t.Error("View should show 'q: Exit view' when stopped")
	}
}

func TestVMRunningModelKeyQWhenRunning(t *testing.T) {
	m := setupRunningModel(t, "running")
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
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
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
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
	initialOffset := m.vp.YOffset()
	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
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
	viewContent := m.View().Content
	if !strings.Contains(viewContent, "[STOPPED]") {
		t.Error("View should show [STOPPED] after VMStoppedMsg")
	}
}

func TestVMRunningModelMultipleStatusUpdates(t *testing.T) {
	m := setupRunningModel(t, "starting")
	updated, _ := m.Update(VMStatusUpdateMsg{Status: "running"})
	m = updated.(*VMRunningModel)
	updated, _ = m.Update(VMStatusUpdateMsg{Status: "running"})
	m = updated.(*VMRunningModel)
	if m.status != "running" {
		t.Errorf("Expected status 'running', got '%s'", m.status)
	}
}

func TestVMRunningModelStatusUpdateNoThreads(t *testing.T) {
	m := setupRunningModel(t, "starting")
	updated, _ := m.Update(VMStatusUpdateMsg{Status: "running"})
	m = updated.(*VMRunningModel)
	if m.status != "running" {
		t.Errorf("Expected status 'running', got '%s'", m.status)
	}
}

func TestVMRunningModelStartTimeZero(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	viewContent := m.View().Content
	if !strings.Contains(viewContent, "[RUNNING]") {
		t.Error("View should show [RUNNING]")
	}
}

func TestVMRunningModelViewContainsStatusIndicator(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	viewContent := m.View().Content
	if !strings.Contains(viewContent, "Status:") {
		t.Error("View should contain 'Status:' label")
	}
}

func TestVMRunningModelViewContainsQEMUOutput(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.updateViewport()
	viewContent := m.View().Content
	if !strings.Contains(viewContent, "QEMU Output") {
		t.Error("View should contain 'QEMU Output' separator")
	}
}

func TestVMRunningModelRunnerNilSafe(t *testing.T) {
	m := setupRunningModel(t, "running")
	if m.Runner() != nil {
		t.Error("Expected Runner() to return nil")
	}
}

func TestVMRunningModelMetricsUpdate(t *testing.T) {
	m := setupRunningModel(t, "running")

	// Send a metrics update with per-vCPU data
	metricsMsg := VMMetricsUpdateMsg{
		Metrics: vm.Metrics{
			Status: "running",
			VCPUs: []vm.VCPUStat{
				{ThreadID: 100, CPUTimeNs: 5000},  // 50.00%
				{ThreadID: 101, CPUTimeNs: 2500},  // 25.00%
			},
		},
	}
	updated, cmd := m.Update(metricsMsg)
	m = updated.(*VMRunningModel)

	if len(m.metrics.VCPUs) != 2 {
		t.Fatalf("Expected 2 vCPUs in metrics, got %d", len(m.metrics.VCPUs))
	}
	if m.metrics.VCPUs[0].CPUTimeNs != 5000 {
		t.Errorf("Expected CPUTimeNs=5000 for vCPU 0, got %d", m.metrics.VCPUs[0].CPUTimeNs)
	}
	if cmd == nil {
		t.Error("Expected non-nil command (pollMetrics) after metrics update")
	}
}

func TestVMRunningModelMetricsRendering(t *testing.T) {
	m := setupRunningModel(t, "running")

	// Set metrics with per-vCPU data
	m.metrics = vm.Metrics{
		Status: "running",
		VCPUs: []vm.VCPUStat{
			{ThreadID: 100, CPUTimeNs: 5000},  // 50.00%
			{ThreadID: 101, CPUTimeNs: 2500},  // 25.00%
		},
	}

	m.updateViewport()
	viewContent := m.View().Content

	// Should contain per-vCPU percentages
	if !strings.Contains(viewContent, "50.0%") {
		t.Error("View should contain '50.0%' for vCPU 0")
	}
	if !strings.Contains(viewContent, "25.0%") {
		t.Error("View should contain '25.0%' for vCPU 1")
	}
	if !strings.Contains(viewContent, "75.0%") {
		t.Error("View should contain '75.0%' for aggregate total")
	}
}

func TestVMRunningModelMetricsEmpty(t *testing.T) {
	m := setupRunningModel(t, "running")

	// Send empty metrics (no vCPUs)
	metricsMsg := VMMetricsUpdateMsg{
		Metrics: vm.Metrics{Status: "running"},
	}
	updated, cmd := m.Update(metricsMsg)
	m = updated.(*VMRunningModel)

	if len(m.metrics.VCPUs) != 0 {
		t.Errorf("Expected 0 vCPUs, got %d", len(m.metrics.VCPUs))
	}
	if cmd == nil {
		t.Error("Expected non-nil command after metrics update")
	}

	m.updateViewport()
	viewContent := m.View().Content
	// Should NOT contain vCPU% label when no vCPUs
	if strings.Contains(viewContent, "vCPU%:") {
		t.Error("View should not contain 'vCPU%:' when no vCPU data")
	}
}

// S4: a VMMetricsUpdateMsg carrying host CPU% and RSS is stored on the model
// and triggers another pollMetrics() tick.
func TestVMRunningModelHostMetricsUpdate(t *testing.T) {
	m := setupRunningModel(t, "running")

	metricsMsg := VMMetricsUpdateMsg{
		Metrics: vm.Metrics{
			Status:        "running",
			HostRSSBytes:  100 * 1024 * 1024, // 100 MiB
			HostCPUJiffies: 520,               // 5.20%
		},
	}
	updated, cmd := m.Update(metricsMsg)
	m = updated.(*VMRunningModel)

	if m.metrics.HostRSSBytes != 100*1024*1024 {
		t.Errorf("Expected HostRSSBytes=%d, got %d", 100*1024*1024, m.metrics.HostRSSBytes)
	}
	if m.metrics.HostCPUJiffies != 520 {
		t.Errorf("Expected HostCPUJiffies=520, got %d", m.metrics.HostCPUJiffies)
	}
	if cmd == nil {
		t.Error("Expected non-nil command (pollMetrics) after host metrics update")
	}
}

// S4: the rendered view contains the host CPU% and RSS strings, in
// human-readable units, alongside the per-vCPU guest metrics.
func TestVMRunningModelHostMetricsRendering(t *testing.T) {
	m := setupRunningModel(t, "running")

	// 5.20% host CPU (HostCPUJiffies=520 → 520/100=5.20)
	// 100 MiB RSS = 100 * 1024 * 1024 bytes
	m.metrics = vm.Metrics{
		Status:         "running",
		HostRSSBytes:   100 * 1024 * 1024,
		HostCPUJiffies: 520,
	}

	m.updateViewport()
	viewContent := m.View().Content

	if !strings.Contains(viewContent, "Host:") {
		t.Error("View should contain 'Host:' label for host metrics")
	}
	if !strings.Contains(viewContent, "5.2%") {
		t.Error("View should contain '5.2%' for host CPU")
	}
	if !strings.Contains(viewContent, "100") {
		t.Error("View should contain '100' for host RSS value")
	}
	if !strings.Contains(viewContent, "MiB") {
		t.Error("View should contain 'MiB' unit for host RSS")
	}
}

// S4: when host fields are zero (cold snapshot, PID=0, or /proc unreadable),
// the view should NOT render the Host: line.
func TestVMRunningModelHostMetricsEmpty(t *testing.T) {
	m := setupRunningModel(t, "running")

	// No host data
	m.metrics = vm.Metrics{Status: "running"}

	m.updateViewport()
	viewContent := m.View().Content

	if strings.Contains(viewContent, "Host:") {
		t.Error("View should not contain 'Host:' when no host data")
	}
}
func TestVMRunningModelSetSizeUpdatesViewport(t *testing.T) {
	m := setupRunningModel(t, "running")
	m.SetSize(100, 30)
	m.updateViewport()
	viewContent := m.View().Content
	if !strings.Contains(viewContent, "[RUNNING]") {
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

// --- Async VM Start Tests ---

func TestVMStartedMsgHandlerSetsRunner(t *testing.T) {
	m := setupRunningModel(t, "starting")

	// Create a real runner (with nil config, not used in this test)
	runner := vm.NewVMRunner(&domain.VM{Name: "test-vm", ID: "1"}, nil, vm.RunConfig{}, false)

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
	// When runner is nil, waitForLog should return a command that returns nil (no-op),
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
