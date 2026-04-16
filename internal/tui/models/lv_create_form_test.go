package models

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestParseVGSOutput(t *testing.T) {
	out := "ubuntu-vg\t500.00g\t300.00g\t2\nvg0\t100.00g\t10.00g\t5\n"
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
}

func TestLVCreateValidate(t *testing.T) {
	m := NewLVCreateFormModel()
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "10.00g"}}
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
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "100.00g"}}

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
	m.volumeGroups = []VolumeGroup{{Name: "ubuntu-vg", Free: "100.00g"}}
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
