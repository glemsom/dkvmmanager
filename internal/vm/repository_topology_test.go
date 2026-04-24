package vm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// mockRepositoryForTopology creates a repository with a temporary config file
func mockRepositoryForTopology(t *testing.T) (*Repository, func()) {
	t.Helper()
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Create initial config file
	if err := os.WriteFile(configPath, []byte("vms: {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create repository using the public constructor
	repo, err := NewRepository(configPath)
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return repo, cleanup
}

func TestGetCPUTopology(t *testing.T) {
	repo, cleanup := mockRepositoryForTopology(t)
	defer cleanup()

	// Set CPU topology config directly using viper
	repo.vip.Set("cpu_topology", map[string]interface{}{
		"enabled":          true,
		"selected_cpus":   []interface{}{4, 5, 6, 7},
		"use_host_topology": true,
	})

	topo, err := repo.GetCPUTopology()
	if err != nil {
		t.Fatalf("GetCPUTopology() returned error: %v", err)
	}

	if !topo.Enabled {
		t.Error("Expected Enabled=true")
	}
	if len(topo.SelectedCPUs) != 4 {
		t.Errorf("SelectedCPUs length = %d, want 4", len(topo.SelectedCPUs))
	}
	if topo.SelectedCPUs[0] != 4 {
		t.Errorf("SelectedCPUs[0] = %d, want 4", topo.SelectedCPUs[0])
	}
	if !topo.UseHostTopology {
		t.Error("Expected UseHostTopology=true")
	}
}

func TestSaveCPUTopology(t *testing.T) {
	repo, cleanup := mockRepositoryForTopology(t)
	defer cleanup()

	topo := models.CPUTopology{
		Enabled:         true,
		SelectedCPUs:    []int{0, 1, 2, 3},
		UseHostTopology: true,
	}

	if err := repo.SaveCPUTopology(topo); err != nil {
		t.Fatalf("SaveCPUTopology() returned error: %v", err)
	}

	// Reload config to verify persistence
	if err := repo.vip.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	loaded, err := repo.GetCPUTopology()
	if err != nil {
		t.Fatalf("GetCPUTopology() returned error after reload: %v", err)
	}

	if !loaded.Enabled {
		t.Error("Loaded Enabled should be true")
	}
	if len(loaded.SelectedCPUs) != 4 {
		t.Errorf("Loaded SelectedCPUs length = %d, want 4", len(loaded.SelectedCPUs))
	}
	if loaded.SelectedCPUs[0] != 0 || loaded.SelectedCPUs[3] != 3 {
		t.Errorf("Loaded SelectedCPUs = %v, want [0,1,2,3]", loaded.SelectedCPUs)
	}
	if !loaded.UseHostTopology {
		t.Error("Loaded UseHostTopology should be true")
	}
}

func TestSaveCPUTopologyUseHostTopologyFalse(t *testing.T) {
	repo, cleanup := mockRepositoryForTopology(t)
	defer cleanup()

	topo := models.CPUTopology{
		Enabled:         true,
		SelectedCPUs:    []int{0, 2, 4, 6},
		UseHostTopology: false,
	}

	if err := repo.SaveCPUTopology(topo); err != nil {
		t.Fatalf("SaveCPUTopology() returned error: %v", err)
	}

	loaded, err := repo.GetCPUTopology()
	if err != nil {
		t.Fatalf("GetCPUTopology() returned error: %v", err)
	}

	if loaded.UseHostTopology {
		t.Error("Loaded UseHostTopology should be false")
	}
}

func TestSaveCPUTopologyRoundTripMultiple(t *testing.T) {
	repo, cleanup := mockRepositoryForTopology(t)
	defer cleanup()

	// Test multiple round-trips
	testCases := []models.CPUTopology{
		{Enabled: true, SelectedCPUs: []int{0}, UseHostTopology: true},
		{Enabled: false, SelectedCPUs: []int{1, 2, 3, 4}, UseHostTopology: false},
		{Enabled: true, SelectedCPUs: []int{0, 1, 2, 3, 4, 5, 6, 7}, UseHostTopology: true},
	}

	for i, topo := range testCases {
		if err := repo.SaveCPUTopology(topo); err != nil {
			t.Fatalf("SaveCPUTopology() test case %d returned error: %v", i, err)
		}

		loaded, err := repo.GetCPUTopology()
		if err != nil {
			t.Fatalf("GetCPUTopology() test case %d returned error: %v", i, err)
		}

		if loaded.Enabled != topo.Enabled {
			t.Errorf("Case %d: Enabled = %v, want %v", i, loaded.Enabled, topo.Enabled)
		}
		if len(loaded.SelectedCPUs) != len(topo.SelectedCPUs) {
			t.Errorf("Case %d: SelectedCPUs length = %d, want %d", i, len(loaded.SelectedCPUs), len(topo.SelectedCPUs))
		}
		if loaded.UseHostTopology != topo.UseHostTopology {
			t.Errorf("Case %d: UseHostTopology = %v, want %v", i, loaded.UseHostTopology, topo.UseHostTopology)
		}
	}
}