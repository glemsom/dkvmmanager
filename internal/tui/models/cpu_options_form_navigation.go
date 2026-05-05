// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
)

// cpuOptFocusData carries per-position metadata for the framework.
type cpuOptFocusData struct {
	kind      int // maps to form.FocusKind (int values: FocusText=0, FocusToggle=1, FocusButton=3)
	fieldName string
}

// BuildPositions returns the current list of navigable positions.
func (m *CPUOptionsFormModel) BuildPositions() []form.FocusPos {
	var positions []form.FocusPos

	// Hypervisor Stealth section header
	positions = append(positions, form.FocusPos{
		Kind: form.FocusHeader, Label: "== Hypervisor Stealth ==", Key: "header_stealth",
	})

	// Hypervisor Stealth section
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Hides VM from guest", Key: "HideKVM",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HideKVM"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusText, Label: "Custom hypervisor vendor ID", Key: "VendorID",
		Data: cpuOptFocusData{kind: int(form.FocusText), fieldName: "VendorID"},
	})

	// Hyper-V Enlightenments section header
	positions = append(positions, form.FocusPos{
		Kind: form.FocusHeader, Label: "== Hyper-V Enlightenments ==", Key: "header_hyperv",
	})

	// Hyper-V Enlightenments section
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Expose TSC/APIC frequencies", Key: "HVFrequency",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVFrequency"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Relaxed timing checks", Key: "HVRelaxed",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVRelaxed"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Guest reset capability", Key: "HVReset",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVReset"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Hypervisor runtime info", Key: "HVRuntime",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVRuntime"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusText, Label: "Paravirtualized spinlocks", Key: "HVSpinlocks",
		Data: cpuOptFocusData{kind: int(form.FocusText), fieldName: "HVSpinlocks"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Synthetic timers", Key: "HVStimer",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVStimer"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Synthetic interrupt controller", Key: "HVSyncIC",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVSyncIC"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Reference TSC page", Key: "HVTime",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVTime"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Exit-less EOI processing", Key: "HVVapic",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVVapic"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Virtual CPU index", Key: "HVVPIndex",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVVPIndex"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "SMT perf counter isolation", Key: "HVNoNonarchCoresharing",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVNoNonarchCoresharing"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Paravirtualized TLB flush", Key: "HVTLBFlush",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVTLBFlush"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Extended TLB flush ranges", Key: "HVTLBFlushExt",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVTLBFlushExt"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Paravirtualized IPI", Key: "HVIPI",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVIPI"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Hyper-V nested APIC virt", Key: "HVAVIC",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "HVAVIC"},
	})

	// Advanced CPU Features section header
	positions = append(positions, form.FocusPos{
		Kind: form.FocusHeader, Label: "== Advanced CPU Features ==", Key: "header_advanced",
	})

	// Advanced CPU Features section
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "AMD topology extension", Key: "TopoExt",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "TopoExt"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Expose host L3 cache info", Key: "L3Cache",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "L3Cache"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "x2APIC mode (>255 vCPUs)", Key: "X2APIC",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "X2APIC"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Expose all host features (no live migration)", Key: "Migratable",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "Migratable"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Invariant TSC", Key: "InvTSC",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "InvTSC"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Use UTC time for RTC", Key: "RTCUTC",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "RTCUTC"},
	})
	positions = append(positions, form.FocusPos{
		Kind: form.FocusToggle, Label: "Allow guest C/P-state control", Key: "CPUPM",
		Data: cpuOptFocusData{kind: int(form.FocusToggle), fieldName: "CPUPM"},
	})

	// Save button
	positions = append(positions, form.FocusPos{
		Kind: form.FocusButton, Label: "Save", Key: "save",
		Data: cpuOptFocusData{kind: int(form.FocusButton), fieldName: "save"},
	})

	return positions
}

// cpuOptPos is a legacy position type for backward-compatible test access.
type cpuOptPos struct {
	kind      int
	fieldName string
}

// currentPos returns the current focused position (for backward compatibility).
func (m *CPUOptionsFormModel) currentPos() cpuOptPos {
	if m.focusIndex < 0 || m.focusIndex >= len(m.positions) {
		if len(m.positions) > 0 {
			p := m.positions[0]
			return cpuOptPos{kind: int(p.Kind), fieldName: p.Key}
		}
		return cpuOptPos{}
	}
	p := m.positions[m.focusIndex]
	return cpuOptPos{kind: int(p.Kind), fieldName: p.Key}
}

// Constants for backward compatibility with existing tests.
const (
	cpuOptToggle = int(form.FocusToggle) // toggle position kind
	cpuOptText   = int(form.FocusText)   // text field position kind
	cpuOptSave   = int(form.FocusButton) // save button position kind
)

// moveFocus moves focus by delta in the flat positions list (backward compat for tests).
func (m *CPUOptionsFormModel) moveFocus(delta int) {
	m.focusIndex += delta
	if m.focusIndex < 0 {
		m.focusIndex = 0
	}
	if m.focusIndex >= len(m.positions) {
		m.focusIndex = len(m.positions) - 1
	}
}
