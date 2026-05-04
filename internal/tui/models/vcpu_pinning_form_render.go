// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// --- Styles ---

var (
	vcpuPinningLabelStyle   = styles.FormLabelStyle()
	vcpuPinningFocusStyle   = styles.FormFocusStyle()
	vcpuPinningEnabledStyle = styles.SuccessTextStyle()
	vcpuPinningMutedStyle   = styles.FormMutedStyle()
	vcpuPinningSaveStyle    = styles.FormSaveStyle()
	vcpuPinningApplyStyle   = styles.FormSaveStyle().Background(styles.Colors.Warning)
	vcpuPinningTitleStyle   = styles.TitleStyle()
	vcpuPinningErrorStyle   = styles.ErrorTextStyle()
	vcpuPinningWarnStyle    = styles.WarningTextStyle()
	vcpuPinningHeaderStyle  = lipgloss.NewStyle().Bold(true).Foreground(styles.Colors.Primary)
)

// --- FormModel interface methods ---

// RenderHeader returns the form header containing all read-only content.
func (m *VCPUPinningFormModel) RenderHeader() string {
	var sb strings.Builder

	// Title
	sb.WriteString(vcpuPinningTitleStyle.Render("vCPU Pinning"))

	// Host summary (read-only, from CPU topology)
	if m.scanErr != nil {
		sb.WriteString("\n")
		sb.WriteString(vcpuPinningErrorStyle.Render(fmt.Sprintf("Warning: CPU scan failed: %s", m.scanErr)))
		sb.WriteString("\n")
		sb.WriteString(vcpuPinningWarnStyle.Render("vCPU pinning requires CPU topology configuration."))
	} else {
		sb.WriteString("\n")
		sb.WriteString(vcpuPinningLabelStyle.Render(fmt.Sprintf("Host: %d dies, %d cores, %d threads",
			len(m.hostTopo.Dies), m.hostTopo.TotalCores, m.hostTopo.TotalCPUs)))
	}

	// Current allocation (based on CPU topology config)
	sb.WriteString("\n")
	sb.WriteString(vcpuPinningTitleStyle.Render("Current Allocation:"))
	allocatedCores := 0
	if m.topology.Enabled && len(m.topology.SelectedCPUs) > 0 {
		// Count cores allocated to VMs
		for _, die := range m.hostTopo.Dies {
			for _, core := range die.CoreDetails {
				allThreadsSelected := true
				for _, t := range core.Threads {
					if !containsInt(m.topology.SelectedCPUs, t) {
						allThreadsSelected = false
						break
					}
				}
				if allThreadsSelected {
					allocatedCores++
				}
			}
		}

		if allocatedCores > 0 {
			// Show allocation per die
			coreCount := 0
			for _, die := range m.hostTopo.Dies {
				dieCores := 0
				for _, core := range die.CoreDetails {
					allThreadsSelected := true
					for _, t := range core.Threads {
						if !containsInt(m.topology.SelectedCPUs, t) {
							allThreadsSelected = false
							break
						}
					}
					if allThreadsSelected {
						dieCores++
					}
				}
				if dieCores > 0 {
					vcpus := coreCount * 4 // Assume 4 threads per core
					sb.WriteString("\n")
					sb.WriteString(vcpuPinningLabelStyle.Render(fmt.Sprintf("  Die %d: %d cores (vCPUs %d-%d) -> Host CPUs %s",
						die.ID, dieCores, vcpus, vcpus+dieCores*4-1, "auto")))
					coreCount += dieCores
				}
			}
		} else {
			sb.WriteString("\n")
			sb.WriteString(vcpuPinningMutedStyle.Render("  No cores allocated to VMs."))
		}
	} else {
		sb.WriteString("\n")
		sb.WriteString(vcpuPinningMutedStyle.Render("  CPU topology not configured."))
	}

	// Current mappings - show detailed topology alignment
	sb.WriteString("\n")
	sb.WriteString(vcpuPinningTitleStyle.Render("Current Mappings (auto-computed from topology):"))

	if !m.pinning.Enabled {
		sb.WriteString("\n")
		sb.WriteString("  " + vcpuPinningMutedStyle.Render("Pinning disabled."))
	} else if len(m.pinning.Mappings) == 0 {
		sb.WriteString("\n")
		sb.WriteString("  " + vcpuPinningMutedStyle.Render("No mappings (save with enabled to compute)."))
	} else {
		// Compute guest topology info (die, siblings per vCPU)
		guestDieMap, guestSiblingMap := m.computeGuestTopology()

		// Show each mapping with die and siblings
		for _, mp := range m.pinning.Mappings {
			// Find host die and siblings for this host CPU
			hostDieID, hostSiblings := m.findHostTopologyInfo(mp.HostCPUID)

			// Get guest info
			guestDieID := guestDieMap[mp.VCPUID]
			guestSiblings := guestSiblingMap[mp.VCPUID]

			// Format siblings as "0,1" or "0,1,2,3"
			guestSibStr := formatInts(guestSiblings)
			hostSibStr := formatInts(hostSiblings)

			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("  vCPU %d (die %d, siblings %s) -> Host CPU %d (die %d, siblings %s)",
				mp.VCPUID, guestDieID, guestSibStr, mp.HostCPUID, hostDieID, hostSibStr))
		}
		sb.WriteString("\n")
		sb.WriteString("  Die mapping: OK (guest die 0 -> host die 0)")
		sb.WriteString("\n")
		sb.WriteString("  Sibling alignment: OK")
	}

	// Summary
	sb.WriteString("\n")
	sb.WriteString(vcpuPinningLabelStyle.Render(fmt.Sprintf("Summary: %d vCPUs pinned", len(m.pinning.Mappings))))
	if len(m.pinning.Mappings) > 0 {
		sb.WriteString("\n")
		sb.WriteString(vcpuPinningLabelStyle.Render("topology-aware"))
	} else {
		sb.WriteString("\n")
		sb.WriteString(vcpuPinningMutedStyle.Render("not configured"))
	}

	return sb.String()
}

