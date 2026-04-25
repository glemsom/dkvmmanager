// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VMStartedMsg is sent when a VM starts successfully
type VMStartedMsg struct {
	VMName string
	VMID   string
}

// VMStoppedMsg is sent when a VM stops
type VMStoppedMsg struct {
	VMName string
	VMID   string
	Reason string
}

// VMLogMsg is sent when a new log line is available from the VM
type VMLogMsg struct {
	Line string
}

// VMStatusUpdateMsg is sent periodically to refresh VM status
type VMStatusUpdateMsg struct {
	Status  string
	Threads []int
}

// VMRunningModel displays a running VM with log output and status
type VMRunningModel struct {
	// VM info
	vm     *models.VM
	runner *vm.VMRunner

	// Log viewport
	vp    viewport.Model
	ready bool

	// Log lines accumulated
	logLines    []string
	maxLogLines int

	// Status
	status    string
	threads   []int
	startTime time.Time

	// Dimensions
	width  int
	height int

	// Info panel height (calculated)
	infoHeight int
}

// NewVMRunningModel creates a new VM running model
func NewVMRunningModel(vmObj *models.VM, runner *vm.VMRunner) *VMRunningModel {
	return &VMRunningModel{
		vm:          vmObj,
		runner:      runner,
		maxLogLines: 500,
	}
}

// Init initializes the model
func (m *VMRunningModel) Init() tea.Cmd {
	return tea.Batch(
		m.waitForLog(),
		m.waitForVMExit(),
		m.pollStatus(),      // periodic (500ms)
		m.initialStatus(),   // immediate
	)
}

// initialStatus performs an immediate status check and returns a command
func (m *VMRunningModel) initialStatus() tea.Cmd {
	return func() tea.Msg {
		if m.runner == nil {
			return VMStatusUpdateMsg{Status: "starting"}
		}
		client := m.runner.QMPClient()
		if client == nil {
			return VMStatusUpdateMsg{Status: "starting"}
		}
		status, err := client.QueryStatus()
		if err != nil {
			return VMStatusUpdateMsg{Status: "unknown"}
		}
		// Try to get thread info
		var threads []int
		cpus, err := client.QueryCPUsFast()
		if err == nil {
			for _, cpu := range cpus {
				threads = append(threads, cpu.ThreadID)
			}
		}
		return VMStatusUpdateMsg{Status: status, Threads: threads}
	}
}

// waitForLog returns a tea.Cmd that reads one log line from the runner
func (m *VMRunningModel) waitForLog() tea.Cmd {
	return func() tea.Msg {
		// Guard against nil runner and nil channel
		if m.runner == nil {
			return VMLogMsg{Line: ""}
		}

		line, ok := <-m.runner.LogChan()
		if !ok {
			return VMLogMsg{Line: ""}
		}
		return VMLogMsg{Line: line}
	}
}

// waitForVMExit returns a tea.Cmd that waits for the VM process to exit
func (m *VMRunningModel) waitForVMExit() tea.Cmd {
	return func() tea.Msg {
		// Guard against nil runner
		if m.runner == nil {
			return VMStoppedMsg{
				VMName: m.vm.Name,
				VMID:   m.vm.ID,
				Reason: "no runner",
			}
		}

		<-m.runner.Done()
		err := m.runner.ExitError()
		reason := "exited"
		if err != nil {
			reason = fmt.Sprintf("exited: %v", err)
		}
		return VMStoppedMsg{
			VMName: m.vm.Name,
			VMID:   m.vm.ID,
			Reason: reason,
		}
	}
}

// pollStatus returns a tea.Cmd that polls VM status periodically
func (m *VMRunningModel) pollStatus() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		// Guard against nil runner (shouldn't happen in practice but defensive)
		if m.runner == nil {
			return VMStatusUpdateMsg{Status: "starting"}
		}

		client := m.runner.QMPClient()
		if client == nil {
			return VMStatusUpdateMsg{Status: "starting"}
		}

		status, err := client.QueryStatus()
		if err != nil {
			return VMStatusUpdateMsg{Status: "unknown"}
		}

		// Try to get thread info
		var threads []int
		cpus, err := client.QueryCPUsFast()
		if err == nil {
			for _, cpu := range cpus {
				threads = append(threads, cpu.ThreadID)
			}
		}

		return VMStatusUpdateMsg{Status: status, Threads: threads}
	})
}

