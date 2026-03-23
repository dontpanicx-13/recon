package details

import tea "github.com/charmbracelet/bubbletea"

type Action int

const (
	ActionNone Action = iota
	ActionClose
	ActionExportJSON
	ActionExportMarkdown
)

func (m Model) Update(msg tea.Msg) (Model, Action) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, ActionClose
		case "j":
			return m, ActionExportJSON
		case "m":
			return m, ActionExportMarkdown
		case "w":
			m.ScrollDetail(-1)
		case "s":
			m.ScrollDetail(1)
		case "W":
			m.ScrollDetail(-5)
		case "S":
			m.ScrollDetail(5)
		case "up":
			m.MoveSelection(-1)
		case "down":
			m.MoveSelection(1)
		case "home":
			m.DetailTop = 0
		case "end":
			m.DetailTop = 999999
		}
	}
	return m, ActionNone
}
