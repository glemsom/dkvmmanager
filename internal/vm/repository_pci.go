package vm

import (
	"github.com/glemsom/dkvmmanager/internal/models"
)

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
