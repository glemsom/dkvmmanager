package vm

import (
	"github.com/glemsom/dkvmmanager/internal/models"
)

// GetCPUTopology returns the global CPU topology configuration
func (r *Repository) GetCPUTopology() (models.CPUTopology, error) {
	var topo models.CPUTopology

	if !r.vip.IsSet("cpu_topology") {
		return topo, nil
	}

	data := r.vip.GetStringMap("cpu_topology")
	topo.Enabled = getBool(data, "enabled")
	topo.SelectedCPUs = parseIntSlice(data, "selected_cpus")

	return topo, nil
}

// SaveCPUTopology saves the global CPU topology configuration
func (r *Repository) SaveCPUTopology(topo models.CPUTopology) error {
	selectedCPUs := make([]interface{}, len(topo.SelectedCPUs))
	for i, v := range topo.SelectedCPUs {
		selectedCPUs[i] = v
	}

	data := map[string]interface{}{
		"enabled":       topo.Enabled,
		"selected_cpus": selectedCPUs,
	}

	r.vip.Set("cpu_topology", data)
	return r.save()
}
