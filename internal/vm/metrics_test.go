package vm

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
)

// mockQMPClient implements QMPClientInterface for testing.
type mockQMPClient struct {
	status  string
	cpus    []VCPUInfo
	qError  error
}

func (m *mockQMPClient) QueryStatus() (string, error) {
	if m.qError != nil {
		return "", m.qError
	}
	return m.status, nil
}

func (m *mockQMPClient) QueryCPUs() ([]VCPUInfo, error) {
	if m.qError != nil {
		return nil, m.qError
	}
	return m.cpus, nil
}

func (m *mockQMPClient) QueryCPUsFast() ([]QMPVCPUInfo, error) {
	if m.qError != nil {
		return nil, m.qError
	}
	var out []QMPVCPUInfo
	for _, c := range m.cpus {
		out = append(out, QMPVCPUInfo{
			CPU:      c.CPU,
			ThreadID: c.ThreadID,
		})
	}
	return out, nil
}

func (m *mockQMPClient) Close() error          { return nil }
func (m *mockQMPClient) Quit() error            { return nil }
func (m *mockQMPClient) Events() <-chan QMPEvent { return nil }

func TestSnapshotColdSnapshot(t *testing.T) {
	vmObj := &models.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{})

	// Set up mock QMP client
	mock := &mockQMPClient{
		status: "running",
		cpus: []VCPUInfo{
			{CPU: 0, ThreadID: 100},
			{CPU: 1, ThreadID: 101},
		},
	}
	runner.qmpClient = mock

	// Set up a PID so /proc reading is attempted (but our mock proc returns 0)
	// We need a valid PID for readThreadCPUTime to be called
	runner.mu.Lock()
	runner.cmdProcess = nil // PID=0 means no /proc reading
	runner.mu.Unlock()

	// Override proc reader to return known values
	runner.readThreadCPUTime = func(pid, tid int) (int64, error) {
		// Return some base values
		return int64(tid) * 10000000, nil // 10ms * tid
	}

	// First snapshot (cold)
	m, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() unexpected error: %v", err)
	}

	if m.Status != "running" {
		t.Errorf("expected status 'running', got '%s'", m.Status)
	}
	if len(m.VCPUs) != 2 {
		t.Fatalf("expected 2 vCPUs, got %d", len(m.VCPUs))
	}
	// Cold snapshot: CPU% should be 0
	for i, s := range m.VCPUs {
		if s.CPUTimeNs != 0 {
			t.Errorf("cold snapshot: vCPU[%d].CPUTimeNs expected 0, got %d", i, s.CPUTimeNs)
		}
	}
}

func TestSnapshotWarmSnapshot(t *testing.T) {
	vmObj := &models.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{})

	mock := &mockQMPClient{
		status: "running",
		cpus: []VCPUInfo{
			{CPU: 0, ThreadID: 200},
		},
	}
	runner.qmpClient = mock

	// Set a fake PID so proc reading is attempted
	runner.mu.Lock()
	runner.cmdProcess, _ = os.FindProcess(1) // init process PID always exists
	runner.mu.Unlock()

	// Mock proc reader: returns increasing CPU times
	var callCount int
	var mu sync.Mutex
	runner.readThreadCPUTime = func(pid, tid int) (int64, error) {
		mu.Lock()
		callCount++
		c := callCount
		mu.Unlock()
		if c <= 2 {
			return 100_000_000, nil // 100ms in ns (first snapshot)
		}
		return 200_000_000, nil // 200ms in ns (second snapshot)
	}

	// Cold snapshot (establishes baseline)
	_, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("first Snapshot() error: %v", err)
	}

	// Wait a bit so delta time is measurable
	time.Sleep(10 * time.Millisecond)

	// Warm snapshot (should compute delta)
	m, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("second Snapshot() error: %v", err)
	}

	if len(m.VCPUs) != 1 {
		t.Fatalf("expected 1 vCPU, got %d", len(m.VCPUs))
	}
	// CPU% should be non-zero since we advanced the CPU time
	if m.VCPUs[0].CPUTimeNs == 0 {
		t.Error("warm snapshot: expected non-zero CPU% after CPU time advance")
	}
}

