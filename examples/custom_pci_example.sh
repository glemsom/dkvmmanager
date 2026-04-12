#!/bin/bash
#
# Example custom start/stop script for PCI passthrough
#
# Usage:
#   ./custom_pci_example.sh start 0000:01:00.0 0000:02:00.0
#   ./custom_pci_example.sh stop 0000:01:00.0 0000:02:00.0
#
# Arguments:
#   $1 - Command: "start" or "stop"
#   $2+ - PCI device addresses (e.g., "0000:01:00.0")
#

set -e

# Exit codes:
#   0 - Success
#   1 - Invalid arguments
#   2 - No PCI devices provided
#   3 - Failed to rebind device

# Validate argument count
if [ $# -lt 1 ]; then
	echo "Usage: $0 <start|stop> [device1] [device2] ..."
	exit 1
fi

COMMAND="$1"
shift

# Validate command
if [ "$COMMAND" != "start" ] && [ "$COMMAND" != "stop" ]; then
	echo "Error: First argument must be 'start' or 'stop'"
	exit 1
fi

# Check if any devices provided
if [ $# -eq 0 ]; then
	echo "No PCI devices provided - nothing to do"
	exit 0
fi

# Function to rebind a PCI device to vfio-pci
rebind_device() {
	local pci_addr="$1"
	local driver="vfio-pci"

	# Extract domain:bus:slot.func from address (e.g., 0000:01:00.0 -> 0000:01:00.0)
	local sysfs_device="/sys/bus/pci/devices/$pci_addr"

	if [ ! -d "$sysfs_device" ]; then
		echo "Warning: Device $pci_addr not found in /sys/bus/pci/devices/"
		return 1
	fi

	# Get current driver (if any)
	local current_driver
	if [ -L "$sysfs_device/driver" ]; then
		current_driver=$(basename "$(readlink -f "$sysfs_device/driver")")
	fi

	if [ "$COMMAND" = "start" ]; then
		# Unbind from current driver (if bound)
		if [ -n "$current_driver" ] && [ -w "$sysfs_device/driver/unbind" ]; then
			echo "Unbinding $pci_addr from $current_driver"
			echo "$pci_addr" >"$sysfs_device/driver/unbind" 2>/dev/null || true
		fi

		# Bind to vfio-pci
		if [ -w "/sys/bus/pci/drivers/$driver/bind" ]; then
			echo "Binding $pci_addr to $driver"
			echo "$pci_addr" >"/sys/bus/pci/drivers/$driver/bind"
			echo "Success: $pci_addr bound to vfio-pci"
		else
			echo "Error: Cannot bind to $driver - driver may not be loaded"
			exit 3
		fi
	elif [ "$COMMAND" = "stop" ]; then
		# Unbind from vfio-pci (no rebind needed on stop, but good practice to clean up)
		if [ -w "$sysfs_device/driver/unbind" ]; then
			echo "Unbinding $pci_addr from $driver"
			echo "$pci_addr" >"$sysfs_device/driver/unbind" 2>/dev/null || true
		fi
		echo "Success: $pci_addr unbound from vfio-pci"
	fi
}

# Process each device
echo "========================================="
echo "Custom PCI Script - $COMMAND"
echo "========================================="
echo "Received ${#} PCI device(s):"
for dev in "$@"; do
	echo "  - $dev"
done
echo ""

for pci_addr in "$@"; do
	echo "Processing $pci_addr..."
	rebind_device "$pci_addr"
	echo ""
done

echo "========================================="
echo "Done - $COMMAND complete for all devices"
echo "========================================="

exit 0
