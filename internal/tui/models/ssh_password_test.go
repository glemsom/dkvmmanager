package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	tform "github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// TestNewSSHPasswordFormModel tests form initialization
func TestNewSSHPasswordFormModel(t *testing.T) {
	form := NewSSHPasswordFormModel()

	if form == nil {
		t.Fatal("NewSSHPasswordFormModel returned nil")
	}
	if form.focusIndex != 0 {
		t.Errorf("Initial focusIndex = %d, want 0", form.focusIndex)
	}
	if form.newPassword != "" {
		t.Errorf("Initial newPassword should be empty, got %q", form.newPassword)
	}
	if form.confirmPassword != "" {
		t.Errorf("Initial confirmPassword should be empty, got %q", form.confirmPassword)
	}
}

// TestSSHPasswordFormRebuildPositions tests correct number of positions
func TestSSHPasswordFormRebuildPositions(t *testing.T) {
	form := NewSSHPasswordFormModel()

	// 2 text fields + 1 apply button = 3
	if len(form.positions) != 3 {
		t.Errorf("Expected 3 positions, got %d", len(form.positions))
	}
	if form.positions[0].Key != "newPassword" {
		t.Errorf("Position 0 field = %q, want newPassword", form.positions[0].Key)
	}
	if form.positions[1].Key != "confirmPassword" {
		t.Errorf("Position 1 field = %q, want confirmPassword", form.positions[1].Key)
	}
	if form.positions[2].Kind != tform.FocusButton {
		t.Errorf("Position 2 kind = %d, want FocusButton", form.positions[2].Kind)
	}
}

// TestSSHPasswordFormCharInput tests typing characters into password fields
func TestSSHPasswordFormCharInput(t *testing.T) {
	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Focus is on newPassword (index 0)
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	form = model.(*SSHPasswordFormModel)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	form = model.(*SSHPasswordFormModel)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	form = model.(*SSHPasswordFormModel)

	if form.newPassword != "sec" {
		t.Errorf("newPassword = %q, want %q", form.newPassword, "sec")
	}
}

// TestSSHPasswordFormBackspace tests backspace in password fields
func TestSSHPasswordFormBackspace(t *testing.T) {
	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Type "abc"
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	form = model.(*SSHPasswordFormModel)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	form = model.(*SSHPasswordFormModel)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	form = model.(*SSHPasswordFormModel)

	// Backspace removes 'c'
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	form = model.(*SSHPasswordFormModel)

	if form.newPassword != "ab" {
		t.Errorf("newPassword = %q, want %q after backspace", form.newPassword, "ab")
	}
}

// TestSSHPasswordFormNavigation tests Tab navigation between fields
func TestSSHPasswordFormNavigation(t *testing.T) {
	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Start at newPassword (index 0)
	if form.currentPos().fieldName != "newPassword" {
		t.Fatalf("Expected initial field newPassword, got %s", form.currentPos().fieldName)
	}

	// Tab moves to confirmPassword (index 1)
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*SSHPasswordFormModel)
	if form.currentPos().fieldName != "confirmPassword" {
		t.Errorf("Expected confirmPassword after Tab, got %s", form.currentPos().fieldName)
	}

	// Tab moves to apply (index 2)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*SSHPasswordFormModel)
	if form.currentPos().kind != sshPwApply {
		t.Errorf("Expected apply button after Tab, got kind %d", form.currentPos().kind)
	}

	// Shift+Tab moves back to confirmPassword
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	form = model.(*SSHPasswordFormModel)
	if form.currentPos().fieldName != "confirmPassword" {
		t.Errorf("Expected confirmPassword after Shift+Tab, got %s", form.currentPos().fieldName)
	}
}

// TestSSHPasswordFormDelete tests delete key in password fields
func TestSSHPasswordFormDelete(t *testing.T) {
	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Type "abc"
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	form = model.(*SSHPasswordFormModel)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	form = model.(*SSHPasswordFormModel)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	form = model.(*SSHPasswordFormModel)

	// Move cursor to start
	form.setCursorOffset("newPassword", 0)

	// Delete removes 'a'
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyDelete})
	form = model.(*SSHPasswordFormModel)

	if form.newPassword != "bc" {
		t.Errorf("newPassword = %q, want %q after delete", form.newPassword, "bc")
	}
}

