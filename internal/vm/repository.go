// Package vm provides virtual machine management functionality
package vm

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/spf13/viper"
)

// Repository handles VM metadata storage using Viper
type Repository struct {
	vip        *viper.Viper
	configFile string
}

// NewRepository creates a new VM repository
func NewRepository(configFile string) (*Repository, error) {
	// Ensure the directory exists
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create new viper instance for VM storage
	vip := viper.New()
	vip.SetConfigFile(configFile)
	vip.SetConfigType("yaml")

	// Try to load existing config
	if err := vip.ReadInConfig(); err != nil {
		// Config doesn't exist yet, that's ok
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read VM config: %w", err)
		}
	}

	return &Repository{vip: vip, configFile: configFile}, nil
}

// save saves the configuration to file
func (r *Repository) save() error {
	dir := filepath.Dir(r.configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return r.vip.WriteConfig()
}

// === VM Persistence ===

// ListVMs returns all configured VMs
func (r *Repository) ListVMs() ([]models.VM, error) {
	var vms []models.VM

	// Get all VMs from the config
	vmsMap := r.vip.GetStringMap("vms")
	if vmsMap == nil {
		return vms, nil
	}

	for id, value := range vmsMap {
		vmData, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		vm := r.unmarshalVM(id, vmData)
		vms = append(vms, vm)
	}

	return vms, nil
}

// GetVM returns a VM by ID
func (r *Repository) GetVM(id string) (*models.VM, error) {
	vmKey := fmt.Sprintf("vms.%s", id)
	if !r.vip.IsSet(vmKey) {
		return nil, fmt.Errorf("VM not found: %s", id)
	}

	vmData := r.vip.GetStringMap(vmKey)
	vm := r.unmarshalVM(id, vmData)

	return &vm, nil
}

// SaveVM saves a VM to the config
func (r *Repository) SaveVM(vm *models.VM) error {
	if vm.ID == "" {
		return fmt.Errorf("VM ID is required")
	}

	vm.UpdatedAt = time.Now()

	// Build the VM data map
	vmData := map[string]interface{}{
		"id":           vm.ID,
		"name":         vm.Name,
		"created_at":   vm.CreatedAt.Format(time.RFC3339),
		"updated_at":   vm.UpdatedAt.Format(time.RFC3339),
		"harddisks":    vm.HardDisks,
		"cdroms":       vm.CDROMs,
		"gpu_rom":      vm.GPUROM,
		"mac":          vm.MAC,
		"network_mode": vm.NetworkMode,
		"vnc_listen":   vm.VNCListen,
		"tpm_enabled":  vm.TPMEnabled,
	}

	// Get the current vms map, update the entry, and set it back.
	// This preserves all other VMs in the config (unlike Set() for individual
	// nested keys which can corrupt Viper's internal map representation).
	vmsMap := r.vip.GetStringMap("vms")
	if vmsMap == nil {
		vmsMap = make(map[string]interface{})
	}
	vmsMap[vm.ID] = vmData
	r.vip.Set("vms", vmsMap)

	return r.save()
}

// DeleteVM deletes a VM from the config
func (r *Repository) DeleteVM(id string) error {
	vmKey := fmt.Sprintf("vms.%s", id)
	if !r.vip.IsSet(vmKey) {
		return fmt.Errorf("VM not found: %s", id)
	}

	// Get current config and manually remove the key
	vmsMap := r.vip.GetStringMap("vms")
	delete(vmsMap, id)
	r.vip.Set("vms", vmsMap)

	return r.save()
}

// FindNextAvailableID finds the next available VM ID (0-9)
func (r *Repository) FindNextAvailableID() (int, error) {
	vms, err := r.ListVMs()
	if err != nil {
		return 0, err
	}

	used := make(map[int]bool)
	for _, vm := range vms {
		var id int
		if _, err := fmt.Sscanf(vm.ID, "%d", &id); err == nil {
			used[id] = true
		}
	}

	for i := 0; i <= 9; i++ {
		if !used[i] {
			return i, nil
		}
	}

	return 0, fmt.Errorf("maximum number of VMs (10) reached")
}

// === CPU Options ===

// GetCPUOptions returns the global CPU options configuration
func (r *Repository) GetCPUOptions() (models.CPUOptions, error) {
	var opts models.CPUOptions

	if !r.vip.IsSet("cpu_options") {
		return opts, nil // Return defaults if not set
	}

	data := r.vip.GetStringMap("cpu_options")
	opts.HideKVM = getBool(data, "hide_kvm")
	opts.VendorID = getString(data, "vendor_id")
	opts.HVFrequency = getBool(data, "hv_frequency")
	opts.HVRelaxed = getBool(data, "hv_relaxed")
	opts.HVReset = getBool(data, "hv_reset")
	opts.HVRuntime = getBool(data, "hv_runtime")
	opts.HVSpinlocks = getString(data, "hv_spinlocks")
	opts.HVStimer = getBool(data, "hv_stimer")
	opts.HVSyncIC = getBool(data, "hv_synic")
	opts.HVTime = getBool(data, "hv_time")
	opts.HVVapic = getBool(data, "hv_vapic")
	opts.HVVPIndex = getBool(data, "hv_vpindex")
	opts.HVNoNonarchCoresharing = getBool(data, "hv_no_nonarch_coresharing")
	opts.HVTLBFlush = getBool(data, "hv_tlbflush")
	opts.HVTLBFlushExt = getBool(data, "hv_tlbflush_ext")
	opts.HVIPI = getBool(data, "hv_ipi")
	opts.HVAVIC = getBool(data, "hv_avic")
	opts.TopoExt = getBool(data, "topoext")
	opts.L3Cache = getBool(data, "l3_cache")
	opts.X2APIC = getBool(data, "x2apic")
	opts.Migratable = getBool(data, "migratable")
	opts.InvTSC = getBool(data, "invtsc")
	opts.RTCUTC = getBool(data, "rtc_utc")

	return opts, nil
}

// SaveCPUOptions saves the global CPU options configuration
func (r *Repository) SaveCPUOptions(opts models.CPUOptions) error {
	data := map[string]interface{}{
		"hide_kvm":                  opts.HideKVM,
		"vendor_id":                 opts.VendorID,
		"hv_frequency":              opts.HVFrequency,
		"hv_relaxed":                opts.HVRelaxed,
		"hv_reset":                  opts.HVReset,
		"hv_runtime":                opts.HVRuntime,
		"hv_spinlocks":              opts.HVSpinlocks,
		"hv_stimer":                 opts.HVStimer,
		"hv_synic":                  opts.HVSyncIC,
		"hv_time":                   opts.HVTime,
		"hv_vapic":                  opts.HVVapic,
		"hv_vpindex":                opts.HVVPIndex,
		"hv_no_nonarch_coresharing": opts.HVNoNonarchCoresharing,
		"hv_tlbflush":               opts.HVTLBFlush,
		"hv_tlbflush_ext":           opts.HVTLBFlushExt,
		"hv_ipi":                    opts.HVIPI,
		"hv_avic":                   opts.HVAVIC,
		"topoext":                   opts.TopoExt,
		"l3_cache":                  opts.L3Cache,
		"x2apic":                    opts.X2APIC,
		"migratable":                opts.Migratable,
		"invtsc":                    opts.InvTSC,
		"rtc_utc":                   opts.RTCUTC,
	}

	r.vip.Set("cpu_options", data)
	return r.save()
}

// === PCI Passthrough ===

// GetPCIPassthroughConfig returns the global PCI passthrough configuration
func (r *Repository) GetPCIPassthroughConfig() (models.PCIPassthroughConfig, error) {
	var cfg models.PCIPassthroughConfig

	if !r.vip.IsSet("pci_passthrough") {
		return cfg, nil // Return empty config if not set
	}

	data := r.vip.GetStringMap("pci_passthrough")
	devicesRaw, ok := data["devices"]
	if !ok {
		return cfg, nil
	}

	// Normalize device list: Viper may return either []interface{} (from YAML)
	// or []map[string]interface{} (from in-memory Set), so handle both.
	var deviceMaps []map[string]interface{}
	switch v := devicesRaw.(type) {
	case []interface{}:
		for _, item := range v {
			if devMap, ok := item.(map[string]interface{}); ok {
				deviceMaps = append(deviceMaps, devMap)
			}
		}
	case []map[string]interface{}:
		deviceMaps = v
	default:
		return cfg, nil
	}

	for _, devMap := range deviceMaps {
		dev := models.PCIPassthroughDevice{
			Address:   getString(devMap, "address"),
			ROMPath:   getString(devMap, "rom_path"),
			Vendor:    getString(devMap, "vendor"),
			Device:    getString(devMap, "device"),
			Name:      getString(devMap, "name"),
			ClassCode: getString(devMap, "class_code"),
		}
		cfg.Devices = append(cfg.Devices, dev)
	}

	return cfg, nil
}

// SavePCIPassthroughConfig saves the global PCI passthrough configuration
func (r *Repository) SavePCIPassthroughConfig(cfg models.PCIPassthroughConfig) error {
	devices := make([]map[string]interface{}, len(cfg.Devices))
	for i, dev := range cfg.Devices {
		devices[i] = map[string]interface{}{
			"address":    dev.Address,
			"rom_path":   dev.ROMPath,
			"vendor":     dev.Vendor,
			"device":     dev.Device,
			"name":       dev.Name,
			"class_code": dev.ClassCode,
		}
	}

	data := map[string]interface{}{
		"devices": devices,
	}

	r.vip.Set("pci_passthrough", data)
	return r.save()
}

// === USB Passthrough ===

// GetUSBPassthroughConfig returns the global USB passthrough configuration
func (r *Repository) GetUSBPassthroughConfig() (models.USBPassthroughConfig, error) {
	var cfg models.USBPassthroughConfig

	if !r.vip.IsSet("usb_passthrough") {
		return cfg, nil // Return empty config if not set
	}

	data := r.vip.GetStringMap("usb_passthrough")
	devicesRaw, ok := data["devices"]
	if !ok {
		return cfg, nil
	}

	// Normalize device list: Viper may return either []interface{} (from YAML)
	// or []map[string]interface{} (from in-memory Set), so handle both.
	var deviceMaps []map[string]interface{}
	switch v := devicesRaw.(type) {
	case []interface{}:
		for _, item := range v {
			if devMap, ok := item.(map[string]interface{}); ok {
				deviceMaps = append(deviceMaps, devMap)
			}
		}
	case []map[string]interface{}:
		deviceMaps = v
	default:
		return cfg, nil
	}

	for _, devMap := range deviceMaps {
		dev := models.USBPassthroughDevice{
			Vendor:  getString(devMap, "vendor"),
			Product: getString(devMap, "product"),
			Name:    getString(devMap, "name"),
			BusID:   getString(devMap, "bus_id"),
		}
		cfg.Devices = append(cfg.Devices, dev)
	}

	return cfg, nil
}

// SaveUSBPassthroughConfig saves the global USB passthrough configuration
func (r *Repository) SaveUSBPassthroughConfig(cfg models.USBPassthroughConfig) error {
	devices := make([]map[string]interface{}, len(cfg.Devices))
	for i, dev := range cfg.Devices {
		devices[i] = map[string]interface{}{
			"vendor":  dev.Vendor,
			"product": dev.Product,
			"name":    dev.Name,
			"bus_id":  dev.BusID,
		}
	}

	data := map[string]interface{}{
		"devices": devices,
	}

	r.vip.Set("usb_passthrough", data)
	return r.save()
}

// === CPU Topology ===

// GetCPUTopology returns the global CPU topology configuration
func (r *Repository) GetCPUTopology() (models.CPUTopology, error) {
	var topo models.CPUTopology

	if !r.vip.IsSet("cpu_topology") {
		return topo, nil
	}

	data := r.vip.GetStringMap("cpu_topology")
	topo.Enabled = getBool(data, "enabled")
	topo.SelectedCPUs = parseIntSlice(data, "selected_cpus")
	topo.UseHostTopology = getBool(data, "use_host_topology")

	return topo, nil
}

// SaveCPUTopology saves the global CPU topology configuration
func (r *Repository) SaveCPUTopology(topo models.CPUTopology) error {
	selectedCPUs := make([]interface{}, len(topo.SelectedCPUs))
	for i, v := range topo.SelectedCPUs {
		selectedCPUs[i] = v
	}

	data := map[string]interface{}{
		"enabled":           topo.Enabled,
		"selected_cpus":     selectedCPUs,
		"use_host_topology": topo.UseHostTopology,
	}

	r.vip.Set("cpu_topology", data)
	return r.save()
}

// === vCPU Pinning ===

// GetVCPUPinningGlobal returns the global vCPU pinning configuration.
func (r *Repository) GetVCPUPinningGlobal() (models.VCPUPinningGlobal, error) {
	var cfg models.VCPUPinningGlobal
	if !r.vip.IsSet("vcpu_pinning") {
		return cfg, nil
	}
	data := r.vip.GetStringMap("vcpu_pinning")
	cfg.Enabled = getBool(data, "enabled")
	mappingsData, ok := data["mappings"]
	if !ok {
		return cfg, nil
	}
	list, ok := mappingsData.([]interface{})
	if !ok {
		return cfg, nil
	}
	cfg.Mappings = make([]models.VCPUToHostMapping, 0, len(list))
	for _, item := range list {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		cfg.Mappings = append(cfg.Mappings, models.VCPUToHostMapping{
			VCPUID:    getInt(entry, "vcpu_id"),
			HostCPUID: getInt(entry, "host_cpu_id"),
		})
	}
	return cfg, nil
}

// SaveVCPUPinningGlobal saves the global vCPU pinning configuration.
func (r *Repository) SaveVCPUPinningGlobal(cfg models.VCPUPinningGlobal) error {
	mappings := make([]interface{}, 0, len(cfg.Mappings))
	for _, m := range cfg.Mappings {
		mappings = append(mappings, map[string]interface{}{
			"vcpu_id":     m.VCPUID,
			"host_cpu_id": m.HostCPUID,
		})
	}
	r.vip.Set("vcpu_pinning", map[string]interface{}{
		"enabled":  cfg.Enabled,
		"mappings": mappings,
	})
	return r.save()
}

// === Start/Stop Script ===

// GetStartStopScript returns the start/stop script configuration
func (r *Repository) GetStartStopScript() (models.StartStopScript, error) {
	var cfg models.StartStopScript

	if !r.vip.IsSet("custom_script") {
		return cfg, nil // Return defaults if not set
	}

	data := r.vip.GetStringMap("custom_script")
	cfg.UseBuiltin = getBool(data, "use_builtin")
	cfg.StartScript = getString(data, "start_script")
	cfg.StopScript = getString(data, "stop_script")

	return cfg, nil
}

// SaveStartStopScript saves the start/stop script configuration
func (r *Repository) SaveStartStopScript(cfg models.StartStopScript) error {
	data := map[string]interface{}{
		"use_builtin":  cfg.UseBuiltin,
		"start_script": cfg.StartScript,
		"stop_script":  cfg.StopScript,
	}

	r.vip.Set("custom_script", data)
	return r.save()
}

// === Helpers ===

// unmarshalVM converts a map to a VM struct
func (r *Repository) unmarshalVM(id string, data map[string]interface{}) models.VM {
	vm := models.VM{
		ID:   id,
		Name: getString(data, "name"),
	}

	// Parse times
	if createdAt, ok := data["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			vm.CreatedAt = t
		}
	}
	if updatedAt, ok := data["updated_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			vm.UpdatedAt = t
		}
	}

	// If no timestamps, set defaults
	if vm.CreatedAt.IsZero() {
		vm.CreatedAt = time.Now()
	}
	if vm.UpdatedAt.IsZero() {
		vm.UpdatedAt = time.Now()
	}

	// Parse arrays - handle both []string and []interface{} types
	vm.HardDisks = parseStringSlice(data, "harddisks")
	vm.CDROMs = parseStringSlice(data, "cdroms")

	vm.GPUROM = getString(data, "gpu_rom")
	vm.MAC = getString(data, "mac")
	vm.NetworkMode = getString(data, "network_mode")
	vm.VNCListen = getString(data, "vnc_listen")
	vm.TPMEnabled = getBool(data, "tpm_enabled")

	return vm
}

