// Package vm provides virtual machine management functionality
package vm

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
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

	return NewManagerWithRepository(cfg, repo)
}

// NewManagerWithRepository creates a Manager with an existing repository.
func NewManagerWithRepository(cfg *config.Config, repo *Repository) (*Manager, error) {
	return &Manager{
		cfg:        cfg,
		repository: repo,
	}, nil
}

// Repository returns the underlying VM repository for config persistence.
func (m *Manager) Repository() *Repository {
	return m.repository
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

// ApplyVFIOIDsToKernel reads the current PCI passthrough config, builds the
// vfio-pci.ids parameter string, and writes it to the grub.cfg file.
// It returns an error if the config cannot be read, or the grub.cfg cannot be updated.
func (m *Manager) ApplyVFIOIDsToKernel() error {
	if debugMode {
		log.Println("[DEBUG] ApplyVFIOIDsToKernel: starting")
	}

	// Get current PCI passthrough config
	pciCfg, err := m.repository.GetPCIPassthroughConfig()
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

	// Remount /media/usb as rw before modifying grub.cfg.
	// DKVM Hypervisor keeps the USB filesystem read-only by default.
	remountPath := detectMountPath(grubPath)
	if remountPath != "" {
		if err := remountFilesystem(remountPath, "rw"); err != nil {
			return fmt.Errorf("remount %s as rw: %w", remountPath, err)
		}
		// Always restore read-only after the update, regardless of outcome.
		defer func() {
			if err := remountFilesystem(remountPath, "ro"); err != nil {
				log.Printf("[WARN] ApplyVFIOIDsToKernel: failed to remount %s as ro: %v", remountPath, err)
			}
		}()
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

// ApplyCPUParamsToKernel reads the current vCPU pinning config, builds CPU isolation
// parameter strings (isolcpus, nohz_full, rcu_nocbs), and writes them to grub.cfg.
// It returns an error if the config cannot be read, or the grub.cfg cannot be updated.
func (m *Manager) ApplyCPUParamsToKernel() error {
	if debugMode {
		log.Println("[DEBUG] ApplyCPUParamsToKernel: starting")
	}

	// Get current vCPU pinning config
	pinning, err := m.repository.GetVCPUPinningGlobal()
	if err != nil {
		if debugMode {
			log.Printf("[DEBUG] ApplyCPUParamsToKernel: failed to get vCPU pinning config: %v", err)
		}
		return fmt.Errorf("get vCPU pinning config: %w", err)
	}

	if debugMode {
		log.Printf("[DEBUG] ApplyCPUParamsToKernel: pinning enabled=%v, mappings=%d",
			pinning.Enabled, len(pinning.Mappings))
	}

	// Build sorted, deduplicated host CPU list from mappings
	cpuList := buildHostCPUList(pinning.Mappings)

	// Build parameter values
	var isolcpus, nohzFull, rcuNoCBS string
	if cpuList != "" {
		isolcpus = "domain,managed_irq," + cpuList
		nohzFull = cpuList
		rcuNoCBS = cpuList
	}

	if debugMode {
		log.Printf("[DEBUG] ApplyCPUParamsToKernel: cpuList=%q, isolcpus=%q, nohzFull=%q, rcuNoCBS=%q",
			cpuList, isolcpus, nohzFull, rcuNoCBS)
	}

	// Get grub config path
	grubPath := m.cfg.GrubConfigPath
	if grubPath == "" {
		grubPath = "/media/usb/boot/grub/grub.cfg"
	}

	if debugMode {
		log.Printf("[DEBUG] ApplyCPUParamsToKernel: grubPath = %s", grubPath)
	}

	// Remount /media/usb as rw before modifying grub.cfg.
	remountPath := detectMountPath(grubPath)
	if remountPath != "" {
		if err := remountFilesystem(remountPath, "rw"); err != nil {
			return fmt.Errorf("remount %s as rw: %w", remountPath, err)
		}
		defer func() {
			if err := remountFilesystem(remountPath, "ro"); err != nil {
				log.Printf("[WARN] ApplyCPUParamsToKernel: failed to remount %s as ro: %v", remountPath, err)
			}
		}()
	}

	// Update grub.cfg
	if err := UpdateGrubCPUParams(isolcpus, nohzFull, rcuNoCBS, grubPath); err != nil {
		if debugMode {
			log.Printf("[DEBUG] ApplyCPUParamsToKernel: UpdateGrubCPUParams failed: %v", err)
		}
		return fmt.Errorf("update grub.cfg: %w", err)
	}

	if debugMode {
		log.Println("[DEBUG] ApplyCPUParamsToKernel: completed successfully")
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

// buildHostCPUList extracts sorted, deduplicated host CPU IDs from pinning mappings
// and returns them as a comma-separated string (e.g., "0,1,2,3").
func buildHostCPUList(mappings []models.VCPUToHostMapping) string {
	cpuSet := make(map[int]bool)
	for _, m := range mappings {
		cpuSet[m.HostCPUID] = true
	}
	cpus := make([]int, 0, len(cpuSet))
	for cpu := range cpuSet {
		cpus = append(cpus, cpu)
	}
	sort.Ints(cpus)

	if len(cpus) == 0 {
		return ""
	}
	var parts []string
	for _, c := range cpus {
		parts = append(parts, strconv.Itoa(c))
	}
	return strings.Join(parts, ",")
}

// detectMountPath checks if a file path resides under /media/usb and returns
// the mount point path, or an empty string if no remount is needed.
func detectMountPath(filePath string) string {
	if strings.HasPrefix(filePath, "/media/usb") {
		return "/media/usb"
	}
	return ""
}

// remountFilesystem remounts the given mount point with the specified options
// (e.g., "rw" or "ro"). It uses the mount command with the remount flag.
func remountFilesystem(mountPath, mode string) error {
	if debugMode {
		log.Printf("[DEBUG] remountFilesystem: mount -o remount,%s %s", mode, mountPath)
	}
	cmd := exec.Command("mount", "-o", "remount,"+mode, mountPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount -o remount,%s %s: %s: %w", mode, mountPath, string(output), err)
	}
	if debugMode {
		log.Printf("[DEBUG] remountFilesystem: successfully remounted %s as %s", mountPath, mode)
	}
	return nil
}
