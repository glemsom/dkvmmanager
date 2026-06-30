// Package vm provides virtual machine management functionality
package vm

import (
	"fmt"
	"sort"

	"github.com/glemsom/dkvmmanager/internal/domain"
)

// vCPUDevice represents a single vCPU device declaration for QEMU.
type vCPUDevice struct {
	apicID    int
	socketID  int
	dieID     int
	guestCore int // Guest-visible core ID (sequential within die)
	threadID  int // 0 or 1
	id        string
}

// GenerateAsymmetricCPUDevices generates -device declarations for asymmetric
// CPU topology. Returns the list of -device args and the APIC ID of the first
// CPU (which is auto-created by -smp 1 and must NOT be declared as -device).
//
// The algorithm:
//  1. Resolves each selected CPU's topology via CPUIndexToTopology
//  2. For each die, collects unique host core IDs from selected CPUs and sorts them
//  3. Builds a mapping: hostCoreID → guestCoreID (sequential within each die, starting at 0)
//  4. For each selected host CPU, computes guestCoreID, threadID, and apicID
//  5. The first CPU (lowest APIC ID, always 0) is returned separately
//  6. Returns -device declarations for all remaining CPUs (skipping the first)
func GenerateAsymmetricCPUDevices(
	selectedCPUs []int,
	hostTopo domain.HostCPUTopology,
) (deviceArgs []string, firstAPICID int, err error) {
	if len(selectedCPUs) == 0 {
		return nil, 0, fmt.Errorf("no CPUs selected for asymmetric topology")
	}

	if len(hostTopo.Dies) == 0 {
		return nil, 0, fmt.Errorf("host topology has no dies")
	}

	// Resolve each selected CPU's topology coordinates using the shared helper.
	type cpuLocation struct {
		dieID    int
		hostCore int
		threadID int
	}
	cpuLocMap := make(map[int]cpuLocation, len(selectedCPUs))
	for _, cpu := range selectedCPUs {
		dieID, coreID, threadID, err := CPUIndexToTopology(cpu, hostTopo)
		if err != nil {
			return nil, 0, fmt.Errorf("selected CPU %d not found in host topology", cpu)
		}
		cpuLocMap[cpu] = cpuLocation{
			dieID:    dieID,
			hostCore: coreID,
			threadID: threadID,
		}
	}

	// Verify that (die=0, core=0, thread=0) is present —
	// this is required because -smp 1 auto-creates it.
	hasDie0Core0Thread0 := false
	for _, cpu := range selectedCPUs {
		loc := cpuLocMap[cpu]
		if loc.dieID == 0 && loc.hostCore == 0 && loc.threadID == 0 {
			hasDie0Core0Thread0 = true
			break
		}
	}
	if !hasDie0Core0Thread0 {
		return nil, 0, fmt.Errorf("asymmetric topology requires selecting the host CPU at (die=0, core=0, thread=0); " +
			"this CPU is auto-created by -smp 1 and must be in the selected set")
	}

	// For each die, collect unique host core IDs from selected CPUs and sort them.
	// This creates the hostCoreID → guestCoreID mapping (sequential among selected).
	type dieCoreInfo struct {
		dieID            int
		selectedHostCores []int // sorted unique host core IDs on this die
	}
	dieCoreMap := make(map[int]*dieCoreInfo) // dieID → info
	for _, cpu := range selectedCPUs {
		loc := cpuLocMap[cpu]
		info, ok := dieCoreMap[loc.dieID]
		if !ok {
			info = &dieCoreInfo{dieID: loc.dieID}
			dieCoreMap[loc.dieID] = info
		}
		info.selectedHostCores = append(info.selectedHostCores, loc.hostCore)
	}

	// Sort and deduplicate host core IDs per die
	for _, info := range dieCoreMap {
		sort.Ints(info.selectedHostCores)
		uniq := info.selectedHostCores[:0]
		for i, c := range info.selectedHostCores {
			if i == 0 || c != info.selectedHostCores[i-1] {
				uniq = append(uniq, c)
			}
		}
		info.selectedHostCores = uniq
	}

	// Compute maxCoresPerDie = maximum core count across all dies
	maxCoresPerDie := 0
	for _, die := range hostTopo.Dies {
		if die.Cores > maxCoresPerDie {
			maxCoresPerDie = die.Cores
		}
	}
	// Also check selected core counts per die (in case host die cores are not set)
	for _, info := range dieCoreMap {
		if len(info.selectedHostCores) > maxCoresPerDie {
			maxCoresPerDie = len(info.selectedHostCores)
		}
	}
	if maxCoresPerDie == 0 {
		maxCoresPerDie = 1
	}

	threadsPerCore := hostTopo.ThreadsPerCore
	if threadsPerCore == 0 {
		threadsPerCore = 1
	}

	numDies := len(hostTopo.Dies)
	if numDies == 0 {
		numDies = 1
	}

	// Build the list of vCPU devices
	var devices []vCPUDevice
	for _, hostCPU := range selectedCPUs {
		loc := cpuLocMap[hostCPU]
		info := dieCoreMap[loc.dieID]

		// Map host core ID to guest core ID (sequential within die)
		guestCoreID := -1
		for i, hc := range info.selectedHostCores {
			if hc == loc.hostCore {
				guestCoreID = i
				break
			}
		}
		if guestCoreID == -1 {
			return nil, 0, fmt.Errorf("internal error: host core %d on die %d not found in mapping", loc.hostCore, loc.dieID)
		}

		// Compute APIC ID using QEMU's formula (now includes dies factor, see issue #65)
		apicID := computeAPICID(0 /* socketID */, numDies, loc.dieID, guestCoreID, loc.threadID, maxCoresPerDie, threadsPerCore)

		devices = append(devices, vCPUDevice{
			apicID:    apicID,
			socketID:  0,
			dieID:     loc.dieID,
			guestCore: guestCoreID,
			threadID:  loc.threadID,
			id:        fmt.Sprintf("cpu-host%d", hostCPU),
		})
	}

	// Sort devices by APIC ID to ensure deterministic ordering
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].apicID < devices[j].apicID
	})

	// The first device (APIC ID 0) is auto-created by -smp 1
	firstAPICID = devices[0].apicID
	if firstAPICID != 0 {
		return nil, 0, fmt.Errorf("expected first APIC ID to be 0, got %d", firstAPICID)
	}

	// Generate -device declarations for all CPUs except the first
	for i := 1; i < len(devices); i++ {
		dev := devices[i]
		devArg := fmt.Sprintf(
			"host-x86_64-cpu,apic-id=%d,socket-id=%d,die-id=%d,core-id=%d,thread-id=%d",
			dev.apicID, dev.socketID, dev.dieID, dev.guestCore, dev.threadID,
		)
		deviceArgs = append(deviceArgs, devArg)
	}

	return deviceArgs, firstAPICID, nil
}

// computeAPICID computes the QEMU APIC ID from topology coordinates.
//
// Formula (x86, QEMU 10.1.5):
//
//	apic-id = (socket-id × dies × cores-per-die × threads) +
//	          (die-id × cores-per-die × threads) +
//	          (guest-core-id × threads) +
//	          thread-id
//
// Where cores-per-die is the maximum cores across all dies.
func computeAPICID(socketID, numDies, dieID, guestCoreID, threadID, maxCoresPerDie, threadsPerCore int) int {
	return (socketID * numDies * maxCoresPerDie * threadsPerCore) +
		(dieID * maxCoresPerDie * threadsPerCore) +
		(guestCoreID * threadsPerCore) +
		threadID
}
