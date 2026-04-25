package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glemsom/dkvmmanager/internal/models"
	"github.com/glemsom/dkvmmanager/internal/tui/styles"
)

// VMCardView renders VMs as bordered cards instead of a table
type VMCardView struct {
	vms    []models.VM
	cursor int
	width  int
	height int
	focused bool
}

// NewVMCardView creates a new VM card view
func NewVMCardView(vms []models.VM, width, height int) *VMCardView {
	return &VMCardView{
		vms:    vms,
		cursor: 0,
		width:  width,
		height: height,
		focused: false,
	}
}

// SetSize updates the card view dimensions
func (c *VMCardView) SetSize(width, height int) {
	c.width = width
	c.height = height
}

// SetFocused sets whether the card view has focus
func (c *VMCardView) SetFocused(focused bool) {
	c.focused = focused
}

// Update handles messages for the card view
func (c *VMCardView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if c.cursor > 0 {
				c.cursor--
			}
		case "down", "j":
			if c.cursor < len(c.vms)-1 {
				c.cursor++
			}
		}
	}
	return c, nil
}

// View renders all VM cards
func (c *VMCardView) View() string {
	if len(c.vms) == 0 {
		return lipgloss.NewStyle().
			Foreground(styles.Colors.Muted).
			Italic(true).
			Render("No virtual machines configured.")
	}

	var cards []string
	for i, vm := range c.vms {
		cards = append(cards, c.renderCard(vm, i == c.cursor))
	}

	return strings.Join(cards, "\n")
}

// renderCard renders a single VM as a bordered card
func (c *VMCardView) renderCard(vm models.VM, selected bool) string {
	cardWidth := c.width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	// Determine border color based on selection and focus state
	var borderColor lipgloss.Color
	if selected {
		if c.focused {
			borderColor = styles.Colors.Primary
		} else {
			borderColor = styles.Colors.Foreground
		}
	} else {
		borderColor = styles.Colors.Muted
	}

	// Top border with title
	title := vm.Name
	titlePadded := " " + title + " "
	remaining := cardWidth - 2 - lipgloss.Width(titlePadded)
	if remaining < 0 {
		remaining = 0
	}
	topBorder := lipgloss.NewStyle().Foreground(borderColor).
		Render("╭─") +
		lipgloss.NewStyle().Foreground(borderColor).Bold(true).
			Render(titlePadded) +
		lipgloss.NewStyle().Foreground(borderColor).
			Render(strings.Repeat("─", remaining)+"╮")

	// Status line
	statusIcon := styles.StatusIndicator("stopped")
	statusText := "Stopped"
	// Simple heuristic: if VM has a MAC, it was configured for running
	if vm.MAC != "" {
		statusIcon = styles.StatusIndicator("running")
		statusText = "Configured"
	}

	tpmStatus := "No"
	if vm.TPMEnabled {
		tpmStatus = "Yes"
	}

	infoLine := fmt.Sprintf(" %s %s │ Disks: %d │ MAC: %s │ TPM: %s",
		statusIcon, statusText, len(vm.HardDisks), vm.MAC, tpmStatus)
	if vm.MAC == "" {
		infoLine = fmt.Sprintf(" %s %s │ Disks: %d │ MAC: - │ TPM: %s",
			statusIcon, statusText, len(vm.HardDisks), tpmStatus)
	}

	// Pad the info line
	infoPadded := lipgloss.NewStyle().
		Width(cardWidth).
		MaxWidth(cardWidth).
		Render(infoLine)

	sideStyle := lipgloss.NewStyle().Foreground(borderColor)
	body := sideStyle.Render("│") + infoPadded + sideStyle.Render("│")

	// Bottom border
	bottomBorder := sideStyle.Render("╰" + strings.Repeat("─", cardWidth) + "╯")

	card := topBorder + "\n" + body + "\n" + bottomBorder

	// Apply selected highlight
	if selected {
		card = lipgloss.NewStyle().
			Bold(true).
			Render(card)
	}

	return card
}

// SelectedVM returns the currently selected VM
func (c *VMCardView) SelectedVM() *models.VM {
	if len(c.vms) == 0 || c.cursor < 0 || c.cursor >= len(c.vms) {
		return nil
	}
	return &c.vms[c.cursor]
}

// Cursor returns the current cursor position
func (c *VMCardView) Cursor() int {
	return c.cursor
}

// SetCursor sets the cursor position with bounds checking
func (c *VMCardView) SetCursor(index int) {
	if index >= 0 && index < len(c.vms) {
		c.cursor = index
	}
}

// SetVMs updates the VM list
func (c *VMCardView) SetVMs(vms []models.VM) {
	c.vms = vms
	if c.cursor >= len(vms) && len(vms) > 0 {
		c.cursor = len(vms) - 1
	}
}

// Init initializes the card view
func (c *VMCardView) Init() tea.Cmd {
	return nil
}
