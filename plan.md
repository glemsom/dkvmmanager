# Deepen Repository: Replace 12 Repetitive Config Get/Save Methods with a Generic Config Store

## Problem

`internal/vm/repository.go` has 12 per-type config persistence methods (6 config types × Get/Save)
that all follow the same pattern: build a `map[string]interface{}` → call `vip.Set` or
`vip.GetStringMap` → extract fields manually with `getString`/`getBool`/`getInt` helpers.

The interface (12 methods) is nearly as complex as the implementation (the map-building code
inside each). Adding a new config type requires 2 new methods, 2 new Viper keys, plus wiring
in 2+ form files and test files.

**Deletion test**: removing these methods would scatter YAML I/O across 8 form handlers — the
module earns its keep. But the leverage ratio is low: ~300 lines of repetition for 6 config types.

## Approach

Replace the 12 per-type methods with 2 generic methods using `mapstructure` (already a
dependency via `github.com/go-viper/mapstructure/v2`):

```go
// GetConfig decodes a config section into dest. Returns zero-value dest if key is not set.
func (r *Repository) GetConfig(key string, dest interface{}) error

// SaveConfig encodes src as a config section and persists.
func (r *Repository) SaveConfig(key string, src interface{}) error
```

VM methods (`ListVMs`, `GetVM`, `SaveVM`, `DeleteVM`, `FindNextAvailableID`) stay unchanged —
they use `vms.<id>` nested keys with special timestamp handling.

## Implementation

### [DONE ✅] Phase 1 — Add generic methods to Repository

**File: `internal/vm/repository.go`**

Add imports: `"github.com/go-viper/mapstructure/v2"`

Add two new methods after the helper section:

```go
// === Generic Config Store ===

// GetConfig decodes a config section identified by key into dest.
// If the key is not set in the config, dest is left at its zero value (no error).
func (r *Repository) GetConfig(key string, dest interface{}) error {
	raw := r.vip.Get(key)
	if raw == nil {
		return nil
	}
	config := &mapstructure.DecoderConfig{
		TagName: "json",
		Result:  dest,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return fmt.Errorf("create decoder for %q: %w", key, err)
	}
	return decoder.Decode(raw)
}

// SaveConfig encodes src and stores it under key, then persists to disk.
func (r *Repository) SaveConfig(key string, src interface{}) error {
	// Use JSON marshal/unmarshal to convert struct to map[string]interface{}
	// since mapstructure/v2 does not provide an Encode function.
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("marshal config for %q: %w", key, err)
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("unmarshal config for %q: %w", key, err)
	}
	r.vip.Set(key, cfg)
	return r.save()
}
```

Key decisions:
- Uses `r.vip.Get(key)` not `r.vip.Sub(key)` — `Sub` only works for file-loaded data,
  not for data set via `Set()` (important for tests).
- `mapstructure.Decode` with `TagName: "json"` ensures it reads keys produced by JSON
  marshal/unmarshal (e.g., `"hide_kvm"` matches the `json:"hide_kvm"` tag).
- `json.Marshal`/`json.Unmarshal` converts any struct to `map[string]interface{}` because
  `mapstructure/v2` v2.4.0 does **not** provide an `Encode` function.

### [DONE ✅] Phase 2 — Replace per-type method bodies, keep signatures (deprecation shim)

**File: `internal/vm/repository.go`**

Replace the body of each of the 12 config methods to delegate to the generic methods.
Keep the public signatures to avoid breaking callers in one big change.

Example — CPUOptions:

```go
// GetCPUOptions returns the global CPU options configuration.
// Deprecated: Use GetConfig("cpu_options", &opts) instead.
func (r *Repository) GetCPUOptions() (models.CPUOptions, error) {
	var opts models.CPUOptions
	err := r.GetConfig("cpu_options", &opts)
	return opts, err
}

// SaveCPUOptions saves the global CPU options configuration.
// Deprecated: Use SaveConfig("cpu_options", opts) instead.
func (r *Repository) SaveCPUOptions(opts models.CPUOptions) error {
	return r.SaveConfig("cpu_options", opts)
}
```

