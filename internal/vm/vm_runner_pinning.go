package vm

import (
	"fmt"
	"sort"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// ApplyVCPUPinning applies thread affinity based on QMP CPU topology and configured mapping.
func (r *VMRunner) ApplyVCPUPinning(pinning models.VCPUPinningGlobal) error {
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

	guestDie := -1
	for idx, v := range vcpus {
		hostCPU, ok := mappingByVCPU[idx]
		if !ok {
			continue
		}
		if guestDie == -1 {
			guestDie = v.Props.DieID
		} else if guestDie != v.Props.DieID {
			return fmt.Errorf("guest vcpu dies mixed in running VM: %d and %d", guestDie, v.Props.DieID)
		}
		if err := PinThreadToCores(v.ThreadID, []int{hostCPU}); err != nil {
			return fmt.Errorf("vcpu %d thread %d pin to host cpu %d: %w", idx, v.ThreadID, hostCPU, err)
		}
	}
	return nil
}
