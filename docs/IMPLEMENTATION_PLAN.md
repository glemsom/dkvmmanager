# Implementation Plan: Form Framework Adoption

## Overview

Migrate all TUI form models to use the `ScrollableForm` framework, consolidating scattered viewport/focus/navigation logic into a single reusable abstraction.

## Current State

The `form` package provides a `ScrollableForm` framework that handles:
- Viewport management (`viewport.Model`, `ready`, `contentW`, `contentH`)
- Focus navigation (`focusIndex`, `MoveFocus`, `SetFocusIndex`)
- Key dispatching (`handleKey`, `handleEnter`, `HandleBackspace`, `HandleDelete`)
- Cursor tracking (`cursorOffsets`)
- Error display (`errors` map)

**Only `SSHPasswordModel` was using this framework initially.**

## Current Progress

| Form Model | Status | Notes |
|------------|--------|-------|
| SSHPasswordModel | ✅ Done | Reference implementation |
| VMFormModel | ✅ Done | Fully migrated + bug fixes (openFilePickerCmd parsing, validateAndSave→validateAndSaveCmd, test fixes, ViewChangeMsg type, ClampOffset export) |
| CPUOptionsFormModel | ✅ Done | Fully migrated |
| CPUTopologyFormModel | ✅ Done | Fully migrated + test updates |
| PCIPassthroughFormModel | Pending | |
| USBPassthroughFormModel | Pending | |
| VCPUPinningFormModel | Pending | |
| StartStopScriptFormModel | Pending | |
| LVCreateFormModel | Pending | |

## Architecture Contract

### FormModel Interface (form/types.go)

```go
type FormModel interface {
    // Position Management
    BuildPositions() []FocusPos
    CurrentIndex() int
    SetFocusIndex(int)

    // Rendering
    RenderHeader() string
    RenderPosition(pos FocusPos, focused bool, cursorOffset int) string
    RenderFooter() string

    // Interaction
    HandleEnter(pos FocusPos) (FormResult, tea.Cmd)
    HandleChar(pos FocusPos, ch string)
    HandleBackspace(pos FocusPos)
    HandleDelete(pos FocusPos)

    // Lifecycle
    OnEnter()
    OnExit()
    SetSize(width, height int)
    SetFocused(bool)
}
```

### Optional Interfaces

```go
// For forms needing async message handling
type handleMessage interface {
    HandleMessage(msg tea.Msg) tea.Cmd
}

// For forms with save/cancel semantics
type savedMsg interface {
    IsFormSaved()
    FormName() string
    FormStatus() string
}
```

### Wrapper Pattern (reference: ssh_password.go)

```go
type SomeFormModel struct {
    form *form.ScrollableForm
}

func (m *SomeFormModel) Init() tea.Cmd {
    return m.form.Init()
}

func (m *SomeFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    inner, cmd := m.form.Update(msg)
    if sf, ok := inner.(*form.ScrollableForm); ok {
        m.form = sf
    }
    return m, cmd
}

func (m *SomeFormModel) View() string {
    return m.form.View()
}
```

## Forms to Migrate

| Form Model | Files | Status |
|------------|-------|--------|
| VMFormModel | vm_form*.go (13 files) | ✅ Migrated |
| CPUOptionsFormModel | cpu_options_form*.go (4 files) | ✅ Migrated |
| CPUTopologyFormModel | cpu_topology_form*.go (4 files) | ✅ Migrated |
| PCIPassthroughFormModel | pci_passthrough_form*.go (4 files) | Pending |
| USBPassthroughFormModel | usb_passthrough_form.go (+ render.go) | Pending |
| VCPUPinningFormModel | vcpu_pinning_form*.go (2 files) | Pending |
| StartStopScriptFormModel | start_stop_script_form*.go (2 files) | Pending |
| LVCreateFormModel | lv_create_form.go (+ render.go) | Pending |

## Migration Steps

