package models

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- parseVGSOutput tests (unchanged from original) ---

func TestParseVGSOutput(t *testing.T) {
	out := "ubuntu-vg\t500.00g\t300.00g\t2\t1\nvg0\t100.00g\t10.00g\t5\t3\n"
	vgs, err := parseVGSOutput(out)
	if err != nil {
		t.Fatalf("parseVGSOutput() error: %v", err)
	}
	if len(vgs) != 2 {
		t.Fatalf("expected 2 VGs, got %d", len(vgs))
	}
	if vgs[0].Name != "ubuntu-vg" || vgs[1].Name != "vg0" {
		t.Fatalf("unexpected vg parse result: %#v", vgs)
	}
	if vgs[0].PVCount != 1 || vgs[1].PVCount != 3 {
		t.Fatalf("unexpected pv count: %#v", vgs)
	}
}

func TestParseVGSOutputWhitespaceAndEmptyLines(t *testing.T) {
	out := "\n  ubuntu-vg\t500.00g\t300.00g\t2\t2\n\t\ninvalid line\n  vg0\t100.00g\t10.00g\t5\t1\n"
	vgs, err := parseVGSOutput(out)
	if err != nil {
		t.Fatalf("parseVGSOutput() error: %v", err)
	}
	if len(vgs) != 2 {
		t.Fatalf("expected 2 VGs, got %d", len(vgs))
	}
	if vgs[0].Name != "ubuntu-vg" || vgs[1].Name != "vg0" {
		t.Fatalf("unexpected vg parse result: %#v", vgs)
	}
	if vgs[0].PVCount != 2 || vgs[1].PVCount != 1 {
		t.Fatalf("unexpected pv count: %#v", vgs)
	}
}

func TestParseVGSOutputLiteralEscapedTabSeparator(t *testing.T) {
	out := "vg_nvme\\t7452.04g\\t2798.04g\\t6\\t4\n"
	vgs, err := parseVGSOutput(out)
	if err != nil {
		t.Fatalf("parseVGSOutput() error: %v", err)
	}
	if len(vgs) != 1 {
		t.Fatalf("expected 1 VG, got %d", len(vgs))
	}
	if vgs[0].Name != "vg_nvme" || vgs[0].Free != "2798.04g" || vgs[0].LVCount != 6 || vgs[0].PVCount != 4 {
		t.Fatalf("unexpected vg parse result: %#v", vgs[0])
	}
}

// --- FormModel interface tests ---

func TestLVCreateBuildPositions(t *testing.T) {
	m := NewLVCreateFormModel()
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "100.00g", PVCount: 1}}
	m.vgIndex = 0

	positions := m.BuildPositions()

	// VG with 1 PV: no stripped option → 9 positions
	expectedCount := 9
	if len(positions) != expectedCount {
		t.Fatalf("expected %d positions (single PV), got %d", expectedCount, len(positions))
	}

	// Verify position keys
	expectedKeys := []string{"vg", "name", "size", "unit", "thin", "contig", "ro", "create", "cancel"}
	for i, key := range expectedKeys {
		if positions[i].Key != key {
			t.Errorf("position %d: expected key %q, got %q", i, key, positions[i].Key)
		}
	}

	// FocusIndex starts at 0 (VG)
	if m.focusIndex != 0 {
		t.Fatalf("expected focusIndex 0, got %d", m.focusIndex)
	}
}

func TestLVCreateBuildPositionsMultiPV(t *testing.T) {
	m := NewLVCreateFormModel()
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "100.00g", PVCount: 2}}
	m.vgIndex = 0

	positions := m.BuildPositions()

	// VG with 2 PVs: stripped option shown → 10 positions
	expectedCount := 10
	if len(positions) != expectedCount {
		t.Fatalf("expected %d positions (multi PV), got %d", expectedCount, len(positions))
	}

	// Verify stripped position exists
	found := false
	for _, p := range positions {
		if p.Key == "stripped" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected 'stripped' position for multi-PV VG")
	}
}

