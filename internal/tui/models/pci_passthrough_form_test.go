package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// --- Test Fixtures ---

// mockPCIDevices returns a deterministic set of PCI devices for testing
// IOMMU groups: Group 1 (GPU + audio), Group 2 (USB), Group -1 (ungrouped)
func mockPCIDevices() []models.PCIDevice {
	return []models.PCIDevice{
		{
			Address:    "0000:01:00.0",
			Vendor:     "10de",
			Device:     "1b80",
			ClassCode:  "0300",
			Name:       "NVIDIA GeForce GTX 1080",
			IsGPU:      true,
			IsUSB:      false,
			IOMMUGroup: 1,
		},
		{
			Address:    "0000:01:00.1",
			Vendor:     "10de",
			Device:     "10f0",
			ClassCode:  "0403",
			Name:       "NVIDIA GP104 High Definition Audio Controller",
			IsGPU:      false,
			IsUSB:      false,
			IOMMUGroup: 1,
		},
		{
			Address:    "0000:00:14.0",
			Vendor:     "8086",
			Device:     "a12f",
			ClassCode:  "0c03",
			Name:       "Intel USB 3.0 xHCI Controller",
			IsGPU:      false,
			IsUSB:      true,
			IOMMUGroup: 2,
		},
		{
			Address:    "0000:00:1f.0",
			Vendor:     "8086",
			Device:     "a150",
			ClassCode:  "0680",
			Name:       "Intel LPC Controller",
			IsGPU:      false,
			IsUSB:      false,
			IOMMUGroup: -1,
		},
		{
			Address:    "0000:03:00.0",
			Vendor:     "144d",
			Device:     "a808",
			ClassCode:  "0108",
			Name:       "Samsung NVMe SSD Controller SM981",
			IsGPU:      false,
			IsUSB:      false,
			IOMMUGroup: 3,
		},
	}
}

// newTestPCIForm creates a PCI passthrough form with mock devices (no real VM manager needed)
func newTestPCIForm(t *testing.T) *PCIPassthroughFormModel {
	t.Helper()
	vmManager := createTestVMManager(t)

	formModel, err := NewPCIPassthroughFormModel(vmManager)
	if err != nil {
		// In CI, real scanning may fail; construct form manually with mock devices
		formModel = &PCIPassthroughFormModel{
			vmManager: vmManager,
			devices:   mockPCIDevices(),
			selected:  make(map[string]bool),
			errors:    make(map[string]string),
		}
		formModel.buildIOMMUGroups()
		formModel.BuildPositions()
	}

	// Apply mock devices regardless of scan result
	formModel.devices = mockPCIDevices()
	formModel.buildIOMMUGroups()
	formModel.BuildPositions()

	// Apply window size to enable viewport
	formModel.Update(tea.WindowSizeMsg{Width: 80, Height: 25})

	return formModel
}

// --- Phase 1a: ROM Removal Tests ---

// TestPCIFORMNoROMInRender verifies rendered output contains no ROM labels
func TestPCIFORMNoROMInRender(t *testing.T) {
	m := newTestPCIForm(t)

	// Select a device
	m.selected["0000:01:00.0"] = true
	m.BuildPositions()
	m.syncViewport()

	view := m.View()
	if strings.Contains(view, "ROM:") || strings.Contains(view, "rom_path") || strings.Contains(view, "(optional ROM)") {
		t.Errorf("View should not contain any ROM references.\nView:\n%s", view)
	}
}

// TestPCIFORMOneLinePerDevice verifies each device occupies exactly one toggle line
// (no ROM path fields remaining). Group headers add extra positions but each device
// should have exactly one toggle position.
func TestPCIFORMOneLinePerDevice(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()

	// Count toggle positions — should equal number of devices
	toggleCount := 0
	for _, pos := range positions {
		if pos.Kind == form.FocusToggle {
			toggleCount++
		}
	}

	if toggleCount != len(m.devices) {
		t.Errorf("Expected %d toggle positions (one per device), got %d", len(m.devices), toggleCount)
	}

	// Also verify save button exists
	saveCount := 0
	for _, pos := range positions {
		if pos.Kind == form.FocusButton && pos.Key == "save" {
			saveCount++
		}
	}
	if saveCount != 1 {
		t.Errorf("Expected 1 save button position, got %d", saveCount)
	}
}