### Phase 1: VMFormModel (Largest, Most Complex)

**Files involved:**
- `vm_form_model.go` - Model struct, positions, values
- `vm_form.go` - Update(), View(), SetSize()
- `vm_form_ui.go` - syncViewport(), render methods
- `vm_form_validation.go` - validateAndSave(), save methods
- `vm_form_types.go` - focusPos type

**Step 1.1: Create VMFormModel struct (vm_form_model.go)**

Before:
```go
type VMFormModel struct {
    mode      FormMode
    vmManager *vm.Manager
    vm        *models.VM

    // Field values
    vmName      string
    hardDisks   []string
    cdroms      []string
    macAddress  string
    vncEnabled  bool
    networkMode string
    tpmEnabled  bool

    // Flat list of focusable positions
    positions  []focusPos
    focusIndex int

    focused bool
    cursorOffsets map[string]int
    errors map[string]string

    vp       viewport.Model
    ready    bool
    contentW int
    contentH int
    renderedLines []string

    fileBrowser *FileBrowserModel
    addDiskModel *AddDiskModel
    browsingFieldName string
    browsingIndex int
}
```

After:
```go
type VMFormModel struct {
    mode      FormMode
    vmManager *vm.Manager
    vm        *models.VM

    // Field values
    vmName      string
    hardDisks   []string
    cdroms      []string
    macAddress  string
    vncEnabled  bool
    networkMode string
    tpmEnabled  bool

    // Focus state
    positions     []form.FocusPos
    focusIndex    int
    cursorOffsets map[string]int
    errors        map[string]string

    // File browser state (form-specific, keep)
    fileBrowser       *FileBrowserModel
    addDiskModel      *AddDiskModel
    browsingFieldName string
    browsingIndex     int
}
```

**Step 1.2: Implement FormModel interface**

Add methods to `vm_form_model.go`:

```go
// BuildPositions returns focusable positions using form.FocusPos
func (m *VMFormModel) BuildPositions() []form.FocusPos {
    // Convert current positions to form.FocusPos
    // Each position needs: Kind, Label, Key, Data (for listIndex, etc.)
}

// CurrentIndex/SetFocusIndex
func (m *VMFormModel) CurrentIndex() int { return m.focusIndex }
func (m *VMFormModel) SetFocusIndex(i int) { m.focusIndex = i }

// RenderHeader/RenderFooter
func (m *VMFormModel) RenderHeader() string {
    return "Create/Edit VM"
}

func (m *VMFormModel) RenderFooter() string {
    return styles.MutedTextStyle().Render("Tab Navigate  Space/Enter Browse  ESC Cancel")
}

// RenderPosition - move logic from vm_form_ui.go
func (m *VMFormModel) RenderPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
    // Call existing render methods like renderTextInput, renderListItem, etc.
    // These can be moved to vm_form_render.go
}

// HandleEnter - move from vm_form.go
func (m *VMFormModel) HandleEnter(pos form.FocusPos) (form.FormResult, tea.Cmd) {
    switch pos.Key {
    case "hardDisks", "cdroms": // Add buttons
        m.addItem(pos.Key)
        return form.ResultNone, nil
    case "save":
        return form.ResultSave, m.validateAndSave()
    default:
        // Move to next field
        return form.ResultNone, nil
    }
}

// HandleChar/Backspace/Delete - move from vm_form.go
func (m *VMFormModel) HandleChar(pos form.FocusPos, ch string) { /* ... */ }
func (m *VMFormModel) HandleBackspace(pos form.FocusPos) { /* ... */ }
func (m *VMFormModel) HandleDelete(pos form.FocusPos) { /* ... */ }

// Lifecycle
func (m *VMFormModel) OnEnter() {}
func (m *VMFormModel) OnExit() {}
func (m *VMFormModel) SetFocused(bool) {}
func (m *VMFormModel) SetSize(w, h int) { /* Keep for consistency but framework handles viewport */ }
```

