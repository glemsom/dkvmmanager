package models

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// LVCreateUpdatedMsg indicates LV creation succeeded.
type LVCreateUpdatedMsg struct{}

type VolumeGroup struct {
	Name    string
	Size    string
	Free    string
	LVCount int
}

type lvCreateFocus int

const (
	lvFocusVG lvCreateFocus = iota
	lvFocusName
	lvFocusSize
	lvFocusUnit
	lvFocusThin
	lvFocusContig
	lvFocusRO
	lvFocusCreate
	lvFocusCancel
)

var lvNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// LVCreateFormModel is the create LV dialog.
type LVCreateFormModel struct {
	volumeGroups []VolumeGroup
	vgIndex      int
	volumeName   string
	sizeValue    string
	unitIndex    int
	isThinPool   bool
	isContiguous bool
	isReadOnly   bool
	focusIndex   int
	errors       map[string]string
	preview      string

	vp       viewport.Model
	ready    bool
	contentW int
	contentH int
}

func NewLVCreateFormModel() *LVCreateFormModel {
	return &LVCreateFormModel{
		volumeName: "my-data-volume",
		sizeValue:  "100",
		errors:     map[string]string{},
	}
}

func (m *LVCreateFormModel) Init() tea.Cmd { return m.loadVolumeGroupsCmd() }

func (m *LVCreateFormModel) SetSize(w, h int) {
	m.contentW, m.contentH = w, h
	if !m.ready {
		m.vp = viewport.New(w, h)
		m.ready = true
	} else {
		m.vp.Width = w
		m.vp.Height = h
	}
	m.syncViewport()
}

func (m *LVCreateFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case lvVGsLoadedMsg:
		m.volumeGroups = msg.vgs
		if msg.err != nil {
			m.errors["vg"] = "Failed to load VGs: " + msg.err.Error()
		}
		if len(m.volumeGroups) == 0 {
			m.errors["vg"] = "No volume groups found"
		}
		m.syncViewport()
	}
	return m, nil
}

func (m *LVCreateFormModel) View() string {
	if !m.ready {
		return "Loading form..."
	}
	return m.vp.View()
}

type lvVGsLoadedMsg struct {
	vgs []VolumeGroup
	err error
}

func (m *LVCreateFormModel) loadVolumeGroupsCmd() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("vgs", "--noheadings", "-o", "vg_name,vg_size,vg_free,lv_count", "--units", "g", "--separator", "\t")
		out, err := cmd.Output()
		if err != nil {
			return lvVGsLoadedMsg{err: err}
		}
		vgs, pErr := parseVGSOutput(string(out))
		if pErr != nil {
			return lvVGsLoadedMsg{err: pErr}
		}
		return lvVGsLoadedMsg{vgs: vgs}
	}
}

func parseVGSOutput(output string) ([]VolumeGroup, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	out := make([]VolumeGroup, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}
		cnt, _ := strconv.Atoi(strings.TrimSpace(parts[3]))
		out = append(out, VolumeGroup{
			Name:    strings.TrimSpace(parts[0]),
			Size:    strings.TrimSpace(parts[1]),
			Free:    strings.TrimSpace(parts[2]),
			LVCount: cnt,
		})
	}
	return out, nil
}

func (m *LVCreateFormModel) selectedVG() string {
	if m.vgIndex >= 0 && m.vgIndex < len(m.volumeGroups) {
		return m.volumeGroups[m.vgIndex].Name
	}
	return ""
}

func (m *LVCreateFormModel) units() []string { return []string{"GiB", "TiB", "MiB"} }

func (m *LVCreateFormModel) buildCommand() string {
	vg := m.selectedVG()
	suffix := "G"
	switch m.units()[m.unitIndex] {
	case "TiB":
		suffix = "T"
	case "MiB":
		suffix = "M"
	}
	cmd := fmt.Sprintf("lvcreate -L %s%s -n %s %s", m.sizeValue, suffix, m.volumeName, vg)
	if m.isThinPool {
		cmd += " --type thin"
	}
	if m.isContiguous {
		cmd += " --contiguous y"
	}
	if m.isReadOnly {
		cmd += " -p r"
	}
	return cmd
}

