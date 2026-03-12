package newscan

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type blinkMsg struct{}

func BlinkCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return blinkMsg{}
	})
}