// --- Phase 1b: Device Line Reordering Tests ---

// TestPCIFORMAddressFirst verifies PCI address appears right after toggle indicator
func TestPCIFORMAddressFirst(t *testing.T) {
	m := newTestPCIForm(t)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 25})

	positions := m.BuildPositions()

	// Find the first toggle position index
	firstToggleIdx := -1
	for i, pos := range positions {
		if pos.Kind == form.FocusToggle {
			firstToggleIdx = i
			break
		}
	}
	if firstToggleIdx < 0 {
		t.Fatal("No toggle positions found")
	}

	// Focus on the first toggle to get it rendered
	m.SetFocusIndex(firstToggleIdx)
	m.syncViewport()

	lines := m.renderAllLines()

	// Find the device line for the first toggle (contains its PCI address)
	addr := "0000:01:00.0"
	var deviceLine string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && strings.Contains(trimmed, addr) &&
			!strings.HasPrefix(trimmed, "Tab ") && !strings.Contains(trimmed, "IOMMU Group") &&
			!strings.Contains(trimmed, "Ungrouped") && !strings.Contains(trimmed, "Save") {
			deviceLine = trimmed
			break
		}
	}

	if deviceLine == "" {
		t.Fatalf("No device line found for address %s", addr)
	}

	// After the toggle indicator [X] or [ ], the PCI address should appear
	// Format: [ ] 0000:01:00.0 [GPU] NVIDIA ...
	// Check that address appears before the device type tag and name
	toggleEnd := strings.Index(deviceLine, "]")
	if toggleEnd == -1 {
		t.Fatalf("No toggle indicator found in device line: %q", deviceLine)
	}

	afterToggle := deviceLine[toggleEnd+1:]

	// Address should be the first thing after the toggle
	addrIdx := strings.Index(afterToggle, addr)
	tagIdx := strings.Index(afterToggle, "[GPU]")
	nameIdx := strings.Index(afterToggle, "NVIDIA")

	if addrIdx == -1 {
		t.Fatalf("PCI address %q not found in device line: %s", addr, deviceLine)
	}
	if tagIdx != -1 && addrIdx > tagIdx {
		t.Errorf("PCI address should appear before type tag. Address at %d, tag at %d", addrIdx, tagIdx)
	}
	if nameIdx != -1 && addrIdx > nameIdx {
		t.Errorf("PCI address should appear before device name. Address at %d, name at %d", addrIdx, nameIdx)
	}
}

// TestPCIFORMAllDeviceLinesHaveAddress verifies every device toggle shows the PCI address
func TestPCIFORMAllDeviceLinesHaveAddress(t *testing.T) {
	m := newTestPCIForm(t)
	m.syncViewport()

	lines := m.renderAllLines()
	expectedAddrs := map[string]bool{
		"0000:01:00.0": false,
		"0000:01:00.1": false,
		"0000:00:14.0": false,
		"0000:00:1f.0": false,
		"0000:03:00.0": false,
	}

	for _, line := range lines {
		for addr := range expectedAddrs {
			if strings.Contains(line, addr) {
				expectedAddrs[addr] = true
			}
		}
	}

	for addr, found := range expectedAddrs {
		if !found {
			t.Errorf("PCI address %s not found in any rendered line", addr)
		}
	}
}

// TestPCIFORMDeviceTypeTag verifies [GPU] and [USB] tags render correctly
func TestPCIFORMDeviceTypeTag(t *testing.T) {
	m := newTestPCIForm(t)
	m.syncViewport()

	lines := m.renderAllLines()

	// GPU tag
	gpuFound := false
	// USB tag
	usbFound := false
	for _, line := range lines {
		if strings.Contains(line, "[GPU]") {
			gpuFound = true
		}
		if strings.Contains(line, "[USB]") {
			usbFound = true
		}
	}

	if !gpuFound {
		t.Error("Expected [GPU] tag for GPU device in rendered output")
	}
	if !usbFound {
		t.Error("Expected [USB] tag for USB device in rendered output")
	}
}

// --- Phase 1c: IOMMU Grouping Tests ---

