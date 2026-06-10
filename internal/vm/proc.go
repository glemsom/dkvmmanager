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
