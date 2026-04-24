package vm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// createMockSysfs creates a fake sysfs tree for testing CPU topology detection.
// Simulates a system with 2 dies, 4 cores per die, 2 threads per core (16 logical CPUs total).
func createMockSysfs(t *testing.T) string {
	t.Helper()
	base := t.TempDir()
	cpuBase := filepath.Join(base, "sys", "devices", "system", "cpu")
	os.MkdirAll(cpuBase, 0755)

	// Create 16 CPUs: 2 dies, 8 cores per die, 2 threads per core
	// Die 0: CPUs 0-7 (cores 0-3, each with 2 threads)
	// Die 1: CPUs 8-15 (cores 0-3, each with 2 threads)
	for cpuID := 0; cpuID < 16; cpuID++ {
		cpuDir := filepath.Join(cpuBase, "cpu"+itoa(cpuID))
		topoDir := filepath.Join(cpuDir, "topology")
		os.MkdirAll(topoDir, 0755)

		dieID := cpuID / 8
		coreID := (cpuID % 8) / 2

		// Write die_id
		os.WriteFile(filepath.Join(topoDir, "die_id"), []byte(itoa(dieID)+"\n"), 0644)
		// Write core_id
		os.WriteFile(filepath.Join(topoDir, "core_id"), []byte(itoa(coreID)+"\n"), 0644)

		// Write thread_siblings_list
		thread0 := cpuID
		thread1 := cpuID + 1
		if cpuID%2 == 1 {
			thread0 = cpuID - 1
			thread1 = cpuID
		}
		os.WriteFile(filepath.Join(topoDir, "thread_siblings_list"), []byte(itoa(thread0)+"-"+itoa(thread1)+"\n"), 0644)

		// Write online (for all except cpu0 which is always online)
		if cpuID > 0 {
			os.WriteFile(filepath.Join(cpuDir, "online"), []byte("1\n"), 0644)
		}

		// Create L3 cache dir (shared per die, so put it on first CPU of each die)
		if cpuID%8 == 0 {
			cacheDir := filepath.Join(cpuDir, "cache", "index0")
			os.MkdirAll(cacheDir, 0755)
			os.WriteFile(filepath.Join(cacheDir, "level"), []byte("3\n"), 0644)
			os.WriteFile(filepath.Join(cacheDir, "size"), []byte("32M\n"), 0644)
		}
	}

	return cpuBase
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}

func TestCPUScannerScanTopology(t *testing.T) {
	sysfsPath := createMockSysfs(t)
	scanner := NewCPUScannerWithPath(sysfsPath)

	topo, err := scanner.ScanTopology()
	if err != nil {
		t.Fatalf("ScanTopology() returned error: %v", err)
	}

	// Verify total CPUs
	if topo.TotalCPUs != 16 {
		t.Errorf("TotalCPUs = %d, want 16", topo.TotalCPUs)
	}

	// Verify total cores (8 physical cores)
	if topo.TotalCores != 8 {
		t.Errorf("TotalCores = %d, want 8", topo.TotalCores)
	}

	// Verify threads per core
	if topo.ThreadsPerCore != 2 {
		t.Errorf("ThreadsPerCore = %d, want 2", topo.ThreadsPerCore)
	}

	// Verify die count
	if len(topo.Dies) != 2 {
		t.Fatalf("Dies count = %d, want 2", len(topo.Dies))
	}

	// Die 0
	die0 := topo.Dies[0]
	if die0.ID != 0 {
		t.Errorf("Die[0].ID = %d, want 0", die0.ID)
	}
	if die0.Cores != 4 {
		t.Errorf("Die[0].Cores = %d, want 4", die0.Cores)
	}
	if len(die0.LogicalCPUs) != 8 {
		t.Errorf("Die[0].LogicalCPUs length = %d, want 8", len(die0.LogicalCPUs))
	}
	if die0.L3CacheKB != 32768 {
		t.Errorf("Die[0].L3CacheKB = %d, want 32768", die0.L3CacheKB)
	}

	// Die 1
	die1 := topo.Dies[1]
	if die1.ID != 1 {
		t.Errorf("Die[1].ID = %d, want 1", die1.ID)
	}
	if die1.Cores != 4 {
		t.Errorf("Die[1].Cores = %d, want 4", die1.Cores)
	}
	if len(die1.LogicalCPUs) != 8 {
		t.Errorf("Die[1].LogicalCPUs length = %d, want 8", len(die1.LogicalCPUs))
	}
}

