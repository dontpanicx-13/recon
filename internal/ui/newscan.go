package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type portsMode int

const (
	portsPreset portsMode = iota
	portsRange
	portsList
)

type portsPresetKind int

const (
	presetTop100 portsPresetKind = iota
	presetTop1000
	presetAll
)

type profileKind int

const (
	profileQuick profileKind = iota
	profileDefault
	profileFull
	profileCustom
)

type fieldID int

const (
	fieldTargets fieldID = iota
	fieldPortsMode
	fieldPortsPreset
	fieldPortsRange
	fieldPortsList
	fieldProfile
	fieldConcurrency
	fieldTimeout
	fieldBanner
	fieldTLS
	fieldRDNS
	fieldLabel
	fieldStart
)

type NewScanModel struct {
	targetsInput textinput.Model
	portsRange   textinput.Model
	portsList    textinput.Model
	concurrency  textinput.Model
	timeoutMs    textinput.Model
	label        textinput.Model

	portsMode   portsMode
	portsPreset portsPresetKind
	profile     profileKind

	toggleBanner bool
	toggleTLS    bool
	toggleRDNS   bool

	focusedField fieldID
	disabled     bool
	lastErrors   []string
	width        int
	height       int
}

func NewNewScanModel() NewScanModel {
	m := NewScanModel{
		portsMode:    portsPreset,
		portsPreset:  presetTop100,
		profile:      profileDefault,
		focusedField: fieldTargets,
	}

	m.targetsInput = newTextInput("Targets", "")
	m.portsRange = newTextInput("Range", "1-1024")
	m.portsList = newTextInput("List", "22,80,443")
	m.concurrency = newTextInput("Concurrency", "100")
	m.timeoutMs = newTextInput("Timeout", "1000")
	m.label = newTextInput("Label", "")

	m.applyFocus()
	return m
}

