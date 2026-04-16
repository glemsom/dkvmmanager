# Implementation Plan: Enter on "Volume Group" opens dropdown in Create Logical Volume

## Goal
In **Configuration → Create Logical Volume**, pressing **Enter** while focus is on **Volume Group** should open a dropdown/list of available LVM VGs and allow selecting one.

Also address current issue where VG appears blank and left/right does not show expected groups.

## Constraints & Project Rules
1. Keep TUI compatible with **80x25** terminal.
2. Follow project workflow: **create/approve golden file before implementation**.
3. Add tests using TDD.
4. Delegate test/build execution to `tester` subagent.

---

## Phase 1 — Visual Spec (Golden First)
1. Add a new golden file in:
   - `internal/tui/models/testdata/lv_create_form_vg_dropdown_open.golden`
2. Golden should show:
   - LV form with focus on `Volume Group`.
   - Expanded dropdown/list panel under VG field.
   - At least 2 sample VGs (one selected, one unselected).
   - Error line area if no VGs found (for empty state golden in a second file if needed).
3. Share the golden output for approval before code changes.

---

## Phase 2 — Data Loading & Diagnostics
1. Review/strengthen VG loading in `lv_create_form.go`:
   - command: `vgs --noheadings -o vg_name,vg_size,vg_free,lv_count --units g --separator "\t"`
2. Improve error capture so failures are visible:
   - if `vgs` exits non-zero, include stderr in error message.
3. Confirm parser handles:
   - extra spaces
n   - empty lines
   - missing columns safely
4. Preserve loaded VGs in model state and ensure default selection index is valid when data arrives.

---

## Phase 3 — Model State for Dropdown
1. Extend `LVCreateFormModel` with dropdown state, e.g.:
   - `vgDropdownOpen bool`
   - optional `vgDropdownIndex int` (if different from selected `vgIndex`)
2. Add helper methods:
   - `openVGDropdown()`
   - `closeVGDropdown()`
   - `confirmVGSelection()`
   - `moveVGSelection(delta int)`

---

## Phase 4 — Keyboard Behavior Changes
1. Update `handleKey` in `lv_create_form.go`:
   - **Enter** when focus is `lvFocusVG`:
     - if closed → open dropdown
     - if open → confirm highlighted VG and close
   - **Esc** when dropdown is open: close dropdown only (do not leave form)
   - **Up/Down** while dropdown open on VG: move within VG list
   - **Left/Right** while dropdown open: optional no-op or same as up/down (define explicitly)
2. Keep existing submission behavior:
   - Enter should still submit from non-VG fields (unless UX changed intentionally).
   - Ensure pressing Enter on VG no longer submits form immediately.

---

## Phase 5 — Rendering Changes
1. Update `renderLines()`:
   - show VG field with indicator:
     - closed: `▼`
     - open: `▲` (or similar)
2. When open, render dropdown rows below VG line:
   - selected row prefixed with `>`
   - include helpful details (name + free size)
3. Clip/truncate lines to fit current viewport width.
4. Keep total visual stable for 80x25; rely on viewport scrolling if needed.

---

## Phase 6 — Tests (TDD)
Add/extend tests in `internal/tui/models/lv_create_form_test.go`:
1. `TestLVCreateEnterOnVGOpensDropdown`
2. `TestLVCreateEnterOnVGWhenOpenConfirmsSelection`
3. `TestLVCreateEscClosesDropdownOnly`
4. `TestLVCreateUpDownNavigatesVGDropdown`
5. `TestLVCreateEnterOnNonVGStillSubmits`
6. `TestParseVGSOutput_WhitespaceAndEmptyLines`
7. `TestLVCreateRenderDropdownOpen` (string assertions or golden-based render test)

Optional:
- Add a regression test for blank VG when valid VGs are loaded.

---

## Phase 7 — Integration Wiring Check
1. Verify no regressions in `MainModel` delegation path for `ViewLVCreate` (`key_handlers.go`).
2. Ensure `lvVGsLoadedMsg` still handled correctly.
3. Confirm status/error messages are visible and not overwritten unexpectedly.

---

## Phase 8 — Validation
1. Run targeted tests for LV form.
2. Run full test suite (`go test ./...`) via `tester` subagent.
3. Build via `tester` subagent (`make build` or `go build ./...`).

---

## Deliverables
1. Approved golden file(s) for dropdown behavior.
2. Updated LV create form model with Enter-to-open dropdown UX.
3. Test coverage for keyboard handling, rendering, and parser robustness.
4. Passing tests and successful build.
