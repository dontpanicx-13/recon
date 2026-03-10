package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	width   int
	height  int
	newScan NewScanModel
}

func InitalModel() model {
	return model{
		newScan: NewNewScanModel(),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.newScan, cmd = m.newScan.Update(msg)

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, cmd
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	usableWidth := m.width - 4
	usableHeight := m.height - 4

	if usableWidth < 20 {
		usableWidth = 20
	}
	if usableHeight < 10 {
		usableHeight = 10
	}

	topHeight := (usableHeight * 6) / 10
	bottomHeight := usableHeight - topHeight
	leftWidth := (usableWidth * 35) / 100
	rightWidth := usableWidth - leftWidth

	if topHeight < 6 {
		topHeight = 6
	}
	if bottomHeight < 4 {
		bottomHeight = 4
	}
	if leftWidth < 20 {
		leftWidth = 20
		rightWidth = m.width - leftWidth
	}

	newScan := m.newScan.View(leftWidth-1, topHeight)

	panel := lipgloss.NewStyle().Padding(1, 2)

	running := panel.
		Width(rightWidth-1).
		Height(topHeight).
		Render("Running / Logs\n\nNo active scan.")

	history := panel.
		Width(usableWidth).
		Height(bottomHeight-1).
		Render("Scan History\n\n(no scans yet)")

	vertLine := strings.Repeat("│\n", topHeight-1) + "│"
	vert := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Render(vertLine)

	horiz := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Render(strings.Repeat("─", usableWidth))

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, newScan, vert, running)
	return lipgloss.JoinVertical(lipgloss.Left, topRow, horiz, history)
}
