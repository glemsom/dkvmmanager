package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// LayoutType represents the type of layout to use
type LayoutType int

const (
	LayoutDualPane      LayoutType = iota // Side-by-side layout
	LayoutVerticalStack                   // Stacked layout for narrow terminals
)

// Pane identifiers for focus management
const (
	PaneLeft  = 0
	PaneRight = 1
)

// Layout calculation constants
const (
	BorderWidth       = 2 // Top and bottom borders
	WideTerminalGap   = 4
	MediumTerminalGap = 2
	NarrowTerminalGap = 1
)

// LayoutConfig holds the configuration for a layout
type LayoutConfig struct {
	Type       LayoutType
	LeftWidth  int
	RightWidth int
	Gap        int
}

// DualPane represents a dual-pane layout component
type DualPane struct {
	leftContent  string
	rightContent string
	width        int
	height       int
	focusedPane  int // PaneLeft or PaneRight
	focused      bool // Whether the dual pane container has focus
}

// NewDualPane creates a new DualPane instance
func NewDualPane() *DualPane {
	return &DualPane{
		leftContent:  "",
		rightContent: "",
		width:        0,
		height:       0,
		focusedPane:  PaneLeft,
		focused:      false,
	}
}

// SetLeftContent sets the content for the left pane
func (d *DualPane) SetLeftContent(content string) {
	d.leftContent = content
}

// SetRightContent sets the content for the right pane
func (d *DualPane) SetRightContent(content string) {
	d.rightContent = content
}

// SetDimensions sets the width and height for the layout
func (d *DualPane) SetDimensions(width, height int) {
	d.width = width
	d.height = height
}

// SetFocusedPane sets which pane is focused (PaneLeft or PaneRight)
func (d *DualPane) SetFocusedPane(pane int) {
	if pane == PaneLeft || pane == PaneRight {
		d.focusedPane = pane
	}
}

// GetFocusedPane returns the currently focused pane
func (d *DualPane) GetFocusedPane() int {
	return d.focusedPane
}

// CalculateLayout calculates the layout configuration based on terminal width
func CalculateLayout(width int) LayoutConfig {
	// Wide terminal: >120 columns - side-by-side with generous spacing
	if width > 120 {
		return LayoutConfig{
			Type:       LayoutDualPane,
			LeftWidth:  (width - BorderWidth - WideTerminalGap) / 2,
			RightWidth: (width - BorderWidth - WideTerminalGap) / 2,
			Gap:        WideTerminalGap,
		}
	}

	// Medium terminal: 80-120 columns - side-by-side with minimal spacing
	if width >= 80 {
		return LayoutConfig{
			Type:       LayoutDualPane,
			LeftWidth:  (width - BorderWidth - MediumTerminalGap) / 2,
			RightWidth: (width - BorderWidth - MediumTerminalGap) / 2,
			Gap:        MediumTerminalGap,
		}
	}

	// Narrow terminal: <80 columns - vertical stack
	return LayoutConfig{
		Type:       LayoutVerticalStack,
		LeftWidth:  width - BorderWidth,
		RightWidth: width - BorderWidth,
		Gap:        NarrowTerminalGap,
	}
}

// Render renders the dual-pane layout
func (d *DualPane) Render() string {
	if d.width == 0 || d.height == 0 {
		return ""
	}

	config := CalculateLayout(d.width)

	switch config.Type {
	case LayoutDualPane:
		return d.renderDualPane(config)
	case LayoutVerticalStack:
		return d.renderVerticalStack(config)
	default:
		return d.renderDualPane(config)
	}
}

// applyFocusStyle applies focus styling to the appropriate pane style
func (d *DualPane) applyFocusStyle(leftStyle, rightStyle lipgloss.Style) (lipgloss.Style, lipgloss.Style) {
	if d.focused {
		if d.focusedPane == PaneLeft {
			leftStyle = styles.ActiveLayeredPanelStyle()
			rightStyle = styles.LayeredPanelStyle()
		} else {
			leftStyle = styles.LayeredPanelStyle()
			rightStyle = styles.ActiveLayeredPanelStyle()
		}
	} else {
		// When the container is not focused, both panes use the unfocused style
		leftStyle = styles.LayeredPanelStyle()
		rightStyle = styles.LayeredPanelStyle()
	}
	// Apply dimensions to both styles
	leftStyle = leftStyle.Width(leftStyle.GetWidth()).Height(leftStyle.GetHeight())
	rightStyle = rightStyle.Width(rightStyle.GetWidth()).Height(rightStyle.GetHeight())
	return leftStyle, rightStyle
}

// renderDualPane renders side-by-side panes
func (d *DualPane) renderDualPane(config LayoutConfig) string {
	// Calculate pane heights (account for borders)
	paneHeight := d.height - BorderWidth

	// Apply focus styling first to get properly configured styles
	leftStyle, rightStyle := d.applyFocusStyle(lipgloss.Style{}, lipgloss.Style{})

	// Override dimensions
	leftStyle = leftStyle.Width(config.LeftWidth).Height(paneHeight)
	rightStyle = rightStyle.Width(config.RightWidth).Height(paneHeight)

	// Render panes
	leftPane := leftStyle.Render(d.leftContent)
	rightPane := rightStyle.Render(d.rightContent)

	// Join panes horizontally with gap
	gap := strings.Repeat(" ", config.Gap)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, gap, rightPane)
}

// renderVerticalStack renders panes stacked vertically
func (d *DualPane) renderVerticalStack(config LayoutConfig) string {
	// Calculate pane heights (split available height)
	paneHeight := (d.height - BorderWidth*2) / 2 // 2 borders per pane

	// Apply focus styling first to get properly configured styles
	topStyle, bottomStyle := d.applyFocusStyle(lipgloss.Style{}, lipgloss.Style{})

	// Override dimensions
	topStyle = topStyle.Width(config.LeftWidth).Height(paneHeight)
	bottomStyle = bottomStyle.Width(config.RightWidth).Height(paneHeight)

	// Render panes
	topPane := topStyle.Render(d.leftContent)
	bottomPane := bottomStyle.Render(d.rightContent)

	// Join panes vertically with gap
	gap := strings.Repeat("\n", config.Gap)
	return lipgloss.JoinVertical(lipgloss.Left, topPane, gap, bottomPane)
}

// RenderWithContent renders the dual-pane layout with provided content
func RenderDualPane(leftContent, rightContent string, width, height, focusedPane int) string {
	dp := NewDualPane()
	dp.SetLeftContent(leftContent)
	dp.SetRightContent(rightContent)
	dp.SetDimensions(width, height)
	dp.SetFocusedPane(focusedPane)
	return dp.Render()
}

// GetLayoutType returns the layout type for a given width
func GetLayoutType(width int) LayoutType {
	config := CalculateLayout(width)
	return config.Type
}

// IsDualPane returns true if the layout is dual-pane
func IsDualPane(width int) bool {
	return GetLayoutType(width) == LayoutDualPane
}

// IsVerticalStack returns true if the layout is vertical stack
func IsVerticalStack(width int) bool {
	return GetLayoutType(width) == LayoutVerticalStack
}