func TestCPUScannerSingleDie(t *testing.T) {
	base := t.TempDir()
	cpuBase := filepath.Join(base, "sys", "devices", "system", "cpu")
	os.MkdirAll(cpuBase, 0755)

	// Create 4 CPUs on a single die
	for cpuID := 0; cpuID < 4; cpuID++ {
		cpuDir := filepath.Join(cpuBase, "cpu"+itoa(cpuID))
		topoDir := filepath.Join(cpuDir, "topology")
		os.MkdirAll(topoDir, 0755)

		os.WriteFile(filepath.Join(topoDir, "die_id"), []byte("0\n"), 0644)
		os.WriteFile(filepath.Join(topoDir, "core_id"), []byte(itoa(cpuID)+"\n"), 0644)
		os.WriteFile(filepath.Join(topoDir, "thread_siblings_list"), []byte(itoa(cpuID)+"\n"), 0644)

		if cpuID > 0 {
			os.WriteFile(filepath.Join(cpuDir, "online"), []byte("1\n"), 0644)
		}
	}

	scanner := NewCPUScannerWithPath(cpuBase)
	topo, err := scanner.ScanTopology()
	if err != nil {
		t.Fatalf("ScanTopology() returned error: %v", err)
	}

	if len(topo.Dies) != 1 {
		t.Fatalf("Dies count = %d, want 1", len(topo.Dies))
	}
	if topo.TotalCPUs != 4 {
		t.Errorf("TotalCPUs = %d, want 4", topo.TotalCPUs)
	}
	if topo.TotalCores != 4 {
		t.Errorf("TotalCores = %d, want 4", topo.TotalCores)
	}
	if topo.ThreadsPerCore != 1 {
		t.Errorf("ThreadsPerCore = %d, want 1", topo.ThreadsPerCore)
	}
}

func TestCPUScannerEmptyDir(t *testing.T) {
	base := t.TempDir()
	cpuBase := filepath.Join(base, "sys", "devices", "system", "cpu")
	os.MkdirAll(cpuBase, 0755)

	scanner := NewCPUScannerWithPath(cpuBase)
	_, err := scanner.ScanTopology()
	if err == nil {
		t.Error("Expected error for empty CPU directory, got nil")
	}
}

func TestCPUScannerNonexistentDir(t *testing.T) {
	scanner := NewCPUScannerWithPath("/nonexistent/path")
	_, err := scanner.ScanTopology()
	if err == nil {
		t.Error("Expected error for nonexistent sysfs path, got nil")
	}
}

func TestParseCacheSizeKB(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"32M", 32768},
		{"96M", 98304},
		{"1024K", 1024},
		{"1G", 1048576},
		{"512K", 512},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		result := parseCacheSizeKB(tt.input)
		if result != tt.expected {
			t.Errorf("parseCacheSizeKB(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestParseIntList(t *testing.T) {
	tests := []struct {
		input    string
		expected []int
	}{
		{"0-3", []int{0, 1, 2, 3}},
		{"0,1,2,3", []int{0, 1, 2, 3}},
		{"0-1,4-5", []int{0, 1, 4, 5}},
		{"7", []int{7}},
		{"", nil},
	}

	for _, tt := range tests {
		result := parseIntList(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("parseIntList(%q) length = %d, want %d", tt.input, len(result), len(tt.expected))
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("parseIntList(%q)[%d] = %d, want %d", tt.input, i, v, tt.expected[i])
			}
		}
	}
}

