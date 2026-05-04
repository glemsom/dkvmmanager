// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"os"
	"path/filepath"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

const dkvmDataMountPath = "/media/dkvmdata"

// isMountPoint checks whether the given path is a mount point by comparing
// the device ID of the path with the device ID of its parent directory.
func isMountPoint(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	parentDir := filepath.Dir(path)
	parentInfo, err := os.Stat(parentDir)
	if err != nil {
		return false, err
	}
	// If device IDs differ, it's a mount point
	return info.Sys().(*syscall.Stat_t).Dev != parentInfo.Sys().(*syscall.Stat_t).Dev, nil
}

// MountPointWarningModel displays a warning modal when /media/dkvmdata
// is not a mount point, informing the user that the DKVM hypervisor
// auto-mounts filesystems with LABEL=dkvmdata.
type MountPointWarningModel struct {
	// Selected option (0 = OK)
	selectedIndex int
}

// NewMountPointWarningModel creates a new mount point warning model.
func NewMountPointWarningModel() *MountPointWarningModel {
	return &MountPointWarningModel{
		selectedIndex: 0,
	}
}

// Init initializes the model.
func (m *MountPointWarningModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages.
func (m *MountPointWarningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

// handleKeyPress handles keyboard input.
func (m *MountPointWarningModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", " ", "esc":
		// Dismiss the warning and return to the main menu
		return m, func() tea.Msg {
			return ViewChangeMsg{View: ViewMainMenu}
		}
	}

	return m, nil
}

// View returns the view for the model.
func (m *MountPointWarningModel) View() string {
	warningTitle := lipgloss.NewStyle().
		Foreground(styles.Colors.Warning).
		Bold(true).
		Padding(0, 1).
		Render("⚠ Mount Point Warning")

	warningStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Warning).
		Bold(true)

	bodyStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Text)

	noteStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted).
		Italic(true)

	buttonSelected := lipgloss.NewStyle().
		Foreground(styles.Colors.Background).
		Background(styles.Colors.Primary).
		Bold(true).
		Padding(0, 2).
		Render("OK")

	var output string
	output += warningTitle + "\n\n"
	output += warningStyle.Render("Warning:") + " " + bodyStyle.Render("/media/dkvmdata is not a mount point.") + "\n\n"
	output += bodyStyle.Render("The DKVM hypervisor will auto-mount filesystems with LABEL=dkvmdata.") + "\n"
	output += bodyStyle.Render("To resolve this, create a filesystem with the label 'dkvmdata' and restart.") + "\n\n"
	output += noteStyle.Render("Note: Most VM settings done in the TUI cannot be persisted without this mount.") + "\n\n"

	output += "  " + buttonSelected + "\n\n"
	output += noteStyle.Render("Press Enter, Space, or ESC to dismiss")

	return styles.ModalStyle().Render(output)
}
