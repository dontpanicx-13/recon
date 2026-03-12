package statusbar

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"recon/internal/ui/theme"
)

func (m Model) View(width int, uiTheme theme.Theme) string {
	sp := m.spinner
	sp.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.AccentFg)).
		Background(lipgloss.Color(uiTheme.AccentBg))

	toolStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.StatusFg)).
		Background(lipgloss.Color(uiTheme.StatusBg))
	uptimeStyle := toolStyle
	nowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.AccentFg)).
		Background(lipgloss.Color(uiTheme.AccentBg))
	fillStyle := toolStyle

	spinnerSegment := strings.TrimRight(sp.View(), " ")
	toolSegment := toolStyle.Render(" " + m.toolName + " " + m.toolVersion + " ")
	left := spinnerSegment + toolSegment

	uptimeSegment := uptimeStyle.Render(" Uptime " + formatUptime(time.Since(m.started)) + " ")
	nowSegment := nowStyle.Render(" Now " + time.Now().Format("15:04:05") + " ")
	right := uptimeSegment + nowSegment

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spaceCount := width - leftWidth - rightWidth
	if spaceCount < 0 {
		spaceCount = 0
	}

	statusLine := left + fillStyle.Render(strings.Repeat(" ", spaceCount)) + right
	lineWidth := lipgloss.Width(statusLine)
	if lineWidth < width {
		statusLine += fillStyle.Render(strings.Repeat(" ", width-lineWidth))
	}
	return statusLine
}

func formatUptime(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	secs := int(d.Seconds())
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	if h > 0 {
		return formatTwo(h) + ":" + formatTwo(m) + ":" + formatTwo(s)
	}
	return formatTwo(m) + ":" + formatTwo(s)
}

func formatTwo(v int) string {
	if v < 10 {
		return "0" + strconv.Itoa(v)
	}
	return strconv.Itoa(v)
}
