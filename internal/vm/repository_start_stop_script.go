package vm

import (
	"github.com/glemsom/dkvmmanager/internal/models"
)

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
