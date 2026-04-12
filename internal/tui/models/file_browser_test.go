package models

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func setupFileBrowserTest(t *testing.T, fileType FileType) (*FileBrowserModel, string) {
	t.Helper()
	tmpDir := t.TempDir()

	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	os.Mkdir(filepath.Join(tmpDir, "another"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "disk.qcow2"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "install.iso"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "data.txt"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte{}, 0644)

	m := NewFileBrowserModel(fileType)
	m.SetDirectory(tmpDir)
	return m, tmpDir
}

func TestNewFileBrowserModel(t *testing.T) {
	m := NewFileBrowserModel(FileTypeAll)

	if !m.active {
		t.Error("Expected active to be true")
	}
	if m.fileType != FileTypeAll {
		t.Errorf("Expected fileType FileTypeAll, got %d", m.fileType)
	}
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", m.selectedIndex)
	}

	homeDir, _ := os.UserHomeDir()
	if homeDir != "" && m.currentDir != homeDir {
		t.Errorf("Expected currentDir to be home dir '%s', got '%s'", homeDir, m.currentDir)
	}
}

func TestNewFileBrowserModelFileTypeISO(t *testing.T) {
	m := NewFileBrowserModel(FileTypeISO)

	if m.fileType != FileTypeISO {
		t.Errorf("Expected fileType FileTypeISO, got %d", m.fileType)
	}
	if !m.active {
		t.Error("Expected active to be true")
	}
}

func TestFileBrowserSetDirectory(t *testing.T) {
	m, tmpDir := setupFileBrowserTest(t, FileTypeAll)

	if m.currentDir != tmpDir {
		t.Errorf("Expected currentDir '%s', got '%s'", tmpDir, m.currentDir)
	}
}

func TestFileBrowserSetDirectoryInvalid(t *testing.T) {
	m := NewFileBrowserModel(FileTypeAll)
	originalDir := m.currentDir

	m.SetDirectory("/nonexistent/path/that/does/not/exist")

	if m.currentDir != originalDir {
		t.Errorf("Expected currentDir to remain '%s', got '%s'", originalDir, m.currentDir)
	}
}

func TestFileBrowserListDirectoryAll(t *testing.T) {
	m, tmpDir := setupFileBrowserTest(t, FileTypeAll)

	entries, err := m.listDirectory(tmpDir)
	if err != nil {
		t.Fatalf("listDirectory() error: %v", err)
	}

	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}

	if !contains(names, "subdir") {
		t.Error("Expected 'subdir' in entries")
	}
	if !contains(names, "another") {
		t.Error("Expected 'another' in entries")
	}
	if !contains(names, "disk.qcow2") {
		t.Error("Expected 'disk.qcow2' in entries")
	}
	if !contains(names, "install.iso") {
		t.Error("Expected 'install.iso' in entries")
	}
	if !contains(names, "data.txt") {
		t.Error("Expected 'data.txt' in entries")
	}
}

func TestFileBrowserListDirectoryHiddenSkipped(t *testing.T) {
	m, tmpDir := setupFileBrowserTest(t, FileTypeAll)

	entries, err := m.listDirectory(tmpDir)
	if err != nil {
		t.Fatalf("listDirectory() error: %v", err)
	}

	for _, e := range entries {
		if e.Name == ".." {
			continue
		}
		if strings.HasPrefix(e.Name, ".") {
			t.Errorf("Hidden file '%s' should be excluded", e.Name)
		}
	}
}

func TestFileBrowserListDirectoryParentEntry(t *testing.T) {
	m, tmpDir := setupFileBrowserTest(t, FileTypeAll)

	entries, err := m.listDirectory(tmpDir)
	if err != nil {
		t.Fatalf("listDirectory() error: %v", err)
	}

	if entries[0].Name != ".." {
		t.Errorf("Expected first entry '..', got '%s'", entries[0].Name)
	}
	if !entries[0].IsDir {
		t.Error("Expected '..' entry to be a directory")
	}
	if entries[0].Path != filepath.Dir(tmpDir) {
		t.Errorf("Expected '..' path '%s', got '%s'", filepath.Dir(tmpDir), entries[0].Path)
	}
}

func TestFileBrowserListDirectoryDiskImage(t *testing.T) {
	m, tmpDir := setupFileBrowserTest(t, FileTypeDiskImage)

	entries, err := m.listDirectory(tmpDir)
	if err != nil {
		t.Fatalf("listDirectory() error: %v", err)
	}

	for _, e := range entries {
		if e.Name == ".." || e.IsDir {
			continue
		}
		if !isDiskImageFile(e.Name) {
			t.Errorf("With FileTypeDiskImage, should only show disk images, got '%s'", e.Name)
		}
	}

	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	if contains(names, "data.txt") {
		t.Error("Should not include 'data.txt' with FileTypeDiskImage")
	}
	if contains(names, "install.iso") {
		t.Error("Should not include 'install.iso' with FileTypeDiskImage")
	}
	if !contains(names, "disk.qcow2") {
		t.Error("Expected 'disk.qcow2' with FileTypeDiskImage")
	}
}

