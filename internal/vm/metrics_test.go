package vm

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/domain"
)

// mockQMPClient implements QMPClientInterface for testing.
type mockQMPClient struct {
	status  string
	cpus    []VCPUInfo
	blocks  []QMPBlockDeviceStats
	balloon uint64
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

func (m *mockQMPClient) QueryBlockStats() ([]QMPBlockDeviceStats, error) {
	if m.qError != nil {
		return nil, m.qError
	}
	return m.blocks, nil
}

func (m *mockQMPClient) QueryBalloon() (uint64, error) {
	if m.qError != nil {
		return 0, m.qError
	}
	return m.balloon, nil
}

func (m *mockQMPClient) Close() error            { return nil }
func (m *mockQMPClient) Quit() error             { return nil }
func (m *mockQMPClient) Events() <-chan QMPEvent { return nil }

func TestSnapshotColdSnapshot(t *testing.T) {
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

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
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

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

	// Mock proc reader: returns increasing CPU times per snapshot.
	// Uses a snapshot counter rather than call-count to avoid coupling
	// to how many times readThreadCPUTime is called per snapshot.
	var snapshotCount int
	var muSnap sync.Mutex
	runner.readThreadCPUTime = func(pid, tid int) (int64, error) {
		muSnap.Lock()
		s := snapshotCount
		muSnap.Unlock()
		if s == 0 {
			return 100_000_000, nil // 100ms in ns (first snapshot)
		}
		return 200_000_000, nil // 200ms in ns (second snapshot)
	}

	// Cold snapshot (establishes baseline)
	_, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("first Snapshot() error: %v", err)
	}

	// Advance snapshot counter so the warm snapshot sees higher CPU time.
	muSnap.Lock()
	snapshotCount++
	muSnap.Unlock()

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
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

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
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

	mock := &mockQMPClient{
		qError: fmt.Errorf("mock failure"),
	}
	runner.qmpClient = mock

	_, err := runner.Snapshot()
	if err == nil {
		t.Error("expected error from Snapshot() with failing QMP")
	}
}

// S5: Snapshot() populates BlockDevices and BalloonBytes from the QMP
// client. The fields are populated with raw counters; the per-disk
// B/s and IOPS math lives in the warm-snapshot test below.
func TestSnapshotBlockAndBalloonPopulated(t *testing.T) {
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

	mock := &mockQMPClient{
		status: "running",
		cpus:   []VCPUInfo{{CPU: 0, ThreadID: 100}},
		blocks: []QMPBlockDeviceStats{
			{Device: "drive0", Stats: QMPBlockDeviceIO{
				RDBytes: 1024, WRBytes: 2048, RDOps: 10, WROps: 20,
			}},
		},
		balloon: 2147483648,
	}
	runner.qmpClient = mock

	// PID=0 to avoid /proc reads
	runner.mu.Lock()
	runner.cmdProcess = nil
	runner.mu.Unlock()

	runner.readProcessRSS = func(pid int) (uint64, error) { return 0, nil }
	runner.readProcessCPUJiffies = func(pid int) (uint64, error) { return 0, nil }

	m, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() unexpected error: %v", err)
	}

	if len(m.BlockDevices) != 1 {
		t.Fatalf("Expected 1 BlockDevice, got %d", len(m.BlockDevices))
	}
	if m.BlockDevices[0].Device != "drive0" {
		t.Errorf("Expected device='drive0', got '%s'", m.BlockDevices[0].Device)
	}
	if m.BlockDevices[0].RDBytes != 1024 {
		t.Errorf("Expected RDBytes=1024, got %d", m.BlockDevices[0].RDBytes)
	}
	if m.BlockDevices[0].WRBytes != 2048 {
		t.Errorf("Expected WRBytes=2048, got %d", m.BlockDevices[0].WRBytes)
	}
	if m.BlockDevices[0].RDOps != 10 {
		t.Errorf("Expected RDOps=10, got %d", m.BlockDevices[0].RDOps)
	}
	if m.BlockDevices[0].WROps != 20 {
		t.Errorf("Expected WROps=20, got %d", m.BlockDevices[0].WROps)
	}
	if m.BalloonBytes != 2147483648 {
		t.Errorf("Expected BalloonBytes=2147483648, got %d", m.BalloonBytes)
	}
}

