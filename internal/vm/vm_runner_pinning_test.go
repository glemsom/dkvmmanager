package vm

import (
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/domain"
)

func TestApplyVCPUPinningNoClient(t *testing.T) {
	r := NewVMRunner(&domain.VM{ID: "1", Name: "t"}, &config.Config{DataFolder: t.TempDir(), QEMUPath: "/bin/true"}, RunConfig{}, false)
	err := r.ApplyVCPUPinning(domain.VCPUPinningGlobal{Enabled: true, Mappings: []domain.VCPUToHostMapping{{VCPUID: 0, HostCPUID: 0}}})
	if err == nil {
		t.Fatal("expected error when qmp client is nil")
	}
}

func TestApplyVCPUPinningDisabled(t *testing.T) {
	r := NewVMRunner(&domain.VM{ID: "1", Name: "t"}, &config.Config{DataFolder: t.TempDir(), QEMUPath: "/bin/true"}, RunConfig{}, false)
	if err := r.ApplyVCPUPinning(domain.VCPUPinningGlobal{}); err != nil {
		t.Fatalf("disabled pinning should be no-op: %v", err)
	}
}
