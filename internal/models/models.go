// Package models provides data models for DKVM Manager
package models

import "time"

// VM represents a virtual machine configuration
type VM struct {
	ID        string    `json:"id" yaml:"id"`
	Name      string    `json:"name" yaml:"name"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`

	// Storage
	HardDisks []string `json:"harddisks" yaml:"harddisks"` // Block devices or file paths
	CDROMs    []string `json:"cdroms" yaml:"cdroms"`       // ISO file paths

	// GPU (deprecated: use PCIPassthroughConfig per-device ROM instead)
	GPUROM string `json:"gpu_rom" yaml:"gpu_rom"` // Deprecated: VBIOS ROM file, use PCIPassthroughConfig

	// Network
	MAC         string `json:"mac" yaml:"mac"`
	NetworkMode string `json:"network_mode" yaml:"network_mode"` // "bridge" or "nat" (default: "nat")
	VNCListen   string `json:"vnc_listen" yaml:"vnc_listen"`     // VNC binding (e.g., "0.0.0.0:0")

	// TPM
	TPMEnabled bool `json:"tpm_enabled" yaml:"tpm_enabled"` // Enable emulated TPM via swtpm
}

// PCIDevice represents a PCI device discovered on the host
type PCIDevice struct {
	Address    string `json:"address" yaml:"address"`         // e.g., "0000:01:00.0"
	Vendor     string `json:"vendor" yaml:"vendor"`           // e.g., "10de"
	Device     string `json:"device" yaml:"device"`           // e.g., "1b80"
	ClassCode  string `json:"class_code" yaml:"class_code"`   // e.g., "0300"
	Name       string `json:"name" yaml:"name"`               // e.g., "NVIDIA GeForce GTX 1080"
	IsGPU      bool   `json:"is_gpu" yaml:"is_gpu"`           // True if VGA/GPU device
	IsUSB      bool   `json:"is_usb" yaml:"is_usb"`           // True if USB controller
	IOMMUGroup int    `json:"iommu_group" yaml:"iommu_group"` // IOMMU group number, -1 if none
}

// PCIPassthroughDevice represents a PCI device selected for passthrough with per-device ROM
type PCIPassthroughDevice struct {
	Address   string `json:"address" yaml:"address"`       // PCI address (e.g., "0000:01:00.0")
	ROMPath   string `json:"rom_path" yaml:"rom_path"`     // Optional VBIOS ROM file path
	Vendor    string `json:"vendor" yaml:"vendor"`         // Vendor ID
	Device    string `json:"device" yaml:"device"`         // Device ID
	Name      string `json:"name" yaml:"name"`             // Human-readable name
	ClassCode string `json:"class_code" yaml:"class_code"` // PCI class code
}

// PCIPassthroughConfig holds the global PCI passthrough configuration
type PCIPassthroughConfig struct {
	Devices []PCIPassthroughDevice `json:"devices" yaml:"devices"`
}

// USBDevice represents a USB device
type USBDevice struct {
	ID       string `json:"id" yaml:"id"`           // e.g., "1-1"
	Vendor   string `json:"vendor" yaml:"vendor"`   // e.g., "046d"
	Product  string `json:"product" yaml:"product"` // e.g., "c52b"
	Name     string `json:"name" yaml:"name"`       // e.g., "Logitech Unifying Receiver"
	Selected bool   `json:"selected" yaml:"selected"`
}

// USBPassthroughDevice represents a USB device selected for passthrough
type USBPassthroughDevice struct {
	Vendor  string `json:"vendor" yaml:"vendor"`   // Vendor ID (e.g., "046d")
	Product string `json:"product" yaml:"product"` // Product ID (e.g., "c52b")
	Name    string `json:"name" yaml:"name"`       // Human-readable name
	BusID   string `json:"bus_id" yaml:"bus_id"`   // Bus device ID (e.g., "1-1")
}

// USBPassthroughConfig holds the global USB passthrough configuration
type USBPassthroughConfig struct {
	Devices []USBPassthroughDevice `json:"devices" yaml:"devices"`
}

// CPUCore represents a physical CPU core with its thread IDs
type CPUCore struct {
	ID      int   `json:"id" yaml:"id"`           // Physical core ID
	Threads []int `json:"threads" yaml:"threads"` // Logical CPU IDs for this core's threads
	DieID   int   `json:"die_id" yaml:"die_id"`   // Die this core belongs to
}

// CPUDie represents a single CPU die with its topology details
type CPUDie struct {
	ID          int       `json:"id" yaml:"id"`                     // Die ID
	Cores       int       `json:"cores" yaml:"cores"`               // Physical cores on this die
	Threads     int       `json:"threads" yaml:"threads"`           // Threads per core
	LogicalCPUs []int     `json:"logical_cpus" yaml:"logical_cpus"` // Logical CPU IDs on this die
	L3CacheKB   int       `json:"l3_cache_kb" yaml:"l3_cache_kb"`   // L3 cache size in KB
	CoreDetails []CPUCore `json:"core_details" yaml:"core_details"` // Per-core thread details
}

