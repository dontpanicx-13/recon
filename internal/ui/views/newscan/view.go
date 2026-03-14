package newscan

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"recon/internal/ui/theme"
)

func (m NewScanModel) View(width, height int, active bool) string {
	m.width = width
	m.height = height

	panel := lipgloss.NewStyle().
		Padding(1, 2).
		Width(width).
		Height(height)

	titleWidth := width - 4
	if titleWidth < 10 {
		titleWidth = 10
	}
	titleFg := m.theme.StatusFg
	titleBg := m.theme.StatusBg
	underline := strings.Repeat(" ", titleWidth)
	if active {
		titleFg = m.theme.AccentFg
		titleBg = m.theme.AccentBg
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
		Render("NEW SCAN")
	underlineLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.AccentBg)).
		Background(lipgloss.Color(m.theme.AppBg)).
		Width(titleWidth).
		Render(underline)

	body := m.renderBody()
	return panel.Render(lipgloss.JoinVertical(lipgloss.Left, title, underlineLine, "", body))
}

func (m NewScanModel) renderBody() string {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
	focusStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(m.theme.AccentFg)).
		Background(lipgloss.Color(m.theme.AccentBg)).
		Padding(0, 1)
	toggleFocusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.AccentFg)).
		Background(lipgloss.Color(m.theme.AccentBg)).
		Padding(0, 1)
	toggleOn := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AccentBg))
	toggleOff := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
	errorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(m.theme.AccentFg)).
		Background(lipgloss.Color(m.theme.AccentBg)).
		Padding(0, 1)

	contentWidth := m.width - 6
	if contentWidth < 10 {
		contentWidth = 10
	}
	m.applyInputWidths(contentWidth)

	if m.pickingFile {
		return m.renderFilePickerOverlay()
	}

	lines := []string{
		labelStyle.Render("Targets"),
		m.renderInput(m.targetsInput, fieldTargets),
		"",
		m.renderSelectValue("File", fieldFilePicker),
		"",
		labelStyle.Render("Ports"),
		m.renderSelect("Mode", m.portsModeLabel(), fieldPortsMode),
	}

	switch m.portsMode {
	case portsPreset:
		lines = append(lines, m.renderSelect("Preset", m.portsPresetLabel(), fieldPortsPreset))
	case portsRange:
		lines = append(lines, m.renderInputRow("Range", m.portsRange, fieldPortsRange))
	case portsList:
		lines = append(lines, m.renderInputRow("List", m.portsList, fieldPortsList))
	}

	lines = append(lines,
		"",
		labelStyle.Render("Profile"),
		m.renderSelect("Profile", m.profileLabel(), fieldProfile),
		m.renderInputRow("Concurrency", m.concurrency, fieldConcurrency),
		m.renderInputRow("Timeout(ms)", m.timeoutMs, fieldTimeout),
		"",
		labelStyle.Render("Options"),
		m.renderToggle("Banner grabbing", m.toggleBanner, fieldBanner, toggleOn, toggleOff, toggleFocusStyle),
		m.renderToggle("TLS analysis", m.toggleTLS, fieldTLS, toggleOn, toggleOff, toggleFocusStyle),
		m.renderToggle("Reverse DNS", m.toggleRDNS, fieldRDNS, toggleOn, toggleOff, toggleFocusStyle),
		"",
		labelStyle.Render("Label (optional)"),
		m.renderInput(m.label, fieldLabel),
		"",
	)

	errors := m.validate()
	startLine := m.renderStart(errors, mutedStyle, focusStyle)
	lines = append(lines, startLine)

	if len(errors) > 0 {
		for _, err := range errors {
			lines = append(lines, errorStyle.Render("- "+err))
		}
	}

	lines = append(lines, "")

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m NewScanModel) renderFilePickerOverlay() string {
	modalWidth := min(70, max(30, m.width-8))
	modalHeight := min(18, max(10, m.height-8))
	m.filePicker.Height = max(8, modalHeight-4)

	var status strings.Builder
	status.WriteString("Pick a file:")
	if m.pickerErr != nil {
		status.Reset()
		status.WriteString(m.filePicker.Styles.DisabledFile.Render(m.pickerErr.Error()))
	} else if m.pickerSelected != "" {
		status.Reset()
		status.WriteString("Selected file: " + m.filePicker.Styles.Selected.Render(m.pickerSelected))
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(m.theme.AccentFg)).
		Render(status.String())
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Render("↑/↓ move  ← back  → open  Enter select  Esc close")
	body := lipgloss.JoinVertical(lipgloss.Left, header, help, "", m.filePicker.View())

	box := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.theme.AccentBg)).
		Background(lipgloss.Color(m.theme.AppBg)).
		Render(body)

	return lipgloss.Place(m.width-4, m.height-4, lipgloss.Center, lipgloss.Center, box)
}

