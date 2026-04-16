// Package hugepages provides utilities for managing hugepages allocation.
// This is used for VM memory backing with 2MB hugepages.
package hugepages

import (
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
)

// Config holds hugepages configuration
type Config struct {
	MemMB   int64 // Memory in MB (default: 8192)
	PageSize int64 // Hugepage size (default: 2MB)
}

// NewConfig creates a config for default 8GB hugepages
func NewConfig() *Config {
	return &Config{
		MemMB:    DefaultMemoryMB,
		PageSize: HugepageSize2MB,
	}
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