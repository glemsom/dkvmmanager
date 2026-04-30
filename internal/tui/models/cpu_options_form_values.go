// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

// getToggleValue returns the boolean value for a toggle field
func (m *CPUOptionsFormModel) getToggleValue(fieldName string) bool {
	switch fieldName {
	case "HideKVM":
		return m.options.HideKVM
	case "HVFrequency":
		return m.options.HVFrequency
	case "HVRelaxed":
		return m.options.HVRelaxed
	case "HVReset":
		return m.options.HVReset
	case "HVRuntime":
		return m.options.HVRuntime
	case "HVStimer":
		return m.options.HVStimer
	case "HVSyncIC":
		return m.options.HVSyncIC
	case "HVTime":
		return m.options.HVTime
	case "HVVapic":
		return m.options.HVVapic
	case "HVVPIndex":
		return m.options.HVVPIndex
	case "HVNoNonarchCoresharing":
		return m.options.HVNoNonarchCoresharing
	case "HVTLBFlush":
		return m.options.HVTLBFlush
	case "HVTLBFlushExt":
		return m.options.HVTLBFlushExt
	case "HVIPI":
		return m.options.HVIPI
	case "HVAVIC":
		return m.options.HVAVIC
	case "TopoExt":
		return m.options.TopoExt
	case "L3Cache":
		return m.options.L3Cache
	case "X2APIC":
		return m.options.X2APIC
	case "Migratable":
		return m.options.Migratable
	case "InvTSC":
		return m.options.InvTSC
	case "RTCUTC":
		return m.options.RTCUTC
	case "CPUPM":
		return m.options.CPUPM
	}
	return false
}

// toggleValue toggles a boolean field
func (m *CPUOptionsFormModel) toggleValue(fieldName string) {
	switch fieldName {
	case "HideKVM":
		m.options.HideKVM = !m.options.HideKVM
	case "HVFrequency":
		m.options.HVFrequency = !m.options.HVFrequency
	case "HVRelaxed":
		m.options.HVRelaxed = !m.options.HVRelaxed
	case "HVReset":
		m.options.HVReset = !m.options.HVReset
	case "HVRuntime":
		m.options.HVRuntime = !m.options.HVRuntime
	case "HVStimer":
		m.options.HVStimer = !m.options.HVStimer
	case "HVSyncIC":
		m.options.HVSyncIC = !m.options.HVSyncIC
	case "HVTime":
		m.options.HVTime = !m.options.HVTime
	case "HVVapic":
		m.options.HVVapic = !m.options.HVVapic
	case "HVVPIndex":
		m.options.HVVPIndex = !m.options.HVVPIndex
	case "HVNoNonarchCoresharing":
		m.options.HVNoNonarchCoresharing = !m.options.HVNoNonarchCoresharing
	case "HVTLBFlush":
		m.options.HVTLBFlush = !m.options.HVTLBFlush
	case "HVTLBFlushExt":
		m.options.HVTLBFlushExt = !m.options.HVTLBFlushExt
	case "HVIPI":
		m.options.HVIPI = !m.options.HVIPI
	case "HVAVIC":
		m.options.HVAVIC = !m.options.HVAVIC
	case "TopoExt":
		m.options.TopoExt = !m.options.TopoExt
	case "L3Cache":
		m.options.L3Cache = !m.options.L3Cache
	case "X2APIC":
		m.options.X2APIC = !m.options.X2APIC
	case "Migratable":
		m.options.Migratable = !m.options.Migratable
	case "InvTSC":
		m.options.InvTSC = !m.options.InvTSC
	case "RTCUTC":
		m.options.RTCUTC = !m.options.RTCUTC
	case "CPUPM":
		m.options.CPUPM = !m.options.CPUPM
	}
}

// getTextValue returns the text value for a field
func (m *CPUOptionsFormModel) getTextValue(fieldName string) string {
	switch fieldName {
	case "VendorID":
		return m.options.VendorID
	case "HVSpinlocks":
		return m.options.HVSpinlocks
	}
	return ""
}

// setTextValue sets the text value for a field
func (m *CPUOptionsFormModel) setTextValue(fieldName string, val string) {
	switch fieldName {
	case "VendorID":
		m.options.VendorID = val
	case "HVSpinlocks":
		m.options.HVSpinlocks = val
	}
}
