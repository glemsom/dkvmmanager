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

// save saves the configuration to file
func (r *Repository) save() error {
	dir := filepath.Dir(r.configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return r.vip.WriteConfig()
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
