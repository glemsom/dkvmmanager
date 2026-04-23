package vm

import (
	"time"

	"github.com/glemsom/dkvmmanager/internal/models"
)

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
