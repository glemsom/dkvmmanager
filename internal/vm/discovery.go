package vm

import "github.com/glemsom/dkvmmanager/internal/models"

// HostDiscovery provides host hardware scanning capabilities.
type HostDiscovery interface {
	ScanPCIDevices() ([]models.PCIDevice, error)
	ScanUSBDevices() ([]models.USBDevice, error)
	ScanCPUTopology() (models.HostCPUTopology, error)
}

// DefaultHostDiscovery is the production implementation using system scanners.
type DefaultHostDiscovery struct{}

func (d *DefaultHostDiscovery) ScanPCIDevices() ([]models.PCIDevice, error) {
	return NewPCIScanner().ScanDevices()
}

func (d *DefaultHostDiscovery) ScanUSBDevices() ([]models.USBDevice, error) {
	return NewUSBScanner().ScanDevices()
}

func (d *DefaultHostDiscovery) ScanCPUTopology() (models.HostCPUTopology, error) {
	return NewCPUScanner().ScanTopology()
}
