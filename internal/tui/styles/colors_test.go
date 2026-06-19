package styles

import (
	"image/color"
	"os"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func isLinuxTerm() bool {
	term := os.Getenv("TERM")
	return term == "linux" || strings.HasPrefix(term, "linux-")
}

func mutedColorExpected() string {
	if isLinuxTerm() {
		return "7"
	}
	return "8"
}

func brightToBaseExpected(bright, base string) string {
	if isLinuxTerm() {
		return base
	}
	return bright
}

func TestColorPalette(t *testing.T) {
	tests := []struct {
		name     string
		color    color.Color
		expected string
	}{
		{"Primary", Colors.Primary, "6"},
		{"PrimaryDim", Colors.PrimaryDim, brightToBaseExpected("14", "4")},
		{"Secondary", Colors.Secondary, brightToBaseExpected("13", "5")},
		{"Success", Colors.Success, brightToBaseExpected("10", "2")},
		{"Warning", Colors.Warning, brightToBaseExpected("11", "3")},
		{"Error", Colors.Error, brightToBaseExpected("9", "1")},
		{"Muted", Colors.Muted, mutedColorExpected()},
		{"Background", Colors.Background, "0"},
		{"Border", Colors.Border, "8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color != lipgloss.Color(tt.expected) {
				t.Errorf("Colors.%s = %v, want %s", tt.name, tt.color, tt.expected)
			}
		})
	}
}

func TestStatusColors(t *testing.T) {
	tests := []struct {
		name     string
		color    color.Color
		expected string
	}{
		{"Running", StatusColors.Running, brightToBaseExpected("10", "2")},
		{"Stopped", StatusColors.Stopped, mutedColorExpected()},
		{"Error", StatusColors.Error, brightToBaseExpected("9", "1")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color != lipgloss.Color(tt.expected) {
				t.Errorf("StatusColors.%s = %v, want %s", tt.name, tt.color, tt.expected)
			}
		})
	}
}

func TestLayeredPanelStyle(t *testing.T) {
	style := LayeredPanelStyle()

	// Verify the style has a border
	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("LayeredPanelStyle() should render content")
	}

	// Verify the style has padding
	if style.GetPaddingLeft() != 2 {
		t.Errorf("LayeredPanelStyle() padding left = %d, want 2", style.GetPaddingLeft())
	}
	if style.GetPaddingTop() != 1 {
		t.Errorf("LayeredPanelStyle() padding top = %d, want 1", style.GetPaddingTop())
	}
}

func TestActiveLayeredPanelStyle(t *testing.T) {
	style := ActiveLayeredPanelStyle()

	// Verify the style has a border
	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("ActiveLayeredPanelStyle() should render content")
	}

	// Verify it has a border (different from inactive)
	if !strings.Contains(rendered, "─") && !strings.Contains(rendered, "┌") {
		t.Error("ActiveLayeredPanelStyle() should have border characters")
	}
}

func TestPanelWithTitleStyle(t *testing.T) {
	title := "Test Panel"
	style := PanelWithTitleStyle(title)

	rendered := style.Render(title)
	if rendered == "" {
		t.Error("PanelWithTitleStyle() should render content")
	}
}

func TestTextStyles(t *testing.T) {
	tests := []struct {
		name     string
		styleFn  func() lipgloss.Style
		testText string
	}{
		{"NormalTextStyle", NormalTextStyle, "Normal text"},
		{"SelectedTextStyle", SelectedTextStyle, "Selected text"},
		{"FocusedTextStyle", FocusedTextStyle, "Focused text"},
		{"DisabledTextStyle", DisabledTextStyle, "Disabled text"},
		{"TitleStyle", TitleStyle, "Title"},
		{"SubtitleStyle", SubtitleStyle, "Subtitle"},
		{"ErrorTextStyle", ErrorTextStyle, "Error text"},
		{"WarningTextStyle", WarningTextStyle, "Warning text"},
		{"SuccessTextStyle", SuccessTextStyle, "Success text"},
		{"MutedTextStyle", MutedTextStyle, "Muted text"},
		{"HeaderStyle", HeaderStyle, "Header"},
		{"FooterStyle", FooterStyle, "Footer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := tt.styleFn()
			rendered := style.Render(tt.testText)
			if rendered == "" {
				t.Errorf("%s() should render content", tt.name)
			}
		})
	}
}

