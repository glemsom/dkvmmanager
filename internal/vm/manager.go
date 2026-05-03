// Package vm provides virtual machine management functionality
package vm

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
)

// Manager handles virtual machine operations
type Manager struct {
	cfg        *config.Config
	repository *Repository
}

// NewManager creates a new VM manager
func NewManager(cfg *config.Config) (*Manager, error) {
	repo, err := NewRepository(cfg.VMsConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM repository: %w", err)
	}

	return &Manager{
		cfg:        cfg,
		repository: repo,
	}, nil
}

// GetConfig returns the manager configuration
func (m *Manager) GetConfig() *config.Config {
	return m.cfg
}

// ListVMs returns all configured VMs
func (m *Manager) ListVMs() ([]models.VM, error) {
	return m.repository.ListVMs()
}

// GetVM returns a VM by ID
func (m *Manager) GetVM(id string) (*models.VM, error) {
	return m.repository.GetVM(id)
}

// CreateVM creates a new VM
func (m *Manager) CreateVM(name string) (*models.VM, error) {
	// Find next available VM ID
	vmID, err := m.repository.FindNextAvailableID()
	if err != nil {
		return nil, err
	}

	// Create VM directory in vms folder
	vmDir := m.GetVMDataPath(vmID)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		return nil, err
	}

	// Copy OVMF files to VM directory
	if err := m.copyOVMFFiles(vmDir); err != nil {
		return nil, fmt.Errorf("failed to copy OVMF files: %w", err)
	}

	// Create VM config
	vm := models.VM{
		ID:        fmt.Sprintf("%d", vmID),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		MAC:       generateMAC(),
	}

	if err := m.repository.SaveVM(&vm); err != nil {
		return nil, err
	}

	return &vm, nil
}

// CreateVMWithMAC creates a new VM with a specific MAC address (for testing)
func (m *Manager) CreateVMWithMAC(name, mac string) (*models.VM, error) {
	// Find next available VM ID
	vmID, err := m.repository.FindNextAvailableID()
	if err != nil {
		return nil, err
	}

	// Create VM directory in vms folder
	vmDir := m.GetVMDataPath(vmID)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		return nil, err
	}

	// Copy OVMF files to VM directory
	if err := m.copyOVMFFiles(vmDir); err != nil {
		return nil, fmt.Errorf("failed to copy OVMF files: %w", err)
	}

	// Create VM config with specified MAC
	vm := models.VM{
		ID:        fmt.Sprintf("%d", vmID),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		MAC:       mac,
	}

	if err := m.repository.SaveVM(&vm); err != nil {
		return nil, err
	}

	return &vm, nil
}

// SaveVM saves a VM configuration
func (m *Manager) SaveVM(vm *models.VM) error {
	return m.repository.SaveVM(vm)
}

// DeleteVM deletes a VM by ID (metadata only)
func (m *Manager) DeleteVM(id string) error {
	// Delete from repository (metadata only)
	return m.repository.DeleteVM(id)
}

// GetVMDataPath returns the data folder path for a VM by numeric ID
func (m *Manager) GetVMDataPath(vmID int) string {
	return filepath.Join(m.cfg.DataFolder, "vms", fmt.Sprintf("%d", vmID))
}

// GetVMDataPathByID returns the data folder path for a VM by string ID
func (m *Manager) GetVMDataPathByID(id string) string {
	return filepath.Join(m.cfg.DataFolder, "vms", id)
}

// GenerateMAC generates a random MAC address
func (m *Manager) GenerateMAC() string {
	return generateMAC()
}

// GetCPUOptions returns the global CPU options configuration
func (m *Manager) GetCPUOptions() (models.CPUOptions, error) {
	return m.repository.GetCPUOptions()
}

// SaveCPUOptions saves the global CPU options configuration
func (m *Manager) SaveCPUOptions(opts models.CPUOptions) error {
	return m.repository.SaveCPUOptions(opts)
}

// GetPCIPassthroughConfig returns the global PCI passthrough configuration
func (m *Manager) GetPCIPassthroughConfig() (models.PCIPassthroughConfig, error) {
	return m.repository.GetPCIPassthroughConfig()
}

// SavePCIPassthroughConfig saves the global PCI passthrough configuration
func (m *Manager) SavePCIPassthroughConfig(cfg models.PCIPassthroughConfig) error {
	return m.repository.SavePCIPassthroughConfig(cfg)
}

// ScanPCIDevices scans the host for available PCI devices
func (m *Manager) ScanPCIDevices() ([]models.PCIDevice, error) {
	scanner := NewPCIScanner()
	return scanner.ScanDevices()
}

// GetUSBPassthroughConfig returns the global USB passthrough configuration
func (m *Manager) GetUSBPassthroughConfig() (models.USBPassthroughConfig, error) {
	return m.repository.GetUSBPassthroughConfig()
}

// SaveUSBPassthroughConfig saves the global USB passthrough configuration
func (m *Manager) SaveUSBPassthroughConfig(cfg models.USBPassthroughConfig) error {
	return m.repository.SaveUSBPassthroughConfig(cfg)
}

// ScanUSBDevices scans the host for available USB devices
func (m *Manager) ScanUSBDevices() ([]models.USBDevice, error) {
	scanner := NewUSBScanner()
	return scanner.ScanDevices()
}

