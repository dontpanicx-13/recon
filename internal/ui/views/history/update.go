package history

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.lastWidth = msg.Width
		m.lastHeight = msg.Height
		m.clampTop()
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			m.MoveSelection(-1)
			return m, nil
		case "down":
			m.MoveSelection(1)
			return m, nil
		case "pgup":
			m.MoveSelection(-5)
			return m, nil
		case "pgdown":
			m.MoveSelection(5)
			return m, nil
		case "home":
			m.JumpTop()
			return m, nil
		case "end":
			m.JumpBottom()
			return m, nil
		}
	}
	return m, nil
}