func TestFileBrowserListDirectoryISO(t *testing.T) {
	m, tmpDir := setupFileBrowserTest(t, FileTypeISO)

	entries, err := m.listDirectory(tmpDir)
	if err != nil {
		t.Fatalf("listDirectory() error: %v", err)
	}

	for _, e := range entries {
		if e.Name == ".." || e.IsDir {
			continue
		}
		if !isISOFile(e.Name) {
			t.Errorf("With FileTypeISO, should only show .iso files, got '%s'", e.Name)
		}
	}

	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name
	}
	if contains(names, "data.txt") {
		t.Error("Should not include 'data.txt' with FileTypeISO")
	}
	if contains(names, "disk.qcow2") {
		t.Error("Should not include 'disk.qcow2' with FileTypeISO")
	}
	if !contains(names, "install.iso") {
		t.Error("Expected 'install.iso' with FileTypeISO")
	}
}

func TestFileBrowserListDirectorySortsDirsFirst(t *testing.T) {
	m, tmpDir := setupFileBrowserTest(t, FileTypeAll)

	entries, err := m.listDirectory(tmpDir)
	if err != nil {
		t.Fatalf("listDirectory() error: %v", err)
	}

	seenFile := false
	for _, e := range entries {
		if e.Name == ".." {
			continue
		}
		if !e.IsDir && !seenFile {
			seenFile = true
		}
		if e.IsDir && seenFile {
			t.Errorf("Directory '%s' appears after a file — dirs should come first", e.Name)
		}
	}
}

func TestFileBrowserListDirectorySortsAlphabetically(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "charlie.txt"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "alpha.txt"), []byte{}, 0644)
	os.WriteFile(filepath.Join(tmpDir, "bravo.txt"), []byte{}, 0644)

	m := &FileBrowserModel{fileType: FileTypeAll}
	entries, err := m.listDirectory(tmpDir)
	if err != nil {
		t.Fatalf("listDirectory() error: %v", err)
	}

	var fileNames []string
	for _, e := range entries {
		if e.Name == ".." || e.IsDir {
			continue
		}
		fileNames = append(fileNames, e.Name)
	}

	expected := []string{"alpha.txt", "bravo.txt", "charlie.txt"}
	if len(fileNames) != len(expected) {
		t.Fatalf("Expected %d files, got %d", len(expected), len(fileNames))
	}
	for i, name := range expected {
		if fileNames[i] != name {
			t.Errorf("Expected file[%d]='%s', got '%s'", i, name, fileNames[i])
		}
	}
}

func TestFileBrowserListDirectoryError(t *testing.T) {
	m := &FileBrowserModel{fileType: FileTypeAll}

	_, err := m.listDirectory("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
}

func TestIsDiskImageFile(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"disk.qcow2", true},
		{"disk.img", true},
		{"disk.raw", true},
		{"disk.vmdk", true},
		{"disk.qcow", true},
		{"disk.vdi", true},
		{"disk.vhdx", true},
		{"DISK.QCOW2", true},
		{"disk.iso", false},
		{"disk.txt", false},
		{"noextension", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDiskImageFile(tt.name)
			if got != tt.expected {
				t.Errorf("isDiskImageFile(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestIsISOFile(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"install.iso", true},
		{"INSTALL.ISO", true},
		{"disk.qcow2", false},
		{"file.txt", false},
		{"file.iso.bak", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isISOFile(tt.name)
			if got != tt.expected {
				t.Errorf("isISOFile(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestFileBrowserHandleKeyPressUp(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true},
		{Name: "file1.txt", IsDir: false},
		{Name: "file2.txt", IsDir: false},
	}

	m := &FileBrowserModel{
		currentDir:    "/tmp",
		files:         entries,
		selectedIndex: 1,
		fileType:      FileTypeAll,
		active:        true,
	}

	updated, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(*FileBrowserModel)

	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after up, got %d", m.selectedIndex)
	}
}

func TestFileBrowserHandleKeyPressDown(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true},
		{Name: "file1.txt", IsDir: false},
		{Name: "file2.txt", IsDir: false},
	}

	m := &FileBrowserModel{
		currentDir:    "/tmp",
		files:         entries,
		selectedIndex: 0,
		fileType:      FileTypeAll,
		active:        true,
	}

	updated, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(*FileBrowserModel)

	if m.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after down, got %d", m.selectedIndex)
	}
}

func TestFileBrowserHandleKeyPressVimKeys(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true},
		{Name: "file1.txt", IsDir: false},
		{Name: "file2.txt", IsDir: false},
	}

	m := &FileBrowserModel{
		currentDir:    "/tmp",
		files:         entries,
		selectedIndex: 1,
		fileType:      FileTypeAll,
		active:        true,
	}

	updated, _ := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = updated.(*FileBrowserModel)

	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after 'k', got %d", m.selectedIndex)
	}

	m.selectedIndex = 1
	updated, _ = m.handleKeyPress(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = updated.(*FileBrowserModel)

	if m.selectedIndex != 2 {
		t.Errorf("Expected selectedIndex 2 after 'j', got %d", m.selectedIndex)
	}
}

