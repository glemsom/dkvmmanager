package vm

import (
	"fmt"
	"sort"

	"github.com/glemsom/dkvmmanager/internal/domain"
)

// ComputePinningFromTopology derives 1:1 host CPU mappings from selected topology.
// The mapping order is topology-aware: die -> core -> thread.
func ComputePinningFromTopology(topo domain.CPUTopology, host domain.HostCPUTopology) (domain.VCPUPinningGlobal, error) {
	result := domain.VCPUPinningGlobal{Enabled: false, Mappings: nil}
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
	for _, die := range host.Dies {
		for _, core := range die.CoreDetails {
			for threadIdx, cpuID := range core.Threads {
				if selected[cpuID] {
					threads = append(threads, hostThread{dieID: die.ID, coreID: core.ID, thread: threadIdx, hostCPU: cpuID})
				}
			}
		}
	}

	if len(threads) != len(topo.SelectedCPUs) {
		return result, fmt.Errorf("selected CPUs do not match detected host topology")
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

	mappings := make([]domain.VCPUToHostMapping, 0, len(threads))
	for i, th := range threads {
		mappings = append(mappings, domain.VCPUToHostMapping{VCPUID: i, HostCPUID: th.hostCPU})
	}

	result.Mappings = mappings
	return result, nil
}