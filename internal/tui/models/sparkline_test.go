package models

import (
	"testing"
	"unicode"
)

func TestRenderBrailleSparkline_EmptyInput(t *testing.T) {
	if got := RenderBrailleSparkline(nil, 5); got != "" {
		t.Errorf("nil input: got %q, want empty", got)
	}
	if got := RenderBrailleSparkline([]float64{}, 5); got != "" {
		t.Errorf("empty slice: got %q, want empty", got)
	}
}

func TestRenderBrailleSparkline_ZeroWidth(t *testing.T) {
	if got := RenderBrailleSparkline([]float64{1, 2, 3}, 0); got != "" {
		t.Errorf("zero width: got %q, want empty", got)
	}
	if got := RenderBrailleSparkline([]float64{1, 2, 3}, -1); got != "" {
		t.Errorf("negative width: got %q, want empty", got)
	}
}

func TestRenderBrailleSparkline_SingleValue(t *testing.T) {
	result := RenderBrailleSparkline([]float64{42}, 3)
	if len([]rune(result)) != 3 {
		t.Errorf("single value width=3: got %d chars, want 3", len([]rune(result)))
	}
	// All chars should be identical (same value repeated)
	runes := []rune(result)
	for i := 1; i < len(runes); i++ {
		if runes[i] != runes[0] {
			t.Errorf("single value: char %d (%U) differs from char 0 (%U)", i, runes[i], runes[0])
		}
	}
	// Should not be blank (mid-height for flat line)
	if runes[0] == 0x2800 {
		t.Error("single value flat line rendered as blank char")
	}
}

func TestRenderBrailleSparkline_FlatLine(t *testing.T) {
	// All zeros
	result := RenderBrailleSparkline([]float64{0, 0, 0, 0}, 4)
	if len([]rune(result)) != 4 {
		t.Errorf("flat line: got %d chars, want 4", len([]rune(result)))
	}
	// Should not be blank
	runes := []rune(result)
	if runes[0] == 0x2800 {
		t.Error("flat line rendered as blank char, expected visible baseline")
	}
	// All chars identical
	for i := 1; i < len(runes); i++ {
		if runes[i] != runes[0] {
			t.Errorf("flat line: char %d differs from char 0", i)
		}
	}
}

func TestRenderBrailleSparkline_FlatLineNonZero(t *testing.T) {
	// All same non-zero value
	result := RenderBrailleSparkline([]float64{5, 5, 5}, 2)
	runes := []rune(result)
	if len(runes) != 2 {
		t.Errorf("flat non-zero: got %d chars, want 2", len(runes))
	}
	if runes[0] == 0x2800 {
		t.Error("flat non-zero line rendered as blank char")
	}
	if runes[0] != runes[1] {
		t.Error("flat non-zero: chars differ")
	}
}

func TestRenderBrailleSparkline_IncreasingValues(t *testing.T) {
	// Values from min to max: should show rising pattern
	values := []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	result := RenderBrailleSparkline(values, 5)
	runes := []rune(result)
	if len(runes) != 5 {
		t.Errorf("increasing: got %d chars, want 5", len(runes))
	}
	// First char should be lowest, last char highest
	first := runes[0]
	last := runes[len(runes)-1]
	if last < first {
		t.Error("increasing: last char lower than first, expected rising trend")
	}
}

func TestRenderBrailleSparkline_DecreasingValues(t *testing.T) {
	values := []float64{100, 90, 80, 70, 60, 50, 40, 30, 20, 10, 0}
	result := RenderBrailleSparkline(values, 5)
	runes := []rune(result)
	if len(runes) != 5 {
		t.Errorf("decreasing: got %d chars, want 5", len(runes))
	}
	first := runes[0]
	last := runes[len(runes)-1]
	if last > first {
		t.Error("decreasing: last char higher than first, expected falling trend")
	}
}

func TestRenderBrailleSparkline_WidthControlsLength(t *testing.T) {
	values := []float64{0, 10, 20, 30, 40, 50}
	for _, w := range []int{1, 2, 3, 6} {
		result := RenderBrailleSparkline(values, w)
		if len([]rune(result)) != w {
			t.Errorf("width=%d: got %d chars", w, len([]rune(result)))
		}
	}
}

