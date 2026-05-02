package form

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Mock FormModel for testing ---

type mockFormModel struct {
	positions    []FocusPos
	focusIndex   int
	header       string
	footer       string
	positionOut  string
	enterResult  FormResult
	enterCmd     tea.Cmd
	setSizeW     int
	setSizeH     int
	setFocused   *bool
	onEnterCalled  bool
	onExitCalled   bool
	handleCharPos  FocusPos
	handleCharCh   string
	handleBsPos    FocusPos
	handleDelPos   FocusPos
}

func (m *mockFormModel) BuildPositions() []FocusPos {
	if m.positions == nil {
		return []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
			{Kind: FocusText, Label: "Email", Key: "email"},
			{Kind: FocusButton, Label: "Save", Key: "save"},
		}
	}
	return m.positions
}

func (m *mockFormModel) CurrentIndex() int             { return m.focusIndex }
func (m *mockFormModel) SetFocusIndex(i int)            { m.focusIndex = i }
func (m *mockFormModel) RenderHeader() string           { return m.header }
func (m *mockFormModel) RenderFooter() string           { return m.footer }
func (m *mockFormModel) RenderPosition(pos FocusPos, focused bool, cursorOffset int) string {
	return m.positionOut
}
func (m *mockFormModel) HandleEnter(pos FocusPos) (FormResult, tea.Cmd) {
	return m.enterResult, m.enterCmd
}
func (m *mockFormModel) HandleChar(pos FocusPos, ch string) {
	m.handleCharPos = pos
	m.handleCharCh = ch
}
func (m *mockFormModel) HandleBackspace(pos FocusPos) {
	m.handleBsPos = pos
}
func (m *mockFormModel) HandleDelete(pos FocusPos) {
	m.handleDelPos = pos
}
func (m *mockFormModel) OnEnter()               { m.onEnterCalled = true }
func (m *mockFormModel) OnExit()                { m.onExitCalled = true }
func (m *mockFormModel) SetSize(w, h int)       { m.setSizeW = w; m.setSizeH = h }
func (m *mockFormModel) SetFocused(f bool)      { m.setFocused = &f }

// --- Tracer Bullet: ScrollableForm viewport initialization ---

func TestScrollableForm_WindowSizeMsg_InitializesViewport(t *testing.T) {
	mock := &mockFormModel{}
	sf := NewScrollableForm(mock)

	// Before WindowSizeMsg, viewport is not ready
	if sf.Ready() {
		t.Fatal("expected viewport to not be ready before WindowSizeMsg")
	}

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}
	result, cmd := sf.Update(msg)
	sf = result.(*ScrollableForm)

	// After WindowSizeMsg, viewport is ready
	if !sf.Ready() {
		t.Fatal("expected viewport to be ready after WindowSizeMsg")
	}

	// Size was propagated to the form model
	if mock.setSizeW != 80 {
		t.Errorf("expected SetSize width 80, got %d", mock.setSizeW)
	}
	if mock.setSizeH != 24 {
		t.Errorf("expected SetSize height 24, got %d", mock.setSizeH)
	}

	// No command should be returned
	if cmd != nil {
		t.Errorf("expected nil cmd from WindowSizeMsg, got %v", cmd)
	}
}

func TestScrollableForm_WindowSizeMsg_SingleSet(t *testing.T) {
	// Only the first WindowSizeMsg should initialize the viewport
	// Subsequent ones should just update dimensions
	mock := &mockFormModel{}
	sf := NewScrollableForm(mock)

	// First WindowSizeMsg
	msg1 := tea.WindowSizeMsg{Width: 80, Height: 24}
	result, _ := sf.Update(msg1)
	sf = result.(*ScrollableForm)

	if !sf.Ready() {
		t.Fatal("expected ready after first WindowSizeMsg")
	}

	// Second WindowSizeMsg should not reset ready state
	msg2 := tea.WindowSizeMsg{Width: 100, Height: 30}
	result, _ = sf.Update(msg2)
	sf = result.(*ScrollableForm)

	if !sf.Ready() {
		t.Fatal("expected ready to remain true after second WindowSizeMsg")
	}
}

