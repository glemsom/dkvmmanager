package models

import (
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// TestVCPUPinningFormModelInterface verifies that VCPUPinningFormModel
// satisfies the form.FormModel interface at compile time.
func TestVCPUPinningFormModelInterface(t *testing.T) {
	var _ form.FormModel = (*VCPUPinningFormModel)(nil)
}

// TestVCPUPinningBuildPositions verifies the positions are built correctly.
func TestVCPUPinningBuildPositions(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)

	positions := m.BuildPositions()
	if len(positions) != 3 {
		t.Fatalf("expected 3 positions, got %d", len(positions))
	}

	// Position 0: Toggle
	if positions[0].Kind != form.FocusToggle {
		t.Errorf("position 0: expected FocusToggle, got %v", positions[0].Kind)
	}
	if positions[0].Key != "enabled" {
		t.Errorf("position 0: expected key 'enabled', got '%s'", positions[0].Key)
	}

	// Position 1: Save button
	if positions[1].Kind != form.FocusButton {
		t.Errorf("position 1: expected FocusButton, got %v", positions[1].Kind)
	}
	if positions[1].Key != "save" {
		t.Errorf("position 1: expected key 'save', got '%s'", positions[1].Key)
	}

	// Position 2: Apply to Kernel button
	if positions[2].Kind != form.FocusButton {
		t.Errorf("position 2: expected FocusButton, got %v", positions[2].Kind)
	}
	if positions[2].Key != "apply_kernel" {
		t.Errorf("position 2: expected key 'apply_kernel', got '%s'", positions[2].Key)
	}
}

// TestVCPUPinningNavigation tests Tab/Shift+Tab navigation.
func TestVCPUPinningNavigation(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)

	// Start at position 0 (toggle)
	if m.CurrentIndex() != 0 {
		t.Fatalf("expected initial focusIndex 0, got %d", m.CurrentIndex())
	}

	// Tab → position 1 (save)
	m.moveFocus(1)
	if m.CurrentIndex() != 1 {
		t.Errorf("after tab: expected focusIndex 1, got %d", m.CurrentIndex())
	}

	// Tab → position 2 (apply_kernel)
	m.moveFocus(1)
	if m.CurrentIndex() != 2 {
		t.Errorf("after 2nd tab: expected focusIndex 2, got %d", m.CurrentIndex())
	}

	// Tab at end → stays at 2 (clamp)
	m.moveFocus(1)
	if m.CurrentIndex() != 2 {
		t.Errorf("tab at end: expected focusIndex 2, got %d", m.CurrentIndex())
	}

	// Shift+Tab → position 1
	m.moveFocus(-1)
	if m.CurrentIndex() != 1 {
		t.Errorf("after shift+tab: expected focusIndex 1, got %d", m.CurrentIndex())
	}

	// Shift+Tab → position 0
	m.moveFocus(-1)
	if m.CurrentIndex() != 0 {
		t.Errorf("after 2nd shift+tab: expected focusIndex 0, got %d", m.CurrentIndex())
	}

	// Shift+Tab at start → stays at 0 (clamp)
	m.moveFocus(-1)
	if m.CurrentIndex() != 0 {
		t.Errorf("shift+tab at start: expected focusIndex 0, got %d", m.CurrentIndex())
	}
}

// TestVCPUPinningToggle tests toggling pinning via Enter.
func TestVCPUPinningToggle(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)

	// Start disabled
	if m.pinning.Enabled {
		t.Fatal("expected pinning to start disabled")
	}

	// Toggle via Enter on position 0
	pos := m.positions[0]
	result, cmd := m.HandleEnter(pos)
	if result != form.ResultNone {
		t.Errorf("toggle enter: expected ResultNone, got %v", result)
	}
	if cmd != nil {
		t.Errorf("toggle enter: expected nil cmd, got %v", cmd)
	}
	if !m.pinning.Enabled {
		t.Error("expected pinning to be enabled after toggle")
	}

	// Toggle again to disable
	result, cmd = m.HandleEnter(pos)
	if result != form.ResultNone {
		t.Errorf("toggle enter 2nd: expected ResultNone, got %v", result)
	}
	if m.pinning.Enabled {
		t.Error("expected pinning to be disabled after 2nd toggle")
	}
}

// TestVCPUPinningHandleCharNoOp verifies HandleChar is a no-op.
func TestVCPUPinningHandleCharNoOp(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)

	// HandleChar should be a no-op (no text fields)
	m.HandleChar(m.positions[0], "a")
	m.HandleChar(m.positions[1], "x")

	if m.pinning.Enabled {
		t.Error("HandleChar should not toggle pinning")
	}
}

// TestVCPUPinningHandleBackspaceNoOp verifies HandleBackspace is a no-op.
func TestVCPUPinningHandleBackspaceNoOp(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)
	m.HandleBackspace(m.positions[0]) // should not panic
}

// TestVCPUPinningHandleDeleteNoOp verifies HandleDelete is a no-op.
func TestVCPUPinningHandleDeleteNoOp(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)
	m.HandleDelete(m.positions[0]) // should not panic
}