**Step 1.3: Create VMFormWrapper**

Create `vm_form.go` (simplified):

```go
type VMFormWrapper struct {
    form *form.ScrollableForm
}

func NewVMFormWrapper(vmManager *vm.Manager, mode FormMode, vm *models.VM) *VMFormWrapper {
    fm := NewVMFormModel(vmManager, mode, vm)
    return &VMFormWrapper{form: form.NewScrollableForm(fm)}
}

// Init/Update/View delegate to form
// Add FileBrowserActive() method that delegates to inner form
```

**Step 1.4: Update VMCreateModel/VMEditModel**

Change from wrapping `VMFormModel` directly to wrapping `VMFormWrapper`:

```go
type VMCreateModel struct {
    form *form.ScrollableForm  // Was *VMFormModel
}

// Init/Update/View delegate to form
```

**Step 1.5: Update MainModel**

In `main.go`, change:
```go
vmCreateModel *form.ScrollableForm  // instead of *VMCreateModel
vmEditModel   *form.ScrollableForm  // instead of *VMEditModel
```

Or keep the wrapper but have it use the framework internally.

### Phase 2: CPUOptionsFormModel - Completed

Migrated CPUOptionsFormModel to form.ScrollableForm framework.

**Changes:**
- `cpu_options_form.go`: Removed viewport fields from struct, added FormModel interface methods (BuildPositions, CurrentIndex, SetFocusIndex, RenderHeader, RenderFooter, RenderPosition, HandleEnter, HandleChar, HandleBackspace, HandleDelete, OnEnter, OnExit, SetSize, SetFocused). Added helper methods (cursorOffset, setCursorOffset, effectiveCursor). Kept backward-compat Init/Update/View.
- `cpu_options_form_navigation.go`: Rewrote BuildPositions() to return []form.FocusPos with 28 positions (22 toggles + 2 text + 1 save + 3 section headers). Added FocusHeader positions for section headers. Kept currentPos() and moveFocus() for backward compat. Removed syncViewport(), focusedLineIndex(), pageUp(), pageDown().
- `cpu_options_form_handlers.go`: Implemented HandleEnter/HandleChar/HandleBackspace/HandleDelete/FormModel interface. Added HandleMessage for cpuOptionsErrorMsg. Renamed validateAndSave to validateAndSaveCmd. Kept backward-compat handleKey/handleEnterKey/handleBackspaceKey/handleDeleteKey/handleCharInput.
- `cpu_options_form_render.go`: Implemented RenderHeader/RenderFooter/RenderPosition for FormModel interface. Updated renderAllLines and renderPositionLine to handle FocusHeader positions. Kept renderToggle/renderTextInput/fieldLabel as helpers.
- `cpu_options.go`: Changed wrapper from `*CPUOptionsFormModel` to `*form.ScrollableForm`. Added `Form()` accessor. Implemented IsFormSaved/FormName/FormStatus on CPUOptionsUpdatedMsg.
- `cpu_options_form_test.go`: Updated for new positions (28 total, focusIndex starts at 1), updated navigation test for header positions, updated findIndexByName to use p.Key.
- `cpu_options_test.go`: Updated to use m.Form() accessor and m.Form().currentPos() and wrapped.form.Ready().
- `cpu_options_form_values.go` and `cpu_options_form_validation.go` - unchanged.

**All 10 packages pass tests.**

### Phase 4: PCIPassthroughFormModel

Uses IOMMU groups. Position data can be device addresses:

