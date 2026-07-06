// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"log"
	"sort"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/glemsom/dkvmmanager/internal/domain"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VMSelectModel implements SubViewModel for selecting a VM from a list.
type VMSelectModel struct {
	vmManager *vm.Manager
	vms       []domain.VM
	vmList    list.Model
	mode      string // "edit" or "delete"
	debugMode bool
	width     int
	height    int
}

// NewVMSelectModel creates a new VMSelectModel.
func NewVMSelectModel(mgr *vm.Manager, vms []domain.VM, mode string, debugMode bool) *VMSelectModel {
	// Sort VMs by ID to ensure deterministic ordering
	sorted := make([]domain.VM, len(vms))
	copy(sorted, vms)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})

	// Build list adapter
	vmListAdapter := buildVMListAdapter(sorted)
	vmDelegate := VMListItemDelegate{}
	vmList := list.New(vmListAdapter, vmDelegate, 80, 20)
	vmList.SetShowTitle(false)
	vmList.SetShowStatusBar(false)
	vmList.SetFilteringEnabled(false)
	vmList.SetShowHelp(false)

	return &VMSelectModel{
		vmManager: mgr,
		vms:       sorted,
		vmList:    vmList,
		mode:      mode,
		debugMode: debugMode,
		width:     80,
		height:    20,
	}
}

// Init implements tea.Model.
func (m *VMSelectModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
// When user presses Enter/Space on a VM, it transitions to the next view
// by returning a VMSelectedMsg.
func (m *VMSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter", "space":
			selectedIndex := m.vmList.Index()
			if selectedIndex >= 0 && selectedIndex < len(m.vms) {
				selectedVM := m.vms[selectedIndex]
				if m.debugMode {
					log.Printf("[DEBUG] VMSelectModel: selected VM %s (ID: %s) for %s", selectedVM.Name, selectedVM.ID, m.mode)
				}
				return m, func() tea.Msg {
					return VMSelectedMsg{
						VMID: selectedVM.ID,
						Mode: m.mode,
					}
				}
			}
			return m, nil
		case "esc":
			return m, func() tea.Msg {
				return ViewChangeMsg{View: ViewConfigMenu}
			}
		}
	}

	var cmd tea.Cmd
	m.vmList, cmd = m.vmList.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m *VMSelectModel) View() tea.View {
	m.vmList.SetSize(m.width-4, m.height-2)
	output := m.vmList.View()
	output += "\n\n\u2191/\u2193 Navigate  Space/Enter Select  ESC Cancel\n"
	return tea.NewView(output)
}

// SetSize implements SubViewModel.
func (m *VMSelectModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// FileBrowserActive implements SubViewModel.
func (m *VMSelectModel) FileBrowserActive() bool {
	return false
}

// VMSelectedMsg is sent when a VM is selected from VMSelectModel.
type VMSelectedMsg struct {
	VMID string
	Mode string // "edit" or "delete"
}
