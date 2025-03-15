package tui

//todo update README.md

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"time"
)

type Tab interface {
	tea.Model
	toString() string
	setActive(bool) Tab
	isActivated() bool
	getFooter() string
}

type MainStage struct {
	currentTab    int
	tabs          []Tab
	height, width int
}

const newDownloadTabId = 0
const queuesTabId = 2
const downloadsTabId = 1

func NewMainStage() MainStage {
	return MainStage{
		height:     20,
		width:      100,
		currentTab: 0,
		tabs:       append([]Tab{}, NewNewDownloadTab(), NewDownloadsTab(), NewQueuesTab()),
	}
}

func Start() {
	p := tea.NewProgram(NewMainStage(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println("Error starting program:", err)
		os.Exit(1)
	}
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m MainStage) Init() tea.Cmd {
	for i, tab := range m.tabs {
		n, _ := tab.Update(tickMsg(time.Now()))
		m.tabs[i] = n.(Tab)
	}
	return tickCmd()
}

func (m MainStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case exitQueuesMsg:
		// Deactivate the current tab so that tabs menu is shown.
		m.tabs[m.currentTab] = m.tabs[m.currentTab].setActive(false)
		return m, nil

	case tea.WindowSizeMsg: // Handles terminal resizing
		m.height = msg.Height // Update stored height dynamically
		m.width = msg.Width   // Update stored width dynamically
	}
	for _, tab := range m.tabs {
		if tab.isActivated() {
			n, cmd := tab.Update(msg)
			m.tabs[m.currentTab] = n.(Tab)
			return m, cmd
		}
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.currentTab = (m.currentTab + 1) % len(m.tabs)
		case "up":
			m.currentTab = (m.currentTab + len(m.tabs) - 1) % len(m.tabs)
		case "right":
			m.tabs[m.currentTab] = m.tabs[m.currentTab].setActive(true)
		}
	}

	return m, tickCmd()
}

var (
	selectedTabStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("48")).Background(lipgloss.Color("")).Bold(true)
	unselectedTabStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Background(lipgloss.Color(""))
	footerStyle        = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")). // Dark gray color
				Background(lipgloss.Color("")).    // Black background
				Padding(0, 1)
)

func (m MainStage) View() string {
	// Generate tab labels
	var renderedTabs []string
	for i, tab := range m.tabs {
		if i == m.currentTab {
			renderedTabs = append(renderedTabs, selectedTabStyle.Render(fmt.Sprintf("%12s |", tab.toString())))
		} else {
			renderedTabs = append(renderedTabs, unselectedTabStyle.Render(fmt.Sprintf("%12s |", tab.toString())))
		}
	}

	// Tab navigation bar
	tabs := lipgloss.JoinVertical(lipgloss.Top, renderedTabs...)
	content := m.tabs[m.currentTab].View()

	// Footer text
	footerText := fmt.Sprintf(" Active Tab: %s %s", m.tabs[m.currentTab].toString(), m.tabs[m.currentTab].getFooter())
	footer := footerStyle.Render(footerText)

	// Properly position footer at the **BOTTOM** of the terminal
	body := lipgloss.JoinHorizontal(lipgloss.Left, tabs, "     ", content)

	// Calculate available space for content
	contentHeight := lipgloss.Height(body)
	remainingSpace := m.height - contentHeight - 2 // Adjusted spacing

	// Ensure at least some space before footer
	if remainingSpace < 0 {
		remainingSpace = 0
	}

	// Use `lipgloss.Place` to enforce bottom positioning
	return body + lipgloss.Place(m.width, max(m.height-lipgloss.Height(body), 0), lipgloss.Left, lipgloss.Bottom, footer)
}
