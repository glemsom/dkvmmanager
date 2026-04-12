// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/components"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
	"github.com/glemsom/dkvmmanager/internal/version"
)

func (m *MainModel) view() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	// 1. Header
	header := m.renderHeader()

	// 2. Tab bar
	tabBar := m.tabModel.RenderTabs(m.windowWidth)

	// 3. Breadcrumbs (only if non-empty)
	breadcrumbsView := ""
	if m.breadcrumbs.Len() > 0 {
		breadcrumbsView = m.breadcrumbs.Render()
	}

	// 4. Content area
	content := m.renderActiveContent()

	// 5. Status bar
	statusBarView := m.statusBar.Render(m.windowWidth)

	// Assemble
	var parts []string
	parts = append(parts, header)
	parts = append(parts, "")
	parts = append(parts, tabBar)
	if breadcrumbsView != "" {
		parts = append(parts, breadcrumbsView)
	}
	parts = append(parts, content)
	parts = append(parts, statusBarView)

	output := strings.Join(parts, "\n")

	// Modal backdrop dim: when a sub-view is active (except VMRunning which
	// has its own full display), apply faint styling to the entire output
	// to create a dialog overlay effect.
	if m.isSubViewActive() && m.currentView != ViewVMRunning {
		output = lipgloss.NewStyle().Faint(true).Render(output)
	}

	return output
}

func (m *MainModel) renderHeader() string {
	title := lipgloss.NewStyle().Foreground(styles.Colors.Primary).Render("DKVM Manager")
	versionStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted).
		Italic(true)
	ver := versionStyle.Render("v" + version.Version)
	gap := m.windowWidth - lipgloss.Width(title) - lipgloss.Width(ver)
	if gap < 1 {
		gap = 1
	}
	return title + strings.Repeat(" ", gap) + ver
}

func (m *MainModel) renderActiveContent() string {
	// Sub-views take over the content area
	switch m.currentView {
	case ViewVMCreate:
		if m.vmCreateModel != nil {
			return m.vmCreateModel.View()
		}
		return "Loading..."
	case ViewVMEdit:
		if m.vmEditModel != nil {
			return m.vmEditModel.View()
		}
		return "Loading..."
	case ViewVMDelete:
		if m.vmDeleteModel != nil {
			return m.vmDeleteModel.View()
		}
		return "Loading..."
	case ViewVMSelect:
		return m.renderVMSelectView()
	case ViewCPUOptions:
		if m.cpuOptionsModel != nil {
			return m.cpuOptionsModel.View()
		}
		return "Loading..."
	case ViewPCIPassthrough:
		if m.pciPassthroughModel != nil {
			return m.pciPassthroughModel.View()
		}
		return "Loading..."
	case ViewUSBPassthrough:
		if m.usbPassthroughModel != nil {
			return m.usbPassthroughModel.View()
		}
		return "Loading..."
	case ViewCPUTopology:
		if m.cpuTopologyModel != nil {
			return m.cpuTopologyModel.View()
		}
		return "Loading..."
	case ViewVCPUPinning:
		if m.vcpuPinningModel != nil {
			return m.vcpuPinningModel.View()
		}
		return "Loading..."
	case ViewSSHPassword:
		if m.sshPasswordModel != nil {
			return m.sshPasswordModel.View()
		}
		return "Loading..."
	case ViewStartStopScript:
		if m.startStopScriptFormModel != nil {
			return m.startStopScriptFormModel.View()
		}
		return "Loading..."
	case ViewVMRunning:
		if m.vmRunningModel != nil {
			return m.vmRunningModel.View()
		}
		return "Loading..."
	}

	// Tab-based content
	switch m.tabModel.GetActiveTab() {
	case components.TabVMs:
		return m.renderVMsTab()
	case components.TabConfiguration:
		return m.renderConfigTab()
	case components.TabPower:
		return m.renderPowerTab()
	default:
		return m.renderVMsTab()
	}
}

