package tui

import (
	"IDM/internal"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type exitQueuesMsg struct{}

type QueuesTab struct {
	queues           []internal.Queue
	queueCursor      int
	contentCursor    int
	buttonsCursor    int
	isActive         bool
	controlContent   bool
	creatingNewQueue bool
	inputs           []textinput.Model
}

func temporaryRandomQueues() []internal.Queue {
	return []internal.Queue{
		{Id: "queue1", Directory: "Downloads/queue1", MaxConcurrentDownloads: 5, BandwidthLimit: 1000, NumberOfTriesLimit: 3, StartTime: time.Now(), EndTime: time.Now().Add(2 * time.Hour)},
		{Id: "queue2", Directory: "Downloads/queue2", MaxConcurrentDownloads: 10, BandwidthLimit: 2000, NumberOfTriesLimit: 2, StartTime: time.Now(), EndTime: time.Now().Add(3 * time.Hour)},
		{Id: "queue3", Directory: "Downloads/queue3", MaxConcurrentDownloads: 7, BandwidthLimit: 1500, NumberOfTriesLimit: 4, StartTime: time.Now(), EndTime: time.Now().Add(1 * time.Hour)},
	}
}

func NewQueuesTab() *QueuesTab {
	queues := temporaryRandomQueues()
	inputs := make([]textinput.Model, 7)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = ""
		inputs[i].Prompt = ""
	}
	tab := &QueuesTab{
		queues:           queues,
		buttonsCursor:    -1,
		queueCursor:      0,
		contentCursor:    0,
		controlContent:   false,
		creatingNewQueue: false,
		inputs:           inputs,
	}
	tab.setInputs()
	return tab
}

func (q QueuesTab) updateCurrentQueue() QueuesTab {
	q.queues[q.queueCursor].Directory = q.inputs[0].Value()
	q.queues[q.queueCursor].NumberOfTriesLimit, _ = strconv.Atoi(q.inputs[1].Value())
	q.queues[q.queueCursor].BandwidthLimit, _ = strconv.Atoi(q.inputs[2].Value())
	q.queues[q.queueCursor].MaxConcurrentDownloads, _ = strconv.Atoi(q.inputs[3].Value())
	q.queues[q.queueCursor].StartTime, _ = time.Parse("2006-01-02 15:04:05", q.inputs[4].Value())
	q.queues[q.queueCursor].EndTime, _ = time.Parse("2006-01-02 15:04:05", q.inputs[5].Value())

	internalQueue := internal.QueuesList[q.queues[q.queueCursor].Id]
	internalQueue.EditQueue(
		q.queues[q.queueCursor].Directory,
		q.queues[q.queueCursor].NumberOfTriesLimit,
		q.queues[q.queueCursor].MaxConcurrentDownloads,
		q.queues[q.queueCursor].StartTime,
		q.queues[q.queueCursor].EndTime,
		q.queues[q.queueCursor].BandwidthLimit,
	)
	return q
}

func (q *QueuesTab) setInputs() {
	if len(q.queues) == 0 {
		return
	}
	q.inputs[0].SetValue(q.queues[q.queueCursor].Directory)
	q.inputs[0].CursorEnd()
	q.inputs[1].SetValue(strconv.Itoa(q.queues[q.queueCursor].NumberOfTriesLimit))
	q.inputs[1].CursorEnd()
	q.inputs[2].SetValue(strconv.Itoa(q.queues[q.queueCursor].BandwidthLimit))
	q.inputs[2].CursorEnd()
	q.inputs[3].SetValue(strconv.Itoa(q.queues[q.queueCursor].MaxConcurrentDownloads))
	q.inputs[3].CursorEnd()
	q.inputs[4].SetValue(q.queues[q.queueCursor].StartTime.Format("2006-01-02 15:04:05"))
	q.inputs[4].CursorEnd()
	q.inputs[5].SetValue(q.queues[q.queueCursor].EndTime.Format("2006-01-02 15:04:05"))
	q.inputs[5].CursorEnd()
}

func (q *QueuesTab) deleteQueue() {
	if len(q.queues) == 0 {
		return
	}
	queueID := q.queues[q.queueCursor].Id

	newDL := make([]*internal.Download, 0)
	for _, dl := range internal.DownloadsList {
		if dl.QueueName != queueID {
			newDL = append(newDL, dl)
		}
	}
	internal.DownloadsList = newDL

	q.queues = internal.DeleteQueue(queueID)
	if q.queueCursor >= len(q.queues) {
		q.queueCursor = max(0, len(q.queues)-1)
	}
	q.setInputs()
}