func parseSizeG(s string) float64 {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimSuffix(s, "g")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func (m *LVCreateFormModel) validate() bool {
	m.errors = map[string]string{}
	if m.selectedVG() == "" {
		m.errors["vg"] = "Volume Group is required"
	}
	if strings.TrimSpace(m.volumeName) == "" {
		m.errors["name"] = "Volume Name is required"
	} else if len(m.volumeName) > 128 || !lvNameRe.MatchString(m.volumeName) {
		m.errors["name"] = "Invalid name (a-z A-Z 0-9 _ -)"
	}
	s, err := strconv.ParseFloat(m.sizeValue, 64)
	if err != nil || s <= 0 {
		m.errors["size"] = "Size must be a positive number"
	}
	if m.vgIndex < len(m.volumeGroups) && m.vgIndex >= 0 {
		free := parseSizeG(m.volumeGroups[m.vgIndex].Free)
		if free > 0 && s > free {
			m.errors["size"] = "Size exceeds VG free space"
		}
	}
	return len(m.errors) == 0
}

type lvCreateErrorMsg struct{ err string }

func (m *LVCreateFormModel) createCmd() tea.Cmd {
	cmdStr := m.buildCommand()
	if dryRunMode {
		m.preview = "Would execute: " + cmdStr
		return func() tea.Msg { return LVCreateUpdatedMsg{} }
	}
	parts := strings.Fields(cmdStr)
	return func() tea.Msg {
		cmd := exec.Command(parts[0], parts[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			msg := strings.TrimSpace(string(out))
			if msg == "" {
				msg = err.Error()
			}
			return lvCreateErrorMsg{err: msg}
		}
		return LVCreateUpdatedMsg{}
	}
}

func (m *LVCreateFormModel) renderLines() []string {
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	focus := styles.FormFocusStyle()
	muted := styles.FormMutedStyle()
	lines := []string{
		"Create Logical Volume",
		"────────────────────────────────────────────────────────────────",
		"",
	}
	vg := ""
	if m.selectedVG() != "" {
		vg = m.selectedVG()
	} else {
		vg = "ubuntu-vg"
	}
	vgLine := fmt.Sprintf("Volume Group: [%-20s ▼]", vg)
	if m.focusIndex == int(lvFocusVG) {
		vgLine = focus.Render(vgLine)
	} else {
		vgLine = label.Render(vgLine)
	}
	lines = append(lines, vgLine)
	nameLine := fmt.Sprintf("Volume Name:  [%-20s]", m.volumeName)
	if m.focusIndex == int(lvFocusName) {
		nameLine = focus.Render(nameLine)
	}
	lines = append(lines, nameLine)
	sizeLine := fmt.Sprintf("Size:         [%-20s ▼]  %s", m.sizeValue, m.units()[m.unitIndex])
	if m.focusIndex == int(lvFocusSize) || m.focusIndex == int(lvFocusUnit) {
		sizeLine = focus.Render(sizeLine)
	}
	lines = append(lines, sizeLine, "", "Options:")
	cb := func(v bool) string { if v { return "[x]" }; return "[ ]" }
	lines = append(lines,
		fmt.Sprintf("  %s Thin pool              %s Contiguous", cb(m.isThinPool), cb(m.isContiguous)),
		fmt.Sprintf("  %s Read-only", cb(m.isReadOnly)),
	)
	if len(m.errors) > 0 {
		lines = append(lines, "")
		for _, k := range []string{"vg", "name", "size"} {
			if e := m.errors[k]; e != "" {
				lines = append(lines, styles.ErrorTextStyle().Render("- "+e))
			}
		}
	}
	if m.preview != "" {
		lines = append(lines, "", muted.Render(m.preview))
	}
	lines = append(lines, "", "[Space/Enter] Create    [ESC] Cancel")
	return lines
}

func (m *LVCreateFormModel) syncViewport() {
	content := strings.Join(m.renderLines(), "\n")
	m.vp.SetContent(content)
}

func (m *LVCreateFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab", "down":
		m.focusIndex = (m.focusIndex + 1) % 9
	case "shift+tab", "up":
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = 8
		}
	case "left":
		switch lvCreateFocus(m.focusIndex) {
		case lvFocusVG:
			if len(m.volumeGroups) > 0 {
				m.vgIndex--
				if m.vgIndex < 0 {
					m.vgIndex = len(m.volumeGroups) - 1
				}
			}
		case lvFocusUnit:
			m.unitIndex--
			if m.unitIndex < 0 {
				m.unitIndex = len(m.units()) - 1
			}
		}
	case "right":
		switch lvCreateFocus(m.focusIndex) {
		case lvFocusVG:
			if len(m.volumeGroups) > 0 {
				m.vgIndex = (m.vgIndex + 1) % len(m.volumeGroups)
			}
		case lvFocusUnit:
			m.unitIndex = (m.unitIndex + 1) % len(m.units())
		}
	case " ", "enter":
		switch lvCreateFocus(m.focusIndex) {
		case lvFocusThin:
			m.isThinPool = !m.isThinPool
		case lvFocusContig:
			m.isContiguous = !m.isContiguous
		case lvFocusRO:
			m.isReadOnly = !m.isReadOnly
		case lvFocusCreate:
			if m.validate() {
				m.syncViewport()
				return m, m.createCmd()
			}
		case lvFocusCancel:
			return m, func() tea.Msg { return ViewChangeMsg{View: ViewConfigMenu} }
		}
	case "esc":
		return m, func() tea.Msg { return ViewChangeMsg{View: ViewConfigMenu} }
	case "backspace":
		switch lvCreateFocus(m.focusIndex) {
		case lvFocusName:
			if len(m.volumeName) > 0 {
				m.volumeName = m.volumeName[:len(m.volumeName)-1]
			}
		case lvFocusSize:
			if len(m.sizeValue) > 0 {
				m.sizeValue = m.sizeValue[:len(m.sizeValue)-1]
			}
		}
	default:
		if len(msg.Runes) > 0 {
			r := string(msg.Runes)
			switch lvCreateFocus(m.focusIndex) {
			case lvFocusName:
				if len(m.volumeName) < 128 {
					m.volumeName += r
				}
			case lvFocusSize:
				if (r[0] >= '0' && r[0] <= '9') || r == "." {
					m.sizeValue += r
				}
			}
		}
	}
	m.syncViewport()
	return m, nil
}
