// Package vm provides virtual machine management functionality
package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// SetPCIPassthroughConfig sets the PCI passthrough configuration for this VM
func (r *VMRunner) SetPCIPassthroughConfig(cfg models.PCIPassthroughConfig) {
	r.pciPassthroughConfig = cfg
}

// SetUSBPassthroughConfig sets the USB passthrough configuration for this VM
func (r *VMRunner) SetUSBPassthroughConfig(cfg models.USBPassthroughConfig) {
	r.usbPassthroughConfig = cfg
}

// SetCPUOptions sets the CPU feature options for this VM
func (r *VMRunner) SetCPUOptions(opts models.CPUOptions) {
	r.cpuOptions = opts
}

// SetCPUTopology sets the global CPU topology for this VM
func (r *VMRunner) SetCPUTopology(topo models.CPUTopology) {
	r.cpuTopology = topo
}

// buildQEMUArgs constructs the QEMU command line arguments from VM config
func (r *VMRunner) buildQEMUArgs(vmDataDir string) []string {
	var args []string

	// Machine and acceleration
	args = append(args,
		"-name", fmt.Sprintf("%s,debug-threads=on", r.vm.Name),
		"-nodefaults", "-no-user-config",
		"-accel", "accel=kvm,kernel-irqchip=split",
		"-machine", "q35,mem-merge=off,vmport=off,dump-guest-core=off",
	)

	// QMP socket
	args = append(args,
		"-qmp", fmt.Sprintf("unix:%s,server=on,wait=off", r.socketPath),
	)

	// Memory and clock
	rtcBase := "localtime"
	if r.cpuOptions.RTCUTC {
		rtcBase = "utc"
	}
	args = append(args,
		"-mem-prealloc",
		"-overcommit", "mem-lock=on,cpu-pm=on",
		"-rtc", fmt.Sprintf("base=%s,clock=vm,driftfix=slew", rtcBase),
		"-serial", "none",
		"-parallel", "none",
	)

	// Networking
	mac := r.vm.MAC
	if mac == "" {
		mac = "52:54:00:00:00:00"
	}
	netMode := r.vm.NetworkMode
	if netMode == "" {
		netMode = "nat"
	}
	switch netMode {
	case "bridge":
		if r.cfg.NetworkBridge != "" {
			args = append(args,
				"-netdev", fmt.Sprintf("bridge,id=hostnet0,br=%s", r.cfg.NetworkBridge),
				"-device", fmt.Sprintf("virtio-net-pci,netdev=hostnet0,id=net0,mac=%s", mac),
			)
		} else {
			// Fall back to NAT when no bridge is configured
			args = append(args,
				"-netdev", "user,id=hostnet0",
				"-device", fmt.Sprintf("virtio-net-pci,netdev=hostnet0,id=net0,mac=%s", mac),
			)
		}
	case "nat":
		args = append(args,
			"-netdev", "user,id=hostnet0",
			"-device", fmt.Sprintf("virtio-net-pci,netdev=hostnet0,id=net0,mac=%s", mac),
		)
	default:
		// Unknown mode — fall back to NAT
		args = append(args,
			"-netdev", "user,id=hostnet0",
			"-device", fmt.Sprintf("virtio-net-pci,netdev=hostnet0,id=net0,mac=%s", mac),
		)
	}

	// Hugepages
	memMB := 8192 // Default 8GB
	args = append(args,
		"-object", fmt.Sprintf("memory-backend-memfd,id=mem,size=%dM,hugetlb=on,hugetlbsize=2M,prealloc=on", memMB),
		"-machine", "memory-backend=mem",
	)

	// Disable S3/S4 sleep states
	args = append(args,
		"-global", "ICH9-LPC.disable_s3=1",
		"-global", "ICH9-LPC.disable_s4=1",
		"-global", "kvm-pit.lost_tick_policy=discard",
	)

	// TPM
	tpmSock := filepath.Join(vmDataDir, "tpm.sock")
	if _, err := os.Stat(tpmSock); err == nil {
		args = append(args,
			"-chardev", fmt.Sprintf("socket,id=chrtpm,path=%s,server=on,wait=off", tpmSock),
			"-tpmdev", "emulator,id=tpm0,chardev=chrtpm",
			"-device", "tpm-tis,tpmdev=tpm0",
		)
	}

	// Guest agent
	args = append(args,
		"-device", "virtio-serial-pci,id=virtio-serial0",
		"-chardev", "socket,id=guestagent,path=/tmp/qga.sock,server=on,wait=off",
		"-device", "virtserialport,chardev=guestagent,name=org.qemu.guest_agent.0",
	)

	// Boot options
	args = append(args,
		"-boot", "menu=on,splash-time=5000",
		"-fw_cfg", "opt/ovmf/X-PciMmio64Mb,string=65536",
	)

	// Graphics
	if r.vm.VNCListen == "" {
		args = append(args, "-nographic", "-vga", "none")
	} else {
		args = append(args, "-vga", "std", "-vnc", r.vm.VNCListen)
	}

	// CPU topology - use global CPUTopology if enabled
	if r.cpuTopology.Enabled && len(r.cpuTopology.SelectedCPUs) > 0 {
		numCPUs := len(r.cpuTopology.SelectedCPUs)
		args = append(args, "-smp", fmt.Sprintf("%d,sockets=1,cores=%d,threads=1", numCPUs, numCPUs))
	}

	// OVMF firmware
	codePath := filepath.Join(vmDataDir, "OVMF_CODE.fd")
	varsPath := filepath.Join(vmDataDir, "OVMF_VARS.fd")
	args = append(args,
		"-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", codePath),
		"-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", varsPath),
	)

	// Hard disks
	if len(r.vm.HardDisks) > 0 {
		args = append(args, "-device", "virtio-scsi-pci,id=scsi")
		for i, disk := range r.vm.HardDisks {
			if disk == "" {
				continue
			}
			args = append(args,
				"-drive", fmt.Sprintf("if=none,cache=none,aio=native,discard=unmap,detect-zeroes=unmap,format=raw,file=%s,id=drive%d", disk, i),
				"-device", fmt.Sprintf("scsi-hd,drive=drive%d", i),
			)
		}
	}

	// CDROMs
	for _, cd := range r.vm.CDROMs {
		if cd == "" {
			continue
		}
		args = append(args, "-drive", fmt.Sprintf("file=%s,media=cdrom", cd))
	}

	// PCI Passthrough
	pciDevices := r.pciPassthroughConfig.Devices
	// Backward compatibility: if no PCI config but GPUROM is set, use legacy passthrough
	if len(pciDevices) == 0 && r.vm.GPUROM != "" {
		pciDevices = []models.PCIPassthroughDevice{
			{Address: "0000:01:00.0", ROMPath: r.vm.GPUROM},
		}
	}

	// Build mapping of base device (domain:bus:device) to root port index
	// This ensures multifunction devices share the same root port
	baseDeviceToPort := make(map[string]int) // "domain:bus:dev" -> port number
	portCount := 0
	for _, dev := range pciDevices {
		// Extract base device (domain:bus:device) by removing function number
		baseDev := extractBaseDevice(dev.Address)
		if _, exists := baseDeviceToPort[baseDev]; !exists {
			portCount++
			baseDeviceToPort[baseDev] = portCount
		}
	}

	// Create PCIe root ports for each unique base device
	for portNum := 1; portNum <= portCount; portNum++ {
		busName := fmt.Sprintf("root_port%d", portNum)
		portArgs := fmt.Sprintf("pcie-root-port,id=%s,slot=%d,chassis=%d", busName, portNum, portNum)
		args = append(args, "-device", portArgs)
	}

	for i, dev := range pciDevices {
		baseDev := extractBaseDevice(dev.Address)
		portNum := baseDeviceToPort[baseDev]
		busName := fmt.Sprintf("root_port%d", portNum)
		funcNum := extractFunctionNumber(dev.Address)

		// Detect multifunction: check if next device shares same bus:device
		isMultifunctionGroup := false
		if i+1 < len(pciDevices) {
			isMultifunctionGroup = IsMultifunction(dev.Address, pciDevices[i+1].Address)
		} else if i > 0 {
			// Check if previous device was same bus:device (this is the last function)
			isMultifunctionGroup = IsMultifunction(pciDevices[i-1].Address, dev.Address)
		}

		// Build device args
		// - Primary function (func 0) in multifunction group: addr=00.0,multifunction=on
		// - Secondary functions (func 1+) in multifunction group: addr=00.N where N is function number
		// - Single function devices: addr=00.0
		devArgs := fmt.Sprintf("vfio-pci,host=%s,bus=%s", dev.Address, busName)
		if funcNum == 0 {
			// Primary function gets the slot address
			devArgs += ",addr=00.0"
			if isMultifunctionGroup {
				devArgs += ",multifunction=on"
			}
		} else if funcNum > 0 && isMultifunctionGroup {
			// Secondary functions in multifunction group: use addr=00.N
			devArgs += fmt.Sprintf(",addr=00.%d", funcNum)
		}
		// Non-multifunction devices (single function) also get addr=00.0
		if funcNum != 0 && !isMultifunctionGroup {
			devArgs += ",addr=00.0"
		}

		if dev.ROMPath != "" {
			devArgs += fmt.Sprintf(",romfile=%s", dev.ROMPath)
		}

		args = append(args, "-device", devArgs)
	}

	// USB Passthrough - one xHCI controller per device
	// Note: Each USB device needs its own xHCI controller (no id specified),
	// and usb-host must NOT specify bus= (QEMU auto-attaches to the previous xHCI)
	for _, dev := range r.usbPassthroughConfig.Devices {
		args = append(args, "-device", "qemu-xhci")
		devArgs := fmt.Sprintf("usb-host,vendorid=0x%s,productid=0x%s", dev.Vendor, dev.Product)
		args = append(args, "-device", devArgs)
	}

	// CPU options
	cpuOptsStr := r.buildCPUOptsString()
	if cpuOptsStr != "" {
		args = append(args, "-cpu", fmt.Sprintf("host,%s", cpuOptsStr))
	} else {
		args = append(args, "-cpu", "host")
	}

	return args
}