// HostCPUTopology represents the detected host CPU topology
type HostCPUTopology struct {
	Dies           []CPUDie `json:"dies" yaml:"dies"`                         // CPU dies
	TotalCores     int      `json:"total_cores" yaml:"total_cores"`           // Total physical cores
	TotalCPUs      int      `json:"total_cpus" yaml:"total_cpus"`             // Total logical CPUs
	ThreadsPerCore int      `json:"threads_per_core" yaml:"threads_per_core"` // Threads per core
}

// CPUTopology represents the global CPU topology configuration
type CPUTopology struct {
	Enabled        bool  `json:"enabled" yaml:"enabled"`             // Whether custom topology is enabled
	SelectedCPUs  []int `json:"selected_cpus" yaml:"selected_cpus"` // Logical CPU IDs allocated to VMs
	UseHostTopology bool `json:"use_host_topology" yaml:"use_host_topology"` // Use host topology layout (dies/sockets) vs flat
}

// VCPUPinningGlobal holds global vCPU pinning configuration.
type VCPUPinningGlobal struct {
	Enabled  bool                `json:"enabled" yaml:"enabled"`
	Mappings []VCPUToHostMapping `json:"mappings" yaml:"mappings"`
}

// VCPUToHostMapping maps a guest vCPU index to a host logical CPU ID.
type VCPUToHostMapping struct {
	VCPUID    int `json:"vcpu_id" yaml:"vcpu_id"`
	HostCPUID int `json:"host_cpu_id" yaml:"host_cpu_id"`
}

// CPUOptions represents CPU performance options
type CPUOptions struct {
	HideKVM                bool   `json:"hide_kvm" yaml:"hide_kvm"`
	VendorID               string `json:"vendor_id" yaml:"vendor_id"`
	HVFrequency            bool   `json:"hv_frequency" yaml:"hv_frequency"`
	HVRelaxed              bool   `json:"hv_relaxed" yaml:"hv_relaxed"`
	HVReset                bool   `json:"hv_reset" yaml:"hv_reset"`
	HVRuntime              bool   `json:"hv_runtime" yaml:"hv_runtime"`
	HVSpinlocks            string `json:"hv_spinlocks" yaml:"hv_spinlocks"`
	HVStimer               bool   `json:"hv_stimer" yaml:"hv_stimer"`
	HVSyncIC               bool   `json:"hv_synic" yaml:"hv_synic"`
	HVTime                 bool   `json:"hv_time" yaml:"hv_time"`
	HVVapic                bool   `json:"hv_vapic" yaml:"hv_vapic"`
	HVVPIndex              bool   `json:"hv_vpindex" yaml:"hv_vpindex"`
	HVNoNonarchCoresharing bool   `json:"hv_no_nonarch_coresharing" yaml:"hv_no_nonarch_coresharing"`
	HVTLBFlush             bool   `json:"hv_tlbflush" yaml:"hv_tlbflush"`
	HVTLBFlushExt          bool   `json:"hv_tlbflush_ext" yaml:"hv_tlbflush_ext"`
	HVIPI                  bool   `json:"hv_ipi" yaml:"hv_ipi"`
	HVAVIC                 bool   `json:"hv_avic" yaml:"hv_avic"`
	TopoExt                bool   `json:"topoext" yaml:"topoext"`
	L3Cache                bool   `json:"l3_cache" yaml:"l3_cache"`
	X2APIC                 bool   `json:"x2apic" yaml:"x2apic"`
	Migratable             bool   `json:"migratable" yaml:"migratable"`
	InvTSC                 bool   `json:"invtsc" yaml:"invtsc"`
	RTCUTC                 bool   `json:"rtc_utc" yaml:"rtc_utc"`
	CPUPM                  bool   `json:"cpu_pm" yaml:"cpu_pm"`
}

// VMStatus represents the current status of a running VM
type VMStatus struct {
	VMID       string    `json:"vm_id" yaml:"vm_id"`
	Name       string    `json:"name" yaml:"name"`
	Running    bool      `json:"running" yaml:"running"`
	PID        int       `json:"pid" yaml:"pid"`
	StartedAt  time.Time `json:"started_at" yaml:"started_at"`
	CPUThreads []int     `json:"cpu_threads" yaml:"cpu_threads"` // vCPU thread IDs
	USBDevices []string  `json:"usb_devices" yaml:"usb_devices"`
	PCIDevices []string  `json:"pci_devices" yaml:"pci_devices"`
}

// StartStopScript represents start/stop scripts configuration
type StartStopScript struct {
	UseBuiltin  bool   `json:"use_builtin" yaml:"use_builtin"` // Toggle: true=builtin, false=custom
	StartScript string `json:"start_script" yaml:"start_script"`
	StopScript  string `json:"stop_script" yaml:"stop_script"`
}
