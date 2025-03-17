package tui

import (
	"IDM/internal"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedInputStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#18FFFF")).Bold(true)
	unselectedInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
)

type NewDownloadTab struct {
	url, queue, saveAs textinput.Model
	cursor             int
	isActive           bool
	errorMsg           string
}

func NewNewDownloadTab() NewDownloadTab {
	n := NewDownloadTab{}
	n.url = textinput.New()
	n.url.Placeholder = ""
	n.url.Prompt = ""
	n.url.Focus()

	n.queue = textinput.New()
	n.queue.Placeholder = ""
	n.queue.Prompt = ""

	n.saveAs = textinput.New()
	n.saveAs.Placeholder = ""
	n.saveAs.Prompt = ""

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

	n.queue = textinput.New()
	n.queue.Placeholder = ""

	n.saveAs = textinput.New()
	n.saveAs.Placeholder = ""

	n.cursor = -1
	return nil
}

func (n *NewDownloadTab) newDownload() error {
	url := n.url.Value()
	queueID := n.queue.Value()
	fileName := n.saveAs.Value()

	if url == "" {
		n.errorMsg = "URL cannot be empty"
		return fmt.Errorf(n.errorMsg)
	}
	if queueID == "" {
		n.errorMsg = "Queue cannot be empty"
		return fmt.Errorf(n.errorMsg)
	}
	if fileName == "" {
		n.errorMsg = "File name cannot be empty"
		return fmt.Errorf(n.errorMsg)
	}

	queue := internal.GetQueue(queueID)
	if queue == nil {
		n.errorMsg = fmt.Sprintf("Queue %s not found", queueID)
		return fmt.Errorf(n.errorMsg)
	}

	// Create new download instance
	download := internal.NewDownload(url, fileName, queue.Directory)
	err := queue.AddDownload(download)
	if err != nil {
		n.errorMsg = fmt.Sprintf("Failed to add download: %v", err)
		return err
	}

	n.url.SetValue("")
	n.queue.SetValue("")
	n.saveAs.SetValue("")
	n.errorMsg = "Download added to queue"
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
			n.newDownload()
		}
	}

	var cmd tea.Cmd
	switch n.cursor {
	case 0:
		n.url, cmd = n.url.Update(msg)
		n.url.Focus()
	case 1:
		n.queue, cmd = n.queue.Update(msg)
		n.queue.Focus()
	case 2:
		n.saveAs, cmd = n.saveAs.Update(msg)
		n.saveAs.Focus()
	}
	if cmd == nil {
		return n, tickCmd()
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

	errMsg := errorStyle.Render(n.errorMsg)

	return lipgloss.JoinVertical(lipgloss.Top,
		urlView, queueView, saveAsView, errMsg,
	)
}

func (n NewDownloadTab) toString() string {
	return "New Download"
}

func (n NewDownloadTab) getFooter() string {
	return "Use '↑ / ↓' to navigate, " + "Press Enter to start download"
}
