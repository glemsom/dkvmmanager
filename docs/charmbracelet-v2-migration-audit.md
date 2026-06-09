# Charmbracelet v2 Migration Audit

> Issue #42 — Full audit of all v1 Charmbracelet API usage across the DKVM Manager
> codebase, with a precise migration checklist.

**Audit date:** 2026-06-09
**Auditor:** Agent (pi)

---

## Summary

| Library | v1 Path | v2 Path | Files Affected |
|---------|---------|---------|----------------|
| Bubble Tea | `github.com/charmbracelet/bubbletea` | `charm.land/bubbletea/v2` | 75 source files, ~17 test files |
| Bubbles | `github.com/charmbracelet/bubbles/...` | `charm.land/bubbles/v2/...` | 17 files (list, table, viewport) |
| Lip Gloss | `github.com/charmbracelet/lipgloss` | `charm.land/lipgloss/v2` | 33 source files (~396 call sites) |

**Total Go files scanned:** 175 (102 source + 72 test + 1 main.go)

---

## 1. Import Paths

### Need updating

| Pattern | Source Files | Test Files | Total |
|---------|-------------|------------|-------|
| `github.com/charmbracelet/bubbletea` | 50 imports | ~17 files | 67+ |
| `github.com/charmbracelet/bubbles/list` | 4 files | 1 file | 5 |
| `github.com/charmbracelet/bubbles/table` | 1 file | 0 | 1 |
| `github.com/charmbracelet/bubbles/viewport` | 11 files | 2 files | 13 |
| `github.com/charmbracelet/lipgloss` | 28 imports | 4 files | 32+ |

**Target:** `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`

---

## 2. `View() string` → `View() tea.View`

### 32 methods need updating (source files, excluding tests)

These are in:

- `internal/tui/components/`: `vm_cards.go`, `vm_details.go`, `vm_list_view.go`, `vm_table.go`
- `internal/tui/models/`: `cpu_topology.go`, `cpu_topology_form.go`, `cpu_options.go`, `cpu_options_form.go`, `main.go`, `vm_delete.go`, `disk_selector.go` (×2), `file_browser.go`, `lvm_volume.go`, `vm_running.go`, `ssh_password.go`, `ssh_password_form.go`, `start_stop_script.go`, `lv_create.go`, `lv_create_form.go`, `pci_passthrough.go`, `pci_passthrough_form.go`, `start_stop_script_form.go`, `usb_passthrough.go`, `usb_passthrough_form.go`, `vcpu_pinning.go`, `vcpu_pinning_form.go`, `mount_point_warning.go`, `vm_edit.go`, `vm_create.go`, `form/form.go`
- `internal/tui/models/vm_selection.go`: `renderVMSelectView()` (returns string, used as helper, not tea.Model interface)

**Pattern:** All `func (m *Model) View() string` → `func (m *Model) View() tea.View` returning `tea.NewView(string)`.

Note: `renderVMSelectView()` is a helper returning `string`, not a `tea.Model` interface method, so it may not need to change return type — but its callers in `View()` will wrap it in `tea.NewView()`.

---

## 3. `tea.KeyMsg` Usage

### 31 key handling functions across source files

| File | Function |
|------|----------|
| `components/vm_cards.go` | `Update()` switch |
| `models/cpu_topology_form.go` | `Update()` switch, `handleKey()` |
| `models/disk_selector_handlers.go` | `BlockDeviceModel.handleKeyPress()`, `AddDiskModel.handleKeyPress()` |
| `models/vm_delete.go` | `handleKeyPress()` |
| `models/file_browser.go` | `handleKeyPress()` |
| `models/lvm_volume.go` | `handleKeyPress()` |
| `models/vm_running.go` | `handleKeyPress()` |
| `models/form/form.go` | `handleKey()` |
| `models/ssh_password_form.go` | `handleKey()` |
| `models/cpu_options_form_handlers.go` | `handleKey()` |
| `models/pci_passthrough_form.go` | `handleKey()` |
| `models/start_stop_script_form.go` | `Update()` switch |
| `models/usb_passthrough_form.go` | `handleKey()` |
| `models/vcpu_pinning_form.go` | `handleKey()` |
| `models/mount_point_warning.go` | `handleKeyPress()` |
| `models/key_handlers.go` | 2x type assertions + `handleKeyPress()` |
| ... and more | |