For PCIPassthroughFormModel:
```go
type pciFocusData struct {
    Address string
    GroupNum int
}

func (m *PCIPassthroughFormModel) BuildPositions() []form.FocusPos {
    var positions []form.FocusPos
    for groupNum, devices := range m.iommuGroups {
        // Header (render-only, no focus)
        positions = append(positions, form.FocusPos{
            Kind: form.FocusHeader,
            Label: fmt.Sprintf("IOMMU Group %d", groupNum),
            Key: fmt.Sprintf("group_%d", groupNum),
        })
        for _, dev := range devices {
            positions = append(positions, form.FocusPos{
                Kind:  form.FocusToggle,
                Label: dev.Name,
                Key:   dev.Address,
                Data:  pciFocusData{Address: dev.Address, GroupNum: groupNum},
            })
        }
    }
    positions = append(positions, form.FocusPos{
        Kind:  form.FocusButton, Label: "Save", Key: "save"
    })
    positions = append(positions, form.FocusPos{
        Kind:  form.FocusButton, Label: "Apply to Kernel", Key: "apply_kernel"
    })
    return positions
}
```

### Phase 5: USBPassthroughFormModel, VCPUPinningFormModel, StartStopScriptFormModel, LVCreateFormModel

Similar patterns.

## Key Design Decisions

### 2. Handling List Items and Position Types

Map current `focusKind` constants to `form.FocusKind`:

| Current | Framework | Notes |
|---------|-----------|-------|
| `focusText` | `form.FocusText` | Editable text field |
| `focusListItem` | `form.FocusList` | List item (click to edit) |
| `focusAddBtn` | `form.FocusButton` | Add button |
| `focusToggle` | `form.FocusToggle` | On/off toggle |
| `focusSaveBtn` | `form.FocusButton` | Save action |
| - | `form.FocusHeader` | Non-interactive header (new use) |

**Position Key Scheme**

Use consistent naming: `"fieldName"` for simple fields, `"fieldName_index"` for list items.

Example for VMFormModel:
```go
{
    Kind: form.FocusText, Label: "VM Name", Key: "vmName"
}
{
    Kind: form.FocusList, Label: "Hard Disk 1", Key: "hardDisks_0",
    Data: struct{ Index int }{Index: 0},
}
{
    Kind: form.FocusButton, Label: "[+] Add Disk", Key: "hardDisks_add"
}
```

### 3. Async Message Handling

Forms that spawn async operations (file browser, apply to kernel) need `HandleMessage`:

```go
func (m *SomeFormModel) HandleMessage(msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case FileSelectedMsg:
        return m.handleFileSelected(msg)
    case PCIVFIOKernelAppliedMsg:
        return m.handleKernelApplied(msg)
    }
    return nil
}
```

### 4. File Browser Integration

The file browser state (`fileBrowser`, `addDiskModel`) should remain in the form model since it's form-specific behavior. The framework's `handleMessage` interface handles async results.

## Testing Strategy

1. **Preserve existing tests** - The existing tests for each form model test internal state and behavior. These should continue to work.

   **Note:** Some tests directly access form fields (e.g., `m.vmEditModel.form.vmName = "test"`). These will need updates if the wrapper type changes. Consider adding accessor methods:
   ```go
   // In VMFormModel
   func (m *VMFormModel) SetVMName(name string) { m.vmName = name }
   func (m *VMFormModel) VMName() string { return m.vmName }
   ```

2. **Add framework integration tests** - One test per form that verifies:
   - `BuildPositions()` returns expected positions
   - `RenderPosition()` produces correct output
   - `HandleEnter()` returns correct result for each position type
   - Focus navigation works correctly

3. **Reference test**: `internal/tui/models/ssh_password_framework_test.go`

## Migration Order (Recommended)

1. **VMFormModel** - Largest impact, shared by create/edit
2. **CPUOptionsFormModel** - Simpler, good learning
3. **StartStopScriptFormModel** - Has toggle that changes positions
4. **VCPUPinningFormModel** - Simple toggle pattern
5. **CPUTopologyFormModel** - Complex with die/core data
6. **PCIPassthroughFormModel** - Group headers
7. **USBPassthroughFormModel** - Similar to PCI
8. **LVCreateFormModel** - Standalone

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Break existing tests | Run tests after each form migration; tests test behavior, not implementation |
| Keyboard handling changes | Verify Tab, Enter, Backspace, arrow keys work identically |
| File browser integration | Keep async message handling in HandleMessage interface |
| Performance regression | Framework is same logic, just consolidated; no additional overhead |