func TestCPUScannerAsymmetricL3(t *testing.T) {
	base := t.TempDir()
	cpuBase := filepath.Join(base, "sys", "devices", "system", "cpu")
	os.MkdirAll(cpuBase, 0755)

	// Simulate AMD Ryzen 9950X3D: die 0 has 32M L3, die 1 has 96M L3
	for cpuID := 0; cpuID < 16; cpuID++ {
		cpuDir := filepath.Join(cpuBase, "cpu"+itoa(cpuID))
		topoDir := filepath.Join(cpuDir, "topology")
		os.MkdirAll(topoDir, 0755)

		dieID := cpuID / 8
		coreID := (cpuID % 8) / 2

		os.WriteFile(filepath.Join(topoDir, "die_id"), []byte(itoa(dieID)+"\n"), 0644)
		os.WriteFile(filepath.Join(topoDir, "core_id"), []byte(itoa(coreID)+"\n"), 0644)
		os.WriteFile(filepath.Join(topoDir, "thread_siblings_list"), []byte(itoa(cpuID)+"\n"), 0644)

		if cpuID > 0 {
			os.WriteFile(filepath.Join(cpuDir, "online"), []byte("1\n"), 0644)
		}

		// Create cache for first CPU of each die
		if cpuID%8 == 0 {
			cacheDir := filepath.Join(cpuDir, "cache", "index0")
			os.MkdirAll(cacheDir, 0755)
			os.WriteFile(filepath.Join(cacheDir, "level"), []byte("3\n"), 0644)
			if dieID == 0 {
				os.WriteFile(filepath.Join(cacheDir, "size"), []byte("32M\n"), 0644)
			} else {
				os.WriteFile(filepath.Join(cacheDir, "size"), []byte("96M\n"), 0644)
			}
		}
	}

	scanner := NewCPUScannerWithPath(cpuBase)
	topo, err := scanner.ScanTopology()
	if err != nil {
		t.Fatalf("ScanTopology() returned error: %v", err)
	}

	if len(topo.Dies) != 2 {
		t.Fatalf("Dies count = %d, want 2", len(topo.Dies))
	}

	if topo.Dies[0].L3CacheKB != 32768 {
		t.Errorf("Die[0].L3CacheKB = %d, want 32768 (32M)", topo.Dies[0].L3CacheKB)
	}
	if topo.Dies[1].L3CacheKB != 98304 {
		t.Errorf("Die[1].L3CacheKB = %d, want 98304 (96M)", topo.Dies[1].L3CacheKB)
	}
}

func TestCPUIndexToTopology(t *testing.T) {
	// Build a mock host topology matching the mock sysfs created by createMockSysfs
	hostTopo := models.HostCPUTopology{
		TotalCores:      8,
		TotalCPUs:     16,
		ThreadsPerCore: 2,
		Dies: []models.CPUDie{
			{
				ID:     0,
				Cores:  4,
				Threads: 2,
				CoreDetails: []models.CPUCore{
					{ID: 0, DieID: 0, Threads: []int{0, 1}},
					{ID: 1, DieID: 0, Threads: []int{2, 3}},
					{ID: 2, DieID: 0, Threads: []int{4, 5}},
					{ID: 3, DieID: 0, Threads: []int{6, 7}},
				},
			},
			{
				ID:     1,
				Cores:  4,
				Threads: 2,
				CoreDetails: []models.CPUCore{
					{ID: 0, DieID: 1, Threads: []int{8, 9}},
					{ID: 1, DieID: 1, Threads: []int{10, 11}},
					{ID: 2, DieID: 1, Threads: []int{12, 13}},
					{ID: 3, DieID: 1, Threads: []int{14, 15}},
				},
			},
		},
	}

	tests := []struct {
		name       string
		cpuID      int
		wantDie   int
		wantCore  int
		wantThread int
		wantErr   bool
	}{
		{"CPU 0", 0, 0, 0, 0, false},
		{"CPU 1", 1, 0, 0, 1, false},
		{"CPU 2", 2, 0, 1, 0, false},
		{"CPU 3", 3, 0, 1, 1, false},
		{"CPU 7", 7, 0, 3, 1, false},
		{"CPU 8", 8, 1, 0, 0, false},
		{"CPU 15", 15, 1, 3, 1, false},
		{"invalid CPU", 99, 0, 0, 0, true},
		{"negative CPU", -1, 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dieID, coreID, threadID, err := CPUIndexToTopology(tt.cpuID, hostTopo)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for CPU %d", tt.cpuID)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error for CPU %d: %v", tt.cpuID, err)
			}
			if dieID != tt.wantDie {
				t.Errorf("CPU %d: dieID = %d, want %d", tt.cpuID, dieID, tt.wantDie)
			}
			if coreID != tt.wantCore {
				t.Errorf("CPU %d: coreID = %d, want %d", tt.cpuID, coreID, tt.wantCore)
			}
			if threadID != tt.wantThread {
				t.Errorf("CPU %d: threadID = %d, want %d", tt.cpuID, threadID, tt.wantThread)
			}
		})
	}
}

func TestCPUIndexToTopologyEmpty(t *testing.T) {
	hostTopo := models.HostCPUTopology{}

	_, _, _, err := CPUIndexToTopology(0, hostTopo)
	if err == nil {
		t.Error("Expected error for empty topology")
	}
}
