package history

import "recon/internal/store"

type Model struct {
	items      []store.ManifestItem
	selected   int
	top        int
	lastHeight int
	lastWidth  int
	errMessage string
	prompt     string
}

func NewModel() Model {
	return Model{}
}

func (m *Model) SetItems(items []store.ManifestItem) {
	m.items = items
	if m.selected >= len(m.items) {
		m.selected = max(0, len(m.items)-1)
	}
	m.clampTop()
}

func (m *Model) SetError(err error) {
	if err == nil {
		m.errMessage = ""
		return
	}
	m.errMessage = err.Error()
}

func (m *Model) SetPrompt(value string) {
	m.prompt = value
}

func (m Model) Prompt() string {
	return m.prompt
}

func (m Model) Error() string {
	return m.errMessage
}

func (m Model) Items() []store.ManifestItem {
	return m.items
}

func (m Model) Selected() (store.ManifestItem, bool) {
	if m.selected < 0 || m.selected >= len(m.items) {
		return store.ManifestItem{}, false
	}
	return m.items[m.selected], true
}

func (m *Model) MoveSelection(delta int) {
	if len(m.items) == 0 {
		m.selected = 0
		m.top = 0
		return
	}
	m.selected += delta
	if m.selected < 0 {
		m.selected = 0
	}
	if m.selected >= len(m.items) {
		m.selected = len(m.items) - 1
	}
	m.clampTop()
}

func (m *Model) JumpTop() {
	if len(m.items) == 0 {
		m.selected = 0
		m.top = 0
		return
	}
	m.selected = 0
	m.top = 0
}

func (m *Model) JumpBottom() {
	if len(m.items) == 0 {
		m.selected = 0
		m.top = 0
		return
	}
	m.selected = len(m.items) - 1
	m.clampTop()
}

func (m *Model) clampTop() {
	if len(m.items) == 0 {
		m.top = 0
		return
	}
	visible := max(1, m.lastHeight-2)
	if m.selected < m.top {
		m.top = m.selected
	}
	if m.selected >= m.top+visible {
		m.top = m.selected - visible + 1
	}
	maxTop := max(0, len(m.items)-visible)
	if m.top > maxTop {
		m.top = maxTop
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