func (q *QueuesTab) setActive(b bool) Tab {
	q.isActive = b
	if b {
		q.queueCursor = 0
		q.setInputs()
	}
	return q
}

func (q *QueuesTab) isActivated() bool {
	return q.isActive
}

func (q *QueuesTab) toString() string {
	return "Queues"
}

func (q *QueuesTab) Init() tea.Cmd {
	return nil
}

func (q *QueuesTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if len(q.queues) > 0 && q.queueCursor < 0 {
			q.queueCursor = 0
		}
		if !q.creatingNewQueue {
			switch msg.String() {
			case "up":
				if q.buttonsCursor != -1 {
					q.buttonsCursor = -1
					q.contentCursor = len(q.inputs) - 1
				} else if !q.controlContent {
					q.queueCursor = (q.queueCursor + len(q.queues) - 1) % len(q.queues)
					q.setInputs()
				} else if q.contentCursor > 0 {
					q.contentCursor--
				}
			case "down":
				if !q.controlContent {
					q.queueCursor = (q.queueCursor + 1) % len(q.queues)
					q.setInputs()
				} else if q.contentCursor < len(q.inputs)-1 {
					q.contentCursor++
				} else {
					q.buttonsCursor = 0
				}
			case "right":
				if q.buttonsCursor == -1 {
					q.controlContent = true
				} else {
					q.buttonsCursor = (q.buttonsCursor + 1) % 2
				}
			case "left":
				if q.buttonsCursor != -1 {
					if q.buttonsCursor > 0 {
						q.buttonsCursor--
					} else {
						q.buttonsCursor = 1
					}
				} else if q.controlContent {
					q.controlContent = false
					q.contentCursor = -1
				} else {
					return q, func() tea.Msg { return exitQueuesMsg{} }
				}
			case "n":
				q.creatingNewQueue = true
				for i := range q.inputs {
					q.inputs[i].SetValue("")
				}
				q.contentCursor = 0
				q.buttonsCursor = -1
			case "enter":
				if q.buttonsCursor != -1 {
					if q.buttonsCursor == 0 {
						currentQueue := q.queues[q.queueCursor]
						if currentQueue.CancelFunc == nil {
							go internal.ScheduleQueueDownloads(&currentQueue)
						} else if currentQueue.Paused {
							go currentQueue.ResumeQueue()
						} else {
							go currentQueue.PauseQueue()
						}
					} else if q.buttonsCursor == 1 {
						go q.deleteQueue()
					}
					q.buttonsCursor = -1
				} else if q.controlContent {
					q.inputs[q.contentCursor], _ = q.inputs[q.contentCursor].Update(msg)
					q.updateCurrentQueue()
				}
			default:
				if q.controlContent && q.contentCursor < len(q.inputs) {
					q.inputs[q.contentCursor], _ = q.inputs[q.contentCursor].Update(msg)
					q.updateCurrentQueue()
				}
			}
		} else {
			switch msg.String() {
			case "up":
				if q.contentCursor > 0 {
					q.contentCursor--
				}
			case "down":
				if q.contentCursor < len(q.inputs)-1 {
					q.contentCursor++
				}
			case "right":
				if q.contentCursor == len(q.inputs)-1 {
					q.buttonsCursor = 0
				}
			case "left":
				q.creatingNewQueue = false
			case "enter":
				if q.buttonsCursor != -1 {
					if q.buttonsCursor%2 == 0 {
						id := q.inputs[0].Value()
						directory := q.inputs[1].Value()
						retries, _ := strconv.Atoi(q.inputs[2].Value())
						bwLimit, _ := strconv.Atoi(q.inputs[3].Value())
						maxconcurrent, _ := strconv.Atoi(q.inputs[4].Value())
						startTime, _ := time.Parse("15:04", q.inputs[5].Value())
						endTime, _ := time.Parse("15:04", q.inputs[6].Value())
						now := time.Now()
						startTime = time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), 0, 0, now.Location())
						endTime = time.Date(now.Year(), now.Month(), now.Day(), endTime.Hour(), endTime.Minute(), 0, 0, now.Location())

						newQueue := internal.NewQueue(
							id,
							directory,
							retries,
							bwLimit,
							maxconcurrent,
							startTime,
							endTime,
						)
						newQueue.NumberOfTriesLimit = retries
						newQueue.BandwidthLimit = bwLimit
						newQueue.MaxConcurrentDownloads = maxconcurrent
						q.queues = append(q.queues, *newQueue)
						q.creatingNewQueue = false
						q.controlContent = false
						q.queueCursor = len(q.queues) - 1
						q.setInputs()
					} else {
						q.creatingNewQueue = false
					}
					return q, nil
				} else {
					q.inputs[q.contentCursor], _ = q.inputs[q.contentCursor].Update(msg)
				}
			default:
				if q.buttonsCursor == -1 && q.contentCursor < len(q.inputs) {
					q.inputs[q.contentCursor], _ = q.inputs[q.contentCursor].Update(msg)
				}
			}
		}
	case tickMsg:
		var sortedQueues []internal.Queue
		for _, queue := range internal.QueuesList {
			sortedQueues = append(sortedQueues, *queue)
		}
		sort.Slice(sortedQueues, func(i, j int) bool {
			return sortedQueues[i].Id < sortedQueues[j].Id
		})
		q.queues = sortedQueues
		if !q.creatingNewQueue && len(q.queues) > 0 {
			if q.queueCursor >= len(q.queues) {
				q.queueCursor = 0
			}
			q.setInputs()
		}
		return q, tickCmd()
	}
	return q, nil
}

