package details

import "recon/internal/scanner"

type Model struct {
	Scan      scanner.ScanResult
	Hosts     []scanner.Host
	Selected  int
	DetailTop int
	Message   string
}

func NewModel() Model {
	return Model{}
}

func (m *Model) SetScan(scan scanner.ScanResult) {
	m.Scan = scan
	m.Hosts = append([]scanner.Host(nil), scan.Hosts...)
	m.Selected = 0
	m.DetailTop = 0
	m.Message = ""
}

func (m *Model) MoveSelection(delta int) {
	if len(m.Hosts) == 0 {
		m.Selected = 0
		m.DetailTop = 0
		return
	}
	m.Selected += delta
	if m.Selected < 0 {
		m.Selected = 0
	}
	if m.Selected >= len(m.Hosts) {
		m.Selected = len(m.Hosts) - 1
	}
	m.DetailTop = 0
	m.Message = ""
}

func (m *Model) ScrollDetail(delta int) {
	m.DetailTop += delta
	if m.DetailTop < 0 {
		m.DetailTop = 0
	}
}

func (m *Model) SetMessage(msg string) {
	m.Message = msg
}

func (m Model) SelectedHost() (scanner.Host, bool) {
	if m.Selected < 0 || m.Selected >= len(m.Hosts) {
		return scanner.Host{}, false
	}
	return m.Hosts[m.Selected], true
}
