package hugepages

import (
	"os"
	"strings"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg.MemMB != DefaultMemoryMB {
		t.Errorf("Expected MemMB=%d, got %d", DefaultMemoryMB, cfg.MemMB)
	}

	if cfg.PageSize != HugepageSize2MB {
		t.Errorf("Expected PageSize=%d, got %d", HugepageSize2MB, cfg.PageSize)
	}
}

func TestRequiredPages(t *testing.T) {
	tests := []struct {
		name     string
		memMB    int64
		pageSize int64
		expected int64
	}{
		{
			name:     "8GB with 2MB pages",
			memMB:    8192,
			pageSize: 2 * 1024 * 1024,
			expected: 4096,
		},
		{
			name:     "4GB with 2MB pages",
			memMB:    4096,
			pageSize: 2 * 1024 * 1024,
			expected: 2048,
		},
		{
			name:     "16GB with 2MB pages",
			memMB:    16384,
			pageSize: 2 * 1024 * 1024,
			expected: 8192,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				MemMB:    tt.memMB,
				PageSize: tt.pageSize,
			}

			result := cfg.RequiredPages()
			if result != tt.expected {
				t.Errorf("Expected %d pages, got %d", tt.expected, result)
			}
		})
	}
}

func TestCheckReadsFromProc(t *testing.T) {
	// This test verifies Check() can read from /proc/sys/vm/nr_hugepages
	// It may fail in environments without the file, which is expected
	result, err := Check()

	// If the file doesn't exist (e.g., Windows or some containers), skip
	if os.IsNotExist(err) {
		t.Skip("nr_hugepages not available on this system")
	}

	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	// AvailablePages should be non-negative
	if result.AvailablePages < 0 {
		t.Errorf("AvailablePages should be non-negative, got %d", result.AvailablePages)
	}
}

func TestCheckResultIsSufficient(t *testing.T) {
	result := &CheckResult{
		AvailablePages: 5000,
		RequiredPages:  4096,
	}
	result.IsSufficient = result.AvailablePages >= result.RequiredPages

	if !result.IsSufficient {
		t.Error("Expected IsSufficient=true when available >= required")
	}

	result.AvailablePages = 2000
	result.IsSufficient = result.AvailablePages >= result.RequiredPages

	if result.IsSufficient {
		t.Error("Expected IsSufficient=false when available < required")
	}
}

func TestFormatError(t *testing.T) {
	result := &CheckResult{
		AvailablePages: 1000,
		RequiredPages:  4096,
	}

	errMsg := FormatError(result)

	expected := "insufficient hugepages: have 1000, need 4096 (try: echo 4096 > /proc/sys/vm/nr_hugepages)"
	if errMsg != expected {
		t.Errorf("Expected error message %q, got %q", expected, errMsg)
	}
}

func TestGetTotalSystemMemoryMB(t *testing.T) {
	totalMB, err := GetTotalSystemMemoryMB()
	if err != nil {
		// Skip if /proc/meminfo not available (e.g., non-Linux)
		if os.IsNotExist(err) {
			t.Skip("/proc/meminfo not available")
		}
		t.Fatalf("GetTotalSystemMemoryMB failed: %v", err)
	}

	if totalMB <= 0 {
		t.Errorf("Expected positive total memory, got %d", totalMB)
	}

	// Sanity check: total memory should be at least a few GB on modern systems
	// Allow small embedded systems but warn
	if totalMB < 1024 {
		t.Logf("Warning: total memory (%d MB) seems low", totalMB)
	}
}

func TestNewAutoConfig(t *testing.T) {
	cfg, err := NewAutoConfig()
	if err != nil {
		// Skip if cannot read memory (non-Linux or no /proc/meminfo)
		if os.IsNotExist(err) || strings.Contains(err.Error(), "failed to read /proc/meminfo") {
			t.Skip("Cannot read /proc/meminfo: ", err)
		}
		t.Fatalf("NewAutoConfig failed: %v", err)
	}

	// Verify memory is reasonable
	if cfg.MemMB <= 0 {
		t.Errorf("Expected positive MemMB, got %d", cfg.MemMB)
	}

	// Verify page size is 2MB
	if cfg.PageSize != HugepageSize2MB {
		t.Errorf("Expected PageSize=%d, got %d", HugepageSize2MB, cfg.PageSize)
	}

	// Verify memory is aligned to 2MB boundary
	if cfg.MemMB%2 != 0 {
		t.Errorf("MemMB should be aligned to 2MB boundary, got %d", cfg.MemMB)
	}

	// Get total memory to verify reserve logic
	totalMB, err := GetTotalSystemMemoryMB()
	if err != nil {
		t.Skip("Cannot verify reserve: cannot read total memory")
	}

	// VM memory should be <= total - 4GB (with alignment)
	expectedMax := (totalMB - ReservedOSMemoryMB) / 2 * 2
	if cfg.MemMB > expectedMax {
		t.Errorf("MemMB (%d) exceeds expected max (%d) after reserving %d MB", cfg.MemMB, expectedMax, ReservedOSMemoryMB)
	}
}