**Patterns used:**

- **String matching** (most common): `msg.String()` / `key.String()` / `km.String()` — e.g., `case "esc":`, `case "q":`, etc.
- **Type switch on `tea.KeyMsg`**: `case tea.KeyMsg:` in Update() functions
- **Type assertion style**: `if km, ok := msg.(tea.KeyMsg); ok && km.String() == "esc"` (in `key_handlers.go`)

### 224 `tea.KeyMsg{...}` struct literals in test files

All need to change to `tea.KeyPressMsg{...}` or the equivalent new struct.

### Key type constants used in tests

Common ones: `tea.KeyUp`, `tea.KeyDown`, `tea.KeyEnter`, `tea.KeyEsc`, `tea.KeyCtrlC`, `tea.KeyTab`, `tea.KeySpace`, `tea.KeyRunes`, `tea.KeyBackspace`

**Migration:** 
- `case tea.KeyMsg:` → `case tea.KeyPressMsg:` 
- `msg.Type` → `msg.Code`
- `msg.Runes` → `msg.Text` (now `string`, not `[]rune`)
- `msg.Alt` → `msg.Mod` with `msg.Mod.Contains(tea.ModAlt)`
- `tea.KeyCtrlC` → `msg.String() == "ctrl+c"` or check `msg.Code` + `msg.Mod`
- Space bar: `case " ":` → `case "space":`
- `tea.KeyRunes` no longer exists — check `len(msg.Text) > 0` instead

---

## 4. `tea.MouseMsg` Usage

### 8 occurrences across 8 form files

| File | Line |
|------|------|
| `cpu_topology_form.go` | `case tea.MouseMsg:` |
| `form/form.go` | `case tea.MouseMsg:` |
| `ssh_password_form.go` | `case tea.MouseMsg:` |
| `cpu_options_form.go` | `case tea.MouseMsg:` |
| `pci_passthrough_form.go` | `case tea.MouseMsg:` |
| `start_stop_script_form.go` | `case tea.MouseMsg:` |
| `usb_passthrough_form.go` | `case tea.MouseMsg:` |
| `vcpu_pinning_form.go` | `case tea.MouseMsg:` |

**Migration:** Rename mouse button constants per v2 guide (detailed changes to be confirmed).

---

## 5. Bubbles Component API Patterns

### list (5 files)

- `internal/tui/models/init.go` — `list.New(...)` (3 calls for menu, config, power lists)
- `internal/tui/models/vm_selection.go` — `list.New(...)` (2 calls)
- `internal/tui/models/list_adapter.go` — `list.Model`, `list.Item`, custom delegates
- `internal/tui/models/list_adapter_test.go` — `list.New(...)` (4 calls)
- `internal/tui/models/key_handlers.go` — `list` package imported

**Migration notes:**
- `list.New()` signature unchanged
- `DefaultStyles()` → `DefaultStyles(isDark bool)` — need a `isDark` parameter
- `NewDefaultItemStyles()` → `NewDefaultItemStyles(isDark bool)`
- Filter styles restructured: `styles.FilterPrompt` → `styles.Filter.Focused.Prompt` / `styles.Filter.Blurred.Prompt`

### table (1 source file)

- `internal/tui/components/vm_table.go` — full usage pattern

**Migration notes:**
- `table.New(...)` signature unchanged
- `table.DefaultStyles()` → `table.DefaultStyles(isDark bool)`
- `.SetWidth(width)` stays same (already uses setter)
- `.SetHeight(height)` stays same
- `.SetStyles(s)` stays same
- `.SetRows(rows)` stays same
- `.UpdateViewport(...)` — check for new API
- `.Width =` → `.SetWidth(width)` — already using setter, good

### viewport (13 files)

- Used in form framework + individual form models

**Migration notes:**
- `viewport.New(w, h)` signature unchanged
- `.Width = w` → `.SetWidth(w)`
- `.Height = h` → `.SetHeight(h)`
- Getter methods: `.Width()` and `.Height()` now available

### Files with direct field assignment on viewport:

| File | Line | v1 Pattern | v2 Pattern |
|------|------|-----------|------------|
| `cpu_topology_form.go` | `m.vp.Width = w` | field | `.SetWidth(w)` |
| `vm_running.go` | `m.vp.Width = m.width` | field | `.SetWidth(m.width)` |
| `form/form.go` | `sf.vp.Width = w` | field | `.SetWidth(w)` |
| `ssh_password_form.go` | `m.vp.Width = w` | field | `.SetWidth(w)` |
| `cpu_options_form.go` | `m.vp.Width = w` | field | `.SetWidth(w)` |
| `pci_passthrough_form.go` | `m.vp.Width = w` | field | `.SetWidth(w)` |
| `start_stop_script_form.go` | `m.vp.Width = w` | field | `.SetWidth(w)` |
| `usb_passthrough_form.go` | `m.vp.Width = w` | field | `.SetWidth(w)` |
| `vcpu_pinning_form.go` | `m.vp.Width = w` | field | `.SetWidth(w)` |
| Same files for `.Height = h` | — | field | `.SetHeight(h)` |

Note: `cpu_topology_form.go` also does `m.vp.Width = w` and `m.vp.Height = h` — 2 field assignments.

---

## 6. Lip Gloss Patterns

### ~396 lipgloss call sites in source (not test) files

### `lipgloss.Color` as a type (8 variable/struct/return-type declarations)

Files needing `lipgloss.Color` → `color.Color`:

- `internal/tui/theme/theme.go` — `Theme` struct fields (20 fields of type `lipgloss.Color`)
- `internal/tui/styles/colors.go` — `StatusColors` struct (3 fields), `var color lipgloss.Color` in `StatusIndicator()`
- `internal/tui/components/statusbar.go` — `var color lipgloss.Color`
- `internal/tui/components/vm_cards.go` — `var borderColor lipgloss.Color`
- `internal/tui/models/ssh_password_form_validation.go` — `func strengthLabel(...) (string, lipgloss.Color)`
- `internal/tui/styles/styles.go` — `lipgloss.Color(fmt.Sprintf("%d", colorIdx))` (function call, not type)

### `lipgloss.Color("4")` style calls (38 call sites)

In v2, `lipgloss.Color` becomes a function returning `color.Color`, but the call syntax `lipgloss.Color("4")` stays the same. Only variable/field/return-type declarations need type changes.

### `lipgloss.Style` usage (~300+ call sites)

- `lipgloss.NewStyle()` — unchanged in v2
- `.Foreground(c)` — parameter type changes from `TerminalColor` to `color.Color`
- `.Background(c)` — same parameter type change
- `.BorderForeground(c)` — same parameter type change
- `.BorderBackground(c)` — same parameter type change
- Style getter methods in tests (60 occurrences):
  - `.GetBold()`, `.GetForeground()`, `.GetBackground()`, `.GetItalic()`, `.GetUnderline()`, `.GetStrikethrough()`, `.GetPaddingLeft()`, `.GetPaddingTop()`
  - These may change in v2 — need to check

### `lipgloss.JoinHorizontal()` / `lipgloss.JoinVertical()` — used in `dualpane.go`

### `lipgloss.Place()` — used in `mount_point_warning.go`

### `lipgloss.Width()` — widely used (~20+ call sites)

### `lipgloss.RoundedBorder()` / `lipgloss.NormalBorder()` — used in styles

### `string(tt.color)` conversion in test file

`internal/tui/styles/colors_test.go` does `string(tt.color)` which assumes `lipgloss.Color` is a string type. In v2 this won't work — need alternative assertion.

---

## 7. `tea.WindowSizeMsg` Struct Literals in Tests

### 40 struct literals in test files

Most common pattern:
```go
form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
```

The struct `tea.WindowSizeMsg` itself still exists in v2 but since `Width` and `Height` become getter/setter methods on components, direct struct literals should still compile. **No change needed** for struct literal syntax.

---

## 8. Program Setup (`tui.go`)

| v1 | v2 Change |
|----|-----------|
| `tea.WithAltScreen()` | Remove option → set `v.AltScreen = true` in `View()` |
| `tea.WithOutput(os.Stderr)` | Remove option → set `v.Output = os.Stderr` in `View()`? Check v2 API |
| `tea.NewProgram(m, opts...)` | Probably same |
| `p.Run()` | Same |