var (
	style              = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	queueSelectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CC99"))
	selectedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#18FFFF")).Bold(true)
)

func (q *QueuesTab) View() string {
	if q.creatingNewQueue {
		form := ""
		labels := []string{"Name", "Directory", "Retries Limit", "Bandwidth Limit", "Max Concurrent Files Limit", "Start Time", "End Time"}
		for i, label := range labels {
			if q.contentCursor == i && q.buttonsCursor == -1 {
				form += selectedStyle.Render(label+": "+q.inputs[i].View()) + "\n"
				q.inputs[i].Focus()
			} else {
				form += style.Render(label+": "+q.inputs[i].View()) + "\n"
			}
		}
		buttons := []string{"Create", "Cancel"}
		for i, button := range buttons {
			if q.buttonsCursor != -1 && q.buttonsCursor%len(buttons) == i {
				form += selectedStyle.Render("["+button+"]") + "  "
			} else {
				form += style.Render("["+button+"]") + "  "
			}
		}
		return form + "\n" + "Press ← to go back to the queue list."
	}
	var queueList, queueContent string
	for i, queue := range q.queues {
		if q.queueCursor == i {
			queueList += queueSelectedStyle.Render(fmt.Sprintf("%8s |", queue.Id)) + "\n"
		} else {
			queueList += style.Render(fmt.Sprintf("%8s |", queue.Id)) + "\n"
		}
	}
	labels := []string{"Directory", "Retries Limit", "Bandwidth Limit", "Max Concurrent Files Limit", "Start Time", "End Time"}
	for i, label := range labels {
		if q.controlContent && q.contentCursor == i {
			queueContent += selectedStyle.Render(label+": "+q.inputs[i].View()) + "\n"
			q.inputs[i].Focus()
		} else {
			queueContent += style.Render(label+": "+q.inputs[i].View()) + "\n"
		}
	}
	stateButton := "[State]"
	deleteButton := "[Delete]"
	currentQueue := q.queues[q.queueCursor]
	if currentQueue.CancelFunc == nil {
		stateButton = "[Start]"
	} else if currentQueue.Paused {
		stateButton = "[Resume]"
	} else {
		stateButton = "[Pause]"
	}
	var buttonsLine string
	if q.buttonsCursor == 0 {
		buttonsLine += selectedStyle.Render(stateButton) + "  "
	} else {
		buttonsLine += style.Render(stateButton) + "  "
	}
	if q.buttonsCursor == 1 {
		buttonsLine += selectedStyle.Render(deleteButton)
	} else {
		buttonsLine += style.Render(deleteButton)
	}
	queueContent += "\n" + buttonsLine + "\n"

	footer := "Press 'n' for New Queue | '→' to edit fields | '←' to go back | '↑/↓' to navigate | 'Enter' to select"
	return lipgloss.JoinHorizontal(lipgloss.Top, queueList, "   ", queueContent) + "\n" + footer
}

func (q *QueuesTab) getFooter() string {
	return "Press '→' to select, '←' to go back, '↑ / ↓' to navigate"
}
