package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewCPUOptionsModel(t *testing.T) {
	mgr := createTestVMManager(t)
	m := NewCPUOptionsModel(mgr)

	if m == nil {
		t.Fatal("NewCPUOptionsModel returned nil")
	}
	if m.form == nil {
		t.Fatal("Expected non-nil form")
	}
}

func TestCPUOptionsModelInit(t *testing.T) {
	mgr := createTestVMManager(t)
	m := NewCPUOptionsModel(mgr)

	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil (delegates to form which returns nil)")
	}
}

func TestCPUOptionsModelUpdate(t *testing.T) {
	mgr := createTestVMManager(t)
	m := NewCPUOptionsModel(mgr)

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if updated == nil {
		t.Fatal("Update() returned nil model")
	}

	wrapped, ok := updated.(*CPUOptionsModel)
	if !ok {
		t.Fatalf("Update() should return *CPUOptionsModel, got %T", updated)
	}
	if wrapped.form == nil {
		t.Error("Wrapped model form should not be nil after Update()")
	}
	_ = cmd
}

func TestCPUOptionsModelUpdateWindowSize(t *testing.T) {
	mgr := createTestVMManager(t)
	m := NewCPUOptionsModel(mgr)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 50})
	wrapped := updated.(*CPUOptionsModel)

	if !wrapped.form.ready {
		t.Error("Form should be ready after WindowSizeMsg is forwarded")
	}
}

func TestCPUOptionsModelView(t *testing.T) {
	mgr := createTestVMManager(t)
	m := NewCPUOptionsModel(mgr)

	view := m.View()
	if view == "" {
		t.Error("View() should return non-empty string")
	}
}

func TestCPUOptionsModelViewContainsHeader(t *testing.T) {
	mgr := createTestVMManager(t)
	m := NewCPUOptionsModel(mgr)

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = updated.(*CPUOptionsModel)

	view := m.View()
	if !strings.Contains(view, "Hypervisor Stealth") {
		t.Error("View() should contain 'Hypervisor Stealth' section header")
	}
}

func TestCPUOptionsModelUpdateDelegatesKeyPress(t *testing.T) {
	mgr := createTestVMManager(t)
	m := NewCPUOptionsModel(mgr)

	initialField := m.form.currentPos().fieldName

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	wrapped := updated.(*CPUOptionsModel)

	newField := wrapped.form.currentPos().fieldName
	if newField == initialField {
		t.Error("Key press (Tab) should be delegated to form, focus should change")
	}
}
