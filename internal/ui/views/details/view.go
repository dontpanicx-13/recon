package details

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"recon/internal/scanner"
	"recon/internal/ui/theme"
)

func Render(width, height int, uiTheme theme.Theme, m Model) string {
	if width < 4 || height < 4 {
		return ""
	}
	innerW := width - 2
	innerH := height - 2

	styles := uiStyles(uiTheme)
	lines := make([]string, 0, height)
	lines = append(lines, styles.Border.Render("╭"+strings.Repeat("─", innerW)+"╮"))

	headerLines := renderHeader(innerW, styles, m)
	summaryLines := renderSummary(innerW, styles, m)
	hostLeftWidth := max(16, innerW/3)
	hostRightWidth := innerW - hostLeftWidth - 1
	listLines := renderHostList(hostLeftWidth, styles, m)
	detailLines := renderHostDetail(hostRightWidth, styles, m)
	bodyLines := renderBody(innerW, innerH, hostLeftWidth, hostRightWidth, listLines, detailLines, styles, m)

	content := append(headerLines, summaryLines...)
	content = append(content, bodyLines...)
	for len(content) < innerH {
		content = append(content, "")
	}
	if len(content) > innerH {
		content = content[:innerH]
	}
	for _, line := range content {
		lines = append(lines, styles.Border.Render("│")+padANSI(line, innerW)+styles.Border.Render("│"))
	}

	lines = append(lines, styles.Border.Render("╰"+strings.Repeat("─", innerW)+"╯"))
	return strings.Join(lines, "\n")
}

type uiParts struct {
	Border  lipgloss.Style
	Title   lipgloss.Style
	Text    lipgloss.Style
	Hint    lipgloss.Style
	Select  lipgloss.Style
	Key     lipgloss.Style
	Message lipgloss.Style
}

func uiStyles(uiTheme theme.Theme) uiParts {
	return uiParts{
		Border: lipgloss.NewStyle().Foreground(lipgloss.Color(uiTheme.AccentBg)),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(uiTheme.StatusFg)),
		Text: lipgloss.NewStyle().Foreground(lipgloss.Color(uiTheme.StatusFg)),
		Hint: lipgloss.NewStyle().Foreground(lipgloss.Color(uiTheme.ControlsFg)),
		Select: lipgloss.NewStyle().
			Foreground(lipgloss.Color(uiTheme.AppBg)).
			Background(lipgloss.Color(uiTheme.ControlsFg)),
		Key: lipgloss.NewStyle().
			Foreground(lipgloss.Color(uiTheme.AppBg)).
			Background(lipgloss.Color(uiTheme.ControlsFg)).
			Padding(0, 1),
		Message: lipgloss.NewStyle().
			Foreground(lipgloss.Color(uiTheme.AppBg)).
			Background(lipgloss.Color(uiTheme.AccentBg)).
			Padding(0, 1),
	}
}

func renderHeader(width int, styles uiParts, m Model) []string {
	title := styles.Title.Render(fmt.Sprintf(" Scan Detail  —  %s  —  %s ", formatTargets(m.Scan.Config), formatDate(m.Scan.Meta.Date)))
	actions := styles.Hint.Render("[J] Export JSON  [M] Export Markdown  [W/S] Scroll  [Esc] close")
	return []string{
		centerLine(title, width),
		centerLine(actions, width),
		strings.Repeat("─", width),
	}
}

func renderSummary(width int, styles uiParts, m Model) []string {
	cfg := m.Scan.Config
	meta := m.Scan.Meta
	sum := m.Scan.Summary
	lines := []string{
		styles.Text.Render("Config"),
		renderKV(styles, "Targets", formatTargets(cfg)),
		renderKV(styles, "Ports", formatPorts(cfg.Ports)),
		renderKV(styles, "Concurrency", itoa(cfg.Concurrency)) + "  " + renderKV(styles, "Timeout", fmt.Sprintf("%dms", cfg.TimeoutMS)),
		"",
		styles.Text.Render("Options"),
		renderKV(styles, "Banner", onOff(cfg.BannerGrabbing)) + "  " + renderKV(styles, "TLS", onOff(cfg.TLSAnalysis)) + "  " + renderKV(styles, "rDNS", onOff(cfg.ReverseDNS)),
		"",
		styles.Text.Render("Meta"),
		renderKV(styles, "Status", orDash(meta.Status)) + "  " + renderKV(styles, "Date", formatDate(meta.Date)) + "  " + renderKV(styles, "Duration", formatDuration(meta.DurationMS)),
		"",
		styles.Text.Render("Summary"),
		renderKV(styles, "Hosts", fmt.Sprintf("%d/%d", sum.HostsFound, sum.HostsTotal)) + "  " + renderKV(styles, "Open ports", itoa(sum.OpenPorts)) + "  " + renderKV(styles, "Probed", itoa(sum.PortsProbed)),
		strings.Repeat("─", width),
	}
	if m.Message != "" {
		lines = append([]string{styles.Message.Render(m.Message)}, lines...)
	}
	return lines
}

