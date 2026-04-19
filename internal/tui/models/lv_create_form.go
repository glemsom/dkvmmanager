package models

import (
	"fmt"
	"log"
	"os"
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
	PVCount int
}

type lvCreateFocus int

const (
	lvFocusVG lvCreateFocus = iota
	lvFocusName
	lvFocusSize
	lvFocusUnit
	lvFocusThin
	lvFocusStripped
	lvFocusContig
	lvFocusRO
	lvFocusCreate
	lvFocusCancel
)

var lvNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
var lvmSizeRe = regexp.MustCompile(`^[<>]?\d+(?:\.\d+)?[a-zA-Z]?$`)

// LVCreateFormModel is the create LV dialog.
type LVCreateFormModel struct {
	volumeGroups     []VolumeGroup
	vgIndex          int
	vgDropdownOpen   bool
	vgDropdownIndex  int
	volumeName       string
	sizeValue        string
	unitIndex        int
	isThinPool       bool
	isStripped       bool
	isContiguous     bool
	isReadOnly       bool
	focusIndex       int
	errors           map[string]string
	preview          string

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

func (m *LVCreateFormModel) Init() tea.Cmd {
	return m.loadVolumeGroupsCmd()
}

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
		} else {
			delete(m.errors, "vg")
		}
		if len(m.volumeGroups) == 0 {
			m.errors["vg"] = "No volume groups found"
			m.vgIndex = -1
			m.vgDropdownIndex = -1
		} else {
			if m.vgIndex < 0 || m.vgIndex >= len(m.volumeGroups) {
				m.vgIndex = 0
			}
			if m.vgDropdownIndex < 0 || m.vgDropdownIndex >= len(m.volumeGroups) {
				m.vgDropdownIndex = m.vgIndex
			}
			// Auto-enable stripped if first VG has more than 1 PV
			if m.selectedVGPVCount() > 1 {
				m.isStripped = true
			}
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
		if debugMode {
			if p, err := exec.LookPath("vgs"); err == nil {
				log.Printf("[DEBUG] LV create: vgs path=%s", p)
			} else {
				log.Printf("[DEBUG] LV create: vgs not found in PATH=%q (%v)", os.Getenv("PATH"), err)
			}
			log.Printf("[DEBUG] LV create: uid=%d euid=%d", os.Getuid(), os.Geteuid())
		}

		argsPrimary := []string{"--noheadings", "-o", "vg_name,vg_size,vg_free,lv_count,pv_count", "--units", "g", "--separator", "\t"}
		cmd := exec.Command("vgs", argsPrimary...)
		cmd.Env = append(os.Environ(), "LVM_SUPPRESS_FD_WARNINGS=1")
		out, err := cmd.CombinedOutput()
		if debugMode {
			log.Printf("[DEBUG] LV create: running: vgs %s", strings.Join(argsPrimary, " "))
			log.Printf("[DEBUG] LV create: primary output raw=%q", strings.TrimSpace(string(out)))
		}
		if err != nil {
			stderr := strings.TrimSpace(string(out))
			if debugMode {
				log.Printf("[DEBUG] LV create: primary vgs failed: %v", err)
			}
			if stderr != "" {
				return lvVGsLoadedMsg{err: fmt.Errorf("%w: %s", err, stderr)}
			}
			return lvVGsLoadedMsg{err: err}
		}

		vgs, pErr := parseVGSOutput(string(out))
		if pErr != nil {
			return lvVGsLoadedMsg{err: pErr}
		}
		if len(vgs) > 0 {
			if debugMode {
				log.Printf("[DEBUG] LV create: parsed %d volume groups (primary)", len(vgs))
			}
			return lvVGsLoadedMsg{vgs: vgs}
		}

		// Fallback for environments that ignore separator flags or format unexpectedly.
		argsFallback := []string{"--noheadings", "-o", "vg_name,vg_size,vg_free,lv_count,pv_count", "--units", "g"}
		fallbackCmd := exec.Command("vgs", argsFallback...)
		fallbackCmd.Env = append(os.Environ(), "LVM_SUPPRESS_FD_WARNINGS=1")
		fallbackOut, fallbackErr := fallbackCmd.CombinedOutput()
		if debugMode {
			log.Printf("[DEBUG] LV create: running fallback: vgs %s", strings.Join(argsFallback, " "))
			log.Printf("[DEBUG] LV create: fallback output raw=%q", strings.TrimSpace(string(fallbackOut)))
		}
		if fallbackErr != nil {
			stderr := strings.TrimSpace(string(fallbackOut))
			if stderr != "" {
				return lvVGsLoadedMsg{err: fmt.Errorf("fallback %w: %s", fallbackErr, stderr)}
			}
			return lvVGsLoadedMsg{err: fallbackErr}
		}

		fallbackVGS, fallbackParseErr := parseVGSOutput(string(fallbackOut))
		if fallbackParseErr != nil {
			return lvVGsLoadedMsg{err: fallbackParseErr}
		}
		if debugMode {
			log.Printf("[DEBUG] LV create: parsed %d volume groups (fallback)", len(fallbackVGS))
		}
		return lvVGsLoadedMsg{vgs: fallbackVGS}
	}
}

