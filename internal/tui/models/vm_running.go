// Package models provides the BubbleTea models for the DKVM Manager TUI
package models

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
	"github.com/glemsom/dkvmmanager/internal/vm"
)

// VMStartedMsg is sent when a VM starts successfully
type VMStartedMsg struct {
	Runner *vm.VMRunner
	VMName string
	VMID   string
}

// VMStartErrorMsg is sent when a VM fails to start
type VMStartErrorMsg struct {
	VMName string
	Err    error
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

// LogSeedMsg is sent after seeding the log buffer from the persisted file.
// It carries the lines read from the persisted log to initialize the view.
type LogSeedMsg struct {
	Lines []string
}

// VMStatusUpdateMsg is sent periodically to refresh VM status
type VMStatusUpdateMsg struct {
	Status  string
	Threads []int
}

// VMMetricsUpdateMsg is sent periodically (2 s cadence) with a full Metrics
// snapshot. It is routed directly to VMRunningModel, bypassing the view
// registry (same pattern as VMStatusUpdateMsg).
type VMMetricsUpdateMsg struct {
	Metrics vm.Metrics
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

	// Subscribed log channel (from runner.Subscribe())
	logSub <-chan string

	// Status
	status  string
	threads []int

	// Status polling tracking
	pollingSince time.Time

	// Metrics snapshot (latest, updated every 2s)
	metrics vm.Metrics

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
		m.seedAndSubscribe(), // seed from persisted log, then subscribe
		m.waitForVMExit(),
		m.pollStatus(),      // periodic (500ms)
		m.initialStatus(),   // immediate
		m.pollMetrics(),     // periodic (2s) — decoupled from status poll
	)
}

// seedAndSubscribe reads the last N lines from the persisted log, seeds the
// ring buffer, then subscribes to the live log stream. The drain-on-subscribe
// semantic in the runner ensures that any lines buffered between the file read
// and the subscription are delivered before new live lines.
func (m *VMRunningModel) seedAndSubscribe() tea.Cmd {
	return func() tea.Msg {
		if m.runner == nil {
			return nil
		}

		// Read the last 500 lines from the persisted log
		lines, _ := m.runner.RecentLog(500)

		// Subscribe to live log (drains any buffered lines first)
		m.logSub = m.runner.Subscribe()

		return LogSeedMsg{Lines: lines}
	}
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

// waitForLog returns a tea.Cmd that reads one log line from the subscribed channel.
// The subscription is set up by seedAndSubscribe during Init.
func (m *VMRunningModel) waitForLog() tea.Cmd {
	return func() tea.Msg {
		// Guard against nil runner or not-yet-subscribed channel
		if m.runner == nil || m.logSub == nil {
			return nil // no-op until runner and subscription are set
		}

		line, ok := <-m.logSub
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
			return nil // no-op until runner is set
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
			// QMP not connected yet - use fallback logic:
			// If the VM process is running and has been for a while, assume running.
			// This handles cases where QMP socket may not be available or QEMU
			// is taking longer than expected to create it.
			if m.runner.IsRunning() {
				// If pollingSince is set and we've been trying for a while,
				// assume the VM is running even without QMP
				if !m.pollingSince.IsZero() && time.Since(m.pollingSince) > 5*time.Second {
					return VMStatusUpdateMsg{Status: "running"}
				}
			}
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

// pollMetrics returns a tea.Cmd that polls VM metrics on a 2 s cadence,
// decoupled from the 500 ms status poll. The metrics tick bypasses the
// view registry (same pattern as VMStatusUpdateMsg).
func (m *VMRunningModel) pollMetrics() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		if m.runner == nil {
			return VMMetricsUpdateMsg{}
		}

		snap, err := m.runner.Snapshot()
		if err != nil {
			// Degraded: return empty metrics; view will show N/A
			return VMMetricsUpdateMsg{}
		}
		return VMMetricsUpdateMsg{Metrics: snap}
	})
}