// TestPCIFORMIOMMUGroupsBuilt verifies IOMMU groups are correctly indexed
func TestPCIFORMIOMMUGroupsBuilt(t *testing.T) {
	m := newTestPCIForm(t)

	if m.iommuGroups == nil {
		t.Fatal("iommuGroups map should not be nil")
	}

	// We expect 4 groups: 1 (2 devices), 2 (1 device), 3 (1 device), -1 (1 device = ungrouped)
	expectedGroups := 4
	if len(m.iommuGroups) != expectedGroups {
		t.Errorf("Expected %d IOMMU groups, got %d", expectedGroups, len(m.iommuGroups))
	}

	// Group 1 should have 2 devices
	if len(m.iommuGroups[1]) != 2 {
		t.Errorf("Expected 2 devices in IOMMU group 1, got %d", len(m.iommuGroups[1]))
	}

	// Ungrouped (-1) should have 1 device
	if len(m.iommuGroups[-1]) != 1 {
		t.Errorf("Expected 1 ungrouped device, got %d", len(m.iommuGroups[-1]))
	}
}

// TestPCIFORMGroupHeadersRendered verifies group headers appear in rendered output
func TestPCIFORMGroupHeadersRendered(t *testing.T) {
	m := newTestPCIForm(t)
	m.syncViewport()

	lines := m.renderAllLines()

	var headerCount int
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "IOMMU Group") ||
			strings.Contains(trimmed, "Ungrouped") {
			headerCount++
		}
	}

	// Should have 4 headers (groups 1, 2, 3, and ungrouped)
	if headerCount != 4 {
		t.Errorf("Expected 4 group headers in output, got %d", headerCount)
	}
}

// TestPCIFORMGroupHeaderFormat verifies group header shows correct device count
func TestPCIFORMGroupHeaderFormat(t *testing.T) {
	m := newTestPCIForm(t)
	m.syncViewport()

	lines := m.renderAllLines()

	// Group 1 header should mention 2 devices
	found := false
	for _, line := range lines {
		if strings.Contains(line, "IOMMU Group 1") && strings.Contains(line, "2 devices") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'IOMMU Group 1' header with '2 devices' in output")
	}
}

// TestPCIFORMGroupAutoSelect verifies toggling one device in a group selects all in that group
func TestPCIFORMGroupAutoSelect(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()

	// Find the toggle position for 0000:01:00.0 (in IOMMU group 1)
	for i, pos := range positions {
		if pos.Kind == form.FocusToggle && pos.Key == "0000:01:00.0" {
			m.SetFocusIndex(i)
			break
		}
	}

	// Toggle the device
	m.toggleDevice("0000:01:00.0")

	// Both devices in group 1 should be selected
	if !m.selected["0000:01:00.0"] {
		t.Error("Toggled device 0000:01:00.0 should be selected")
	}
	if !m.selected["0000:01:00.1"] {
		t.Error("Sibling device 0000:01:00.1 in same IOMMU group should be auto-selected")
	}

	// Devices in other groups should not be affected
	if m.selected["0000:00:14.0"] {
		t.Error("USB device in different group should not be selected")
	}
}

// TestPCIFORMGroupAutoDeselect verifies toggling off a device deselects the entire group
func TestPCIFORMGroupAutoDeselect(t *testing.T) {
	m := newTestPCIForm(t)

	// Pre-select both devices in group 1
	m.selected["0000:01:00.0"] = true
	m.selected["0000:01:00.1"] = true

	// Toggle off one device
	m.toggleDevice("0000:01:00.0")

	// Both should be deselected
	if m.selected["0000:01:00.0"] {
		t.Error("Toggled-off device 0000:01:00.0 should be deselected")
	}
	if m.selected["0000:01:00.1"] {
		t.Error("Sibling device 0000:01:00.1 should also be deselected")
	}
}

// TestPCIFORMUngroupedDeviceNoAutoSelect verifies ungrouped devices (IOMMU -1) are toggled individually
func TestPCIFORMUngroupedDeviceNoAutoSelect(t *testing.T) {
	m := newTestPCIForm(t)

	// Toggle the ungrouped device
	m.toggleDevice("0000:00:1f.0")

	// Only this device should be selected
	if !m.selected["0000:00:1f.0"] {
		t.Error("Ungrouped device should be selected after toggle")
	}
	// No other device should be affected
	if m.selected["0000:01:00.0"] {
		t.Error("Device in IOMMU group 1 should not be auto-selected when toggling ungrouped device")
	}
}

