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

	// Debug mode
	debugMode bool
}

// NewVMRunningModel creates a new VM running model
func NewVMRunningModel(vmObj *models.VM, runner *vm.VMRunner) *VMRunningModel {
	return &VMRunningModel{
		vm:          vmObj,
		runner:      runner,
		maxLogLines: 500,
		debugMode:   debugMode,
	}
}

// Init initializes the model
func (m *VMRunningModel) Init() tea.Cmd {
	return tea.Batch(
		m.waitForLog(),
		m.waitForVMExit(),
		m.pollStatus(),
	)
}

// waitForLog returns a tea.Cmd that reads one log line from the runner
func (m *VMRunningModel) waitForLog() tea.Cmd {
	return func() tea.Msg {
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
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
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
		if !m.ready {
			m.vp = viewport.New(msg.Width, m.height)
			m.ready = true
		} else {
			m.vp.Width = msg.Width
			m.vp.Height = m.height
		}
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

// updateViewport syncs the log content to the viewport
func (m *VMRunningModel) updateViewport() {
	content := m.renderContent()
	m.vp.SetContent(content)
}

// View returns the view for the model
func (m *VMRunningModel) View() string {
	if !m.ready {
		return "Loading..."
	}
	return m.vp.View()
}

// renderContent builds the full content string
func (m *VMRunningModel) renderContent() string {
	statusRunning := lipgloss.NewStyle().
		Foreground(styles.Colors.Success).
		Bold(true)

	statusStopped := lipgloss.NewStyle().
		Foreground(styles.Colors.Error).
		Bold(true)

	statusStarting := lipgloss.NewStyle().
		Foreground(styles.Colors.Warning).
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted)

	logStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	var output strings.Builder

	// Status line
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
	output.WriteString(fmt.Sprintf("Status: %s", statusStr))

	// Uptime
	if !m.startTime.IsZero() && m.runner != nil && m.runner.IsRunning() {
		uptime := time.Since(m.startTime).Truncate(time.Second)
		output.WriteString(fmt.Sprintf("  Uptime: %s", mutedStyle.Render(uptime.String())))
	}
	output.WriteString("\n")

	// Thread info with progress bar
	if len(m.threads) > 0 {
		threadStrs := make([]string, len(m.threads))
		for i, t := range m.threads {
			threadStrs[i] = fmt.Sprintf("%d", t)
		}
		output.WriteString(fmt.Sprintf("vCPU Threads: %s\n",
			mutedStyle.Render(strings.Join(threadStrs, ", "))))

		// Show log activity as a progress bar (log lines / max capacity)
		logFraction := float64(len(m.logLines)) / float64(m.maxLogLines)
		if logFraction > 1 {
			logFraction = 1
		}
		output.WriteString(styles.RenderProgressBar(
			logFraction, 20, "Log Buffer",
			fmt.Sprintf("%d/%d lines", len(m.logLines), m.maxLogLines),
		))
		output.WriteString("\n")
	}

	output.WriteString("\n")

	// Log separator
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
		output.WriteString(mutedStyle.Render("q: Stop VM  Ctrl+C: Force Kill  Scroll: ↑/↓"))
	} else {
		output.WriteString(mutedStyle.Render("q: Exit view  ↑/↓: Scroll"))
	}

	return output.String()
}

// SetSize updates the model dimensions
func (m *VMRunningModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	if !m.ready {
		m.vp = viewport.New(w, h)
		m.ready = true
	} else {
		m.vp.Width = w
		m.vp.Height = h
	}
	m.updateViewport()
}

// Runner returns the underlying VM runner
func (m *VMRunningModel) Runner() *vm.VMRunner {
	return m.runner
}