// ScanCPUTopology scans the host for CPU topology information
func (m *Manager) ScanCPUTopology() (models.HostCPUTopology, error) {
	scanner := NewCPUScanner()
	return scanner.ScanTopology()
}

// GetCPUTopology returns the global CPU topology configuration
func (m *Manager) GetCPUTopology() (models.CPUTopology, error) {
	return m.repository.GetCPUTopology()
}

// SaveCPUTopology saves the global CPU topology configuration
func (m *Manager) SaveCPUTopology(topo models.CPUTopology) error {
	return m.repository.SaveCPUTopology(topo)
}

// GetVCPUPinningGlobal returns global vCPU pinning configuration.
func (m *Manager) GetVCPUPinningGlobal() (models.VCPUPinningGlobal, error) {
	return m.repository.GetVCPUPinningGlobal()
}

// SaveVCPUPinningGlobal saves global vCPU pinning configuration.
func (m *Manager) SaveVCPUPinningGlobal(p models.VCPUPinningGlobal) error {
	return m.repository.SaveVCPUPinningGlobal(p)
}

// GetStartStopScript returns the start/stop script configuration
func (m *Manager) GetStartStopScript() (models.StartStopScript, error) {
	return m.repository.GetStartStopScript()
}

// SaveStartStopScript saves the start/stop script configuration
func (m *Manager) SaveStartStopScript(cfg models.StartStopScript) error {
	return m.repository.SaveStartStopScript(cfg)
}

// ApplyVFIOIDsToKernel reads the current PCI passthrough config, builds the
// vfio-pci.ids parameter string, and writes it to the grub.cfg file.
// It returns an error if the config cannot be read, or the grub.cfg cannot be updated.
func (m *Manager) ApplyVFIOIDsToKernel() error {
	if debugMode {
		log.Println("[DEBUG] ApplyVFIOIDsToKernel: starting")
	}

	// Get current PCI passthrough config
	pciCfg, err := m.GetPCIPassthroughConfig()
	if err != nil {
		if debugMode {
			log.Printf("[DEBUG] ApplyVFIOIDsToKernel: failed to get PCI config: %v", err)
		}
		return fmt.Errorf("get PCI passthrough config: %w", err)
	}

	if debugMode {
		log.Printf("[DEBUG] ApplyVFIOIDsToKernel: got %d PCI devices from config", len(pciCfg.Devices))
		for i, dev := range pciCfg.Devices {
			log.Printf("[DEBUG] ApplyVFIOIDsToKernel: device[%d] = %s:%s (%s)", i, dev.Vendor, dev.Device, dev.Address)
		}
	}

	// Build vfio-pci.ids string
	vfioIDs := BuildVFIOIDs(pciCfg.Devices)

	if debugMode {
		log.Printf("[DEBUG] ApplyVFIOIDsToKernel: built vfioIDs = %q", vfioIDs)
	}

	// Get grub config path
	grubPath := m.cfg.GrubConfigPath
	if grubPath == "" {
		grubPath = "/media/usb/boot/grub/grub.cfg"
	}

	if debugMode {
		log.Printf("[DEBUG] ApplyVFIOIDsToKernel: grubPath = %s", grubPath)
	}

	// Update grub.cfg
	if err := UpdateGrubVFIOIDs(vfioIDs, grubPath); err != nil {
		if debugMode {
			log.Printf("[DEBUG] ApplyVFIOIDsToKernel: UpdateGrubVFIOIDs failed: %v", err)
		}
		return fmt.Errorf("update grub.cfg: %w", err)
	}

	if debugMode {
		log.Println("[DEBUG] ApplyVFIOIDsToKernel: completed successfully")
	}
	return nil
}

// copyOVMFFiles copies OVMF_CODE.fd and OVMF_VARS.fd from the configured
// BIOS paths to the VM directory.
func (m *Manager) copyOVMFFiles(vmDir string) error {
	// Copy BIOS code file
	if err := copyFile(m.cfg.BIOSCode, filepath.Join(vmDir, "OVMF_CODE.fd")); err != nil {
		return fmt.Errorf("failed to copy OVMF_CODE.fd: %w", err)
	}

	// Copy BIOS vars file
	if err := copyFile(m.cfg.BIOSVars, filepath.Join(vmDir, "OVMF_VARS.fd")); err != nil {
		return fmt.Errorf("failed to copy OVMF_VARS.fd: %w", err)
	}

	return nil
}

// copyFile copies a single file from src to dst. It returns nil if the
// source file doesn't exist (allows tests to work without real BIOS files).
func copyFile(src, dst string) error {
	// Check if source exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // Source doesn't exist, skip copy
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()


	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()


	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy content from %s to %s: %w", src, dst, err)
	}

	// Sync to ensure data is written
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file %s: %w", dst, err)
	}

	return nil
}

func generateMAC() string {
	// Generate a random MAC address with the local bit set
	// Using a random locally-administered MAC address
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to static MAC if random fails
		return "de:ad:be:ef:00:00"
	}
	// Set local bit (bit 1) to indicate locally administered MAC
	bytes[0] = bytes[0]&0xFE | 0x02
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5])
}
