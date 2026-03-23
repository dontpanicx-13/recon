package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"recon/internal/logger"
	"recon/internal/scanner"
	"recon/internal/store"
	"recon/internal/tlsinfo"
	"recon/internal/ui/theme"
	"recon/internal/ui/views/history"
	"recon/internal/ui/views/newscan"
	"recon/internal/ui/views/running"
	"recon/internal/ui/views/statusbar"
)

type model struct {
	width   int
	height  int
	newScan newscan.NewScanModel
	running running.Model
	history history.Model
	status  statusbar.Model
	active  viewID

	scanRunner *scanRunner
	scanLabel  string
	log        *logger.Logger
	store      *store.Store
}

type viewID int

const (
	viewNewScan viewID = iota
	viewLogs
	viewHistory
)

func InitalModel(toolName, toolVersion string) model {
	log, _ := logger.NewDefault()
	historyModel := history.NewModel()
	storeHandle, err := store.Default()
	if err != nil {
		historyModel.SetError(err)
	} else {
		manifest, loadErr := storeHandle.LoadManifest()
		if loadErr != nil {
			historyModel.SetError(loadErr)
		} else {
			historyModel.SetItems(manifest.Scans)
		}
	}
	return model{
		newScan: newscan.NewModel(),
		running: running.NewModel(),
		history: historyModel,
		status:  statusbar.NewModel(toolName, toolVersion),
		active:  viewNewScan,
		log:     log,
		store:   storeHandle,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return tea.Batch(m.status.Init(), newscan.BlinkCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case newscan.StartScanMsg:
		if m.running.Active() {
			return m, nil
		}
		runner := &scanRunner{events: make(chan scanner.Event, 256)}
		ctx, cancel := context.WithCancel(context.Background())
		m.scanRunner = runner
		m.scanLabel = msg.Label
		m.running.StartScan(cancel, msg.PreLogs)
		m.active = viewLogs
		m.newScan.SetDisabled(true)
		m.logScanStart(msg.Config)
		for _, line := range msg.PreLogs {
			m.logLine("dns", line)
		}

		return m, tea.Batch(runScanCmd(ctx, msg.Config, runner), listenScanCmd(runner))
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case scanEventMsg:
		m.running.HandleScanEvent(msg.Event)
		m.logScanEvent(msg.Event)
		if m.scanRunner != nil {
			return m, listenScanCmd(m.scanRunner)
		}
	case scanDoneMsg:
		m.running.HandleScanDone(msg.Result, msg.Err)
		m.logScanDone(msg.Result, msg.Err)
		m.persistScan(msg.Result, msg.Err)
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
		}
	}

	var cmd tea.Cmd
	var statusCmd tea.Cmd
	m.status, statusCmd = m.status.Update(msg)
	if _, ok := msg.(tea.KeyMsg); !ok || m.active == viewNewScan {
		m.newScan, cmd = m.newScan.Update(msg)
	}
	if _, ok := msg.(tea.KeyMsg); !ok || m.active == viewLogs {
		m.running, _ = m.running.Update(msg)
	}
	if _, ok := msg.(tea.KeyMsg); !ok || m.active == viewHistory {
		m.history, _ = m.history.Update(msg)
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
	panelPadV := 2
	panelChrome := 3 // title + underline + spacer
	historyPanel := lipgloss.NewStyle().Padding(0, 2)
	historyPadV := 0
	historyChrome := 2 // title + underline (no spacer)

	runningTitle := m.renderPanelTitle("RUNNING / LOGS", rightWidth-1, uiTheme, m.active == viewLogs)
	runningInnerWidth := rightWidth - 5
	if runningInnerWidth < 10 {
		runningInnerWidth = 10
	}
	runningBodyHeight := topHeight - panelPadV - panelChrome
	if runningBodyHeight < 1 {
		runningBodyHeight = 1
	}
	runningBody := m.running.View(runningInnerWidth, runningBodyHeight, uiTheme)
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
	historyBodyHeight := historyHeight - historyPadV - historyChrome
	if historyBodyHeight < 1 {
		historyBodyHeight = 1
	}
	historyBody := m.history.View(usableWidth-4, historyBodyHeight, uiTheme, m.active == viewHistory)
	history := historyPanel.
		Width(usableWidth).
		Height(historyHeight).
		Render(lipgloss.JoinVertical(lipgloss.Left, historyTitle, historyBody))

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

func (m *model) logScanStart(cfg scanner.ScanConfig) {
	if m.log == nil {
		return
	}
	m.log.Info("scan_start", map[string]any{
		"targets":       cfg.Targets,
		"ports":         cfg.Ports,
		"profile":       cfg.Profile,
		"concurrency":   cfg.Concurrency,
		"timeout_ms":    cfg.TimeoutMS,
		"banner_grab":   cfg.BannerGrabbing,
		"tls_analysis":  cfg.TLSAnalysis,
		"reverse_dns":   cfg.ReverseDNS,
		"targets_count": len(cfg.Targets),
		"ports_count":   len(cfg.Ports),
	})
}

func (m *model) logScanEvent(evt scanner.Event) {
	if m.log == nil {
		return
	}
	fields := map[string]any{
		"host": evt.Host,
		"port": evt.Port,
		"kind": string(evt.Kind),
	}
	switch evt.Kind {
	case scanner.EventScanStart:
		fields["ports_total"] = evt.PortsTotal
		fields["hosts_total"] = evt.HostsTotal
	case scanner.EventHostStart:
		fields["ports_total"] = evt.PortsTotal
	case scanner.EventPort:
		fields["state"] = evt.State
		fields["service"] = evt.Service
		fields["ports_probed"] = evt.PortsProbed
		fields["ports_total"] = evt.PortsTotal
		if evt.Err != nil {
			fields["error"] = evt.Err.Error()
			m.log.Warn("port_error", fields)
			return
		}
		if evt.TLS != nil && evt.TLS.Note != "" {
			fields["tls_note"] = evt.TLS.Note
			m.log.Warn("tls_note", fields)
			return
		}
	case scanner.EventHostDone:
		fields["hosts_completed"] = evt.HostsCompleted
		fields["hosts_total"] = evt.HostsTotal
	case scanner.EventScanDone:
		fields["hosts_completed"] = evt.HostsCompleted
		fields["hosts_total"] = evt.HostsTotal
		fields["ports_probed"] = evt.PortsProbed
		fields["ports_total"] = evt.PortsTotal
		fields["elapsed_ms"] = int(evt.Elapsed.Milliseconds())
	}
	m.log.Info("scan_event", fields)
}

func (m *model) logScanDone(result scanner.ScanResult, err error) {
	if m.log == nil {
		return
	}
	if err != nil {
		m.log.Error("scan_failed", map[string]any{"error": err.Error()})
		return
	}
	m.log.Info("scan_done", map[string]any{
		"status":       result.Meta.Status,
		"duration_ms":  result.Meta.DurationMS,
		"hosts_total":  result.Summary.HostsTotal,
		"hosts_found":  result.Summary.HostsFound,
		"ports_total":  result.Summary.PortsTotal,
		"ports_probed": result.Summary.PortsProbed,
		"open_ports":   result.Summary.OpenPorts,
	})
}

func (m *model) persistScan(result scanner.ScanResult, err error) {
	if err != nil {
		return
	}
	if m.store == nil {
		return
	}
	if result.ScanID == "" {
		return
	}
	if _, saveErr := m.store.SaveScan(result, m.scanLabel); saveErr != nil {
		m.history.SetError(saveErr)
		if m.log != nil {
			m.log.Error("store_save_failed", map[string]any{"error": saveErr.Error()})
		}
		return
	}
	manifest, loadErr := m.store.LoadManifest()
	if loadErr != nil {
		m.history.SetError(loadErr)
		if m.log != nil {
			m.log.Error("store_manifest_failed", map[string]any{"error": loadErr.Error()})
		}
		return
	}
	m.history.SetError(nil)
	m.history.SetItems(manifest.Scans)
}

func (m *model) logLine(event, line string) {
	if m.log == nil {
		return
	}
	m.log.Info(event, map[string]any{"message": line})
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
		s.TLSInspector = tlsinfo.Inspect
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
