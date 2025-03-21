package tui

import (
	"IDM/internal"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type exitDownloadsMsg struct{}

type DownloadsTab struct {
	downloads     []*internal.Download
	cursor        int
	isActive      bool
	pageStart     int
	showOptions   bool
	optionsCursor int
}

var (
	tableSize   = 10
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#18FFFF")).
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder())
	rowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)
	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Padding(0, 1)
)

var colWidths = []int{12, 10, 20, 10, 10, 10, 25}

func NewDownloadsTab() DownloadsTab {
	return DownloadsTab{downloads: internal.DownloadsList, cursor: -1, pageStart: 0}
}

func calculateTimeLeft(d internal.Download) string {
	if d.Status == internal.InProgress && d.Progress > 0 {
		downloaded := float64(d.FileSize) * (float64(d.Progress) / 100.0)
		elapsed := time.Since(d.StartTime).Seconds()
		if elapsed == 0 {
			return "N/A"
		}
		rate := downloaded / elapsed
		if rate <= 0 {
			return "N/A"
		}
		remainingBytes := float64(d.FileSize) - downloaded
		remainingSeconds := remainingBytes / rate
		return fmt.Sprintf("%.0fs", remainingSeconds)
	}
	return "N/A"
}

func getQueueName(download *internal.Download) string {
	if download.QueueName != "" {
		return download.QueueName
	}
	return "Unknown"
}

func getActions(status internal.Status) []string {
	switch status {
	case internal.Completed:
		return []string{"Delete"}
	case internal.Failed:
		return []string{"Retry", "Delete"}
	case internal.Paused:
		return []string{"Resume", "Cancel", "Delete"}
	case internal.InProgress:
		return []string{"Pause", "Cancel", "Delete"}
	case internal.Pending:
		return []string{"Start", "Delete", "Cancel"}
	case internal.Cancelled:
		return []string{"Delete"}
	default:
		return []string{"Start", "Delete", "Cancel"}
	}
}
func (d DownloadsTab) RenderTable() string {
	columns := []string{"Filename", "Queue", "Progress", "Time Left", "Status", "Speed", "Action"}
	var headerRow []string
	for i, col := range columns {
		headerRow = append(headerRow, lipgloss.NewStyle().Width(colWidths[i]).MaxWidth(colWidths[i]).Render(col))
	}
	table := headerStyle.Render(strings.Join(headerRow, " ")) + "\n"

	progressBar := progress.New(
		progress.WithWidth(colWidths[2]-2),
		progress.WithGradient("#0077BE", "#39FF14"),
		progress.WithFillCharacters('▬', '▬'),
	)

	for i, entry := range d.downloads {
		if i < d.pageStart || i >= d.pageStart+tableSize {
			continue
		}

		queueName := getQueueName(entry)
		actions := getActions(entry.Status)
		var actionStr string

		if len(actions) > 0 {
			if d.cursor == i && d.showOptions {
				var parts []string
				for j, act := range actions {
					if j == d.optionsCursor {
						parts = append(parts, selectedRowStyle.Render(act))
					} else {
						parts = append(parts, rowStyle.Render(act))
					}
				}
				actionStr = "[" + strings.Join(parts, "|") + "]"
			} else {
				actionStr = rowStyle.Render("[" + strings.Join(actions, "|") + "]")
			}
		}

		// Speed display: show "N/A" if not in progress
		speedStr := "N/A"
		if entry.Status == internal.InProgress {
			speedStr = formatSpeed(entry.Speed)
		}

		rowStyler := rowStyle
		if d.cursor == i {
			rowStyler = selectedRowStyle
		}

		columns := []string{
			rowStyler.Width(colWidths[0]).MaxWidth(colWidths[0]).Render(entry.FileName),
			rowStyler.Width(colWidths[1]).MaxWidth(colWidths[1]).Render(queueName),
			rowStyler.MaxWidth(colWidths[2]).Render(progressBar.ViewAs(float64(entry.Progress) / 100.0)),
			rowStyler.Width(colWidths[3]).Render(calculateTimeLeft(*entry)), // Ensure it's gray
			rowStyler.Width(colWidths[4]).Render(string(entry.Status)),      // Ensure it's gray
			rowStyler.Width(colWidths[5]).Render(speedStr),                  // Ensure it's gray
			actionStr,
		}
		rowContent := lipgloss.JoinHorizontal(lipgloss.Left, strings.Join(columns, " "))
		table += "  " + rowContent + "\n"
	}

	return table
}