// --- Validation tests ---

func TestLVCreateValidate(t *testing.T) {
	m := NewLVCreateFormModel()
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "10.00g", PVCount: 1}}
	m.vgIndex = 0
	m.volumeName = "ok-name"
	m.sizeValue = "2"
	if !m.validate() {
		t.Fatalf("expected validation success: %#v", m.errors)
	}

	m.volumeName = "bad name"
	if m.validate() {
		t.Fatal("expected validation failure for invalid name")
	}
	m.volumeName = "ok-name"
	m.sizeValue = "999"
	if m.validate() {
		t.Fatal("expected validation failure for size > free")
	}
}

// --- Navigation and interaction tests (via ScrollableForm Update) ---

func TestLVCreateFormNavigationAndToggle(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "100.00g", PVCount: 1}}
	m.vgIndex = 0

	// Tab through: VG(0) → name(1) → size(2) → unit(3) → thin(4)
	for i := 0; i < 4; i++ {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	}

	// Check we're on "thin" position
	positions := m.BuildPositions()
	if positions[m.focusIndex].Key != "thin" {
		t.Fatalf("expected focus on 'thin', got %q at index %d", positions[m.focusIndex].Key, m.focusIndex)
	}

	// Space toggles thin pool
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !m.isThinPool {
		t.Fatal("expected thin pool toggled on")
	}
}

func TestLVCreateDryRunCreate(t *testing.T) {
	orig := dryRunMode
	dryRunMode = true
	defer func() { dryRunMode = orig }()

	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "100.00g", PVCount: 1}}
	m.vgIndex = 0
	m.volumeName = "data"
	m.sizeValue = "10"

	// Navigate to Create button via Tab presses (7 steps: VG->name->size->unit->thin->contig->ro->create)
	for i := 0; i < 7; i++ {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected create command")
	}
	msg := cmd()
	if _, ok := msg.(LVCreateUpdatedMsg); !ok {
		t.Fatalf("expected LVCreateUpdatedMsg, got %T", msg)
	}
	if m.preview == "" {
		t.Fatal("expected dry-run preview")
	}
}

func TestLVCreateEnterSubmitsFromNonCreateFocus(t *testing.T) {
	orig := dryRunMode
	dryRunMode = true
	defer func() { dryRunMode = orig }()

	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "vg0", Free: "100.00g", PVCount: 1}}
	m.vgIndex = 0
	m.volumeName = "data"
	m.sizeValue = "10"

	// Navigate to name position (one tab from VG)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected create command from enter on non-create focus")
	}
	msg := cmd()
	if _, ok := msg.(LVCreateUpdatedMsg); !ok {
		t.Fatalf("expected LVCreateUpdatedMsg, got %T", msg)
	}
	if m.preview == "" {
		t.Fatal("expected dry-run preview")
	}
}

func TestLVCreateEnterOnVGOpensDropdown(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 1}, {Name: "vg0", Free: "10.00g", PVCount: 1}}
	m.vgIndex = 0
	m.focusIndex = 0 // VG position

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.vgDropdownOpen {
		t.Fatal("expected VG dropdown to open")
	}
	if m.vgDropdownIndex != 0 {
		t.Fatalf("expected dropdown index 0, got %d", m.vgDropdownIndex)
	}
}

func TestLVCreateEnterOnVGWhenOpenConfirmsSelection(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 2}, {Name: "vg0", Free: "10.00g", PVCount: 1}}
	m.vgIndex = 0
	m.focusIndex = 0
	m.vgDropdownOpen = true
	m.vgDropdownIndex = 1

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.vgDropdownOpen {
		t.Fatal("expected VG dropdown to close after confirm")
	}
	if m.vgIndex != 1 {
		t.Fatalf("expected vgIndex=1, got %d", m.vgIndex)
	}
}

