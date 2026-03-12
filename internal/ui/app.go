package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"recon/internal/ui/theme"
	"recon/internal/ui/views/newscan"
	"recon/internal/ui/views/statusbar"
)

type model struct {
	width   int
	height  int
	newScan newscan.NewScanModel
	status  statusbar.Model
	active  viewID
}

type viewID int

const (
	viewNewScan viewID = iota
	viewLogs
	viewHistory
)

func InitalModel(toolName, toolVersion string) model {
	return model{
		newScan: newscan.NewModel(),
		status:  statusbar.NewModel(toolName, toolVersion),
		active:  viewNewScan,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return tea.Batch(m.status.Init(), newscan.BlinkCmd())
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
		case "alt+up":
			m.active = viewNewScan
			return m, nil
		case "alt+right":
			m.active = viewLogs
			return m, nil
		case "alt+down":
			m.active = viewHistory
			return m, nil

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	var statusCmd tea.Cmd
	m.status, statusCmd = m.status.Update(msg)
	if _, ok := msg.(tea.KeyMsg); !ok || m.active == viewNewScan {
		m.newScan, cmd = m.newScan.Update(msg)
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, tea.Batch(cmd, statusCmd)
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	uiTheme := theme.Load()

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
	if bottomHeight < 5 {
		bottomHeight = 5
	}
	if leftWidth < 20 {
		leftWidth = 20
		rightWidth = m.width - leftWidth
	}

	newScan := m.newScan.View(leftWidth-1, topHeight, m.active == viewNewScan)

	panel := lipgloss.NewStyle().Padding(1, 2)

	runningTitle := m.renderPanelTitle("RUNNING / LOGS", rightWidth-1, uiTheme, m.active == viewLogs)
	runningBody := "No active scan."
	running := panel.
		Width(rightWidth - 1).
		Height(topHeight).
		Render(lipgloss.JoinVertical(lipgloss.Left, runningTitle, "", runningBody))

	statusHeight := 1
	historyHeight := bottomHeight - statusHeight
	if historyHeight < 3 {
		historyHeight = 3
	}

	historyTitle := m.renderPanelTitle("SCAN HISTORY", usableWidth, uiTheme, m.active == viewHistory)
	historyBody := "(no scans yet)"
	history := panel.
		Width(usableWidth).
		Height(historyHeight - 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, historyTitle, "", historyBody))

	vertLine := strings.Repeat("│\n", topHeight-1) + "│"
	vert := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Render(vertLine)

	horiz := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Render(strings.Repeat("─", usableWidth))

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, newScan, vert, running)
	status := m.status.View(m.width, uiTheme)
	used := lipgloss.Height(topRow) + lipgloss.Height(horiz) + lipgloss.Height(history) + lipgloss.Height(status)
	remaining := m.height - used
	if remaining < 0 {
		remaining = 0
	}
	filler := makeFiller(m.width, uiTheme.AppBg, remaining)

	content := lipgloss.JoinVertical(lipgloss.Left, topRow, horiz, history, filler, status)
	base := lipgloss.NewStyle().
		Background(lipgloss.Color(uiTheme.AppBg)).
		Width(m.width).
		Height(m.height)
	return base.Render(content)
}

func (m model) renderPanelTitle(text string, width int, uiTheme theme.Theme, active bool) string {
	titleWidth := width - 4
	if titleWidth < 10 {
		titleWidth = 10
	}
	titleFg := uiTheme.StatusFg
	titleBg := uiTheme.StatusBg
	underline := strings.Repeat(" ", titleWidth)
	if active {
		titleFg = uiTheme.AccentFg
		titleBg = uiTheme.AccentBg
		underline = strings.Repeat("─", titleWidth)
	}
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(titleFg)).
		Background(lipgloss.Color(titleBg)).
		Padding(0, 1).
		Width(titleWidth).
		Height(1).
		Align(lipgloss.Center).
		Render(text)
	underlineLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.AccentBg)).
		Background(lipgloss.Color(uiTheme.AppBg)).
		Width(titleWidth).
		Render(underline)
	return lipgloss.JoinVertical(lipgloss.Left, title, underlineLine)
}

func makeFiller(width int, bg string, lines int) string {
	if lines <= 0 {
		return ""
	}
	line := strings.Repeat(" ", width)
	block := line
	for i := 1; i < lines; i++ {
		block += "\n" + line
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Render(block)
}
