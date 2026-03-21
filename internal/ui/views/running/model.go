package running

import (
	"context"
	"strings"
	"time"

	"recon/internal/scanner"
)

type Model struct {
	scanActive        bool
	scanStatus        string
	scanStart         time.Time
	scanElapsed       time.Duration
	scanPortsTotal    int
	scanPortsProbed   int
	scanHostsTotal    int
	scanHostsDone     int
	scanOpenPorts     int
	scanLogs          []string
	scanLastError     string
	scanLogTop        int
	scanLogFollow     bool
	scanCancel        context.CancelFunc
	lastAvailableLogs int
	hostOpenPorts     map[string]int
}

func NewModel() Model {
	return Model{
		scanLogFollow: true,
	}
}

func (m Model) Active() bool {
	return m.scanActive
}

func (m *Model) StartScan(cancel context.CancelFunc, preLogs []string) {
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
	m.hostOpenPorts = make(map[string]int)

	for _, line := range preLogs {
		m.addLog(line)
	}
}

func (m *Model) HandleScanEvent(evt scanner.Event) {
	switch evt.Kind {
	case scanner.EventScanStart:
		m.scanPortsTotal = evt.PortsTotal
		m.scanHostsTotal = evt.HostsTotal
		m.addLog("[ scan ] started")
	case scanner.EventHostStart:
		m.addLog("[ probe ] " + evt.Host + "  scanning " + itoa(evt.PortsTotal) + " ports...")
	case scanner.EventPort:
		m.scanPortsProbed = evt.PortsProbed
		m.scanPortsTotal = evt.PortsTotal
		if evt.State == scanner.PortOpen {
			m.scanOpenPorts++
			m.hostOpenPorts[evt.Host] = m.hostOpenPorts[evt.Host] + 1
		}
		switch evt.State {
		case scanner.PortOpen:
			m.addLog("[ open ] " + evt.Host + ":" + itoa(evt.Port))
		case scanner.PortClosed:
			m.addLog("[ closed ] " + evt.Host + ":" + itoa(evt.Port))
		case scanner.PortFiltered:
			m.addLog("[ filtered ] " + evt.Host + ":" + itoa(evt.Port) + "  (timeout)")
		default:
			m.addLog("[ " + evt.State + " ] " + evt.Host + ":" + itoa(evt.Port))
		}
		if evt.Banner != nil && *evt.Banner != "" {
			m.addLog("[ service ] " + evt.Host + ":" + itoa(evt.Port) + "  \u2192  " + *evt.Banner)
		} else if evt.Service != "" {
			m.addLog("[ service ] " + evt.Host + ":" + itoa(evt.Port) + "  \u2192  " + evt.Service)
		}
		if evt.TLS != nil && (evt.TLS.TLSVersion != "" || evt.TLS.CommonName != "" || evt.TLS.Note != "") {
			desc := evt.TLS.TLSVersion
			if desc == "" {
				desc = evt.TLS.CommonName
			}
			if desc == "" {
				desc = evt.TLS.Note
			}
			if desc == "" {
				desc = "TLS detected"
			}
			m.addLog("[ tls ] " + evt.Host + ":" + itoa(evt.Port) + "  \u2192  " + desc)
		}
	case scanner.EventHostDone:
		m.scanHostsDone = evt.HostsCompleted
		m.scanHostsTotal = evt.HostsTotal
		openCount := m.hostOpenPorts[evt.Host]
		m.addLog("[ done ] " + evt.Host + "  \u2014  " + itoa(openCount) + " open ports")
	case scanner.EventScanDone:
		m.scanElapsed = evt.Elapsed
		m.addLog("[ scan ] done in " + formatDuration(evt.Elapsed))
	}
}

func (m *Model) HandleScanDone(result scanner.ScanResult, err error) {
	m.scanActive = false
	if err != nil {
		m.scanStatus = "failed"
		m.scanLastError = err.Error()
	} else {
		m.scanStatus = result.Meta.Status
		m.scanElapsed = time.Duration(result.Meta.DurationMS) * time.Millisecond
	}
	m.scanCancel = nil
	m.scanLogFollow = true
}

func (m *Model) CancelScan() {
	if m.scanActive && m.scanCancel != nil {
		m.scanCancel()
		m.scanStatus = "aborted"
	}
}

func (m *Model) addLog(line string) {
	line = sanitizeLogLine(line)
	if line == "" {
		return
	}
	m.scanLogs = append(m.scanLogs, line)
	if len(m.scanLogs) > 200 {
		m.scanLogs = m.scanLogs[len(m.scanLogs)-200:]
	}
}

func sanitizeLogLine(line string) string {
	line = strings.ReplaceAll(line, "\r", "")
	line = strings.ReplaceAll(line, "\n", " ")
	line = strings.TrimSpace(line)
	if len(line) > 200 {
		line = line[:200]
	}
	return line
}