func TestScrollableForm_SetSize(t *testing.T) {
	mock := &mockFormModel{}
	sf := NewScrollableForm(mock)

	sf.SetSize(90, 28)

	if mock.setSizeW != 90 {
		t.Errorf("expected SetSize width 90, got %d", mock.setSizeW)
	}
	if mock.setSizeH != 28 {
		t.Errorf("expected SetSize height 28, got %d", mock.setSizeH)
	}

	if !sf.Ready() {
		t.Fatal("expected viewport ready after SetSize")
	}
}

// --- Focus navigation through ScrollableForm ---

func TestScrollableForm_MoveFocusDown(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
			{Kind: FocusText, Label: "Email", Key: "email"},
			{Kind: FocusButton, Label: "Save", Key: "save"},
		},
	}
	sf := NewScrollableForm(mock)

	if sf.FocusIndex() != 0 {
		t.Fatalf("expected initial focus index 0, got %d", sf.FocusIndex())
	}

	sf.MoveFocus(1)
	if sf.FocusIndex() != 1 {
		t.Errorf("expected focus index 1 after +1, got %d", sf.FocusIndex())
	}

	sf.MoveFocus(1)
	if sf.FocusIndex() != 2 {
		t.Errorf("expected focus index 2 after another +1, got %d", sf.FocusIndex())
	}

	// Clamping at end
	sf.MoveFocus(1)
	if sf.FocusIndex() != 2 {
		t.Errorf("expected focus index clamped at 2, got %d", sf.FocusIndex())
	}
}

func TestScrollableForm_MoveFocusUp(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
			{Kind: FocusText, Label: "Email", Key: "email"},
			{Kind: FocusButton, Label: "Save", Key: "save"},
		},
		focusIndex: 2,
	}
	sf := NewScrollableForm(mock)

	sf.MoveFocus(-1)
	if sf.FocusIndex() != 1 {
		t.Errorf("expected focus index 1 after -1, got %d", sf.FocusIndex())
	}

	// Clamping at start
	sf.MoveFocus(-1)
	sf.MoveFocus(-1)
	if sf.FocusIndex() != 0 {
		t.Errorf("expected focus index clamped at 0, got %d", sf.FocusIndex())
	}
}

// --- Key dispatching ---

func TestScrollableForm_TabKey_MovesFocusForward(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
			{Kind: FocusText, Label: "Email", Key: "email"},
			{Kind: FocusButton, Label: "Save", Key: "save"},
		},
	}
	sf := NewScrollableForm(mock)

	msg := tea.KeyMsg{Type: tea.KeyTab, Runes: []rune{'\t'}}
	result, _ := sf.Update(msg)
	sf = result.(*ScrollableForm)

	if sf.FocusIndex() != 1 {
		t.Errorf("expected Tab to move focus to index 1, got %d", sf.FocusIndex())
	}
}

func TestScrollableForm_ShiftTab_MovesFocusBackward(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
			{Kind: FocusText, Label: "Email", Key: "email"},
			{Kind: FocusButton, Label: "Save", Key: "save"},
		},
		focusIndex: 2,
	}
	sf := NewScrollableForm(mock)

	// In bubbletea, shift+tab is detected via msg.String() == "shift+tab".
	// When Type is KeyRunes, String() returns string(Runes).
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("shift+tab")}
	result, _ := sf.Update(msg)
	sf = result.(*ScrollableForm)

	if sf.FocusIndex() != 1 {
		t.Errorf("expected Shift+Tab to move focus to index 1, got %d", sf.FocusIndex())
	}
}

func TestScrollableForm_UpArrow_MovesFocusUp(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
			{Kind: FocusText, Label: "Email", Key: "email"},
			{Kind: FocusButton, Label: "Save", Key: "save"},
		},
		focusIndex: 1,
	}
	sf := NewScrollableForm(mock)

	msg := tea.KeyMsg{Type: tea.KeyUp}
	result, _ := sf.Update(msg)
	sf = result.(*ScrollableForm)

	if sf.FocusIndex() != 0 {
		t.Errorf("expected Up to move focus to index 0, got %d", sf.FocusIndex())
	}
}