// buildCPUOptsString builds CPU feature flags string from global CPU options
// for use with -cpu host,<flags> QEMU argument
func (r *VMRunner) buildCPUOptsString() string {
	var flags []string

	opts := r.cpuOptions
	if opts.HideKVM {
		flags = append(flags, "kvm=off")
	}
	if opts.VendorID != "" {
		flags = append(flags, fmt.Sprintf("-hypervisor,vendor_id=%s", opts.VendorID))
	}
	if opts.HVFrequency {
		flags = append(flags, "+hv-frequencies")
	}
	if opts.HVRelaxed {
		flags = append(flags, "+hv-relaxed")
	}
	if opts.HVReset {
		flags = append(flags, "+hv-reset")
	}
	if opts.HVRuntime {
		flags = append(flags, "+hv-runtime")
	}
	if opts.HVSpinlocks != "" {
		flags = append(flags, fmt.Sprintf("+hv-spinlocks=%s", opts.HVSpinlocks))
	}
	if opts.HVStimer {
		flags = append(flags, "+hv-stimer")
	}
	if opts.HVSyncIC {
		flags = append(flags, "+hv-synic")
	}
	if opts.HVTime {
		flags = append(flags, "+hv-time")
	}
	if opts.HVVapic {
		flags = append(flags, "+hv-vapic")
	}
	if opts.HVVPIndex {
		flags = append(flags, "+hv-vpindex")
	}
	if opts.HVNoNonarchCoresharing {
		flags = append(flags, "hv-no-nonarch-coresharing=on")
	}
	if opts.HVTLBFlush {
		flags = append(flags, "+hv-tlbflush")
	}
	if opts.HVTLBFlushExt {
		flags = append(flags, "+hv-tlbflush-ext")
	}
	if opts.HVIPI {
		flags = append(flags, "+hv-ipi")
	}
	if opts.HVAVIC {
		flags = append(flags, "+hv-avic")
	}
	if opts.TopoExt {
		flags = append(flags, "+topoext")
	}
	if opts.L3Cache {
		flags = append(flags, "+l3-cache")
	}
	if opts.X2APIC {
		flags = append(flags, "+x2apic")
	}
	if opts.Migratable {
		flags = append(flags, "migratable=no")
	}
	if opts.InvTSC {
		flags = append(flags, "+invtsc")
	}

	return strings.Join(flags, ",")
}



// extractBaseDevice extracts the base device address (domain:bus:device) from a PCI address
// by removing the function number. E.g., "0000:03:00.1" -> "0000:03:00"
func extractBaseDevice(addr string) string {
	parts := strings.Split(addr, ".")
	if len(parts) != 2 {
		return addr // Return as-is if not in expected format
	}
	return parts[0]
}

// extractFunctionNumber extracts the function number from a PCI address.
// Returns -1 if invalid. E.g., "0000:03:00.1" -> 1
func extractFunctionNumber(addr string) int {
	parts := strings.Split(addr, ".")
	if len(parts) != 2 {
		return -1 // Invalid format
	}
	var fn int
	_, err := fmt.Sscanf(parts[1], "%d", &fn)
	if err != nil {
		return -1
	}
	return fn
}
