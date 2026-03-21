package running

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"recon/internal/ui/theme"
)

func (m *Model) View(width, height int, uiTheme theme.Theme) string {
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

	elapsed := m.scanElapsed
	if m.scanActive {
		elapsed = time.Since(m.scanStart)
	}

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.AppBg)).
		Background(lipgloss.Color(uiTheme.ControlsFg)).
		Padding(0, 1)
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.StatusFg)).
		Background(lipgloss.Color(uiTheme.StatusBg)).
		Padding(0, 1)

	labelWidth := 12
	if width < 36 {
		labelWidth = 8
	}
	valueWidth := max(10, width-labelWidth-2)

	rows := []struct {
		Label string
		Value string
	}{
		{"STATUS", status},
		{"HOSTS", fmt.Sprintf("%d / %d", m.scanHostsDone, m.scanHostsTotal)},
		{"PORTS", fmt.Sprintf("%d / %d", m.scanPortsProbed, m.scanPortsTotal)},
		{"OPEN", fmt.Sprintf("%d", m.scanOpenPorts)},
		{"ELAPSED", formatDuration(elapsed)},
	}

	for _, row := range rows {
		label := labelStyle.Width(labelWidth).Render(row.Label)
		value := valueStyle.Width(valueWidth).Render(row.Value)
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Left, label, value))
	}

	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.StatusBg)).
		Render(strings.Repeat("─", width))
	lines = append(lines, sep)

	if m.scanLastError != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(uiTheme.AccentBg)).
			Bold(true)
		lines = append(lines, errorStyle.Render("Error: "+m.scanLastError))
	}

	if m.scanActive {
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(uiTheme.StatusFg))
		lines = append(lines, helpStyle.Render("Press C or Enter to cancel"))
	}

	lines = append(lines, "")
	logStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.ControlsFg))
	availableLogs := height - len(lines)
	if availableLogs < 1 {
		availableLogs = 1
	}
	m.lastAvailableLogs = availableLogs
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
			lines = append(lines, logStyle.Render(truncateLine(line, width)))
		}
	}
	return lipgloss.NewStyle().Width(width).Height(height).Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m *Model) scrollLogs(delta int) {
	if len(m.scanLogs) == 0 {
		return
	}
	available := m.lastAvailableLogs
	if available <= 0 {
		available = 1
	}
	maxTop := maxLogTop(m.scanLogs, available)
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

func (m Model) clampLogTop(availableLogs int) int {
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

func maxLogTop(logs []string, availableLogs int) int {
	if availableLogs <= 0 {
		return 0
	}
	if len(logs) <= availableLogs {
		return 0
	}
	return len(logs) - availableLogs
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncateLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(line) <= width {
		return line
	}
	if width <= 1 {
		r := []rune(line)
		if len(r) == 0 {
			return ""
		}
		return string(r[:1])
	}
	r := []rune(line)
	if len(r) <= width {
		return line
	}
	return string(r[:width-1]) + "…"
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

func itoa(v int) string {
	return strconv.Itoa(v)
}
