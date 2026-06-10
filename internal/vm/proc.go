// Package vm provides virtual machine management functionality
package vm

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// clkTck is the number of clock ticks per second (USER_HZ).
// On x86 Linux this is always 100, giving 10 ms per jiffy = 10_000_000 ns.
const clkTck = 100
const nsPerJiffy = 1_000_000_000 / clkTck // 10_000_000 ns

// readThreadCPUTime reads the total CPU time (utime + stime) for a thread
// from /proc/<pid>/task/<tid>/stat. Returns the time in nanoseconds.
func readThreadCPUTime(pid, tid int) (int64, error) {
	path := fmt.Sprintf("/proc/%d/task/%d/stat", pid, tid)
	return readStatCPUTime(path)
}

// readProcessCPUTime reads the total CPU time (utime + stime) for a process
// from /proc/<pid>/stat. Returns the time in jiffies (not ns) for delta math
// consistency with how RSS is handled.
func readProcessCPUJiffies(pid int) (uint64, error) {
	path := fmt.Sprintf("/proc/%d/stat", pid)
	return readStatCPUJiffies(path)
}

// readProcessRSS reads VmRSS (resident memory size) for a process from
// /proc/<pid>/status. Returns the size in bytes.
//
// VmRSS appears in /proc/<pid>/status as e.g. "VmRSS:	  12345 kB".
// On error (missing file, malformed line, no VmRSS), returns a wrapped error
// so the caller can decide to log or ignore.
func readProcessRSS(pid int) (uint64, error) {
	path := fmt.Sprintf("/proc/%d/status", pid)
	return readStatusRSS(path)
}

// readStatusRSS parses the VmRSS line from a /proc/<pid>/status file.
// Returns the RSS in bytes (kB value × 1024).
func readStatusRSS(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read %s: %w", path, err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "VmRSS:") {
			continue
		}
		fields := strings.Fields(line)
		// fields[0] = "VmRSS:"
		// fields[1] = numeric value in kB
		// fields[2] = "kB" unit
		if len(fields) < 3 {
			return 0, fmt.Errorf("malformed VmRSS line in %s: %q", path, line)
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse VmRSS value in %s: %w", path, err)
		}
		return kb * 1024, nil
	}

	return 0, fmt.Errorf("VmRSS line not found in %s", path)
}

// readStatCPUTime reads /proc stat file and returns utime+stime in nanoseconds.
func readStatCPUTime(path string) (int64, error) {
	jiffies, err := readStatCPUJiffies(path)
	if err != nil {
		return 0, err
	}
	return int64(jiffies) * nsPerJiffy, nil
}

// readStatCPUJiffies reads /proc/<...>/stat and returns utime+stime in jiffies.
// The stat file format is space-separated, with the comm field in parens
// which may contain spaces. Fields of interest are at positions 14 (utime)
// and 15 (stime) when splitting after the closing paren.
func readStatCPUJiffies(path string) (uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Find the closing paren of the comm field: "... (comm) STATE ..."
	closeParen := strings.LastIndexByte(string(data), ')')
	if closeParen < 0 {
		return 0, fmt.Errorf("malformed stat file %s: no closing paren", path)
	}

	fields := strings.Fields(string(data[closeParen+1:]))
	// fields[0] = state, fields[1..] = remaining
	// utime = fields[11], stime = fields[12] (0-indexed from after state)
	if len(fields) < 13 {
		return 0, fmt.Errorf("malformed stat file %s: only %d fields after comm", path, len(fields))
	}

	utime, err := strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse utime in %s: %w", path, err)
	}
	stime, err := strconv.ParseUint(fields[12], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse stime in %s: %w", path, err)
	}

	return utime + stime, nil
}