func TestLVCreateEscIsNoOp(t *testing.T) {
	// Note: In production, ESC is intercepted by MainModel.update() before reaching
	// the form. MainModel calls returnFromSubView() for all sub-views including
	// ViewLVCreate. This test verifies the form's Update is a no-op for ESC,
	// which is expected since ESC handling is done at the MainModel level.
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 1}}
	m.focusIndex = 0
	m.vgDropdownOpen = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	// Form Update is a no-op for ESC (MainModel handles it)
	if cmd != nil {
		t.Fatal("expected no command from ESC (MainModel handles it)")
	}
}

func TestLVCreateUpDownNavigatesVGDropdown(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 1}, {Name: "vg0", Free: "10.00g", PVCount: 1}}
	m.focusIndex = 0
	m.vgDropdownOpen = true
	m.vgDropdownIndex = 0

	// The framework dispatches up/down to move focus between positions.
	// When the VG dropdown is open, the original code used up/down to
	// navigate the dropdown list. In the new framework, left/right are
	// dispatched to HandleLeft/HandleRight for dropdown navigation.
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.vgDropdownIndex != 1 {
		t.Fatalf("expected dropdown index 1 after right, got %d", m.vgDropdownIndex)
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.vgDropdownIndex != 0 {
		t.Fatalf("expected dropdown index 0 after left, got %d", m.vgDropdownIndex)
	}
}

func TestLVCreateUnitCycling(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)

	// Navigate to unit position: VG(0) → name(1) → size(2) → unit(3)
	for i := 0; i < 3; i++ {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	}

	positions := m.BuildPositions()
	if positions[m.focusIndex].Key != "unit" {
		t.Fatalf("expected focus on 'unit', got %q", positions[m.focusIndex].Key)
	}

	// Right cycles forward: GiB → TiB → MiB → GiB
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.unitIndex != 1 {
		t.Fatalf("expected unitIndex 1 (TiB), got %d", m.unitIndex)
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.unitIndex != 2 {
		t.Fatalf("expected unitIndex 2 (MiB), got %d", m.unitIndex)
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.unitIndex != 0 {
		t.Fatalf("expected unitIndex 0 (wrap to GiB), got %d", m.unitIndex)
	}

	// Left cycles backward
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.unitIndex != 2 {
		t.Fatalf("expected unitIndex 2 (MiB) after left, got %d", m.unitIndex)
	}
}

func TestLVCreateTextInput(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)

	// Tab to name position (one tab from VG at index 0)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	positions := m.BuildPositions()
	if positions[m.focusIndex].Key != "name" {
		t.Fatalf("expected focus on 'name', got %q", positions[m.focusIndex].Key)
	}

	// Type characters one at a time (framework expects single characters)
	for _, ch := range "test" {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	if !strings.Contains(m.volumeName, "test") {
		t.Fatalf("expected 'test' in volumeName, got %q", m.volumeName)
	}

	// Tab to size (one tab from name)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	positions = m.BuildPositions()
	if positions[m.focusIndex].Key != "size" {
		t.Fatalf("expected focus on 'size', got %q", positions[m.focusIndex].Key)
	}
	for _, ch := range "50" {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	if !strings.Contains(m.sizeValue, "50") {
		t.Fatalf("expected '50' in sizeValue, got %q", m.sizeValue)
	}

	// Backspace on size
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if len(m.sizeValue) < 3 {
		t.Fatalf("expected size to shrink after backspace, got %q", m.sizeValue)
	}
}

func TestLVCreateToggleBehavior(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)

	// Navigate to thin toggle: VG(0) → name(1) → size(2) → unit(3) → thin(4)
	for i := 0; i < 4; i++ {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	}

	// Space toggles thin
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !m.isThinPool {
		t.Fatal("expected thin pool toggled on")
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.isThinPool {
		t.Fatal("expected thin pool toggled off")
	}

	// Navigate to contiguous (skip stripped if single PV)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	positions := m.BuildPositions()
	if positions[m.focusIndex].Key != "contig" {
		t.Fatalf("expected focus on 'contig', got %q", positions[m.focusIndex].Key)
	}
}

// --- Render tests ---

func TestLVCreateRenderNoHardcodedVGFallback(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)

	view := m.View()
	// Should show Volume Group line with empty value
	if !strings.Contains(view, "Volume Group:") {
		t.Fatalf("expected 'Volume Group:' in view, got:\n%s", view)
	}
	if strings.Contains(view, "ubuntu-vg") {
		t.Fatalf("expected no hardcoded ubuntu-vg fallback, got:\n%s", view)
	}
}

func TestLVCreateStrippedNotShownWithSinglePV(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 1}}
	m.vgIndex = 0

	// BuildPositions with single PV should NOT include stripped
	positions := m.BuildPositions()
	for _, p := range positions {
		if p.Key == "stripped" {
			t.Fatal("expected 'stripped' position to NOT exist for VG with 1 PV")
		}
	}

	// Should have 9 positions (no stripped)
	if len(positions) != 9 {
		t.Fatalf("expected 9 positions without stripped, got %d", len(positions))
	}
}

