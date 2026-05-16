# Option 4: Keep ">" Cursor — Eliminate Line Shifting

## Problem

Currently, the selected line shifts 1 character to the left relative to unselected lines. This happens because:

| State | Prefix | `PaddingLeft` | Total left offset |
|---|---|---|---|
| **Selected** | `">  "` (3 chars) | `0` | **3 chars** |
| **Unselected** | `"  "` (2 chars) | `2` | **4 chars** |

The selected line starts 1 character left of all other lines, creating a visually jarring "jump" when navigating.

## Fix Principle

Every line uses exactly **2 characters** of uniform gutter. The first character is `>` when selected, or a space when unselected. The second character is always a space. All items use the same `PaddingLeft(0)`.

| State | Prefix | `PaddingLeft` | Total left offset |
|---|---|---|---|
| **Selected** | `"> "` (2 chars) | `0` | **2 chars** |
| **Unselected** | `"  "` (2 chars) | `0` | **2 chars** |

No shift. The `>` appears in the same column on every selected line.

---

## Files to Change

### 1. `internal/tui/models/list_adapter.go` — `MenuItemDelegate.Render`

**Lines 62-76 (selected + unselected branches):**

```go
// Current — selected:
style = styles.ListItemSelectedStyle().
    PaddingLeft(0)
str = ">  " + item.MenuItem.Title

// Current — unselected:
style = lipgloss.NewStyle().
    Foreground(styles.Colors.Muted).
    Background(styles.Colors.Background).
    PaddingLeft(2)
str = "  " + item.MenuItem.Title
```

**Replace with:**

```go
// Selected:
style = styles.ListItemSelectedStyle().
    PaddingLeft(0)
str = "> " + item.MenuItem.Title   // ">" + single space

// Unselected:
style = lipgloss.NewStyle().
    Foreground(styles.Colors.Muted).
    Background(styles.Colors.Background).
    PaddingLeft(0)
str = "  " + item.MenuItem.Title   // two spaces, same total width
```

Also fix the **disabled** branch (line 66) if its offset should match:

```go
// Current:
style = styles.ListItemDisabledStyle().
    PaddingLeft(0).
    Background(styles.Colors.Background)
str := "  " + item.MenuItem.Title   // This is already 2-space prefix ✓
```

The disabled branch already uses `"  "` with `PaddingLeft(0)` — so it's already correct and won't shift. Good.

---

### 2. `internal/tui/models/list_adapter.go` — `VMListItemDelegate.Render`

**Lines 107-119 (selected + unselected branches):**

```go
// Current — selected:
style = styles.ListItemSelectedStyle().
    PaddingLeft(0)
str = ">  " + item.VM.Name

// Current — unselected:
style = lipgloss.NewStyle().
    Foreground(styles.Colors.Muted).
    Background(styles.Colors.Background).
    PaddingLeft(2)
str = "  " + item.VM.Name
```

**Replace with** (same pattern as MenuItemDelegate):

```go
// Selected:
style = styles.ListItemSelectedStyle().
    PaddingLeft(0)
str = "> " + item.VM.Name

// Unselected:
style = lipgloss.NewStyle().
    Foreground(styles.Colors.Muted).
    Background(styles.Colors.Background).
    PaddingLeft(0)
str = "  " + item.VM.Name
```

---

### 3. `internal/tui/components/vm_list_view.go` — `VMListView.View`

**Lines 91-103 (cursor variables + rendering loop):**

```go
// Current:
normalCursor := " "
selectedCursor := ">"
if v.focused {
    selectedCursor = lipgloss.NewStyle().
        Foreground(styles.Colors.Primary).
        Bold(true).
        Render(">")
} else {
    selectedCursor = lipgloss.NewStyle().
        Foreground(styles.Colors.Muted).
        Render(">")
}
```

The issue here isn't in the `PaddingLeft` but in how the line is assembled on line 135:

```go
// Line 135 — current:
line := cursor + " " + statusIcon + " " + nameStyle.Render(vm.Name)
```

- For selected: `cursor` = `">"` → line starts with `"> "` (2 chars: `>` + space from the literal) ✓  
- For unselected: `cursor` = `" "` → line starts with `"  "` (2 chars: space + space from the literal) ✓

Actually, on closer inspection, **`vm_list_view.go` already has no shift**. Both selected and unselected lines have a 2-character prefix:
- Selected: `">"` (1 char) + `" "` (literal space) = 2 chars ✓
- Unselected: `" "` (1 char) + `" "` (literal space) = 2 chars ✓

**No change needed here.** The perceived shift in the Configuration tab must be coming entirely from `list_adapter.go`.

---

### 4. `internal/tui/components/vm_table.go` — `VMTable.View`

**Lines 99-108:**

```go
// Current:
copied[cursor][0] = "> " + copied[cursor][0]
```

This prepends `"> "` to the first column of only the selected row. For consistency with the no-shift approach, every row should have the same gutter. Change to:

```go
// Add a 2-char gutter to all rows, with "> " on the selected row
for i, row := range copied {
    if i == cursor {
        copied[i][0] = "> " + row[0]
    } else {
        copied[i][0] = "  " + row[0]
    }
}
```

This ensures all rows start with exactly 2 characters of gutter, eliminating any horizontal shift on the selected row.

---

## Summary of Changes

| File | Lines | Change |
|---|---|---|
| `list_adapter.go` — `MenuItemDelegate.Render` | 66-76 | `">  "` → `"> "` for selected; drop `PaddingLeft(2)` on unselected (use `PaddingLeft(0)`) |
| `list_adapter.go` — `VMListItemDelegate.Render` | 107-119 | Same as above |
| `vm_list_view.go` — `VMListView.View` | 91-135 | No change needed (already aligned) |
| `vm_table.go` — `VMTable.View` | 99-108 | Apply `"  "` gutter to all rows, `"> "` to selected only |

## Result

In the Configuration tab (and all other list views), every line starts exactly 2 characters from the left edge. The selected line shows `>` in the first character position, while unselected lines show a space. No text jumps when navigating up/down.
