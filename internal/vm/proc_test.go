package vm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadStatCPUJiffies(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stat")

	// Write a simulated /proc/<pid>/task/<tid>/stat file
	// Format: PID (comm) STATE ppid pgrp session tty_nr tpgid flags minflt cminflt
	//         majflt cmajflt UTIME STIME cutime cstime ...
	data := "12345 (qemu-system-x86) S 1 12345 12345 0 -1 4194560 0 0 0 0 1500 2500 0 0 20 0 8 0 123456789 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	jiffies, err := readStatCPUJiffies(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// utime=1500, stime=2500 -> total=4000
	if jiffies != 4000 {
		t.Errorf("expected 4000 jiffies, got %d", jiffies)
	}
}

func TestReadStatCPUJiffiesRealProc(t *testing.T) {
	// Use the test process's own /proc/self/stat which always exists on Linux.
	// Just verify parsing succeeds (no error) — zero CPU time is valid for
	// short-lived processes.
	_, err := readStatCPUJiffies("/proc/self/stat")
	if err != nil {
		t.Fatalf("failed to read /proc/self/stat: %v", err)
	}
}

func TestReadStatCPUJiffiesMissingFile(t *testing.T) {
	_, err := readStatCPUJiffies("/proc/nonexistent/stat")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestReadStatCPUJiffiesMalformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stat")

	// Missing closing paren
	data := "12345 (qemu-system-x86 S 1\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := readStatCPUJiffies(path)
	if err == nil {
		t.Error("expected error for malformed stat file")
	}
}

func TestNSecPerJiffy(t *testing.T) {
	expected := int64(10_000_000) // 10ms per jiffy at CLK_TCK=100
	if nsPerJiffy != expected {
		t.Errorf("expected nsPerJiffy=%d, got %d", expected, nsPerJiffy)
	}
}

func TestReadProcessRSS(t *testing.T) {
	// Use the test process's own /proc/self/status which always exists on Linux.
	// Verify parsing succeeds and returns a value (RSS for the test process).
	rss, err := readProcessRSS(os.Getpid())
	if err != nil {
		t.Fatalf("readProcessRSS(/proc/self) failed: %v", err)
	}
	if rss == 0 {
		t.Error("expected non-zero RSS for the test process")
	}
	// RSS should be a multiple of 1024 (kB-aligned) and at least one page (4 KiB)
	if rss%1024 != 0 {
		t.Errorf("expected RSS to be kB-aligned (multiple of 1024), got %d", rss)
	}
	if rss < 4096 {
		t.Errorf("expected RSS >= 4 KiB, got %d", rss)
	}
}

func TestReadStatusRSS(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "status")

	// Realistic /proc/<pid>/status content
	data := `Name:	qemu-system-x86
Umask:	0022
State:	S (sleeping)
VmRSS:	  12345 kB
VmSize:	 65536 kB
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	rss, err := readStatusRSS(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 12345 kB = 12345 * 1024 bytes
	expected := uint64(12345) * 1024
	if rss != expected {
		t.Errorf("expected %d bytes, got %d", expected, rss)
	}
}

func TestReadStatusRSSMissingVmRSS(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "status")

	// Status file without VmRSS line
	data := `Name:	qemu-system-x86
State:	S (sleeping)
`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := readStatusRSS(path)
	if err == nil {
		t.Error("expected error when VmRSS line is missing")
	}
}

func TestReadStatusRSSMalformed(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "status")

	// Malformed VmRSS line (missing kB suffix)
	data := "VmRSS:	12345\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := readStatusRSS(path)
	if err == nil {
		t.Error("expected error for malformed VmRSS line")
	}
}

func TestReadProcessRSSMissingPID(t *testing.T) {
	// A PID that is very unlikely to exist. Should return a clean error
	// (os.ErrNotExist wrapped), not a panic.
	_, err := readProcessRSS(999999999)
	if err == nil {
		t.Error("expected error for missing PID, got nil")
	}
	// The error should mention the failure (so the caller can decide
	// to log or ignore), but it must not be a panic.
}

func TestReadProcessCPUJiffiesRealProc(t *testing.T) {
	// Use the test process's own /proc/self/stat. Wrapper test ensures
	// the readProcessCPUJiffies path (not just readStatCPUJiffies) works.
	jiffies, err := readProcessCPUJiffies(os.Getpid())
	if err != nil {
		t.Fatalf("readProcessCPUJiffies failed: %v", err)
	}
	// Jiffies can be 0 for very short-lived processes — just verify the
	// call returned without error and a non-negative value.
	_ = jiffies
}

func TestReadProcessCPUJiffiesMissingPID(t *testing.T) {
	// Missing PID must return a clean error (no panic).
	_, err := readProcessCPUJiffies(999999999)
	if err == nil {
		t.Error("expected error for missing PID, got nil")
	}
}