// TestPCIFORMNavigationIncludesGroupHeaders verifies that Tab/Up/Down navigate through group headers.
// After migration to form.ScrollableForm, FocusHeader positions are focusable tab-stops.
func TestPCIFORMNavigationIncludesGroupHeaders(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()

	// Positions can be FocusToggle, FocusHeader, or FocusButton
	for i, pos := range positions {
		if pos.Kind != form.FocusToggle && pos.Kind != form.FocusButton && pos.Kind != form.FocusHeader {
			t.Errorf("Position %d has unexpected Kind %v (expected FocusToggle, FocusButton, or FocusHeader)", i, pos.Kind)
		}
	}

	// Count expected FocusHeader positions (should be 4: groups 1, 2, 3, and ungrouped)
	headerCount := 0
	for _, pos := range positions {
		if pos.Kind == form.FocusHeader {
			headerCount++
		}
	}
	if headerCount != 4 {
		t.Errorf("Expected 4 FocusHeader positions, got %d", headerCount)
	}

	// Verify that navigating via Tab DOES land on group headers (they are now focusable)
	m.SetFocusIndex(0)
	visited := make(map[int]bool)
	headerVisited := false
	for {
		visited[m.focusIndex] = true
		pos := positions[m.focusIndex]
		if pos.Kind == form.FocusHeader {
			headerVisited = true
		}
		m.moveFocus(1)
		if visited[m.focusIndex] || m.focusIndex >= len(positions)-1 {
			break
		}
	}

	if !headerVisited {
		t.Error("Expected Tab navigation to visit at least one FocusHeader position")
	}
}

// TestPCIFORMSavePreservesSelectedDevices verifies saving config includes all selected devices
func TestPCIFORMSavePreservesSelectedDevices(t *testing.T) {
	m := newTestPCIForm(t)

	// Select a device from group 1 (should auto-select its sibling)
	m.selected["0000:01:00.0"] = true
	// Manually ensure the group is fully selected (in case toggleDevice logic is wrong)
	m.selected["0000:01:00.1"] = true

	// Navigate to save button
	positions := m.BuildPositions()
	m.SetFocusIndex(len(positions) - 1)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		// Validation may fail in test env (no lspci); skip the test
		t.Skip("PCI validation failed (lspci not available in test env)")
	}

	msg := cmd()
	if _, ok := msg.(PCIPassthroughUpdatedMsg); !ok {
		t.Errorf("Expected PCIPassthroughUpdatedMsg, got %T", msg)
	}

	// Verify saved config
	saved, err := m.vmManager.GetPCIPassthroughConfig()
	if err != nil {
		t.Fatalf("Failed to load saved PCI config: %v", err)
	}

	if len(saved.Devices) != 2 {
		t.Errorf("Expected 2 saved devices, got %d", len(saved.Devices))
	}

	addrSet := make(map[string]bool)
	for _, d := range saved.Devices {
		addrSet[d.Address] = true
	}
	if !addrSet["0000:01:00.0"] || !addrSet["0000:01:00.1"] {
		t.Errorf("Saved devices should include both IOMMU group 1 devices")
	}
}

// TestPCIFORMNoROMInSavedConfig verifies saved devices have no ROM path set
func TestPCIFORMNoROMInSavedConfig(t *testing.T) {
	m := newTestPCIForm(t)

	m.selected["0000:01:00.0"] = true
	m.selected["0000:01:00.1"] = true
	positions := m.BuildPositions()
	m.SetFocusIndex(len(positions) - 1)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		// Validation may fail in test env (no lspci); skip the test
		t.Skip("PCI validation failed (lspci not available in test env)")
	}
	cmd()

	saved, err := m.vmManager.GetPCIPassthroughConfig()
	if err != nil {
		t.Fatalf("Failed to load saved PCI config: %v", err)
	}

	for _, d := range saved.Devices {
		if d.ROMPath != "" {
			t.Errorf("Device %s has ROMPath=%q, expected empty", d.Address, d.ROMPath)
		}
	}
}

