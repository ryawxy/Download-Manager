package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedInputStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	unselectedInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type NewDownloadTab struct {
	url, queue, saveAs textinput.Model
	cursor             int
	isActive           bool
}

func NewNewDownloadTab() NewDownloadTab {
	n := NewDownloadTab{}
	n.url = textinput.New()
	n.url.Placeholder = ""
	n.url.Focus()

	n.queue = textinput.New()
	n.queue.Placeholder = ""

	n.saveAs = textinput.New()
	n.saveAs.Placeholder = ""

	n.cursor = -1
	n.isActive = false
	return n
}

func (n NewDownloadTab) setActive(b bool) Tab {
	n.isActive = b
	if b {
		n.cursor = 0
	} else {
		n.cursor = -1
	}
	return n
}

func (n NewDownloadTab) isActivated() bool {
	return n.isActive
}

func (n NewDownloadTab) Init() tea.Cmd {
	n.url = textinput.New()
	n.url.Placeholder = ""
	n.url.Focus()

	n.queue = textinput.New()
	n.queue.Placeholder = ""

	n.saveAs = textinput.New()
	n.saveAs.Placeholder = ""

	n.cursor = -1
	return nil
}

func (n NewDownloadTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if n.cursor > 0 {
				n.cursor--
			} else {
				n.cursor = 2
			}
		case "down":
			if n.cursor < 2 {
				n.cursor++
			} else {
				n.cursor = 0
			}
		case "left":
			n.isActive = false
			n.cursor = -1
		case "enter":
			//todo call new_download function
		}
	}

	var cmd tea.Cmd
	switch n.cursor {
	case 0:
		n.url, cmd = n.url.Update(msg)
	case 1:
		n.queue, cmd = n.queue.Update(msg)
	case 2:
		n.saveAs, cmd = n.saveAs.Update(msg)
	}

	return n, cmd
}

func (n NewDownloadTab) View() string {
	urlView := n.url.View()
	queueView := n.queue.View()
	saveAsView := n.saveAs.View()

	if n.cursor == 0 {
		urlView = selectedInputStyle.Render("URL: " + urlView)
	} else {
		urlView = unselectedInputStyle.Render("URL: " + urlView)
	}

	if n.cursor == 1 {
		queueView = selectedInputStyle.Render("Queue: " + queueView)
	} else {
		queueView = unselectedInputStyle.Render("Queue: " + queueView)
	}

	if n.cursor == 2 {
		saveAsView = selectedInputStyle.Render("Save as: " + saveAsView)
	} else {
		saveAsView = unselectedInputStyle.Render("Save as: " + saveAsView)
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		urlView, queueView, saveAsView,
	)
}

func (n NewDownloadTab) toString() string {
	return "New Download"
}