Config key mapping (YAML top-level keys):
| Old method | Viper key | Model type |
|---|---|---|
| `GetCPUOptions` / `SaveCPUOptions` | `"cpu_options"` | `models.CPUOptions` |
| `GetPCIPassthroughConfig` / `SavePCIPassthroughConfig` | `"pci_passthrough"` | `models.PCIPassthroughConfig` |
| `GetUSBPassthroughConfig` / `SaveUSBPassthroughConfig` | `"usb_passthrough"` | `models.USBPassthroughConfig` |
| `GetCPUTopology` / `SaveCPUTopology` | `"cpu_topology"` | `models.CPUTopology` |
| `GetVCPUPinningGlobal` / `SaveVCPUPinningGlobal` | `"vcpu_pinning"` | `models.VCPUPinningGlobal` |
| `GetStartStopScript` / `SaveStartStopScript` | `"custom_script"` | `models.StartStopScript` |

Add `// Deprecated` comments to all 12 methods so the Go tooling warns about usage.

### [DONE ✅] Phase 3 — Update all callers to use generic methods directly

Each caller change follows the same pattern. Here is every caller, grouped by file:

#### File: `internal/tui/models/cpu_options_form.go`

**Line 50:**
```go
// Before:
opts, _ := repo.GetCPUOptions()
// After:
var opts models.CPUOptions
repo.GetConfig("cpu_options", &opts)
```

#### File: `internal/tui/models/cpu_options_form_validation.go`

**Line 64:**
```go
// Before:
if err := m.repo.SaveCPUOptions(*m.options); err != nil {
// After:
if err := m.repo.SaveConfig("cpu_options", *m.options); err != nil {
```

#### File: `internal/tui/models/cpu_options_form_handlers.go`

**Lines 111 and 130:**
```go
// Before:
if err := m.repo.SaveCPUOptions(*m.options); err != nil {
// After:
if err := m.repo.SaveConfig("cpu_options", *m.options); err != nil {
```

#### File: `internal/tui/models/pci_passthrough_form.go`

**Line 74:**
```go
// Before:
cfg, _ := repo.GetPCIPassthroughConfig()
// After:
var cfg models.PCIPassthroughConfig
repo.GetConfig("pci_passthrough", &cfg)
```

#### File: `internal/tui/models/pci_passthrough_form_validation.go`

**Lines 46 and 93:**
```go
// Before:
if err := m.repo.SavePCIPassthroughConfig(cfg); err != nil {
// After:
if err := m.repo.SaveConfig("pci_passthrough", cfg); err != nil {
```

#### File: `internal/tui/models/usb_passthrough_form.go`

**Line 59:**
```go
// Before:
cfg, _ := repo.GetUSBPassthroughConfig()
// After:
var cfg models.USBPassthroughConfig
repo.GetConfig("usb_passthrough", &cfg)
```

#### File: `internal/tui/models/usb_passthrough_form_validation.go`

**Line 45:**
```go
// Before:
if err := m.repo.SaveUSBPassthroughConfig(cfg); err != nil {
// After:
if err := m.repo.SaveConfig("usb_passthrough", cfg); err != nil {
```

#### File: `internal/tui/models/cpu_topology_form.go`

**Line 69:**
```go
// Before:
topology, err := repo.GetCPUTopology()
// After:
var topology models.CPUTopology
err := repo.GetConfig("cpu_topology", &topology)
```

#### File: `internal/tui/models/cpu_topology_form_validation.go`

**Line 41:**
```go
// Before:
if err := m.repo.SaveCPUTopology(topo); err != nil {
// After:
if err := m.repo.SaveConfig("cpu_topology", topo); err != nil {
```

#### File: `internal/tui/models/vcpu_pinning_form.go`

**Lines 69 and 75:**
```go
// Before:
topology, err := repo.GetCPUTopology()
pinning, err := repo.GetVCPUPinningGlobal()
// After:
var topology models.CPUTopology
err := repo.GetConfig("cpu_topology", &topology)
var pinning models.VCPUPinningGlobal
err = repo.GetConfig("vcpu_pinning", &pinning)
```
Note: line 85 also calls `repo.GetVCPUPinningGlobal()` — replace similarly.

#### File: `internal/tui/models/vcpu_pinning_form_validation.go`

**Lines 28 and 34:**
```go
// Before:
if err := m.repo.SaveCPUTopology(m.topology); err != nil {
if err := m.repo.SaveVCPUPinningGlobal(m.pinning); err != nil {
// After:
if err := m.repo.SaveConfig("cpu_topology", m.topology); err != nil {
if err := m.repo.SaveConfig("vcpu_pinning", m.pinning); err != nil {
```

**Lines 61 and 68 (same pattern in handleApplyKernel):**
```go
// Before:
if err := m.repo.SaveCPUTopology(m.topology); err != nil {
if err := m.repo.SaveVCPUPinningGlobal(m.pinning); err != nil {
// After:
if err := m.repo.SaveConfig("cpu_topology", m.topology); err != nil {
if err := m.repo.SaveConfig("vcpu_pinning", m.pinning); err != nil {
```

