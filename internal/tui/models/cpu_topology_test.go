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

	if !wrapper.form.Ready() {
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
	if view != "Loading..." {
		t.Errorf("Expected 'Loading...', got %q", view)
	}

	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view = model.View()
	if view == "Loading..." {
		t.Error("Should not show loading after WindowSizeMsg")
	}
}

// TestCPUTopologyFormAccessor tests the Form() accessor
func TestCPUTopologyFormAccessor(t *testing.T) {
	vmManager := createTestVMManager(t)

	model, err := NewCPUTopologyModel(vmManager)
	if err != nil {
		t.Fatalf("NewCPUTopologyModel returned error: %v", err)
	}

	f := model.Form()
	if f == nil {
		t.Fatal("Expected Form() to return non-nil")
	}

	if f.focusIndex != 0 {
		t.Errorf("Initial focusIndex = %d, want 0", f.focusIndex)
	}
}

// TestCPUTopologyUpdatedMsgImplementsFormSavedMsg verifies the message interface
func TestCPUTopologyUpdatedMsgImplementsFormSavedMsg(t *testing.T) {
	var msg any = CPUTopologyUpdatedMsg{}
	if _, ok := msg.(interface{ IsFormSaved() }); !ok {
		t.Error("CPUTopologyUpdatedMsg does not implement IsFormSaved()")
	}
	if _, ok := msg.(interface{ FormName() string }); !ok {
		t.Error("CPUTopologyUpdatedMsg does not implement FormName()")
	}
	if _, ok := msg.(interface{ FormStatus() string }); !ok {
		t.Error("CPUTopologyUpdatedMsg does not implement FormStatus()")
	}
}