func (m NewScanModel) renderInput(input textinput.Model, field fieldID) string {
	if m.focusedField == field {
		return focusTextInput(input, m.theme)
	}
	return blurTextInput(input)
}

func (m NewScanModel) renderSelect(label, value string, field fieldID) string {
	lineLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(label + ":")
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB")).Background(lipgloss.Color("#374151")).Padding(0, 1)
	if m.focusedField == field {
		valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.AccentFg)).
			Background(lipgloss.Color(m.theme.AccentBg)).
			Padding(0, 1)
	}
	return fmt.Sprintf("%s %s", lineLabel, valueStyle.Render(value))
}

func (m NewScanModel) renderSelectValue(value string, field fieldID) string {
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB")).Background(lipgloss.Color("#374151")).Padding(0, 1)
	if m.focusedField == field {
		valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.AccentFg)).
			Background(lipgloss.Color(m.theme.AccentBg)).
			Padding(0, 1)
	}
	return valueStyle.Render(value)
}

func (m NewScanModel) renderInputRow(label string, input textinput.Model, field fieldID) string {
	lineLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(label + ":")
	return fmt.Sprintf("%s %s", lineLabel, m.renderInput(input, field))
}

func (m NewScanModel) renderToggle(label string, enabled bool, field fieldID, onStyle, offStyle, focusStyle lipgloss.Style) string {
	box := "▯"
	style := offStyle
	if enabled {
		box = "▮"
		style = onStyle
	}
	rendered := style.Render(box + " " + label)
	if m.focusedField == field {
		rendered = focusStyle.Render(box + " " + label)
	}
	return rendered
}

func (m NewScanModel) renderStart(errors []string, mutedStyle, focusStyle lipgloss.Style) string {
	if m.disabled {
		return mutedStyle.Render("Scan in progress...")
	}
	label := "[ START SCAN ]"
	if len(errors) > 0 {
		return mutedStyle.Render(label)
	}
	if m.focusedField == fieldStart {
		marker := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.theme.AccentBg)).
			Render(markerGlyph(m.blinkOn))
		return lipgloss.JoinHorizontal(lipgloss.Left, marker, " ", focusStyle.Render(label))
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(m.theme.AccentFg)).
		Background(lipgloss.Color(m.theme.AccentBg)).
		Padding(0, 1).
		Render(label)
}

func (m NewScanModel) portsModeLabel() string {
	switch m.portsMode {
	case portsRange:
		return "Range"
	case portsList:
		return "List"
	default:
		return "Preset"
	}
}

func (m NewScanModel) portsPresetLabel() string {
	switch m.portsPreset {
	case presetTop1000:
		return "Top 1000"
	case presetAll:
		return "All"
	default:
		return "Top 100"
	}
}

func (m NewScanModel) profileLabel() string {
	switch m.profile {
	case profileQuick:
		return "Quick"
	case profileFull:
		return "Full"
	case profileCustom:
		return "Custom"
	default:
		return "Default"
	}
}

func (m *NewScanModel) applyInputWidths(width int) {
	m.targetsInput.Width = width
	m.portsRange.Width = width
	m.portsList.Width = width
	m.concurrency.Width = max(10, width/2)
	m.timeoutMs.Width = max(10, width/2)
	m.label.Width = width
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func focusTextInput(input textinput.Model, theme theme.Theme) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.AccentFg)).
		Background(lipgloss.Color(theme.AccentBg)).
		Padding(0, 1)
	value := input.View()
	return style.Render(value)
}

func blurTextInput(input textinput.Model) string {
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	style := lipgloss.NewStyle().Background(lipgloss.Color("#374151")).Padding(0, 1)
	value := input.View()
	return style.Render(value)
}

func markerGlyph(on bool) string {
	if on {
		return "█"
	}
	return " "
}
