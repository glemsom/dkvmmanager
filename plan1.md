# Form Framework Consolidation Plan

## Overview

Consolidate the ~7 form implementations (VMForm, CPUOptionsForm, CPUTopologyForm, PCIPassthroughForm, USBPassthroughForm, VCPUPinningForm, SSHPasswordForm, StartStopScriptForm, LVCreateForm) into a single framework that handles common concerns while allowing forms to customize rendering and behavior.

## Goals

1. **Eliminate boilerplate** - Remove ~500 lines of duplicate focus management, viewport sync, and key handling from each form
2. **Improve testability** - Framework handles navigation, forms only test domain logic
3. **Enable faster feature development** - New forms require only field definitions and validation, not full framework code
4. **Centralize bug fixes** - Fix focus navigation once, all forms benefit

## Architecture

### Core Framework (`internal/tui/models/form/`)

```
form/
├── form.go          # Core FormModel interface and ScrollableForm implementation
├── focus.go         # FocusPos, FocusKind, focus navigation logic
├── viewport.go      # Viewport synchronization and line indexing
├── keybinds.go      # Key binding definitions and default handlers
└── messages.go      # Standard message types
```

### FormModel Interface

Forms implement this interface to integrate with the framework:

```go
// FormModel defines the contract for forms using the scrollable form framework
type FormModel interface {
    // Position Management
    BuildPositions() []FocusPos
    GetCurrentIndex() int
    SetFocusIndex(int)
    
    // Rendering
    RenderHeader() string
    RenderPosition(pos FocusPos, focused bool) string
    RenderFooter() string
    
    // Interaction
    HandleEnter(pos FocusPos) (FormResult, tea.Cmd)
    HandleChar(pos FocusPos, ch string)
    HandleBackspace(pos FocusPos)
    HandleDelete(pos FocusPos)
    
    // Lifecycle
    OnEnter()          // Called when form becomes active
    OnExit()           // Called when form is dismissed
    SetSize(width, height int)
    SetFocused(bool)
}
```

### Focus Position System

```go
// FocusKind defines the type of focusable element
type FocusKind int

const (
    FocusText     FocusKind = iota  // Editable text field
    FocusToggle                      // Boolean toggle
    FocusList                          // Selectable list item
    FocusButton                        // Action button (save, add, etc.)
    FocusHeader                        // Non-interactive header (display only)
    FocusCustom                        // Custom rendering handled by form
)

// FocusPos represents one navigable position in a form
type FocusPos struct {
    Kind   FocusKind
    Label  string     // Human-readable label
    Key    string     // Unique identifier for cursor/error tracking
    Data   any        // Form-specific data (device info, field name, etc.)
}

// FormResult indicates what action the form wants to take
type FormResult int
const (
    ResultNone     FormResult = iota  // No special action
    ResultSave                          // Save and exit
    ResultCancel                        // Cancel and exit
    ResultCustom                        // Custom action (handled by cmd)
)
```

### Key Bindings

```go
var DefaultKeyBindings = KeyBindings{
    // Navigation
    Tab:          []string{"tab"},
    ShiftTab:     []string{"shift+tab"},
    Up:           []string{"up"},
    Down:         []string{"down"},
    PageUp:       []string{"pgup"},        // Optional - forms can opt out
    PageDown:     []string{"pgdown"},      // Optional - forms can opt out
    
    // Actions
    Enter:        []string{"enter"},
    Space:        []string{" "},
    Backspace:    []string{"backspace"},
    Delete:       []string{"delete"},
    CharInput:    "default", // Any single character
}
```

## Implementation Steps

### Phase 1: Framework Foundation

1. ✅ Create `form/form.go` with `ScrollableForm` struct
   - Own the viewport model
   - Implement `Update()` dispatching to form methods
   - Implement `View()` delegating to form's render methods
   - Handle `tea.WindowSizeMsg` (single set, no resize)
   - **Done**: `ScrollableForm` struct with `tea.Model` implementation, viewport ownership, key dispatching, focus navigation, cursor tracking

2. ✅ Create `form/focus.go`
   - `moveFocus(delta int)` with clamping (no wrap)
   - `focusedLineIndex()` calculation
   - `clampOffset()` for viewport positioning
   - **Done**: 9 focus navigation tests + viewport utility tests, all passing

