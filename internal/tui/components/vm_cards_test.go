package components

import (
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

func makeTestVMs(count int) []models.VM {
	vms := make([]models.VM, count)
	for i := range vms {
		vms[i] = models.VM{Name: "test-vm"}
	}
	return vms
}

func TestSetCursorValidIndex(t *testing.T) {
	vms := makeTestVMs(3)
	c := NewVMCardView(vms, 80, 24)

	c.SetCursor(1)
	if c.Cursor() != 1 {
		t.Errorf("Expected cursor 1, got %d", c.Cursor())
	}

	c.SetCursor(2)
	if c.Cursor() != 2 {
		t.Errorf("Expected cursor 2, got %d", c.Cursor())
	}
}

func TestSetCursorBoundsChecking(t *testing.T) {
	vms := makeTestVMs(3)
	c := NewVMCardView(vms, 80, 24)

	// Negative index should be ignored
	c.SetCursor(-1)
	if c.Cursor() != 0 {
		t.Errorf("Expected cursor 0 after negative set, got %d", c.Cursor())
	}

	// Index equal to length should be ignored
	c.SetCursor(3)
	if c.Cursor() != 0 {
		t.Errorf("Expected cursor 0 after out-of-bounds set, got %d", c.Cursor())
	}

	// Index beyond length should be ignored
	c.SetCursor(100)
	if c.Cursor() != 0 {
		t.Errorf("Expected cursor 0 after far out-of-bounds set, got %d", c.Cursor())
	}
}

func TestSetCursorEmptyVMs(t *testing.T) {
	c := NewVMCardView(nil, 80, 24)

	c.SetCursor(0)
	if c.Cursor() != 0 {
		t.Errorf("Expected cursor 0 on empty VMs, got %d", c.Cursor())
	}
}

func TestSetCursorPreservesOnSetVMs(t *testing.T) {
	vms := makeTestVMs(5)
	c := NewVMCardView(vms, 80, 24)

	c.SetCursor(3)
	if c.Cursor() != 3 {
		t.Fatalf("Expected cursor 3, got %d", c.Cursor())
	}

	// SetVMs with fewer items should clamp cursor
	c.SetVMs(makeTestVMs(2))
	if c.Cursor() != 1 {
		t.Errorf("Expected cursor clamped to 1, got %d", c.Cursor())
	}
}
