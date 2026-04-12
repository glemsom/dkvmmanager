package vm

import (
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
)

// TestUSBPassthroughConfigPersistence tests that USB passthrough config
// can be saved and loaded correctly
func TestUSBPassthroughConfigPersistence(t *testing.T) {
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
	loadedCfg, err := mgr.GetUSBPassthroughConfig()
	if err != nil {
		t.Fatalf("GetUSBPassthroughConfig error: %v", err)
	}
	if len(loadedCfg.Devices) != 0 {
		t.Errorf("Expected 0 devices initially, got %d", len(loadedCfg.Devices))
	}

	// Save a config with one device
	saveCfg := models.USBPassthroughConfig{
		Devices: []models.USBPassthroughDevice{
			{
				Vendor:  "046d",
				Product: "c52b",
				Name:    "Logitech Unifying Receiver",
				BusID:   "1-1",
			},
		},
	}

	err = mgr.SaveUSBPassthroughConfig(saveCfg)
	if err != nil {
		t.Fatalf("SaveUSBPassthroughConfig error: %v", err)
	}

	// Load and verify
	loadedCfg, err = mgr.GetUSBPassthroughConfig()
	if err != nil {
		t.Fatalf("GetUSBPassthroughConfig error: %v", err)
	}

	if len(loadedCfg.Devices) != 1 {
		t.Fatalf("Expected 1 device, got %d", len(loadedCfg.Devices))
	}

	dev := loadedCfg.Devices[0]
	if dev.Vendor != "046d" {
		t.Errorf("Vendor = %s, want 046d", dev.Vendor)
	}
	if dev.Product != "c52b" {
		t.Errorf("Product = %s, want c52b", dev.Product)
	}
	if dev.Name != "Logitech Unifying Receiver" {
		t.Errorf("Name = %s, want Logitech Unifying Receiver", dev.Name)
	}
	if dev.BusID != "1-1" {
		t.Errorf("BusID = %s, want 1-1", dev.BusID)
	}
}

// TestUSBPassthroughConfigMultipleDevices tests saving multiple USB devices
func TestUSBPassthroughConfigMultipleDevices(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	saveCfg := models.USBPassthroughConfig{
		Devices: []models.USBPassthroughDevice{
			{
				Vendor:  "046d",
				Product: "c52b",
				Name:    "Logitech Unifying Receiver",
				BusID:   "1-1",
			},
			{
				Vendor:  "045e",
				Product: "028e",
				Name:    "Microsoft Corp. Xbox360 Controller",
				BusID:   "3-2",
			},
		},
	}

	err = mgr.SaveUSBPassthroughConfig(saveCfg)
	if err != nil {
		t.Fatalf("SaveUSBPassthroughConfig error: %v", err)
	}

	loadedCfg, err := mgr.GetUSBPassthroughConfig()
	if err != nil {
		t.Fatalf("GetUSBPassthroughConfig error: %v", err)
	}

	if len(loadedCfg.Devices) != 2 {
		t.Fatalf("Expected 2 devices, got %d", len(loadedCfg.Devices))
	}

	// Verify both devices
	if loadedCfg.Devices[0].Vendor != "046d" {
		t.Errorf("Device 0 vendor = %s, want 046d", loadedCfg.Devices[0].Vendor)
	}
	if loadedCfg.Devices[1].Vendor != "045e" {
		t.Errorf("Device 1 vendor = %s, want 045e", loadedCfg.Devices[1].Vendor)
	}
}

// TestUSBPassthroughConfigOverwrite tests overwriting existing USB config
func TestUSBPassthroughConfigOverwrite(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		DataFolder:    tmpDir,
		VMsConfigFile: filepath.Join(tmpDir, "dkvmmanager", "vms.yaml"),
	}

	mgr, err := NewManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Save initial config
	saveCfg := models.USBPassthroughConfig{
		Devices: []models.USBPassthroughDevice{
			{Vendor: "046d", Product: "c52b", Name: "Logitech", BusID: "1-1"},
		},
	}
	if err := mgr.SaveUSBPassthroughConfig(saveCfg); err != nil {
		t.Fatalf("SaveUSBPassthroughConfig error: %v", err)
	}

	// Overwrite with different config
	newCfg := models.USBPassthroughConfig{
		Devices: []models.USBPassthroughDevice{
			{Vendor: "045e", Product: "028e", Name: "Xbox Controller", BusID: "3-2"},
		},
	}
	if err := mgr.SaveUSBPassthroughConfig(newCfg); err != nil {
		t.Fatalf("SaveUSBPassthroughConfig error: %v", err)
	}

	loadedCfg, err := mgr.GetUSBPassthroughConfig()
	if err != nil {
		t.Fatalf("GetUSBPassthroughConfig error: %v", err)
	}

	if len(loadedCfg.Devices) != 1 {
		t.Fatalf("Expected 1 device after overwrite, got %d", len(loadedCfg.Devices))
	}

	if loadedCfg.Devices[0].Vendor != "045e" {
		t.Errorf("Vendor = %s, want 045e after overwrite", loadedCfg.Devices[0].Vendor)
	}
}