func (m NewScanModel) Update(msg tea.Msg) (NewScanModel, tea.Cmd) {
	if m.disabled {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.focusedField = m.nextField(true)
			m.applyFocus()
			return m, nil
		case "shift+tab":
			m.focusedField = m.nextField(false)
			m.applyFocus()
			return m, nil
		case "left":
			if m.isSelectField(m.focusedField) {
				m.handleSelect(-1)
				return m, nil
			}
		case "right":
			if m.isSelectField(m.focusedField) {
				m.handleSelect(1)
				return m, nil
			}
		case "enter":
			if m.focusedField == fieldStart {
				m.lastErrors = m.validate()
				return m, nil
			}
		case " ":
			if m.isToggleField(m.focusedField) {
				m.handleToggle()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m, cmd = m.updateFocusedInput(msg)
	return m, cmd
}

func (m NewScanModel) View(width, height int) string {
	m.width = width
	m.height = height

	panel := lipgloss.NewStyle().
		Padding(1, 2).
		Width(width).
		Height(height)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#4C1D95")).
		Padding(0, 1).
		Render("NEW SCAN")

	body := m.renderBody()

	return panel.Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
}

/*F59E0B 10B981*/
func (m NewScanModel) renderBody() string {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB"))
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
	focusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#111827")).Background(lipgloss.Color("#F59E0B")).Padding(0, 1)
	toggleOn := lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	toggleOff := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
	errorStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#DC2626")).
		Padding(0, 1)

	contentWidth := m.width - 6
	if contentWidth < 10 {
		contentWidth = 10
	}
	m.applyInputWidths(contentWidth)

	lines := []string{
		labelStyle.Render("Targets"),
		m.renderInput(m.targetsInput, fieldTargets),
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
		m.renderToggle("Banner grabbing", m.toggleBanner, fieldBanner, toggleOn, toggleOff, focusStyle),
		m.renderToggle("TLS analysis", m.toggleTLS, fieldTLS, toggleOn, toggleOff, focusStyle),
		m.renderToggle("Reverse DNS", m.toggleRDNS, fieldRDNS, toggleOn, toggleOff, focusStyle),
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

	controls := mutedStyle.Render("Tab/Shift+Tab (move) • Left/Right (switch) • Space (toggle) • Enter (start)")
	lines = append(lines, "", controls)

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m NewScanModel) renderInput(input textinput.Model, field fieldID) string {
	if m.focusedField == field {
		return focusTextInput(input)
	}
	return blurTextInput(input)
}

func (m NewScanModel) renderSelect(label, value string, field fieldID) string {
	lineLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(label + ":")
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB")).Background(lipgloss.Color("#374151")).Padding(0, 1)
	if m.focusedField == field {
		valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#111827")).Background(lipgloss.Color("#F59E0B")).Padding(0, 1)
	}
	return fmt.Sprintf("%s %s", lineLabel, valueStyle.Render(value))
}

func (m NewScanModel) renderInputRow(label string, input textinput.Model, field fieldID) string {
	lineLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(label + ":")
	return fmt.Sprintf("%s %s", lineLabel, m.renderInput(input, field))
}

func (m NewScanModel) renderToggle(label string, enabled bool, field fieldID, onStyle, offStyle, focusStyle lipgloss.Style) string {
	box := "[ ]"
	style := offStyle
	if enabled {
		box = "■"
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
		return focusStyle.Render(label)
	}
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E5E7EB")).Background(lipgloss.Color("#10B981")).Padding(0, 1).Render(label)
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

func (m NewScanModel) nextField(forward bool) fieldID {
	order := []fieldID{
		fieldTargets,
		fieldPortsMode,
		fieldPortsPreset,
		fieldPortsRange,
		fieldPortsList,
		fieldProfile,
		fieldConcurrency,
		fieldTimeout,
		fieldBanner,
		fieldTLS,
		fieldRDNS,
		fieldLabel,
		fieldStart,
	}

	start := 0
	for i, field := range order {
		if field == m.focusedField {
			start = i
			break
		}
	}

	step := 1
	if !forward {
		step = -1
	}

	for i := 1; i <= len(order); i++ {
		idx := (start + step*i + len(order)) % len(order)
		if m.isFieldActive(order[idx]) {
			return order[idx]
		}
	}
	return m.focusedField
}

func (m NewScanModel) isFieldActive(field fieldID) bool {
	switch field {
	case fieldPortsPreset:
		return m.portsMode == portsPreset
	case fieldPortsRange:
		return m.portsMode == portsRange
	case fieldPortsList:
		return m.portsMode == portsList
	default:
		return true
	}
}

func (m NewScanModel) isSelectField(field fieldID) bool {
	return field == fieldPortsMode || field == fieldPortsPreset || field == fieldProfile
}

func (m NewScanModel) isToggleField(field fieldID) bool {
	return field == fieldBanner || field == fieldTLS || field == fieldRDNS
}

func (m *NewScanModel) handleSelect(delta int) {
	switch m.focusedField {
	case fieldPortsMode:
		m.portsMode = portsMode((int(m.portsMode) + delta + 3) % 3)
	case fieldPortsPreset:
		m.portsPreset = portsPresetKind((int(m.portsPreset) + delta + 3) % 3)
	case fieldProfile:
		m.setProfile(profileKind((int(m.profileSelectIndex()) + delta + 3) % 3))
	}
}

func (m *NewScanModel) profileSelectIndex() int {
	switch m.profile {
	case profileQuick:
		return 0
	case profileFull:
		return 2
	default:
		return 1
	}
}

func (m *NewScanModel) setProfile(profile profileKind) {
	switch profile {
	case profileQuick:
		m.profile = profileQuick
		m.concurrency.SetValue("200")
		m.timeoutMs.SetValue("300")
	case profileFull:
		m.profile = profileFull
		m.concurrency.SetValue("50")
		m.timeoutMs.SetValue("3000")
	default:
		m.profile = profileDefault
		m.concurrency.SetValue("100")
		m.timeoutMs.SetValue("1000")
	}
}

func (m *NewScanModel) handleToggle() {
	switch m.focusedField {
	case fieldBanner:
		m.toggleBanner = !m.toggleBanner
	case fieldTLS:
		m.toggleTLS = !m.toggleTLS
	case fieldRDNS:
		m.toggleRDNS = !m.toggleRDNS
	}
}

func (m NewScanModel) updateFocusedInput(msg tea.Msg) (NewScanModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focusedField {
	case fieldTargets:
		m.targetsInput, cmd = m.targetsInput.Update(msg)
	case fieldPortsRange:
		m.portsRange, cmd = m.portsRange.Update(msg)
	case fieldPortsList:
		m.portsList, cmd = m.portsList.Update(msg)
	case fieldConcurrency:
		prev := m.concurrency.Value()
		m.concurrency, cmd = m.concurrency.Update(msg)
		if m.concurrency.Value() != prev {
			m.profile = profileCustom
		}
	case fieldTimeout:
		prev := m.timeoutMs.Value()
		m.timeoutMs, cmd = m.timeoutMs.Update(msg)
		if m.timeoutMs.Value() != prev {
			m.profile = profileCustom
		}
	case fieldLabel:
		m.label, cmd = m.label.Update(msg)
	}
	return m, cmd
}

func (m *NewScanModel) applyFocus() {
	m.targetsInput.Blur()
	m.portsRange.Blur()
	m.portsList.Blur()
	m.concurrency.Blur()
	m.timeoutMs.Blur()
	m.label.Blur()

	switch m.focusedField {
	case fieldTargets:
		m.targetsInput.Focus()
	case fieldPortsRange:
		m.portsRange.Focus()
	case fieldPortsList:
		m.portsList.Focus()
	case fieldConcurrency:
		m.concurrency.Focus()
	case fieldTimeout:
		m.timeoutMs.Focus()
	case fieldLabel:
		m.label.Focus()
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

func (m NewScanModel) validate() []string {
	var errs []string

	if strings.TrimSpace(m.targetsInput.Value()) == "" {
		errs = append(errs, "Targets is required.")
	}

	switch m.portsMode {
	case portsRange:
		if !validRange(m.portsRange.Value()) {
			errs = append(errs, "Ports range must be like 1-1024 within 1-65535.")
		}
	case portsList:
		if !validList(m.portsList.Value()) {
			errs = append(errs, "Ports list must be comma-separated numbers within 1-65535.")
		}
	}

	if !validPositiveInt(m.concurrency.Value()) {
		errs = append(errs, "Concurrency must be a positive number.")
	}

	if !validPositiveInt(m.timeoutMs.Value()) {
		errs = append(errs, "Timeout must be a positive number.")
	}

	return errs
}

func validPositiveInt(value string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	return err == nil && n > 0
}

func validRange(value string) bool {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 2 {
		return false
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return false
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return false
	}
	return start >= 1 && end <= 65535 && start <= end
}

func validList(value string) bool {
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return false
		}
		port, err := strconv.Atoi(part)
		if err != nil || port < 1 || port > 65535 {
			return false
		}
	}
	return true
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func newTextInput(placeholder, value string) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = placeholder
	input.SetValue(value)
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB"))
	input.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
	input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#111827")).Background(lipgloss.Color("#F59E0B"))
	input.CharLimit = 0
	return input
}

func focusTextInput(input textinput.Model) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#111827")).Background(lipgloss.Color("#F59E0B")).Padding(0, 1)
	value := input.View()
	return style.Render(value)
}

func blurTextInput(input textinput.Model) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E7EB")).Background(lipgloss.Color("#374151")).Padding(0, 1)
	value := input.View()
	return style.Render(value)
}
