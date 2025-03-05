package tui

import (
	"IDM/internal/queue"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
)

// QueuesTab todo show at most 10 queues at a time
// QueuesTab todo implement newQueue button
// QueuesTab todo show the contents of queues
// QueuesTab todo implement buttons

type QueuesTab struct {
	queues                     []queue.Queue
	queueCursor, contentCursor int
	isActive                   bool
	controlContent             bool
	inputs                     []textinput.Model
}

func temporaryRandomQueues() []queue.Queue {
	return []queue.Queue{
		{Id: "queue1", Directory: "Downloads/queue1", NumberOfFilesLimit: 5, BandwidthLimit: 1000, NumberOfTriesLimit: 3, StartTime: time.Now(), EndTime: time.Now().Add(2 * time.Hour)},
		{Id: "queue2", Directory: "Downloads/queue2", NumberOfFilesLimit: 10, BandwidthLimit: 2000, NumberOfTriesLimit: 2, StartTime: time.Now(), EndTime: time.Now().Add(3 * time.Hour)},
		{Id: "queue3", Directory: "Downloads/queue3", NumberOfFilesLimit: 7, BandwidthLimit: 1500, NumberOfTriesLimit: 4, StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)},
		{Id: "newQueue", Directory: "", NumberOfFilesLimit: 0, BandwidthLimit: 0, NumberOfTriesLimit: 0, StartTime: time.Now(), EndTime: time.Now()},
	}
}

func NewQueuesTab() QueuesTab {
	queues := temporaryRandomQueues()
	inputs := make([]textinput.Model, 5)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = ""
		inputs[i].Prompt = ""
	}
	return QueuesTab{
		queues:         queues,
		queueCursor:    -1,
		contentCursor:  -1,
		controlContent: false,
		inputs:         inputs,
	}
}

func (q QueuesTab) setActive(b bool) Tab {
	q.isActive = b
	if b {
		q.queueCursor = 0
	} else {
		q.queueCursor = -1
	}
	return q
}

func (q QueuesTab) isActivated() bool {
	return q.isActive
}

func (q QueuesTab) toString() string {
	return "Queues"
}

func (q QueuesTab) Init() tea.Cmd {
	return nil
}

func (q QueuesTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if !q.controlContent && q.queueCursor > 0 {
				q.queueCursor--
			} else if q.controlContent && q.contentCursor > 0 {
				q.contentCursor--
			}
		case "down":
			if !q.controlContent && q.queueCursor < len(q.queues)-1 {
				q.queueCursor++
			} else if q.controlContent && q.contentCursor < len(q.inputs)-1 {
				q.contentCursor++
			}
		case "right":
			q.controlContent = true
			q.contentCursor = 0
		case "left":
			if q.controlContent {
				q.controlContent = false
				q.contentCursor = -1
			} else {
				q.isActive = false
				q.queueCursor = -1
			}
		case "ctrl+c":
			return q, tea.Quit
		}
	}

	var cmd tea.Cmd
	if q.controlContent {
		q.inputs[q.contentCursor], cmd = q.inputs[q.contentCursor].Update(msg)
	}

	return q, cmd
}

var (
	style         = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#18FFFF")).Bold(true)
)

func (q QueuesTab) View() string {

	var queueList, queueContent string
	for i, queue := range q.queues {
		if q.queueCursor == i {
			queueList += selectedStyle.Render(fmt.Sprintf("%8s |", queue.Id)) + "\n"
		} else {
			queueList += style.Render(fmt.Sprintf("%8s |", queue.Id)) + "\n"
		}
	}

	if q.controlContent {
		queueContent += selectedStyle.Render("Queue Settings:") + "\n\n"
		for i, label := range []string{"Default Download Path", "Number of Files Limit", "Bandwidth Limit", "Number of Retries Limit", "Start Time"} {
			if q.contentCursor == i {
				queueContent += selectedStyle.Render(label+": "+q.inputs[i].View()) + "\n"
				q.inputs[i].Focus()
			} else {
				queueContent += style.Render(label+": "+q.inputs[i].View()) + "\n"
			}
		}
		queueContent += "\n[startAll]  [pause_all]  [delete]"
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, queueList, "   ", queueContent)
}
