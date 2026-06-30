package vm

import (
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/domain"
)

// TestConfigStore_RoundTrip verifies each config type survives a save-and-reload cycle
// through the generic GetConfig/SaveConfig methods.
func TestConfigStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	repo, err := NewRepository(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("CPUOptions", func(t *testing.T) {
		in := domain.CPUOptions{HideKVM: true, HVRelaxed: true, VendorID: "GenuineIntel"}
		if err := repo.SaveConfig("cpu_options", in); err != nil {
			t.Fatal(err)
		}
		var out domain.CPUOptions
		if err := repo.GetConfig("cpu_options", &out); err != nil {
			t.Fatal(err)
		}
		if out.HideKVM != in.HideKVM || out.VendorID != in.VendorID {
			t.Errorf("round-trip mismatch: %+v vs %+v", in, out)
		}
	})

	t.Run("PCIPassthroughConfig", func(t *testing.T) {
		in := domain.PCIPassthroughConfig{
			Devices: []domain.PCIPassthroughDevice{
				{Address: "0000:01:00.0", Vendor: "10de", Device: "1b80", Name: "GPU"},
			},
		}
		if err := repo.SaveConfig("pci_passthrough", in); err != nil {
			t.Fatal(err)
		}
		var out domain.PCIPassthroughConfig
		if err := repo.GetConfig("pci_passthrough", &out); err != nil {
			t.Fatal(err)
		}
		if len(out.Devices) != 1 || out.Devices[0].Address != "0000:01:00.0" {
			t.Errorf("round-trip mismatch: %+v", out)
		}
	})

	t.Run("USBPassthroughConfig", func(t *testing.T) {
		in := domain.USBPassthroughConfig{
			Devices: []domain.USBPassthroughDevice{
				{Vendor: "046d", Product: "c52b", Name: "Unifying Receiver"},
			},
		}
		if err := repo.SaveConfig("usb_passthrough", in); err != nil {
			t.Fatal(err)
		}
		var out domain.USBPassthroughConfig
		if err := repo.GetConfig("usb_passthrough", &out); err != nil {
			t.Fatal(err)
		}
		if len(out.Devices) != 1 || out.Devices[0].Vendor != "046d" {
			t.Errorf("round-trip mismatch: %+v", out)
		}
	})

	t.Run("CPUTopology", func(t *testing.T) {
		in := domain.CPUTopology{Enabled: true, SelectedCPUs: []int{0, 1, 2, 3}}
		if err := repo.SaveConfig("cpu_topology", in); err != nil {
			t.Fatal(err)
		}
		var out domain.CPUTopology
		if err := repo.GetConfig("cpu_topology", &out); err != nil {
			t.Fatal(err)
		}
		if !out.Enabled || len(out.SelectedCPUs) != 4 {
			t.Errorf("round-trip mismatch: %+v", out)
		}
	})

	t.Run("VCPUPinningGlobal", func(t *testing.T) {
		in := domain.VCPUPinningGlobal{
			Enabled: true,
			Mappings: []domain.VCPUToHostMapping{
				{VCPUID: 0, HostCPUID: 4},
			},
		}
		if err := repo.SaveConfig("vcpu_pinning", in); err != nil {
			t.Fatal(err)
		}
		var out domain.VCPUPinningGlobal
		if err := repo.GetConfig("vcpu_pinning", &out); err != nil {
			t.Fatal(err)
		}
		if !out.Enabled || len(out.Mappings) != 1 || out.Mappings[0].HostCPUID != 4 {
			t.Errorf("round-trip mismatch: %+v", out)
		}
	})

	t.Run("StartStopScript", func(t *testing.T) {
		in := domain.StartStopScript{UseBuiltin: false, StartScript: "echo start", StopScript: "echo stop"}
		if err := repo.SaveConfig("custom_script", in); err != nil {
			t.Fatal(err)
		}
		var out domain.StartStopScript
		if err := repo.GetConfig("custom_script", &out); err != nil {
			t.Fatal(err)
		}
		if out.UseBuiltin || out.StartScript != "echo start" {
			t.Errorf("round-trip mismatch: %+v", out)
		}
	})

	t.Run("GetUnsetKeyReturnsZeroValue", func(t *testing.T) {
		var opts domain.CPUOptions
		if err := repo.GetConfig("nonexistent_key", &opts); err != nil {
			t.Fatal(err)
		}
		// Should be zero value, no error
	})

	t.Run("OverwriteExistingKey", func(t *testing.T) {
		in1 := domain.CPUOptions{HideKVM: true}
		in2 := domain.CPUOptions{HVRelaxed: true}
		if err := repo.SaveConfig("cpu_options", in1); err != nil {
			t.Fatal(err)
		}
		if err := repo.SaveConfig("cpu_options", in2); err != nil {
			t.Fatal(err)
		}
		var out domain.CPUOptions
		if err := repo.GetConfig("cpu_options", &out); err != nil {
			t.Fatal(err)
		}
		if out.HideKVM || !out.HVRelaxed {
			t.Errorf("overwrite failed: %+v", out)
		}
	})
}
