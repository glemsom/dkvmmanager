package vm

import (
	"github.com/glemsom/dkvmmanager/internal/models"
)

// GetCPUOptions returns the global CPU options configuration
func (r *Repository) GetCPUOptions() (models.CPUOptions, error) {
	var opts models.CPUOptions

	if !r.vip.IsSet("cpu_options") {
		return opts, nil // Return defaults if not set
	}

	data := r.vip.GetStringMap("cpu_options")
	opts.HideKVM = getBool(data, "hide_kvm")
	opts.VendorID = getString(data, "vendor_id")
	opts.HVFrequency = getBool(data, "hv_frequency")
	opts.HVRelaxed = getBool(data, "hv_relaxed")
	opts.HVReset = getBool(data, "hv_reset")
	opts.HVRuntime = getBool(data, "hv_runtime")
	opts.HVSpinlocks = getString(data, "hv_spinlocks")
	opts.HVStimer = getBool(data, "hv_stimer")
	opts.HVSyncIC = getBool(data, "hv_synic")
	opts.HVTime = getBool(data, "hv_time")
	opts.HVVapic = getBool(data, "hv_vapic")
	opts.HVVPIndex = getBool(data, "hv_vpindex")
	opts.HVNoNonarchCoresharing = getBool(data, "hv_no_nonarch_coresharing")
	opts.HVTLBFlush = getBool(data, "hv_tlbflush")
	opts.HVTLBFlushExt = getBool(data, "hv_tlbflush_ext")
	opts.HVIPI = getBool(data, "hv_ipi")
	opts.HVAVIC = getBool(data, "hv_avic")
	opts.TopoExt = getBool(data, "topoext")
	opts.L3Cache = getBool(data, "l3_cache")
	opts.X2APIC = getBool(data, "x2apic")
	opts.Migratable = getBool(data, "migratable")
	opts.InvTSC = getBool(data, "invtsc")
	opts.RTCUTC = getBool(data, "rtc_utc")

	return opts, nil
}

// SaveCPUOptions saves the global CPU options configuration
func (r *Repository) SaveCPUOptions(opts models.CPUOptions) error {
	data := map[string]interface{}{
		"hide_kvm":                  opts.HideKVM,
		"vendor_id":                 opts.VendorID,
		"hv_frequency":              opts.HVFrequency,
		"hv_relaxed":                opts.HVRelaxed,
		"hv_reset":                  opts.HVReset,
		"hv_runtime":                opts.HVRuntime,
		"hv_spinlocks":              opts.HVSpinlocks,
		"hv_stimer":                 opts.HVStimer,
		"hv_synic":                  opts.HVSyncIC,
		"hv_time":                   opts.HVTime,
		"hv_vapic":                  opts.HVVapic,
		"hv_vpindex":                opts.HVVPIndex,
		"hv_no_nonarch_coresharing": opts.HVNoNonarchCoresharing,
		"hv_tlbflush":               opts.HVTLBFlush,
		"hv_tlbflush_ext":           opts.HVTLBFlushExt,
		"hv_ipi":                    opts.HVIPI,
		"hv_avic":                   opts.HVAVIC,
		"topoext":                   opts.TopoExt,
		"l3_cache":                  opts.L3Cache,
		"x2apic":                    opts.X2APIC,
		"migratable":                opts.Migratable,
		"invtsc":                    opts.InvTSC,
		"rtc_utc":                   opts.RTCUTC,
	}

	r.vip.Set("cpu_options", data)
	return r.save()
}
