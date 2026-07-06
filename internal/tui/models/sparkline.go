package models

import "math"

// renderBrailleSparkline renders values as a braille sparkline of width braille characters.
// Each braille character represents 2 value buckets (left column=bucket n, right column=bucket n+1).
// Input is downsampled to width*2 buckets via averaging if larger, or upsampled via linear
// interpolation if smaller.
//
// The braille height mapping uses 8 dot levels (0–8) corresponding to the min–max range.
// A flat line (min==max) renders at mid-height for a visible baseline.
func renderBrailleSparkline(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}

	// Resample to exactly width*2 buckets
	buckets := width * 2
	sampled := resample(values, buckets)

	// Find min/max for normalization
	min, max := minMax(sampled)

	// Build braille chars
	result := make([]rune, width)
	for i := 0; i < width; i++ {
		leftVal := sampled[i*2]
		rightVal := sampled[i*2+1]
		result[i] = valuePairToBraille(leftVal, rightVal, min, max)
	}

	return string(result)
}

// resample adjusts values to exactly n buckets using averaging (downsample) or
// linear interpolation (upsample).
func resample(values []float64, n int) []float64 {
	if len(values) == n {
		out := make([]float64, n)
		copy(out, values)
		return out
	}

	out := make([]float64, n)

	if len(values) > n {
		// Downsample: partition into n groups and average each
		for i := 0; i < n; i++ {
			start := i * len(values) / n
			end := (i + 1) * len(values) / n
			if end > len(values) {
				end = len(values)
			}
			if end <= start {
				out[i] = values[start]
				continue
			}
			var sum float64
			for _, v := range values[start:end] {
				sum += v
			}
			out[i] = sum / float64(end-start)
		}
	} else {
		// Upsample: linear interpolation
		for i := 0; i < n; i++ {
			pos := float64(i) * float64(len(values)-1) / float64(n-1)
			idx := int(pos)
			frac := pos - float64(idx)
			if idx >= len(values)-1 {
				out[i] = values[len(values)-1]
			} else {
				out[i] = values[idx] + frac*(values[idx+1]-values[idx])
			}
		}
	}

	return out
}

// minMax returns the min and max of a float64 slice.
func minMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	min := values[0]
	max := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// valuePairToBraille maps two values to a single braille character using min/max range.
func valuePairToBraille(left, right, min, max float64) rune {
	// Handle flat line (min == max) – show mid-height for visible baseline
	if min == max {
		return brailleFromLevels(4, 4)
	}

	// Normalize to 0..8
	leftLevel := normalizeToLevel(left, min, max)
	rightLevel := normalizeToLevel(right, min, max)

	return brailleFromLevels(leftLevel, rightLevel)
}


// normalizeToLevel maps v in [min,max] to 0..8.
func normalizeToLevel(v, min, max float64) int {
	normalized := (v - min) / (max - min)
	level := int(math.Round(normalized * 8))
	if level < 0 {
		level = 0
	}
	if level > 8 {
		level = 8
	}
	return level
}

// brailleFromLevels returns a braille rune for the given left/right column levels (0–8).
// Level 0 = no dots; level 8 = all 4 dots in that column.
func brailleFromLevels(leftLevel, rightLevel int) rune {
	offset := 0

	// Left column dots (bottom to top): dot 7 (bit 6), dot 3 (bit 2), dot 2 (bit 1), dot 1 (bit 0)
	leftDots := levelToDotCount(leftLevel)
	leftBits := []int{6, 2, 1, 0}
	for i := 0; i < leftDots; i++ {
		offset |= 1 << leftBits[i]
	}

	// Right column dots (bottom to top): dot 8 (bit 7), dot 6 (bit 5), dot 5 (bit 4), dot 4 (bit 3)
	rightDots := levelToDotCount(rightLevel)
	rightBits := []int{7, 5, 4, 3}
	for i := 0; i < rightDots; i++ {
		offset |= 1 << rightBits[i]
	}

	return rune(0x2800 + offset)
}

// levelToDotCount maps a level 0–8 to a dot count 0–4.
// 0→0, 1→0, 2→1, 3→1, 4→2, 5→2, 6→3, 7→3, 8→4
func levelToDotCount(level int) int {
	return (level*4 + 4) / 8
}
