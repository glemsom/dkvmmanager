package models

import (
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/tui/components"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files")

// ansiRegexp strips ANSI escape codes from terminal output before golden comparison.
var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI color/style escape codes from a string.
func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

// goldenPath returns the path to a golden file in testdata/.
func goldenPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("testdata", name+".golden")
}

// assertGolden compares stripped View() output against a golden file.
// Set UPDATE_GOLDEN=1 or use -update-golden flag to regenerate.
func assertGolden(t *testing.T, name string, view string) {
	t.Helper()

	stripped := stripANSI(view)
	path := goldenPath(t, name)

	if *updateGolden || os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(stripped), 0644); err != nil {
			t.Fatalf("Failed to write golden file %s: %v", path, err)
		}
		return
	}

	golden, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Golden file not found: %s\nRun with UPDATE_GOLDEN=1 to create it", path)
	}

	if stripped != string(golden) {
		t.Errorf("Golden file mismatch for %s\n--- Expected (golden) ---\n%s\n--- Got (stripped) ---\n%s",
			name, string(golden), stripped)
	}
}

// TestGoldenMainMenu verifies the main menu view doesn't change unexpectedly.
func TestGoldenMainMenu(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	view := m.View()
	assertGolden(t, "main_menu", view)
}

// TestGoldenMainMenuWithVMs verifies main menu with VMs listed.
func TestGoldenMainMenuWithVMs(t *testing.T) {
	m := setupTestModelForScenarios(t)

	view := m.View()
	assertGolden(t, "main_menu_with_vms", view)
}

// TestGoldenConfigTab verifies the Configuration tab rendering.
func TestGoldenConfigTab(t *testing.T) {
	m := setupTestModelForScenarios(t)

	m.tabModel.SetActiveTab(components.TabConfiguration)
	view := m.View()
	assertGolden(t, "config_tab", view)
}

// TestGoldenPowerTab verifies the Power tab rendering.
func TestGoldenPowerTab(t *testing.T) {
	m := setupTestModelForScenarios(t)

	m.tabModel.SetActiveTab(components.TabPower)
	view := m.View()
	assertGolden(t, "power_tab", view)
}

// TestGoldenLVCreateForm verifies LV create form rendering.
func TestGoldenLVCreateForm(t *testing.T) {
	m := setupTestModelForScenarios(t)
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.lvCreateFormModel = NewLVCreateFormModel()
	m.lvCreateFormModel.SetSize(m.windowWidth-4, m.contentHeight()-2)
	m.currentView = ViewLVCreate
	m.breadcrumbs.AddItem("Configuration", "config", 1)
	m.breadcrumbs.AddItem("Create Logical Volume", "lv_create", 1)

	view := m.View()
	assertGolden(t, "lv_create_form", view)
}

// TestGoldenVMSelectEdit verifies VM selection view for editing.
func TestGoldenVMSelectEdit(t *testing.T) {
	m := setupTestModelForScenarios(t)
	m.windowWidth = 80
	m.windowHeight = 30

	m.showVMSelectionWithMode("edit", "No VMs available to edit")

	view := m.View()
	assertGolden(t, "vm_select_edit", view)
}

// TestGoldenVMSelectDelete verifies VM selection view for deletion.
func TestGoldenVMSelectDelete(t *testing.T) {
	m := setupTestModelForScenarios(t)
	m.windowWidth = 80
	m.windowHeight = 30

	m.showVMSelectionWithMode("delete", "No VMs available to delete")

	view := m.View()
	assertGolden(t, "vm_select_delete", view)
}

// TestGoldenVMCreateLoading verifies loading state for VM create sub-view.
func TestGoldenVMCreateLoading(t *testing.T) {
	m := setupTestModel(t)
	m.currentView = ViewVMCreate
	m.vmCreateModel = nil
	m.windowWidth = 80
	m.windowHeight = 30

	view := m.View()
	assertGolden(t, "vm_create_loading", view)
}

// TestGoldenVMEditLoading verifies loading state for VM edit sub-view.
func TestGoldenVMEditLoading(t *testing.T) {
	m := setupTestModel(t)
	m.currentView = ViewVMEdit
	m.vmEditModel = nil
	m.windowWidth = 80
	m.windowHeight = 30

	view := m.View()
	assertGolden(t, "vm_edit_loading", view)
}

// TestGoldenVMDeleteLoading verifies loading state for VM delete sub-view.
func TestGoldenVMDeleteLoading(t *testing.T) {
	m := setupTestModel(t)
	m.currentView = ViewVMDelete
	m.vmDeleteModel = nil
	m.windowWidth = 80
	m.windowHeight = 30

	view := m.View()
	assertGolden(t, "vm_delete_loading", view)
}

// TestGoldenQuitting verifies the quitting view.
func TestGoldenQuitting(t *testing.T) {
	m := setupTestModel(t)
	m.quitting = true

	view := m.View()
	assertGolden(t, "quitting", view)
}

// TestGoldenEmptyVMs verifies the empty VM list rendering.
func TestGoldenEmptyVMs(t *testing.T) {
	m := setupTestModel(t)
	m.windowWidth = 80
	m.windowHeight = 30

	view := m.View()
	assertGolden(t, "empty_vms", view)
}

// TestGoldenBreadcrumbsWithSubView verifies breadcrumb rendering in sub-view.
func TestGoldenBreadcrumbsWithSubView(t *testing.T) {
	m := setupTestModelForScenarios(t)
	m.tabModel.SetActiveTab(components.TabConfiguration)
	m.currentView = ViewVMCreate
	m.vmCreateModel = NewVMCreateModel(m.vmManager)
	m.breadcrumbs.AddItem("Configuration", "config", 1)
	m.breadcrumbs.AddItem("Add new VM", "vm_create", 1)

	view := m.View()
	assertGolden(t, "breadcrumbs_subview", view)
}

// TestStripANSI verifies the ANSI stripping function works correctly.
func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no ANSI codes",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "single ANSI code",
			input:    "\x1b[31mred text\x1b[0m",
			expected: "red text",
		},
		{
			name:     "multiple ANSI codes",
			input:    "\x1b[1;32mbold green\x1b[0m normal \x1b[34mblue\x1b[0m",
			expected: "bold green normal blue",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripANSI(tt.input)
			if got != tt.expected {
				t.Errorf("stripANSI() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestGoldenFileUpdateFlag verifies the golden path helper.
func TestGoldenFileUpdateFlag(t *testing.T) {
	path := goldenPath(t, "test_name")
	if !strings.Contains(path, "testdata") {
		t.Errorf("Expected golden path to contain 'testdata', got %s", path)
	}
	if !strings.Contains(path, "test_name.golden") {
		t.Errorf("Expected golden path to contain 'test_name.golden', got %s", path)
	}
}