// S5: Snapshot() should be resilient when the guest has no balloon driver
// (QueryBalloon returns 0 with no error via graceful degradation) and when
// query-blockstats returns an empty array.
func TestSnapshotBlockAndBalloonGracefulNoBalloon(t *testing.T) {
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

	mock := &mockQMPClient{
		status:  "running",
		cpus:    []VCPUInfo{{CPU: 0, ThreadID: 100}},
		blocks:  nil, // no disks attached / empty
		balloon: 0,   // no balloon driver
	}
	runner.qmpClient = mock

	runner.mu.Lock()
	runner.cmdProcess = nil
	runner.mu.Unlock()

	runner.readProcessRSS = func(pid int) (uint64, error) { return 0, nil }
	runner.readProcessCPUJiffies = func(pid int) (uint64, error) { return 0, nil }

	m, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() should not error on no-balloon / no-block case: %v", err)
	}
	if m.BalloonBytes != 0 {
		t.Errorf("Expected BalloonBytes=0, got %d", m.BalloonBytes)
	}
	if len(m.BlockDevices) != 0 {
		t.Errorf("Expected 0 BlockDevices, got %d", len(m.BlockDevices))
	}
}

// S5: warm Snapshot() computes per-disk B/s and IOPS from deltas.
// On the first call (cold), the B/s and IOPS fields are zero. On the second
// call, they reflect the delta in raw counters divided by the wall-clock delta.
func TestSnapshotBlockDelta(t *testing.T) {
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

	var blockCall int
	mock := &mockQMPClient{
		status: "running",
		cpus:   []VCPUInfo{{CPU: 0, ThreadID: 100}},
		blocks: []QMPBlockDeviceStats{
			{Device: "drive0", Stats: QMPBlockDeviceIO{
				RDBytes: 0, WRBytes: 0, RDOps: 0, WROps: 0,
			}},
		},
	}
	runner.qmpClient = mock

	// Mutate raw counters across calls to simulate I/O activity.
	// blockCall is incremented BEFORE the rdBytes/wrBytes/rdOps/wrOps
	// functions are evaluated. We use blockCall-1 so the FIRST call
	// sees "0" (cold snapshot, no delta) and the SECOND call sees the
	// "advanced" values (warm snapshot, non-zero delta).
	runner.qmpClient = &countingBlockClient{
		mock: mock,
		onCall: func() {
			blockCall++
		},
		rdBytes: func() uint64 {
			// Call 1 (cold): 0; Call 2+ (warm): 1 MiB
			if blockCall <= 1 {
				return 0
			}
			return 1024 * 1024
		},
		wrBytes: func() uint64 {
			if blockCall <= 1 {
				return 0
			}
			return 2 * 1024 * 1024
		},
		rdOps: func() uint64 {
			if blockCall <= 1 {
				return 0
			}
			return 100
		},
		wrOps: func() uint64 {
			if blockCall <= 1 {
				return 0
			}
			return 200
		},
	}

	runner.mu.Lock()
	runner.cmdProcess = nil
	runner.mu.Unlock()
	runner.readProcessRSS = func(pid int) (uint64, error) { return 0, nil }
	runner.readProcessCPUJiffies = func(pid int) (uint64, error) { return 0, nil }

	// Cold snapshot — deltas should be 0.
	m1, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("first Snapshot() error: %v", err)
	}
	if len(m1.BlockDevices) != 1 {
		t.Fatalf("cold: expected 1 BlockDevice, got %d", len(m1.BlockDevices))
	}
	if m1.BlockDevices[0].RDBps != 0 || m1.BlockDevices[0].WRBps != 0 {
		t.Errorf("cold: expected Bps=0, got r=%d w=%d",
			m1.BlockDevices[0].RDBps, m1.BlockDevices[0].WRBps)
	}
	if m1.BlockDevices[0].RDIOPS != 0 || m1.BlockDevices[0].WRIOPS != 0 {
		t.Errorf("cold: expected IOPS=0, got r=%d w=%d",
			m1.BlockDevices[0].RDIOPS, m1.BlockDevices[0].WRIOPS)
	}

	// Wait a measurable wall-clock delta
	time.Sleep(20 * time.Millisecond)

	// Warm snapshot — deltas should be non-zero.
	m2, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("second Snapshot() error: %v", err)
	}
	if len(m2.BlockDevices) != 1 {
		t.Fatalf("warm: expected 1 BlockDevice, got %d", len(m2.BlockDevices))
	}
	// 1 MiB read delta in 20ms wall = 50 MiB/s. Allow a wide tolerance for CI.
	if m2.BlockDevices[0].RDBps == 0 {
		t.Error("warm: expected non-zero RDBps after read delta")
	}
	if m2.BlockDevices[0].WRBps == 0 {
		t.Error("warm: expected non-zero WRBps after write delta")
	}
	if m2.BlockDevices[0].RDIOPS == 0 {
		t.Error("warm: expected non-zero RDIOPS after read ops delta")
	}
	if m2.BlockDevices[0].WRIOPS == 0 {
		t.Error("warm: expected non-zero WRIOPS after write ops delta")
	}
}