func (d DownloadsTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if d.cursor > 0 {
				d.cursor--
			} else {
				d.cursor = len(d.downloads) - 1
			}
		case "down":
			if d.cursor < len(d.downloads)-1 {
				d.cursor++
			} else {
				d.cursor = 0
			}
		case "right":
			if !d.showOptions {
				d.showOptions = true
				d.optionsCursor = 0
			} else {
				actions := getActions(d.downloads[d.cursor].Status)
				if len(actions) > 0 {
					if d.optionsCursor < len(actions)-1 {
						d.optionsCursor++
					} else {
						d.optionsCursor = 0
					}
				}
			}
		case "left":
			if d.showOptions {
				d.showOptions = false
				d.optionsCursor = -1
			} else {
				d.cursor = -1
				d.isActive = false
				return d, func() tea.Msg { return exitDownloadsMsg{} }
			}
		case "enter":
			if d.showOptions {
				actions := getActions(d.downloads[d.cursor].Status)
				if len(actions) > 0 && d.optionsCursor < len(actions) {
					selectedAction := actions[d.optionsCursor]
					switch selectedAction {
					case "Start":
						go d.downloads[d.cursor].StartDownload()
					case "Pause":
						go d.downloads[d.cursor].PauseDownload()
					case "Resume":
						go d.downloads[d.cursor].ResumeDownload()
					case "Cancel":
						go d.downloads[d.cursor].CancelDownload()
					case "Retry":
						//go d.downloads[d.cursor].Retry()
					case "Delete":
						selectedDL := d.downloads[d.cursor]
						if queue, exists := internal.QueuesList[selectedDL.QueueName]; exists {
							err := queue.RemoveDownload(selectedDL.FileName)
							if err == nil {
								newDownloads := make([]*internal.Download, 0)
								for _, dl := range internal.DownloadsList {
									if dl.FileName != selectedDL.FileName {
										newDownloads = append(newDownloads, dl)
									}
								}
								internal.DownloadsList = newDownloads
								d.downloads = internal.DownloadsList
								if d.cursor >= len(d.downloads) {
									d.cursor = max(0, len(d.downloads)-1)
								}
								if d.cursor >= len(d.downloads) {
									d.cursor = max(0, len(d.downloads)-1)
								}
							}
						}
					default:
						fmt.Printf("Selected action: %s on %s\n", selectedAction, d.downloads[d.cursor].FileName)
					}
				}
				d.showOptions = false
			} else {
				d.showOptions = true
				d.optionsCursor = 0
			}
		}
	case tickMsg:
		d.downloads = make([]*internal.Download, len(internal.DownloadsList))
		copy(d.downloads, internal.DownloadsList)
		for i, dl := range d.downloads {
			if queue, exists := internal.QueuesList[dl.QueueName]; exists {
				for _, qDL := range queue.Downloads {
					if qDL.FileName == dl.FileName {
						d.downloads[i].Progress = qDL.Progress
						d.downloads[i].Status = qDL.Status
						d.downloads[i].FileSize = qDL.FileSize
						break
					}
				}
			}
		}
		return d, tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return d, nil
}

func (d DownloadsTab) View() string {
	return d.RenderTable()
}

func (d DownloadsTab) setActive(b bool) Tab {
	d.isActive = b
	if b {
		d.cursor = 0
	} else {
		d.cursor = -1
	}
	return d
}

func (d DownloadsTab) isActivated() bool {
	return d.isActive
}

func (d DownloadsTab) toString() string {
	return "Downloads"
}

func (d DownloadsTab) Init() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (d DownloadsTab) getFooter() string {
	return "Use '↑/↓' to navigate rows, '→' to select actions, 'Enter' to execute"
}

func formatSpeed(speed float64) string {
	if speed < 1024 {
		return fmt.Sprintf("%.0f B/s", speed)
	} else if speed < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", speed/1024)
	} else {
		return fmt.Sprintf("%.1f MB/s", speed/(1024*1024))
	}
}
