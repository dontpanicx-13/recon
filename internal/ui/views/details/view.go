package details

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"recon/internal/ui/theme"
)

func Render(width, height int, uiTheme theme.Theme) string {
	if width < 4 || height < 4 {
		return ""
	}
	innerW := width - 2
	innerH := height - 2

	borderStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.AccentBg))
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(uiTheme.AccentFg)).
		Background(lipgloss.Color(uiTheme.AccentBg))
	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.StatusFg))
	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(uiTheme.ControlsFg))

	lines := make([]string, 0, height)
	lines = append(lines, borderStyle.Render("╭"+strings.Repeat("─", innerW)+"╮"))

	content := []string{
		titleStyle.Render(" DETAIL "),
		"",
		textStyle.Render("Hola, esta es la vista detalle."),
		"",
		hintStyle.Render("Esc para cerrar"),
	}
	for len(content) < innerH {
		content = append(content, "")
	}
	if len(content) > innerH {
		content = content[:innerH]
	}
	for _, line := range content {
		lines = append(lines, borderStyle.Render("│")+padANSI(line, innerW)+borderStyle.Render("│"))
	}

	lines = append(lines, borderStyle.Render("╰"+strings.Repeat("─", innerW)+"╯"))
	return strings.Join(lines, "\n")
}

func padANSI(value string, width int) string {
	diff := width - ansi.StringWidth(value)
	if diff <= 0 {
		return ansi.Truncate(value, width, "")
	}
	return value + strings.Repeat(" ", diff)
}