func renderBody(width, height, leftW, rightW int, left, right []string, styles uiParts, m Model) []string {
	rows := height - 6
	if rows < 1 {
		rows = 1
	}
	if len(left) < rows {
		for len(left) < rows {
			left = append(left, "")
		}
	}
	if len(right) < rows {
		for len(right) < rows {
			right = append(right, "")
		}
	}
	if len(left) > rows {
		left = left[:rows]
	}
	if len(right) > rows {
		// Keep for scrolling
	}
	top := m.DetailTop
	maxTop := max(0, len(right)-rows)
	if top > maxTop {
		top = maxTop
	}
	right = right[top:]
	if len(right) > rows {
		right = right[:rows]
	}
	if len(right) < rows {
		for len(right) < rows {
			right = append(right, "")
		}
	}
	lines := make([]string, 0, rows)
	for i := 0; i < rows; i++ {
		l := padANSI(left[i], leftW)
		r := padANSI(right[i], rightW)
		lines = append(lines, l+"│"+r)
	}
	return lines
}

func renderHostList(width int, styles uiParts, m Model) []string {
	lines := []string{styles.Text.Render("Host list")}
	if len(m.Hosts) == 0 {
		lines = append(lines, styles.Text.Render("No hosts"))
		return lines
	}
	for i, host := range m.Hosts {
		label := host.IP
		if host.Hostname != "" {
			label += "  " + host.Hostname
		}
		label = "  " + trimInline(label, width-2)
		if i == m.Selected {
			lines = append(lines, styles.Select.Render(padANSI(label, width)))
			continue
		}
		lines = append(lines, styles.Text.Render(label))
	}
	return lines
}

func renderHostDetail(width int, styles uiParts, m Model) []string {
	host, ok := m.SelectedHost()
	if !ok {
		return []string{styles.Text.Render("No host selected")}
	}
	lines := []string{}
	header := host.IP
	if host.Hostname != "" {
		header += "  " + host.Hostname
	}
	lines = append(lines, styles.Text.Render("## "+header))
	lines = append(lines, styles.Text.Render(""))
	lines = append(lines, styles.Text.Render("| Port | State | Service | Banner |"))
	ports := append([]scanner.PortState(nil), host.Ports...)
	sort.Slice(ports, func(i, j int) bool { return ports[i].Port < ports[j].Port })
	for _, port := range ports {
		service := orDash(port.ServiceGuess)
		banner := "—"
		if port.Banner != nil && *port.Banner != "" {
			banner = trimInline(*port.Banner, max(10, width-30))
		}
		lines = append(lines, styles.Text.Render(fmt.Sprintf("| %d | %s | %s | %s |", port.Port, port.State, service, banner)))
	}
	lines = append(lines, styles.Text.Render(""))
	for _, port := range ports {
		if port.TLS == nil {
			continue
		}
		lines = append(lines, styles.Text.Render(fmt.Sprintf("### TLS — %d", port.Port)))
		lines = append(lines, styles.Text.Render(fmt.Sprintf("CN: %s", orDash(port.TLS.CommonName))))
		lines = append(lines, styles.Text.Render(fmt.Sprintf("Issuer: %s", orDash(port.TLS.Issuer))))
		lines = append(lines, styles.Text.Render(fmt.Sprintf("Expires: %s", orDash(port.TLS.Expires))))
		lines = append(lines, styles.Text.Render(fmt.Sprintf("Version: %s / %s", orDash(port.TLS.TLSVersion), orDash(port.TLS.Cipher))))
		if port.TLS.Note != "" {
			lines = append(lines, styles.Text.Render(fmt.Sprintf("Note: %s", trimInline(port.TLS.Note, width-4))))
		}
		lines = append(lines, styles.Text.Render(""))
	}
	return lines
}

func padANSI(value string, width int) string {
	diff := width - ansi.StringWidth(value)
	if diff <= 0 {
		return ansi.Truncate(value, width, "")
	}
	return value + strings.Repeat(" ", diff)
}

func formatTargets(cfg scanner.ScanConfig) string {
	if len(cfg.Targets) == 0 {
		return "—"
	}
	if len(cfg.Targets) <= 3 {
		return strings.Join(cfg.Targets, ", ")
	}
	return fmt.Sprintf("%d targets", len(cfg.Targets))
}

func formatPorts(ports []int) string {
	if len(ports) == 0 {
		return "—"
	}
	if len(ports) <= 10 {
		out := make([]string, 0, len(ports))
		for _, p := range ports {
			out = append(out, fmt.Sprintf("%d", p))
		}
		return strings.Join(out, ", ")
	}
	return fmt.Sprintf("%d ports", len(ports))
}

func formatDate(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t.Format("2006-01-02")
	}
	return trimInline(value, 10)
}

func formatDuration(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	d := time.Duration(ms) * time.Millisecond
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

func trimInline(value string, maxLen int) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.TrimSpace(value)
	if maxLen <= 0 {
		return ""
	}
	if ansi.StringWidth(value) <= maxLen {
		return value
	}
	return ansi.Truncate(value, maxLen-1, "") + "…"
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

func orDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "—"
	}
	return value
}

func itoa(v int) string {
	return fmt.Sprintf("%d", v)
}

func centerLine(value string, width int) string {
	vw := ansi.StringWidth(value)
	if vw >= width {
		return ansi.Truncate(value, width, "")
	}
	left := (width - vw) / 2
	right := width - vw - left
	return strings.Repeat(" ", left) + value + strings.Repeat(" ", right)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderKV(styles uiParts, label, value string) string {
	if value == "" {
		value = "—"
	}
	key := styles.Key.Render(label)
	return key + " " + styles.Text.Render(value)
}