func (m *MainModel) renderVMsTab() string {
	if len(m.menuItems) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Colors.Muted).
			Italic(true)
		content := emptyStyle.Render("No virtual machines configured.") + "\n\n" +
			"Press 'n' or go to Configuration tab to create a new VM."

		// Pad to match list-based tabs. The config/power lists call SetSize with
		// (width-4, contentHeight()-2). The -2 accounts for list internal borders.
		// Count newlines manually since lipgloss.Height may not match.
		targetLines := m.contentHeight() - 2
		currentLines := strings.Count(content, "\n") + 1
		if currentLines < targetLines {
			padding := targetLines - currentLines
			content += strings.Repeat("\n", padding)
		}

		return content
	}

	if m.windowWidth < 80 {
		m.menuList.SetSize(m.windowWidth-4, m.contentHeight()-2)
		return m.menuList.View()
	}

	leftWidth := max(30, int(float64(m.windowWidth)*0.4))
	rightWidth := m.windowWidth - leftWidth - 3
	height := m.contentHeight() - 2

	m.menuList.SetSize(leftWidth, height)

	var selectedVM *models.VM
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.menuItems) {
		selectedVM = m.menuItems[m.selectedIndex].VMData
	}

	leftPanel := m.menuList.View()

	if selectedVM == nil {
		placeholderStyle := lipgloss.NewStyle().
			Foreground(styles.Colors.Muted).
			Italic(true).
			Width(rightWidth).
			Height(height)
		return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", placeholderStyle.Render("Select a VM"))
	}

	rightPanel := renderVMDetail(selectedVM, rightWidth, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
}

func renderVMDetail(vm *models.VM, width, height int) string {
	headerStyle := styles.DetailHeaderStyle()
	labelStyle := styles.DetailLabelStyle()
	valueStyle := styles.DetailValueStyle()
	sectionStyle := styles.DetailSectionStyle()
	panelStyle := styles.DetailPanelStyle().Width(width).Height(height)

	var b strings.Builder

	b.WriteString(headerStyle.Render(vm.Name))
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("ID: "))
	b.WriteString(valueStyle.Render(vm.ID))
	b.WriteString("\n\n")

	b.WriteString(sectionStyle.Render("Storage"))
	b.WriteString("\n")

	if len(vm.HardDisks) == 0 && len(vm.CDROMs) == 0 {
		b.WriteString(valueStyle.Render("  None"))
		b.WriteString("\n")
	} else {
		for _, disk := range vm.HardDisks {
			b.WriteString(valueStyle.Render("  ● " + disk))
			b.WriteString("\n")
		}
		for _, cd := range vm.CDROMs {
			b.WriteString(valueStyle.Render("  ◑ " + cd))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("Network"))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("  MAC: "))
	b.WriteString(valueStyle.Render(vm.MAC))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("  Mode: "))
	b.WriteString(valueStyle.Render(vm.NetworkMode))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("  VNC: "))
	if vm.VNCListen != "" {
		b.WriteString(valueStyle.Render("enabled (" + vm.VNCListen + ")"))
	} else {
		b.WriteString(valueStyle.Render("disabled"))
	}

	return panelStyle.Render(b.String())
}

func (m *MainModel) renderConfigTab() string {
	height := m.contentHeight() - 2
	m.configList.SetSize(m.windowWidth-4, height)
	content := m.configList.View()

	// Pad to match VM tab height (contentHeight, not contentHeight-2)
	targetLines := m.contentHeight()
	currentLines := strings.Count(content, "\n") + 1
	if currentLines < targetLines {
		padding := targetLines - currentLines
		content += strings.Repeat("\n", padding)
	}


	return content
}

func (m *MainModel) renderPowerTab() string {
	height := m.contentHeight() - 2
	m.powerList.SetSize(m.windowWidth-4, height)
	content := m.powerList.View()

	// Pad to match VM tab height (contentHeight, not contentHeight-2)
	targetLines := m.contentHeight()
	currentLines := strings.Count(content, "\n") + 1
	if currentLines < targetLines {
		padding := targetLines - currentLines
		content += strings.Repeat("\n", padding)
	}

	return content
}

// contentHeight calculates available height for content area
func (m *MainModel) contentHeight() int {
	// header: 1, blank: 1, tab bar: 1, separator: 1, status bar: 1 = 5 fixed lines
	// + breadcrumbs: 1 if present
	fixed := 5
	if m.breadcrumbs.Len() > 0 {
		fixed++
	}
	height := m.windowHeight - fixed
	if height < 3 {
		height = 3
	}
	return height
}