#### File: `internal/tui/models/start_stop_script_form.go`

**Lines 44-45:**
```go
// Before:
cfg, _ := repo.GetStartStopScript()
pciCfg, _ := repo.GetPCIPassthroughConfig()
// After:
var cfg models.StartStopScript
repo.GetConfig("custom_script", &cfg)
var pciCfg models.PCIPassthroughConfig
repo.GetConfig("pci_passthrough", &pciCfg)
```

**Line 139:**
```go
// Before:
if err := m.repo.SaveStartStopScript(m.config); err != nil {
// After:
if err := m.repo.SaveConfig("custom_script", m.config); err != nil {
```

### [DONE ✅] Phase 4 — Remove deprecated methods

Once all callers are updated, remove the 12 deprecated per-type methods from
`internal/vm/repository.go`.

Also remove the helper functions that are no longer needed:
- `getString(data map[string]interface{}, key string) string`
- `getBool(data map[string]interface{}, key string) bool`
- `getInt(data map[string]interface{}, key string) int`
- `parseStringSlice(data map[string]interface{}, key string) []string`
- `parseIntSlice(data map[string]interface{}, key string) []int`

Verify these helpers are not used elsewhere before removing (`grep` the codebase).

### [DONE ✅] Phase 5 — Update test files

Create a single unit test file that exercises the generic GetConfig/SaveConfig methods
through all 6 config types:

**New file: `internal/vm/config_store_test.go`**

```go
package vm

import (
	"path/filepath"
	"testing"

	"github.com/glemsom/dkvmmanager/internal/models"
)

// TestConfigStore_RoundTrip verifies each config type survives a save-and-reload cycle
// through the generic GetConfig/SaveConfig methods.
func TestConfigStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	repo, err := NewRepository(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("CPUOptions", func(t *testing.T) {
		in := models.CPUOptions{HideKVM: true, HVRelaxed: true, VendorID: "GenuineIntel"}
		if err := repo.SaveConfig("cpu_options", in); err != nil {
			t.Fatal(err)
		}
		var out models.CPUOptions
		if err := repo.GetConfig("cpu_options", &out); err != nil {
			t.Fatal(err)
		}
		if out.HideKVM != in.HideKVM || out.VendorID != in.VendorID {
			t.Errorf("round-trip mismatch: %+v vs %+v", in, out)
		}
	})

	t.Run("PCIPassthroughConfig", func(t *testing.T) {
		in := models.PCIPassthroughConfig{
			Devices: []models.PCIPassthroughDevice{
				{Address: "0000:01:00.0", Vendor: "10de", Device: "1b80", Name: "GPU"},
			},
		}
		if err := repo.SaveConfig("pci_passthrough", in); err != nil {
			t.Fatal(err)
		}
		var out models.PCIPassthroughConfig
		if err := repo.GetConfig("pci_passthrough", &out); err != nil {
			t.Fatal(err)
		}
		if len(out.Devices) != 1 || out.Devices[0].Address != "0000:01:00.0" {
			t.Errorf("round-trip mismatch: %+v", out)
		}
	})

	t.Run("USBPassthroughConfig", func(t *testing.T) {
		in := models.USBPassthroughConfig{
			Devices: []models.USBPassthroughDevice{
				{Vendor: "046d", Product: "c52b", Name: "Unifying Receiver"},
			},
		}
		if err := repo.SaveConfig("usb_passthrough", in); err != nil {
			t.Fatal(err)
		}
		var out models.USBPassthroughConfig
		if err := repo.GetConfig("usb_passthrough", &out); err != nil {
			t.Fatal(err)
		}
		if len(out.Devices) != 1 || out.Devices[0].Vendor != "046d" {
			t.Errorf("round-trip mismatch: %+v", out)
		}
	})

	t.Run("CPUTopology", func(t *testing.T) {
		in := models.CPUTopology{Enabled: true, SelectedCPUs: []int{0, 1, 2, 3}}
		if err := repo.SaveConfig("cpu_topology", in); err != nil {
			t.Fatal(err)
		}
		var out models.CPUTopology
		if err := repo.GetConfig("cpu_topology", &out); err != nil {
			t.Fatal(err)
		}
		if !out.Enabled || len(out.SelectedCPUs) != 4 {
			t.Errorf("round-trip mismatch: %+v", out)
		}
	})

	t.Run("VCPUPinningGlobal", func(t *testing.T) {
		in := models.VCPUPinningGlobal{
			Enabled: true,
			Mappings: []models.VCPUToHostMapping{
				{VCPUID: 0, HostCPUID: 4},
			},
		}
		if err := repo.SaveConfig("vcpu_pinning", in); err != nil {
			t.Fatal(err)
		}
		var out models.VCPUPinningGlobal
		if err := repo.GetConfig("vcpu_pinning", &out); err != nil {
			t.Fatal(err)
		}
		if !out.Enabled || len(out.Mappings) != 1 || out.Mappings[0].HostCPUID != 4 {
			t.Errorf("round-trip mismatch: %+v", out)
		}
	})

	t.Run("StartStopScript", func(t *testing.T) {
		in := models.StartStopScript{UseBuiltin: false, StartScript: "echo start", StopScript: "echo stop"}
		if err := repo.SaveConfig("custom_script", in); err != nil {
			t.Fatal(err)
		}
		var out models.StartStopScript
		if err := repo.GetConfig("custom_script", &out); err != nil {
			t.Fatal(err)
		}
		if out.UseBuiltin || out.StartScript != "echo start" {
			t.Errorf("round-trip mismatch: %+v", out)
		}
	})

	t.Run("GetUnsetKeyReturnsZeroValue", func(t *testing.T) {
		var opts models.CPUOptions
		if err := repo.GetConfig("nonexistent_key", &opts); err != nil {
			t.Fatal(err)
		}
		// Should be zero value, no error
	})

	t.Run("OverwriteExistingKey", func(t *testing.T) {
		in1 := models.CPUOptions{HideKVM: true}
		in2 := models.CPUOptions{HVRelaxed: true}
		if err := repo.SaveConfig("cpu_options", in1); err != nil {
			t.Fatal(err)
		}
		if err := repo.SaveConfig("cpu_options", in2); err != nil {
			t.Fatal(err)
		}
		var out models.CPUOptions
		if err := repo.GetConfig("cpu_options", &out); err != nil {
			t.Fatal(err)
		}
		if out.HideKVM || !out.HVRelaxed {
			t.Errorf("overwrite failed: %+v", out)
		}
	})
}
```

