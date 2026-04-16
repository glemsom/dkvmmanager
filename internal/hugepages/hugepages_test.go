package hugepages

import (
	"os"
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