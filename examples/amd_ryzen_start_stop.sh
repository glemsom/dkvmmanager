#!/bin/bash
#
# AMD Ryzen GPU passthrough start/stop script
#
# Usage:
#   ./amd_ryzen_start_stop.sh start 0000:01:00.0 0000:02:00.0
#   ./amd_ryzen_start_stop.sh stop 0000:01:00.0 0000:02:00.0
#
# Features:
#   - Blocks iGPU from loading amdgpu driver using driver_override
#   - Reloads amdgpu driver for dGPU before passthrough
#   - VFIO-PCI binding for passthrough devices
#
# Arguments:
#   $1 - Command: "start" or "stop"
#   $2+ - PCI device addresses (e.g., "0000:01:00.0")

set -e

# Exit codes:
#   0 - Success
#   1 - Invalid arguments
#   2 - Failed to rebind device

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

PASSDEVICES="$@"

# ════════════════════════════════════════════════════════════════════════════════════
# Helper Functions (inspired by amd_9000_StartStop.sh)
# ════════════════════════════════════════════════════════════════════════════════════

# Get the current driver for a PCI device
get_current_driver() {
	local device=$1
	local sysfs_device="/sys/bus/pci/devices/${device}"

	if [ -L "$sysfs_device/driver" ]; then
		basename "$(readlink -f "$sysfs_device/driver")"
	else
		echo ""
	fi
}

# Check if device is a VGA device
is_vga_device() {
	local device=$1
	local sysfs_device="/sys/bus/pci/devices/${device}"

	if [ -f "$sysfs_device/class" ]; then
		local class
		class=$(cat "$sysfs_device/class")
		# VGA compatible controller: 0x0300xx, 3D controller: 0x0302xx
		[[ "$class" == 0x03* ]]
	else
		return 1
	fi
}

# Unbind driver from device
unbind_driver() {
	local device=$1
	local driver=$2
	local sysfs_device="/sys/bus/pci/devices/${device}"

	if [ -n "$driver" ] && [ -e "/sys/bus/pci/drivers/${driver}/unbind" ]; then
		echo "Unbinding driver '$driver' from ${device}"
		echo "${device}" >"/sys/bus/pci/drivers/${driver}/unbind" 2>/dev/null || true
	fi
}

# Bind driver to device
bind_driver() {
	local device=$1
	local driver=$2

	if [ -e "/sys/bus/pci/drivers/${driver}/bind" ]; then
		echo "Binding driver '$driver' to ${device}"
		echo "${device}" >"/sys/bus/pci/drivers/${driver}/bind" 2>/dev/null || true
	fi
}

# ════════════════════════════════════════════════════════════════════════════════════════════
# Start: Protect iGPU, reload amdgpu driver, bind to vfio-pci
# ════════════════════════════════════════════════════════════════════════════════════

