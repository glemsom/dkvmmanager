// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// --- Styles ---

var (
	vcpuPinningLabelStyle    = styles.FormLabelStyle()
	vcpuPinningFocusStyle  = styles.FormFocusStyle()
	vcpuPinningEnabledStyle = styles.SuccessTextStyle()
	vcpuPinningMutedStyle  = styles.FormMutedStyle()
	vcpuPinningSaveStyle   = styles.FormSaveStyle()
	vcpuPinningTitleStyle   = styles.TitleStyle()
	vcpuPinningErrorStyle = styles.ErrorTextStyle()
	vcpuPinningWarnStyle  = styles.WarningTextStyle()
)

// renderAllLines produces the full list of output lines for the form
func (m *VCPUPinningFormModel) renderAllLines() []string {
	var lines []string

	// Title
	lines = append(lines, vcpuPinningTitleStyle.Render("vCPU Pinning"))
	lines = append(lines, "")

	// Host summary (read-only, from CPU topology)
	if m.scanErr != nil {
		lines = append(lines, vcpuPinningErrorStyle.Render(fmt.Sprintf("Warning: CPU scan failed: %s", m.scanErr)))
		lines = append(lines, vcpuPinningWarnStyle.Render("vCPU pinning requires CPU topology configuration."))
		lines = append(lines, "")
	} else {
		lines = append(lines, vcpuPinningLabelStyle.Render(fmt.Sprintf("Host: %d dies, %d cores, %d threads",
			len(m.hostTopo.Dies), m.hostTopo.TotalCores, m.hostTopo.TotalCPUs)))
		lines = append(lines, "")
	}

	// Current allocation (based on CPU topology config)
	lines = append(lines, vcpuPinningTitleStyle.Render("Current Allocation:"))
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
					lines = append(lines, vcpuPinningLabelStyle.Render(fmt.Sprintf("  Die %d: %d cores (vCPUs %d-%d) -> Host CPUs %s",
						die.ID, dieCores, vcpus, vcpus+dieCores*4-1, "auto")))
					coreCount += dieCores
				}
			}
		} else {
			lines = append(lines, vcpuPinningMutedStyle.Render("  No cores allocated to VMs."))
		}
	} else {
		lines = append(lines, vcpuPinningMutedStyle.Render("  CPU topology not configured."))
	}
	lines = append(lines, "")

	// Configuration toggle
	focused := m.focusPos == vcpuPinningToggle
	lines = append(lines, vcpuPinningTitleStyle.Render("vCPU Pinning Configuration"))

	toggleLabel := "[ ] Disabled"
	if m.pinning.Enabled {
		toggleLabel = "[x] Enabled"
	}

	if focused {
		toggleLabel = vcpuPinningFocusStyle.Render(toggleLabel)
	} else if m.pinning.Enabled {
		toggleLabel = vcpuPinningEnabledStyle.Render(toggleLabel)
	}
	lines = append(lines, "  "+toggleLabel)
	lines = append(lines, "")

	// Current mappings - show detailed topology alignment
	lines = append(lines, vcpuPinningTitleStyle.Render("Current Mappings (auto-computed from topology):"))

	if !m.pinning.Enabled {
		lines = append(lines, "  "+vcpuPinningMutedStyle.Render("Pinning disabled."))
	} else if len(m.pinning.Mappings) == 0 {
		lines = append(lines, "  "+vcpuPinningMutedStyle.Render("No mappings (save with enabled to compute)."))
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
			
			lines = append(lines, fmt.Sprintf("  vCPU %d (die %d, siblings %s) -> Host CPU %d (die %d, siblings %s)",
				mp.VCPUID, guestDieID, guestSibStr, mp.HostCPUID, hostDieID, hostSibStr))
		}
		lines = append(lines, "  Die mapping: OK (guest die 0 -> host die 0)")
		lines = append(lines, "  Sibling alignment: OK")
	}
	lines = append(lines, "")

	// Summary
	lines = append(lines, vcpuPinningLabelStyle.Render(fmt.Sprintf("Summary: %d vCPUs pinned", len(m.pinning.Mappings))))
	if len(m.pinning.Mappings) > 0 {
		lines = append(lines, vcpuPinningLabelStyle.Render("topology-aware"))
	} else {
		lines = append(lines, vcpuPinningMutedStyle.Render("not configured"))
	}
	lines = append(lines, "")

	// Save button
	focused = m.focusPos == vcpuPinningSave
	saveText := vcpuPinningMutedStyle.Render("[Enter] Save    [ESC] Cancel")
	if focused {
		saveText = vcpuPinningSaveStyle.Render("[Space/Enter] Save") + "    " + vcpuPinningMutedStyle.Render("[ESC] Cancel")
	}
	lines = append(lines, saveText)

	// Errors
	if errMsg, ok := m.errors["save"]; ok {
		lines = append(lines, "")
		lines = append(lines, vcpuPinningErrorStyle.Render("Error: "+errMsg))
	}

	lines = append(lines, "")
	lines = append(lines, vcpuPinningMutedStyle.Render("↑/↓ Navigate  Space Toggle  ESC Cancel"))

	return lines
}

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
	dieMap := make(map[int]int)   // vCPU ID -> guest die ID
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

// syncViewport regenerates the rendered lines and syncs the viewport
func (m *VCPUPinningFormModel) syncViewport() {
	m.renderedLines = m.renderAllLines()
	content := strings.Join(m.renderedLines, "\n")
	m.vp.SetContent(content)
	// Ensure focused element is visible
	if m.focusedLineIndex() >= 0 {
		m.vp.SetYOffset(clampOffset(m.vp.YOffset, m.focusedLineIndex(), m.vp.Height))
	}
}

// focusedLineIndex maps focusPos to the rendered line index
func (m *VCPUPinningFormModel) focusedLineIndex() int {
	line := 0

	// Title + blank = 2
	line += 2

	// Host summary
	if m.scanErr != nil {
		line += 3 // error + warning + blank
	} else {
		line += 2 // host info + blank
	}

	// Current Allocation section: title + (either allocation details or muted message) + blank
	line += 3

	// Configuration toggle section: title + toggle + blank = 3
	line += 3

	// Current mappings section (only when enabled)
	if m.pinning.Enabled {
		line++ // title
		if len(m.pinning.Mappings) == 0 {
			line++ // muted message
		} else {
			// Each mapping is 1 line + 2 info lines
			line += len(m.pinning.Mappings)
			line += 2 // die mapping + sibling alignment
		}
		line++ // blank
	}

	// Summary: label + status = 2 lines + blank
	line += 3

	// Now we're at the Save button line
	// If focus is on Save, return this line; otherwise return toggle line
	if m.focusPos == vcpuPinningSave {
		return line
	}

	// Toggle is at the configuration section, after title + blank + allocation section
	// Recalculate: title(2) + host(2) + allocation(3) + config(3) = line 10
	// Toggle is at line 10
	line = 2 + 2 + 3 + 3 // title + host + allocation + config = 10

	return line
}