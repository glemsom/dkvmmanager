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
