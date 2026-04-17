package vm

import (
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

func TestParseCPUList(t *testing.T) {
	cpus, err := ParseCPUList("0,1,2-4,4,6")
	if err != nil {
		t.Fatalf("ParseCPUList error: %v", err)
	}
	expected := []int{0, 1, 2, 3, 4, 6}
	if len(cpus) != len(expected) {
		t.Fatalf("len=%d want=%d", len(cpus), len(expected))
	}
	for i := range expected {
		if cpus[i] != expected[i] {
			t.Fatalf("cpus[%d]=%d want=%d", i, cpus[i], expected[i])
		}
	}
}

func TestParseCPUListInvalid(t *testing.T) {
	if _, err := ParseCPUList("3-1"); err == nil {
		t.Fatal("expected error for reversed range")
	}
	if _, err := ParseCPUList("a"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestComputePinningFromTopology(t *testing.T) {
	host := testHostTopoSingleDie()
	topo := models.CPUTopology{Enabled: true, SelectedCPUs: []int{4, 5, 6, 7}}
	pin, err := ComputePinningFromTopology(topo, host)
	if err != nil {
		t.Fatalf("ComputePinningFromTopology: %v", err)
	}
	if len(pin.Mappings) != 4 {
		t.Fatalf("mappings=%d want=4", len(pin.Mappings))
	}
	want := []int{4, 5, 6, 7}
	for i, m := range pin.Mappings {
		if m.VCPUID != i || m.HostCPUID != want[i] {
			t.Fatalf("mapping[%d]=(%d->%d) want (%d->%d)", i, m.VCPUID, m.HostCPUID, i, want[i])
		}
	}
}

func TestComputePinningFromTopologyMixedDies(t *testing.T) {
	host := testHostTopoTwoDies()
	topo := models.CPUTopology{Enabled: true, SelectedCPUs: []int{0, 1, 8, 9}}
	pin, err := ComputePinningFromTopology(topo, host)
	if err != nil {
		t.Fatalf("ComputePinningFromTopology: %v", err)
	}
	if len(pin.Mappings) != 4 {
		t.Fatalf("mappings=%d want=4", len(pin.Mappings))
	}
	// Verify sorted by die -> core -> thread
	// Die 0, Core 0, Threads 0,1 then Die 1, Core 4, Threads 8,9
	wantHostCPUs := []int{0, 1, 8, 9}
	for i, m := range pin.Mappings {
		if m.VCPUID != i {
			t.Fatalf("mapping[%d].VCPUID=%d want=%d", i, m.VCPUID, i)
		}
		if m.HostCPUID != wantHostCPUs[i] {
			t.Fatalf("mapping[%d].HostCPUID=%d want=%d", i, m.HostCPUID, wantHostCPUs[i])
		}
	}
}

func testHostTopoSingleDie() models.HostCPUTopology {
	return models.HostCPUTopology{Dies: []models.CPUDie{{ID: 0, CoreDetails: []models.CPUCore{{ID: 2, DieID: 0, Threads: []int{4, 5}}, {ID: 3, DieID: 0, Threads: []int{6, 7}}}}}}
}

func testHostTopoTwoDies() models.HostCPUTopology {
	return models.HostCPUTopology{Dies: []models.CPUDie{
		{ID: 0, CoreDetails: []models.CPUCore{{ID: 0, DieID: 0, Threads: []int{0, 1}}}},
		{ID: 1, CoreDetails: []models.CPUCore{{ID: 4, DieID: 1, Threads: []int{8, 9}}}},
	}}
}
