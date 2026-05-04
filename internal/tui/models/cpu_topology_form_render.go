// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"

	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/models/form"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// --- Styles ---

var (
	cpuTopoLabelStyle    = styles.FormLabelStyle()
	cpuTopoFocusStyle    = styles.FormFocusStyle()
	cpuTopoSelectedStyle = styles.SuccessTextStyle()
	cpuTopoMutedStyle    = styles.FormMutedStyle()
	cpuTopoSaveStyle     = styles.FormSaveStyle()
	cpuTopoDieStyle      = styles.TitleStyle()
	cpuTopoCacheStyle    = styles.WarningTextStyle()
	cpuTopoErrorStyle    = styles.ErrorTextStyle()
	cpuTopoWarnStyle     = styles.WarningTextStyle()
	cpuTopoHostStyle     = styles.WarningTextStyle()
	cpuTopoThreadStyle   = styles.MutedTextStyle()
)

// --- Backward-compatible rendering helpers ---

// renderAllLines produces the full list of output lines for the viewport (backward compat).
func (m *CPUTopologyFormModel) renderAllLines() []string {
	var lines []string

	if m.scanErr != nil {
		lines = append(lines, cpuTopoFocusStyle.Render("CPU Topology"))
		lines = append(lines, "")
		lines = append(lines, cpuTopoErrorStyle.Render(fmt.Sprintf("Warning: CPU scan failed: %s", m.scanErr)))
		lines = append(lines, cpuTopoWarnStyle.Render("Topology configuration unavailable."))
		lines = append(lines, "")

		saveText := cpuTopoMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
		if len(m.positions) > 0 && m.positions[m.focusIndex].Kind == form.FocusButton {
			saveText = cpuTopoSaveStyle.Render("[Space/Enter] Save") + "    " + cpuTopoMutedStyle.Render("[ESC] Cancel")
		}
		lines = append(lines, saveText)
		return lines
	}

	if len(m.hostTopo.Dies) == 0 {
		lines = append(lines, cpuTopoFocusStyle.Render("CPU Topology"))
		lines = append(lines, "")
		lines = append(lines, cpuTopoMutedStyle.Render("No CPU topology data available."))
		lines = append(lines, "")
		saveText := cpuTopoMutedStyle.Render("[Space/Enter] Save    [ESC] Cancel")
		if len(m.positions) > 0 && m.positions[m.focusIndex].Kind == form.FocusButton {
			saveText = cpuTopoSaveStyle.Render("[Enter] Save") + "    " + cpuTopoMutedStyle.Render("[ESC] Cancel")
		}
		lines = append(lines, saveText)
		return lines
	}

	// Header
	lines = append(lines, cpuTopoFocusStyle.Render("CPU Topology"))
	lines = append(lines, "")
	lines = append(lines, cpuTopoLabelStyle.Render(fmt.Sprintf("Host: %d dies, %d cores, %d threads",
		len(m.hostTopo.Dies), m.hostTopo.TotalCores, m.hostTopo.TotalCPUs)))
	lines = append(lines, "")

	// Count allocated cores
	allocatedCores := 0
	for _, pos := range m.positions {
		if pos.Kind == form.FocusToggle {
			d := pos.Data.(cpuTopoFocusData)
			key := coreKey(d.dieID, d.coreID)
			if m.coreSelected[key] {
				allocatedCores++
			}
		}
	}

	// Render grouped by die
	lastDieID := -1
	for i, pos := range m.positions {
		focused := (i == m.focusIndex)

		switch pos.Kind {
		case form.FocusToggle:
			d := pos.Data.(cpuTopoFocusData)
			if d.dieID != lastDieID {
				if lastDieID != -1 {
					lines = append(lines, "")
				}
				lines = append(lines, cpuTopoDieStyle.Render(d.dieLabel))
				lastDieID = d.dieID
			}
			key := coreKey(d.dieID, d.coreID)
			lines = append(lines, m.renderCoreLine(d.coreInfo, m.coreSelected[key], focused))

		case form.FocusButton:
			lines = append(lines, "")
			hostCores := m.hostTopo.TotalCores - allocatedCores
			lines = append(lines, cpuTopoLabelStyle.Render(fmt.Sprintf(
				"Summary: %d cores for VMs, %d for host",
				allocatedCores, hostCores)))
			if hostCores == 0 {
				lines = append(lines, cpuTopoErrorStyle.Render("Warning: No cores reserved for host — system may become unresponsive"))
			}
			lines = append(lines, "")
			saveText := cpuTopoMutedStyle.Render("[Enter] Save    [ESC] Cancel")
			if focused {
				saveText = cpuTopoSaveStyle.Render("[Space/Enter] Save") + "    " + cpuTopoMutedStyle.Render("[ESC] Cancel")
			}
			lines = append(lines, saveText)
		}
	}

	if errMsg, ok := m.errors["save"]; ok {
		lines = append(lines, "")
		lines = append(lines, cpuTopoErrorStyle.Render("Error: "+errMsg))
	}

	lines = append(lines, "")
	lines = append(lines, cpuTopoMutedStyle.Render("Tab Navigate  PgUp/PgDown Scroll  Space Toggle  ESC Cancel"))

	return lines
}

// renderCoreLine renders a physical core with its thread info
func (m *CPUTopologyFormModel) renderCoreLine(core *models.CPUCore, selected, focused bool) string {
	prefix := "  "
	if focused {
		prefix = cpuTopoFocusStyle.Render("> ")
	}

	var togglePart string
	if selected {
		if focused {
			togglePart = cpuTopoFocusStyle.Render("[ VM ]")
		} else {
			togglePart = cpuTopoSelectedStyle.Render("[ VM ]")
		}
	} else {
		if focused {
			togglePart = cpuTopoFocusStyle.Render("[HOST]")
		} else {
			togglePart = cpuTopoHostStyle.Render("[HOST]")
		}
	}

	threadLabel := ""
	if core != nil && len(core.Threads) > 0 {
		threadIDs := make([]string, len(core.Threads))
		for i, t := range core.Threads {
			threadIDs[i] = fmt.Sprintf("%d", t)
		}
		threadLabel = cpuTopoThreadStyle.Render(
			fmt.Sprintf("  [%d threads: %s]", len(core.Threads), strings.Join(threadIDs, ",")))
	}

	coreLabel := cpuTopoLabelStyle.Render(fmt.Sprintf("Core %d", core.ID))

	return prefix + togglePart + " " + coreLabel + threadLabel
}
