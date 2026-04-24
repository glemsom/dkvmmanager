package models

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/models"
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

	form, err := NewPCIPassthroughFormModel(vmManager)
	if err != nil {
		// In CI, real scanning may fail; construct form manually with mock devices
		form = &PCIPassthroughFormModel{
			vmManager: vmManager,
			devices:   mockPCIDevices(),
			selected:  make(map[string]bool),
			errors:    make(map[string]string),
		}
		form.buildIOMMUGroups()
		form.rebuildPositions()
	}

	// Apply mock devices regardless of scan result
	form.devices = mockPCIDevices()
	form.buildIOMMUGroups()
	form.rebuildPositions()

	// Apply window size to enable viewport
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 25})

	return form
}

// --- Phase 1a: ROM Removal Tests ---

// TestPCIFORMNoROMInRender verifies rendered output contains no ROM labels
func TestPCIFORMNoROMInRender(t *testing.T) {
	form := newTestPCIForm(t)

	// Select a device
	form.selected["0000:01:00.0"] = true
	form.rebuildPositions()
	form.syncViewport()

	view := form.View()
	if strings.Contains(view, "ROM:") || strings.Contains(view, "rom_path") || strings.Contains(view, "(optional ROM)") {
		t.Errorf("View should not contain any ROM references.\nView:\n%s", view)
	}
}

// TestPCIFORMOneLinePerDevice verifies each device occupies exactly one toggle line
// (no ROM path fields remaining). Group headers add extra positions but each device
// should have exactly one toggle position.
func TestPCIFORMOneLinePerDevice(t *testing.T) {
	form := newTestPCIForm(t)

	// Count toggle positions — should equal number of devices
	toggleCount := 0
	for _, pos := range form.positions {
		if pos.kind == pciToggle {
			toggleCount++
		}
	}

	if toggleCount != len(form.devices) {
		t.Errorf("Expected %d toggle positions (one per device), got %d", len(form.devices), toggleCount)
	}

	// Also verify save button exists
	saveCount := 0
	for _, pos := range form.positions {
		if pos.kind == pciSave {
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
	form := newTestPCIForm(t)
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 25})

	// Find the first toggle position index
	firstToggleIdx := -1
	for i, pos := range form.positions {
		if pos.kind == pciToggle {
			firstToggleIdx = i
			break
		}
	}
	if firstToggleIdx < 0 {
		t.Fatal("No toggle positions found")
	}

	// Focus on the first toggle to get it rendered
	form.focusIndex = firstToggleIdx
	form.syncViewport()

	lines := form.renderAllLines()

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
	form := newTestPCIForm(t)
	form.syncViewport()

	lines := form.renderAllLines()
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
	form := newTestPCIForm(t)
	form.syncViewport()

	lines := form.renderAllLines()

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
	form := newTestPCIForm(t)

	if form.iommuGroups == nil {
		t.Fatal("iommuGroups map should not be nil")
	}

	// We expect 4 groups: 1 (2 devices), 2 (1 device), 3 (1 device), -1 (1 device = ungrouped)
	expectedGroups := 4
	if len(form.iommuGroups) != expectedGroups {
		t.Errorf("Expected %d IOMMU groups, got %d", expectedGroups, len(form.iommuGroups))
	}

	// Group 1 should have 2 devices
	if len(form.iommuGroups[1]) != 2 {
		t.Errorf("Expected 2 devices in IOMMU group 1, got %d", len(form.iommuGroups[1]))
	}

	// Ungrouped (-1) should have 1 device
	if len(form.iommuGroups[-1]) != 1 {
		t.Errorf("Expected 1 ungrouped device, got %d", len(form.iommuGroups[-1]))
	}
}

