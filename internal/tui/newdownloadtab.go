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
	url, queue, saveAs, name textinput.Model
	cursor                   int
	isActive                 bool
	errorMsg                 string
	successMsg               string
	buttonMode               bool
	buttonIndex              int
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
	n.saveAs.Placeholder = "Enter Local Path"
	n.saveAs.Prompt = ""

	n.name = textinput.New()
	n.name.Placeholder = "Enter Desired Name"
	n.name.Prompt = ""

	n.cursor = -1
	n.isActive = false
	n.errorMsg = ""
	n.successMsg = ""
	n.buttonMode = false
	n.buttonIndex = 0
	return n
}

func (n NewDownloadTab) setActive(b bool) Tab {
	n.isActive = b
	if b {
		n.cursor = 0
		n.buttonMode = false
		n.buttonIndex = 0
	} else {
		n.cursor = -1
		n.buttonMode = false
		n.buttonIndex = 0
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
	n.saveAs.Placeholder = "Enter Local Path"

	n.name = textinput.New()
	n.name.Placeholder = "Enter Desired Name"

	n.cursor = -1
	n.errorMsg = ""
	n.successMsg = ""
	n.buttonMode = false
	n.buttonIndex = 0
	return nil
}

func (n *NewDownloadTab) newDownload() {
	url := strings.TrimSpace(n.url.Value())
	queueName := strings.TrimSpace(n.queue.Value())
	saveAs := strings.TrimSpace(n.saveAs.Value())
	name := strings.TrimSpace(n.name.Value())

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
		FileName:  name,
	}
	newDownload.NewDownloadManager(internal.WORKERS, selectedQueue.TokenBucket)

	err := newDownload.GetFileSizeAndName()
	if err != nil {
		n.errorMsg = "Invalid URL or unreachable resource"
		return
	}
	if name != "" {
		newDownload.FileName = name
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
}

func (n NewDownloadTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if n.buttonMode {
			switch msg.String() {
			case "left":
				if n.buttonIndex > 0 {
					n.buttonIndex--
				}
			case "right":
				if n.buttonIndex < 1 {
					n.buttonIndex++
				}
			case "up":
				n.buttonMode = false
				n.buttonIndex = 0
				n.cursor = 3
			case "enter":
				if n.buttonIndex == 0 {
					n.newDownload()
					n.url.SetValue("")
					n.queue.SetValue("")
					n.saveAs.SetValue("")
					n.name.SetValue("")
					n.errorMsg = ""
					n.successMsg = ""
				} else {
					n.url.SetValue("")
					n.queue.SetValue("")
					n.saveAs.SetValue("")
					n.name.SetValue("")
					n.errorMsg = ""
					n.successMsg = ""
				}
				n.buttonMode = false
				n.buttonIndex = 0
				n.cursor = 0
			}
			return n, nil
		}
		switch msg.String() {
		case "up":
			if n.cursor > 0 {
				n.cursor--
			} else {
				n.cursor = 3
			}
		case "down":
			if n.cursor < 3 {
				n.cursor++
			} else {
				n.buttonMode = true
				n.buttonIndex = 0
			}
		case "left":
			n.isActive = false
			n.cursor = -1
		case "enter":
			n.newDownload()
		default:
			switch n.cursor {
			case 0:
				n.url, _ = n.url.Update(msg)
				n.url.Focus()
			case 1:
				n.queue, _ = n.queue.Update(msg)
				n.queue.Focus()
			case 2:
				n.saveAs, _ = n.saveAs.Update(msg)
				n.saveAs.Focus()
			case 3:
				n.name, _ = n.name.Update(msg)
				n.name.Focus()
			}
		}
	}
	return n, nil
}

func (n NewDownloadTab) View() string {
	urlView := n.url.View()
	queueView := n.queue.View()
	saveAsView := n.saveAs.View()
	nameView := n.name.View()

	if n.cursor == 0 && !n.buttonMode {
		urlView = selectedInputStyle.Render("URL: " + urlView)
	} else {
		urlView = unselectedInputStyle.Render("URL: " + urlView)
	}
	if n.cursor == 1 && !n.buttonMode {
		queueView = selectedInputStyle.Render("Queue: " + queueView)
	} else {
		queueView = unselectedInputStyle.Render("Queue: " + queueView)
	}
	if n.cursor == 2 && !n.buttonMode {
		saveAsView = selectedInputStyle.Render("Save as: " + saveAsView)
	} else {
		saveAsView = unselectedInputStyle.Render("Save as: " + saveAsView)
	}
	if n.cursor == 3 && !n.buttonMode {
		nameView = selectedInputStyle.Render("Name: " + nameView)
	} else {
		nameView = unselectedInputStyle.Render("Name: " + nameView)
	}

	errMsg := errorStyle.Render(n.errorMsg)
	successMsg := successStyle.Render(n.successMsg)

	fields := lipgloss.JoinVertical(lipgloss.Top, urlView, queueView, saveAsView, nameView)
	var layout string
	var createButton, cancelButton string
	if n.buttonMode {
		if n.buttonIndex == 0 {
			createButton = selectedInputStyle.Render("[Create]")
			cancelButton = unselectedInputStyle.Render("[Cancel]")
		} else {
			createButton = unselectedInputStyle.Render("[Create]")
			cancelButton = selectedInputStyle.Render("[Cancel]")
		}
		buttons := lipgloss.JoinHorizontal(lipgloss.Center, createButton, "   ", cancelButton)
		layout = lipgloss.JoinVertical(lipgloss.Top, fields, buttons, errMsg, successMsg)
	} else {
		createButton = unselectedInputStyle.Render("[Create]")
		cancelButton = unselectedInputStyle.Render("[Cancel]")
		buttons := lipgloss.JoinHorizontal(lipgloss.Center, createButton, "   ", cancelButton)
		layout = lipgloss.JoinVertical(lipgloss.Top, fields, buttons, errMsg, successMsg)
	}
	return layout
}

func (n NewDownloadTab) toString() string {
	return "New Download"
}

func (n NewDownloadTab) getFooter() string {
	return "Use '↑/↓' to navigate fields. When on the last field, press down to select buttons; use left/right to choose and Enter to execute."
}