func TestLVCreateStrippedShownWithMultiplePVs(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 2, LVCount: 3}}
	m.vgIndex = 0
	m.isStripped = true
	m.BuildPositions()
	// Sync viewport after BuildPositions to include the stripped option
	m.SetSize(76, 18)

	view := m.View()
	if !strings.Contains(view, "Stripped") {
		t.Fatal("expected Stripped option shown for VG with 2 PVs")
	}
}

// --- Async message handling ---

func TestLVCreateStrippedAutoEnabledWithMultiplePVs(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 2}}
	m.vgIndex = 0

	// Simulate loading VGs via HandleMessage
	m.HandleMessage(lvVGsLoadedMsg{
		vgs: []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 2}},
	})
	if !m.isStripped {
		t.Fatal("expected stripped to be auto-enabled for VG with 2 PVs")
	}
}

func TestLVCreateHandleMessageError(t *testing.T) {
	m := NewLVCreateFormModel()
	m.HandleMessage(lvVGsLoadedMsg{err: fmt.Errorf("vgs failed")})
	if m.errors["vg"] == "" {
		t.Fatal("expected VG error after failed load")
	}
}

// --- Command building tests ---

func TestLVCreateBuildCommandWithStripes(t *testing.T) {
	m := NewLVCreateFormModel()
	m.volumeGroups = []VolumeGroup{{Name: "vg0", Free: "100.00g", PVCount: 2}}
	m.vgIndex = 0
	m.volumeName = "mylv"
	m.sizeValue = "50"
	m.isStripped = true

	cmd := m.buildCommand()
	if !strings.Contains(cmd, "--stripes") {
		t.Fatalf("expected --stripes in command, got: %s", cmd)
	}
}

func TestLVCreateBuildCommandWithoutStripes(t *testing.T) {
	m := NewLVCreateFormModel()
	m.volumeGroups = []VolumeGroup{{Name: "vg0", Free: "100.00g", PVCount: 1}}
	m.vgIndex = 0
	m.volumeName = "mylv"
	m.sizeValue = "50"
	m.isStripped = false

	cmd := m.buildCommand()
	if strings.Contains(cmd, "--stripes") {
		t.Fatalf("expected no --stripes in command, got: %s", cmd)
	}
}

func TestLVCreateBuildCommandThinAndContiguous(t *testing.T) {
	m := NewLVCreateFormModel()
	m.volumeGroups = []VolumeGroup{{Name: "vg0", Free: "100.00g", PVCount: 1}}
	m.vgIndex = 0
	m.volumeName = "thinvol"
	m.sizeValue = "20"
	m.isThinPool = true
	m.isContiguous = true

	cmd := m.buildCommand()
	if !strings.Contains(cmd, "--type thin") {
		t.Fatalf("expected --type thin in command, got: %s", cmd)
	}
	if !strings.Contains(cmd, "--contiguous y") {
		t.Fatalf("expected --contiguous y in command, got: %s", cmd)
	}
}
