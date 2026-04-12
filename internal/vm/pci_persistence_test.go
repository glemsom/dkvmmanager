package vm

import (
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
)

// TestPCIPassthroughConfigPersistence tests that PCI passthrough config
// can be saved and loaded correctly
func TestPCIPassthroughConfigPersistence(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Initially, config should be empty
	loadedCfg, err := mgr.GetPCIPassthroughConfig()
	if err != nil {
		t.Fatalf("GetPCIPassthroughConfig error: %v", err)
	}
	if len(loadedCfg.Devices) != 0 {
		t.Errorf("Expected 0 devices initially, got %d", len(loadedCfg.Devices))
	}

	// Save a config with one device
	saveCfg := models.PCIPassthroughConfig{
		Devices: []models.PCIPassthroughDevice{
			{
				Address:   "0000:01:00.0",
				ROMPath:   "/roms/gpu.rom",
				Vendor:    "10de",
				Device:    "1b80",
				Name:      "NVIDIA GeForce GTX 1080",
				ClassCode: "0300",
			},
		},
	}

	err = mgr.SavePCIPassthroughConfig(saveCfg)
	if err != nil {
		t.Fatalf("SavePCIPassthroughConfig error: %v", err)
	}

	// Load it back
	loadedCfg, err = mgr.GetPCIPassthroughConfig()
	if err != nil {
		t.Fatalf("GetPCIPassthroughConfig after save error: %v", err)
	}

	if len(loadedCfg.Devices) != 1 {
		t.Fatalf("Expected 1 device after save, got %d", len(loadedCfg.Devices))
	}

	dev := loadedCfg.Devices[0]
	if dev.Address != "0000:01:00.0" {
		t.Errorf("Expected Address '0000:01:00.0', got '%s'", dev.Address)
	}
	if dev.ROMPath != "/roms/gpu.rom" {
		t.Errorf("Expected ROMPath '/roms/gpu.rom', got '%s'", dev.ROMPath)
	}
	if dev.Vendor != "10de" {
		t.Errorf("Expected Vendor '10de', got '%s'", dev.Vendor)
	}
	if dev.Device != "1b80" {
		t.Errorf("Expected Device '1b80', got '%s'", dev.Device)
	}
	if dev.Name != "NVIDIA GeForce GTX 1080" {
		t.Errorf("Expected Name 'NVIDIA GeForce GTX 1080', got '%s'", dev.Name)
	}
	if dev.ClassCode != "0300" {
		t.Errorf("Expected ClassCode '0300', got '%s'", dev.ClassCode)
	}
}

// TestPCIPassthroughConfigMultipleDevices tests saving/loading multiple devices
func TestPCIPassthroughConfigMultipleDevices(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Save config with multiple devices
	saveCfg := models.PCIPassthroughConfig{
		Devices: []models.PCIPassthroughDevice{
			{
				Address:   "0000:01:00.0",
				Vendor:    "10de",
				Device:    "1b80",
				Name:      "NVIDIA GeForce GTX 1080",
				ClassCode: "0300",
			},
			{
				Address:   "0000:02:00.0",
				ROMPath:   "/roms/usb.rom",
				Vendor:    "1b73",
				Device:    "1100",
				Name:      "Fresco Logic USB 3.0",
				ClassCode: "0c03",
			},
		},
	}

	err = mgr.SavePCIPassthroughConfig(saveCfg)
	if err != nil {
		t.Fatalf("SavePCIPassthroughConfig error: %v", err)
	}

	loadedCfg, err := mgr.GetPCIPassthroughConfig()
	if err != nil {
		t.Fatalf("GetPCIPassthroughConfig error: %v", err)
	}

	if len(loadedCfg.Devices) != 2 {
		t.Fatalf("Expected 2 devices, got %d", len(loadedCfg.Devices))
	}

	// Verify first device
	if loadedCfg.Devices[0].Address != "0000:01:00.0" {
		t.Errorf("Device 0: expected Address '0000:01:00.0', got '%s'", loadedCfg.Devices[0].Address)
	}
	if loadedCfg.Devices[0].ROMPath != "" {
		t.Errorf("Device 0: expected empty ROMPath, got '%s'", loadedCfg.Devices[0].ROMPath)
	}

	// Verify second device
	if loadedCfg.Devices[1].Address != "0000:02:00.0" {
		t.Errorf("Device 1: expected Address '0000:02:00.0', got '%s'", loadedCfg.Devices[1].Address)
	}
	if loadedCfg.Devices[1].ROMPath != "/roms/usb.rom" {
		t.Errorf("Device 1: expected ROMPath '/roms/usb.rom', got '%s'", loadedCfg.Devices[1].ROMPath)
	}
}

// TestPCIPassthroughConfigOverwrite tests that saving a new config replaces the old one
func TestPCIPassthroughConfigOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Save first config
	cfg1 := models.PCIPassthroughConfig{
		Devices: []models.PCIPassthroughDevice{
			{Address: "0000:01:00.0", Vendor: "10de", Device: "1b80", Name: "GPU 1", ClassCode: "0300"},
		},
	}
	if err := mgr.SavePCIPassthroughConfig(cfg1); err != nil {
		t.Fatalf("First save error: %v", err)
	}

	// Save second config (overwrite)
	cfg2 := models.PCIPassthroughConfig{
		Devices: []models.PCIPassthroughDevice{
			{Address: "0000:02:00.0", Vendor: "1002", Device: "67df", Name: "GPU 2", ClassCode: "0300"},
		},
	}
	if err := mgr.SavePCIPassthroughConfig(cfg2); err != nil {
		t.Fatalf("Second save error: %v", err)
	}

	// Should have the second config, not the first
	loadedCfg, err := mgr.GetPCIPassthroughConfig()
	if err != nil {
		t.Fatalf("GetPCIPassthroughConfig error: %v", err)
	}

	if len(loadedCfg.Devices) != 1 {
		t.Fatalf("Expected 1 device after overwrite, got %d", len(loadedCfg.Devices))
	}

	if loadedCfg.Devices[0].Address != "0000:02:00.0" {
		t.Errorf("Expected Address '0000:02:00.0' after overwrite, got '%s'", loadedCfg.Devices[0].Address)
	}
}

// TestPCIPassthroughConfigEmptySave tests saving an empty config (deselecting all)
func TestPCIPassthroughConfigEmptySave(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// First save a config with a device
	saveCfg := models.PCIPassthroughConfig{
		Devices: []models.PCIPassthroughDevice{
			{Address: "0000:01:00.0", Vendor: "10de", Device: "1b80", Name: "GPU", ClassCode: "0300"},
		},
	}
	if err := mgr.SavePCIPassthroughConfig(saveCfg); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Now save empty config
	emptyCfg := models.PCIPassthroughConfig{Devices: []models.PCIPassthroughDevice{}}
	if err := mgr.SavePCIPassthroughConfig(emptyCfg); err != nil {
		t.Fatalf("Empty save error: %v", err)
	}

	loadedCfg, err := mgr.GetPCIPassthroughConfig()
	if err != nil {
		t.Fatalf("GetPCIPassthroughConfig error: %v", err)
	}

	if len(loadedCfg.Devices) != 0 {
		t.Errorf("Expected 0 devices after empty save, got %d", len(loadedCfg.Devices))
	}
}
