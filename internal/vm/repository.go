// Package vm provides virtual machine management functionality
package vm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/go-viper/mapstructure/v2"
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

// === Generic Config Store ===

// GetConfig decodes a config section identified by key into dest.
// If the key is not set in the config, dest is left at its zero value (no error).
func (r *Repository) GetConfig(key string, dest interface{}) error {
	raw := r.vip.Get(key)
	if raw == nil {
		return nil
	}
	// Use json tag name so struct tags (json:"...") match keys produced by SaveConfig
	// which uses JSON marshal/unmarshal for struct→map conversion.
	config := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  dest,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("create decoder for %q: %w", key, err)
	}
	return decoder.Decode(raw)
}

// SaveConfig encodes src and stores it under key, then persists to disk.
func (r *Repository) SaveConfig(key string, src interface{}) error {
	// Use JSON marshal/unmarshal to convert struct to map[string]interface{}
	// since mapstructure/v2 does not provide an Encode function.
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("marshal config for %q: %w", key, err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("unmarshal config for %q: %w", key, err)
	}
	r.vip.Set(key, cfg)
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
