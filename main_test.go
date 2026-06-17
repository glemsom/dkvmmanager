package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Issue #66: The tea.LogToFile returned file handle must be saved and closed
// on shutdown to avoid leaking the log file until process exit.

func TestDebugLogFileInitialNil(t *testing.T) {
	// Without calling setupDebugLog, the file handle should be nil.
	// This also verifies closeDebugLog is safe when no file is open.
	if debugLogFile != nil {
		t.Error("expected debugLogFile to be nil before setupDebugLog")
	}
	// Must not panic when file is nil
	closeDebugLog()
}

func TestSetupDebugLogSavesFile(t *testing.T) {
	debugLogFile = nil

	dir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := setupDebugLog(); err != nil {
		t.Fatalf("setupDebugLog() should succeed in temp dir: %v", err)
	}
	// debug.log file should exist
	if _, err := os.Stat(filepath.Join(dir, "debug.log")); err != nil {
		t.Errorf("debug.log not created: %v", err)
	}
	// File handle must not be nil (it was saved, not discarded)
	if debugLogFile == nil {
		t.Fatal("debugLogFile is nil — the file handle was discarded with _")
	}

	// Close and verify it's cleared
	closeDebugLog()
	if debugLogFile != nil {
		t.Error("expected debugLogFile to be nil after closeDebugLog")
	}
}

func TestSetupDebugLogCleansUpOnFailure(t *testing.T) {
	// Verify that if setupDebugLog cannot create a debug log,
	// it returns an error and debugLogFile stays nil.
	// We make CWD non-writable to force at least the first candidate
	// to fail; fallback paths may still succeed in a normal environment.
	// The key assertion is that the function doesn't panic and returns
	// either success (with non-nil file) or error (with nil file).
	debugLogFile = nil

	dir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := os.Chmod(dir, 0444); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0755)

	err = setupDebugLog()
	if err == nil && debugLogFile == nil {
		t.Error("setupDebugLog succeeded but debugLogFile is nil — closer was discarded")
	}
	if err != nil && debugLogFile != nil {
		t.Error("debugLogFile should be nil when setupDebugLog fails")
	}
}

func TestCloseDebugLogIdempotent(t *testing.T) {
	debugLogFile = nil

	dir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origWd)

	if err := setupDebugLog(); err != nil {
		t.Fatal(err)
	}
	// Must not panic on double close
	closeDebugLog()
	closeDebugLog()
}