func TestSnapshotNoQMPClient(t *testing.T) {
	vmObj := &models.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{})

	// No QMP client set (nil)

	m, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() unexpected error: %v", err)
	}
	if m.Status != "starting" {
		t.Errorf("expected status 'starting' without QMP client, got '%s'", m.Status)
	}
	if len(m.VCPUs) != 0 {
		t.Errorf("expected 0 vCPUs without QMP client, got %d", len(m.VCPUs))
	}
}

func TestSnapshotQMPError(t *testing.T) {
	vmObj := &models.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{})

	mock := &mockQMPClient{
		qError: fmt.Errorf("mock failure"),
	}
	runner.qmpClient = mock

	_, err := runner.Snapshot()
	if err == nil {
		t.Error("expected error from Snapshot() with failing QMP")
	}
}

func TestPID(t *testing.T) {
	vmObj := &models.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{})

	if pid := runner.PID(); pid != 0 {
		t.Errorf("expected PID=0 for unstarted runner, got %d", pid)
	}
}

// S4 acceptance criteria: given a snapshot with PID=0, host fields are 0
// and the function does not error. /proc is not consulted.
func TestSnapshotHostFieldsZeroOnNoPID(t *testing.T) {
	vmObj := &models.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{})

	mock := &mockQMPClient{
		status: "running",
		cpus:   []VCPUInfo{{CPU: 0, ThreadID: 100}},
	}
	runner.qmpClient = mock

	// PID=0 by leaving cmdProcess nil
	runner.mu.Lock()
	runner.cmdProcess = nil
	runner.mu.Unlock()

	// proc readers are the production ones — they would fail or return
	// stale data if called, which would corrupt the test. Override them
	// to fail the test loudly if Snapshot calls them with PID=0.
	procCalled := false
	var procMu sync.Mutex
	runner.readProcessRSS = func(pid int) (uint64, error) {
		procMu.Lock()
		procCalled = true
		procMu.Unlock()
		return 0, fmt.Errorf("proc reader should not be called when PID=0")
	}
	runner.readProcessCPUJiffies = func(pid int) (uint64, error) {
		procMu.Lock()
		procCalled = true
		procMu.Unlock()
		return 0, fmt.Errorf("proc reader should not be called when PID=0")
	}

	m, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() unexpected error: %v", err)
	}

	if m.HostRSSBytes != 0 {
		t.Errorf("expected HostRSSBytes=0 when PID=0, got %d", m.HostRSSBytes)
	}
	if m.HostCPUJiffies != 0 {
		t.Errorf("expected HostCPUJiffies=0 when PID=0, got %d", m.HostCPUJiffies)
	}

	procMu.Lock()
	called := procCalled
	procMu.Unlock()
	if called {
		t.Error("Snapshot() should not call proc readers when PID=0")
	}
}

// S4: Snapshot() returns 0/0 with no error when PID is valid but /proc is
// unreadable (e.g. process exited between PID discovery and read).
func TestSnapshotHostFieldsZeroOnUnreadableProc(t *testing.T) {
	vmObj := &models.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{})

	mock := &mockQMPClient{
		status: "running",
		cpus:   []VCPUInfo{{CPU: 0, ThreadID: 100}},
	}
	runner.qmpClient = mock

	// Use a known-nonexistent PID via os.FindProcess (no /proc/<pid>/status)
	runner.mu.Lock()
	runner.cmdProcess, _ = os.FindProcess(999999999)
	runner.mu.Unlock()

	// Override to production readers — they will fail cleanly.
	runner.readProcessRSS = readProcessRSS
	runner.readProcessCPUJiffies = readProcessCPUJiffies

	m, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() should not surface /proc errors: %v", err)
	}

	if m.HostRSSBytes != 0 {
		t.Errorf("expected HostRSSBytes=0 when /proc unreadable, got %d", m.HostRSSBytes)
	}
	if m.HostCPUJiffies != 0 {
		t.Errorf("expected HostCPUJiffies=0 when /proc unreadable, got %d", m.HostCPUJiffies)
	}
}

