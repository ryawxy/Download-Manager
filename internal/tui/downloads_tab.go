package tui

import tea "github.com/charmbracelet/bubbletea"

type DownloadsTab struct {
}

func (d DownloadsTab) setActive(b bool) Tab {
	return d
}

func (d DownloadsTab) isActivated() bool {
	return false
}

func (d DownloadsTab) toString() string {
	return "Downloads"
}

func (d DownloadsTab) Init() tea.Cmd {
	return nil
}

func (d DownloadsTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return d, nil
}

func (d DownloadsTab) View() string {
	return "Downloads"
}
