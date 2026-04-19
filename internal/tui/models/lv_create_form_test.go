package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

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

func TestLVCreateFormNavigationAndToggle(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "100.00g", PVCount: 1}}

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab, Runes: []rune{'\t'}}) // name
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab, Runes: []rune{'\t'}}) // size
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab, Runes: []rune{'\t'}}) // unit
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab, Runes: []rune{'\t'}}) // thin
	if m.focusIndex != int(lvFocusThin) {
		t.Fatalf("expected focus thin, got %d", m.focusIndex)
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
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
	m.focusIndex = int(lvFocusCreate)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune{'\n'}})
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
	m.focusIndex = int(lvFocusName)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune{'\n'}})
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
	m.focusIndex = int(lvFocusVG)

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune{'\n'}})
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
	m.focusIndex = int(lvFocusVG)
	m.vgDropdownOpen = true
	m.vgDropdownIndex = 1

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter, Runes: []rune{'\n'}})
	if m.vgDropdownOpen {
		t.Fatal("expected VG dropdown to close after confirm")
	}
	if m.vgIndex != 1 {
		t.Fatalf("expected vgIndex=1, got %d", m.vgIndex)
	}
}

func TestLVCreateEscClosesDropdownOnly(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 1}}
	m.focusIndex = int(lvFocusVG)
	m.vgDropdownOpen = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.vgDropdownOpen {
		t.Fatal("expected VG dropdown to close")
	}
	if cmd != nil {
		t.Fatal("expected no view change command when closing dropdown")
	}
}

func TestLVCreateUpDownNavigatesVGDropdown(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 1}, {Name: "vg0", Free: "10.00g", PVCount: 1}}
	m.focusIndex = int(lvFocusVG)
	m.vgDropdownOpen = true
	m.vgDropdownIndex = 0

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.vgDropdownIndex != 1 {
		t.Fatalf("expected dropdown index 1 after down, got %d", m.vgDropdownIndex)
	}
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.vgDropdownIndex != 0 {
		t.Fatalf("expected dropdown index 0 after up, got %d", m.vgDropdownIndex)
	}
}

func TestLVCreateRenderNoHardcodedVGFallback(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)

	view := m.View()
	if !strings.Contains(view, "Volume Group: [                     ▼]") {
		t.Fatalf("expected empty VG placeholder, got:\n%s", view)
	}
	if strings.Contains(view, "ubuntu-vg") {
		t.Fatalf("expected no hardcoded ubuntu-vg fallback, got:\n%s", view)
	}
}

func TestLVCreateRenderDropdownOpen(t *testing.T) {
	m := NewLVCreateFormModel()
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 1}, {Name: "vg0", Free: "10.00g", PVCount: 1}}
	m.vgIndex = 0
	m.vgDropdownOpen = true
	m.vgDropdownIndex = 0
	m.SetSize(76, 18)
	m.syncViewport()

	view := stripANSI(m.View())
	assertGolden(t, "lv_create_form_vg_dropdown_open", view)
}

func TestLVCreateStrippedAutoEnabledWithMultiplePVs(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	// VG with 2 PVs should have stripped auto-enabled
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 2}}
	m.vgIndex = 0

	// Simulate loading VGs (triggers auto-enable)
	_, _ = m.Update(lvVGsLoadedMsg{
		vgs: []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 2}},
	})
	if !m.isStripped {
		t.Fatal("expected stripped to be auto-enabled for VG with 2 PVs")
	}
}

func TestLVCreateStrippedNotShownWithSinglePV(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	// VG with 1 PV should NOT show stripped option
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 1}}
	m.vgIndex = 0
	m.syncViewport()

	view := m.View()
	if strings.Contains(view, "Stripped") {
		t.Fatal("expected Stripped option hidden for VG with 1 PV")
	}
}

func TestLVCreateStrippedShownWithMultiplePVs(t *testing.T) {
	m := NewLVCreateFormModel()
	m.SetSize(76, 18)
	// VG with 2 PVs should show stripped option
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "300.00g", PVCount: 2, LVCount: 3}}
	m.vgIndex = 0
	m.isStripped = true
	m.syncViewport()

	view := m.View()
	if !strings.Contains(view, "Stripped") {
		t.Fatal("expected Stripped option shown for VG with 2 PVs")
	}
}

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
