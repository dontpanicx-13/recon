package statusbar

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	spinner     spinner.Model
	started     time.Time
	toolName    string
	toolVersion string
}

func NewModel(toolName, toolVersion string) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return Model{
		spinner:     sp,
		started:     time.Now(),
		toolName:    toolName,
		toolVersion: toolVersion,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}
