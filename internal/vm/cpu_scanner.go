// Package vm provides virtual machine management functionality
package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/glemsom/dkvmmanager/internal/models"
)

const (
	// SysfsCPUBasePath is the base path for CPU entries in sysfs
	SysfsCPUBasePath = "/sys/devices/system/cpu"
)

// CPUScanner scans the host for CPU topology information
type CPUScanner struct {
	sysfsPath string
}

// NewCPUScanner creates a new CPU scanner using default paths
func NewCPUScanner() *CPUScanner {
	return &CPUScanner{
		sysfsPath: SysfsCPUBasePath,
	}
}

// NewCPUScannerWithPath creates a new CPU scanner with a custom sysfs path (for testing)
func NewCPUScannerWithPath(sysfsPath string) *CPUScanner {
	return &CPUScanner{
		sysfsPath: sysfsPath,
	}
}

// ScanTopology detects the host CPU topology from sysfs
func (s *CPUScanner) ScanTopology() (models.HostCPUTopology, error) {
	cpuDirs, err := s.findOnlineCPUDirs()
	if err != nil {
		return models.HostCPUTopology{}, fmt.Errorf("failed to find CPU directories: %w", err)
	}

	if len(cpuDirs) == 0 {
		return models.HostCPUTopology{}, fmt.Errorf("no CPU entries found in %s", s.sysfsPath)
	}

	// Collect per-CPU topology info
	type cpuInfo struct {
		id     int
		coreID int
		dieID  int
		thread bool // true if this is a sibling (not the first thread)
	}
	var cpus []cpuInfo

	// Track core_id -> first logical CPU to determine thread siblings
	coreFirstCPU := make(map[string]int) // key: "dieID:coreID"
	threadsPerCore := 0

	for _, dir := range cpuDirs {
		cpuID, err := s.readCPUID(dir)
		if err != nil {
			continue
		}

		coreID := s.readSysfsInt(filepath.Join(dir, "topology", "core_id"))
		dieID := s.readSysfsInt(filepath.Join(dir, "topology", "die_id"))

		// Read thread siblings to determine threads per core
		siblings := s.readThreadSiblings(dir)
		key := fmt.Sprintf("%d:%d", dieID, coreID)

		isThread := false
		if first, ok := coreFirstCPU[key]; ok {
			isThread = cpuID != first
			if threadsPerCore == 0 {
				threadsPerCore = len(siblings)
			}
		} else {
			coreFirstCPU[key] = cpuID
			if threadsPerCore == 0 && len(siblings) > 0 {
				threadsPerCore = len(siblings)
			}
		}

		cpus = append(cpus, cpuInfo{
			id:     cpuID,
			coreID: coreID,
			dieID:  dieID,
			thread: isThread,
		})
	}

	if threadsPerCore == 0 {
		threadsPerCore = 1
	}

	// Group by die
	dieMap := make(map[int]*models.CPUDie)
	coreThreads := make(map[string][]int) // key: "dieID:coreID" -> thread IDs

	for _, cpu := range cpus {
		die, ok := dieMap[cpu.dieID]
		if !ok {
			die = &models.CPUDie{
				ID:        cpu.dieID,
				Threads:   threadsPerCore,
				L3CacheKB: s.readL3CacheKB(cpu.dieID),
			}
			dieMap[cpu.dieID] = die
		}

		die.LogicalCPUs = append(die.LogicalCPUs, cpu.id)

		// Track threads per physical core
		key := fmt.Sprintf("%d:%d", cpu.dieID, cpu.coreID)
		coreThreads[key] = append(coreThreads[key], cpu.id)

		// Count physical cores (only first thread per core)
		if !cpu.thread {
			die.Cores++
		}
	}

	// Build core details for each die
	for _, die := range dieMap {
		seenCores := make(map[int]bool)
		for _, cpuID := range die.LogicalCPUs {
			// Find core ID for this CPU
			for _, cpu := range cpus {
				if cpu.id == cpuID && !seenCores[cpu.coreID] {
					seenCores[cpu.coreID] = true
					key := fmt.Sprintf("%d:%d", die.ID, cpu.coreID)
					threads := coreThreads[key]
					sort.Ints(threads)
					die.CoreDetails = append(die.CoreDetails, models.CPUCore{
						ID:      cpu.coreID,
						Threads: threads,
						DieID:   die.ID,
					})
					break
				}
			}
		}
		// Sort core details by core ID
		sort.Slice(die.CoreDetails, func(i, j int) bool {
			return die.CoreDetails[i].ID < die.CoreDetails[j].ID
		})
	}

	// Convert map to sorted slice
	dies := make([]models.CPUDie, 0, len(dieMap))
	var dieIDs []int
	for id := range dieMap {
		dieIDs = append(dieIDs, id)
	}
	sort.Ints(dieIDs)
	for _, id := range dieIDs {
		// Sort logical CPUs within each die
		sort.Ints(dieMap[id].LogicalCPUs)
		dies = append(dies, *dieMap[id])
	}

	// Compute totals
	totalCores := 0
	totalCPUs := 0
	for _, die := range dies {
		totalCores += die.Cores
		totalCPUs += len(die.LogicalCPUs)
	}

	return models.HostCPUTopology{
		Dies:           dies,
		TotalCores:     totalCores,
		TotalCPUs:      totalCPUs,
		ThreadsPerCore: threadsPerCore,
	}, nil
}