func TestFileBrowserHandleKeyPressESC(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true},
		{Name: "file1.txt", IsDir: false},
	}

	m := &FileBrowserModel{
		currentDir:    "/tmp",
		files:         entries,
		selectedIndex: 0,
		fileType:      FileTypeAll,
		active:        true,
	}

	updated, cmd := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(*FileBrowserModel)

	if m.active {
		t.Error("Expected active=false after ESC")
	}
	if cmd == nil {
		t.Fatal("Expected command after ESC")
	}

	msg := cmd()
	fsm, ok := msg.(FileSelectedMsg)
	if !ok {
		t.Fatalf("Expected FileSelectedMsg, got %T", msg)
	}
	if !fsm.Canceled {
		t.Error("Expected Canceled=true")
	}
}

func TestFileBrowserHandleKeyPressCtrlC(t *testing.T) {
	entries := []FileEntry{
		{Name: "..", IsDir: true},
		{Name: "file1.txt", IsDir: false},
	}

	m := &FileBrowserModel{
		currentDir:    "/tmp",
		files:         entries,
		selectedIndex: 0,
		fileType:      FileTypeAll,
		active:        true,
	}

	updated, cmd := m.handleKeyPress(tea.KeyMsg{Type: tea.KeyCtrlC})
	m = updated.(*FileBrowserModel)

	if m.active {
		t.Error("Expected active=false after ctrl+c")
	}
	if cmd == nil {
		t.Fatal("Expected command after ctrl+c")
	}

	msg := cmd()
	fsm, ok := msg.(FileSelectedMsg)
	if !ok {
		t.Fatalf("Expected FileSelectedMsg, got %T", msg)
	}
	if !fsm.Canceled {
		t.Error("Expected Canceled=true")
	}
}

func TestFileBrowserHandleEnterDirectory(t *testing.T) {
	m, tmpDir := setupFileBrowserTest(t, FileTypeAll)

	entries, err := m.listDirectory(tmpDir)
	if err != nil {
		t.Fatalf("listDirectory() error: %v", err)
	}
	m.files = entries

	dirIdx := -1
	for i, e := range entries {
		if e.IsDir && e.Name != ".." {
			dirIdx = i
			break
		}
	}
	if dirIdx < 0 {
		t.Fatal("No directory entry found")
	}
	m.selectedIndex = dirIdx

	updated, cmd := m.handleEnter()
	m = updated.(*FileBrowserModel)

	if m.currentDir != entries[dirIdx].Path {
		t.Errorf("Expected currentDir '%s', got '%s'", entries[dirIdx].Path, m.currentDir)
	}
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after entering dir, got %d", m.selectedIndex)
	}
	if cmd == nil {
		t.Error("Expected loadDirectory command after entering dir")
	}
}

func TestFileBrowserHandleEnterFile(t *testing.T) {
	m, tmpDir := setupFileBrowserTest(t, FileTypeAll)

	entries, err := m.listDirectory(tmpDir)
	if err != nil {
		t.Fatalf("listDirectory() error: %v", err)
	}
	m.files = entries

	fileIdx := -1
	for i, e := range entries {
		if !e.IsDir {
			fileIdx = i
			break
		}
	}
	if fileIdx < 0 {
		t.Fatal("No file entry found")
	}
	m.selectedIndex = fileIdx

	updated, cmd := m.handleEnter()
	m = updated.(*FileBrowserModel)

	if m.active {
		t.Error("Expected active=false after selecting file")
	}
	if m.selectedPath != entries[fileIdx].Path {
		t.Errorf("Expected selectedPath '%s', got '%s'", entries[fileIdx].Path, m.selectedPath)
	}
	if cmd == nil {
		t.Fatal("Expected command after selecting file")
	}

	msg := cmd()
	fsm, ok := msg.(FileSelectedMsg)
	if !ok {
		t.Fatalf("Expected FileSelectedMsg, got %T", msg)
	}
	if fsm.Canceled {
		t.Error("Expected Canceled=false")
	}
	if fsm.Path != entries[fileIdx].Path {
		t.Errorf("Expected Path '%s', got '%s'", entries[fileIdx].Path, fsm.Path)
	}
}

