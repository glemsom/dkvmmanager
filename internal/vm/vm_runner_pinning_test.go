package vm

import (
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
)

func TestApplyVCPUPinningNoClient(t *testing.T) {
	r := NewVMRunner(&models.VM{ID: "1", Name: "t"}, &config.Config{DataFolder: t.TempDir(), QEMUPath: "/bin/true"}, RunConfig{})
	err := r.ApplyVCPUPinning(models.VCPUPinningGlobal{Enabled: true, Mappings: []models.VCPUToHostMapping{{VCPUID: 0, HostCPUID: 0}}})
	if err == nil {
		t.Fatal("expected error when qmp client is nil")
	}
}

func TestApplyVCPUPinningDisabled(t *testing.T) {
	r := NewVMRunner(&models.VM{ID: "1", Name: "t"}, &config.Config{DataFolder: t.TempDir(), QEMUPath: "/bin/true"}, RunConfig{})
	if err := r.ApplyVCPUPinning(models.VCPUPinningGlobal{}); err != nil {
		t.Fatalf("disabled pinning should be no-op: %v", err)
	}
}