// findOnlineCPUDirs returns sysfs directories for online CPUs (cpu0, cpu1, ...)
func (s *CPUScanner) findOnlineCPUDirs() ([]string, error) {
	entries, err := os.ReadDir(s.sysfsPath)
	if err != nil {
		return nil, err
	}

	type cpuDir struct {
		path string
		id   int
	}
	var dirs []cpuDir
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "cpu") {
			continue
		}
		// Skip "cpuidle", "cpufreq", "cpufreq", etc.
		suffix := entry.Name()[3:]
		id, err := strconv.Atoi(suffix)
		if err != nil {
			continue
		}
		// Check if CPU is online
		onlinePath := filepath.Join(s.sysfsPath, entry.Name(), "online")
		if data, err := os.ReadFile(onlinePath); err == nil {
			if strings.TrimSpace(string(data)) != "1" {
				continue
			}
		}
		// cpu0 may not have online file (always online)
		dirs = append(dirs, cpuDir{path: filepath.Join(s.sysfsPath, entry.Name()), id: id})
	}

	// Sort by numeric CPU ID
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].id < dirs[j].id
	})

	result := make([]string, len(dirs))
	for i, d := range dirs {
		result[i] = d.path
	}
	return result, nil
}

// readCPUID extracts the numeric CPU ID from a directory path like /sys/.../cpu0
func (s *CPUScanner) readCPUID(dir string) (int, error) {
	baseName := filepath.Base(dir)
	suffix := baseName[3:] // strip "cpu"
	return strconv.Atoi(suffix)
}

// readSysfsInt reads an integer value from a sysfs file
func (s *CPUScanner) readSysfsInt(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	val, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return val
}

// readThreadSiblings reads the thread_siblings_list from topology
func (s *CPUScanner) readThreadSiblings(cpuDir string) []int {
	path := filepath.Join(cpuDir, "topology", "thread_siblings_list")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return parseIntList(strings.TrimSpace(string(data)))
}

// readL3CacheKB reads the L3 cache size for a given die from sysfs
func (s *CPUScanner) readL3CacheKB(dieID int) int {
	// L3 cache is shared per die. Find it by scanning cache/index* for type "Unified"
	// We look at cpu0's cache dirs (or the first cpu of the die)
	cpuDirs, err := s.findOnlineCPUDirs()
	if err != nil || len(cpuDirs) == 0 {
		return 0
	}

	// Find a CPU in this die
	targetDir := ""
	for _, dir := range cpuDirs {
		die := s.readSysfsInt(filepath.Join(dir, "topology", "die_id"))
		if die == dieID {
			targetDir = dir
			break
		}
	}
	if targetDir == "" {
		targetDir = cpuDirs[0]
	}

	cacheDir := filepath.Join(targetDir, "cache")
	cacheEntries, err := os.ReadDir(cacheDir)
	if err != nil {
		return 0
	}

	for _, entry := range cacheEntries {
		if !entry.IsDir() {
			continue
		}
		indexPath := filepath.Join(cacheDir, entry.Name())

		// Check if this is L3 cache
		typeData, err := os.ReadFile(filepath.Join(indexPath, "level"))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(typeData)) != "3" {
			continue
		}

		// Read size
		sizeData, err := os.ReadFile(filepath.Join(indexPath, "size"))
		if err != nil {
			continue
		}
		return parseCacheSizeKB(strings.TrimSpace(string(sizeData)))
	}

	return 0
}

// parseIntList parses a comma-separated list of integers and ranges (e.g., "0-3,7")
func parseIntList(s string) []int {
	var result []int
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, "-") {
			rangeParts := strings.SplitN(p, "-", 2)
			start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if err1 == nil && err2 == nil {
				for i := start; i <= end; i++ {
					result = append(result, i)
				}
			}
		} else {
			val, err := strconv.Atoi(p)
			if err == nil {
				result = append(result, val)
			}
		}
	}
	return result
}

// parseCacheSizeKB parses sysfs cache size strings like "32M", "96M", "1024K" to KB
func parseCacheSizeKB(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	suffix := s[len(s)-1]
	numStr := s[:len(s)-1]

	val, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}

	switch suffix {
	case 'K', 'k':
		return val
	case 'M', 'm':
		return val * 1024
	case 'G', 'g':
		return val * 1024 * 1024
	default:
		// Assume bytes, convert to KB
		v, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return v / 1024
	}
}

// CPUIndexToTopology maps a logical CPU index to host topology coordinates.
// Uses HostCPUTopology to find die-id, core-id, thread-id.
// Returns error if CPU ID is not found in the topology.
func CPUIndexToTopology(cpuID int, host models.HostCPUTopology) (dieID, coreID, threadID int, err error) {
	// Search through all dies and cores for this CPU
	for _, die := range host.Dies {
		for _, core := range die.CoreDetails {
			for i, t := range core.Threads {
				if t == cpuID {
					return die.ID, core.ID, i, nil
				}
			}
		}
	}
	return 0, 0, 0, fmt.Errorf("CPU %d not found in host topology", cpuID)
}
