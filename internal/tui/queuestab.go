package tui

import (
	"IDM/internal"
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"time"
)

//if you were in the mood, cleanup controls
// QueuesTab todo implement newQueue button

type QueuesTab struct {
	queues                                    []internal.Queue
	queueCursor, contentCursor, buttonsCursor int
	isActive                                  bool
	controlContent                            bool
	inputs                                    []textinput.Model
}

func temporaryRandomQueues() []internal.Queue {
	return []internal.Queue{
		{Id: "queue1", Directory: "Downloads/queue1", NumberOfFilesLimit: 5, BandwidthLimit: 1000, NumberOfTriesLimit: 3, StartTime: time.Now(), EndTime: time.Now().Add(2 * time.Hour)},
		{Id: "queue2", Directory: "Downloads/queue2", NumberOfFilesLimit: 10, BandwidthLimit: 2000, NumberOfTriesLimit: 2, StartTime: time.Now(), EndTime: time.Now().Add(3 * time.Hour)},
		{Id: "queue3", Directory: "Downloads/queue3", NumberOfFilesLimit: 7, BandwidthLimit: 1500, NumberOfTriesLimit: 4, StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)},
	}
}

func getNewQueue() internal.Queue {
	return internal.Queue{
		Id:                 "newQueue",
		Directory:          "Downloads",
		NumberOfFilesLimit: -1,
		BandwidthLimit:     -1,
		NumberOfTriesLimit: -1,
		StartTime:          time.Now(),
		EndTime:            time.Now(),
	}
}

func NewQueuesTab() QueuesTab {
	queues := temporaryRandomQueues()
	queues = append(queues, getNewQueue())
	inputs := make([]textinput.Model, 6)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = ""
		inputs[i].Prompt = ""
	}
	return QueuesTab{
		queues:         queues,
		buttonsCursor:  -1,
		queueCursor:    -1,
		contentCursor:  -1,
		controlContent: false,
		inputs:         inputs,
	}
}

func (q QueuesTab) updateCurrentQueue() {
	//todo we probably need to call backend here!!!
	q.queues[q.queueCursor].Directory = q.inputs[0].Value()
	q.queues[q.queueCursor].NumberOfFilesLimit, _ = strconv.Atoi(q.inputs[1].Value())
	q.queues[q.queueCursor].BandwidthLimit, _ = strconv.Atoi(q.inputs[2].Value())
	q.queues[q.queueCursor].NumberOfTriesLimit, _ = strconv.Atoi(q.inputs[3].Value())
	q.queues[q.queueCursor].StartTime, _ = time.Parse("2006-01-02 15:04:05", q.inputs[4].Value())
	q.queues[q.queueCursor].EndTime, _ = time.Parse("2006-01-02 15:04:05", q.inputs[5].Value())
}

func (q QueuesTab) setInputs() {
	q.inputs[0].SetValue(q.queues[q.queueCursor].Directory)
	q.inputs[0].CursorEnd()
	q.inputs[1].SetValue(strconv.Itoa(q.queues[q.queueCursor].NumberOfFilesLimit))
	q.inputs[1].CursorEnd()
	q.inputs[2].SetValue(strconv.Itoa(q.queues[q.queueCursor].BandwidthLimit))
	q.inputs[2].CursorEnd()
	q.inputs[3].SetValue(strconv.Itoa(q.queues[q.queueCursor].NumberOfTriesLimit))
	q.inputs[3].CursorEnd()
	q.inputs[4].SetValue(q.queues[q.queueCursor].StartTime.Format("2006-01-02 15:04:05"))
	q.inputs[4].CursorEnd()
	q.inputs[5].SetValue(q.queues[q.queueCursor].EndTime.Format("2006-01-02 15:04:05"))
	q.inputs[5].CursorEnd()
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
			if !q.controlContent {
				q.queueCursor = (q.queueCursor + len(q.queues) - 1) % len(q.queues)
				q.setInputs()
			} else {
				if q.contentCursor == len(q.inputs) {
					q.buttonsCursor = -1
				}
				q.contentCursor = (q.contentCursor + len(q.inputs)) % (len(q.inputs) + 1)
				if q.contentCursor == len(q.inputs) {
					q.buttonsCursor = 0
				}
			}
		case "down":
			if !q.controlContent {
				q.queueCursor = (q.queueCursor + 1) % len(q.queues)
				q.setInputs()
			} else {
				if q.contentCursor == len(q.inputs) {
					q.buttonsCursor = -1
				}
				q.contentCursor = (q.contentCursor + 1) % (len(q.inputs) + 1)
				if q.contentCursor == len(q.inputs) {
					q.buttonsCursor = 0
				}
			}
		case "right":
			if q.contentCursor == len(q.inputs) {
				q.buttonsCursor = (q.buttonsCursor + 1) % 3
			} else {
				q.controlContent = true
				q.contentCursor = 0
			}
		case "left":
			if q.controlContent {
				if q.contentCursor == len(q.inputs) {
					q.buttonsCursor = (q.buttonsCursor + 2) % 3
				} else {
					q.controlContent = false
					q.contentCursor = -1
				}
			} else {
				q.isActive = false
				q.queueCursor = -1
			}
		}
	case tickMsg:
		q.queues = make([]internal.Queue, 0)
		for _, queue := range internal.QueuesList {
			q.queues = append(q.queues, *queue)
		}
		q.queues = append(q.queues, getNewQueue())
		return q, tickCmd()
	}
	var cmd tea.Cmd
	if q.controlContent && q.contentCursor < len(q.inputs) {
		q.inputs[q.contentCursor], _ = q.inputs[q.contentCursor].Update(msg)
		q.updateCurrentQueue()
	}
	if cmd == nil {
		return q, tickCmd()
	}
	return q, cmd
}

var (
	style              = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	queueSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CC99"))
	selectedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#18FFFF")).Bold(true)
)

func (q QueuesTab) View() string {

	var queueList, queueContent string
	for i, queue := range q.queues {
		if q.queueCursor == i {
			queueList += queueSelectedStyle.Render(fmt.Sprintf("%8s |", queue.Id)) + "\n"
		} else {
			queueList += style.Render(fmt.Sprintf("%8s |", queue.Id)) + "\n"
		}
	}

	for i, label := range []string{"Default Download Path", "Number of Files Limit", "Bandwidth Limit", "Number of Retries Limit", "Start Time", "Finish Time"} {
		if q.contentCursor == i {
			queueContent += selectedStyle.Render(label+": "+q.inputs[i].View()) + "\n"
			q.inputs[i].Focus()
		} else {
			queueContent += style.Render(label+": "+q.inputs[i].View()) + "\n"
		}
	}
	buttons := []string{"Start All", "Pause All", "Delete"}
	if q.queueCursor == len(q.queues)-1 {
		buttons = []string{"Create"}
	}
	for i, button := range buttons {
		if q.buttonsCursor != -1 && q.buttonsCursor%len(buttons) == i {
			queueContent += selectedStyle.Render("["+button+"]") + "  "
		} else {
			queueContent += style.Render("["+button+"]") + "  "
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, queueList, "   ", queueContent)
}

func (q QueuesTab) getFooter() string {
	return "Press '→' to select, '←' to go back, '↑ / ↓' to navigate"
}
