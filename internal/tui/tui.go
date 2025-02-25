package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedTabStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("48")).Background(lipgloss.Color("")).Bold(true)
	unselectedTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Background(lipgloss.Color(""))
)

type Tab interface {
	tea.Model
	toString() string
}

type MainStage struct {
	currentTab int
	tabs       []Tab
}

func NewMainStage() MainStage {
	return MainStage{
		currentTab: 0,
		tabs:       append([]Tab{}, NewDownloadTab{}, DownloadsTab{}, QueuesTab{}),
	}
}

func (m MainStage) Init() tea.Cmd {
	return nil
}
func (m MainStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}
func (m MainStage) View() string {
	var renderedTabs []string
	for i, tab := range m.tabs {
		if i == m.currentTab {
			renderedTabs = append(renderedTabs, selectedTabStyle.Render(fmt.Sprintf("%12s |", tab.toString())))
		} else {
			renderedTabs = append(renderedTabs, unselectedTabStyle.Render(fmt.Sprintf("%12s |", tab.toString())))
		}
	}
	tabs := lipgloss.JoinVertical(lipgloss.Top, renderedTabs...)
	content := m.tabs[m.currentTab].View()
	return lipgloss.JoinHorizontal(lipgloss.Left, tabs, "        ", content)
}