func TestRenderBrailleSparkline_AllDotLevelsReachable(t *testing.T) {
	// Test that all 8 individual dots in the braille cell can be reached.
	// We produce a value sequence that exercises both columns at various heights.
	// Left column dots (bottom→top): dot 7 (bit 6), dot 3 (bit 2), dot 2 (bit 1), dot 1 (bit 0)
	// Right column dots (bottom→top): dot 8 (bit 7), dot 6 (bit 5), dot 5 (bit 4), dot 4 (bit 3)
	//
	// Generate values that create left-level=1 and right-level=1 through left-level=4, right-level=4
	// to cover all dot positions.
	//
	// With min=0, max=100:
	//   level 1 (1 dot) = value between >0 and <=12.5
	//   level 2 (1 dot) = value between >12.5 and <=25
	//   level 3 (2 dots) = value between >25 and <=37.5
	//   level 4 (2 dots) = value between >37.5 and <=50
	//   level 5 (3 dots) = value between >50 and <=62.5
	//   level 6 (3 dots) = value between >62.5 and <=75
	//   level 7 (4 dots) = value between >75 and <=87.5
	//   level 8 (4 dots) = value between >87.5 and <=100
	//
	// Actually levelToDotCount: 0→0, 1→0, 2→1, 3→1, 4→2, 5→2, 6→3, 7→3, 8→4
	// So we need levels 2,4,6,8 to get 1,2,3,4 dots.
	// For left column 1 dot (level 2): bit 6 (dot 7)
	// For left column 2 dots (level 4): bits 6, 2 (dots 7, 3)
	// For left column 3 dots (level 6): bits 6, 2, 1 (dots 7, 3, 2)
	// For left column 4 dots (level 8): bits 6, 2, 1, 0 (dots 7, 3, 2, 1)
	// For right column 1 dot (level 2): bit 7 (dot 8)
	// For right column 2 dots (level 4): bits 7, 5 (dots 8, 6)
	// For right column 3 dots (level 6): bits 7, 5, 4 (dots 8, 6, 5)
	// For right column 4 dots (level 8): bits 7, 5, 4, 3 (dots 8, 6, 5, 4)



	// Use 4 braille chars, each with a specific left+right level combo
	// Char 0: left level 0 (0 dots), right level 2 (1 dot)
	// Char 1: left level 4 (2 dots), right level 4 (2 dots)
	// Char 2: left level 6 (3 dots), right level 6 (3 dots)
	// Char 3: left level 8 (4 dots), right level 8 (4 dots)
	//
	// With min=0, max=100:
	// bucket 0-1: level 2 → value = 2*100/8 = 25, but need level 2 = ceil? levelToDotCount(2) = (2*4+4)/8 = 12/8 = 1
	// Actually level=2 is value >= 2*100/8 = 25. So value of 25 gives normalized=0.25, level=int(round(0.25*8))=int(2.0)=2.
	// Wait, round(0.25*8) = round(2.0) = 2. Good.
	// level 2 → 1 dot
	// Level 4 → value = 50. round(0.5*8) = 4. ✓
	// Level 6 → value = 75. round(0.75*8) = round(6.0) = 6. ✓
	// Level 8 → value = 100. round(1.0*8) = 8. ✓

	// Build 8 values (4 chars * 2 buckets each)
	// Start at 0 so min=0, giving clean level mapping
	v := []float64{0, 25, 50, 50, 75, 75, 100, 100}
	result := RenderBrailleSparkline(v, 4)
	runes := []rune(result)
	if len(runes) != 4 {
		t.Fatalf("expected 4 runes, got %d", len(runes))
	}

	// Verify each braille char has the right dot pattern
	// Char 0: left=0 (level 0, 0 dots), right=25 (level 2, 1 dot) → bit 7 only → offset=128
	expected0 := rune(0x2800 + 128)
	if runes[0] != expected0 {
		t.Errorf("char 0 (left=0dots, right=1dot): got %U, want %U", runes[0], expected0)
	}

	// Char 1: left=50 (level 4, 2 dots), right=50 (level 4, 2 dots) → bits 6,2,7,5 = 64+4+128+32 = 228
	expected1 := rune(0x2800 + 228)
	if runes[1] != expected1 {
		t.Errorf("char 1 (2 dots each): got %U, want %U", runes[1], expected1)
	}

	// Char 2: left=75 (level 6, 3 dots), right=75 (level 6, 3 dots) → bits 6,2,1,7,5,4 = 64+4+2+128+32+16 = 246
	expected2 := rune(0x2800 + 246)
	if runes[2] != expected2 {
		t.Errorf("char 2 (3 dots each): got %U, want %U", runes[2], expected2)
	}

	// Char 3: left=100 (level 8, 4 dots), right=100 (level 8, 4 dots) → all 8 bits → 255
	expected3 := rune(0x2800 + 255)
	if runes[3] != expected3 {
		t.Errorf("char 3 (4 dots each): got %U, want %U", runes[3], expected3)
	}
}