// TestPCIFORMGroupHeadersRendered verifies group headers appear in rendered output
func TestPCIFORMGroupHeadersRendered(t *testing.T) {
	form := newTestPCIForm(t)
	form.syncViewport()

	lines := form.renderAllLines()

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
	form := newTestPCIForm(t)
	form.syncViewport()

	lines := form.renderAllLines()

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
	form := newTestPCIForm(t)

	// Find the toggle position for 0000:01:00.0 (in IOMMU group 1)
	for i, pos := range form.positions {
		if pos.kind == pciToggle && pos.deviceAddr == "0000:01:00.0" {
			form.focusIndex = i
			break
		}
	}

	// Toggle the device
	form.toggleDevice("0000:01:00.0")

	// Both devices in group 1 should be selected
	if !form.selected["0000:01:00.0"] {
		t.Error("Toggled device 0000:01:00.0 should be selected")
	}
	if !form.selected["0000:01:00.1"] {
		t.Error("Sibling device 0000:01:00.1 in same IOMMU group should be auto-selected")
	}

	// Devices in other groups should not be affected
	if form.selected["0000:00:14.0"] {
		t.Error("USB device in different group should not be selected")
	}
}

// TestPCIFORMGroupAutoDeselect verifies toggling off a device deselects the entire group
func TestPCIFORMGroupAutoDeselect(t *testing.T) {
	form := newTestPCIForm(t)

	// Pre-select both devices in group 1
	form.selected["0000:01:00.0"] = true
	form.selected["0000:01:00.1"] = true

	// Toggle off one device
	form.toggleDevice("0000:01:00.0")

	// Both should be deselected
	if form.selected["0000:01:00.0"] {
		t.Error("Toggled-off device 0000:01:00.0 should be deselected")
	}
	if form.selected["0000:01:00.1"] {
		t.Error("Sibling device 0000:01:00.1 should also be deselected")
	}
}

// TestPCIFORMUngroupedDeviceNoAutoSelect verifies ungrouped devices (IOMMU -1) are toggled individually
func TestPCIFORMUngroupedDeviceNoAutoSelect(t *testing.T) {
	form := newTestPCIForm(t)

	// Toggle the ungrouped device
	form.toggleDevice("0000:00:1f.0")

	// Only this device should be selected
	if !form.selected["0000:00:1f.0"] {
		t.Error("Ungrouped device should be selected after toggle")
	}
	// No other device should be affected
	if form.selected["0000:01:00.0"] {
		t.Error("Device in IOMMU group 1 should not be auto-selected when toggling ungrouped device")
	}
}

// TestPCIFORMNavigationSkipsGroupHeaders verifies Tab/Up/Down skip group header positions
func TestPCIFORMNavigationSkipsGroupHeaders(t *testing.T) {
	form := newTestPCIForm(t)

	// Positions can be pciToggle, pciGroupHeader, or pciSave
	for i, pos := range form.positions {
		if pos.kind != pciToggle && pos.kind != pciSave && pos.kind != pciGroupHeader {
			t.Errorf("Position %d has unexpected kind %d (expected pciToggle, pciSave, or pciGroupHeader)", i, pos.kind)
		}
	}

	// Group headers exist in positions but are skipped during navigation.
	// Verify that navigating via Tab never lands on a group header.
	form.focusIndex = 0
	if form.currentPos().kind == pciGroupHeader {
		form.moveFocus(1) // move past initial header
	}
	visited := make(map[int]bool)
	for {
		visited[form.focusIndex] = true
		pos := form.currentPos()
		if pos.kind == pciGroupHeader {
			t.Errorf("Focus should never land on pciGroupHeader index %d", form.focusIndex)
		}
		form.moveFocus(1)
		if visited[form.focusIndex] || form.focusIndex >= len(form.positions)-1 {
			break
		}
	}
}

