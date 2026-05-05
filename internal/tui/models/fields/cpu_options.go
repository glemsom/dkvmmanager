package fields

// FieldKind represents the type of form field
type FieldKind int

const (
	FieldToggle FieldKind = iota
	FieldText
)

// FieldMeta holds metadata for a form field
type FieldMeta struct {
	Name string
	Kind FieldKind
}

// CPUOptionsFields defines the field order and types for CPU options.
// This is the single source of truth for CPU options form fields.
var CPUOptionsFields = []FieldMeta{
	{Name: "HideKVM", Kind: FieldToggle},
	{Name: "VendorID", Kind: FieldText},
	{Name: "HVFrequency", Kind: FieldToggle},
	{Name: "HVRelaxed", Kind: FieldToggle},
	{Name: "HVReset", Kind: FieldToggle},
	{Name: "HVRuntime", Kind: FieldToggle},
	{Name: "HVSpinlocks", Kind: FieldText},
	{Name: "HVStimer", Kind: FieldToggle},
	{Name: "HVSyncIC", Kind: FieldToggle},
	{Name: "HVTime", Kind: FieldToggle},
	{Name: "HVVapic", Kind: FieldToggle},
	{Name: "HVVPIndex", Kind: FieldToggle},
	{Name: "HVNoNonarchCoresharing", Kind: FieldToggle},
	{Name: "HVTLBFlush", Kind: FieldToggle},
	{Name: "HVTLBFlushExt", Kind: FieldToggle},
	{Name: "HVIPI", Kind: FieldToggle},
	{Name: "HVAVIC", Kind: FieldToggle},
	{Name: "TopoExt", Kind: FieldToggle},
	{Name: "L3Cache", Kind: FieldToggle},
	{Name: "X2APIC", Kind: FieldToggle},
	{Name: "Migratable", Kind: FieldToggle},
	{Name: "InvTSC", Kind: FieldToggle},
	{Name: "RTCUTC", Kind: FieldToggle},
	{Name: "CPUPM", Kind: FieldToggle},
}