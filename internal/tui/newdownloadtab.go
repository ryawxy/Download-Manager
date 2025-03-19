package tui

import (
	"IDM/internal"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var (
	selectedInputStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#18FFFF")).Bold(true)
	unselectedInputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errorStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Bold(true)
	successStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Bold(true)
)

type NewDownloadTab struct {
	url, queue, saveAs textinput.Model
	cursor             int
	isActive           bool
	errorMsg           string
	successMsg         string
}

func NewNewDownloadTab() NewDownloadTab {
	n := NewDownloadTab{}
	n.url = textinput.New()
	n.url.Placeholder = "Enter URL"
	n.url.Prompt = ""
	n.url.Focus()

	n.queue = textinput.New()
	n.queue.Placeholder = "Enter Queue Name"
	n.queue.Prompt = ""

	n.saveAs = textinput.New()
	n.saveAs.Placeholder = "Enter Filename"
	n.saveAs.Prompt = ""

	n.cursor = -1
	n.isActive = false
	n.errorMsg = ""
	n.successMsg = ""

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
	n.url.Placeholder = "Enter URL"

	n.queue = textinput.New()
	n.queue.Placeholder = "Enter Queue Name"

	n.saveAs = textinput.New()
	n.saveAs.Placeholder = "Enter Filename"

	n.cursor = -1
	n.errorMsg = ""
	n.successMsg = ""

	return nil
}

func (n *NewDownloadTab) newDownload() {
	url := strings.TrimSpace(n.url.Value())
	queueName := strings.TrimSpace(n.queue.Value())
	saveAs := strings.TrimSpace(n.saveAs.Value())

	if url == "" || queueName == "" {
		n.errorMsg = "All fields must be filled!"
		n.successMsg = ""
		return
	}

	var selectedQueue *internal.Queue
	for _, q := range internal.QueuesList {
		if q.Id == queueName {
			selectedQueue = q
			break
		}
	}

	if selectedQueue == nil {
		n.errorMsg = fmt.Sprintf("Queue '%s' not found!", queueName)
		n.successMsg = ""
		return
	}
	if saveAs == "" {
		saveAs = selectedQueue.Directory
	}
	newDownload := internal.Download{
		URL:       url,
		Directory: saveAs,
		QueueName: queueName,
		Status:    "Pending",
	}
	newDownload.NewDownloadManager(internal.WORKERS, selectedQueue.TokenBucket)
	err := newDownload.GetFileSizeAndName()
	if err != nil {
		n.errorMsg = "Invalid URL or unreachable resource"
		return
	}

	err = selectedQueue.AddDownload(&newDownload)
	if err != nil {
		n.errorMsg = "Failed to add to queue"
		return
	}
	internal.DownloadsList = append(internal.DownloadsList, &newDownload)

	fmt.Println(len(selectedQueue.Downloads))

	n.successMsg = fmt.Sprintf("Added '%s' to queue '%s'", newDownload.FileName, queueName)
	n.errorMsg = ""

	n.url.SetValue("")
	n.queue.SetValue("")
	n.saveAs.SetValue("")
	n.cursor = 0
	n.successMsg = ""
	n.errorMsg = ""
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
	successMsg := successStyle.Render(n.successMsg)

	return lipgloss.JoinVertical(lipgloss.Top,
		urlView, queueView, saveAsView, errMsg, successMsg,
	)
}

func (n NewDownloadTab) toString() string {
	return "New Download"
}

func (n NewDownloadTab) getFooter() string {
	return "Use '↑ / ↓' to navigate, Press Enter to add download"
}
