package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestHighlightStyle(t *testing.T) {
	style := HighlightStyle()

	if !style.GetBold() {
		t.Error("HighlightStyle() should be bold")
	}

	if style.GetForeground() != Colors.Primary {
		t.Errorf("HighlightStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestAccentStyle(t *testing.T) {
	style := AccentStyle()

	if style.GetForeground() != Colors.Primary {
		t.Errorf("AccentStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestInverseStyle(t *testing.T) {
	style := InverseStyle()

	if style.GetBackground() != Colors.Primary {
		t.Errorf("InverseStyle() background = %v, want %v", style.GetBackground(), Colors.Primary)
	}

	if style.GetForeground() != Colors.Background {
		t.Errorf("InverseStyle() foreground = %v, want %v", style.GetForeground(), Colors.Background)
	}
}

func TestDimStyle(t *testing.T) {
	style := DimStyle()

	if style.GetForeground() != Colors.Muted {
		t.Errorf("DimStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestItalicStyle(t *testing.T) {
	style := ItalicStyle()

	if !style.GetItalic() {
		t.Error("ItalicStyle() should be italic")
	}
}

func TestUnderlineStyle(t *testing.T) {
	style := UnderlineStyle()

	if !style.GetUnderline() {
		t.Error("UnderlineStyle() should be underlined")
	}
}

func TestStrikethroughStyle(t *testing.T) {
	style := StrikethroughStyle()

	if !style.GetStrikethrough() {
		t.Error("StrikethroughStyle() should have strikethrough")
	}

	if style.GetForeground() != Colors.Muted {
		t.Errorf("StrikethroughStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestBadgeStyle(t *testing.T) {
	style := BadgeStyle()

	if !style.GetBold() {
		t.Error("BadgeStyle() should be bold")
	}

	if style.GetBackground() != Colors.Primary {
		t.Errorf("BadgeStyle() background = %v, want %v", style.GetBackground(), Colors.Primary)
	}

	if style.GetForeground() != Colors.Background {
		t.Errorf("BadgeStyle() foreground = %v, want %v", style.GetForeground(), Colors.Background)
	}
}

func TestTagStyle(t *testing.T) {
	style := TagStyle()

	if style.GetForeground() != Colors.Primary {
		t.Errorf("TagStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestSeparatorStyle(t *testing.T) {
	style := SeparatorStyle()

	if style.GetForeground() != Colors.Muted {
		t.Errorf("SeparatorStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestProgressBarStyle(t *testing.T) {
	style := ProgressBarStyle()

	if style.GetBackground() != Colors.Muted {
		t.Errorf("ProgressBarStyle() background = %v, want %v", style.GetBackground(), Colors.Muted)
	}
}

func TestProgressBarFillStyle(t *testing.T) {
	style := ProgressBarFillStyle()

	if style.GetBackground() != Colors.Primary {
		t.Errorf("ProgressBarFillStyle() background = %v, want %v", style.GetBackground(), Colors.Primary)
	}
}

func TestInputStyle(t *testing.T) {
	style := InputStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("InputStyle() should render content")
	}
}

func TestInputFocusedStyle(t *testing.T) {
	style := InputFocusedStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("InputFocusedStyle() should render content")
	}
}

func TestInputErrorStyle(t *testing.T) {
	style := InputErrorStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("InputErrorStyle() should render content")
	}
}

func TestTableHeaderStyle(t *testing.T) {
	style := TableHeaderStyle()

	if !style.GetBold() {
		t.Error("TableHeaderStyle() should be bold")
	}

	if style.GetForeground() != Colors.Primary {
		t.Errorf("TableHeaderStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestTableRowStyle(t *testing.T) {
	style := TableRowStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("TableRowStyle() should render content")
	}
}

func TestTableSelectedRowStyle(t *testing.T) {
	style := TableSelectedRowStyle()

	if !style.GetBold() {
		t.Error("TableSelectedRowStyle() should be bold")
	}

	if style.GetForeground() != Colors.Primary {
		t.Errorf("TableSelectedRowStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestTooltipStyle(t *testing.T) {
	style := TooltipStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("TooltipStyle() should render content")
	}
}

func TestModalStyle(t *testing.T) {
	style := ModalStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("ModalStyle() should render content")
	}
}

func TestNotificationStyle(t *testing.T) {
	style := NotificationStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("NotificationStyle() should render content")
	}
}

func TestErrorNotificationStyle(t *testing.T) {
	style := ErrorNotificationStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("ErrorNotificationStyle() should render content")
	}
}

func TestSuccessNotificationStyle(t *testing.T) {
	style := SuccessNotificationStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("SuccessNotificationStyle() should render content")
	}
}

func TestWarningNotificationStyle(t *testing.T) {
	style := WarningNotificationStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("WarningNotificationStyle() should render content")
	}
}

func TestScrollbarStyle(t *testing.T) {
	style := ScrollbarStyle()

	if style.GetForeground() != Colors.Muted {
		t.Errorf("ScrollbarStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestScrollbarThumbStyle(t *testing.T) {
	style := ScrollbarThumbStyle()

	if style.GetForeground() != Colors.Primary {
		t.Errorf("ScrollbarThumbStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestLinkStyle(t *testing.T) {
	style := LinkStyle()

	if !style.GetUnderline() {
		t.Error("LinkStyle() should be underlined")
	}

	if style.GetForeground() != Colors.Primary {
		t.Errorf("LinkStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestCodeStyle(t *testing.T) {
	style := CodeStyle()

	if style.GetForeground() != Colors.Secondary {
		t.Errorf("CodeStyle() foreground = %v, want %v", style.GetForeground(), Colors.Secondary)
	}
}

func TestQuoteStyle(t *testing.T) {
	style := QuoteStyle()

	if !style.GetItalic() {
		t.Error("QuoteStyle() should be italic")
	}

	if style.GetForeground() != Colors.Muted {
		t.Errorf("QuoteStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestHeadingStyle(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		expected lipgloss.Color
	}{
		{"Level 1", 1, Colors.Primary},
		{"Level 2", 2, Colors.Secondary},
		{"Level 3", 3, lipgloss.Color("252")},
		{"Default", 4, Colors.Primary},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := HeadingStyle(tt.level)

			if !style.GetBold() {
				t.Errorf("HeadingStyle(%d) should be bold", tt.level)
			}

			if style.GetForeground() != tt.expected {
				t.Errorf("HeadingStyle(%d) foreground = %v, want %v", tt.level, style.GetForeground(), tt.expected)
			}
		})
	}
}

func TestCaptionStyle(t *testing.T) {
	style := CaptionStyle()

	if !style.GetItalic() {
		t.Error("CaptionStyle() should be italic")
	}

	if style.GetForeground() != Colors.Muted {
		t.Errorf("CaptionStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestTimestampStyle(t *testing.T) {
	style := TimestampStyle()

	if style.GetForeground() != Colors.Muted {
		t.Errorf("TimestampStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestCounterStyle(t *testing.T) {
	style := CounterStyle()

	if !style.GetBold() {
		t.Error("CounterStyle() should be bold")
	}

	if style.GetForeground() != Colors.Primary {
		t.Errorf("CounterStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestDividerStyle(t *testing.T) {
	style := DividerStyle()

	if style.GetForeground() != Colors.Muted {
		t.Errorf("DividerStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestRenderDivider(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		expected bool
	}{
		{"Positive width", 10, true},
		{"Zero width", 0, false},
		{"Negative width", -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderDivider(tt.width)
			if tt.expected && result == "" {
				t.Errorf("RenderDivider(%d) should return non-empty string", tt.width)
			}
			if !tt.expected && result != "" {
				t.Errorf("RenderDivider(%d) should return empty string", tt.width)
			}
		})
	}
}

func TestRenderVerticalDivider(t *testing.T) {
	tests := []struct {
		name     string
		height   int
		expected bool
	}{
		{"Positive height", 5, true},
		{"Zero height", 0, false},
		{"Negative height", -3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderVerticalDivider(tt.height)
			if tt.expected && result == "" {
				t.Errorf("RenderVerticalDivider(%d) should return non-empty string", tt.height)
			}
			if !tt.expected && result != "" {
				t.Errorf("RenderVerticalDivider(%d) should return empty string", tt.height)
			}
		})
	}
}

func TestStylesRendering(t *testing.T) {
	// Test that all styles can render without panicking
	testText := "Test Content"

	styles := []struct {
		name string
		fn   func() lipgloss.Style
	}{
		{"HighlightStyle", HighlightStyle},
		{"AccentStyle", AccentStyle},
		{"InverseStyle", InverseStyle},
		{"DimStyle", DimStyle},
		{"ItalicStyle", ItalicStyle},
		{"UnderlineStyle", UnderlineStyle},
		{"StrikethroughStyle", StrikethroughStyle},
		{"BadgeStyle", BadgeStyle},
		{"TagStyle", TagStyle},
		{"SeparatorStyle", SeparatorStyle},
		{"ProgressBarStyle", ProgressBarStyle},
		{"ProgressBarFillStyle", ProgressBarFillStyle},
		{"InputStyle", InputStyle},
		{"InputFocusedStyle", InputFocusedStyle},
		{"InputErrorStyle", InputErrorStyle},
		{"TableHeaderStyle", TableHeaderStyle},
		{"TableRowStyle", TableRowStyle},
		{"TableSelectedRowStyle", TableSelectedRowStyle},
		{"TooltipStyle", TooltipStyle},
		{"ModalStyle", ModalStyle},
		{"NotificationStyle", NotificationStyle},
		{"ErrorNotificationStyle", ErrorNotificationStyle},
		{"SuccessNotificationStyle", SuccessNotificationStyle},
		{"WarningNotificationStyle", WarningNotificationStyle},
		{"ScrollbarStyle", ScrollbarStyle},
		{"ScrollbarThumbStyle", ScrollbarThumbStyle},
		{"LinkStyle", LinkStyle},
		{"CodeStyle", CodeStyle},
		{"QuoteStyle", QuoteStyle},
		{"CaptionStyle", CaptionStyle},
		{"TimestampStyle", TimestampStyle},
		{"CounterStyle", CounterStyle},
		{"DividerStyle", DividerStyle},
	}

	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s() panicked: %v", s.name, r)
				}
			}()

			style := s.fn()
			rendered := style.Render(testText)
			if rendered == "" {
				t.Errorf("%s() returned empty string", s.name)
			}
		})
	}
}
