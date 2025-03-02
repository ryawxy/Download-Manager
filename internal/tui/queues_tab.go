package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type QueuesTab struct {
}

func (q QueuesTab) setActive(b bool) Tab {
	return q
}

func (q QueuesTab) isActivated() bool {
	return false
}

func (q QueuesTab) toString() string {
	return "Queues"
}

func (q QueuesTab) Init() tea.Cmd {
	return nil
}

func (q QueuesTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return q, nil
}

func (q QueuesTab) View() string {
	return ""
}