// Update handles incoming messages
func (m *VMRunningModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.calculateLayout()
		return m, nil

	case LogSeedMsg:
		// Seed the ring buffer from the persisted log tail
		if len(msg.Lines) > 0 {
			// Keep at most maxLogLines
			if len(msg.Lines) > m.maxLogLines {
				msg.Lines = msg.Lines[len(msg.Lines)-m.maxLogLines:]
			}
			m.logLines = msg.Lines
			m.updateViewport()
			m.vp.GotoBottom()
		}
		// Now start reading from the subscribed channel
		return m, m.waitForLog()

	case VMLogMsg:
		if msg.Line != "" {
			// Dedup: skip if line already appears in the tail of the buffer.
			// Prevents overlap between the persisted file tail (seeded via
			// LogSeedMsg) and the staging buffer drained by Subscribe().
			// Check last 20 lines to catch multi-line overlap at the boundary.
			isDup := false
			for i := len(m.logLines) - 1; i >= 0 && i >= len(m.logLines)-20; i-- {
				if m.logLines[i] == msg.Line {
					isDup = true
					break
				}
			}
			if !isDup {
				m.logLines = append(m.logLines, msg.Line)
				if len(m.logLines) > m.maxLogLines {
					m.logLines = m.logLines[len(m.logLines)-m.maxLogLines:]
				}
				m.updateViewport()
				m.vp.GotoBottom()
			}
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

	case VMMetricsUpdateMsg:
		m.metrics = msg.Metrics
		return m, m.pollMetrics()

	case VMStartedMsg:
		m.runner = msg.Runner
		m.status = "starting" // will be updated by initialStatus
		m.pollingSince = time.Now()
		return m, tea.Batch(
			m.seedAndSubscribe(),
			m.waitForVMExit(),
			m.pollStatus(),
			m.initialStatus(),
			m.pollMetrics(),
		)
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
		m.vp = viewport.New(viewport.WithWidth(m.width), viewport.WithHeight(availableHeight))
		// HighPerformanceMode setting removed (no longer supported in bubbles v1)
		m.ready = true
	} else {
		// Update layout
		m.infoHeight = m.calculateInfoHeight()
		availableHeight := m.height - m.infoHeight - 3
		if availableHeight < 5 {
			availableHeight = 5
		}
		m.vp.SetWidth(m.width)
		m.vp.SetHeight(availableHeight)
	}
	m.updateViewport()
}

// calculateInfoHeight returns the height needed for the info panel
func (m *VMRunningModel) calculateInfoHeight() int {
	// Minimum base height
	height := 4 // Base: status line + vCPU info + blank + separator

	// Add lines for vCPU metrics (independent of runner)
	if len(m.metrics.VCPUs) > 0 {
		height++
	}

	// Add lines for PCI devices (if runner available)
	if m.runner != nil {
		if len(m.runner.PCIPassthroughDevices()) > 0 {
			height += 1 + len(m.runner.PCIPassthroughDevices())
		}

		// Add lines for USB devices
		if len(m.runner.USBPassthroughDevices()) > 0 {
			height += 1 + len(m.runner.USBPassthroughDevices())
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
func (m *VMRunningModel) View() tea.View {
	if !m.ready {
		return tea.NewView("Loading...")
	}

	// Render info panel
	infoPanel := m.renderInfoPanel()

	// Render log viewport
	logView := m.vp.View()

	// Combine with separator
	return tea.NewView(infoPanel + "\n" + logView)
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
	case "running", "paused", "postmigrate", "prelaunch":
		statusStr = statusRunning.Render("[RUNNING]")
	case "stopped", "exited":
		statusStr = statusStopped.Render("[STOPPED]")
	case "stopping", "finish":
		statusStr = statusStarting.Render("[STOPPING]")
	default:
		statusStr = statusStarting.Render("[STARTING]")
	}
	b.WriteString(labelStyle.Render("Status: "))
	b.WriteString(statusStr)

	// Uptime
	if m.runner != nil && m.runner.IsRunning() {
		uptime := time.Since(m.runner.StartTime()).Truncate(time.Second)
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

	// === Section: Per-vCPU Metrics (runs on its own cadence, independent of runner) ===
	if len(m.metrics.VCPUs) > 0 {
		b.WriteString(labelStyle.Render("vCPU%: "))
		for i, vcpu := range m.metrics.VCPUs {
			if i > 0 {
				b.WriteString("  ")
			}
			// CPUTimeNs holds CPU% * 100 (fixed-point, two decimals)
			cpuPct := float64(vcpu.CPUTimeNs) / 100.0
			b.WriteString(valueStyle.Render(fmt.Sprintf("#%d: %.1f%%", i, cpuPct)))
		}
		// Aggregate total (sum of per-vCPU percentages)
		var totalPct float64
		for _, vcpu := range m.metrics.VCPUs {
			totalPct += float64(vcpu.CPUTimeNs) / 100.0
		}
		b.WriteString("  ")
		b.WriteString(labelStyle.Render("total: "))
		b.WriteString(valueStyle.Render(fmt.Sprintf("%.1f%%", totalPct)))
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
		Foreground(styles.Colors.Foreground)

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

// FileBrowserActive always returns false; VMRunning has no sub-file-browser.
func (m *VMRunningModel) FileBrowserActive() bool { return false }

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

// startVMCommand returns a tea.Cmd that runs runner.Start() in a goroutine
// to avoid blocking the BubbleTea event loop during VM startup.
// Sends VMStartedMsg on success, VMStartErrorMsg on failure.
func startVMCommand(runner *vm.VMRunner, vmName, vmID string) tea.Cmd {
	ch := make(chan tea.Msg, 1)
	go func() {
		if err := runner.Start(); err != nil {
			ch <- VMStartErrorMsg{VMName: vmName, Err: err}
		} else {
			ch <- VMStartedMsg{Runner: runner, VMName: vmName, VMID: vmID}
		}
	}()
	return func() tea.Msg { return <-ch }
}