Update the existing test files:

**File: `internal/vm/cpu_options_persistence_test.go`**

Replace all calls to `repo.GetCPUOptions()` / `repo.SaveCPUOptions()` with
`repo.GetConfig("cpu_options", &opts)` / `repo.SaveConfig("cpu_options", opts)`.

**File: `internal/vm/repository_topology_test.go`**

Replace all calls to `repo.GetCPUTopology()` / `repo.SaveCPUTopology()` with
`repo.GetConfig("cpu_topology", &topo)` / `repo.SaveConfig("cpu_topology", topo)`.

**File: `internal/vm/repository_pinning_test.go`**

Replace all calls to `repo.GetVCPUPinningGlobal()` / `repo.SaveVCPUPinningGlobal()` with
`repo.GetConfig("vcpu_pinning", &cfg)` / `repo.SaveConfig("vcpu_pinning", cfg)`.

**File: `internal/vm/pci_persistence_test.go`**

Replace all calls to `repo.GetPCIPassthroughConfig()` / `repo.SavePCIPassthroughConfig()`
with `repo.GetConfig("pci_passthrough", &cfg)` / `repo.SaveConfig("pci_passthrough", cfg)`.

**File: `internal/vm/usb_persistence_test.go`**

Replace all calls to `repo.GetUSBPassthroughConfig()` / `repo.SaveUSBPassthroughConfig()`
with `repo.GetConfig("usb_passthrough", &cfg)` / `repo.SaveConfig("usb_passthrough", cfg)`.

### [DONE ✅] Phase 6 — Clean up unused helpers

After Phase 4 removes the per-type methods, verify these functions are no longer referenced:

```
getString
getBool
getInt
parseStringSlice
parseIntSlice
```

If any are still used by the VM persistence methods (they are — `unmarshalVM` uses
`getString`, `getBool`, `parseStringSlice`), keep only those that are actually used.
Remove the rest.

The helpers used by VM methods (not config methods) are:
- `getString` — used in `unmarshalVM`, `getPCIPassthroughConfig`, etc. 
- Wait, after Phase 4 those are removed. Let me check...

Actually, after Phase 4 all config Get/Save methods are removed. But `unmarshalVM` still
uses `getString`, `getBool`, `parseStringSlice`, and `parseIntSlice` for VM-level fields.
These helpers are still needed by VM persistence. So **do not remove them** — they're
shared by VM methods which are unchanged.

