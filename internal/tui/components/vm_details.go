package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

type VMDetailsPanel struct {
	vm     *models.VM
	width  int
	height int
	title  string
	focused bool
}

func NewVMDetailsPanel() *VMDetailsPanel {
	return &VMDetailsPanel{
		vm:     nil,
		width:  0,
		height: 0,
		title:  "VM Details",
		focused: false,
	}
}

func (p *VMDetailsPanel) SetVM(vm *models.VM) {
	p.vm = vm
}

func (p *VMDetailsPanel) SetFocused(focused bool) {
	p.focused = focused
}

func (p *VMDetailsPanel) SetTitle(title string) {
	p.title = title
}

func (p *VMDetailsPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

func (p *VMDetailsPanel) View() string {
	if p.vm == nil {
		return p.renderEmpty()
	}
	return p.renderDetails()
}

func (p *VMDetailsPanel) renderEmpty() string {
	contentWidth := p.width - 4
	if contentWidth < 30 {
		contentWidth = 30
	}

	message := "Select a VM to view details"
	styledMessage := lipgloss.NewStyle().
		Width(contentWidth).
		Foreground(styles.Colors.Muted).
		Italic(true).
		Render(message)

	// Apply panel styling for elevation based on focus
	var panelStyle lipgloss.Style
	if p.focused {
		panelStyle = styles.ActiveLayeredPanelStyle()
	} else {
		panelStyle = styles.LayeredPanelStyle()
	}
	return panelStyle.
		Width(p.width).
		Height(p.height).
		Render(styledMessage)
}

func (p *VMDetailsPanel) renderDetails() string {
	var lines []string

	contentWidth := p.width - 4
	if contentWidth < 30 {
		contentWidth = 30
	}

	titleStyle := styles.HeaderStyle()
	fieldStyle := styles.DetailLabelStyle()
	valueStyle := styles.DetailValueStyle()
	sectionStyle := styles.DetailSectionStyle()

	lines = append(lines, titleStyle.Render(p.title))
	lines = append(lines, "")
	lines = append(lines, p.renderField("Name", p.vm.Name, fieldStyle, valueStyle, contentWidth))
	lines = append(lines, p.renderField("ID", p.vm.ID, fieldStyle, valueStyle, contentWidth))
	lines = append(lines, "")

	// Storage section
	lines = append(lines, sectionStyle.Render(" Storage "))
	if len(p.vm.HardDisks) > 0 {
		lines = append(lines, p.renderArrayField("Hard Disks", p.vm.HardDisks, fieldStyle, valueStyle, contentWidth))
	} else {
		lines = append(lines, p.renderField("Hard Disks", "None", fieldStyle, valueStyle, contentWidth))
	}
	lines = append(lines, "")

	if len(p.vm.CDROMs) > 0 {
		lines = append(lines, p.renderArrayField("CDROMs", p.vm.CDROMs, fieldStyle, valueStyle, contentWidth))
	} else {
		lines = append(lines, p.renderField("CDROMs", "None", fieldStyle, valueStyle, contentWidth))
	}
	lines = append(lines, "")

	// Network section
	lines = append(lines, sectionStyle.Render(" Network "))
	lines = append(lines, p.renderField("MAC Address", p.vm.MAC, fieldStyle, valueStyle, contentWidth))
	lines = append(lines, p.renderField("Network Mode", p.vm.NetworkMode, fieldStyle, valueStyle, contentWidth))
	lines = append(lines, p.renderField("VNC Binding", p.vm.VNCListen, fieldStyle, valueStyle, contentWidth))
	lines = append(lines, "")

	// System section
	lines = append(lines, sectionStyle.Render(" System "))
	lines = append(lines, p.renderField("TPM", formatBool(p.vm.TPMEnabled), fieldStyle, valueStyle, contentWidth))
	lines = append(lines, "")

	// Timestamps section
	lines = append(lines, sectionStyle.Render(" Timestamps "))
	if !p.vm.CreatedAt.IsZero() {
		lines = append(lines, p.renderField("Created", formatTimestamp(p.vm.CreatedAt), fieldStyle, valueStyle, contentWidth))
	}
	if !p.vm.UpdatedAt.IsZero() {
		lines = append(lines, p.renderField("Updated", formatTimestamp(p.vm.UpdatedAt), fieldStyle, valueStyle, contentWidth))
	}

	content := strings.Join(lines, "\n")
	
	// Apply panel styling for elevation based on focus
	var panelStyle lipgloss.Style
	if p.focused {
		panelStyle = styles.ActiveLayeredPanelStyle()
	} else {
		panelStyle = styles.LayeredPanelStyle()
	}
	return panelStyle.
		Width(p.width).
		Height(p.height).
		Render(content)
}

func (p *VMDetailsPanel) renderField(label, value string, fieldStyle, valueStyle lipgloss.Style, width int) string {
	if value == "" {
		value = "-"
	}
	labelPadded := fieldStyle.Render(label + ":")
	valueTruncated := truncate(value, width-2-lipgloss.Width(labelPadded))
	return labelPadded + " " + valueStyle.Render(valueTruncated)
}

func (p *VMDetailsPanel) renderArrayField(label string, values []string, fieldStyle, valueStyle lipgloss.Style, width int) string {
	labelPadded := fieldStyle.Render(label + ":")
	if len(values) == 0 {
		return labelPadded + " " + valueStyle.Render("-")
	}

	firstValue := values[0]
	firstPadded := labelPadded + " " + valueStyle.Render(truncate(firstValue, width-2-lipgloss.Width(labelPadded)))

	var lines []string
	lines = append(lines, firstPadded)

	indent := strings.Repeat(" ", lipgloss.Width(labelPadded)+1)
	for i := 1; i < len(values); i++ {
		truncated := truncate(values[i], width-2-lipgloss.Width(indent))
		lines = append(lines, indent+valueStyle.Render(truncated))
	}

	return strings.Join(lines, "\n")
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen > 3 {
		return s[:maxLen-3] + "..."
	}
	return s[:maxLen]
}

func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}

func (p *VMDetailsPanel) renderActionBar(width int) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Primary).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.Colors.Muted)

	separator := descStyle.Render(" ")

	keyHints := keyStyle.Render("[s]") + descStyle.Render(" Start") +
		separator +
		keyStyle.Render("[e]") + descStyle.Render(" Edit") +
		separator +
		keyStyle.Render("[d]") + descStyle.Render(" Delete")

	borderColor := styles.Colors.Muted
	actionBar := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(width).
		Render(keyHints + "\n\n" +
			lipgloss.NewStyle().
				Foreground(styles.Colors.Background).
				Background(styles.Colors.Primary).
				Bold(true).
				Padding(0, 2).
				MarginRight(1).
				Render("[Start VM]") +
			lipgloss.NewStyle().
				Foreground(styles.Colors.Primary).
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.Colors.Primary).
				Padding(0, 1).
				MarginRight(1).
				Render("[Edit]") +
			lipgloss.NewStyle().
				Foreground(styles.Colors.Error).
				Border(lipgloss.NormalBorder()).
				BorderForeground(styles.Colors.Error).
				Padding(0, 1).
				Render("[Delete]"))

	return actionBar
}

func (p *VMDetailsPanel) SelectedVM() *models.VM {
	return p.vm
}

func (p *VMDetailsPanel) Init() {}
func (p *VMDetailsPanel) Update(msg interface{}) (interface{}, error) {
	return nil, fmt.Errorf("VMDetailsPanel.Update not implemented")
}

func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
