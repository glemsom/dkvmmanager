package vm

import "github.com/glemsom/dkvmmanager/internal/models"

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