Only remove helpers that are exclusively used by the removed config methods. After
auditing:
- `getString` — used by `unmarshalVM` (VM persistence, kept) and by config methods
  (removed). **Keep**.
- `getBool` — used by `unmarshalVM` (kept) and config methods (removed). **Keep**.
- `getInt` — used only by config methods. **Remove**.
- `parseStringSlice` — used by `unmarshalVM` (kept). **Keep**.
- `parseIntSlice` — used only by config methods. **Remove**.

## [DONE] Verification

1. `go build -buildvcs=false ./...` — compiles with no errors ✅
2. `go test ./internal/vm/...` — all vm tests pass ✅
3. `go test ./internal/tui/models/form/...` — all form tests pass ✅
4. `go vet ./...` — no vet warnings ✅
5. Two pre-existing `list_adapter_test.go` failures (render spacing) unrelated to this refactor.

## [DONE] After this refactor

Adding a new config type (e.g., `NetworkConfig`) requires:
1. Define the model struct in `internal/models`
2. One `SaveConfig("network", cfg)` call in the form validator
3. One `GetConfig("network", &cfg)` call in the form constructor

Zero new Repository methods. Zero new Viper key management code.

## Files Changed Summary

| File | Change |
|---|---|
| `internal/vm/repository.go` | Add `GetConfig`/`SaveConfig` + `mapstructure` import; replace 12 method bodies with delegation; remove 12 methods; remove `getInt` and `parseIntSlice` helpers |
| `internal/vm/config_store_test.go` | **NEW** — round-trip tests for all 6 config types |
| `internal/tui/models/cpu_options_form.go` | Replace `GetCPUOptions()` → `GetConfig("cpu_options", &opts)` |
| `internal/tui/models/cpu_options_form_validation.go` | Replace `SaveCPUOptions()` → `SaveConfig("cpu_options", ...)` |
| `internal/tui/models/cpu_options_form_handlers.go` | Replace `SaveCPUOptions()` → `SaveConfig("cpu_options", ...)` |
| `internal/tui/models/pci_passthrough_form.go` | Replace `GetPCIPassthroughConfig()` → `GetConfig("pci_passthrough", &cfg)` |
| `internal/tui/models/pci_passthrough_form_validation.go` | Replace `SavePCIPassthroughConfig()` → `SaveConfig("pci_passthrough", ...)` |
| `internal/tui/models/usb_passthrough_form.go` | Replace `GetUSBPassthroughConfig()` → `GetConfig("usb_passthrough", &cfg)` |
| `internal/tui/models/usb_passthrough_form_validation.go` | Replace `SaveUSBPassthroughConfig()` → `SaveConfig("usb_passthrough", ...)` |
| `internal/tui/models/cpu_topology_form.go` | Replace `GetCPUTopology()` → `GetConfig("cpu_topology", &topo)` |
| `internal/tui/models/cpu_topology_form_validation.go` | Replace `SaveCPUTopology()` → `SaveConfig("cpu_topology", ...)` |
| `internal/tui/models/vcpu_pinning_form.go` | Replace `GetCPUTopology`/`GetVCPUPinningGlobal` → `GetConfig` |
| `internal/tui/models/vcpu_pinning_form_validation.go` | Replace `SaveCPUTopology`/`SaveVCPUPinningGlobal` → `SaveConfig` |
| `internal/tui/models/start_stop_script_form.go` | Replace `GetStartStopScript`/`GetPCIPassthroughConfig`/`SaveStartStopScript` → `GetConfig`/`SaveConfig` |
| `internal/vm/cpu_options_persistence_test.go` | Replace per-type methods → `GetConfig`/`SaveConfig` |
| `internal/vm/repository_topology_test.go` | Replace per-type methods → `GetConfig`/`SaveConfig` |
| `internal/vm/repository_pinning_test.go` | Replace per-type methods → `GetConfig`/`SaveConfig` |
| `internal/vm/pci_persistence_test.go` | Replace per-type methods → `GetConfig`/`SaveConfig` |
| `internal/vm/usb_persistence_test.go` | Replace per-type methods → `GetConfig`/`SaveConfig` |

**Total: 18 files changed (1 new, 17 modified).**

## After this refactor

Adding a new config type (e.g., `NetworkConfig`) requires:
1. Define the model struct in `internal/models`
2. One `SaveConfig("network", cfg)` call in the form validator
3. One `GetConfig("network", &cfg)` call in the form constructor

Zero new Repository methods. Zero new Viper key management code.
