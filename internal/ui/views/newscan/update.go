package newscan

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m NewScanModel) Update(msg tea.Msg) (NewScanModel, tea.Cmd) {
	if m.disabled {
		return m, nil
	}

	switch msg := msg.(type) {
	case blinkMsg:
		m.blinkOn = !m.blinkOn
		return m, BlinkCmd()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	if m.pickingFile {
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
			m.pickingFile = false
			m.applyFocus()
			return m, nil
		}
		var cmd tea.Cmd
		m.filePicker, cmd = m.filePicker.Update(msg)
		if didSelect, path := m.filePicker.DidSelectFile(msg); didSelect {
			m.pickerSelected = path
			m.pickerErr = nil
			m.applyPickedFile(path)
			m.pickingFile = false
			m.applyFocus()
			return m, nil
		}
		if didSelect, path := m.filePicker.DidSelectDisabledFile(msg); didSelect {
			m.setPickerError(path)
			return m, cmd
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			m.focusedField = m.nextField(false)
			m.applyFocus()
			return m, nil
		case "down":
			m.focusedField = m.nextField(true)
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
			if m.isToggleField(m.focusedField) {
				m.handleToggle()
				return m, nil
			}
			if m.focusedField == fieldFilePicker {
				m.pickingFile = true
				m.pickerErr = nil
				m.filePicker.CurrentDirectory = m.filePickerStartDir()
				m.filePicker.AutoHeight = false
				m.filePicker.Height = m.pickerHeight()
				return m, m.filePicker.Init()
			}
			if m.focusedField == fieldStart {
				m.lastErrors = m.validate()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m, cmd = m.updateFocusedInput(msg)
	return m, cmd
}

func (m NewScanModel) nextField(forward bool) fieldID {
	order := []fieldID{
		fieldTargets,
		fieldFilePicker,
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

func (m *NewScanModel) applyPickedFile(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	current := strings.TrimSpace(m.targetsInput.Value())
	if current == "" {
		m.targetsInput.SetValue(path)
		return
	}
	if strings.HasSuffix(current, ",") {
		m.targetsInput.SetValue(current + " " + path)
		return
	}
	m.targetsInput.SetValue(current + ", " + path)
}

func (m NewScanModel) filePickerStartDir() string {
	cwd, err := os.Getwd()
	if err == nil && cwd != "" {
		return filepath.Clean(cwd)
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Clean(home)
	}
	return "."
}

func (m NewScanModel) pickerHeight() int {
	h := m.height
	if h == 0 {
		h = 20
	}
	return max(8, min(18, h-6))
}