// getString safely extracts a string from a map
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

// getBool safely extracts a bool from a map
func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return false
}

// getInt safely extracts an int from a map
func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// parseStringSlice handles parsing string slices from Viper, supporting both []string and []interface{} types
func parseStringSlice(data map[string]interface{}, key string) []string {
	var result []string

	if val, ok := data[key]; ok {
		// Try []string first (Viper preserves this type when reading from YAML)
		if strSlice, ok := val.([]string); ok {
			return strSlice
		}
		// Fall back to []interface{} (older Viper behavior)
		if ifaceSlice, ok := val.([]interface{}); ok {
			for _, item := range ifaceSlice {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
		}
	}

	return result
}

// parseIntSlice handles parsing int slices from Viper, supporting both []int and []interface{} types
func parseIntSlice(data map[string]interface{}, key string) []int {
	var result []int

	if val, ok := data[key]; ok {
		// Try []int first
		if intSlice, ok := val.([]int); ok {
			return intSlice
		}
		// Try []interface{} (from YAML)
		if ifaceSlice, ok := val.([]interface{}); ok {
			for _, item := range ifaceSlice {
				switch v := item.(type) {
				case int:
					result = append(result, v)
				case int64:
					result = append(result, int(v))
				case float64:
					result = append(result, int(v))
				}
			}
		}
	}

	return result
}