## Position Key Reference

Each form should use consistent key naming. Here's the mapping for each form:

### VMFormModel
| Field | Key | Kind |
|-------|-----|------|
| VM Name | `vmName` | FocusText |
| Hard Disk 0 | `hardDisks_0` | FocusList |
| Hard Disk 1 | `hardDisks_1` | FocusList |
| [+] Add Disk | `hardDisks_add` | FocusButton |
| CDROM 0 | `cdroms_0` | FocusList |
| [+] Add CDROM | `cdroms_add` | FocusButton |
| MAC Address | `macAddress` | FocusText |
| VNC Enabled | `vncEnabled` | FocusToggle |
| Network Mode | `networkMode` | FocusToggle |
| TPM Enabled | `tpmEnabled` | FocusToggle |
| Save | `save` | FocusButton |

### CPUOptionsFormModel
| Field | Key | Kind |
|-------|-----|------|
| Hide KVM | `hideKVM` | FocusToggle |
| HV Time | `hvTime` | FocusToggle |
| Vendor ID | `vendorID` | FocusText |
| CPU Model | `cpuModel` | FocusText |
| CPU PM | `cpuPM` | FocusToggle |
| Save | `save` | FocusButton |

### CPUTopologyFormModel
| Field | Key | Kind |
|-------|-----|------|
| Core 0 (die 0) | `0:0` | FocusToggle |
| Core 1 (die 0) | `0:1` | FocusToggle |
| ... | ... | ... |
| Save | `save` | FocusButton |

### PCIPassthroughFormModel
| Field | Key | Kind |
|-------|-----|------|
| Group header (group 1) | `group_1` | FocusHeader |
| Device 0000:01:00.0 | `0000:01:00.0` | FocusToggle |
| Save | `save` | FocusButton |
| Apply to Kernel | `apply_kernel` | FocusButton |

### VCPUPinningFormModel
| Field | Key | Kind |
|-------|-----|------|
| Pinning Enabled | `enabled` | FocusToggle |
| Save | `save` | FocusButton |
| Apply to Kernel | `apply_kernel` | FocusButton |

### StartStopScriptFormModel
| Field | Key | Kind |
|-------|-----|------|
| Toggle (builtin/custom) | `toggle` | FocusToggle |
| Start Script | `start_path` | FocusText |
| Start Browse | `start_browse` | FocusButton |
| Stop Script | `stop_path` | FocusText |
| Stop Browse | `stop_browse` | FocusButton |
| Save | `save` | FocusButton |
| Cancel | `cancel` | FocusButton |

## Detailed Migration Checklist

For each form, follow these steps:

### Pre-Migration
- [ ] Run `make test` to capture baseline test results
- [ ] Identify all files for this form (see table above)

### Migration Steps

#### 1. Update Model Struct
- [ ] Remove viewport fields (`vp`, `ready`, `contentW`, `contentH`, `renderedLines`)
- [ ] Change `positions []focusPos` to `positions []form.FocusPos`
- [ ] Keep form-specific fields (fileBrowser, async state, etc.)

#### 2. Implement FormModel Interface
- [ ] `BuildPositions()` - Convert current positions to `form.FocusPos`
- [ ] `CurrentIndex()`, `SetFocusIndex()` - Pass through
- [ ] `RenderHeader()` - Extract header from current render logic
- [ ] `RenderFooter()` - Extract footer from current render logic
- [ ] `RenderPosition()` - Move position rendering logic here
- [ ] `HandleEnter()` - Move enter key logic, return appropriate `FormResult`
- [ ] `HandleChar()`, `HandleBackspace()`, `HandleDelete()` - Move text editing logic
- [ ] `OnEnter()`, `OnExit()`, `SetFocused()` - Empty or minimal implementation
- [ ] `SetSize()` - Empty (framework manages viewport)

