package models

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// LVCreateUpdatedMsg indicates LV creation succeeded.
type LVCreateUpdatedMsg struct{}

// VolumeGroup represents an LVM Volume Group.
type VolumeGroup struct {
	Name    string
	Size    string
	Free    string
	LVCount int
	PVCount int
}

type lvVGsLoadedMsg struct {
	vgs []VolumeGroup
	err error
}

type lvCreateErrorMsg struct{ err string }

var lvNameRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
var lvmSizeRe = regexp.MustCompile(`^[<>]?\d+(?:\.\d+)?[a-zA-Z]?$`)

// LVCreateFormModel is the create LV dialog.
// It implements form.FormModel for use with the ScrollableForm framework.
type LVCreateFormModel struct {
	// VG state
	volumeGroups    []VolumeGroup
	vgIndex         int
	vgDropdownOpen  bool
	vgDropdownIndex int

	// Form state
	volumeName   string
	sizeValue    string
	unitIndex    int
	isThinPool   bool
	isStripped   bool
	isContiguous bool
	isReadOnly   bool

	// Framework state
	positions     []form.FocusPos
	focusIndex    int
	cursorOffsets map[string]int
	errors        map[string]string
	preview       string

	// ScrollableForm wrapper (for backward-compatible Update/View/Init/SetSize)
	form *form.ScrollableForm
}

func NewLVCreateFormModel() *LVCreateFormModel {
	m := &LVCreateFormModel{
		volumeName:    "my-data-volume",
		sizeValue:     "100",
		errors:        map[string]string{},
		cursorOffsets: make(map[string]int),
	}
	m.form = form.NewScrollableForm(m)
	return m
}

// Form returns the underlying ScrollableForm (for MainModel integration/testing).
func (m *LVCreateFormModel) Form() *form.ScrollableForm { return m.form }

// Preview returns the last preview message (dry-run command string).
func (m *LVCreateFormModel) Preview() string { return m.preview }

// Init implements tea.Model.
func (m *LVCreateFormModel) Init() tea.Cmd {
	return m.loadVolumeGroupsCmd()
}

// SetSize sets the form dimensions.
func (m *LVCreateFormModel) SetSize(w, h int) {
	m.form.SetSize(w, h)
}

// Update implements tea.Model (backward-compatible delegation to ScrollableForm).
func (m *LVCreateFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	inner, cmd := m.form.Update(msg)
	if sf, ok := inner.(*form.ScrollableForm); ok {
		m.form = sf
	}
	return m, cmd
}

// View implements tea.Model.
func (m *LVCreateFormModel) View() string {
	if m.form == nil || !m.form.Ready() {
		return "Loading form..."
	}
	return m.form.View()
}

// --- FormModel Interface ---

func (m *LVCreateFormModel) BuildPositions() []form.FocusPos {
	var positions []form.FocusPos

	positions = append(positions, form.FocusPos{
		Kind: form.FocusCustom, Label: "Volume Group", Key: "vg",
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusText, Label: "Volume Name", Key: "name",
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusCustom, Label: "Size", Key: "size",
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusCustom, Label: "Unit", Key: "unit",
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Thin pool", Key: "thin",
	})
	if m.selectedVGPVCount() > 1 {
		positions = append(positions, form.FocusPos{
			Kind: form.FocusToggle, Label: "Stripped", Key: "stripped",
		})
	}
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Contiguous", Key: "contig",
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Read-only", Key: "ro",
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusButton, Label: "Create", Key: "create",
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusButton, Label: "Cancel", Key: "cancel",
	})
	m.positions = positions
	return positions
}

func (m *LVCreateFormModel) CurrentIndex() int { return m.focusIndex }
func (m *LVCreateFormModel) SetFocusIndex(i int) { m.focusIndex = i }

func (m *LVCreateFormModel) RenderHeader() string {
	return lipgloss.NewStyle().Bold(true).Foreground(styles.Colors.Primary).Render("Create Logical Volume") +
		"\n" + styles.FormMutedStyle().Render("────────────────────────────────────────────────────────────────")
}

func (m *LVCreateFormModel) RenderFooter() string {
	return styles.MutedTextStyle().Render("Tab Navigate  Enter Select/Submit  ESC Cancel")
}

func (m *LVCreateFormModel) RenderPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
	switch pos.Key {
	case "vg":
		return m.renderVGPosition(focused)
	case "name":
		return m.renderNamePosition(focused, cursorOffset)
	case "size":
		return m.renderSizePosition(focused, cursorOffset)
	case "unit":
		return m.renderUnitPosition(focused)
	case "thin", "stripped", "contig", "ro":
		return m.renderTogglePosition(pos, focused)
	case "create", "cancel":
		return m.renderButtonPosition(pos, focused)
	default:
		return pos.Label
	}
}