func parseVGSOutput(output string) ([]VolumeGroup, error) {
	lines := strings.Split(output, "\n")
	out := make([]VolumeGroup, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := splitVGSLine(line)
		if debugMode {
			log.Printf("[DEBUG] LV create: parse line=%q parts=%q", line, parts)
		}
		if len(parts) < 5 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		size := strings.TrimSpace(parts[1])
		free := strings.TrimSpace(parts[2])
		if name == "" {
			continue
		}
		if !lvmSizeRe.MatchString(size) || !lvmSizeRe.MatchString(free) {
			continue
		}
		cnt, cntErr := strconv.Atoi(strings.TrimSpace(parts[3]))
		if cntErr != nil {
			continue
		}
		pvCnt, pvErr := strconv.Atoi(strings.TrimSpace(parts[4]))
		if pvErr != nil {
			pvCnt = 1
		}
		out = append(out, VolumeGroup{
			Name:    name,
			Size:    size,
			Free:    free,
			LVCount: cnt,
			PVCount: pvCnt,
		})
	}
	return out, nil
}

func splitVGSLine(line string) []string {
	// Preferred: real tab separator from `--separator "\t"`.
	if strings.Contains(line, "\t") {
		return strings.Split(line, "\t")
	}
	// Some setups can emit escaped separator as literal "\\t".
	if strings.Contains(line, "\\t") {
		return strings.Split(line, "\\t")
	}
	// Fallback for vgs outputs that are whitespace separated.
	return strings.Fields(line)
}

func (m *LVCreateFormModel) selectedVG() string {
	if m.vgIndex >= 0 && m.vgIndex < len(m.volumeGroups) {
		return m.volumeGroups[m.vgIndex].Name
	}
	return ""
}

func (m *LVCreateFormModel) selectedVGPVCount() int {
	if m.vgIndex >= 0 && m.vgIndex < len(m.volumeGroups) {
		return m.volumeGroups[m.vgIndex].PVCount
	}
	return 0
}

// nextFocus returns the next focus index, skipping Stripped if VG has <= 1 PV.
func (m *LVCreateFormModel) nextFocus(delta int) int {
	maxFocus := 9
	pvCount := m.selectedVGPVCount()
	hasStrippedFocus := pvCount > 1

	newFocus := m.focusIndex + delta
	if newFocus < 0 {
		newFocus = maxFocus - 1
	} else if newFocus >= maxFocus {
		newFocus = 0
	}

	// If we're on Stripped focus but VG has <= 1 PV, skip it
	if newFocus == int(lvFocusStripped) && !hasStrippedFocus {
		if delta > 0 {
			newFocus = int(lvFocusContig)
		} else {
			newFocus = int(lvFocusThin)
		}
	}
	return newFocus
}

func (m *LVCreateFormModel) openVGDropdown() {
	if len(m.volumeGroups) == 0 {
		return
	}
	m.vgDropdownOpen = true
	if m.vgIndex >= 0 && m.vgIndex < len(m.volumeGroups) {
		m.vgDropdownIndex = m.vgIndex
		return
	}
	m.vgDropdownIndex = 0
}

func (m *LVCreateFormModel) closeVGDropdown() {
	m.vgDropdownOpen = false
}

func (m *LVCreateFormModel) confirmVGSelection() {
	if len(m.volumeGroups) == 0 {
		m.closeVGDropdown()
		return
	}
	if m.vgDropdownIndex < 0 {
		m.vgDropdownIndex = 0
	}
	if m.vgDropdownIndex >= len(m.volumeGroups) {
		m.vgDropdownIndex = len(m.volumeGroups) - 1
	}
	m.vgIndex = m.vgDropdownIndex
	// Auto-enable stripped if VG has more than 1 PV
	if m.selectedVGPVCount() > 1 {
		m.isStripped = true
	} else {
		m.isStripped = false
	}
	m.closeVGDropdown()
}

