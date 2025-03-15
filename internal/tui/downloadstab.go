package tui

import (
	"IDM/internal"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DownloadsTab struct {
	downloads     []*internal.Download
	cursor        int
	isActive      bool
	pageStart     int
	showOptions   bool
	optionsCursor int
}

func generateRandomDownloads(n int) []internal.Download {
	downloads := make([]internal.Download, n)
	for i := 0; i < n; i++ {
		downloads[i] = internal.Download{
			FileName:   fmt.Sprintf("File %d", i),
			Status:     internal.InProgress,
			Progress:   int64(i * 10),
			TotalBytes: int64(100),
			StartTime:  time.Now(),
		}
	}
	return downloads
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

func NewDownloadsTab() DownloadsTab {
	return DownloadsTab{downloads: internal.DownloadsList, cursor: 0, pageStart: 0}
}

func calculateTimeLeft(d internal.Download) string {
	if d.Status == internal.InProgress && d.Progress > 0 {
		elapsed := time.Since(d.StartTime).Seconds()
		total := float64(d.TotalBytes) / float64(d.Progress) * 100
		remaining := total - elapsed
		return fmt.Sprintf("%.0fs", remaining)
	}
	return "N/A"
}

var colWidths = []int{15, 10, 30, 12, 12, 15}

func getQueueName(download *internal.Download) string {
	for _, queue := range internal.QueuesList {
		for _, d := range queue.Downloads {
			if d.FileName == download.FileName {
				return queue.Id
			}
		}
	}
	return "Unknown"
}

func getActions(status internal.Status) []string {
	switch status {
	case internal.Completed:
		return []string{"Delete", "Back"}
	case internal.Failed:
		return []string{"Retry", "Delete", "Back"}
	case internal.Paused:
		return []string{"Resume", "Cancel", "Back"}
	case internal.InProgress:
		return []string{"Pause", "Cancel", "Back"}
	default:
		return []string{"Start"}
	}
}

func (d DownloadsTab) RenderTable() string {
	columns := []string{"Filename", "Queue", "Progress", "Time Left", "Status", "Action"}
	var headerRow []string
	for i, col := range columns {
		headerRow = append(headerRow, lipgloss.NewStyle().Width(colWidths[i]).Render(col))
	}
	table := headerStyle.Render(strings.Join(headerRow, " ")) + "\n"

	progressBar := progress.New(
		progress.WithWidth(colWidths[2]*90/100),
		progress.WithGradient("#0077BE", "#39FF14"),
		progress.WithFillCharacters('▬', '▬'),
		progress.WithoutPercentage(),
	)

	// Render rows.
	for i, entry := range d.downloads {
		if i < d.pageStart || i >= d.pageStart+tableSize {
			continue
		}

		queueName := getQueueName(entry)
		actions := getActions(entry.Status)
		defaultAction := ""
		if len(actions) > 0 {
			defaultAction = actions[0]
		}
		rowStr := []string{
			entry.FileName,
			queueName,
			progressBar.ViewAs(float64(entry.Progress) / 100.0),
			calculateTimeLeft(*entry),
			string(entry.Status),
			defaultAction,
		}
		var renderedRow string
		if i == d.cursor {
			renderedRow = selectedRowStyle.Copy().Width(colWidths[0]).Render(rowStr[0]) + " " +
				selectedRowStyle.Copy().Width(colWidths[1]).Render(rowStr[1]) + " " +
				selectedRowStyle.Copy().Width(colWidths[2]).Render(rowStr[2]) + " " +
				selectedRowStyle.Copy().Width(colWidths[3]).Render(rowStr[3]) + " " +
				selectedRowStyle.Copy().Width(colWidths[4]).Render(rowStr[4]) + " "
			if d.showOptions {
				if d.optionsCursor >= len(actions) {
					d.optionsCursor = 0
				}
				var styledActions []string
				for j, action := range actions {
					if j == d.optionsCursor {
						styledActions = append(styledActions, selectedRowStyle.Render(action))
					} else {
						styledActions = append(styledActions, rowStyle.Render(action))
					}
				}
				renderedRow += "[" + strings.Join(styledActions, "|") + "]"
			} else {
				renderedRow += selectedRowStyle.Copy().Width(colWidths[5]).Render(defaultAction)
			}
			table += "  " + renderedRow + "\n"
		} else {
			// Non-selected rows.
			renderedRow = rowStyle.Copy().Width(colWidths[0]).Render(rowStr[0]) + " " +
				rowStyle.Copy().Width(colWidths[1]).Render(rowStr[1]) + " " +
				rowStyle.Copy().Width(colWidths[2]).Render(rowStr[2]) + " " +
				rowStyle.Copy().Width(colWidths[3]).Render(rowStr[3]) + " " +
				rowStyle.Copy().Width(colWidths[4]).Render(rowStr[4]) + " " +
				rowStyle.Copy().Width(colWidths[5]).Render(rowStr[5])
			table += " " + renderedRow + "\n"
		}
	}

	return table
}

func (d DownloadsTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if len(d.downloads) == 0 {
			return d, nil
		}
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
		case "left":
			if d.showOptions {
				actions := getActions(d.downloads[d.cursor].Status)
				if len(actions) > 0 {
					if d.optionsCursor > 0 {
						d.optionsCursor--
					} else {
						d.optionsCursor = len(actions) - 1
					}
				}
			} else {
				d.cursor = -1
				d.isActive = false
			}
		case "right":
			if d.showOptions {
				actions := getActions(d.downloads[d.cursor].Status)
				if len(actions) > 0 {
					if d.optionsCursor < len(actions)-1 {
						d.optionsCursor++
					} else {
						d.optionsCursor = 0
					}
				}
			}
		case "enter":
			if d.showOptions {
				actions := getActions(d.downloads[d.cursor].Status)
				if len(actions) > 0 && d.optionsCursor < len(actions) {
					selectedAction := actions[d.optionsCursor]
					switch selectedAction {
					case "Pause":
						d.downloads[d.cursor].PauseDownload()
					case "Resume":
						d.downloads[d.cursor].ResumeDownload()
					case "Start":
						{
							d.downloads[d.cursor].StartDownload()
						}
						fmt.Println(d.downloads[d.cursor].FileName)
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
		downloadMap := make(map[string]*internal.Download)
		for _, dl := range d.downloads {
			downloadMap[dl.FileName] = dl
		}
		newDownloads := make([]*internal.Download, 0)
		for _, queue := range internal.QueuesList {
			for _, download := range queue.Downloads {
				if existing, found := downloadMap[download.FileName]; found {
					existing.Progress = download.Progress
					existing.Status = download.Status
					existing.TotalBytes = download.TotalBytes
				} else {
					newDownloads = append(newDownloads, download)
				}
			}
		}
		d.downloads = append(d.downloads, newDownloads...)
		return d, tickCmd()
	}

	if d.cursor < d.pageStart {
		d.pageStart = d.cursor
	} else if d.pageStart+tableSize <= d.cursor {
		d.pageStart = d.cursor - tableSize + 1
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
	return nil
}

func (d DownloadsTab) getFooter() string {
	return "Use '↑ / ↓' to navigate, Press 'Enter' to show options"
}
