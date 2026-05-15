// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
)

// SubViewModel is the interface all sub-view models must implement.
// The tea.Model methods provide BubbleTea lifecycle.
// SetSize forwards window resize events.
// FileBrowserActive tells MainModel whether ESC should pass through
// (e.g. when a file picker popup is open inside the form).
type SubViewModel interface {
	tea.Model
	SetSize(width, height int)
	FileBrowserActive() bool
}

// ViewFactory creates a SubViewModel. It receives the MainModel so
// constructors can access vmManager, configRepo, hostDiscovery, etc.
type ViewFactory func(m *MainModel) (SubViewModel, error)

// ViewDef registers a sub-view in the registry.
type ViewDef struct {
	Name            string              // e.g. ViewVMCreate
	Factory         ViewFactory         // constructor
	BreadcrumbLabel string              // e.g. "Add VM"
	ParentTab       components.Tab      // tab to return to on ESC
	ConfigMenuIndex int                 // position in Config tab, -1 if not in config menu
}

// ActiveView is the currently displayed sub-view instance.
type ActiveView struct {
	Def   *ViewDef
	Model SubViewModel
}

// ViewRegistry manages view definitions and the active sub-view.
type ViewRegistry struct {
	definitions     map[string]*ViewDef
	configMenuOrder []*ViewDef // ordered by ConfigMenuIndex, nil for gaps
	configMenuList  []*ViewDef // compacted config menu list (no nils), in display order
	activeView      *ActiveView
}

// NewViewRegistry creates an empty registry.
func NewViewRegistry() *ViewRegistry {
	return &ViewRegistry{
		definitions:     make(map[string]*ViewDef),
		configMenuOrder: nil,
	}
}

// Register adds a view definition. If ConfigMenuIndex >= 0, it's added
// to the config menu at that position. Returns error on duplicate name.
func (r *ViewRegistry) Register(def *ViewDef) error {
	if _, exists := r.definitions[def.Name]; exists {
		return fmt.Errorf("view %q already registered", def.Name)
	}
	r.definitions[def.Name] = def
	if def.ConfigMenuIndex >= 0 {
		r.insertConfigMenu(def)
	}
	return nil
}

// Activate creates a new sub-view instance and sets it as active.
func (r *ViewRegistry) Activate(name string, m *MainModel) (SubViewModel, error) {
	def, ok := r.definitions[name]
	if !ok {
		return nil, fmt.Errorf("unknown view: %s", name)
	}
	model, err := def.Factory(m)
	if err != nil {
		return nil, err
	}
	r.activeView = &ActiveView{Def: def, Model: model}
	return model, nil
}

// ActivateByConfigIndex activates the view at the given config menu position
// (index into configMenuOrder slice, which may have gaps).
func (r *ViewRegistry) ActivateByConfigIndex(index int, m *MainModel) (SubViewModel, error) {
	if index < 0 || index >= len(r.configMenuOrder) || r.configMenuOrder[index] == nil {
		return nil, fmt.Errorf("no view at config menu index %d", index)
	}
	return r.Activate(r.configMenuOrder[index].Name, m)
}

// ActivateByListIndex activates the view at the given list position,
// using the compacted config menu list (no gaps).
func (r *ViewRegistry) ActivateByListIndex(listIndex int, m *MainModel) (SubViewModel, error) {
	if listIndex < 0 || listIndex >= len(r.configMenuList) {
		return nil, fmt.Errorf("no view at list index %d", listIndex)
	}
	return r.Activate(r.configMenuList[listIndex].Name, m)
}

// Active returns the currently active view, or nil.
func (r *ViewRegistry) Active() *ActiveView {
	return r.activeView
}

// ActiveModel returns the currently active SubViewModel, or nil.
func (r *ViewRegistry) ActiveModel() SubViewModel {
	if r.activeView == nil {
		return nil
	}
	return r.activeView.Model
}

// Deactivate clears the active view.
func (r *ViewRegistry) Deactivate() {
	r.activeView = nil
}

// SetActiveModel replaces the model in the active view without going through
// the factory. Used when VMRunningModel is created externally and needs to
// be persisted across view transitions.
func (r *ViewRegistry) SetActiveModel(def *ViewDef, model SubViewModel) {
	r.activeView = &ActiveView{Def: def, Model: model}
}

// IsActive returns true if any sub-view is currently displayed.
func (r *ViewRegistry) IsActive() bool {
	return r.activeView != nil
}

// ActiveDef returns the ViewDef of the active view, or nil.
func (r *ViewRegistry) ActiveDef() *ViewDef {
	if r.activeView == nil {
		return nil
	}
	return r.activeView.Def
}

// ActiveName returns the name of the active view, or "".
func (r *ViewRegistry) ActiveName() string {
	if r.activeView == nil {
		return ""
	}
	return r.activeView.Def.Name
}

// GetDef returns a registered ViewDef by name, or nil.
func (r *ViewRegistry) GetDef(name string) *ViewDef {
	return r.definitions[name]
}

// ConfigMenuCount returns the number of views registered in the config menu.
func (r *ViewRegistry) ConfigMenuCount() int {
	return len(r.configMenuOrder)
}

// insertConfigMenu inserts a def into configMenuOrder at its ConfigMenuIndex
// and rebuilds the compacted configMenuList.
func (r *ViewRegistry) insertConfigMenu(def *ViewDef) {
	// Grow slice if needed
	for len(r.configMenuOrder) <= def.ConfigMenuIndex {
		r.configMenuOrder = append(r.configMenuOrder, nil)
	}
	r.configMenuOrder[def.ConfigMenuIndex] = def
	// Rebuild compacted list (no nils)
	r.configMenuList = r.configMenuList[:0]
	for _, d := range r.configMenuOrder {
		if d != nil {
			r.configMenuList = append(r.configMenuList, d)
		}
	}
}

// BuildConfigMenuItems returns the menu items for the Configuration tab
// from registered views. The last item is always "Save changes".
func (r *ViewRegistry) BuildConfigMenuItems() []MenuItem {
	items := make([]MenuItem, 0, len(r.configMenuList)+1)
	for _, def := range r.configMenuList {
		items = append(items, MenuItem{
			Title: def.BreadcrumbLabel,
			Type:  "INT_CONFIG",
		})
	}
	items = append(items, MenuItem{
		Title: "Save changes",
		Type:  "INT_CONFIG",
	})
	return items
}