// TestVCPUPinningHandleMessageKernelApplied verifies async kernel message handling.
func TestVCPUPinningHandleMessageKernelApplied(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)

	// Success message
	cmd := m.HandleMessage(VCPUCPUKernelAppliedMsg{Success: true})
	if cmd != nil {
		t.Errorf("expected nil cmd, got %v", cmd)
	}
	if m.kernelMsg == "" {
		t.Error("expected kernelMsg to be set on success")
	}
	if m.kernelMsgErr {
		t.Error("expected kernelMsgErr to be false on success")
	}

	// Error message
	cmd = m.HandleMessage(VCPUCPUKernelAppliedMsg{Success: false, Error: "test error"})
	if m.kernelMsg != "test error" {
		t.Errorf("expected kernelMsg 'test error', got '%s'", m.kernelMsg)
	}
	if !m.kernelMsgErr {
		t.Error("expected kernelMsgErr to be true on error")
	}
}

// TestVCPUPinningRenderHeader verifies RenderHeader produces non-empty output.
func TestVCPUPinningRenderHeader(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)

	header := m.RenderHeader()
	if header == "" {
		t.Error("expected non-empty RenderHeader output")
	}
}

// TestVCPUPinningRenderFooter verifies RenderFooter produces non-empty output.
func TestVCPUPinningRenderFooter(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)

	footer := m.RenderFooter()
	if footer == "" {
		t.Error("expected non-empty RenderFooter output")
	}
}

// TestVCPUPinningRenderPosition tests rendering each position type.
func TestVCPUPinningRenderPosition(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)

	// Render toggle focused
	output := m.RenderPosition(m.positions[0], true, -1)
	if output == "" {
		t.Error("expected non-empty toggle render output")
	}

	// Render save button focused
	output = m.RenderPosition(m.positions[1], true, -1)
	if output == "" {
		t.Error("expected non-empty save button render output")
	}

	// Render apply_kernel button focused
	output = m.RenderPosition(m.positions[2], true, -1)
	if output == "" {
		t.Error("expected non-empty apply_kernel button render output")
	}
}

// TestVCPUPinningSetFocusIndex verifies SetFocusIndex works correctly.
func TestVCPUPinningSetFocusIndex(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)

	m.SetFocusIndex(2)
	if m.CurrentIndex() != 2 {
		t.Errorf("expected focusIndex 2, got %d", m.CurrentIndex())
	}

	m.SetFocusIndex(0)
	if m.CurrentIndex() != 0 {
		t.Errorf("expected focusIndex 0, got %d", m.CurrentIndex())
	}
}

// TestVCPUPinningLifecycle verifies lifecycle methods don't panic.
func TestVCPUPinningLifecycle(t *testing.T) {
	m := newTestVCPUPinningFormModel(t)

	m.OnEnter()        // should not panic
	m.OnExit()         // should not panic
	m.SetFocused(true) // should not panic
	m.SetSize(80, 25)  // should not panic
}

// TestVCPUPinningUpdatedMsgImplementsFormSavedMsg verifies the message type.
func TestVCPUPinningUpdatedMsgImplementsFormSavedMsg(t *testing.T) {
	var _ form.FormSavedMsg = VCPUPinningUpdatedMsg{}

	msg := VCPUPinningUpdatedMsg{}
	if msg.FormName() != "vCPU Pinning" {
		t.Errorf("expected FormName 'vCPU Pinning', got '%s'", msg.FormName())
	}
	if msg.FormStatus() != "" {
		t.Errorf("expected empty FormStatus, got '%s'", msg.FormStatus())
	}
}

// newTestVCPUPinningFormModel creates a testable VCPUPinningFormModel.
// Uses the real constructor if possible, falls back to manual construction.
func newTestVCPUPinningFormModel(t *testing.T) *VCPUPinningFormModel {
	t.Helper()

	vmManager := createTestVMManager(t)

	// Try the real constructor (may fail without real CPU topology hardware)
	formModel, err := NewVCPUPinningFormModel(vmManager)
	if err == nil {
		formModel.focusIndex = 0
		formModel.pinning.Enabled = false
		formModel.pinning.Mappings = nil
		return formModel
	}

	// Fallback: construct manually for environments without CPU topology
	formModel = &VCPUPinningFormModel{
		vmManager: vmManager,
		hostTopo: models.HostCPUTopology{
			Dies: []models.CPUDie{
				{
					ID:   0,
					Cores: 4,
					CoreDetails: []models.CPUCore{
						{ID: 0, Threads: []int{0, 1}},
						{ID: 1, Threads: []int{2, 3}},
						{ID: 2, Threads: []int{4, 5}},
						{ID: 3, Threads: []int{6, 7}},
					},
				},
			},
			TotalCores: 4,
			TotalCPUs:  8,
		},
		topology: models.CPUTopology{
			Enabled:      true,
			SelectedCPUs: []int{0, 1, 2, 3},
		},
		pinning:    models.VCPUPinningGlobal{Enabled: false, Mappings: nil},
		errors:     make(map[string]string),
		focusIndex: 0,
		scanErr:    nil,
	}

	formModel.positions = formModel.BuildPositions()
	return formModel
}
