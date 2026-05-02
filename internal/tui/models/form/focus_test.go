package form

import (
	"testing"
)

// --- Tracer Bullet: Basic moveFocus clamps to valid range ---

func TestMoveFocus_ClampsToLowerBound(t *testing.T) {
	positions := makeFakePositions(3)
	idx := 0
	idx = moveFocus(positions, idx, -1)
	if idx != 0 {
		t.Errorf("expected focus index 0 after moving -1 from 0, got %d", idx)
	}
}

func TestMoveFocus_ClampsToUpperBound(t *testing.T) {
	positions := makeFakePositions(3)
	idx := 2
	idx = moveFocus(positions, idx, 1)
	if idx != 2 {
		t.Errorf("expected focus index 2 after moving +1 from last, got %d", idx)
	}
}

func TestMoveFocus_MovesForward(t *testing.T) {
	positions := makeFakePositions(5)
	idx := 1
	idx = moveFocus(positions, idx, 1)
	if idx != 2 {
		t.Errorf("expected focus index 2, got %d", idx)
	}
}

func TestMoveFocus_MovesBackward(t *testing.T) {
	positions := makeFakePositions(5)
	idx := 3
	idx = moveFocus(positions, idx, -1)
	if idx != 2 {
		t.Errorf("expected focus index 2, got %d", idx)
	}
}

func TestMoveFocus_MovesMultipleSteps(t *testing.T) {
	positions := makeFakePositions(5)
	idx := 0
	idx = moveFocus(positions, idx, 3)
	if idx != 3 {
		t.Errorf("expected focus index 3, got %d", idx)
	}
}

func TestMoveFocus_ClampsWhenDeltaExceeds(t *testing.T) {
	positions := makeFakePositions(5)
	idx := 4
	idx = moveFocus(positions, idx, 10)
	if idx != 4 {
		t.Errorf("expected focus index 4 (clamped), got %d", idx)
	}
}

func TestMoveFocus_ClampsNegativeDelta(t *testing.T) {
	positions := makeFakePositions(5)
	idx := 0
	idx = moveFocus(positions, idx, -10)
	if idx != 0 {
		t.Errorf("expected focus index 0 (clamped), got %d", idx)
	}
}

func TestMoveFocus_ZeroDelta(t *testing.T) {
	positions := makeFakePositions(3)
	idx := 1
	idx = moveFocus(positions, idx, 0)
	if idx != 1 {
		t.Errorf("expected focus index 1 with zero delta, got %d", idx)
	}
}

func TestMoveFocus_EmptyPositions(t *testing.T) {
	var positions []FocusPos
	idx := 0
	idx = moveFocus(positions, idx, 1)
	if idx != 0 {
		t.Errorf("expected focus index 0 with empty positions, got %d", idx)
	}
}

// --- Focused line index calculation ---

func TestFocusedLineIndex_ReturnsPositionIndex(t *testing.T) {
	positions := makeFakePositions(5)
	for i := range positions {
		line := focusedLineIndex(positions, i, 0)
		if line != i {
			t.Errorf("position %d: expected line %d, got %d", i, i, line)
		}
	}
}

// --- Viewport clampOffset ---

func TestClampOffset_AdjustsWhenTargetBelowViewport(t *testing.T) {
	// Target line is above the visible area; offset should decrease
	offset := 10
	targetLine := 5
	viewHeight := 20
	newOffset := clampOffset(offset, targetLine, viewHeight)
	// The target line (5) should be at least at the top of the viewport (offset)
	// So offset should be <= 5
	if newOffset > targetLine {
		t.Errorf("expected offset <= %d, got %d", targetLine, newOffset)
	}
}

func TestClampOffset_AdjustsWhenTargetBelowViewportBottom(t *testing.T) {
	// Target line is below the visible area; offset should increase
	offset := 0
	targetLine := 25
	viewHeight := 20
	newOffset := clampOffset(offset, targetLine, viewHeight)
	// The target line should be visible: offset + viewHeight > targetLine
	if newOffset+viewHeight <= targetLine {
		t.Errorf("expected offset+viewHeight > %d, got offset=%d (sum=%d)", targetLine, newOffset, newOffset+viewHeight)
	}
}

func TestClampOffset_NoChangeWhenTargetVisible(t *testing.T) {
	// Target line is already visible; offset should not change
	offset := 10
	targetLine := 15
	viewHeight := 20
	newOffset := clampOffset(offset, targetLine, viewHeight)
	if newOffset != offset {
		t.Errorf("expected offset to remain %d, got %d", offset, newOffset)
	}
}

func TestClampOffset_NeverNegative(t *testing.T) {
	offset := 0
	targetLine := 0
	viewHeight := 20
	newOffset := clampOffset(offset, targetLine, viewHeight)
	if newOffset < 0 {
		t.Errorf("expected offset >= 0, got %d", newOffset)
	}
}

// Helper

func makeFakePositions(n int) []FocusPos {
	positions := make([]FocusPos, n)
	for i := range positions {
		positions[i] = FocusPos{
			Kind:  FocusText,
			Label: "field",
			Key:   "field",
		}
	}
	return positions
}
