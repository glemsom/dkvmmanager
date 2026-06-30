package vm

import (
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/domain"
)

// TestRunConfigZeroValue verifies that a zero-valued RunConfig is safe to use
func TestRunConfigZeroValue(t *testing.T) {
	var rc RunConfig

	// DryRun must be false (not a nil bool - zero value of bool is false)
	if rc.DryRun {
		t.Error("zero RunConfig should have DryRun=false")
	}

	// All slice fields must be nil (not allocated)
	if rc.PCIPassthroughConfig.Devices != nil {
		t.Error("PCIPassthroughConfig.Devices should be nil for zero RunConfig")
	}
	if rc.USBPassthroughConfig.Devices != nil {
		t.Error("USBPassthroughConfig.Devices should be nil for zero RunConfig")
	}
	if rc.CPUOptions.HideKVM {
		t.Error("CPUOptions.HideKVM should be false for zero RunConfig")
	}
	if rc.CPUTopology.Enabled {
		t.Error("CPUTopology.Enabled should be false for zero RunConfig")
	}
	if rc.VCPUPinning.Enabled {
		t.Error("VCPUPinning.Enabled should be false for zero RunConfig")
	}
	if rc.StartStopScript.UseBuiltin {
		t.Error("StartStopScript.UseBuiltin should be false for zero RunConfig")
	}
	if rc.StartStopScript.StartScript != "" {
		t.Error("StartStopScript.StartScript should be empty for zero RunConfig")
	}

	// HostCPUTopology should be zero value (empty dies)
	if rc.HostCPUTopology.Dies != nil {
		t.Error("HostCPUTopology.Dies should be nil for zero RunConfig")
	}
}

// TestRunConfigStructFields verifies all 8 exported fields exist on the struct
// by checking types via assignment. This is a compile-time check.
func TestRunConfigStructFields(t *testing.T) {
	rc := RunConfig{}

	// 1. PCIPassthroughConfig
	_ = domain.PCIPassthroughConfig(rc.PCIPassthroughConfig)

	// 2. USBPassthroughConfig
	_ = domain.USBPassthroughConfig(rc.USBPassthroughConfig)

	// 3. CPUOptions
	_ = domain.CPUOptions(rc.CPUOptions)

	// 4. CPUTopology
	_ = domain.CPUTopology(rc.CPUTopology)

	// 5. HostCPUTopology
	_ = domain.HostCPUTopology(rc.HostCPUTopology)

	// 6. VCPUPinning
	_ = domain.VCPUPinningGlobal(rc.VCPUPinning)

	// 7. StartStopScript
	_ = domain.StartStopScript(rc.StartStopScript)

	// 8. DryRun
	_ = bool(rc.DryRun)
}

// TestLoadRunConfigFromRepoEmpty verifies that LoadRunConfigFromRepo returns
// zero-valued fields when the repository has no config keys set.
func TestLoadRunConfigFromRepoEmpty(t *testing.T) {
	dir := t.TempDir()
	repo, err := NewRepository(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	rc := LoadRunConfigFromRepo(repo)

	// All fields should be zero-valued (no panic)
	if rc.DryRun {
		t.Error("expected DryRun=false with empty repo")
	}
	if rc.PCIPassthroughConfig.Devices != nil {
		t.Error("expected nil Devices with empty repo")
	}
	if rc.USBPassthroughConfig.Devices != nil {
		t.Error("expected nil USB Devices with empty repo")
	}
	if rc.CPUOptions.HideKVM {
		t.Error("expected HideKVM=false with empty repo")
	}
	if rc.CPUTopology.Enabled {
		t.Error("expected CPUTopology.Enabled=false with empty repo")
	}
	if rc.VCPUPinning.Enabled {
		t.Error("expected VCPUPinning.Enabled=false with empty repo")
	}
	if rc.StartStopScript.StartScript != "" {
		t.Error("expected empty StartScript with empty repo")
	}
}

// TestLoadRunConfigFromRepoPopulated verifies that LoadRunConfigFromRepo loads
// each config section from the correct Viper keys.
func TestLoadRunConfigFromRepoPopulated(t *testing.T) {
	dir := t.TempDir()
	repo, err := NewRepository(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	// Save configs using the same keys that LoadRunConfigFromRepo will read
	pciCfg := domain.PCIPassthroughConfig{
		Devices: []domain.PCIPassthroughDevice{
			{Address: "0000:01:00.0", Vendor: "10de", Device: "1b80", Name: "GPU"},
		},
	}
	if err := repo.SaveConfig("pci_passthrough", pciCfg); err != nil {
		t.Fatal(err)
	}

	usbCfg := domain.USBPassthroughConfig{
		Devices: []domain.USBPassthroughDevice{
			{Vendor: "046d", Product: "c52b", Name: "Unifying Receiver"},
		},
	}
	if err := repo.SaveConfig("usb_passthrough", usbCfg); err != nil {
		t.Fatal(err)
	}

	cpuOpts := domain.CPUOptions{HideKVM: true, HVRelaxed: true, VendorID: "GenuineIntel"}
	if err := repo.SaveConfig("cpu_options", cpuOpts); err != nil {
		t.Fatal(err)
	}

	cpuTopo := domain.CPUTopology{Enabled: true, SelectedCPUs: []int{0, 1, 2, 3}}
	if err := repo.SaveConfig("cpu_topology", cpuTopo); err != nil {
		t.Fatal(err)
	}

	vcpuPin := domain.VCPUPinningGlobal{
		Enabled: true,
		Mappings: []domain.VCPUToHostMapping{
			{VCPUID: 0, HostCPUID: 4},
		},
	}
	if err := repo.SaveConfig("vcpu_pinning", vcpuPin); err != nil {
		t.Fatal(err)
	}

	scriptCfg := domain.StartStopScript{UseBuiltin: false, StartScript: "echo start", StopScript: "echo stop"}
	if err := repo.SaveConfig("custom_script", scriptCfg); err != nil {
		t.Fatal(err)
	}

	// Now load
	rc := LoadRunConfigFromRepo(repo)

	// Verify each field
	if len(rc.PCIPassthroughConfig.Devices) != 1 || rc.PCIPassthroughConfig.Devices[0].Address != "0000:01:00.0" {
		t.Errorf("PCIPassthroughConfig mismatch: %+v", rc.PCIPassthroughConfig)
	}
	if len(rc.USBPassthroughConfig.Devices) != 1 || rc.USBPassthroughConfig.Devices[0].Vendor != "046d" {
		t.Errorf("USBPassthroughConfig mismatch: %+v", rc.USBPassthroughConfig)
	}
	if !rc.CPUOptions.HideKVM || !rc.CPUOptions.HVRelaxed || rc.CPUOptions.VendorID != "GenuineIntel" {
		t.Errorf("CPUOptions mismatch: %+v", rc.CPUOptions)
	}
	if !rc.CPUTopology.Enabled || len(rc.CPUTopology.SelectedCPUs) != 4 {
		t.Errorf("CPUTopology mismatch: %+v", rc.CPUTopology)
	}
	if !rc.VCPUPinning.Enabled || len(rc.VCPUPinning.Mappings) != 1 || rc.VCPUPinning.Mappings[0].HostCPUID != 4 {
		t.Errorf("VCPUPinning mismatch: %+v", rc.VCPUPinning)
	}
	if rc.StartStopScript.UseBuiltin || rc.StartStopScript.StartScript != "echo start" {
		t.Errorf("StartStopScript mismatch: %+v", rc.StartStopScript)
	}
	if rc.DryRun {
		t.Error("expected DryRun=false (not stored in config)")
	}
}