// Update handles incoming messages
func (m *VMRunningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.calculateLayout()
		return m, nil

	case VMLogMsg:
		if msg.Line != "" {
			m.logLines = append(m.logLines, msg.Line)
			if len(m.logLines) > m.maxLogLines {
				m.logLines = m.logLines[len(m.logLines)-m.maxLogLines:]
			}
			m.updateViewport()
			m.vp.GotoBottom()
		}
		// If line is empty, channel was closed - don't re-poll
		if msg.Line == "" {
			return m, nil
		}
		return m, m.waitForLog()

	case VMStoppedMsg:
		m.status = "stopped"
		m.updateViewport()
		return m, nil

	case VMStatusUpdateMsg:
		m.status = msg.Status
		m.threads = msg.Threads
		return m, m.pollStatus()
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *VMRunningModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "q":
		// 'q' stops the VM if running
		if m.runner != nil && m.runner.IsRunning() {
			m.runner.Stop()
			m.status = "stopping"
			return m, nil
		}
		// If runner exists but VM is stopped, 'q' exits the view
		if m.runner != nil && !m.runner.IsRunning() {
			return m, func() tea.Msg { return VMRunningViewExitMsg{} }
		}

	case "ctrl+c":
		if m.runner != nil && m.runner.IsRunning() {
			m.runner.ForceStop()
			m.status = "stopping"
			return m, nil
		}
	}

	// Delegate scrolling to viewport
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

// calculateLayout calculates the info panel height and adjusts viewport
func (m *VMRunningModel) calculateLayout() {
	// Ensure minimum dimensions
	if m.width < 10 {
		m.width = 80 // Default terminal width
	}

	if !m.ready {
		// Initial setup
		infoHeight := m.calculateInfoHeight()
		m.infoHeight = infoHeight
		availableHeight := m.height - infoHeight - 3 // -3 for footer and separators
		if availableHeight < 5 {
			availableHeight = 5
		}
		m.vp = viewport.New(m.width, availableHeight)
		// HighPerformanceMode setting removed (no longer supported in bubbles v1)
		m.ready = true
	} else {
		// Update layout
		m.infoHeight = m.calculateInfoHeight()
		availableHeight := m.height - m.infoHeight - 3
		if availableHeight < 5 {
			availableHeight = 5
		}
		m.vp.Width = m.width
		m.vp.Height = availableHeight
	}
	m.updateViewport()
}

// calculateInfoHeight returns the height needed for the info panel
func (m *VMRunningModel) calculateInfoHeight() int {
	// Minimum base height
	height := 4 // Base: status line + vCPU info + blank + separator

	// Add lines for PCI devices (if runner available)
	if m.runner != nil {
		if len(m.runner.PCIPassthroughDevices()) > 0 {
			height += 1 + len(m.runner.PCIPassthroughDevices())
		}

		// Add lines for USB devices
		if len(m.runner.USBPassthroughDevices()) > 0 {
			height += 1 + len(m.runner.USBPassthroughDevices())
		}

		// Add line for threads if available
		if len(m.threads) > 0 {
			height += 1
		}
	}

	return height
}

// updateViewport syncs the log content to the viewport
func (m *VMRunningModel) updateViewport() {
	content := m.renderLogContent()
	m.vp.SetContent(content)
}

// View returns the view for the model
func (m *VMRunningModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Render info panel
	infoPanel := m.renderInfoPanel()

	// Render log viewport
	logView := m.vp.View()

	// Combine with separator
	return infoPanel + "\n" + logView
}