// TestPCIFORMEmptyDevices verifies behavior when no devices are found
func TestPCIFORMEmptyDevices(t *testing.T) {
	vmManager := createTestVMManager(t)

	m := &PCIPassthroughFormModel{
		vmManager:   vmManager,
		devices:     []models.PCIDevice{},
		selected:    make(map[string]bool),
		errors:      make(map[string]string),
		iommuGroups: make(map[int][]*models.PCIDevice),
	}
	m.BuildPositions()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 25})

	positions := m.BuildPositions()
	if len(positions) != 2 {
		t.Errorf("Expected 2 positions (save + apply kernel buttons), got %d", len(positions))
	}
	if positions[0].Kind != form.FocusButton || positions[0].Key != "save" {
		t.Errorf("Expected save button as first position, got Kind=%v Key=%s", positions[0].Kind, positions[0].Key)
	}
	if positions[1].Kind != form.FocusButton || positions[1].Key != "apply_kernel" {
		t.Errorf("Expected apply kernel button as second position, got Kind=%v Key=%s", positions[1].Kind, positions[1].Key)
	}
}

// TestPCIFORMViewShowsNoDevicesMessage verifies "no devices" message when scan finds nothing
func TestPCIFORMViewShowsNoDevicesMessage(t *testing.T) {
	vmManager := createTestVMManager(t)

	m := &PCIPassthroughFormModel{
		vmManager:   vmManager,
		devices:     []models.PCIDevice{},
		selected:    make(map[string]bool),
		errors:      make(map[string]string),
		iommuGroups: make(map[int][]*models.PCIDevice),
	}
	m.BuildPositions()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 25})
	m.syncViewport()

	view := m.View()
	if !strings.Contains(view, "No PCI devices found") {
		t.Errorf("Expected 'No PCI devices found' message, got:\n%s", view)
	}
}

// TestPCIFORMToggleViaEnterKey verifies Enter key toggles device on/off
func TestPCIFORMToggleViaEnterKey(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()

	// Find first toggle
	for i, pos := range positions {
		if pos.Kind == form.FocusToggle {
			m.SetFocusIndex(i)
			break
		}
	}

	addr := positions[m.focusIndex].Key

	// Enter to select
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.selected[addr] {
		t.Error("Device should be selected after Enter")
	}

	// Enter to deselect
	m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.selected[addr] {
		t.Error("Device should be deselected after second Enter")
	}
}

// TestPCIFORMToggleViaSpaceKey verifies Space key toggles device on/off
func TestPCIFORMToggleViaSpaceKey(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()

	// Find first toggle
	for i, pos := range positions {
		if pos.Kind == form.FocusToggle {
			m.SetFocusIndex(i)
			break
		}
	}

	addr := positions[m.focusIndex].Key

	// Space to select
	m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !m.selected[addr] {
		t.Error("Device should be selected after Space")
	}
}

// TestPCIFORMNavigationTabCycle verifies Tab cycles through all positions
func TestPCIFORMNavigationTabCycle(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()

	// Tab forward — first position is now a FocusHeader (group header)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(*PCIPassthroughFormModel)

	// After first Tab from index 0, should be at index 1 (second position)
	if m.focusIndex != 1 {
		t.Errorf("After first Tab, expected focusIndex=1, got %d", m.focusIndex)
	}

	// Tab through to the apply kernel button (last position)
	for i := 0; i < len(positions); i++ {
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
		m = model.(*PCIPassthroughFormModel)
	}

	if m.focusIndex != len(positions)-1 {
		t.Errorf("Expected focus at last position, got %d", m.focusIndex)
	}
	if positions[m.focusIndex].Kind != form.FocusButton || positions[m.focusIndex].Key != "apply_kernel" {
		t.Errorf("Expected to land on apply kernel button, got Kind=%v Key=%s", positions[m.focusIndex].Kind, positions[m.focusIndex].Key)
	}
}

// TestPCIFORMNavigationShiftTabBackward verifies Shift+Tab moves focus backward
func TestPCIFORMNavigationShiftTabBackward(t *testing.T) {
	m := newTestPCIForm(t)

	// Tab forward once
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(*PCIPassthroughFormModel)
	firstFocusable := m.focusIndex

	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = model.(*PCIPassthroughFormModel)
	_ = m.focusIndex // second position reached

	// Shift+Tab should go back
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = model.(*PCIPassthroughFormModel)
	if m.focusIndex != firstFocusable {
		t.Errorf("After Shift+Tab, focusIndex = %d, want %d", m.focusIndex, firstFocusable)
	}
}

