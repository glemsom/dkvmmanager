package vm

import (
	"fmt"
	"sort"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// ComputePinningFromTopology derives 1:1 host CPU mappings from selected topology.
// The mapping order is topology-aware: die -> core -> thread.
func ComputePinningFromTopology(topo models.CPUTopology, host models.HostCPUTopology) (models.VCPUPinningGlobal, error) {
	result := models.VCPUPinningGlobal{Enabled: false, Mappings: nil}
	if !topo.Enabled || len(topo.SelectedCPUs) == 0 {
		return result, nil
	}

	selected := make(map[int]bool, len(topo.SelectedCPUs))
	for _, cpu := range topo.SelectedCPUs {
		selected[cpu] = true
	}

	type hostThread struct {
		dieID   int
		coreID  int
		thread  int
		hostCPU int
	}

	var threads []hostThread
	dieSet := make(map[int]bool)
	for _, die := range host.Dies {
		for _, core := range die.CoreDetails {
			for threadIdx, cpuID := range core.Threads {
				if selected[cpuID] {
					threads = append(threads, hostThread{dieID: die.ID, coreID: core.ID, thread: threadIdx, hostCPU: cpuID})
					dieSet[die.ID] = true
				}
			}
		}
	}

	if len(threads) != len(topo.SelectedCPUs) {
		return result, fmt.Errorf("selected CPUs do not match detected host topology")
	}
	if len(dieSet) > 1 {
		return result, fmt.Errorf("cannot mix dies in one VM topology")
	}

	sort.Slice(threads, func(i, j int) bool {
		if threads[i].dieID != threads[j].dieID {
			return threads[i].dieID < threads[j].dieID
		}
		if threads[i].coreID != threads[j].coreID {
			return threads[i].coreID < threads[j].coreID
		}
		return threads[i].thread < threads[j].thread
	})

	mappings := make([]models.VCPUToHostMapping, 0, len(threads))
	for i, th := range threads {
		mappings = append(mappings, models.VCPUToHostMapping{VCPUID: i, HostCPUID: th.hostCPU})
	}

	result.Mappings = mappings
	return result, nil
}