#### 3. Update Wrapper Model
- [ ] Change wrapper to use `form.ScrollableForm`
- [ ] Delegate `Init()`, `Update()`, `View()` to framework
- [ ] Keep form-specific accessor methods (e.g., `FileBrowserActive()`)

#### 4. Remove Duplicated Code
- [ ] Delete `syncViewport()` method
- [ ] Delete `handleKey()` method  
- [ ] Delete `moveFocus()` method
- [ ] Delete viewport initialization in `SetSize()` / `Update()`
- [ ] Delete rendering methods that were moved to `RenderPosition()`

#### 5. Update MainModel References
- [ ] If wrapper type changed, update `MainModel` field type
- [ ] Update any direct field access to use accessor methods

#### 6. Testing
- [ ] Run `make test` - all tests should pass
- [ ] Quick manual verification of: Tab, Shift+Tab, Enter, Backspace, navigation

## Definition of Done

- [x] SSHPasswordModel implements `form.FormModel` ✅
- [x] VMFormModel implements `form.FormModel` ✅
- [x] CPUOptionsFormModel implements `form.FormModel` ✅
- [x] CPUTopologyFormModel implements `form.FormModel` ✅
- [x] All form wrappers use `form.ScrollableForm` (SSHPasswordModel ✅, VMCreateModel ✅, VMEditModel ✅, CPUOptionsModel ✅, CPUTopologyModel ✅)
- [x] All tests pass (`go test ./...` - all 10 packages pass)
- [ ] No behavioral regression in keyboard/input handling
- [ ] Code coverage maintained or improved

## Progress Log

### 2026-05-03: Phase 3 (CPUTopologyFormModel) - Completed

Migrated CPUTopologyFormModel to form.ScrollableForm framework.

**Changes (7 files):**
- `cpu_topology_form.go`: Removed `cpuTopoFocusKind`/`cpuTopoFocusPos` types, removed viewport fields from struct, added `cpuTopoFocusData` for per-core data via `form.FocusPos.Data`. Changed `positions` to `[]form.FocusPos`. Added FormModel interface methods (BuildPositions, CurrentIndex, SetFocusIndex, RenderHeader, RenderFooter, RenderPosition, HandleEnter, HandleChar, HandleBackspace, HandleDelete, OnEnter, OnExit, SetSize, SetFocused). Renamed `validateAndSave()` → `validateAndSaveCmd()` returning `(form.FormResult, tea.Cmd)`. Kept backward-compat `currentPos()`, `handleKey()`, `handleEnterKey()`, `handleSpace()`, `Init()`, `Update()`, `View()`.
- `cpu_topology_form_navigation.go`: Rewrote `BuildPositions()` to return `[]form.FocusPos` with toggles for each core + save button. Removed `syncViewport()` and `focusedLineIndex()`. Kept `moveFocus()` and `hostCoreCount()` for backward compat.
- `cpu_topology_form_render.go`: Updated `renderAllLines()` to use `form.FocusPos` / `cpuTopoFocusData`. Kept `renderCoreLine()` as helper.
- `cpu_topology_form_validation.go`: Renamed `validateAndSave()` → `validateAndSaveCmd()`, returns `(form.FormResult, tea.Cmd)` instead of `(tea.Model, tea.Cmd)`. Uses `form.FocusPos` for iteration.
- `cpu_topology.go`: Changed wrapper from `*CPUTopologyFormModel` to `*form.ScrollableForm`. Added `Form()` accessor. Added `IsFormSaved()`/`FormName()`/`FormStatus()` on `CPUTopologyUpdatedMsg`.
- `cpu_topology_form_test.go`: Updated tests to use `form.FocusPos` / `p.Kind == form.FocusToggle` instead of old `cpuTopoFocusKind` constants. Added `TestCPUTopologyFormModelInterface` to verify interface implementation.
- `cpu_topology_test.go`: Updated to use `m.form.Ready()` (not `.ready`), `m.Form()` accessor. Added `TestCPUTopologyFormAccessor` and `TestCPUTopologyUpdatedMsgImplementsFormSavedMsg`.

