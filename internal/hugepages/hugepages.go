// Package hugepages provides utilities for managing hugepages allocation.
// This is used for VM memory backing with 2MB hugepages.
package hugepages

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// HugepageSize2MB is the default hugepage size: 2MB
	HugepageSize2MB = 2 * 1024 * 1024

	// DefaultMemoryMB is the default memory size: 8GB
	DefaultMemoryMB = 8192

	// ReservedOSMemoryMB is memory reserved for the host OS: 4GB
	ReservedOSMemoryMB = 4096
)

// Config holds hugepages configuration
type Config struct {
	MemMB   int64 // Memory in MB (default: 8192)
	PageSize int64 // Hugepage size (default: 2MB)
}

// NewConfig creates a default hugepages configuration.
func NewConfig() *Config {
	return &Config{
		MemMB:    DefaultMemoryMB,
		PageSize: HugepageSize2MB,
	}
}

// GetTotalSystemMemoryMB reads /proc/meminfo and returns the total system memory in MB.
func GetTotalSystemMemoryMB() (int64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, fmt.Errorf("failed to read /proc/meminfo: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			// Extract the value in kB (MemTotal is in kilobytes)
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			kb, err := strconv.ParseInt(fields[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse MemTotal: %w", err)
			}
			// Convert kB to MB
			return kb / 1024, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error scanning meminfo: %w", err)
	}

	return 0, errors.New("MemTotal not found in /proc/meminfo")
}

// NewAutoConfig creates a config with automatically detected memory.
// It reads total system memory from /proc/meminfo, subtracts ReservedOSMemoryMB (4GB),
// and aligns the result down to the nearest 2MB boundary.
func NewAutoConfig() (*Config, error) {
	totalMB, err := GetTotalSystemMemoryMB()
	if err != nil {
		return nil, fmt.Errorf("failed to get total system memory: %w", err)
	}

	// Reserve 4GB for OS
	availableMB := totalMB - ReservedOSMemoryMB
	if availableMB <= 0 {
		return nil, fmt.Errorf("insufficient memory after reserving %d MB for OS (total: %d MB)", ReservedOSMemoryMB, totalMB)
	}

	// Align down to 2MB boundary (hugepage size)
	// Since hugepages are 2MB, ensure memory size is a multiple of 2MB
	alignedMB := (availableMB / 2) * 2

	// Ensure at least 1 hugepage (2MB) is available
	if alignedMB < 2 {
		return nil, fmt.Errorf("not enough memory for VM after alignment: %d MB", alignedMB)
	}

	return &Config{
		MemMB:    alignedMB,
		PageSize: HugepageSize2MB,
	}, nil
}

// RequiredPages returns the number of hugepages needed for the configured memory
func (c *Config) RequiredPages() int64 {
	return (c.MemMB * 1024 * 1024) / c.PageSize
}

// CheckResult holds the result of a hugepages check
type CheckResult struct {
	AvailablePages int64
	RequiredPages  int64
	IsSufficient   bool
}

// Check reads the current hugepages availability from /proc/sys/vm/nr_hugepages
func Check() (*CheckResult, error) {
	data, err := os.ReadFile("/proc/sys/vm/nr_hugepages")
	if err != nil {
		return nil, fmt.Errorf("failed to read hugepages: %w", err)
	}

	available, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hugepages: %w", err)
	}

	return &CheckResult{
		AvailablePages: available,
		RequiredPages:  0, // Set by caller
		IsSufficient:   false,
	}, nil
}

// ErrInsufficientHugepages is returned when there aren't enough hugepages available
var ErrInsufficientHugepages = errors.New("insufficient hugepages")

// Ensure attempts to allocate required hugepages by writing to nr_hugepages.
// Returns a CheckResult with the final counts after the attempt.
// Note: This may fail if the system doesn't have enough free memory,
// or if not running as root.
func Ensure(cfg *Config) (*CheckResult, error) {
	// First check current availability
	result, err := Check()
	if err != nil {
		return nil, fmt.Errorf("failed to check hugepages: %w", err)
	}

	required := cfg.RequiredPages()
	result.RequiredPages = required

	// Already sufficient?
	if result.AvailablePages >= required {
		result.IsSufficient = true
		return result, nil
	}

	// Try to allocate more
	if err := os.WriteFile("/proc/sys/vm/nr_hugepages", []byte(fmt.Sprintf("%d", required)), 0644); err != nil {
		// Return current state - may still have some hugepages
		return result, fmt.Errorf("failed to allocate hugepages (may require root): %w", err)
	}

	// Re-check after allocation attempt
	result, err = Check()
	if err != nil {
		return nil, fmt.Errorf("failed to re-check hugepages: %w", err)
	}

	result.RequiredPages = required
	result.IsSufficient = result.AvailablePages >= required

	return result, nil
}

// FormatError creates a user-friendly error message for insufficient hugepages
func FormatError(result *CheckResult) string {
	return fmt.Sprintf(
		"insufficient hugepages: have %d, need %d (try: echo %d > /proc/sys/vm/nr_hugepages)",
		result.AvailablePages,
		result.RequiredPages,
		result.RequiredPages,
	)
}