// renderVGPosition renders the Volume Group selector.
func (m *LVCreateFormModel) renderVGPosition(focused bool) string {
	vg := m.selectedVG()
	vgMarker := "▼"
	if m.vgDropdownOpen {
		vgMarker = "▲"
	}
	style := styles.FormMutedStyle()
	if focused {
		style = styles.FormFocusStyle()
	}
	line := fmt.Sprintf("  Volume Group: [%-20s %s]", vg, vgMarker)
	result := style.Render(line)
	if m.vgDropdownOpen {
		for i, candidate := range m.volumeGroups {
			prefix := "    "
			if i == m.vgDropdownIndex {
				prefix = ">   "
			}
			result += "\n" + fmt.Sprintf("    %s%-20s free: %s", prefix, candidate.Name, candidate.Free)
		}
	}
	return result
}

// renderNamePosition renders the Volume Name text input.
func (m *LVCreateFormModel) renderNamePosition(focused bool, cursorOffset int) string {
	style := styles.FormMutedStyle()
	if focused {
		style = styles.FormFocusStyle()
	}
	line := fmt.Sprintf("  Volume Name:  [%-20s]", m.volumeName)
	return style.Render(line)
}

// renderSizePosition renders the Size text input.
func (m *LVCreateFormModel) renderSizePosition(focused bool, cursorOffset int) string {
	style := styles.FormMutedStyle()
	if focused {
		style = styles.FormFocusStyle()
	}
	line := fmt.Sprintf("  Size:         [%-20s]  %s", m.sizeValue, m.units()[m.unitIndex])
	return style.Render(line)
}

// renderUnitPosition renders the Unit selector.
func (m *LVCreateFormModel) renderUnitPosition(focused bool) string {
	style := styles.FormMutedStyle()
	if focused {
		style = styles.FormFocusStyle()
	}
	units := m.units()
	var parts []string
	for i, u := range units {
		if i == m.unitIndex {
			parts = append(parts, "["+u+"]")
		} else {
			parts = append(parts, " "+u+" ")
		}
	}
	return style.Render("  Unit:           " + strings.Join(parts, " "))
}

// renderTogglePosition renders a toggle option.
func (m *LVCreateFormModel) renderTogglePosition(pos form.FocusPos, focused bool) string {
	value := false
	switch pos.Key {
	case "thin":
		value = m.isThinPool
	case "stripped":
		value = m.isStripped
	case "contig":
		value = m.isContiguous
	case "ro":
		value = m.isReadOnly
	}
	cb := "[ ]"
	if value {
		cb = "[x]"
	}
	style := styles.FormMutedStyle()
	if focused {
		style = styles.FormFocusStyle()
	}
	return style.Render("  " + cb + " " + pos.Label)
}

// renderButtonPosition renders a button.
func (m *LVCreateFormModel) renderButtonPosition(pos form.FocusPos, focused bool) string {
	style := styles.FormMutedStyle()
	if focused {
		style = styles.FormFocusStyle()
	}
	return style.Render("  [ " + pos.Label + " ]")
}

func (m *LVCreateFormModel) HandleEnter(pos form.FocusPos) (form.FormResult, tea.Cmd) {
	switch pos.Key {
	case "vg":
		if m.vgDropdownOpen {
			m.confirmVGSelection()
		} else {
			m.openVGDropdown()
		}
		return form.ResultNone, nil
	case "create":
		return m.submitCreate()
	case "cancel":
		return form.ResultCancel, func() tea.Msg { return ViewChangeMsg{View: ViewConfigMenu} }
	default:
		// Enter on any other position also submits
		return m.submitCreate()
	}
}

func (m *LVCreateFormModel) submitCreate() (form.FormResult, tea.Cmd) {
	if m.validate() {
		return form.ResultSave, m.createCmd()
	}
	return form.ResultNone, nil
}

func (m *LVCreateFormModel) HandleChar(pos form.FocusPos, ch string) {
	switch pos.Key {
	case "name":
		if len(m.volumeName) < 128 {
			m.volumeName += ch
		}
	case "size":
		if (ch[0] >= '0' && ch[0] <= '9') || ch == "." {
			m.sizeValue += ch
		}
	}
}

func (m *LVCreateFormModel) HandleBackspace(pos form.FocusPos) {
	switch pos.Key {
	case "name":
		if len(m.volumeName) > 0 {
			m.volumeName = m.volumeName[:len(m.volumeName)-1]
		}
	case "size":
		if len(m.sizeValue) > 0 {
			m.sizeValue = m.sizeValue[:len(m.sizeValue)-1]
		}
	}
}