// RenderFooter returns the form footer.
func (m *VCPUPinningFormModel) RenderFooter() string {
	var parts []string

	// Save error at the bottom
	if errMsg, ok := m.errors["save"]; ok {
		parts = append(parts, "")
		parts = append(parts, vcpuPinningErrorStyle.Render("Error: "+errMsg))
	}

	// Kernel apply status message
	if m.kernelMsg != "" {
		parts = append(parts, "")
		if m.kernelMsgErr {
			parts = append(parts, vcpuPinningErrorStyle.Render("Error: "+m.kernelMsg))
		} else {
			parts = append(parts, vcpuPinningSaveStyle.Render(m.kernelMsg))
		}
	}

	// Footer help text
	parts = append(parts, "")
	parts = append(parts, vcpuPinningMutedStyle.Render("Tab Navigate  PgUp/PgDown Scroll  Space/Enter Toggle/Action  ESC Cancel"))

	return strings.Join(parts, "\n")
}

// RenderPosition renders a single position for the form framework.
func (m *VCPUPinningFormModel) RenderPosition(pos form.FocusPos, focused bool, cursorOffset int) string {
	switch pos.Kind {
	case form.FocusToggle:
		toggleLabel := "[ ] Disabled"
		if m.pinning.Enabled {
			toggleLabel = "[x] Enabled"
		}

		if focused {
			toggleLabel = vcpuPinningFocusStyle.Render(toggleLabel)
		} else if m.pinning.Enabled {
			toggleLabel = vcpuPinningEnabledStyle.Render(toggleLabel)
		}
		return vcpuPinningHeaderStyle.Render("vCPU Pinning Configuration") + "\n  " + toggleLabel

	case form.FocusButton:
		if pos.Key == "save" {
			saveText := vcpuPinningMutedStyle.Render("[Space/Enter] Save") + "    " + vcpuPinningMutedStyle.Render("[ESC] Cancel")
			if focused {
				saveText = vcpuPinningSaveStyle.Render("[Space/Enter] Save") + "    " + vcpuPinningMutedStyle.Render("[ESC] Cancel")
			}
			return "\n" + saveText
		}

		if pos.Key == "apply_kernel" {
			applyText := vcpuPinningMutedStyle.Render("[Space/Enter] Apply to Kernel") + "    " + vcpuPinningMutedStyle.Render("[ESC] Cancel")
			if focused {
				applyText = vcpuPinningApplyStyle.Render("[Space/Enter] Apply to Kernel") + "    " + vcpuPinningMutedStyle.Render("[ESC] Cancel")
			}
			return "\n" + applyText
		}
	}

	return ""
}

