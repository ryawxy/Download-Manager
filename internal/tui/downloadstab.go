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

type exitDownloadsMsg struct{}

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
			FileName:  fmt.Sprintf("File %d", i),
			Status:    internal.InProgress,
			Progress:  int64(i * 10),
			FileSize:  int64(100),
			StartTime: time.Now(),
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

var colWidths = []int{15, 10, 30, 12, 12, 20}

func NewDownloadsTab() DownloadsTab {
	return DownloadsTab{downloads: internal.DownloadsList, cursor: 0, pageStart: 0}
}

func calculateTimeLeft(d internal.Download) string {
	if d.Status == internal.InProgress && d.Progress > 0 {
		elapsed := time.Since(d.StartTime).Seconds()
		total := float64(d.FileSize) / float64(d.Progress) * 100
		remaining := total - elapsed
		return fmt.Sprintf("%.0fs", remaining)
	}
	return "N/A"
}

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
		return []string{"Delete"}
	case internal.Failed:
		return []string{"Retry", "Delete"}
	case internal.Paused:
		return []string{"Resume", "Cancel"}
	case internal.InProgress:
		return []string{"Pause", "Cancel"}
	case internal.Pending:
		return []string{"Start"}
	default:
		return []string{"Start"}
	}
}

func (d DownloadsTab) RenderTable() string {
	// Define header columns including the new "Action" column.
	columns := []string{"Filename", "Queue", "Progress", "Time Left", "Status", "Action"}
	var headerRow []string
	for i, col := range columns {
		headerRow = append(headerRow, lipgloss.NewStyle().Width(colWidths[i]).MaxWidth(colWidths[i]).Render(col))
	}
	table := headerStyle.Render(strings.Join(headerRow, " ")) + "\n"

	progressBar := progress.New(
		progress.WithWidth(colWidths[2]-2), // Account for padding
		progress.WithGradient("#0077BE", "#39FF14"),
		progress.WithFillCharacters('▬', '-'),
	)

	// Render each row.
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
				actionStr = "[" + strings.Join(actions, "|") + "]"
			}
		}

		// Create styled columns with strict width constraints
		columns := []string{
			lipgloss.NewStyle().
				Width(colWidths[0]).
				MaxWidth(colWidths[0]).
				Render(entry.FileName),
			lipgloss.NewStyle().
				Width(colWidths[1]).
				Render(queueName),
			lipgloss.NewStyle().
				Width(colWidths[2]).
				Render(progressBar.ViewAs(float64(entry.Progress) / 100.0)),
			lipgloss.NewStyle().
				Width(colWidths[3]).
				Render(calculateTimeLeft(*entry)),
			lipgloss.NewStyle().
				Width(colWidths[4]).
				Render(string(entry.Status)),
			lipgloss.NewStyle().
				Width(colWidths[5]).
				MaxWidth(colWidths[5]).
				Render(actionStr),
		}

		// Join columns horizontally
		rowContent := lipgloss.JoinHorizontal(
			lipgloss.Left,
			columns[0], " ",
			columns[1], " ",
			columns[2], " ",
			columns[3], " ",
			columns[4], " ",
			columns[5],
		)

		// Apply row styling
		if i == d.cursor {
			table += "  " + selectedRowStyle.Render(rowContent) + "\n"
		} else {
			table += " " + rowStyle.Render(rowContent) + "\n"
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
				actions := getActions(d.downloads[d.cursor].Status)
				if len(actions) > 0 {
					if d.optionsCursor > 0 {
						d.optionsCursor--
					} else {
						d.optionsCursor = len(actions) - 1
					}
				}
			} else {
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
						d.downloads[d.cursor].StartDownload()
					case "Pause":
						d.downloads[d.cursor].PauseDownload()
					case "Resume":
						d.downloads[d.cursor].ResumeDownload()
					case "Cancel":
						d.downloads[d.cursor].CancelDownload()
					case "Retry":
						d.downloads[d.cursor].Retry()
					case "Delete":
						// TODO: Implement deletion logic.
						fmt.Printf("Delete %s\n", d.downloads[d.cursor].FileName)
					default:
						fmt.Printf("Action %s on %s\n", selectedAction, d.downloads[d.cursor].FileName)
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
					existing.FileSize = download.FileSize
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
	return "Use '↑/↓' to navigate rows, '→' to select actions, 'Enter' to execute"
}
