package history

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"recon/internal/store"
	"recon/internal/ui/theme"
)

func (m *Model) View(width, height int, uiTheme theme.Theme, active bool) string {
	m.lastWidth = width
	m.lastHeight = height
	m.clampTop()

	if width < 40 || height < 3 {
		return lipgloss.NewStyle().Width(width).Height(height).Render("Window too small.")
	}

	if m.errMessage != "" {
		errStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(uiTheme.AccentBg)).
			Bold(true)
		return lipgloss.NewStyle().Width(width).Height(height).Render(errStyle.Render("History error: " + m.errMessage))
	}

	if len(m.items) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(uiTheme.ControlsFg))
		return lipgloss.NewStyle().Width(width).Height(height).Render(emptyStyle.Render("No scans yet."))
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(uiTheme.StatusFg)).
		Background(lipgloss.Color(uiTheme.StatusBg))
	rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(uiTheme.ControlsFg))
	selectedRowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.AppBg)).
		Background(lipgloss.Color(uiTheme.ControlsFg))

	cols := calcColumns(width)
	header := renderRow(cols, historyRow{
		ID:       "#",
		Label:    "Label",
		Targets:  "Targets",
		Date:     "Date",
		Duration: "Duration",
		Hosts:    "Hosts",
		Open:     "Open",
		Status:   "Status",
	}, headerStyle, false)

	lines := []string{header}
	visible := max(1, height-1)
	start := m.top
	end := min(len(m.items), start+visible)
	for i := start; i < end; i++ {
		item := m.items[i]
		row := historyRowFromItem(item)
		style := rowStyle
		if i == m.selected {
			style = selectedRowStyle
		}
		lines = append(lines, renderRow(cols, row, style, true))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return lipgloss.NewStyle().Width(width).Height(height).Render(body)
}

type historyRow struct {
	ID       string
	Label    string
	Targets  string
	Date     string
	Duration string
	Hosts    string
	Open     string
	Status   string
}

type columns struct {
	ID       int
	Label    int
	Targets  int
	Date     int
	Duration int
	Hosts    int
	Open     int
	Status   int
}

func calcColumns(width int) columns {
	id := 15
	date := 16
	duration := 9
	hosts := 7
	open := 6
	status := 10

	fixed := id + date + duration + hosts + open + status
	separators := 7
	flex := width - fixed - separators
	if flex < 10 {
		flex = 10
	}
	label := max(5, (flex*2)/5)
	targets := max(5, flex-label)

	return columns{
		ID:       id,
		Label:    label,
		Targets:  targets,
		Date:     date,
		Duration: duration,
		Hosts:    hosts,
		Open:     open,
		Status:   status,
	}
}

func renderRow(cols columns, row historyRow, style lipgloss.Style, truncate bool) string {
	cells := []string{
		cell(row.ID, cols.ID, style, truncate),
		cell(row.Label, cols.Label, style, truncate),
		cell(row.Targets, cols.Targets, style, truncate),
		cell(row.Date, cols.Date, style, truncate),
		cell(row.Duration, cols.Duration, style, truncate),
		cell(row.Hosts, cols.Hosts, style, truncate),
		cell(row.Open, cols.Open, style, truncate),
		cell(row.Status, cols.Status, style, truncate),
	}
	return strings.Join(cells, " ")
}

func cell(value string, width int, style lipgloss.Style, truncate bool) string {
	if truncate {
		value = truncateText(value, width)
	}
	return style.Width(width).Render(value)
}

func historyRowFromItem(item store.ManifestItem) historyRow {
	label := strings.TrimSpace(item.Label)
	if label == "" {
		label = strings.TrimSpace(item.TargetsText)
	}
	if label == "" {
		label = strings.Join(item.Targets, ", ")
	}
	return historyRow{
		ID:       shortID(item.ScanID),
		Label:    label,
		Targets:  strings.Join(item.Targets, ", "),
		Date:     formatDate(item.Date),
		Duration: formatDuration(item.DurationMS),
		Hosts:    fmt.Sprintf("%d", item.HostsFound),
		Open:     fmt.Sprintf("%d", item.OpenPorts),
		Status:   item.Status,
	}
}

func shortID(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 15 {
		return value
	}
	return value[:15]
}

func formatDate(value string) string {
	if value == "" {
		return "-"
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return parsed.Format("2006-01-02 15:04")
}

func formatDuration(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	d := time.Duration(ms) * time.Millisecond
	secs := int(d.Seconds())
	h := secs / 3600
	secs = secs % 3600
	m := secs / 60
	s := secs % 60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
	}
	return fmt.Sprintf("%dm %02ds", m, s)
}

func truncateText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}
	if width <= 1 {
		return string([]rune(value)[:1])
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return string(runes[:width-1]) + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
