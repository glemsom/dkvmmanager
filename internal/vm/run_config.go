package vm

import (
	"github.com/glemsom/dkvmmanager/internal/models"
)

// RunConfig aggregates all optional configuration for VMRunner into a single value.
// A zero-valued RunConfig is safe to use (no panics, all fields at zero value).
type RunConfig struct {
	PCIPassthroughConfig models.PCIPassthroughConfig
	USBPassthroughConfig models.USBPassthroughConfig
	CPUOptions           models.CPUOptions
	CPUTopology          models.CPUTopology
	HostCPUTopology      models.HostCPUTopology
	VCPUPinning          models.VCPUPinningGlobal
	StartStopScript      models.StartStopScript
	DryRun               bool
}

// LoadRunConfigFromRepo loads each config section from the Repository into a RunConfig.
// Config keys used:
//   - "pci_passthrough"  → PCIPassthroughConfig
//   - "usb_passthrough"  → USBPassthroughConfig
//   - "cpu_options"      → CPUOptions
//   - "cpu_topology"     → CPUTopology
//   - "vcpu_pinning"     → VCPUPinningGlobal
//   - "custom_script"    → StartStopScript
//
// HostCPUTopology is populated via DefaultHostDiscovery{}.ScanCPUTopology().
// Empty or missing keys produce zero-valued fields and do not return an error.
func LoadRunConfigFromRepo(repo *Repository) RunConfig {
	var rc RunConfig

	// Load each config section from the repository.
	// Missing keys leave the field at zero value (no error).
	_ = repo.GetConfig("pci_passthrough", &rc.PCIPassthroughConfig)
	_ = repo.GetConfig("usb_passthrough", &rc.USBPassthroughConfig)
	_ = repo.GetConfig("cpu_options", &rc.CPUOptions)
	_ = repo.GetConfig("cpu_topology", &rc.CPUTopology)
	_ = repo.GetConfig("vcpu_pinning", &rc.VCPUPinning)
	_ = repo.GetConfig("custom_script", &rc.StartStopScript)

	// HostCPUTopology is not persisted; it is always discovered from the host.
	topo, err := (&DefaultHostDiscovery{}).ScanCPUTopology()
	if err == nil {
		rc.HostCPUTopology = topo
	}

	// DryRun defaults to false and is not loaded from config.
	return rc
}
