package vm

import "github.com/glemsom/dkvmmanager/internal/domain"

// HostDiscovery provides host hardware scanning capabilities.
type HostDiscovery interface {
	ScanPCIDevices() ([]domain.PCIDevice, error)
	ScanUSBDevices() ([]domain.USBDevice, error)
	ScanCPUTopology() (domain.HostCPUTopology, error)
}

// DefaultHostDiscovery is the production implementation using system scanners.
type DefaultHostDiscovery struct{}

func (d *DefaultHostDiscovery) ScanPCIDevices() ([]domain.PCIDevice, error) {
	return NewPCIScanner().ScanDevices()
}

func (d *DefaultHostDiscovery) ScanUSBDevices() ([]domain.USBDevice, error) {
	return NewUSBScanner().ScanDevices()
}

func (d *DefaultHostDiscovery) ScanCPUTopology() (domain.HostCPUTopology, error) {
	return NewCPUScanner().ScanTopology()
}
