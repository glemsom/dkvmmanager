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
