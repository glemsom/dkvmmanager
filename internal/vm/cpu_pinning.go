package vm

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

func PinThreadToCores(pid int, cores []int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid pid: %d", pid)
	}
	if len(cores) == 0 {
		return fmt.Errorf("no cores specified")
	}

	// Debug: capture affinity BEFORE pinning
	if debugMode {
		before, err := GetThreadAffinity(pid)
		if err != nil {
			log.Printf("[DEBUG] vCPU pinning: could not get current affinity for PID %d: %v", pid, err)
		} else {
			log.Printf("[DEBUG] vCPU pinning: BEFORE - PID %d, current affinity: %v", pid, before)
		}
	}

	var set unix.CPUSet
	set.Zero()
	for _, core := range cores {
		if core < 0 {
			return fmt.Errorf("invalid core: %d", core)
		}
		set.Set(core)
	}
	if err := unix.SchedSetaffinity(pid, &set); err != nil {
		return fmt.Errorf("sched_setaffinity(%d): %w", pid, err)
	}

	// Debug: capture affinity AFTER pinning
	if debugMode {
		after, err := GetThreadAffinity(pid)
		if err != nil {
			log.Printf("[DEBUG] vCPU pinning: could not verify new affinity for PID %d: %v", pid, err)
		} else {
			log.Printf("[DEBUG] vCPU pinning: AFTER - PID %d, new affinity: %v", pid, after)
		}
	}

	return nil
}

func GetThreadAffinity(pid int) ([]int, error) {
	var set unix.CPUSet
	set.Zero()
	if err := unix.SchedGetaffinity(pid, &set); err != nil {
		return nil, fmt.Errorf("sched_getaffinity(%d): %w", pid, err)
	}
	var cores []int
	for i := 0; i < 1024; i++ {
		if set.IsSet(i) {
			cores = append(cores, i)
		}
	}
	return cores, nil
}

func ParseCPUList(cpuStr string) ([]int, error) {
	cpuStr = strings.TrimSpace(cpuStr)
	if cpuStr == "" {
		return nil, nil
	}
	seen := map[int]bool{}
	for _, part := range strings.Split(cpuStr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			bounds := strings.SplitN(part, "-", 2)
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid range start: %s", part)
			}
			end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid range end: %s", part)
			}
			if start > end || start < 0 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			for i := start; i <= end; i++ {
				seen[i] = true
			}
			continue
		}
		v, err := strconv.Atoi(part)
		if err != nil || v < 0 {
			return nil, fmt.Errorf("invalid cpu: %s", part)
		}
		seen[v] = true
	}
	out := make([]int, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	sort.Ints(out)
	return out, nil
}