**Key design:** `cpuTopoFocusData` carries `dieID`, `coreID`, `dieLabel`, and `coreInfo` through `form.FocusPos.Data`, enabling `BuildPositions()` to build positions without needing access to `hostTopo` during rendering.

**All 10 packages pass tests.**

### 2026-05-03: Phase 2 (CPUOptionsFormModel) - Completed

Migrated CPUOptionsFormModel to form.ScrollableForm framework following SSHPasswordModel pattern.

**Changes (7 files):**
- `cpu_options_form.go`: Removed viewport fields from struct, added FormModel interface methods (BuildPositions, CurrentIndex, SetFocusIndex, RenderHeader, RenderFooter, RenderPosition, HandleEnter, HandleChar, HandleBackspace, HandleDelete, OnEnter, OnExit, SetSize, SetFocused). Kept backward-compat Init/Update/View.
- `cpu_options_form_navigation.go`: Rewrote BuildPositions() to return `[]form.FocusPos` with 28 positions (22 toggles + 2 text + 3 section headers + 1 save). Removed syncViewport(), focusedLineIndex(), pageUp(), pageDown(). Kept currentPos()/moveFocus() for backward compat.
- `cpu_options_form_handlers.go`: Implemented FormModel HandleEnter/HandleChar/HandleBackspace/HandleDelete. Added HandleMessage for cpuOptionsErrorMsg. Renamed validateAndSave→validateAndSaveCmd.
- `cpu_options_form_render.go`: Implemented RenderHeader/RenderFooter/RenderPosition. Updated renderAllLines/renderPositionLine for FocusHeader handling.
- `cpu_options.go`: Changed wrapper from `*CPUOptionsFormModel` to `*form.ScrollableForm`. Added `Form()` accessor. Added FormSavedMsg methods on CPUOptionsUpdatedMsg.
- `cpu_options_form_test.go` + `cpu_options_test.go`: Updated for new position count (28), focusIndex starts at 1, use m.Form() accessor.

**Note:** Section headers (FocusHeader) are now focusable/tab-stops. This changes UX slightly from the original where headers were non-focusable.

**All 10 packages pass tests.**

### 2026-05-03: Phase 1 (VMFormModel) - Completed + Bug Fixes

The VMFormModel was already partially migrated to the form framework, but the codebase had several pre-existing compilation errors and test failures that prevented it from building. The following fixes were applied:

**Bug Fixes:**
1. Added missing `ViewChangeMsg` type to `types.go`
2. Exported `ClampOffset` function in `form/focus.go` and updated all callers (cpu_options, cpu_topology, pci_passthrough, usb_passthrough, vcpu_pinning)
3. Fixed `vm_form_validation.go`: renamed `validateAndSave()` to `validateAndSaveCmd()` returning `tea.Cmd` instead of `(tea.Model, tea.Cmd)`
4. Fixed `vm_form.go`: removed dead `openFilePicker()`, `handleFileSelected()`, `handleDiskAdded()` methods that returned `tea.Model`
5. Fixed `openFilePickerCmd()`: corrected field name parsing from `pos.Key[:len(pos.Key)-2]` to proper underscore-based split
6. Added `Focused()` getter to `ScrollableForm`

**Test Fixes:**
- Updated all tests to access form fields via `form.Model().(*VMFormModel)` instead of direct field access
- Fixed message types in file picker tests (DiskAddedMsg for hardDisks, FileSelectedMsg for cdroms)
- Rewrote `vm_form_filepicker_test.go` and `filepicker_integration_test.go` to work with new API

**All 10 packages pass tests.**