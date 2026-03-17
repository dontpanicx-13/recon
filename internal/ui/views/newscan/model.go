package newscan

import (
	"errors"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"recon/internal/target"
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
	fieldFilePicker
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
	filePicker   filepicker.Model
	theme        theme.Theme

	portsMode   portsMode
	portsPreset portsPresetKind
	profile     profileKind

	toggleBanner bool
	toggleTLS    bool
	toggleRDNS   bool

	pickingFile    bool
	pickerSelected string
	pickerErr      error
	focusedField   fieldID
	disabled       bool
	lastErrors     []string
	lastWarnings   []string
	blinkOn        bool
	width          int
	height         int
}

func NewModel() NewScanModel {
	m := NewScanModel{
		portsMode:    portsPreset,
		portsPreset:  presetTop100,
		profile:      profileDefault,
		focusedField: fieldTargets,
		theme:        theme.Load(),
	}

	m.targetsInput = newTextInput(m.theme, "IP, domain, CIDR, list, or file path", "")
	m.portsRange = newTextInput(m.theme, "Range", "1-1024")
	m.portsList = newTextInput(m.theme, "List", "22,80,443")
	m.concurrency = newTextInput(m.theme, "Concurrency", "100")
	m.timeoutMs = newTextInput(m.theme, "Timeout", "1000")
	m.label = newTextInput(m.theme, "Label", "")
	m.filePicker = filepicker.New()
	m.filePicker.ShowHidden = true
	m.filePicker.DirAllowed = true
	m.filePicker.FileAllowed = true
	m.filePicker.AutoHeight = false
	m.filePicker.Height = 12

	m.applyFocus()
	return m
}

func (m *NewScanModel) setPickerError(path string) {
	m.pickerErr = errors.New(path + " is not valid.")
	m.pickerSelected = ""
}

func (m NewScanModel) validate() ([]string, []string) {
	var errs []string
	var warns []string

	if strings.TrimSpace(m.targetsInput.Value()) == "" {
		errs = append(errs, "Targets is required.")
		return errs, warns
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

	if len(errs) > 0 {
		return errs, warns
	}

	parseResult, parseErrs := target.Parse(m.targetsInput.Value(), target.Options{
		ExcludeNetworkBroadcast: true,
	})
	if len(parseErrs) > 0 {
		errs = append(errs, parseErrs...)
	}
	warns = append(warns, parseResult.Warnings...)

	if len(parseResult.Targets) == 0 {
		errs = append(errs, "No valid targets found.")
	}

	return errs, warns
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

func (m *NewScanModel) resetValidation() {
	m.lastErrors = nil
	m.lastWarnings = nil
}
