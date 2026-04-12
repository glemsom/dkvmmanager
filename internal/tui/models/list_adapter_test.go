package models

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/glemsom/dkvmmanager/internal/models"
)

func TestMenuItemAdapterFilterValue(t *testing.T) {
	item := MenuItemAdapter{
		MenuItem: MenuItem{
			Title: "Test VM",
			Type:  "VM",
			VMID:  "vm-001",
		},
	}

	if got := item.FilterValue(); got != "Test VM" {
		t.Errorf("FilterValue() = %q, want %q", got, "Test VM")
	}
}

func TestMenuItemAdapterTitle(t *testing.T) {
	item := MenuItemAdapter{
		MenuItem: MenuItem{
			Title: "My Virtual Machine",
			Type:  "VM",
			VMID:  "vm-002",
		},
	}

	if got := item.Title(); got != "My Virtual Machine" {
		t.Errorf("Title() = %q, want %q", got, "My Virtual Machine")
	}
}

func TestMenuItemAdapterDescription(t *testing.T) {
	tests := []struct {
		name string
		typ  string
		want string
	}{
		{"VM type", "VM", "Virtual Machine"},
		{"Config type", "INT_CONFIG", "Configuration"},
		{"Power type", "INT_POWEROFF", "Power management"},
		{"Shell type", "INT_SHELL", "Shell access"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := MenuItemAdapter{
				MenuItem: MenuItem{
					Title: "Item",
					Type:  tt.typ,
				},
			}

			if got := item.Description(); got != tt.want {
				t.Errorf("Description() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMenuItemAdapterImplementsListItem(t *testing.T) {
	item := MenuItemAdapter{
		MenuItem: MenuItem{
			Title: "Test",
			Type:  "VM",
		},
	}

	// Compile-time check that MenuItemAdapter implements list.Item
	var _ list.Item = item
}

func TestMenuItemDelegateHeight(t *testing.T) {
	d := MenuItemDelegate{}
	if got := d.Height(); got != 1 {
		t.Errorf("Height() = %d, want 1", got)
	}
}

func TestMenuItemDelegateSpacing(t *testing.T) {
	d := MenuItemDelegate{}
	if got := d.Spacing(); got != 0 {
		t.Errorf("Spacing() = %d, want 0", got)
	}
}

func TestMenuItemDelegateUpdate(t *testing.T) {
	d := MenuItemDelegate{}
	cmd := d.Update(nil, nil)
	if cmd != nil {
		t.Error("Update() should return nil cmd")
	}
}

func TestMenuItemDelegateImplementsItemDelegate(t *testing.T) {
	// Compile-time check that MenuItemDelegate implements list.ItemDelegate
	var _ list.ItemDelegate = MenuItemDelegate{}
}

func TestMenuItemDelegateRenderSelected(t *testing.T) {
	d := MenuItemDelegate{}
	item := MenuItemAdapter{
		MenuItem: MenuItem{Title: "Test VM", Type: "VM"},
	}

	// Create a list model with this item selected (index 0)
	items := []list.Item{item}
	m := list.New(items, d, 10, 5)

	var buf bytes.Buffer
	d.Render(&buf, m, 0, item)

	output := buf.String()
	if output == "" {
		t.Fatal("Render() produced empty output for selected item")
	}

	// Selected item should contain ">  " prefix (2 spaces for alignment with unselected items)
	if !bytes.Contains(buf.Bytes(), []byte(">  Test VM")) {
		t.Errorf("Render() selected item should contain '>  Test VM', got %q", output)
	}
}

func TestMenuItemDelegateRenderUnselected(t *testing.T) {
	d := MenuItemDelegate{}
	item1 := MenuItemAdapter{
		MenuItem: MenuItem{Title: "Item 1", Type: "VM"},
	}
	item2 := MenuItemAdapter{
		MenuItem: MenuItem{Title: "Item 2", Type: "VM"},
	}

	items := []list.Item{item1, item2}
	m := list.New(items, d, 10, 5)

	var buf bytes.Buffer
	d.Render(&buf, m, 1, item2)

	output := buf.String()
	if output == "" {
		t.Fatal("Render() produced empty output for unselected item")
	}

	// Unselected item should contain "  " prefix (not "> ")
	if !bytes.Contains(buf.Bytes(), []byte("  Item 2")) {
		t.Errorf("Render() unselected item should contain '  Item 2', got %q", output)
	}
}

func TestMenuItemDelegateRenderDisabled(t *testing.T) {
	d := MenuItemDelegate{}
	item := MenuItemAdapter{
		MenuItem: MenuItem{Title: "Disabled Item", Type: "VM", Disabled: true},
	}

	items := []list.Item{item}
	m := list.New(items, d, 10, 5)

	var buf bytes.Buffer
	d.Render(&buf, m, 0, item)

	output := buf.String()
	if output == "" {
		t.Fatal("Render() produced empty output for disabled item")
	}

	// Disabled item should still show the title
	if !bytes.Contains(buf.Bytes(), []byte("Disabled Item")) {
		t.Errorf("Render() disabled item should contain title, got %q", output)
	}

	// Disabled selected item should NOT have "> " prefix (disabled takes precedence)
	if bytes.Contains(buf.Bytes(), []byte("> Disabled Item")) {
		t.Errorf("Render() disabled item should not have '> ' prefix, got %q", output)
	}
}

func TestMenuItemDelegateRenderInvalidType(t *testing.T) {
	d := MenuItemDelegate{}
	var buf bytes.Buffer

	// Should not panic and should produce no output with nil item
	d.Render(&buf, list.Model{}, 0, nil)

	if buf.Len() != 0 {
		t.Errorf("Render() with nil item should produce no output, got %q", buf.String())
	}
}

func TestVMListAdapterFilterValue(t *testing.T) {
	item := VMListAdapter{
		VM: models.VM{
			ID:   "vm-001",
			Name: "Test VM",
		},
	}

	if got := item.FilterValue(); got != "Test VM" {
		t.Errorf("FilterValue() = %q, want %q", got, "Test VM")
	}
}

func TestVMListAdapterTitle(t *testing.T) {
	item := VMListAdapter{
		VM: models.VM{
			ID:   "vm-002",
			Name: "My Virtual Machine",
		},
	}

	if got := item.Title(); got != "My Virtual Machine" {
		t.Errorf("Title() = %q, want %q", got, "My Virtual Machine")
	}
}

func TestVMListAdapterDescription(t *testing.T) {
	tests := []struct {
		name string
		vm   models.VM
		want string
	}{
		{
			name: "VM with harddisks",
			vm: models.VM{
				ID:        "vm-001",
				Name:      "Test",
				HardDisks: []string{"/dev/sda", "/dev/sdb"},
			},
			want: "ID: vm-001  Disks: 2",
		},
		{
			name: "VM with cdroms",
			vm: models.VM{
				ID:     "vm-002",
				Name:   "Test",
				CDROMs: []string{"/isos/ubuntu.iso"},
			},
			want: "ID: vm-002  Disks: 1",
		},
		{
			name: "VM with both",
			vm: models.VM{
				ID:        "vm-003",
				Name:      "Test",
				HardDisks: []string{"/dev/sda"},
				CDROMs:    []string{"/isos/win.iso"},
			},
			want: "ID: vm-003  Disks: 2",
		},
		{
			name: "VM with no disks",
			vm: models.VM{
				ID:   "vm-004",
				Name: "Test",
			},
			want: "ID: vm-004  Disks: 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := VMListAdapter{VM: tt.vm}
			if got := item.Description(); got != tt.want {
				t.Errorf("Description() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVMListAdapterImplementsListItem(t *testing.T) {
	item := VMListAdapter{
		VM: models.VM{ID: "vm-001", Name: "Test"},
	}

	// Compile-time check that VMListAdapter implements list.Item
	var _ list.Item = item
}

func TestVMListItemDelegateHeight(t *testing.T) {
	d := VMListItemDelegate{}
	if got := d.Height(); got != 1 {
		t.Errorf("Height() = %d, want 1", got)
	}
}

func TestVMListItemDelegateSpacing(t *testing.T) {
	d := VMListItemDelegate{}
	if got := d.Spacing(); got != 0 {
		t.Errorf("Spacing() = %d, want 0", got)
	}
}

func TestVMListItemDelegateUpdate(t *testing.T) {
	d := VMListItemDelegate{}
	cmd := d.Update(nil, nil)
	if cmd != nil {
		t.Error("Update() should return nil cmd")
	}
}

func TestVMListItemDelegateImplementsItemDelegate(t *testing.T) {
	// Compile-time check that VMListItemDelegate implements list.ItemDelegate
	var _ list.ItemDelegate = VMListItemDelegate{}
}

func TestVMListItemDelegateRenderSelected(t *testing.T) {
	d := VMListItemDelegate{}
	item := VMListAdapter{
		VM: models.VM{ID: "vm-001", Name: "Test VM"},
	}

	items := []list.Item{item}
	m := list.New(items, d, 10, 5)

	var buf bytes.Buffer
	d.Render(&buf, m, 0, item)

	output := buf.String()
	if output == "" {
		t.Fatal("Render() produced empty output for selected item")
	}

	if !bytes.Contains(buf.Bytes(), []byte(">  Test VM")) {
		t.Errorf("Render() selected item should contain '>  Test VM', got %q", output)
	}
}

func TestVMListItemDelegateRenderUnselected(t *testing.T) {
	d := VMListItemDelegate{}
	item1 := VMListAdapter{
		VM: models.VM{ID: "vm-001", Name: "Item 1"},
	}
	item2 := VMListAdapter{
		VM: models.VM{ID: "vm-002", Name: "Item 2"},
	}

	items := []list.Item{item1, item2}
	m := list.New(items, d, 10, 5)

	var buf bytes.Buffer
	d.Render(&buf, m, 1, item2)

	output := buf.String()
	if output == "" {
		t.Fatal("Render() produced empty output for unselected item")
	}

	if !bytes.Contains(buf.Bytes(), []byte("  Item 2")) {
		t.Errorf("Render() unselected item should contain '  Item 2', got %q", output)
	}
}

func TestVMListItemDelegateRenderInvalidType(t *testing.T) {
	d := VMListItemDelegate{}
	var buf bytes.Buffer

	d.Render(&buf, list.Model{}, 0, nil)

	if buf.Len() != 0 {
		t.Errorf("Render() with nil item should produce no output, got %q", buf.String())
	}
}

func TestBuildVMListAdapter(t *testing.T) {
	vms := []models.VM{
		{ID: "vm-001", Name: "Alpha"},
		{ID: "vm-002", Name: "Beta"},
		{ID: "vm-003", Name: "Gamma"},
	}

	items := buildVMListAdapter(vms)

	if len(items) != 3 {
		t.Fatalf("buildVMListAdapter() returned %d items, want 3", len(items))
	}

	for i, item := range items {
		vmAdapter, ok := item.(VMListAdapter)
		if !ok {
			t.Errorf("items[%d] is not VMListAdapter", i)
			continue
		}
		if vmAdapter.VM.Name != vms[i].Name {
			t.Errorf("items[%d].Name = %q, want %q", i, vmAdapter.VM.Name, vms[i].Name)
		}
	}
}