// S4: Snapshot() populates host fields when PID is valid and /proc readable.
func TestSnapshotHostFieldsPopulatedOnValidPID(t *testing.T) {
	vmObj := &models.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{})

	mock := &mockQMPClient{
		status: "running",
		cpus:   []VCPUInfo{{CPU: 0, ThreadID: 100}},
	}
	runner.qmpClient = mock

	runner.mu.Lock()
	runner.cmdProcess, _ = os.FindProcess(os.Getpid())
	runner.mu.Unlock()

	// Mock the proc readers to return known values
	runner.readProcessRSS = func(pid int) (uint64, error) {
		return 100 * 1024 * 1024, nil // 100 MiB
	}
	runner.readProcessCPUJiffies = func(pid int) (uint64, error) {
		return 50_000, nil // 50_000 jiffies = 500 seconds of CPU time
	}

	m, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() unexpected error: %v", err)
	}

	if m.HostRSSBytes != 100*1024*1024 {
		t.Errorf("expected HostRSSBytes=%d, got %d", 100*1024*1024, m.HostRSSBytes)
	}
	// Cold snapshot: HostCPUJiffies should be 0 (CPU% unknown until next call)
	if m.HostCPUJiffies != 0 {
		t.Errorf("cold snapshot: expected HostCPUJiffies=0, got %d", m.HostCPUJiffies)
	}
}

// S4: warm Snapshot() computes host CPU% from deltas.
func TestSnapshotHostCPUDelta(t *testing.T) {
	vmObj := &models.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{})

	mock := &mockQMPClient{
		status: "running",
		cpus:   []VCPUInfo{{CPU: 0, ThreadID: 100}},
	}
	runner.qmpClient = mock

	runner.mu.Lock()
	runner.cmdProcess, _ = os.FindProcess(os.Getpid())
	runner.mu.Unlock()

	// Mock with increasing jiffies across calls
	var jiffiesCallCount int
	var jiffiesMu sync.Mutex
	runner.readProcessRSS = func(pid int) (uint64, error) {
		return 100 * 1024 * 1024, nil
	}
	runner.readProcessCPUJiffies = func(pid int) (uint64, error) {
		jiffiesMu.Lock()
		jiffiesCallCount++
		c := jiffiesCallCount
		jiffiesMu.Unlock()
		// First call (cold): 10_000 jiffies
		// Subsequent calls (warm): +100 jiffies each
		// We want the warm call to show non-zero CPU%
		return uint64(10_000 + 100*(c-1)), nil
	}

	// Cold snapshot — establishes baseline
	_, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("first Snapshot() error: %v", err)
	}

	// Wait briefly to ensure measurable wall-clock delta
	time.Sleep(20 * time.Millisecond)

	// Warm snapshot — should compute CPU% from delta
	m, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("second Snapshot() error: %v", err)
	}

	// HostCPUJiffies holds CPU% * 100 (consistent with the S3 VCPUs convention).
	// With a 100-jiffy delta and ~20ms wall clock, the CPU% should be
	// well above 0% (100 jiffies = 1 second of CPU time consumed in 20ms
	// wall time is 5000% — clamped to 100% = 10000 in fixed-point).
	// The important assertion is that it is non-zero.
	if m.HostCPUJiffies == 0 {
		t.Error("warm snapshot: expected non-zero HostCPUJiffies after CPU time advance")
	}
	// And it should be clamped to 100% max (= 10000 fixed-point)
	if m.HostCPUJiffies > 10000 {
		t.Errorf("warm snapshot: expected HostCPUJiffies <= 10000 (100%%), got %d", m.HostCPUJiffies)
	}
}