func (m *LVCreateFormModel) HandleDelete(pos form.FocusPos) {
	// No delete handling needed for LV create
}

func (m *LVCreateFormModel) OnEnter()    {}
func (m *LVCreateFormModel) OnExit()     {}
func (m *LVCreateFormModel) SetFocused(bool) {}

// --- arrowKeyHandler interface ---

func (m *LVCreateFormModel) HandleLeft(pos form.FocusPos) {
	switch pos.Key {
	case "vg":
		if m.vgDropdownOpen {
			m.moveVGSelection(-1)
		} else if len(m.volumeGroups) > 0 {
			m.vgIndex--
			if m.vgIndex < 0 {
				m.vgIndex = len(m.volumeGroups) - 1
			}
		}
	case "unit":
		m.unitIndex--
		if m.unitIndex < 0 {
			m.unitIndex = len(m.units()) - 1
		}
	}
}

func (m *LVCreateFormModel) HandleRight(pos form.FocusPos) {
	switch pos.Key {
	case "vg":
		if m.vgDropdownOpen {
			m.moveVGSelection(1)
		} else if len(m.volumeGroups) > 0 {
			m.vgIndex = (m.vgIndex + 1) % len(m.volumeGroups)
		}
	case "unit":
		m.unitIndex = (m.unitIndex + 1) % len(m.units())
	}
}

// --- spaceHandler interface ---

func (m *LVCreateFormModel) HandleSpace(pos form.FocusPos) {
	switch pos.Key {
	case "thin":
		m.isThinPool = !m.isThinPool
	case "stripped":
		m.isStripped = !m.isStripped
	case "contig":
		m.isContiguous = !m.isContiguous
	case "ro":
		m.isReadOnly = !m.isReadOnly
	case "create":
		if m.validate() {
			m.createCmd() // Execute inline since HandleSpace returns nothing
		}
	}
}

// --- handleMessage interface for async VG loading ---

func (m *LVCreateFormModel) HandleMessage(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
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
		// Rebuild positions since VG changed (may affect Stripped visibility)
		m.BuildPositions()
	case LVCreateUpdatedMsg:
		// Success — handled by MainModel
	case lvCreateErrorMsg:
		m.errors["size"] = msg.err
	}
	return nil
}

// --- Helper methods (preserved from original) ---

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

func (m *LVCreateFormModel) units() []string { return []string{"GiB", "TiB", "MiB"} }

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
	m.BuildPositions() // Rebuild for Stripped visibility
	// Adjust focus if stripped was removed and we were on or past it
	strippedIdx := m.indexOfKey("stripped")
	contigIdx := m.indexOfKey("contig")
	if strippedIdx < 0 && contigIdx >= 0 && m.focusIndex >= contigIdx-1 {
		// Stripped was removed; if focus was on where stripped would be, move it to contig
		m.focusIndex = contigIdx
	}
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

// indexOfKey returns the position index for a given key, or -1 if not found.
func (m *LVCreateFormModel) indexOfKey(key string) int {
	for i, p := range m.positions {
		if p.Key == key {
			return i
		}
	}
	return -1
}

// --- Validation and command building (preserved) ---

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
		// Auto-stripes uses number of PVs in the VG
		stripeCount := m.selectedVGPVCount()
		if stripeCount < 2 {
			stripeCount = 2 // Minimum 2 stripes if enabled
		}
		cmd += fmt.Sprintf(" --stripes %d", stripeCount)
	}
	if m.isContiguous {
		cmd += " --contiguous y"
	}
	if m.isReadOnly {
		cmd += " -p r"
	}
	return cmd
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

// --- VG loading (preserved) ---

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

		// Use literal tab character (not "\t" string) for --separator
		argsPrimary := []string{"--noheadings", "-o", "vg_name,vg_size,vg_free,lv_count,pv_count", "--units", "g", "--separator", "\t"}
		// Convert actual tab to string for logging display
		argsPrimaryDisplay := make([]string, len(argsPrimary))
		for i, a := range argsPrimary {
			if a == "\t" {
				argsPrimaryDisplay[i] = "'<TAB>'"
			} else {
				argsPrimaryDisplay[i] = a
			}
		}
		cmd := exec.Command("vgs", argsPrimary...)
		cmd.Env = append(os.Environ(), "LVM_SUPPRESS_FD_WARNINGS=1")
		out, err := cmd.CombinedOutput()
		if debugMode {
			log.Printf("[DEBUG] LV create: running: vgs %s", strings.Join(argsPrimaryDisplay, " "))
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

func parseSizeG(s string) float64 {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimSuffix(s, "g")
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
