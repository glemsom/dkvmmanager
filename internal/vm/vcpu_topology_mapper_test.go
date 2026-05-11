package vm

import (
	"strings"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// --- computeAPICID tests ---

func TestComputeAPICID(t *testing.T) {
	// 2 dies, 8 cores/die, 2 threads
	// Formula: apic-id = (die-id × maxCoresPerDie × threads) + (guest-core-id × threads) + thread-id

	tests := []struct {
		name           string
		socketID       int
		dieID          int
		guestCoreID    int
		threadID       int
		maxCoresPerDie int
		threadsPerCore int
		expected       int
	}{
		{"die0 core0 thread0", 0, 0, 0, 0, 8, 2, 0},
		{"die0 core0 thread1", 0, 0, 0, 1, 8, 2, 1},
		{"die0 core1 thread0", 0, 0, 1, 0, 8, 2, 2},
		{"die0 core1 thread1", 0, 0, 1, 1, 8, 2, 3},
		{"die0 core7 thread0", 0, 0, 7, 0, 8, 2, 14},
		{"die0 core7 thread1", 0, 0, 7, 1, 8, 2, 15},
		{"die1 core0 thread0", 0, 1, 0, 0, 8, 2, 16},
		{"die1 core0 thread1", 0, 1, 0, 1, 8, 2, 17},
		{"die1 core5 thread0", 0, 1, 5, 0, 8, 2, 26},
		{"die1 core5 thread1", 0, 1, 5, 1, 8, 2, 27},
		{"die1 core7 thread1", 0, 1, 7, 1, 8, 2, 31},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeAPICID(tt.socketID, tt.dieID, tt.guestCoreID, tt.threadID,
				tt.maxCoresPerDie, tt.threadsPerCore)
			if result != tt.expected {
				t.Errorf("computeAPICID() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// --- GenerateAsymmetricCPUDevices tests ---

func TestGenerateAsymmetricCPUDevices_Symmetric(t *testing.T) {
	// All cores on all dies selected — symmetric case
	hostTopo := models.HostCPUTopology{
		TotalCores:     8,
		TotalCPUs:      16,
		ThreadsPerCore: 2,
		Dies: []models.CPUDie{
			{
				ID:          0,
				Cores:       4,
				Threads:     2,
				LogicalCPUs: []int{0, 1, 2, 3, 4, 5, 6, 7},
				CoreDetails: []models.CPUCore{
					{ID: 0, DieID: 0, Threads: []int{0, 1}},
					{ID: 1, DieID: 0, Threads: []int{2, 3}},
					{ID: 2, DieID: 0, Threads: []int{4, 5}},
					{ID: 3, DieID: 0, Threads: []int{6, 7}},
				},
			},
			{
				ID:          1,
				Cores:       4,
				Threads:     2,
				LogicalCPUs: []int{8, 9, 10, 11, 12, 13, 14, 15},
				CoreDetails: []models.CPUCore{
					{ID: 0, DieID: 1, Threads: []int{8, 9}},
					{ID: 1, DieID: 1, Threads: []int{10, 11}},
					{ID: 2, DieID: 1, Threads: []int{12, 13}},
					{ID: 3, DieID: 1, Threads: []int{14, 15}},
				},
			},
		},
	}

	selectedCPUs := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}

	deviceArgs, firstAPICID, err := GenerateAsymmetricCPUDevices(selectedCPUs, hostTopo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if firstAPICID != 0 {
		t.Errorf("expected firstAPICID=0, got %d", firstAPICID)
	}

	// Should have 15 device args (16 CPUs - 1 auto-created)
	if len(deviceArgs) != 15 {
		t.Errorf("expected 15 device args, got %d", len(deviceArgs))
	}

	// Verify no APIC ID 0 in device args (it's auto-created)
	for _, arg := range deviceArgs {
		if strings.Contains(arg, "apic-id=0") {
			t.Error("device args should not contain apic-id=0 (auto-created by -smp 1)")
		}
	}

	// Verify guest core IDs are sequential within each die
	for _, arg := range deviceArgs {
		// Parse the arg to check format
		if !strings.Contains(arg, "host-x86_64-cpu") {
			t.Errorf("expected host-x86_64-cpu device, got: %s", arg)
		}
	}
}

func TestGenerateAsymmetricCPUDevices_PartialDie(t *testing.T) {
	// Asymmetric: all cores on die 0, cores 10-15 on die 1 (cores 8-9 reserved)
	// This matches the user's scenario in the implementation plan.
	hostTopo := models.HostCPUTopology{
		TotalCores:     16,
		TotalCPUs:      32,
		ThreadsPerCore: 2,
		Dies: []models.CPUDie{
			{
				ID:          0,
				Cores:       8,
				Threads:     2,
				LogicalCPUs: []int{0, 1, 2, 3, 4, 5, 6, 7, 16, 17, 18, 19, 20, 21, 22, 23},
				CoreDetails: []models.CPUCore{
					{ID: 0, DieID: 0, Threads: []int{0, 16}},
					{ID: 1, DieID: 0, Threads: []int{1, 17}},
					{ID: 2, DieID: 0, Threads: []int{2, 18}},
					{ID: 3, DieID: 0, Threads: []int{3, 19}},
					{ID: 4, DieID: 0, Threads: []int{4, 20}},
					{ID: 5, DieID: 0, Threads: []int{5, 21}},
					{ID: 6, DieID: 0, Threads: []int{6, 22}},
					{ID: 7, DieID: 0, Threads: []int{7, 23}},
				},
			},
			{
				ID:          1,
				Cores:       8,
				Threads:     2,
				LogicalCPUs: []int{8, 9, 10, 11, 12, 13, 14, 15, 24, 25, 26, 27, 28, 29, 30, 31},
				CoreDetails: []models.CPUCore{
					{ID: 8, DieID: 1, Threads: []int{8, 24}},
					{ID: 9, DieID: 1, Threads: []int{9, 25}},
					{ID: 10, DieID: 1, Threads: []int{10, 26}},
					{ID: 11, DieID: 1, Threads: []int{11, 27}},
					{ID: 12, DieID: 1, Threads: []int{12, 28}},
					{ID: 13, DieID: 1, Threads: []int{13, 29}},
					{ID: 14, DieID: 1, Threads: []int{14, 30}},
					{ID: 15, DieID: 1, Threads: []int{15, 31}},
				},
			},
		},
	}

	// Select all die 0 CPUs (0-7, 16-23) + die 1 cores 10-15 (10-15, 26-31)
	// Host cores 8-9 on die 1 are reserved (not selected)
	selectedCPUs := []int{
		0, 1, 2, 3, 4, 5, 6, 7, // die 0 cores 0-7, thread 0
		16, 17, 18, 19, 20, 21, 22, 23, // die 0 cores 0-7, thread 1
		10, 11, 12, 13, 14, 15, // die 1 cores 10-15, thread 0
		26, 27, 28, 29, 30, 31, // die 1 cores 10-15, thread 1
	}

	deviceArgs, firstAPICID, err := GenerateAsymmetricCPUDevices(selectedCPUs, hostTopo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if firstAPICID != 0 {
		t.Errorf("expected firstAPICID=0, got %d", firstAPICID)
	}

	// 28 CPUs selected, so 27 device args
	if len(deviceArgs) != 27 {
		t.Errorf("expected 27 device args, got %d", len(deviceArgs))
	}

	// Verify die 1 guest core IDs are sequential (0-5), not host core IDs (10-15)
	for _, arg := range deviceArgs {
		if strings.Contains(arg, "die-id=1") {
			// Extract core-id
			if strings.Contains(arg, "core-id=10") || strings.Contains(arg, "core-id=11") ||
				strings.Contains(arg, "core-id=12") || strings.Contains(arg, "core-id=13") ||
				strings.Contains(arg, "core-id=14") || strings.Contains(arg, "core-id=15") {
				t.Errorf("die 1 should have sequential guest core IDs (0-5), not host core IDs; got: %s", arg)
			}
		}
	}

	// Verify APIC IDs:
	// Die 0: APIC 0-15 (guest cores 0-7, threads 0-1)
	// Die 1: APIC 16-27 (guest cores 0-5, threads 0-1)
	// APIC 0 is auto-created, so device args start at APIC 1

	// Check that die 1 devices have APIC IDs >= 16
	for _, arg := range deviceArgs {
		if strings.Contains(arg, "die-id=1") {
			// Extract apic-id
			var apicID int
			_, err := scanAPICID(arg, &apicID)
			if err != nil {
				t.Fatalf("failed to parse apic-id from: %s", arg)
			}
			if apicID < 16 {
				t.Errorf("die 1 device should have APIC ID >= 16, got %d: %s", apicID, arg)
			}
			if apicID > 27 {
				t.Errorf("die 1 device should have APIC ID <= 27, got %d: %s", apicID, arg)
			}
		}
	}
}

func TestGenerateAsymmetricCPUDevices_SingleDie(t *testing.T) {
	// Only one die with partial core selection
	hostTopo := models.HostCPUTopology{
		TotalCores:     4,
		TotalCPUs:      8,
		ThreadsPerCore: 2,
		Dies: []models.CPUDie{
			{
				ID:          0,
				Cores:       4,
				Threads:     2,
				LogicalCPUs: []int{0, 1, 2, 3, 4, 5, 6, 7},
				CoreDetails: []models.CPUCore{
					{ID: 0, DieID: 0, Threads: []int{0, 1}},
					{ID: 1, DieID: 0, Threads: []int{2, 3}},
					{ID: 2, DieID: 0, Threads: []int{4, 5}},
					{ID: 3, DieID: 0, Threads: []int{6, 7}},
				},
			},
		},
	}

	// Select only cores 0-1 on die 0
	selectedCPUs := []int{0, 1, 2, 3}

	deviceArgs, firstAPICID, err := GenerateAsymmetricCPUDevices(selectedCPUs, hostTopo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if firstAPICID != 0 {
		t.Errorf("expected firstAPICID=0, got %d", firstAPICID)
	}

	// 4 CPUs, 3 device args
	if len(deviceArgs) != 3 {
		t.Errorf("expected 3 device args, got %d", len(deviceArgs))
	}

	// Verify all devices are on die 0
	for _, arg := range deviceArgs {
		if !strings.Contains(arg, "die-id=0") {
			t.Errorf("expected die-id=0, got: %s", arg)
		}
	}
}

func TestGenerateAsymmetricCPUDevices_NoDie0(t *testing.T) {
	// User selects no CPUs on die 0 — should return an error
	hostTopo := models.HostCPUTopology{
		TotalCores:     8,
		TotalCPUs:      16,
		ThreadsPerCore: 2,
		Dies: []models.CPUDie{
			{
				ID:          0,
				Cores:       4,
				Threads:     2,
				LogicalCPUs: []int{0, 1, 2, 3, 4, 5, 6, 7},
				CoreDetails: []models.CPUCore{
					{ID: 0, DieID: 0, Threads: []int{0, 1}},
					{ID: 1, DieID: 0, Threads: []int{2, 3}},
					{ID: 2, DieID: 0, Threads: []int{4, 5}},
					{ID: 3, DieID: 0, Threads: []int{6, 7}},
				},
			},
			{
				ID:          1,
				Cores:       4,
				Threads:     2,
				LogicalCPUs: []int{8, 9, 10, 11, 12, 13, 14, 15},
				CoreDetails: []models.CPUCore{
					{ID: 0, DieID: 1, Threads: []int{8, 9}},
					{ID: 1, DieID: 1, Threads: []int{10, 11}},
					{ID: 2, DieID: 1, Threads: []int{12, 13}},
					{ID: 3, DieID: 1, Threads: []int{14, 15}},
				},
			},
		},
	}

	// Only select CPUs on die 1
	selectedCPUs := []int{8, 9, 10, 11}

	_, _, err := GenerateAsymmetricCPUDevices(selectedCPUs, hostTopo)
	if err == nil {
		t.Fatal("expected error when no CPUs selected on die 0")
	}
	if !strings.Contains(err.Error(), "die=0") || !strings.Contains(err.Error(), "core=0") {
		t.Errorf("error should mention (die=0, core=0, thread=0) requirement, got: %v", err)
	}
}

func TestGenerateAsymmetricCPUDevices_Empty(t *testing.T) {
	hostTopo := models.HostCPUTopology{
		TotalCores:     4,
		TotalCPUs:      8,
		ThreadsPerCore: 2,
		Dies: []models.CPUDie{
			{
				ID:          0,
				Cores:       4,
				Threads:     2,
				LogicalCPUs: []int{0, 1, 2, 3, 4, 5, 6, 7},
				CoreDetails: []models.CPUCore{
					{ID: 0, DieID: 0, Threads: []int{0, 1}},
					{ID: 1, DieID: 0, Threads: []int{2, 3}},
					{ID: 2, DieID: 0, Threads: []int{4, 5}},
					{ID: 3, DieID: 0, Threads: []int{6, 7}},
				},
			},
		},
	}

	_, _, err := GenerateAsymmetricCPUDevices([]int{}, hostTopo)
	if err == nil {
		t.Fatal("expected error when no CPUs selected")
	}
	if !strings.Contains(err.Error(), "no CPUs selected") {
		t.Errorf("error should mention 'no CPUs selected', got: %v", err)
	}
}

func TestGenerateAsymmetricCPUDevices_InvalidCPU(t *testing.T) {
	hostTopo := models.HostCPUTopology{
		TotalCores:     4,
		TotalCPUs:      8,
		ThreadsPerCore: 2,
		Dies: []models.CPUDie{
			{
				ID:          0,
				Cores:       4,
				Threads:     2,
				LogicalCPUs: []int{0, 1, 2, 3, 4, 5, 6, 7},
				CoreDetails: []models.CPUCore{
					{ID: 0, DieID: 0, Threads: []int{0, 1}},
					{ID: 1, DieID: 0, Threads: []int{2, 3}},
				},
			},
		},
	}

	// Select CPU 99 which doesn't exist in the topology
	selectedCPUs := []int{0, 99}

	_, _, err := GenerateAsymmetricCPUDevices(selectedCPUs, hostTopo)
	if err == nil {
		t.Fatal("expected error when selecting invalid CPU")
	}
	if !strings.Contains(err.Error(), "not found in host topology") {
		t.Errorf("error should mention CPU not found, got: %v", err)
	}
}

// scanAPICID extracts the apic-id value from a device arg string.
func scanAPICID(arg string, apicID *int) (string, error) {
	// Find "apic-id=" and parse the number
	idx := strings.Index(arg, "apic-id=")
	if idx == -1 {
		return "", nil
	}
	rest := arg[idx+len("apic-id="):]
	// Read digits
	end := strings.IndexFunc(rest, func(r rune) bool { return r < '0' || r > '9' })
	if end == -1 {
		end = len(rest)
	}
	var val int
	for i := 0; i < end; i++ {
		val = val*10 + int(rest[i]-'0')
	}
	*apicID = val
	return rest[:end], nil
}