func (m *LVCreateFormModel) moveVGSelection(delta int) {
	if len(m.volumeGroups) == 0 {
		return
	}
	if m.vgDropdownIndex < 0 || m.vgDropdownIndex >= len(m.volumeGroups) {
		m.vgDropdownIndex = m.vgIndex
		if m.vgDropdownIndex < 0 || m.vgDropdownIndex >= len(m.volumeGroups) {
			m.vgDropdownIndex = 0
		}
	}
	m.vgDropdownIndex += delta
	if m.vgDropdownIndex < 0 {
		m.vgDropdownIndex = len(m.volumeGroups) - 1
	}
	if m.vgDropdownIndex >= len(m.volumeGroups) {
		m.vgDropdownIndex = 0
	}
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
	if m.isStripped {
		cmd += " --stripes"
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
	vg := m.selectedVG()
	vgMarker := "▼"
	if m.vgDropdownOpen {
		vgMarker = "▲"
	}
	vgLine := fmt.Sprintf("Volume Group: [%-20s %s]", vg, vgMarker)
	if m.focusIndex == int(lvFocusVG) {
		vgLine = focus.Render(vgLine)
	} else {
		vgLine = label.Render(vgLine)
	}
	lines = append(lines, vgLine)
	if m.vgDropdownOpen {
		for i, candidate := range m.volumeGroups {
			prefix := "  "
			if i == m.vgDropdownIndex {
				prefix = "> "
			}
			lines = append(lines, fmt.Sprintf("  %s%-20s free: %s", prefix, candidate.Name, candidate.Free))
		}
	}
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
	cb := func(v bool, focused bool) string {
		if v { return "[x]" }
		if focused { return focus.Render("[ ]") }
		return "[ ]"
	}
	lines = append(lines,
		fmt.Sprintf("  %s Thin pool              %s Contiguous", cb(m.isThinPool, m.focusIndex == int(lvFocusThin)), cb(m.isContiguous, m.focusIndex == int(lvFocusContig))),
	)
	if m.selectedVGPVCount() > 1 {
		lines = append(lines, fmt.Sprintf("  %s Stripped", cb(m.isStripped, m.focusIndex == int(lvFocusStripped))))
	}
	lines = append(lines, fmt.Sprintf("  %s Read-only", cb(m.isReadOnly, m.focusIndex == int(lvFocusRO))))
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
	lines = append(lines, "", "[Enter on Create] Create    [ESC] Cancel")
	return lines
}

func (m *LVCreateFormModel) syncViewport() {
	lines := m.renderLines()
	if m.contentW > 0 {
		for i := range lines {
			lines[i] = lipgloss.NewStyle().MaxWidth(m.contentW).Render(lines[i])
		}
	}
	content := strings.Join(lines, "\n")
	m.vp.SetContent(content)
}

func (m *LVCreateFormModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.vgDropdownOpen {
			m.closeVGDropdown()
		}
		m.focusIndex = m.nextFocus(1)
	case "shift+tab":
		if m.vgDropdownOpen {
			m.closeVGDropdown()
		}
		m.focusIndex = m.nextFocus(-1)
	case "down":
		if m.focusIndex == int(lvFocusVG) && m.vgDropdownOpen {
			m.moveVGSelection(1)
			break
		}
		m.focusIndex = m.nextFocus(1)
	case "up":
		if m.focusIndex == int(lvFocusVG) && m.vgDropdownOpen {
			m.moveVGSelection(-1)
			break
		}
		m.focusIndex = m.nextFocus(-1)
	case "left":
		switch lvCreateFocus(m.focusIndex) {
		case lvFocusVG:
			if m.vgDropdownOpen {
				m.moveVGSelection(-1)
			} else if len(m.volumeGroups) > 0 {
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
			if m.vgDropdownOpen {
				m.moveVGSelection(1)
			} else if len(m.volumeGroups) > 0 {
				m.vgIndex = (m.vgIndex + 1) % len(m.volumeGroups)
			}
		case lvFocusUnit:
			m.unitIndex = (m.unitIndex + 1) % len(m.units())
		}
	case " ":
		switch lvCreateFocus(m.focusIndex) {
		case lvFocusThin:
			m.isThinPool = !m.isThinPool
		case lvFocusStripped:
			m.isStripped = !m.isStripped
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
	case "enter":
		if m.focusIndex == int(lvFocusVG) {
			if m.vgDropdownOpen {
				m.confirmVGSelection()
			} else {
				m.openVGDropdown()
			}
			break
		}
		if m.validate() {
			m.syncViewport()
			return m, m.createCmd()
		}
	case "esc":
		if m.vgDropdownOpen {
			m.closeVGDropdown()
			break
		}
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
