package form

// moveFocus moves the focus index by delta, clamping to the valid range [0, len(positions)-1].
// Does not wrap around. Returns the new focus index.
func moveFocus(positions []FocusPos, index int, delta int) int {
	if len(positions) == 0 {
		return 0
	}
	index += delta
	if index < 0 {
		return 0
	}
	if index >= len(positions) {
		return len(positions) - 1
	}
	return index
}

// focusedLineIndex returns the line number (0-based) where a given position
// renders in the form's content. For the basic framework, each position maps
// to exactly one line, so this returns the position index plus any header lines.
func focusedLineIndex(positions []FocusPos, positionIndex int, headerLines int) int {
	return headerLines + positionIndex
}

// clampOffset adjusts the viewport scroll offset so that the target line is
// visible within the viewport. If the target line is above the visible area,
// the offset is reduced. If it's below, the offset is increased. If already
// visible, the offset is unchanged.
func clampOffset(offset int, targetLine int, viewHeight int) int {
	if targetLine < offset {
		return targetLine
	}
	if targetLine >= offset+viewHeight {
		return targetLine - viewHeight + 1
	}
	return offset
}
