package vm

import (
	"fmt"
	"log"
	"sort"

	"github.com/glemsom/dkvmmanager/internal/domain"
)

// ApplyVCPUPinning applies thread affinity based on QMP CPU topology and configured mapping.
func (r *VMRunner) ApplyVCPUPinning(pinning domain.VCPUPinningGlobal) error {
	if !pinning.Enabled || len(pinning.Mappings) == 0 {
		return nil
	}
	client := r.QMPClient()
	if client == nil {
		return fmt.Errorf("qmp client not connected")
	}
	vcpus, err := client.QueryCPUsFast()
	if err != nil {
		return err
	}
	sort.Slice(vcpus, func(i, j int) bool {
		if vcpus[i].Props.DieID != vcpus[j].Props.DieID {
			return vcpus[i].Props.DieID < vcpus[j].Props.DieID
		}
		if vcpus[i].Props.CoreID != vcpus[j].Props.CoreID {
			return vcpus[i].Props.CoreID < vcpus[j].Props.CoreID
		}
		if vcpus[i].Props.ThreadID != vcpus[j].Props.ThreadID {
			return vcpus[i].Props.ThreadID < vcpus[j].Props.ThreadID
		}
		return vcpus[i].CPU < vcpus[j].CPU
	})

	mappingByVCPU := make(map[int]int, len(pinning.Mappings))
	for _, m := range pinning.Mappings {
		mappingByVCPU[m.VCPUID] = m.HostCPUID
	}

	if debugMode {
		log.Printf("[DEBUG] vCPU pinning: starting for VM %q with %d vCPUs, %d mappings",
			r.vm.Name, len(vcpus), len(pinning.Mappings))
	}

	guestDie := -1
	for idx, v := range vcpus {
		hostCPU, ok := mappingByVCPU[idx]
		if !ok {
			continue
		}
		// Track the first die we see for logging, but don't reject multi-die
		// (asymmetric topology explicitly uses multiple dies)
		if guestDie == -1 {
			guestDie = v.Props.DieID
		}

		if debugMode {
			log.Printf("[DEBUG] vCPU pinning: vCPU %d (ThreadID=%d) -> host CPU %d",
				idx, v.ThreadID, hostCPU)
		}

		if err := PinThreadToCores(v.ThreadID, []int{hostCPU}); err != nil {
			return fmt.Errorf("vcpu %d thread %d pin to host cpu %d: %w", idx, v.ThreadID, hostCPU, err)
		}
	}

	if debugMode {
		log.Printf("[DEBUG] vCPU pinning: completed for VM %q", r.vm.Name)
	}

	return nil
}