// renderInfoPanel renders the VM information panel at the top
func (m *VMRunningModel) renderInfoPanel() string {
	mutedStyle := lipgloss.NewStyle().Foreground(styles.Colors.Muted)
	labelStyle := lipgloss.NewStyle().Foreground(styles.Colors.Muted).Bold(false)
	valueStyle := styles.NormalTextStyle()

	// Status styles
	statusRunning := lipgloss.NewStyle().
		Foreground(styles.Colors.Success).
		Bold(true)

	statusStopped := lipgloss.NewStyle().
		Foreground(styles.Colors.Error).
		Bold(true)

	statusStarting := lipgloss.NewStyle().
		Foreground(styles.Colors.Warning).
		Bold(true)

	var b strings.Builder

	// === Section 1: VM Name and Status ===
	b.WriteString(labelStyle.Render("VM: "))
	b.WriteString(valueStyle.Render(m.vm.Name))
	b.WriteString("  ")

	// Status
	var statusStr string
	switch m.status {
	case "running":
		statusStr = statusRunning.Render("[RUNNING]")
	case "stopped", "exited":
		statusStr = statusStopped.Render("[STOPPED]")
	case "stopping":
		statusStr = statusStarting.Render("[STOPPING]")
	default:
		statusStr = statusStarting.Render("[STARTING]")
	}
	b.WriteString(labelStyle.Render("Status: "))
	b.WriteString(statusStr)

	// Uptime
	if !m.startTime.IsZero() && m.runner != nil && m.runner.IsRunning() {
		uptime := time.Since(m.startTime).Truncate(time.Second)
		b.WriteString("  ")
		b.WriteString(labelStyle.Render("Uptime: "))
		b.WriteString(valueStyle.Render(uptime.String()))
	}
	b.WriteString("\n")

	// === Section 2: Guest Resources ===
	// Only show resource info if runner is available
	if m.runner != nil {
		b.WriteString(labelStyle.Render("Memory: "))
		memMB := m.runner.MemoryMB()
		memGB := memMB / 1024
		if memMB%1024 == 0 {
			b.WriteString(valueStyle.Render(fmt.Sprintf("%d GB", memGB)))
		} else {
			b.WriteString(valueStyle.Render(fmt.Sprintf("%.1f GB", float64(memMB)/1024.0)))
		}

		// vCPU count
		vcpuCount := m.runner.VCpuCount()
		b.WriteString("  ")
		b.WriteString(labelStyle.Render("vCPUs: "))
		b.WriteString(valueStyle.Render(fmt.Sprintf("%d", vcpuCount)))

		// vCPU pinning
		pinning := m.runner.VCPUPinning()
		b.WriteString("  ")
		b.WriteString(labelStyle.Render("Pinning: "))
		if pinning.Enabled {
			b.WriteString(valueStyle.Render("enabled"))
		} else {
			b.WriteString(valueStyle.Render("disabled"))
		}
		b.WriteString("\n")

		// === Section 3: PCI Passthrough ===
		pciDevices := m.runner.PCIPassthroughDevices()
		if len(pciDevices) > 0 {
			b.WriteString(labelStyle.Render("PCI: "))
			for i, dev := range pciDevices {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(valueStyle.Render(dev.Address))
				if dev.Name != "" && len(dev.Name) <= 30 {
					b.WriteString(mutedStyle.Render(" (" + dev.Name + ")"))
				}
			}
			b.WriteString("\n")
		}

		// === Section 4: USB Passthrough ===
		usbDevices := m.runner.USBPassthroughDevices()
		if len(usbDevices) > 0 {
			b.WriteString(labelStyle.Render("USB: "))
			for i, dev := range usbDevices {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(valueStyle.Render(dev.BusID))
				if dev.Name != "" && len(dev.Name) <= 30 {
					b.WriteString(mutedStyle.Render(" (" + dev.Name + ")"))
				}
			}
			b.WriteString("\n")
		}

		// === Section 5: TPM ===
		if m.vm.TPMEnabled {
			b.WriteString(labelStyle.Render("TPM: "))
			b.WriteString(valueStyle.Render("enabled"))
			b.WriteString("\n")
		}
	} else {
		// Runner not available, show placeholder values
		b.WriteString(labelStyle.Render("Memory: "))
		b.WriteString(valueStyle.Render("N/A"))
		b.WriteString("  ")
		b.WriteString(labelStyle.Render("vCPUs: "))
		b.WriteString(valueStyle.Render("N/A"))
		b.WriteString("\n")
	}

	// Show thread IDs if available (for both runner present and test)
	if len(m.threads) > 0 {
		b.WriteString(labelStyle.Render("vCPU Threads: "))
		for i, tid := range m.threads {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(valueStyle.Render(fmt.Sprintf("%d", tid)))
		}
		b.WriteString("\n")
	}

	// === Separator line ===
	b.WriteString(mutedStyle.Render("─── QEMU Output ───"))

	return b.String()
}

// renderLogContent renders the log content for the viewport
func (m *VMRunningModel) renderLogContent() string {
	mutedStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted)

	logStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	var output strings.Builder

	// Separator (moved inside viewport so it's visible when scrolling back up)
	output.WriteString(mutedStyle.Render("─── QEMU Output ───"))
	output.WriteString("\n\n")

	// Log lines
	if len(m.logLines) == 0 {
		output.WriteString(mutedStyle.Render("Waiting for output..."))
		output.WriteString("\n")
	} else {
		for _, line := range m.logLines {
			output.WriteString(logStyle.Render(line))
			output.WriteString("\n")
		}
	}

	// Footer
	output.WriteString("\n")
	if m.runner != nil && m.runner.IsRunning() {
		output.WriteString(mutedStyle.Render("q: Stop VM  Ctrl+C: Force Kill  ↑/↓: Scroll"))
	} else {
		output.WriteString(mutedStyle.Render("q: Exit view  ↑/↓: Scroll"))
	}

	return output.String()
}

// SetSize updates the model dimensions
func (m *VMRunningModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.calculateLayout()
}

// Runner returns the underlying VM runner
func (m *VMRunningModel) Runner() *vm.VMRunner {
	return m.runner
}