// countingBlockClient is a QMPClientInterface that delegates everything
// to the inner mock except QueryBlockStats, which builds the array from
// per-call functions so we can simulate counter growth across snapshots.
type countingBlockClient struct {
	mock    *mockQMPClient
	onCall  func()
	rdBytes func() uint64
	wrBytes func() uint64
	rdOps   func() uint64
	wrOps   func() uint64
}

func (c *countingBlockClient) QueryStatus() (string, error)   { return c.mock.QueryStatus() }
func (c *countingBlockClient) QueryCPUs() ([]VCPUInfo, error) { return c.mock.QueryCPUs() }
func (c *countingBlockClient) QueryCPUsFast() ([]QMPVCPUInfo, error) {
	return c.mock.QueryCPUsFast()
}
func (c *countingBlockClient) QueryBalloon() (uint64, error) { return c.mock.QueryBalloon() }
func (c *countingBlockClient) Close() error                  { return nil }
func (c *countingBlockClient) Quit() error                   { return nil }
func (c *countingBlockClient) Events() <-chan QMPEvent       { return nil }
func (c *countingBlockClient) QueryBlockStats() ([]QMPBlockDeviceStats, error) {
	c.onCall()
	return []QMPBlockDeviceStats{
		{Device: "drive0", Stats: QMPBlockDeviceIO{
			RDBytes: c.rdBytes(),
			WRBytes: c.wrBytes(),
			RDOps:   c.rdOps(),
			WROps:   c.wrOps(),
		}},
	}, nil
}

// S5: when the guest has no balloon driver, the QueryBalloon call inside
// Snapshot() must be gracefully tolerated — Snapshot returns the metrics
// with BalloonBytes=0 and no error, and other fields are still populated.
func TestSnapshotBalloonNotActivatedGraceful(t *testing.T) {
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

	// The mock's QueryBalloon returns 0 with no error, mirroring the
	// production "Balloon is not activated" graceful-degradation path.
	mock := &mockQMPClient{
		status: "running",
		cpus:   []VCPUInfo{{CPU: 0, ThreadID: 100}},
		blocks: []QMPBlockDeviceStats{
			{Device: "drive0", Stats: QMPBlockDeviceIO{
				RDBytes: 100, WRBytes: 200, RDOps: 1, WROps: 2,
			}},
		},
		balloon: 0,
	}
	runner.qmpClient = mock

	runner.mu.Lock()
	runner.cmdProcess = nil
	runner.mu.Unlock()
	runner.readProcessRSS = func(pid int) (uint64, error) { return 0, nil }
	runner.readProcessCPUJiffies = func(pid int) (uint64, error) { return 0, nil }

	m, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() should not error when no balloon driver: %v", err)
	}
	if m.BalloonBytes != 0 {
		t.Errorf("Expected BalloonBytes=0 when no balloon driver, got %d", m.BalloonBytes)
	}
	// Other fields are still populated
	if len(m.BlockDevices) != 1 {
		t.Errorf("Expected 1 BlockDevice (should be populated even without balloon), got %d", len(m.BlockDevices))
	}
}