func TestSelectedTextStyle(t *testing.T) {
	style := SelectedTextStyle()

	if !style.GetBold() {
		t.Error("SelectedTextStyle() should be bold")
	}

	if style.GetForeground() != Colors.Primary {
		t.Errorf("SelectedTextStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestFocusedTextStyle(t *testing.T) {
	style := FocusedTextStyle()

	if !style.GetUnderline() {
		t.Error("FocusedTextStyle() should be underlined")
	}

	if style.GetForeground() != Colors.Primary {
		t.Errorf("FocusedTextStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestDisabledTextStyle(t *testing.T) {
	style := DisabledTextStyle()

	if style.GetForeground() != Colors.Muted {
		t.Errorf("DisabledTextStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestButtonStyles(t *testing.T) {
	tests := []struct {
		name     string
		styleFn  func() lipgloss.Style
		testText string
	}{
		{"PrimaryButtonStyle", PrimaryButtonStyle, "Primary"},
		{"SecondaryButtonStyle", SecondaryButtonStyle, "Secondary"},
		{"DisabledButtonStyle", DisabledButtonStyle, "Disabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := tt.styleFn()
			rendered := style.Render(tt.testText)
			if rendered == "" {
				t.Errorf("%s() should render content", tt.name)
			}
		})
	}
}

func TestPrimaryButtonStyle(t *testing.T) {
	style := PrimaryButtonStyle()

	if !style.GetBold() {
		t.Error("PrimaryButtonStyle() should be bold")
	}

	if style.GetPaddingLeft() != 2 {
		t.Errorf("PrimaryButtonStyle() padding left = %d, want 2", style.GetPaddingLeft())
	}
}

func TestSecondaryButtonStyle(t *testing.T) {
	style := SecondaryButtonStyle()

	if !style.GetBold() {
		t.Error("SecondaryButtonStyle() should be bold")
	}

	if style.GetPaddingLeft() != 2 {
		t.Errorf("SecondaryButtonStyle() padding left = %d, want 2", style.GetPaddingLeft())
	}
}

func TestDisabledButtonStyle(t *testing.T) {
	style := DisabledButtonStyle()

	if style.GetBold() {
		t.Error("DisabledButtonStyle() should not be bold")
	}

	if style.GetForeground() != Colors.Muted {
		t.Errorf("DisabledButtonStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestStatusIndicator(t *testing.T) {
	tests := []struct {
		name     string
		status   string
	}{
		{"Running", "running"},
		{"Stopped", "stopped"},
		{"Error", "error"},
		{"Unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indicator := StatusIndicator(tt.status)
			if indicator == "" {
				t.Errorf("StatusIndicator(%q) should return a non-empty string", tt.status)
			}
		})
	}
}

func TestRunningStatusStyle(t *testing.T) {
	style := RunningStatusStyle()

	if !style.GetBold() {
		t.Error("RunningStatusStyle() should be bold")
	}

	// Inverted badge: background = status color, foreground = Background color
	if style.GetBackground() != StatusColors.Running {
		t.Errorf("RunningStatusStyle() background = %v, want %v", style.GetBackground(), StatusColors.Running)
	}
}

func TestStoppedStatusStyle(t *testing.T) {
	style := StoppedStatusStyle()

	if !style.GetBold() {
		t.Error("StoppedStatusStyle() should be bold")
	}

	if style.GetForeground() != Colors.ForegroundDim {
		t.Errorf("StoppedStatusStyle() foreground = %v, want %v", style.GetForeground(), Colors.ForegroundDim)
	}
}

func TestErrorStatusStyle(t *testing.T) {
	style := ErrorStatusStyle()

	if !style.GetBold() {
		t.Error("ErrorStatusStyle() should be bold")
	}

	// Inverted badge: background = status color, foreground = Background color
	if style.GetBackground() != StatusColors.Error {
		t.Errorf("ErrorStatusStyle() background = %v, want %v", style.GetBackground(), StatusColors.Error)
	}
}

func TestBorderStyle(t *testing.T) {
	style := BorderStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("BorderStyle() should render content with border")
	}
}

func TestActiveBorderStyle(t *testing.T) {
	style := ActiveBorderStyle()

	rendered := style.Render("Test")
	if rendered == "" {
		t.Error("ActiveBorderStyle() should render content with border")
	}
}

func TestListItemStyles(t *testing.T) {
	tests := []struct {
		name     string
		styleFn  func() lipgloss.Style
		testText string
	}{
		{"ListItemSelectedStyle", ListItemSelectedStyle, "Selected item"},
		{"ListItemNormalStyle", ListItemNormalStyle, "Normal item"},
		{"ListItemDisabledStyle", ListItemDisabledStyle, "Disabled item"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := tt.styleFn()
			rendered := style.Render(tt.testText)
			if rendered == "" {
				t.Errorf("%s() should render content", tt.name)
			}
		})
	}
}

func TestListItemSelectedStyle(t *testing.T) {
	style := ListItemSelectedStyle()

	if !style.GetBold() {
		t.Error("ListItemSelectedStyle() should be bold")
	}

	if style.GetForeground() != Colors.Primary {
		t.Errorf("ListItemSelectedStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestHelpStyles(t *testing.T) {
	tests := []struct {
		name     string
		styleFn  func() lipgloss.Style
		testText string
	}{
		{"HelpKeyStyle", HelpKeyStyle, "Key"},
		{"HelpDescStyle", HelpDescStyle, "Description"},
		{"HelpSeparatorStyle", HelpSeparatorStyle, " | "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := tt.styleFn()
			rendered := style.Render(tt.testText)
			if rendered == "" {
				t.Errorf("%s() should render content", tt.name)
			}
		})
	}
}

func TestHelpKeyStyle(t *testing.T) {
	style := HelpKeyStyle()

	if !style.GetBold() {
		t.Error("HelpKeyStyle() should be bold")
	}

	if style.GetForeground() != Colors.Primary {
		t.Errorf("HelpKeyStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestColorConsistency(t *testing.T) {
	// Verify that status colors use the same values as the main palette
	if StatusColors.Running != Colors.Success {
		t.Errorf("StatusColors.Running (%v) should match Colors.Success (%v)", StatusColors.Running, Colors.Success)
	}

	if StatusColors.Stopped != Colors.Muted {
		t.Errorf("StatusColors.Stopped (%v) should match Colors.Muted (%v)", StatusColors.Stopped, Colors.Muted)
	}

	if StatusColors.Error != Colors.Error {
		t.Errorf("StatusColors.Error (%v) should match Colors.Error (%v)", StatusColors.Error, Colors.Error)
	}
}

func TestStyleRendering(t *testing.T) {
	// Test that all styles can render without panicking
	testText := "Test Content"

	styles := []struct {
		name string
		fn   func() lipgloss.Style
	}{
		{"LayeredPanelStyle", LayeredPanelStyle},
		{"ActiveLayeredPanelStyle", ActiveLayeredPanelStyle},
		{"NormalTextStyle", NormalTextStyle},
		{"SelectedTextStyle", SelectedTextStyle},
		{"FocusedTextStyle", FocusedTextStyle},
		{"DisabledTextStyle", DisabledTextStyle},
		{"PrimaryButtonStyle", PrimaryButtonStyle},
		{"SecondaryButtonStyle", SecondaryButtonStyle},
		{"DisabledButtonStyle", DisabledButtonStyle},
		{"RunningStatusStyle", RunningStatusStyle},
		{"StoppedStatusStyle", StoppedStatusStyle},
		{"ErrorStatusStyle", ErrorStatusStyle},
		{"TitleStyle", TitleStyle},
		{"SubtitleStyle", SubtitleStyle},
		{"ErrorTextStyle", ErrorTextStyle},
		{"WarningTextStyle", WarningTextStyle},
		{"SuccessTextStyle", SuccessTextStyle},
		{"MutedTextStyle", MutedTextStyle},
		{"HeaderStyle", HeaderStyle},
		{"FooterStyle", FooterStyle},
		{"BorderStyle", BorderStyle},
		{"ActiveBorderStyle", ActiveBorderStyle},
		{"ListItemSelectedStyle", ListItemSelectedStyle},
		{"ListItemNormalStyle", ListItemNormalStyle},
		{"ListItemDisabledStyle", ListItemDisabledStyle},
		{"HelpKeyStyle", HelpKeyStyle},
		{"HelpDescStyle", HelpDescStyle},
		{"HelpSeparatorStyle", HelpSeparatorStyle},
		{"FormFocusStyle", FormFocusStyle},
		{"FormSaveStyle", FormSaveStyle},
		{"FormLabelStyle", FormLabelStyle},
		{"FormInputStyle", FormInputStyle},
		{"FormMutedStyle", FormMutedStyle},
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

func TestStatusIndicatorRendering(t *testing.T) {
	// Test that status indicators render without panicking
	statuses := []string{"running", "stopped", "error", "unknown", ""}

	for _, status := range statuses {
		t.Run("Status_"+status, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("StatusIndicator(%q) panicked: %v", status, r)
				}
			}()

			indicator := StatusIndicator(status)
			if indicator == "" {
				t.Errorf("StatusIndicator(%q) returned empty string", status)
			}
		})
	}
}

func TestFormFocusStyle(t *testing.T) {
	style := FormFocusStyle()

	if !style.GetBold() {
		t.Error("FormFocusStyle() should be bold")
	}

	if style.GetForeground() != Colors.Primary {
		t.Errorf("FormFocusStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestFormSaveStyle(t *testing.T) {
	style := FormSaveStyle()

	if !style.GetBold() {
		t.Error("FormSaveStyle() should be bold")
	}

	if style.GetForeground() != Colors.Success {
		t.Errorf("FormSaveStyle() foreground = %v, want %v", style.GetForeground(), Colors.Success)
	}
}

func TestFormLabelStyle(t *testing.T) {
	style := FormLabelStyle()

	if style.GetForeground() != Colors.Muted {
		t.Errorf("FormLabelStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}

func TestFormInputStyle(t *testing.T) {
	style := FormInputStyle()

	if style.GetForeground() != Colors.Primary {
		t.Errorf("FormInputStyle() foreground = %v, want %v", style.GetForeground(), Colors.Primary)
	}
}

func TestFormMutedStyle(t *testing.T) {
	style := FormMutedStyle()

	if style.GetForeground() != Colors.Muted {
		t.Errorf("FormMutedStyle() foreground = %v, want %v", style.GetForeground(), Colors.Muted)
	}
}