// TestPCIFORMSavePreservesSelectedDevices verifies saving config includes all selected devices
func TestPCIFORMSavePreservesSelectedDevices(t *testing.T) {
	form := newTestPCIForm(t)

	// Select a device from group 1 (should auto-select its sibling)
	form.selected["0000:01:00.0"] = true
	// Manually ensure the group is fully selected (in case toggleDevice logic is wrong)
	form.selected["0000:01:00.1"] = true

	// Navigate to save button
	form.focusIndex = len(form.positions) - 1

	_, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		// Validation may fail in test env (no lspci); skip the test
		t.Skip("PCI validation failed (lspci not available in test env)")
	}

	msg := cmd()
	if _, ok := msg.(PCIPassthroughUpdatedMsg); !ok {
		t.Errorf("Expected PCIPassthroughUpdatedMsg, got %T", msg)
	}

	// Verify saved config
	saved, err := form.vmManager.GetPCIPassthroughConfig()
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
	form := newTestPCIForm(t)

	form.selected["0000:01:00.0"] = true
	form.selected["0000:01:00.1"] = true
	form.focusIndex = len(form.positions) - 1

	_, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		// Validation may fail in test env (no lspci); skip the test
		t.Skip("PCI validation failed (lspci not available in test env)")
	}
	cmd()

	saved, err := form.vmManager.GetPCIPassthroughConfig()
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

	form := &PCIPassthroughFormModel{
		vmManager:  vmManager,
		devices:    []models.PCIDevice{},
		selected:   make(map[string]bool),
		errors:     make(map[string]string),
		iommuGroups: make(map[int][]*models.PCIDevice),
	}
	form.rebuildPositions()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 25})

	if len(form.positions) != 1 {
		t.Errorf("Expected 1 position (save button only), got %d", len(form.positions))
	}
	if form.positions[0].kind != pciSave {
		t.Error("Expected save button as only position")
	}
}

// TestPCIFORMViewShowsNoDevicesMessage verifies "no devices" message when scan finds nothing
func TestPCIFORMViewShowsNoDevicesMessage(t *testing.T) {
	vmManager := createTestVMManager(t)

	form := &PCIPassthroughFormModel{
		vmManager:  vmManager,
		devices:    []models.PCIDevice{},
		selected:   make(map[string]bool),
		errors:     make(map[string]string),
		iommuGroups: make(map[int][]*models.PCIDevice),
	}
	form.rebuildPositions()
	form.Update(tea.WindowSizeMsg{Width: 80, Height: 25})
	form.syncViewport()

	view := form.View()
	if !strings.Contains(view, "No PCI devices found") {
		t.Errorf("Expected 'No PCI devices found' message, got:\n%s", view)
	}
}

// TestPCIFORMToggleViaEnterKey verifies Enter key toggles device on/off
func TestPCIFORMToggleViaEnterKey(t *testing.T) {
	form := newTestPCIForm(t)

	// Find first toggle
	for i, pos := range form.positions {
		if pos.kind == pciToggle {
			form.focusIndex = i
			break
		}
	}

	addr := form.currentPos().deviceAddr

	// Enter to select
	form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !form.selected[addr] {
		t.Error("Device should be selected after Enter")
	}

	// Enter to deselect
	form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if form.selected[addr] {
		t.Error("Device should be deselected after second Enter")
	}
}

// TestPCIFORMToggleViaSpaceKey verifies Space key toggles device on/off
func TestPCIFORMToggleViaSpaceKey(t *testing.T) {
	form := newTestPCIForm(t)

	// Find first toggle
	for i, pos := range form.positions {
		if pos.kind == pciToggle {
			form.focusIndex = i
			break
		}
	}

	addr := form.currentPos().deviceAddr

	// Space to select
	form.Update(tea.KeyMsg{Type: tea.KeySpace})
	if !form.selected[addr] {
		t.Error("Device should be selected after Space")
	}
}

