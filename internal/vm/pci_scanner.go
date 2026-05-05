package vm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/glemsom/dkvmmanager/internal/models"
)

const (
	// SysfsIOMMUBasePath is the base path for IOMMU group symlinks
	SysfsIOMMUBasePath = "/sys/bus/pci/devices"
	// SysfsIOMMUGroupPath is the base path for IOMMU group directories
	SysfsIOMMUGroupPath = "/sys/kernel/iommu_groups"
)

// lspciLineRegex parses lines like:
// 0000:01:00.0 VGA compatible controller [0300]: Advanced Micro Devices, Inc. [AMD/ATI] Navi 48 [1002:7550] (rev c0)
var lspciLineRegex = regexp.MustCompile(
	`^(\S+)\s+(.+?)\s+\[([0-9a-f]+)\]:\s+(.+?)\s+\[([0-9a-f]+):([0-9a-f]+)\]`,
)

// PCIScanner scans the host for PCI devices
type PCIScanner struct {
	sysfsPath string
	lspciPath string
}

// NewPCIScanner creates a new PCI scanner using default paths
func NewPCIScanner() *PCIScanner {
	return &PCIScanner{
		sysfsPath: SysfsIOMMUBasePath,
		lspciPath: "lspci",
	}
}

// ScanDevices scans the host for all PCI devices
func (s *PCIScanner) ScanDevices() ([]models.PCIDevice, error) {
	output, err := s.runLspci()
	if err != nil {
		return nil, fmt.Errorf("failed to run lspci: %w", err)
	}

	devices := s.parseLspciOutput(output)

	// Enrich with IOMMU group info
	for i := range devices {
		group := s.readIOMMUGroup(devices[i].Address)
		devices[i].IOMMUGroup = group
	}

	return devices, nil
}

// runLspci executes lspci -nn -D and returns stdout
func (s *PCIScanner) runLspci() (string, error) {
	cmd := exec.Command(s.lspciPath, "-nn", "-D")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// parseLspciOutput parses the output of lspci -nn -D
func (s *PCIScanner) parseLspciOutput(output string) []models.PCIDevice {
	var devices []models.PCIDevice

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := lspciLineRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		addr := matches[1]
		classCode := matches[3]
		name := strings.TrimSpace(matches[4])
		vendor := matches[5]
		device := matches[6]

		// Class 0300 = VGA compatible controller, 0302 = 3D controller, 0380 = Display controller
		isGPU := classCode == "0300" || classCode == "0302" || classCode == "0380"
		// Class 0c03 = USB controller
		isUSB := classCode == "0c03"
		// Class 0604 = PCI-to-PCI bridge (PCIe switch ports, root ports, downstream ports)
		isBridge := classCode == "0604"

		devices = append(devices, models.PCIDevice{
			Address:    addr,
			Vendor:     vendor,
			Device:     device,
			ClassCode:  classCode,
			Name:       name,
			IsGPU:      isGPU,
			IsUSB:      isUSB,
			IsBridge:   isBridge,
			IOMMUGroup: -1, // Will be filled by IOMMU lookup
		})
	}

	return devices
}

// readIOMMUGroup reads the IOMMU group number for a PCI address from sysfs
// Returns -1 if IOMMU groups are not available
func (s *PCIScanner) readIOMMUGroup(addr string) int {
	linkPath := filepath.Join(s.sysfsPath, addr, "iommu_group")
	target, err := os.Readlink(linkPath)
	if err != nil {
		return -1
	}

	// Target is something like /sys/kernel/iommu_groups/1
	baseName := filepath.Base(target)
	groupNum, err := strconv.Atoi(baseName)
	if err != nil {
		return -1
	}

	return groupNum
}

// GetIOMMUGroupDevices returns all PCI addresses in the given IOMMU group
func GetIOMMUGroupDevices(groupNum int) []string {
	if groupNum < 0 {
		return nil
	}

	groupDir := filepath.Join(SysfsIOMMUGroupPath, strconv.Itoa(groupNum), "devices")
	entries, err := os.ReadDir(groupDir)
	if err != nil {
		return nil
	}

	var addrs []string
	for _, entry := range entries {
		if !entry.IsDir() && entry.Type()&os.ModeSymlink == 0 {
			continue
		}
		addrs = append(addrs, entry.Name())
	}
	return addrs
}

// ValidatePCIDevices checks that the given devices exist and their IOMMU groups are valid
// Returns a list of warnings (non-fatal) and errors (fatal)
func ValidatePCIDevices(devices []models.PCIPassthroughDevice) (warnings []string, errors []string) {
	scanner := NewPCIScanner()
	allDevices, err := scanner.ScanDevices()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to scan PCI devices: %v", err))
		return
	}

	// Build lookup
	deviceMap := make(map[string]models.PCIDevice)
	for _, d := range allDevices {
		deviceMap[d.Address] = d
	}

	for _, dev := range devices {
		found, ok := deviceMap[dev.Address]
		if !ok {
			errors = append(errors, fmt.Sprintf("Device %s not found on host", dev.Address))
			continue
		}

		// Warn if ROM file doesn't exist
		if dev.ROMPath != "" {
			if _, err := os.Stat(dev.ROMPath); err != nil {
				warnings = append(warnings, fmt.Sprintf("ROM file for %s not found: %s", dev.Address, dev.ROMPath))
			}
		}

		// Warn about IOMMU group conflicts
		if found.IOMMUGroup >= 0 {
			groupAddrs := GetIOMMUGroupDevices(found.IOMMUGroup)
			for _, addr := range groupAddrs {
				if addr != dev.Address {
					warnings = append(warnings, fmt.Sprintf(
						"Device %s is in IOMMU group %d with %s — both must be passed through together",
						dev.Address, found.IOMMUGroup, addr))
				}
			}
		}
	}
	return
}

// IsMultifunction returns true if two PCI addresses share the same bus:device
// (differ only in function number)
func IsMultifunction(addr1, addr2 string) bool {
	// PCI addresses are formatted as domain:bus:device.function
	parts1 := strings.Split(addr1, ".")
	parts2 := strings.Split(addr2, ".")
	if len(parts1) != 2 || len(parts2) != 2 {
		return false
	}
	return parts1[0] == parts2[0] // Same domain:bus:device
}