start_vm() {
	echo "========================================="
	echo "AMD Ryzen Start Script"
	echo "========================================="
	echo "Passthrough devices: $PASSDEVICES"
	echo ""

	# Detect VGA device from passthrough devices
	VGA_DEVICE=""
	for device in $PASSDEVICES; do
		if is_vga_device "$device"; then
			VGA_DEVICE="$device"
			echo "Detected VGA device: 0000:${VGA_DEVICE}"
			break
		fi
	done

	# Protect iGPU from amdgpu driver (before loading the module)
	echo "Protecting iGPU from amdgpu driver..."
	for pci_device in /sys/bus/pci/devices/0000:*; do
		device=$(basename "$pci_device")
		class=$(cat "${pci_device}/class" 2>/dev/null)

		# Check if it's a VGA device (class 0x03*)
		if [[ "$class" == 0x03* ]]; then
			# Check if it's NOT in the passthrough list
			if ! echo "$PASSDEVICES" | grep -q "$device"; then
				echo "  Protecting iGPU 0000:${device} from amdgpu driver"
				echo "fake_driver" >"${pci_device}/driver_override" 2>/dev/null || true
			fi
		fi
	done

	# Load amdgpu module if VGA device detected (required before binding)
	if [ -n "$VGA_DEVICE" ]; then
		echo "Loading amdgpu kernel module"
		modprobe amdgpu 2>/dev/null || true
		sleep 1
	fi

	# Step 1: Unbind all passthrough devices from their current drivers
	echo ""
	echo "Step 1: Unbinding passthrough devices from current drivers"
	for device in $PASSDEVICES; do
		current_driver=$(get_current_driver "$device")

		if [ -n "$current_driver" ]; then
			unbind_driver "$device" "$current_driver"
		fi

		# Clean up any existing VFIO IDs
		local pciVendor pciDevice
		pciVendor=$(cat "/sys/bus/pci/devices/${device}/vendor" 2>/dev/null)
		pciDevice=$(cat "/sys/bus/pci/devices/${device}/device" 2>/dev/null)
		if [ -n "$pciVendor" ] && [ -n "$pciDevice" ]; then
			echo "$pciVendor $pciDevice" >/sys/bus/pci/drivers/vfio-pci/remove_id 2>/dev/null || true
		fi
	done

	# Step 2: Special handling for VGA device (AMD requires driver cycle)
	if [ -n "$VGA_DEVICE" ]; then
		echo ""
		echo "Step 2: Performing AMDGPU driver cycle for VGA device 0000:${VGA_DEVICE}"

		# Bind amdgpu driver specifically to this device
		echo "  Loading amdgpu driver on 0000:${VGA_DEVICE}"
		bind_driver "$VGA_DEVICE" "amdgpu"
		sleep 2

		# Unbind amdgpu driver from the device
		echo "  Unloading amdgpu driver from 0000:${VGA_DEVICE}"
		unbind_driver "$VGA_DEVICE" "amdgpu"
	fi

	sleep 2

	# Step 3: Bind all passthrough devices to vfio-pci
	echo ""
	echo "Step 3: Binding passthrough devices to vfio-pci"
	for device in $PASSDEVICES; do
		local pciVendor pciDevice
		pciVendor=$(cat "/sys/bus/pci/devices/${device}/vendor" 2>/dev/null)
		pciDevice=$(cat "/sys/bus/pci/devices/${device}/device" 2>/dev/null)

		if [ -n "$pciVendor" ] && [ -n "$pciDevice" ]; then
			echo "  Binding vfio-pci to ${pciVendor}:${pciDevice} (${device})"
			echo "$pciVendor $pciDevice" >/sys/bus/pci/drivers/vfio-pci/new_id 2>/dev/null || true
			sleep 2
		fi
	done

	echo ""
	echo "========================================="
	echo "AMD Ryzen Start Script completed"
	echo "========================================="
}

# ══════════════════════════════════════════════════════════════════════════��═════════
# Stop: Unbind from vfio-pci, cleanup driver_override
# ══════════════════════════════════════════════════════════════════════════════════════

stop_vm() {
	echo "========================================="
	echo "AMD Ryzen Stop Script"
	echo "========================================="
	echo "Passthrough devices: $PASSDEVICES"
	echo ""

	# Unbind all devices from vfio-pci
	echo "Step 1: Unbinding devices from vfio-pci"
	for device in $PASSDEVICES; do
		local sysfs_device="/sys/bus/pci/devices/${device}"

		if [ -w "$sysfs_device/driver/unbind" ]; then
			echo "  Unbinding vfio-pci from ${device}"
			echo "${device}" >"$sysfs_device/driver/unbind" 2>/dev/null || true
		fi

		# Remove VFIO IDs
		local pciVendor pciDevice
		pciVendor=$(cat "$sysfs_device/vendor" 2>/dev/null)
		pciDevice=$(cat "$sysfs_device/device" 2>/dev/null)
		if [ -n "$pciVendor" ] && [ -n "$pciDevice" ]; then
			echo "$pciVendor $pciDevice" >/sys/bus/pci/drivers/vfio-pci/remove_id 2>/dev/null || true
		fi
	done

	# Cleanup driver_override protections
	echo ""
	echo "Step 2: Cleaning up iGPU driver overrides"
	for pci_device in /sys/bus/pci/devices/0000:*; do
		device=$(basename "$pci_device")
		class=$(cat "${pci_device}/class" 2>/dev/null)

		# Check if it's a VGA device that was protected
		if [[ "$class" == 0x03* ]]; then
			if ! echo "$PASSDEVICES" | grep -q "$device"; then
				if [ -f "$pci_device/driver_override" ]; then
					# Read current override
					local override
					override=$(cat "$pci_device/driver_override" 2>/dev/null)
					if [ "$override" = "fake_driver" ]; then
						echo "  Releasing iGPU 0000:${device} driver override"
						echo "none" >"$pci_device/driver_override" 2>/dev/null || true
					fi
				fi
			fi
		fi
	done

	# Note: We intentionally do NOT reload amdgpu driver here.
	# The system will handle it on next boot or when needed.
	# Unloading amdgpu can cause issues with display if running in GUI.

	echo ""
	echo "========================================="
	echo "AMD Ryzen Stop Script completed"
	echo "========================================="
}

# ════════════════════════════════════════════════════════════════════════════════════
# Main
# ══════════════════════════════════════════════════════════════════════════════════════

if [ "$COMMAND" = "start" ]; then
	start_vm
elif [ "$COMMAND" = "stop" ]; then
	stop_vm
fi

exit 0