// TestPCIFORMNavigationTabCycle verifies Tab cycles through all positions
func TestPCIFORMNavigationTabCycle(t *testing.T) {
	form := newTestPCIForm(t)

	// Tab forward — should skip header and land on first toggle
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*PCIPassthroughFormModel)

	// Should be on a toggle position now
	if form.currentPos().kind != pciToggle {
		t.Errorf("After first Tab, expected to be on a toggle, got kind=%d", form.currentPos().kind)
	}

	// Tab through to the save button (last position)
	for i := 0; i < len(form.positions); i++ {
		model, _ = form.Update(tea.KeyMsg{Type: tea.KeyTab})
		form = model.(*PCIPassthroughFormModel)
	}

	if form.focusIndex != len(form.positions)-1 {
		t.Errorf("Expected focus at last position, got %d", form.focusIndex)
	}
	if form.currentPos().kind != pciSave {
		t.Errorf("Expected to land on save button, got kind=%d", form.currentPos().kind)
	}
}

// TestPCIFORMNavigationShiftTabBackward verifies Shift+Tab moves focus backward
func TestPCIFORMNavigationShiftTabBackward(t *testing.T) {
	form := newTestPCIForm(t)

	// Tab forward twice to get past initial header + first toggle + header to second toggle
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*PCIPassthroughFormModel)
	firstFocusable := form.focusIndex

	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyTab})
	form = model.(*PCIPassthroughFormModel)
	_ = form.focusIndex // second position reached

	// Shift+Tab should go back
	model, _ = form.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	form = model.(*PCIPassthroughFormModel)
	if form.focusIndex != firstFocusable {
		t.Errorf("After Shift+Tab, focusIndex = %d, want %d", form.focusIndex, firstFocusable)
	}
	if form.currentPos().kind != pciToggle {
		t.Errorf("After Shift+Tab, expected toggle position, got kind=%d", form.currentPos().kind)
	}
}

// TestPCIFORMNavigationUpArrow verifies Up arrow moves focus backward
func TestPCIFORMNavigationUpArrow(t *testing.T) {
	form := newTestPCIForm(t)

	// Find the second toggle position
	toggleIdx := 0
	found := 0
	for i, pos := range form.positions {
		if pos.kind == pciToggle {
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

	form.focusIndex = toggleIdx
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyUp})
	form = model.(*PCIPassthroughFormModel)

	// Should have moved back to the previous toggle
	if form.focusIndex >= toggleIdx {
		t.Errorf("After Up from index %d, focus should move back, got %d", toggleIdx, form.focusIndex)
	}
}

// TestPCIFORMNavigationDownArrow verifies Down arrow moves focus forward
func TestPCIFORMNavigationDownArrow(t *testing.T) {
	form := newTestPCIForm(t)

	// Find the first toggle position
	firstToggleIdx := -1
	for i, pos := range form.positions {
		if pos.kind == pciToggle {
			firstToggleIdx = i
			break
		}
	}
	if firstToggleIdx < 0 {
		t.Fatal("No toggle positions found")
	}

	form.focusIndex = firstToggleIdx
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyDown})
	form = model.(*PCIPassthroughFormModel)

	// Should have moved forward (skipping any header)
	if form.focusIndex <= firstToggleIdx {
		t.Errorf("After Down from index %d, focus should move forward, got %d", firstToggleIdx, form.focusIndex)
	}
	if form.currentPos().kind != pciToggle {
		t.Errorf("After Down, expected toggle position, got kind=%d", form.currentPos().kind)
	}
}

// TestPCIFORMNavigationBoundaryUp verifies Up at first position stays at 0
func TestPCIFORMNavigationBoundaryUp(t *testing.T) {
	form := newTestPCIForm(t)

	form.focusIndex = 0
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyUp})
	form = model.(*PCIPassthroughFormModel)

	if form.focusIndex != 0 {
		t.Errorf("Up at position 0 should stay at 0, got %d", form.focusIndex)
	}
}

