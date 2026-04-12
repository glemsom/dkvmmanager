// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// LVMVolumeModel is a model for listing and selecting LVM Logical Volumes
type LVMVolumeModel struct {
	// LVM volumes
	volumes []LVMVolume

	// Selected volume path
	selectedPath string

	// Selection state
	selectedIndex int

	// Error message
	errorMsg string

	// Whether model is active
	active bool
}

// LVMVolume represents an LVM Logical Volume
type LVMVolume struct {
	Path     string
	Name     string
	VG       string
	Size     string
	Attr     string
	Type     string
}

// NewLVMVolumeModel creates a new LVM volume lister model
func NewLVMVolumeModel() *LVMVolumeModel {
	return &LVMVolumeModel{
		selectedIndex: 0,
		active:        true,
	}
}

// Init initializes the model
func (m *LVMVolumeModel) Init() tea.Cmd {
	return m.loadVolumes
}

// LVMVolumeLoadedMsg is sent after loadVolumes completes to trigger a view refresh
type LVMVolumeLoadedMsg struct{}

// loadVolumes loads LVM volumes from the system
func (m *LVMVolumeModel) loadVolumes() tea.Msg {
	volumes, err := m.listLVMVolumes()
	if err != nil {
		m.errorMsg = fmt.Sprintf("Failed to list LVM volumes: %v", err)
		return LVMVolumeLoadedMsg{}
	}

	m.volumes = volumes
	if m.selectedIndex >= len(m.volumes) {
		m.selectedIndex = 0
	}
	return LVMVolumeLoadedMsg{}
}

// listLVMVolumes lists available LVM Logical Volumes using lvs
func (m *LVMVolumeModel) listLVMVolumes() ([]LVMVolume, error) {
	// Use lvs to get LV information
	// Format: LV name, VG name, size, attributes
	cmd := exec.Command("lvs", "--noheadings", "-o", "lv_name,vg_name,lv_size,lv_attr", "--units", "g", "--separator", "\t")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("lvs command failed: %w", err)
	}

	return m.parseLVSOutput(string(output))
}

// parseLVSOutput parses lvs output
func (m *LVMVolumeModel) parseLVSOutput(output string) ([]LVMVolume, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var volumes []LVMVolume

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse tab-separated output
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}

		lvName := strings.TrimSpace(parts[0])
		vgName := strings.TrimSpace(parts[1])
		size := strings.TrimSpace(parts[2])
		attr := strings.TrimSpace(parts[3])

		// Skip snapshot origins (they start with [)
		if strings.HasPrefix(lvName, "[") {
			continue
		}

		// Determine volume type from attributes
		volType := "linear"
		if len(attr) > 6 {
			// attr format: permissions allocation status type
			// r/w = permissions, m/a = allocation, -/c/s = status, p/m/s = type
			switch attr[6:7] {
			case "p":
				volType = "pool"
			case "t":
				volType = "thin"
			case "s":
				volType = "snapshot"
			}
		}

		volumes = append(volumes, LVMVolume{
			Path:     "/dev/" + vgName + "/" + lvName,
			Name:     lvName,
			VG:       vgName,
			Size:     size,
			Attr:     attr,
			Type:     volType,
		})
	}

	// Sort by VG then name
	sort.Slice(volumes, func(i, j int) bool {
		if volumes[i].VG == volumes[j].VG {
			return volumes[i].Name < volumes[j].Name
		}
		return volumes[i].VG < volumes[j].VG
	})

	return volumes, nil
}

// Update handles incoming messages
func (m *LVMVolumeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case LVMVolumeLoadedMsg:
		// Volumes already loaded via side effect in loadVolumes command.
		// This message exists to trigger a view refresh.
		return m, nil
	}
	return m, nil
}

// handleKeyPress handles keyboard input
func (m *LVMVolumeModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errorMsg = ""

	switch msg.String() {
	case "ctrl+c", "esc":
		m.active = false
		return m, func() tea.Msg { return FileSelectedMsg{Path: "", Canceled: true} }

	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		if m.selectedIndex < len(m.volumes)-1 {
			m.selectedIndex++
		}

	case "enter", " ":
		return m.handleEnter()
	}

	return m, nil
}

// handleEnter handles the enter key
func (m *LVMVolumeModel) handleEnter() (tea.Model, tea.Cmd) {
	if len(m.volumes) == 0 {
		return m, nil
	}

	selected := m.volumes[m.selectedIndex]
	m.active = false
	m.selectedPath = selected.Path

	return m, func() tea.Msg { return FileSelectedMsg{Path: selected.Path, Canceled: false} }
}

// View returns the rendered view for LVMVolumeModel
func (m *LVMVolumeModel) View() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Primary).
		Bold(true)

	volNameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("247"))

	sizeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("75"))

	checkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("76"))

	roCheckStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted)

	var output string

	output += headerStyle.Render("Select LVM Volume") + "\n"
	output += "\n"
	output += "Available LVM Logical Volumes:\n"
	output += "\n"

	if len(m.volumes) == 0 {
		output += "  (no LVM volumes found)\n"
	} else {
		for i, vol := range m.volumes {
			var name string
			var checkMark string

			// Check if writable (first char of attr is 'w')
			writable := len(vol.Attr) > 0 && vol.Attr[0:1] == "w"
			if writable {
				name = volNameStyle.Render(vol.VG + "/" + vol.Name)
				checkMark = checkStyle.Render("✓")
			} else {
				name = volNameStyle.Render(vol.VG + "/" + vol.Name)
				checkMark = roCheckStyle.Render("✗")
			}

			if i == m.selectedIndex {
				output += "> " + selectedStyle.Render(name)
			} else {
				output += "  " + name
			}

			output += "  " + sizeStyle.Render(vol.Size)
			output += "  " + typeStyle.Render(vol.Type)
			output += "  " + checkMark
			output += "\n"
		}
	}

	if m.errorMsg != "" {
		output += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.errorMsg)
	}

	output += "\n\n"
	output += helpStyle.Render("↑/↓ Navigate  Space/Enter Select  ESC Cancel") + "\n"

	return output
}

// GetSelectedPath returns the selected volume path
func (m *LVMVolumeModel) GetSelectedPath() string {
	return m.selectedPath
}