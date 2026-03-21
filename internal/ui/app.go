package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"recon/internal/scanner"
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

	scanRunner      *scanRunner
	scanActive      bool
	scanStatus      string
	scanStart       time.Time
	scanElapsed     time.Duration
	scanPortsTotal  int
	scanPortsProbed int
	scanHostsTotal  int
	scanHostsDone   int
	scanOpenPorts   int
	scanLogs        []string
	scanLastError   string
	scanLogTop      int
	scanLogFollow   bool
	scanCancel      context.CancelFunc
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
	case newscan.StartScanMsg:
		if m.scanActive {
			return m, nil
		}
		runner := &scanRunner{events: make(chan scanner.Event, 256)}
		ctx, cancel := context.WithCancel(context.Background())
		m.scanRunner = runner
		m.scanActive = true
		m.scanStatus = "running"
		m.scanLastError = ""
		m.scanLogs = nil
		m.scanPortsTotal = 0
		m.scanPortsProbed = 0
		m.scanHostsTotal = 0
		m.scanHostsDone = 0
		m.scanOpenPorts = 0
		m.scanStart = time.Now()
		m.scanElapsed = 0
		m.scanLogTop = 0
		m.scanLogFollow = true
		m.scanCancel = cancel
		m.active = viewLogs
		m.newScan.SetDisabled(true)

		return m, tea.Batch(runScanCmd(ctx, msg.Config, runner), listenScanCmd(runner))
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case scanEventMsg:
		m.handleScanEvent(msg.Event)
		if m.scanRunner != nil {
			return m, listenScanCmd(m.scanRunner)
		}
	case scanDoneMsg:
		m.scanActive = false
		if msg.Err != nil {
			m.scanStatus = "failed"
			m.scanLastError = msg.Err.Error()
		} else {
			m.scanStatus = msg.Result.Meta.Status
			m.scanElapsed = time.Duration(msg.Result.Meta.DurationMS) * time.Millisecond
		}
		m.scanCancel = nil
		m.scanLogFollow = true
		m.newScan.SetDisabled(false)

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
		case "c":
			if m.scanActive && m.scanCancel != nil {
				m.scanCancel()
				m.scanStatus = "aborted"
				return m, nil
			}
		}
		if m.active == viewLogs {
			switch msg.String() {
			case "up":
				m.scrollLogs(-1)
				return m, nil
			case "down":
				m.scrollLogs(1)
				return m, nil
			case "pgup":
				m.scrollLogs(-5)
				return m, nil
			case "pgdown":
				m.scrollLogs(5)
				return m, nil
			case "home":
				m.scanLogFollow = false
				m.scanLogTop = 0
				return m, nil
			case "end":
				m.scanLogFollow = true
				return m, nil
			case "enter":
				if m.scanActive && m.scanCancel != nil {
					m.scanCancel()
					m.scanStatus = "aborted"
					return m, nil
				}
			}
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
	runningBody := m.renderRunning(rightWidth-3, topHeight-2, uiTheme)
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

type scanRunner struct {
	events chan scanner.Event
}

type scanEventMsg struct {
	Event scanner.Event
}

type scanDoneMsg struct {
	Result scanner.ScanResult
	Err    error
}

func runScanCmd(ctx context.Context, cfg scanner.ScanConfig, runner *scanRunner) tea.Cmd {
	return func() tea.Msg {
		s := scanner.NewScanner()
		result, err := s.Scan(ctx, cfg, func(evt scanner.Event) {
			select {
			case runner.events <- evt:
			default:
			}
		})
		close(runner.events)
		return scanDoneMsg{Result: result, Err: err}
	}
}

func listenScanCmd(runner *scanRunner) tea.Cmd {
	return func() tea.Msg {
		evt, ok := <-runner.events
		if !ok {
			return nil
		}
		return scanEventMsg{Event: evt}
	}
}

func (m *model) handleScanEvent(evt scanner.Event) {
	switch evt.Kind {
	case scanner.EventScanStart:
		m.scanPortsTotal = evt.PortsTotal
		m.scanHostsTotal = evt.HostsTotal
		m.scanLogs = append(m.scanLogs, "[ scan ] started")
	case scanner.EventHostStart:
		m.scanLogs = append(m.scanLogs, fmt.Sprintf("[ probe ] %s scanning %d ports...", evt.Host, evt.PortsTotal))
	case scanner.EventPort:
		m.scanPortsProbed = evt.PortsProbed
		m.scanPortsTotal = evt.PortsTotal
		if evt.State == scanner.PortOpen {
			m.scanOpenPorts++
		}
		line := fmt.Sprintf("[ %s ] %s:%d", evt.State, evt.Host, evt.Port)
		if evt.Service != "" {
			line += " (" + evt.Service + ")"
		}
		m.scanLogs = append(m.scanLogs, line)
	case scanner.EventHostDone:
		m.scanHostsDone = evt.HostsCompleted
		m.scanHostsTotal = evt.HostsTotal
		m.scanLogs = append(m.scanLogs, fmt.Sprintf("[ done ] %s", evt.Host))
	case scanner.EventScanDone:
		m.scanElapsed = evt.Elapsed
		m.scanLogs = append(m.scanLogs, fmt.Sprintf("[ scan ] done in %s", formatDuration(evt.Elapsed)))
	}
	if len(m.scanLogs) > 200 {
		m.scanLogs = m.scanLogs[len(m.scanLogs)-200:]
	}
	if m.scanLogFollow {
		m.scanLogTop = maxLogTop(m.scanLogs, m.lastLogsHeight())
	}
}

func (m model) renderRunning(width, height int, uiTheme theme.Theme) string {
	if width < 10 || height < 3 {
		return ""
	}

	lines := []string{}
	status := "idle"
	if m.scanActive {
		status = "scanning"
	} else if m.scanStatus != "" {
		status = m.scanStatus
	}
	lines = append(lines, fmt.Sprintf("Status: %s", status))

	if m.scanActive || m.scanStatus != "" {
		lines = append(lines, fmt.Sprintf("Hosts: %d / %d", m.scanHostsDone, m.scanHostsTotal))
		lines = append(lines, fmt.Sprintf("Ports: %d / %d", m.scanPortsProbed, m.scanPortsTotal))
		lines = append(lines, fmt.Sprintf("Open ports: %d", m.scanOpenPorts))
		elapsed := m.scanElapsed
		if m.scanActive {
			elapsed = time.Since(m.scanStart)
		}
		lines = append(lines, fmt.Sprintf("Elapsed: %s", formatDuration(elapsed)))
	}

	if m.scanLastError != "" {
		lines = append(lines, "Error: "+m.scanLastError)
	}

	if m.scanActive {
		lines = append(lines, "[ CANCEL ]  (press C or Enter)")
	}

	lines = append(lines, "")
	logStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(uiTheme.ControlsFg))
	availableLogs := height - len(lines)
	if availableLogs < 1 {
		availableLogs = 1
	}
	logs := m.scanLogs
	top := m.clampLogTop(availableLogs)
	if len(logs) == 0 {
		lines = append(lines, logStyle.Render("No active scan."))
	} else {
		end := top + availableLogs
		if end > len(logs) {
			end = len(logs)
		}
		for _, line := range logs[top:end] {
			lines = append(lines, logStyle.Render(line))
		}
	}
	return lipgloss.NewStyle().Width(width).Height(height).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	secs := int(d.Seconds())
	mins := secs / 60
	secs = secs % 60
	if mins >= 60 {
		h := mins / 60
		m := mins % 60
		return fmt.Sprintf("%dh%02dm%02ds", h, m, secs)
	}
	return fmt.Sprintf("%dm%02ds", mins, secs)
}