**`tea.Quit`** — 2 occurrences in `key_handlers.go` — likely still exists in v2.

**`tea.Batch`** — 4 occurrences — likely still exists in v2.

---

## 9. Complete Migration Checklist

### Phase 1: Update Import Paths (Issue #43)
- [ ] Replace `github.com/charmbracelet/bubbletea` → `charm.land/bubbletea/v2` (67+ files)
- [ ] Replace `github.com/charmbracelet/bubbles/...` → `charm.land/bubbles/v2/...` (~17 files)
- [ ] Replace `github.com/charmbracelet/lipgloss` → `charm.land/lipgloss/v2` (~33 files)
- [ ] Update `go.mod` to require new module paths
- [ ] Run `go mod tidy`

### Phase 2: Update View Signatures (part of Issue #43)
- [ ] Change all `View() string` → `View() tea.View` (32 methods)
- [ ] Wrap return values in `tea.NewView(...)`
- [ ] Update program setup: move options to View fields (`AltScreen`, `Output`)
- [ ] Fix all test `View()` implementations if any exist

### Phase 3: Update Key Event Handling (Issue #44)
- [ ] Replace all `case tea.KeyMsg:` with `case tea.KeyPressMsg:` (~31+ occurrences in source)
- [ ] Replace `msg.Type` with `msg.Code`
- [ ] Replace `msg.Runes` with `msg.Text`
- [ ] Replace `msg.Alt` with `msg.Mod` and `msg.Mod.Contains(tea.ModAlt)`
- [ ] Replace all `tea.KeyRunes` checks with `len(msg.Text) > 0`
- [ ] Replace `tea.KeyCtrlC` with `msg.String() == "ctrl+c"` or field matching
- [ ] Replace `case " ":` with `case "space":`
- [ ] Update all 224 `tea.KeyMsg{...}` struct literals in tests to `tea.KeyPressMsg{...}`

### Phase 4: Update Mouse Event Handling (Issue #45)
- [ ] Update 8 `case tea.MouseMsg:` occurrences
- [ ] Rename mouse button constants

### Phase 5: Update Bubbles Component API (Issue #46)
- [ ] **list**: Pass `isDark` to `DefaultStyles()`, update filter style references
- [ ] **table**: Pass `isDark` to `DefaultStyles()`
- [ ] **viewport**: Change `.Width = w` → `.SetWidth(w)` and `.Height = h` → `.SetHeight(h)` (10 files, ~20 occurrences)

### Phase 6: Update Lip Gloss API (part of Issue #46)
- [ ] Change struct field types: `lipgloss.Color` → `color.Color` in `theme.go`, `colors.go`
- [ ] Change return types: `strengthLabel() (string, lipgloss.Color)` → `color.Color`
- [ ] Change local variables: `var color lipgloss.Color` → `var color color.Color`
- [ ] Update `colors_test.go` — remove `string(tt.color)` cast, use proper assertion
- [ ] Check style getter methods in tests (60 occurrences)

### Phase 7: Build, Test, and Final Review (Issue #47)
- [ ] Ensure `go build` passes
- [ ] Ensure `go test ./...` passes
- [ ] Verify TUI renders correctly
- [ ] Run test scenarios
- [ ] Final code review

---

## Dependency Upgrade Order

1. `charm.land/lipgloss/v2` (no deps on the others)
2. `charm.land/bubbletea/v2` (depends on lipgloss v2)
3. `charm.land/bubbles/v2` (depends on bubbletea v2 + lipgloss v2)

All three must be upgraded together — code won't compile with mixed v1/v2.

```bash
go get charm.land/lipgloss/v2@latest
go get charm.land/bubbletea/v2@latest
go get charm.land/bubbles/v2@latest
go mod tidy
```

---

## Appendix: File Manifest

### Files importing `github.com/charmbracelet/bubbletea` (75 source + test)

<details>
<summary>Click to expand</summary>

