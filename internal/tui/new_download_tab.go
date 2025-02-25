package tui

import tea "github.com/charmbracelet/bubbletea"

type NewDownloadTab struct {
}

func (n NewDownloadTab) toString() string {
	return "New Download"
}

func (n NewDownloadTab) Init() tea.Cmd {
	return nil
}

func (n NewDownloadTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return n, nil
}

func (n NewDownloadTab) View() string {
	return ""
}
