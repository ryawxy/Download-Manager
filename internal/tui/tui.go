package tui

//todo implement footer
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
}

type MainStage struct {
	currentTab    int
	tabs          []Tab
	height, width int
}

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

func (m MainStage) Init() tea.Cmd {
	ticker := time.NewTicker(500 * time.Millisecond)
	return func() tea.Msg {
		return tickMsg(<-ticker.C)
	}
}

func (m MainStage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

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

	return m, nil
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
	footerText := fmt.Sprintf(" Active Tab: %s | Use ↑ ↓ to navigate, → to activate | Press Ctrl+C to exit ", m.tabs[m.currentTab].toString())
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