func TestRenderBrailleSparkline_AllBrailleInRange(t *testing.T) {
	// All output chars must be braille pattern Unicode (U+2800–U+28FF)
	values := []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	result := RenderBrailleSparkline(values, 10)
	for i, r := range result {
		if r < 0x2800 || r > 0x28FF {
			t.Errorf("char %d: %U is not in braille range U+2800–U+28FF", i, r)
		}
	}
}

func TestRenderBrailleSparkline_Downsampling(t *testing.T) {
	// More values than buckets — should downsample without error
	values := make([]float64, 100)
	for i := range values {
		values[i] = float64(i)
	}
	result := RenderBrailleSparkline(values, 10)
	if len([]rune(result)) != 10 {
		t.Errorf("downsample: got %d chars, want 10", len([]rune(result)))
	}
}

func TestRenderBrailleSparkline_Upsampling(t *testing.T) {
	// Fewer values than buckets — should upsample without error
	values := []float64{0, 50, 100}
	result := RenderBrailleSparkline(values, 10)
	if len([]rune(result)) != 10 {
		t.Errorf("upsample: got %d chars, want 10", len([]rune(result)))
	}
	// Values monotonic, so output should be too
	runes := []rune(result)
	if runes[0] > runes[len(runes)-1] {
		t.Error("upsample: first > last, expected increasing trend")
	}
}

// Test that each output rune is actually a braille pattern character
func TestRenderBrailleSparkline_OutputIsBraille(t *testing.T) {
	values := []float64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := RenderBrailleSparkline(values, 5)
	for i, r := range result {
		if !unicode.IsPrint(r) {
			t.Errorf("char %d: %U is not printable", i, r)
		}
	}
}

func TestRenderBrailleSparkline_ExactMatchValuesBuckets(t *testing.T) {
	// When len(values) == width*2, no resampling occurs
	values := []float64{0, 100, 50, 50}
	result := RenderBrailleSparkline(values, 2)
	runes := []rune(result)
	if len(runes) != 2 {
		t.Fatalf("got %d chars, want 2", len(runes))
	}
	// First char: left=0, right=100 → left level 0, right level 8
	// left level 0 → 0 dots, right level 8 → 4 dots
	// Only right dots lit: bits 7,5,4,3 → offset = 128+32+16+8 = 184
	expectedFirst := rune(0x2800 + 184)
	if runes[0] != expectedFirst {
		t.Errorf("first char: got %U, want %U", runes[0], expectedFirst)
	}
}

func TestRenderBrailleSparkline_LargeValues(t *testing.T) {
	// Large float64 values shouldn't cause issues
	values := []float64{1e10, 2e10, 3e10, 4e10, 5e10}
	result := RenderBrailleSparkline(values, 3)
	if len([]rune(result)) != 3 {
		t.Errorf("large values: got %d chars, want 3", len([]rune(result)))
	}
}

func TestRenderBrailleSparkline_NegativeValues(t *testing.T) {
	values := []float64{-100, -50, 0, 50, 100}
	result := RenderBrailleSparkline(values, 3)
	if len([]rune(result)) != 3 {
		t.Errorf("negative values: got %d chars, want 3", len([]rune(result)))
	}
}

func TestRenderBrailleSparkline_AllSameLargeValue(t *testing.T) {
	result := RenderBrailleSparkline([]float64{999, 999, 999, 999}, 2)
	runes := []rune(result)
	if len(runes) != 2 {
		t.Fatalf("got %d chars, want 2", len(runes))
	}
	if runes[0] == 0x2800 {
		t.Error("flat large value rendered as blank")
	}
	if runes[0] != runes[1] {
		t.Error("flat large value: chars differ")
	}
}
