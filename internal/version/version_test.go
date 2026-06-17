package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersionMatchesVERSIONFile(t *testing.T) {
	// The VERSION file at the repo root is the canonical source used by
	// release-please. The Go Version constant must match it.
	repoRoot := findRepoRoot(t)
	versionBytes, err := os.ReadFile(filepath.Join(repoRoot, "VERSION"))
	if err != nil {
		t.Fatalf("cannot read VERSION file: %v", err)
	}

	want := string(versionBytes)
	// Trim trailing newline if present
	if len(want) > 0 && want[len(want)-1] == '\n' {
		want = want[:len(want)-1]
	}

	if Version != want {
		t.Errorf("version.Version = %q, VERSION file = %q — they must agree", Version, want)
	}
}

// findRepoRoot walks up from the test directory looking for the VERSION file.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "VERSION")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (VERSION file not found in any parent)")
		}
		dir = parent
	}
}
