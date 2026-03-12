package newscan

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"recon/internal/ui/theme"
)

type portsMode int

type portsPresetKind int

type profileKind int

type fieldID int

const (
	portsPreset portsMode = iota
	portsRange
	portsList
)

const (
	presetTop100 portsPresetKind = iota
	presetTop1000
	presetAll
)

const (
	profileQuick profileKind = iota
	profileDefault
	profileFull
	profileCustom
)

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
	theme        theme.Theme

	portsMode   portsMode
	portsPreset portsPresetKind
	profile     profileKind

	toggleBanner bool
	toggleTLS    bool
	toggleRDNS   bool

	focusedField fieldID
	disabled     bool
	lastErrors   []string
	blinkOn      bool
	width        int
	height       int
}

func NewModel() NewScanModel {
	m := NewScanModel{
		portsMode:    portsPreset,
		portsPreset:  presetTop100,
		profile:      profileDefault,
		focusedField: fieldTargets,
		theme:        theme.Load(),
	}

	m.targetsInput = newTextInput(m.theme, "Targets", "")
	m.portsRange = newTextInput(m.theme, "Range", "1-1024")
	m.portsList = newTextInput(m.theme, "List", "22,80,443")
	m.concurrency = newTextInput(m.theme, "Concurrency", "100")
	m.timeoutMs = newTextInput(m.theme, "Timeout", "1000")
	m.label = newTextInput(m.theme, "Label", "")

	m.applyFocus()
	return m
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

func newTextInput(theme theme.Theme, placeholder, value string) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.Placeholder = placeholder
	input.SetValue(value)
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.AccentFg))
	input.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
	input.Cursor.Style = lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.AccentFg)).
		Background(lipgloss.Color(theme.AccentBg))
	input.CharLimit = 0
	return input
}
