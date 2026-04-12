package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestCPUTopologyModelInit tests wrapper model initialization
func TestCPUTopologyModelInit(t *testing.T) {
	vmManager := createTestVMManager(t)

	model, err := NewCPUTopologyModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyModel returned error: %v", err)
	}

	if model.form == nil {
		t.Fatal("Expected form to be initialized")
	}

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init() should return nil cmd")
	}
}

// TestCPUTopologyModelUpdate tests wrapper model message forwarding
func TestCPUTopologyModelUpdate(t *testing.T) {
	vmManager := createTestVMManager(t)

	model, err := NewCPUTopologyModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyModel returned error: %v", err)
	}

	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	wrapper := updatedModel.(*CPUTopologyModel)

	if !wrapper.form.ready {
		t.Error("Form should be ready after WindowSizeMsg")
	}
}

// TestCPUTopologyModelView tests wrapper model view delegation
func TestCPUTopologyModelView(t *testing.T) {
	vmManager := createTestVMManager(t)

	model, err := NewCPUTopologyModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyModel returned error: %v", err)
	}

	view := model.View()
	if view != "Loading form..." {
		t.Errorf("Expected 'Loading form...', got %q", view)
	}

	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view = model.View()
	if view == "Loading form..." {
		t.Error("Should not show loading after WindowSizeMsg")
	}
}