// TestSSHPasswordValidationEmptyFields tests validation with empty fields
func TestSSHPasswordValidationEmptyFields(t *testing.T) {
	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate to apply and press Enter
	form.focusIndex = len(form.positions) - 1
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*SSHPasswordFormModel)

	if _, ok := form.errors["newPassword"]; !ok {
		t.Error("Expected error for empty newPassword")
	}
	if _, ok := form.errors["confirmPassword"]; !ok {
		t.Error("Expected error for empty confirmPassword")
	}
}

// TestSSHPasswordValidationMismatch tests validation with mismatched passwords
func TestSSHPasswordValidationMismatch(t *testing.T) {
	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	form.newPassword = "password123"
	form.confirmPassword = "different"
	form.focusIndex = len(form.positions) - 1

	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*SSHPasswordFormModel)

	if errMsg, ok := form.errors["confirmPassword"]; !ok {
		t.Error("Expected error for mismatched passwords")
	} else if errMsg != "Passwords do not match" {
		t.Errorf("Error message = %q, want %q", errMsg, "Passwords do not match")
	}
}

// TestSSHPasswordValidationTooShort tests validation with short password
func TestSSHPasswordValidationTooShort(t *testing.T) {
	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	form.newPassword = "abc"
	form.confirmPassword = "abc"
	form.focusIndex = len(form.positions) - 1

	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*SSHPasswordFormModel)

	if errMsg, ok := form.errors["newPassword"]; !ok {
		t.Error("Expected error for short password")
	} else if errMsg != "Password must be at least 6 characters" {
		t.Errorf("Error message = %q, want %q", errMsg, "Password must be at least 6 characters")
	}
}

// TestSSHPasswordValidationSuccess tests that valid input passes validation
func TestSSHPasswordValidationSuccess(t *testing.T) {
	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	form.newPassword = "validpass"
	form.confirmPassword = "validpass"
	form.focusIndex = len(form.positions) - 1

	model, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*SSHPasswordFormModel)

	if len(form.errors) != 0 {
		t.Errorf("Expected no errors, got %v", form.errors)
	}
	if cmd == nil {
		t.Fatal("Expected command after valid apply, got nil")
	}
	if !form.applying {
		t.Error("Expected applying flag to be true")
	}
}

// TestSSHPasswordStrengthWeak tests weak password strength
func TestSSHPasswordStrengthWeak(t *testing.T) {
	score := passwordStrength("abc")
	if score > 1 {
		t.Errorf("passwordStrength(%q) = %d, want <= 1", "abc", score)
	}
	label, _ := strengthLabel(score)
	if label != "Weak" {
		t.Errorf("strengthLabel(%d) = %q, want %q", score, label, "Weak")
	}
}

// TestSSHPasswordStrengthFair tests fair password strength
func TestSSHPasswordStrengthFair(t *testing.T) {
	score := passwordStrength("abcdefgh")
	label, _ := strengthLabel(score)
	if label == "Strong" {
		t.Errorf("strengthLabel(%d) = %q, should not be Strong for simple password", score, label)
	}
}

// TestSSHPasswordStrengthStrong tests strong password strength
func TestSSHPasswordStrengthStrong(t *testing.T) {
	score := passwordStrength("MyP@ssw0rd!")
	if score < 4 {
		t.Errorf("passwordStrength(complex) = %d, want >= 4", score)
	}
	label, _ := strengthLabel(score)
	if label != "Strong" {
		t.Errorf("strengthLabel(%d) = %q, want %q", score, label, "Strong")
	}
}

// TestPasswordMasking tests that passwords are masked in the view
func TestPasswordMasking(t *testing.T) {
	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Type some characters
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	form = model.(*SSHPasswordFormModel)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	form = model.(*SSHPasswordFormModel)
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	form = model.(*SSHPasswordFormModel)

	view := form.View()

	// The password characters should be masked with asterisks, not visible as "sec"
	if strings.Contains(view, "sec") {
		t.Error("View should not contain plaintext password")
	}
	if !strings.Contains(view, "*") {
		t.Error("View should contain asterisk masking characters")
	}
}

// TestSSHPasswordFieldLabels tests that fields have correct labels
func TestSSHPasswordFieldLabels(t *testing.T) {
	form := NewSSHPasswordFormModel()

	if label := form.fieldLabel("newPassword"); label != "New Password" {
		t.Errorf("newPassword label = %q, want %q", label, "New Password")
	}
	if label := form.fieldLabel("confirmPassword"); label != "Confirm Password" {
		t.Errorf("confirmPassword label = %q, want %q", label, "Confirm Password")
	}
}