func TestScrollableForm_DownArrow_MovesFocusDown(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
			{Kind: FocusText, Label: "Email", Key: "email"},
			{Kind: FocusButton, Label: "Save", Key: "save"},
		},
	}
	sf := NewScrollableForm(mock)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	result, _ := sf.Update(msg)
	sf = result.(*ScrollableForm)

	if sf.FocusIndex() != 1 {
		t.Errorf("expected Down to move focus to index 1, got %d", sf.FocusIndex())
	}
}

func TestScrollableForm_EnterKey_DispatchesToFormModel(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
			{Kind: FocusButton, Label: "Save", Key: "save"},
		},
		enterResult: ResultNone, // Enter on text field stays on form
	}
	sf := NewScrollableForm(mock)

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	result, _ := sf.Update(msg)
	sf = result.(*ScrollableForm)

	// Enter dispatches to HandleEnter; ResultNone keeps the form active.
	// The form decides what to do with Enter (e.g., advance focus).
	if sf.FocusIndex() != 0 {
		t.Errorf("expected Enter on text field to keep focus at 0, got %d", sf.FocusIndex())
	}
}

func TestScrollableForm_CharInput_DispatchesToFormModel(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
		},
	}
	sf := NewScrollableForm(mock)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	result, _ := sf.Update(msg)
	sf = result.(*ScrollableForm)

	if mock.handleCharCh != "a" {
		t.Errorf("expected HandleChar to receive 'a', got %q", mock.handleCharCh)
	}
	if mock.handleCharPos.Key != "name" {
		t.Errorf("expected HandleChar pos key 'name', got %q", mock.handleCharPos.Key)
	}
}

func TestScrollableForm_Backspace_DispatchesToFormModel(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
		},
	}
	sf := NewScrollableForm(mock)

	msg := tea.KeyMsg{Type: tea.KeyBackspace}
	result, _ := sf.Update(msg)
	_ = result

	if mock.handleBsPos.Key != "name" {
		t.Errorf("expected HandleBackspace pos key 'name', got %q", mock.handleBsPos.Key)
	}
}

func TestScrollableForm_Delete_DispatchesToFormModel(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
		},
	}
	sf := NewScrollableForm(mock)

	msg := tea.KeyMsg{Type: tea.KeyDelete}
	result, _ := sf.Update(msg)
	_ = result

	if mock.handleDelPos.Key != "name" {
		t.Errorf("expected HandleDelete pos key 'name', got %q", mock.handleDelPos.Key)
	}
}

// --- View rendering ---

func TestScrollableForm_View_ReturnsContent(t *testing.T) {
	mock := &mockFormModel{
		positions: []FocusPos{
			{Kind: FocusText, Label: "Name", Key: "name"},
		},
		header:      "=== My Form ===",
		footer:      "Tab: next | Enter: select",
		positionOut: "[Name: ]",
	}
	sf := NewScrollableForm(mock)

	// Must set size first to initialize viewport
	sf.SetSize(80, 24)

	view := sf.View()

	if view == "" {
		t.Fatal("expected non-empty view")
	}
	// The view should contain the header, position, and footer content
	// concatenated and rendered through the viewport
}

func TestScrollableForm_View_NotReady(t *testing.T) {
	mock := &mockFormModel{}
	sf := NewScrollableForm(mock)

	// Without SetSize/WindowSizeMsg, viewport is not ready
	view := sf.View()
	if view != "Loading..." {
		t.Errorf("expected 'Loading...' when not ready, got %q", view)
	}
}

// --- Lifecycle ---

func TestScrollableForm_FocusedControl(t *testing.T) {
	mock := &mockFormModel{}
	sf := NewScrollableForm(mock)

	sf.SetFocused(false)
	if mock.setFocused == nil || *mock.setFocused != false {
		t.Error("expected SetFocused(false) to propagate to FormModel")
	}

	sf.SetFocused(true)
	if mock.setFocused == nil || *mock.setFocused != true {
		t.Error("expected SetFocused(true) to propagate to FormModel")
	}
}

func TestScrollableForm_Init_CallsOnEnter(t *testing.T) {
	mock := &mockFormModel{}
	sf := NewScrollableForm(mock)

	if mock.onEnterCalled {
		t.Fatal("OnEnter should not be called before Init")
	}

	_ = sf.Init()

	if !mock.onEnterCalled {
		t.Error("expected Init to call FormModel.OnEnter")
	}
}