Source files (*.go excluding *_test.go):
```
internal/tui/components/vm_cards.go
internal/tui/components/vm_table.go
internal/tui/models/cpu_options.go
internal/tui/models/cpu_options_form.go
internal/tui/models/cpu_options_form_handlers.go
internal/tui/models/cpu_topology.go
internal/tui/models/cpu_topology_form.go
internal/tui/models/cpu_topology_form_validation.go
internal/tui/models/debug.go
internal/tui/models/disk_selector.go
internal/tui/models/disk_selector_handlers.go
internal/tui/models/disk_selector_scanner.go
internal/tui/models/file_browser.go
internal/tui/models/form/form.go
internal/tui/models/form/types.go
internal/tui/models/init.go
internal/tui/models/key_handlers.go
internal/tui/models/list_adapter.go
internal/tui/models/lv_create.go
internal/tui/models/lv_create_form.go
internal/tui/models/lvm_volume.go
internal/tui/models/main.go
internal/tui/models/message_handlers.go
internal/tui/models/mount_point_warning.go
internal/tui/models/pci_passthrough.go
internal/tui/models/pci_passthrough_form.go
internal/tui/models/pci_passthrough_form_handlers.go
internal/tui/models/pci_passthrough_form_validation.go
internal/tui/models/ssh_password.go
internal/tui/models/ssh_password_form.go
internal/tui/models/ssh_password_form_validation.go
internal/tui/models/start_stop_script.go
internal/tui/models/start_stop_script_form.go
internal/tui/models/types.go
internal/tui/models/usb_passthrough.go
internal/tui/models/usb_passthrough_form.go
internal/tui/models/usb_passthrough_form_handlers.go
internal/tui/models/usb_passthrough_form_validation.go
internal/tui/models/vcpu_pinning.go
internal/tui/models/vcpu_pinning_form.go
internal/tui/models/vcpu_pinning_form_handlers.go
internal/tui/models/vcpu_pinning_form_validation.go
internal/tui/models/view_registry.go
internal/tui/models/vm_create.go
internal/tui/models/vm_delete.go
internal/tui/models/vm_edit.go
internal/tui/models/vm_form.go
internal/tui/models/vm_form_validation.go
internal/tui/models/vm_running.go
internal/tui/models/vm_selection.go
internal/tui/tui.go
main.go
```
</details>

### Files importing `github.com/charmbracelet/bubbles` (17 files)

```
internal/tui/components/vm_table.go (table)
internal/tui/models/cpu_options_form.go (viewport)
internal/tui/models/cpu_topology_form.go (viewport)
internal/tui/models/form/form.go (viewport)
internal/tui/models/init.go (list)
internal/tui/models/key_handlers.go (list)
internal/tui/models/list_adapter.go (list)
internal/tui/models/list_adapter_test.go (list)
internal/tui/models/pci_passthrough_form.go (viewport)
internal/tui/models/ssh_password_form.go (viewport)
internal/tui/models/start_stop_script_form.go (viewport)
internal/tui/models/types.go (list)
internal/tui/models/usb_passthrough_form.go (viewport)
internal/tui/models/vcpu_pinning_form.go (viewport)
internal/tui/models/vm_running.go (viewport)
internal/tui/models/vm_running_test.go (viewport)
internal/tui/models/vm_selection.go (list)
```

### Files importing `github.com/charmbracelet/lipgloss` (33 source files)

```
internal/tui/components/breadcrumbs.go
internal/tui/components/dualpane.go
internal/tui/components/statusbar.go
internal/tui/components/tabs.go
internal/tui/components/vm_cards.go
internal/tui/components/vm_details.go
internal/tui/components/vm_list_view.go
internal/tui/components/vm_table.go
internal/tui/models/cpu_options_form_render.go
internal/tui/models/disk_selector.go
internal/tui/models/file_browser.go
internal/tui/models/form/render.go (implied)
internal/tui/models/list_adapter.go
internal/tui/models/lv_create_form.go
internal/tui/models/lvm_volume.go
internal/tui/models/mount_point_warning.go
internal/tui/models/pci_passthrough_form_render.go
internal/tui/models/ssh_password_form.go
internal/tui/models/ssh_password_form_validation.go
internal/tui/models/start_stop_script_form_render.go
internal/tui/models/usb_passthrough_form_render.go
internal/tui/models/vcpu_pinning_form_render.go
internal/tui/models/view.go
internal/tui/models/vm_delete.go
internal/tui/models/vm_form_ui.go
internal/tui/styles/colors.go
internal/tui/styles/styles.go
internal/tui/theme/theme.go
```
</details>