func TestPID(t *testing.T) {
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

	if pid := runner.PID(); pid != 0 {
		t.Errorf("expected PID=0 for unstarted runner, got %d", pid)
	}
}

// S4 acceptance criteria: given a snapshot with PID=0, host fields are 0
// and the function does not error. /proc is not consulted.
func TestSnapshotHostFieldsZeroOnNoPID(t *testing.T) {
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

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
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

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
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

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
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

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

// Issue #66: Snapshot() must call readThreadCPUTime exactly once per vCPU
// per snapshot call, not twice (double /proc read bug).
func TestSnapshotReadThreadCPUTimeCalledOncePerVCPU(t *testing.T) {
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

	mock := &mockQMPClient{
		status: "running",
		cpus: []VCPUInfo{
			{CPU: 0, ThreadID: 100},
			{CPU: 1, ThreadID: 101},
		},
	}
	runner.qmpClient = mock

	runner.mu.Lock()
	runner.cmdProcess, _ = os.FindProcess(os.Getpid())
	runner.mu.Unlock()

	var callCount int
	var muCallCount sync.Mutex
	runner.readThreadCPUTime = func(pid, tid int) (int64, error) {
		muCallCount.Lock()
		callCount++
		muCallCount.Unlock()
		return int64(tid) * 10000000, nil
	}

	// One snapshot — callCount should be 2 (one per vCPU), not 4.
	_, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() unexpected error: %v", err)
	}

	muCallCount.Lock()
	got := callCount
	muCallCount.Unlock()

	if got != 2 {
		t.Errorf("expected readThreadCPUTime called 2 times (once per vCPU), got %d — double read bug", got)
	}

	// Second snapshot should also call it 2 times (total 4, not 8)
	_, err = runner.Snapshot()
	if err != nil {
		t.Fatalf("second Snapshot() unexpected error: %v", err)
	}

	muCallCount.Lock()
	got = callCount
	muCallCount.Unlock()

	if got != 4 {
		t.Errorf("expected readThreadCPUTime called 4 times across two snapshots (2 per snap), got %d — double read bug", got)
	}
}

// Issue #66: Snapshot() with no PID (process not started) should not call
// readThreadCPUTime at all, and should not double-read.
func TestSnapshotReadThreadCPUTimeNotCalledWhenNoPID(t *testing.T) {
	vmObj := &domain.VM{Name: "test-vm", ID: "1"}
	cfg := &config.Config{}
	runner := NewVMRunner(vmObj, cfg, RunConfig{}, false)

	mock := &mockQMPClient{
		status: "running",
		cpus: []VCPUInfo{
			{CPU: 0, ThreadID: 100},
		},
	}
	runner.qmpClient = mock

	runner.mu.Lock()
	runner.cmdProcess = nil // PID=0 — no /proc reading
	runner.mu.Unlock()

	var callCount int
	var muCallCount sync.Mutex
	runner.readThreadCPUTime = func(pid, tid int) (int64, error) {
		muCallCount.Lock()
		callCount++
		muCallCount.Unlock()
		return 0, nil
	}

	_, err := runner.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() unexpected error: %v", err)
	}

	muCallCount.Lock()
	got := callCount
	muCallCount.Unlock()

	if got != 0 {
		t.Errorf("expected readThreadCPUTime not called when PID=0, got %d calls", got)
	}
}
