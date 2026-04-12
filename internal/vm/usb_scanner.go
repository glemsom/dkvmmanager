package vm

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/glemsom/dkvmmanager/internal/models"
)

const (
	// SysfsUSBBasePath is the base path for USB device entries
	SysfsUSBBasePath = "/sys/bus/usb/devices"
)

// lsusbLineRegex parses lines like:
// Bus 001 Device 005: ID 046d:c52b Logitech, Inc. Unifying Receiver
var lsusbLineRegex = regexp.MustCompile(
	`^Bus\s+(\d+)\s+Device\s+(\d+):\s+ID\s+([0-9a-fA-F]+):([0-9a-fA-F]+)\s+(.+)$`,
)

// USBScanner scans the host for USB devices
type USBScanner struct {
	sysfsPath string
	lsusbPath string
}

// NewUSBScanner creates a new USB scanner using default paths
func NewUSBScanner() *USBScanner {
	return &USBScanner{
		sysfsPath: SysfsUSBBasePath,
		lsusbPath: "lsusb",
	}
}

// ScanDevices scans the host for all USB devices
func (s *USBScanner) ScanDevices() ([]models.USBDevice, error) {
	output, err := s.runLsusb()
	if err != nil {
		return nil, fmt.Errorf("failed to run lsusb: %w", err)
	}

	devices := s.parseLsusbOutput(output)

	// Enrich with sysfs bus device IDs
	for i := range devices {
		devices[i].ID = s.findBusDeviceID(devices[i].Vendor, devices[i].Product)
	}

	return devices, nil
}

// runLsusb executes lsusb and returns stdout
func (s *USBScanner) runLsusb() (string, error) {
	cmd := exec.Command(s.lsusbPath)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// parseLsusbOutput parses the output of lsusb
func (s *USBScanner) parseLsusbOutput(output string) []models.USBDevice {
	var devices []models.USBDevice

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := lsusbLineRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		vendor := strings.ToLower(matches[3])
		product := strings.ToLower(matches[4])
		name := strings.TrimSpace(matches[5])

		devices = append(devices, models.USBDevice{
			Vendor:  vendor,
			Product: product,
			Name:    name,
		})
	}

	return devices
}

// findBusDeviceID looks up the sysfs bus device ID for a given vendor:product
// by scanning /sys/bus/usb/devices/*/idVendor and idProduct.
// Returns the bus device ID (e.g., "1-1") or empty string if not found.
func (s *USBScanner) findBusDeviceID(vendor, product string) string {
	entries, err := os.ReadDir(s.sysfsPath)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirPath := s.sysfsPath + "/" + entry.Name()

		vendorBytes, err := os.ReadFile(dirPath + "/idVendor")
		if err != nil {
			continue
		}
		productBytes, err := os.ReadFile(dirPath + "/idProduct")
		if err != nil {
			continue
		}

		if strings.TrimSpace(string(vendorBytes)) == vendor &&
			strings.TrimSpace(string(productBytes)) == product {
			return entry.Name()
		}
	}

	return ""
}

// ValidateUSBDevices checks that the given devices exist on the host.
// Returns a list of warnings (non-fatal) and errors (fatal).
func ValidateUSBDevices(devices []models.USBPassthroughDevice) (warnings []string, errors []string) {
	// If no devices to validate, return early
	if len(devices) == 0 {
		return
	}

	scanner := NewUSBScanner()
	allDevices, err := scanner.ScanDevices()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to scan USB devices: %v", err))
		return
	}

	// Build lookup by vendor:product
	type vendorProduct struct {
		vendor  string
		product string
	}
	deviceMap := make(map[vendorProduct]bool)
	for _, d := range allDevices {
		deviceMap[vendorProduct{d.Vendor, d.Product}] = true
	}

	for _, dev := range devices {
		if !deviceMap[vendorProduct{dev.Vendor, dev.Product}] {
			errors = append(errors, fmt.Sprintf(
				"USB device %s [%s:%s] not found on host", dev.Name, dev.Vendor, dev.Product))
		}
	}
	return
}
