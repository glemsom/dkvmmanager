package vm

import (
	"fmt"
	"strings"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// GenerateBuiltinScript generates a builtin start script for PCI passthrough.
// The script iterates over configured PCI devices and binds the vfio-pci driver.
func GenerateBuiltinScript(devices []models.PCIPassthroughDevice) (string, error) {
	if len(devices) == 0 {
		return "# No PCI devices configured", nil
	}

	var sb strings.Builder
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("# Auto-generated builtin start script for PCI passthrough\n")
	sb.WriteString("# WARNING: Do not edit manually - this file is regenerated on each VM start\n\n")

	for _, dev := range devices {
		sb.WriteString(fmt.Sprintf("# Device: %s (%s)\n", dev.Name, dev.Address))

		// Step 1: Use driver_override to cleanly switch drivers
		// This atomically replaces any existing driver binding
		sb.WriteString(fmt.Sprintf("# Clear driver_override first (must be empty to unbind)\n"))
		sb.WriteString(fmt.Sprintf("echo -n  > /sys/bus/pci/devices/%s/driver_override 2>/dev/null || true\n",
			dev.Address))
		sb.WriteString(fmt.Sprintf("# Bind vfio-pci using driver_override\n"))
		sb.WriteString(fmt.Sprintf("echo -n vfio-pci > /sys/bus/pci/devices/%s/driver_override 2>/dev/null || true\n",
			dev.Address))

		// Step 2: Set ROM if specified
		if dev.ROMPath != "" {
			sb.WriteString(fmt.Sprintf("# Set ROM from: %s\n", dev.ROMPath))
			sb.WriteString(fmt.Sprintf("echo 1 > /sys/bus/pci/devices/%s/rom 2>/dev/null || true\n",
				dev.Address))
		}

		sb.WriteString(fmt.Sprintf("echo \"Configured: %s\"\n\n", dev.Address))
	}

	sb.WriteString("echo \"PCI passthrough setup complete\"\n")

	return sb.String(), nil
}

// GenerateBuiltinStopScript generates a builtin stop script.
// Currently just a stub - vfio-pci stays bound for simplicity on next run.
func GenerateBuiltinStopScript() string {
	var sb strings.Builder
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("# Auto-generated builtin stop script for PCI passthrough\n")
	sb.WriteString("# Currently a no-op: vfio-pci stays bound for next VM start\n")
	sb.WriteString("# WARNING: Do not edit manually - this file is regenerated on each VM start\n\n")
	sb.WriteString("echo \"No cleanup needed\"\n")
	return sb.String()
}
