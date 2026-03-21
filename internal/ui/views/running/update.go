package running

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			m.scrollLogs(-1)
			return m, nil
		case "down":
			m.scrollLogs(1)
			return m, nil
		case "pgup":
			m.scrollLogs(-5)
			return m, nil
		case "pgdown":
			m.scrollLogs(5)
			return m, nil
		case "home":
			m.scanLogFollow = false
			m.scanLogTop = 0
			return m, nil
		case "end":
			m.scanLogFollow = true
			return m, nil
		case "enter", "c":
			m.CancelScan()
			return m, nil
		}
	}
	return m, nil
}