func (m *model) scrollLogs(delta int) {
	if len(m.scanLogs) == 0 {
		return
	}
	maxTop := maxLogTop(m.scanLogs, m.lastLogsHeight())
	if maxTop == 0 {
		m.scanLogTop = 0
		m.scanLogFollow = true
		return
	}
	m.scanLogFollow = false
	m.scanLogTop += delta
	if m.scanLogTop < 0 {
		m.scanLogTop = 0
	}
	if m.scanLogTop >= maxTop {
		m.scanLogTop = maxTop
		m.scanLogFollow = true
	}
}

func (m model) clampLogTop(availableLogs int) int {
	if len(m.scanLogs) == 0 {
		return 0
	}
	maxTop := maxLogTop(m.scanLogs, availableLogs)
	if m.scanLogFollow {
		return maxTop
	}
	if m.scanLogTop < 0 {
		return 0
	}
	if m.scanLogTop > maxTop {
		return maxTop
	}
	return m.scanLogTop
}

func (m model) lastLogsHeight() int {
	height := m.height
	if height == 0 {
		return 0
	}
	usableHeight := height - 4
	if usableHeight < 10 {
		usableHeight = 10
	}
	topHeight := (usableHeight * 6) / 10
	if topHeight < 6 {
		topHeight = 6
	}
	return topHeight - 2
}

func maxLogTop(logs []string, availableLogs int) int {
	if availableLogs <= 0 {
		return 0
	}
	if len(logs) <= availableLogs {
		return 0
	}
	return len(logs) - availableLogs
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
