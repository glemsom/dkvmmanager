# Plan: Consolidate Form Message Handling ✅ COMPLETE

## Problem Statement

Form message handling has inconsistent documentation and interface naming:

1. **`ScrollableForm`** (framework layer in `form.go`): Has an internal `handleMessage` interface that delegates custom messages to form models. This works, but the interface is private.

2. **`message_handlers.go`** (main layer): Has a comment suggesting message routing needs manual handling, but the `ScrollableForm` already handles it via the `handleMessage` interface.

3. **Individual `*_form_handlers.go` files**: Some forms implement `HandleMessage`, some don't. There's no clear pattern for which forms should implement it.

### Current State Analysis

From `form.go`:
```go
// handleMessage is an optional interface for forms that need to handle
// custom messages (e.g., async command results).
type handleMessage interface {
    HandleMessage(msg tea.Msg) tea.Cmd
}
```

This interface already exists and works. The `ScrollableForm.Update` method delegates to it:
```go
default:
    if hm, ok := sf.model.(handleMessage); ok {
        if cmd := hm.HandleMessage(msg); cmd != nil {
            return sf, cmd
        }
    }
```

From `message_handlers.go`:
```go
// Note: FileSelectedMsg and DiskAddedMsg are handled by the VMFormModel's
// HandleMessage method, which is delegated to by ScrollableForm.Update.
// No manual routing needed here.
```

This comment is correct but buried in the middle of complex code. It suggests the pattern exists but isn't well documented.

### Friction Points

1. **Discovery problem**: The `handleMessage` interface is private (lowercase), making it invisible to callers
2. **Documentation problem**: The delegation pattern is mentioned in a comment but not as a designed seam
3. **Inconsistency**: Some form handlers have `HandleMessage`, others don't, with no guidance

---

## Solution Overview

Promote the `handleMessage` interface to a public `MessageHandler` interface in `types.go` and update the framework to use it. This creates a well-documented seam between the framework and domain forms.

---

## Implementation Steps

### ~~Step 1: Add Public MessageHandler Interface~~ ✅ DONE

**File**: `internal/tui/models/form/types.go`

Added the public `MessageHandler` interface.

### ~~Step 2: Update ScrollableForm to Use the Public Interface~~ ✅ DONE

**File**: `internal/tui/models/form/form.go`

Removed the private `handleMessage` interface and replaced the type assertion with the public `MessageHandler`.

### ~~Step 3: Update message_handlers.go Comment~~ ✅ DONE

Updated the comment in `handleSubViewMsg` to reference the public `MessageHandler` interface and point to `form/types.go`.

### ~~Step 4: Document HandleMessage Usage in Form Files~~ ✅ DONE

Added `form.MessageHandler` interface implementation comments to all six form files:
- `pci_passthrough_form_handlers.go`
- `vm_form.go`
- `vcpu_pinning_form_handlers.go`
- `lv_create_form.go`
- `cpu_options_form_handlers.go`
- `ssh_password_form.go`

---

## Files Changed Summary

| File | Change |
|------|--------|
| `internal/tui/models/form/types.go` | Add `MessageHandler` interface |
| `internal/tui/models/form/form.go` | Use `types.MessageHandler` instead of private `handleMessage` |
| `internal/tui/models/message_handlers.go` | Update comment to reference the interface |
| `internal/tui/models/pci_passthrough_form_handlers.go` | Add interface implementation comment |
| `internal/tui/models/vm_form.go` | Add interface implementation comment |
| `internal/tui/models/vcpu_pinning_form_handlers.go` | Add interface implementation comment |
| `internal/tui/models/lv_create_form.go` | Add interface implementation comment |
| `internal/tui/models/cpu_options_form_handlers.go` | Add interface implementation comment |
| `internal/tui/models/ssh_password_form.go` | Add interface implementation comment |

---

## Benefits

1. **Locality**: Message handling stays in form files (no change to logic)
2. **Discoverability**: The `MessageHandler` interface is publicly documented
3. **Testability**: Each form's message handling is already tested with the form
4. **Framework consistency**: Single documented seam for message routing

---

## Test Verification

After implementation:
- `make test` should pass unchanged (no logic changes)
- All existing form message handling works identically