// TestSSHPasswordFormWindowSize tests viewport initialization
func TestSSHPasswordFormWindowSize(t *testing.T) {
	form := NewSSHPasswordFormModel()

	if form.ready {
		t.Error("Form should not be ready before WindowSizeMsg")
	}

	model, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	form = model.(*SSHPasswordFormModel)

	if !form.ready {
		t.Error("Form should be ready after WindowSizeMsg")
	}
	if form.vp.Width != 80 {
		t.Errorf("Viewport width = %d, want 80", form.vp.Width)
	}
	if form.vp.Height != 24 {
		t.Errorf("Viewport height = %d, want 24", form.vp.Height)
	}
}

// TestSSHPasswordFormSetSize tests the SetSize method
func TestSSHPasswordFormSetSize(t *testing.T) {
	form := NewSSHPasswordFormModel()

	form.SetSize(100, 30)

	if form.contentW != 100 {
		t.Errorf("contentW = %d, want 100", form.contentW)
	}
	if form.contentH != 30 {
		t.Errorf("contentH = %d, want 30", form.contentH)
	}
	if !form.ready {
		t.Error("Form should be ready after SetSize")
	}

	// Resize again
	form.SetSize(120, 40)
	if form.vp.Width != 120 {
		t.Errorf("Viewport width = %d, want 120 after resize", form.vp.Width)
	}
}

// TestSSHPasswordFormESCWhileApplying tests that ESC is ignored during apply
func TestSSHPasswordFormESCWhileApplying(t *testing.T) {
	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	form.newPassword = "validpass"
	form.confirmPassword = "validpass"
	form.focusIndex = len(form.positions) - 1
	form.applying = true

	// Key input should be ignored while applying
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyEscape})
	form = model.(*SSHPasswordFormModel)

	if !form.applying {
		t.Error("Applying flag should still be true")
	}
}

// TestSSHPasswordFormEmptyViewBeforeReady tests View before ready
func TestSSHPasswordFormEmptyViewBeforeReady(t *testing.T) {
	form := NewSSHPasswordFormModel()

	view := form.View()
	if view != "Loading form..." {
		t.Errorf("View before ready = %q, want %q", view, "Loading form...")
	}
}

// --- Wrapper model tests ---

// TestNewSSHPasswordModel tests wrapper model creation
func TestNewSSHPasswordModel(t *testing.T) {
	m := NewSSHPasswordModel()

	if m == nil {
		t.Fatal("NewSSHPasswordModel returned nil")
	}
	if m.form == nil {
		t.Fatal("Expected non-nil form")
	}
}

