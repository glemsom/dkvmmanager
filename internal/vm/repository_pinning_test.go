package vm

import (
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/config"
	"github.com/glemsom/dkvmmanager/internal/models"
)

func TestRepositoryVCPUPinningPersistence(t *testing.T) {
	repo, err := NewRepository(filepath.Join(t.TempDir(), "vms.yaml"))
	if err != nil {
		t.Fatalf("NewRepository: %v", err)
	}
	in := models.VCPUPinningGlobal{Enabled: true, Mappings: []models.VCPUToHostMapping{{VCPUID: 0, HostCPUID: 4}, {VCPUID: 1, HostCPUID: 5}}}
	if err := repo.SaveConfig("vcpu_pinning", in); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	var out models.VCPUPinningGlobal
	if err := repo.GetConfig("vcpu_pinning", &out); err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if !out.Enabled || len(out.Mappings) != 2 {
		t.Fatalf("unexpected output: %+v", out)
	}

	mgrCfg := &config.Config{DataFolder: t.TempDir(), VMsConfigFile: filepath.Join(t.TempDir(), "vms.yaml")}
	mgr, err := NewManager(mgrCfg)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if err := mgr.Repository().SaveConfig("vcpu_pinning", in); err != nil {
		t.Fatalf("manager SaveConfig: %v", err)
	}
	var mgrOut models.VCPUPinningGlobal
	if err := mgr.Repository().GetConfig("vcpu_pinning", &mgrOut); err != nil {
		t.Fatalf("manager GetConfig: %v", err)
	}
	if !mgrOut.Enabled {
		t.Fatal("manager output should be enabled")
	}
}