3. ✅ Create `form/keybinds.go`
   - Define standard key bindings
   - Map keys to form method calls
   - **Done**: `KeyBindings` struct with default bindings and matcher helpers

4. ✅ Create `form/types.go` (shared types)
   - `FocusKind`, `FocusPos`, `FormResult` types
   - `FormModel` interface
   - **Done**: Core types and interface defined

**Tests**: 33 tests in `form/` package, all passing. `go vet` clean on full project.

### Phase 2: Message Types

4. ✅ Create `form/messages.go`
   - `FormSavedMsg` interface — marker that any form save result implements
   - `IsFormSavedMsg(msg) bool` — detect if any message is a form save result
   - `FormSavedName(msg) string` — extract form name for status messages
   - `FormSavedStatus(msg) string` — extract status (with default fallback)
   - **Done**: 4 tests passing, `go vet` clean. Forms define their own `*UpdatedMsg` types and implement `FormSavedMsg` to participate in uniform handling.

### Phase 3: Form Migration

5. ✅ Migrate SSH Password form to use framework
   - `SSHPasswordFormModel` implements `form.FormModel` interface
   - Wrapped in `form.ScrollableForm` via `SSHPasswordModel`
   - `SSHPasswordUpdatedMsg` implements `form.FormSavedMsg`
   - Added `handleMessage` interface to framework for custom message delegation
   - Fixed framework bug: `handleEnter` now preserves cmd from `HandleEnter`
   - **Tests**: 11 integration tests passing (`ssh_password_framework_test.go`)
   - Old tests backed up (`ssh_password_test.go.bak`) — need updating in Phase 4
   - Removed `ssh_password_form_handlers.go` and `ssh_password_form_navigation.go` (framework handles these)
   - Simplified `ssh_password_form_render.go` (only `renderPasswordInput` remains)

6. Migrate remaining forms one by one
   - Update imports from `vm.*` to use new pattern
   - Remove duplicate code

### Phase 4: Cleanup

7. Delete old form implementation files
8. Update `main.go` message handling if needed
9. Update tests to work with new framework

## Key Design Decisions

### Error Handling
- Forms maintain their own error map
- Framework provides error display via `RenderPosition`
- Errors keyed by field identifier for consistent highlighting

### Cursor Management
- Framework tracks cursor offset per field key
- Cursor position passed to `RenderPosition` for text fields
- Default: text cursor at end; forms can override for special behavior

### Size Management
- Forms sized once when they become active
- `SetSize(width, height int)` called by parent
- No resize handling needed (console app, no dynamic resize)

### Key Customization
- Forms can define custom key handlers if needed
- Base key bindings sufficient for most forms
- Escape key handled by parent (MainModel), not forms

### Message Types
- Each form defines its own result message type
- Type-safe and explicit about what data is returned
- Parent handles uniformly via type switch

## File Changes Summary

| Action | Files |
|--------|-------|
| Create | `internal/tui/models/form/*.go` (5 files) |
| Modify | `internal/tui/models/vm_form_*.go` → single file or minimal wrapper |
| Modify | `internal/tui/models/cpu_options_form*.go` → use framework |
| Modify | `internal/tui/models/cpu_topology_form*.go` → use framework |
| Modify | `internal/tui/models/pci_passthrough*.go` → use framework |
| Modify | `internal/tui/models/usb_passthrough*.go` → use framework |
| Modify | `internal/tui/models/vcpu_pinning*.go` → use framework |
| Modify | `internal/tui/models/ssh_password*.go` → use framework |
| Modify | `internal/tui/models/start_stop_script*.go` → use framework |
| Modify | `internal/tui/models/lv_create*.go` → use framework |
| Delete | Old form infrastructure code after migration |

## Testing Strategy

1. **Framework tests** - Test focus navigation, viewport sync, key handling once
2. **Form tests** - Continue existing tests, they should work unchanged
3. **Integration tests** - Verify message flow from forms to main model

## Rollback Plan

If issues arise, the framework files can be deleted and old form files restored from git history. The framework is additive change.