// TestPCIFORMNavigationUpArrow verifies Up arrow moves focus backward
func TestPCIFORMNavigationUpArrow(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()

	// Find the second toggle position
	toggleIdx := 0
	found := 0
	for i, pos := range positions {
		if pos.Kind == form.FocusToggle {
			found++
			if found == 2 {
				toggleIdx = i
				break
			}
		}
	}
	if found < 2 {
		t.Skip("Need at least 2 toggles")
	}

	m.SetFocusIndex(toggleIdx)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = model.(*PCIPassthroughFormModel)

	// Should have moved back
	if m.focusIndex >= toggleIdx {
		t.Errorf("After Up from index %d, focus should move back, got %d", toggleIdx, m.focusIndex)
	}
}

// TestPCIFORMNavigationDownArrow verifies Down arrow moves focus forward
func TestPCIFORMNavigationDownArrow(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()

	// Find the first toggle position
	firstToggleIdx := -1
	for i, pos := range positions {
		if pos.Kind == form.FocusToggle {
			firstToggleIdx = i
			break
		}
	}
	if firstToggleIdx < 0 {
		t.Fatal("No toggle positions found")
	}

	m.SetFocusIndex(firstToggleIdx)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*PCIPassthroughFormModel)

	// Should have moved forward
	if m.focusIndex <= firstToggleIdx {
		t.Errorf("After Down from index %d, focus should move forward, got %d", firstToggleIdx, m.focusIndex)
	}
}

// TestPCIFORMNavigationBoundaryUp verifies Up at first position stays at 0
func TestPCIFORMNavigationBoundaryUp(t *testing.T) {
	m := newTestPCIForm(t)

	m.SetFocusIndex(0)
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = model.(*PCIPassthroughFormModel)

	if m.focusIndex != 0 {
		t.Errorf("Up at position 0 should stay at 0, got %d", m.focusIndex)
	}
}

// TestPCIFORMNavigationBoundaryDown verifies Down at last position stays at end
func TestPCIFORMNavigationBoundaryDown(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()
	m.SetFocusIndex(len(positions) - 1)
	lastIdx := m.focusIndex
	model, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(*PCIPassthroughFormModel)

	if m.focusIndex != lastIdx {
		t.Errorf("Down at last position %d should stay, got %d", lastIdx, m.focusIndex)
	}
}

// TestPCIFormIOMMUGroupHeaderCount verifies that BuildPositions creates group headers
func TestPCIFormIOMMUGroupHeaderCount(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()

	// Count group header positions
	headerCount := 0
	for _, pos := range positions {
		if pos.Kind == form.FocusHeader {
			headerCount++
		}
	}

	if headerCount != 4 {
		t.Errorf("Expected 4 group header positions (groups 1,2,3,-1), got %d", headerCount)
	}
}

// TestPCIFormGroupHeaderNavigable verifies that when navigating via Tab,
// the focusIndex DOES point to group headers (FocusHeader positions are focusable).
func TestPCIFormGroupHeaderNavigable(t *testing.T) {
	m := newTestPCIForm(t)

	positions := m.BuildPositions()

	m.SetFocusIndex(0)
	visited := make(map[int]bool)
	headerVisitCount := 0

	for {
		visited[m.focusIndex] = true
		pos := positions[m.focusIndex]
		if pos.Kind == form.FocusHeader {
			headerVisitCount++
		}

		// Check if we've reached the apply_kernel button (last position)
		if m.focusIndex >= len(positions)-1 {
			break
		}

		m.moveFocus(1)

		// Stop if we've looped back
		if m.focusIndex <= 0 && len(visited) > 1 {
			break
		}
	}

	// Verify we visited at least the header positions
	if headerVisitCount == 0 {
		t.Error("Expected to visit at least one FocusHeader position during Tab navigation")
	}

	// Verify we visited all toggle positions + buttons
	toggleCount := 0
	saveCount := 0
	applyCount := 0
	for i, pos := range positions {
		if visited[i] {
			if pos.Kind == form.FocusToggle {
				toggleCount++
			} else if pos.Kind == form.FocusButton && pos.Key == "save" {
				saveCount++
			} else if pos.Kind == form.FocusButton && pos.Key == "apply_kernel" {
				applyCount++
			}
		}
	}

	if toggleCount != len(m.devices) {
		t.Errorf("Expected to visit all %d toggles, visited %d", len(m.devices), toggleCount)
	}
	if saveCount != 1 {
		t.Errorf("Expected to visit 1 save button, visited %d", saveCount)
	}
	if applyCount != 1 {
		t.Errorf("Expected to visit 1 apply_kernel button, visited %d", applyCount)
	}
}

