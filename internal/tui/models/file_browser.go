// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// FileType represents the type of file to filter for
type FileType int

const (
	FileTypeAll FileType = iota
	FileTypeDiskImage
	FileTypeISO
)

// FileBrowserModel is a model for browsing and selecting files
type FileBrowserModel struct {
	// Current directory
	currentDir string

	// Selected file path
	selectedPath string

	// Files and directories in current view
	files []FileEntry

	// Selection state
	selectedIndex int

	// File type filter
	fileType FileType

	// Error message
	errorMsg string

	// For returning to previous model
	returnMsg tea.Msg

	// Whether browser is active
	active bool
}

// FileEntry represents a file or directory entry
type FileEntry struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
	Mode  os.FileMode
}

// NewFileBrowserModel creates a new file browser model
func NewFileBrowserModel(fileType FileType) *FileBrowserModel {
	// Start from user's home directory or root
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/"
	}

	return &FileBrowserModel{
		currentDir:    homeDir,
		fileType:      fileType,
		selectedIndex: 0,
		active:        true,
	}
}

// Init initializes the model
func (m *FileBrowserModel) Init() tea.Cmd {
	return m.loadDirectory
}

// Update handles incoming messages
func (m *FileBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case DirectoryLoadedMsg:
		// Directory already loaded via side effect in loadDirectory command.
		// This message exists to trigger a view refresh.
		return m, nil
	}
	return m, nil
}

// handleKeyPress handles keyboard input
func (m *FileBrowserModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errorMsg = ""

	switch msg.String() {
	case "ctrl+c", "esc":
		// Cancel and return
		m.active = false
		return m, func() tea.Msg { return FileSelectedMsg{Path: "", Canceled: true} }

	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		if m.selectedIndex < len(m.files)-1 {
			m.selectedIndex++
		}

	case "enter":
		return m.handleEnter()

	case " ":
		return m.handleEnter()

	case "backspace":
		return m.navigateUp()
	}

	return m, nil
}

// handleEnter handles the enter key
func (m *FileBrowserModel) handleEnter() (tea.Model, tea.Cmd) {
	if len(m.files) == 0 {
		return m, nil
	}

	selected := m.files[m.selectedIndex]

	if selected.IsDir {
		// Navigate into directory
		m.currentDir = selected.Path
		m.selectedIndex = 0
		return m, m.loadDirectory
	}

	// File selected - return path
	m.active = false
	m.selectedPath = selected.Path
	return m, func() tea.Msg { return FileSelectedMsg{Path: selected.Path, Canceled: false} }
}

// navigateUp navigates to the parent directory
func (m *FileBrowserModel) navigateUp() (tea.Model, tea.Cmd) {
	parent := filepath.Dir(m.currentDir)
	if parent == m.currentDir {
		return m, nil // Already at root
	}
	m.currentDir = parent
	m.selectedIndex = 0
	return m, m.loadDirectory
}

// DirectoryLoadedMsg is sent after loadDirectory completes to trigger a view refresh
type DirectoryLoadedMsg struct{}

// loadDirectory loads files from the current directory
func (m *FileBrowserModel) loadDirectory() tea.Msg {
	entries, err := m.listDirectory(m.currentDir)
	if err != nil {
		m.errorMsg = fmt.Sprintf("Failed to read directory: %v", err)
		return DirectoryLoadedMsg{}
	}

	m.files = entries
	if m.selectedIndex >= len(m.files) {
		m.selectedIndex = 0
	}
	return DirectoryLoadedMsg{}
}

// listDirectory lists directory contents with filtering
func (m *FileBrowserModel) listDirectory(dir string) ([]FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []FileEntry

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		isDir := entry.IsDir()
		name := entry.Name()

		// Skip hidden files
		if strings.HasPrefix(name, ".") {
			continue
		}

		// Filter based on file type
		if !isDir {
			switch m.fileType {
			case FileTypeDiskImage:
				// Allow block devices or common disk image extensions
				if !isDiskImageFile(name) && !isBlockDevice(dir, name) {
					continue
				}
			case FileTypeISO:
				// Only allow .iso files
				if !isISOFile(name) {
					continue
				}
			}
		}

		files = append(files, FileEntry{
			Name:  name,
			Path:  filepath.Join(dir, name),
			IsDir: isDir,
			Size:  info.Size(),
			Mode:  info.Mode(),
		})
	}

	// Sort: directories first, then files alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir && !files[j].IsDir {
			return true
		}
		if !files[i].IsDir && files[j].IsDir {
			return false
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	// Add ".." entry if not at root
	if dir != "/" {
		files = append([]FileEntry{{
			Name:  "..",
			Path:  filepath.Dir(dir),
			IsDir: true,
		}}, files...)
	}

	return files, nil
}

// isDiskImageFile checks if filename is a disk image
func isDiskImageFile(name string) bool {
	lower := strings.ToLower(name)
	extensions := []string{".img", ".raw", ".qcow2", ".qcow", ".vmdk", ".vdi", ".vhdx"}
	for _, ext := range extensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// isBlockDevice checks if the path points to a block device
func isBlockDevice(dir, name string) bool {
	path := filepath.Join(dir, name)
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeDevice) != 0
}

// isISOFile checks if filename is an ISO image
func isISOFile(name string) bool {
	return strings.ToLower(filepath.Ext(name)) == ".iso"
}

// View returns the rendered view
func (m *FileBrowserModel) View() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	dirStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("81"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Primary).
		Bold(true)

	fileStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("247"))

	dirMarkerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("75"))

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted)

	var output string

	// Header
	switch m.fileType {
	case FileTypeDiskImage:
		output += headerStyle.Render("Select Disk Image") + "\n"
	case FileTypeISO:
		output += headerStyle.Render("Select ISO Image") + "\n"
	default:
		output += headerStyle.Render("Select File") + "\n"
	}

	output += "\n"
	output += "Current: " + dirStyle.Render(m.currentDir) + "\n"
	output += "\n"

	// File list
	if len(m.files) == 0 {
		output += "  (empty directory)\n"
	} else {
		for i, file := range m.files {
			prefix := "  "
			var name string

			if file.IsDir {
				if file.Name == ".." {
					name = dirMarkerStyle.Render(".. (parent)")
				} else {
					name = dirMarkerStyle.Render(file.Name) + dirMarkerStyle.Render("/")
				}
			} else {
				name = fileStyle.Render(file.Name)
			}

			if i == m.selectedIndex {
				output += selectedStyle.Render("> "+name) + "\n"
			} else {
				output += prefix + name + "\n"
			}
		}
	}

	// Error message
	if m.errorMsg != "" {
		output += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.errorMsg)
	}

	// Help text
	output += "\n\n"
	output += helpStyle.Render("↑/↓ Navigate  Space/Enter Select  Backspace Parent  ESC Cancel") + "\n"

	return output
}

// SetDirectory sets the starting directory
func (m *FileBrowserModel) SetDirectory(dir string) {
	if _, err := os.Stat(dir); err == nil {
		m.currentDir = dir
	}
}

// GetSelectedPath returns the selected file path
func (m *FileBrowserModel) GetSelectedPath() string {
	return m.selectedPath
}

// FileSelectedMsg is a message indicating a file was selected
type FileSelectedMsg struct {
	Path     string
	Canceled bool
}