// TestSSHPasswordModelInit tests Init delegation
func TestSSHPasswordModelInit(t *testing.T) {
	m := NewSSHPasswordModel()

	cmd := m.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

// TestSSHPasswordModelUpdate tests Update delegation
func TestSSHPasswordModelUpdate(t *testing.T) {
	m := NewSSHPasswordModel()

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if updated == nil {
		t.Fatal("Update() returned nil model")
	}

	wrapped, ok := updated.(*SSHPasswordModel)
	if !ok {
		t.Fatalf("Update() should return *SSHPasswordModel, got %T", updated)
	}
	if wrapped.form == nil {
		t.Error("Wrapped model form should not be nil after Update()")
	}
	if !wrapped.form.Ready() {
		t.Error("Form should be ready after WindowSizeMsg")
	}
	_ = cmd
}

// TestSSHPasswordModelView tests View delegation
func TestSSHPasswordModelView(t *testing.T) {
	m := NewSSHPasswordModel()

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = updated.(*SSHPasswordModel)

	view := m.View()
	if view == "" {
		t.Error("View() should return non-empty string")
	}
}

// TestSSHPasswordModelUpdateDelegatesKeyPress tests key delegation
func TestSSHPasswordModelUpdateDelegatesKeyPress(t *testing.T) {
	m := NewSSHPasswordModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m = updated.(*SSHPasswordModel)

	initialField := m.Form().currentPos().fieldName

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(*SSHPasswordModel)

	newField := m.Form().currentPos().fieldName
	if newField == initialField {
		t.Error("Key press (Tab) should be delegated to form, focus should change")
	}
}

// TestSSHPasswordFormApplyDryRun tests dry-run mode
func TestSSHPasswordFormApplyDryRun(t *testing.T) {
	origDryRun := dryRunMode
	dryRunMode = true
	defer func() { dryRunMode = origDryRun }()

	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	form.newPassword = "validpass"
	form.confirmPassword = "validpass"
	form.focusIndex = len(form.positions) - 1

	model, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	form = model.(*SSHPasswordFormModel)

	if cmd == nil {
		t.Fatal("Expected command after valid apply, got nil")
	}

	msg := cmd()
	if _, ok := msg.(SSHPasswordUpdatedMsg); !ok {
		t.Errorf("Expected SSHPasswordUpdatedMsg in dry-run, got %T", msg)
	}
}

// TestSSHPasswordFormApplyInvalidChars tests handling of special characters in passwords
func TestSSHPasswordFormApplyInvalidChars(t *testing.T) {
	origDryRun := dryRunMode
	dryRunMode = true
	defer func() { dryRunMode = origDryRun }()

	form := NewSSHPasswordFormModel()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Special characters should be accepted
	form.newPassword = "p@ss!w0rd#"
	form.confirmPassword = "p@ss!w0rd#"
	form.focusIndex = len(form.positions) - 1

	model, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = model.(*SSHPasswordFormModel)

	if cmd == nil {
		t.Fatal("Expected command after valid apply with special chars, got nil")
	}
}

// TestSSHPasswordStrengthScoring tests various password strength scores
func TestSSHPasswordStrengthScoring(t *testing.T) {
	tests := []struct {
		password string
		maxScore int
	}{
		{"", 0},
		{"a", 1},
		{"abcdefgh", 3},
		{"Abcdefgh", 4},
		{"Abcdefg1", 5},
		{"Abcdefg1!", 5},
	}

	for _, tt := range tests {
		score := passwordStrength(tt.password)
		if score > tt.maxScore {
			t.Errorf("passwordStrength(%q) = %d, want <= %d", tt.password, score, tt.maxScore)
		}
	}
}

// TestStrengthLabelBoundaries tests strength label at boundary values
func TestStrengthLabelBoundaries(t *testing.T) {
	tests := []struct {
		score       int
		expectedLbl string
	}{
		{0, "Weak"},
		{1, "Weak"},
		{2, "Fair"},
		{3, "Fair"},
		{4, "Strong"},
		{5, "Strong"},
	}

	for _, tt := range tests {
		label, _ := strengthLabel(tt.score)
		if label != tt.expectedLbl {
			t.Errorf("strengthLabel(%d) = %q, want %q", tt.score, label, tt.expectedLbl)
		}
	}
}

// TestSSHPasswordCursorOffset tests cursor offset management
func TestSSHPasswordCursorOffset(t *testing.T) {
	form := NewSSHPasswordFormModel()

	// Default offset should be -1 (end)
	if off := form.cursorOffset("newPassword"); off != -1 {
		t.Errorf("Default cursorOffset = %d, want -1", off)
	}

	// Set and get
	form.setCursorOffset("newPassword", 3)
	if off := form.cursorOffset("newPassword"); off != 3 {
		t.Errorf("cursorOffset after set = %d, want 3", off)
	}
}

// TestSSHPasswordEffectiveCursor tests effective cursor calculation
func TestSSHPasswordEffectiveCursor(t *testing.T) {
	form := NewSSHPasswordFormModel()

	// Offset -1 means end
	form.newPassword = "hello"
	form.setCursorOffset("newPassword", -1)
	if c := form.effectiveCursor("newPassword", form.newPassword); c != 5 {
		t.Errorf("effectiveCursor with offset -1 = %d, want 5", c)
	}

	// Offset beyond value length clamps to end
	form.setCursorOffset("newPassword", 100)
	if c := form.effectiveCursor("newPassword", form.newPassword); c != 5 {
		t.Errorf("effectiveCursor with offset 100 = %d, want 5", c)
	}

	// Normal offset
	form.setCursorOffset("newPassword", 2)
	if c := form.effectiveCursor("newPassword", form.newPassword); c != 2 {
		t.Errorf("effectiveCursor with offset 2 = %d, want 2", c)
	}
}
