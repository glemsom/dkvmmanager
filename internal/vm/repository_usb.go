package vm

import (
	"github.com/glemsom/dkvmmanager/internal/models"
)

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
