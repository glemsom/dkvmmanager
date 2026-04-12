package models

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// loadDevices loads block devices from the system
func (m *BlockDeviceModel) loadDevices() tea.Msg {
	devices, err := m.listBlockDevices()
	if err != nil {
		m.errorMsg = fmt.Sprintf("Failed to list block devices: %v", err)
		return BlockDeviceLoadedMsg{}
	}

	m.devices = devices
	if m.selectedIndex >= len(m.devices) {
		m.selectedIndex = 0
	}
	return BlockDeviceLoadedMsg{}
}

// listBlockDevices lists available block devices using lsblk
func (m *BlockDeviceModel) listBlockDevices() ([]BlockDevice, error) {
	// Try lsblk first (preferred)
	cmd := exec.Command("lsblk", "-o", "NAME,SIZE,TYPE,RO", "-n", "-p")
	output, err := cmd.Output()
	if err == nil {
		return m.parseLSBlkOutput(string(output))
	}

	// Fallback to reading /sys/block
	return m.listSysBlock()
}

// parseLSBlkOutput parses lsblk output
func (m *BlockDeviceModel) parseLSBlkOutput(output string) ([]BlockDevice, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var devices []BlockDevice

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse: NAME SIZE TYPE RO
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		name := parts[0]
		size := parts[1]
		devType := parts[2]

		// Only include disk and partition devices
		if devType != "disk" && devType != "part" {
			continue
		}

		readOnly := false
		if len(parts) >= 4 && parts[3] == "1" {
			readOnly = true
		}

		devices = append(devices, BlockDevice{
			Path:     "/dev/" + name,
			Name:     name,
			Size:     size,
			Type:     devType,
			ReadOnly: readOnly,
		})
	}

	// Sort by name
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Name < devices[j].Name
	})

	return devices, nil
}

// listSysBlock lists block devices from /sys/block
func (m *BlockDeviceModel) listSysBlock() ([]BlockDevice, error) {
	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, err
	}

	var devices []BlockDevice
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip loop devices
		if strings.HasPrefix(name, "loop") {
			continue
		}

		devices = append(devices, BlockDevice{
			Path: "/dev/" + name,
			Name: name,
			Size: "unknown",
			Type: "disk",
		})
	}

	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Name < devices[j].Name
	})

	return devices, nil
}