// TestPCIFORMNavigationBoundaryDown verifies Down at last position stays at end
func TestPCIFORMNavigationBoundaryDown(t *testing.T) {
	form := newTestPCIForm(t)

	form.focusIndex = len(form.positions) - 1
	lastIdx := form.focusIndex
	model, _ := form.Update(tea.KeyMsg{Type: tea.KeyDown})
	form = model.(*PCIPassthroughFormModel)

	if form.focusIndex != lastIdx {
		t.Errorf("Down at last position %d should stay, got %d", lastIdx, form.focusIndex)
	}
}

// TestPCIFormIOMMUGroupHeaderCount verifies that rebuildPositions creates group headers
func TestPCIFormIOMMUGroupHeaderCount(t *testing.T) {
	form := newTestPCIForm(t)

	// Count group header positions
	headerCount := 0
	for _, pos := range form.positions {
		if pos.kind == pciGroupHeader {
			headerCount++
		}
	}

	if headerCount != 4 {
		t.Errorf("Expected 4 group header positions (groups 1,2,3,-1), got %d", headerCount)
	}
}

// TestPCIFormGroupHeaderNotNavigable verifies that when navigating via Tab,
// the focusIndex never points to a group header (moveFocus skips them).
func TestPCIFormGroupHeaderNotNavigable(t *testing.T) {
	form := newTestPCIForm(t)

	// Start past any initial group header
	form.focusIndex = 0
	if form.currentPos().kind == pciGroupHeader {
		form.moveFocus(1)
	}
	initialIdx := form.focusIndex
	visited := make(map[int]bool)

	for {
		visited[form.focusIndex] = true
		pos := form.currentPos()
		if pos.kind == pciGroupHeader {
			t.Errorf("Focus landed on group header at index %d (kind=%d, groupNum=%d)",
				form.focusIndex, pos.kind, pos.groupNum)
		}

		// Check if we've reached the save button (last position)
		if form.focusIndex >= len(form.positions)-1 {
			break
		}

		form.moveFocus(1)

		// Stop if we've looped back
		if form.focusIndex <= initialIdx && len(visited) > 1 {
			break
		}
	}

	// Verify we visited at least the toggle positions + save
	toggleCount := 0
	saveCount := 0
	for i, pos := range form.positions {
		if visited[i] {
			if pos.kind == pciToggle {
				toggleCount++
			} else if pos.kind == pciSave {
				saveCount++
			}
		}
	}

	if toggleCount != len(form.devices) {
		t.Errorf("Expected to visit all %d toggles, visited %d", len(form.devices), toggleCount)
	}
	if saveCount != 1 {
		t.Errorf("Expected to visit 1 save button, visited %d", saveCount)
	}
}

// TestPCIFORMSelectMultipleGroups verifies selecting devices from different groups works
func TestPCIFORMSelectMultipleGroups(t *testing.T) {
	form := newTestPCIForm(t)

	// Select group 1
	form.selected["0000:01:00.0"] = true
	form.selected["0000:01:00.1"] = true

	// Select group 3
	form.selected["0000:03:00.0"] = true

	// Navigate to save
	form.focusIndex = len(form.positions) - 1

	_, cmd := form.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		// Validation may fail in test env (no lspci); skip the test
		t.Skip("PCI validation failed (lspci not available in test env)")
	}
	cmd()

	saved, err := form.vmManager.GetPCIPassthroughConfig()
	if err != nil {
		t.Fatalf("Failed to load saved PCI config: %v", err)
	}

	if len(saved.Devices) != 3 {
		t.Errorf("Expected 3 saved devices, got %d", len(saved.Devices))
	}
}

// TestPCIFORMRenderGroupHeaderSelectionStatus verifies group header shows selection state
func TestPCIFORMRenderGroupHeaderSelectionStatus(t *testing.T) {
	form := newTestPCIForm(t)

	// Select all devices in group 1
	form.selected["0000:01:00.0"] = true
	form.selected["0000:01:00.1"] = true

	form.syncViewport()
	lines := form.renderAllLines()

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