// TestPCIFORMSelectMultipleGroups verifies selecting devices from different groups works
func TestPCIFORMSelectMultipleGroups(t *testing.T) {
	m := newTestPCIForm(t)

	// Select group 1
	m.selected["0000:01:00.0"] = true
	m.selected["0000:01:00.1"] = true

	// Select group 3
	m.selected["0000:03:00.0"] = true

	// Navigate to save
	positions := m.BuildPositions()
	m.SetFocusIndex(len(positions) - 1)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		// Validation may fail in test env (no lspci); skip the test
		t.Skip("PCI validation failed (lspci not available in test env)")
	}
	cmd()

	saved, err := m.vmManager.GetPCIPassthroughConfig()
	if err != nil {
		t.Fatalf("Failed to load saved PCI config: %v", err)
	}

	if len(saved.Devices) != 3 {
		t.Errorf("Expected 3 saved devices, got %d", len(saved.Devices))
	}
}

// TestPCIFORMRenderGroupHeaderSelectionStatus verifies group header shows selection state
func TestPCIFORMRenderGroupHeaderSelectionStatus(t *testing.T) {
	m := newTestPCIForm(t)

	// Select all devices in group 1
	m.selected["0000:01:00.0"] = true
	m.selected["0000:01:00.1"] = true

	m.syncViewport()
	lines := m.renderAllLines()

	// Find group 1 header line
	var group1Line string
	for _, line := range lines {
		if strings.Contains(line, "IOMMU Group 1") {
			group1Line = line
			break
		}
	}

	if group1Line == "" {
		t.Fatal("IOMMU Group 1 header not found in rendered output")
	}

	// When all devices in group are selected, header should show some selection indicator
	if !strings.Contains(group1Line, "selected") && !strings.Contains(group1Line, "[") {
		t.Errorf("Group 1 header with all devices selected should show selection status. Got: %s", group1Line)
	}
}

// TestPCIPassthroughFormModelInterface verifies the form implements form.FormModel
func TestPCIPassthroughFormModelInterface(t *testing.T) {
	vmManager := createTestVMManager(t)

	m, err := NewPCIPassthroughFormModel(vmManager)
	if err != nil {
		// In CI, real scanning may fail; construct manually
		m = &PCIPassthroughFormModel{
			vmManager:   vmManager,
			devices:     mockPCIDevices(),
			selected:    make(map[string]bool),
			errors:      make(map[string]string),
		}
		m.buildIOMMUGroups()
		m.BuildPositions()
	} else if len(m.devices) == 0 {
		// Scanning succeeded but found no devices - use mock devices for this test
		m.devices = mockPCIDevices()
		m.buildIOMMUGroups()
		m.BuildPositions()
	}

	// Verify it implements form.FormModel
	var _ form.FormModel = m

	// Verify BuildPositions returns expected positions
	positions := m.BuildPositions()
	if len(positions) == 0 {
		t.Fatal("Expected at least one position")
	}

	// Verify FocusHeader positions exist (for IOMMU groups)
	headerCount := 0
	for _, pos := range positions {
		if pos.Kind == form.FocusHeader {
			headerCount++
		}
	}
	if headerCount == 0 {
		t.Error("Expected at least one FocusHeader position for IOMMU groups")
	}

	// Last position should be apply_kernel button
	lastPos := positions[len(positions)-1]
	if lastPos.Kind != form.FocusButton || lastPos.Key != "apply_kernel" {
		t.Errorf("Expected last position to be apply_kernel button, got Kind=%v Key=%s", lastPos.Kind, lastPos.Key)
	}

	// Second-to-last should be save button
	savePos := positions[len(positions)-2]
	if savePos.Kind != form.FocusButton || savePos.Key != "save" {
		t.Errorf("Expected second-to-last position to be save button, got Kind=%v Key=%s", savePos.Kind, savePos.Key)
	}
}