func TestFileBrowserHandleEnterEmpty(t *testing.T) {
	m := &FileBrowserModel{
		currentDir:    "/tmp",
		files:         []FileEntry{},
		selectedIndex: 0,
		fileType:      FileTypeAll,
		active:        true,
	}

	updated, cmd := m.handleEnter()
	_ = updated

	if cmd != nil {
		t.Error("Expected nil command for empty file list")
	}
}

func TestFileBrowserNavigateUp(t *testing.T) {
	m, tmpDir := setupFileBrowserTest(t, FileTypeAll)

	parentDir := filepath.Dir(tmpDir)

	updated, cmd := m.navigateUp()
	m = updated.(*FileBrowserModel)

	if m.currentDir != parentDir {
		t.Errorf("Expected currentDir '%s', got '%s'", parentDir, m.currentDir)
	}
	if m.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after navigate up, got %d", m.selectedIndex)
	}
	if cmd == nil {
		t.Error("Expected loadDirectory command after navigate up")
	}
}

func TestFileBrowserNavigateUpFromRoot(t *testing.T) {
	m := &FileBrowserModel{
		currentDir:    "/",
		selectedIndex: 0,
		fileType:      FileTypeAll,
		active:        true,
	}

	updated, cmd := m.navigateUp()
	m = updated.(*FileBrowserModel)

	if m.currentDir != "/" {
		t.Errorf("Expected currentDir to remain '/', got '%s'", m.currentDir)
	}
	if cmd != nil {
		t.Error("Expected nil command when navigating up from root")
	}
}

func TestFileBrowserUpdateInactive(t *testing.T) {
	m := &FileBrowserModel{
		currentDir:    "/tmp",
		files:         []FileEntry{},
		selectedIndex: 0,
		fileType:      FileTypeAll,
		active:        false,
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	_ = updated

	if cmd != nil {
		t.Error("Expected nil command for inactive model")
	}
}

func TestFileBrowserViewContainsHeader(t *testing.T) {
	tests := []struct {
		fileType FileType
		header   string
	}{
		{FileTypeAll, "Select File"},
		{FileTypeDiskImage, "Select Disk Image"},
		{FileTypeISO, "Select ISO Image"},
	}

	for _, tt := range tests {
		m := &FileBrowserModel{
			currentDir: "/tmp",
			files:      []FileEntry{},
			fileType:   tt.fileType,
			active:     true,
		}

		view := m.View()
		if !strings.Contains(view, tt.header) {
			t.Errorf("View with fileType %d should contain '%s'", tt.fileType, tt.header)
		}
	}
}

func TestFileBrowserViewContainsCurrentDir(t *testing.T) {
	m := &FileBrowserModel{
		currentDir: "/some/test/path",
		files:      []FileEntry{},
		fileType:   FileTypeAll,
		active:     true,
	}

	view := m.View()
	if !strings.Contains(view, "/some/test/path") {
		t.Error("View should contain current directory path")
	}
}

func TestFileBrowserViewEmptyDir(t *testing.T) {
	m := &FileBrowserModel{
		currentDir: "/tmp",
		files:      []FileEntry{},
		fileType:   FileTypeAll,
		active:     true,
	}

	view := m.View()
	if !strings.Contains(view, "(empty directory)") {
		t.Error("View should show '(empty directory)' when no files")
	}
}

func TestFileBrowserViewContainsHelp(t *testing.T) {
	m := &FileBrowserModel{
		currentDir: "/tmp",
		files:      []FileEntry{},
		fileType:   FileTypeAll,
		active:     true,
	}

	view := m.View()
	if !strings.Contains(view, "Navigate") {
		t.Error("View should contain help text with 'Navigate'")
	}
	if !strings.Contains(view, "ESC Cancel") {
		t.Error("View should contain help text with 'ESC Cancel'")
	}
}

func TestFileBrowserGetSelectedPath(t *testing.T) {
	m := &FileBrowserModel{
		selectedPath: "/tmp/selected.qcow2",
	}

	if m.GetSelectedPath() != "/tmp/selected.qcow2" {
		t.Errorf("Expected '/tmp/selected.qcow2', got '%s'", m.GetSelectedPath())
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