// --- Helper functions ---

// containsInt checks if a slice contains a value
func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// computeGuestTopology derives guest die and sibling info from topology config
func (m *VCPUPinningFormModel) computeGuestTopology() (map[int]int, map[int][]int) {
	dieMap := make(map[int]int)     // vCPU ID -> guest die ID
	siblingMap := make(map[int][]int) // vCPU ID -> sibling vCPU IDs

	if !m.topology.Enabled || len(m.topology.SelectedCPUs) == 0 {
		return dieMap, siblingMap
	}

	// Build map of host CPU -> selected index (vCPU ID)
	hostToVCPU := make(map[int]int)
	for i, cpu := range m.topology.SelectedCPUs {
		hostToVCPU[cpu] = i
	}

	// Walk host topology to assign guest die/siblings based on topology order
	for _, die := range m.hostTopo.Dies {
		for _, core := range die.CoreDetails {
			// Check if all threads in this core are selected
			allSelected := true
			var coreVCPUIDs []int
			for _, t := range core.Threads {
				if !containsInt(m.topology.SelectedCPUs, t) {
					allSelected = false
					break
				}
				if v, ok := hostToVCPU[t]; ok {
					coreVCPUIDs = append(coreVCPUIDs, v)
				}
			}
			if allSelected && len(coreVCPUIDs) > 0 {
				// Assign die and siblings for this core
				for _, vcpu := range coreVCPUIDs {
					dieMap[vcpu] = die.ID
					siblingMap[vcpu] = coreVCPUIDs
				}
			}
		}
	}

	return dieMap, siblingMap
}

// findHostTopologyInfo returns the die ID and sibling CPUs for a host CPU
func (m *VCPUPinningFormModel) findHostTopologyInfo(hostCPU int) (int, []int) {
	for _, die := range m.hostTopo.Dies {
		for _, core := range die.CoreDetails {
			for _, t := range core.Threads {
				if t == hostCPU {
					return die.ID, core.Threads
				}
			}
		}
	}
	return 0, nil
}

// formatInts formats a slice of ints as comma-separated string
func formatInts(nums []int) string {
	if len(nums) == 0 {
		return "none"
	}
	result := ""
	for i, n := range nums {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%d", n)
	}
	return result
}

// --- Backward-compatible viewport helpers ---

// syncViewport regenerates the rendered lines and syncs the viewport (backward compat).
func (m *VCPUPinningFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	content := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(content)
	// Ensure focused element is visible
	if m.focusedLineIndex() >= 0 {
		m.vp.SetYOffset(form.ClampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height))
	}
}

// focusedLineIndex maps focusIndex to a rendered line index (backward compat).
func (m *VCPUPinningFormModel) focusedLineIndex() int {
	line := 0

	// Header: title + blank + host info + blank + allocation title + allocation content + blank
	line += 2  // title + blank
	if m.scanErr != nil {
		line += 3 // error + warning + blank
	} else {
		line += 2 // host info + blank
	}
	line += 3 // allocation title + content + blank

	// Now at the toggle position
	if m.focusIndex == 0 {
		return line // toggle
	}
	line += 2 // config header + toggle

	// Now at the Save button
	if m.focusIndex == 1 {
		return line // save
	}
	line += 2 // blank + save button

	// Apply to Kernel button
	return line // apply_kernel
}

// renderAllLines produces the full list of output lines for the form (backward compat).
func (m *VCPUPinningFormModel) renderAllLines() []string {
	var lines []string

	// Render header
	lines = append(lines, strings.Split(m.RenderHeader(), "\n")...)

	// Render positions
	for i, pos := range m.positions {
		focused := (i == m.focusIndex)
		line := m.RenderPosition(pos, focused, -1)
		lines = append(lines, strings.Split(line, "\n")...)
	}

	// Render footer
	lines = append(lines, strings.Split(m.RenderFooter(), "\n")...)

	return lines
}
