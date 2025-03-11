package tui

import (
	"IDM/internal/download"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DownloadsTab struct {
	downloads     []download.Download
	cursor        int
	isActive      bool
	pageStart     int
	showOptions   bool
	optionsCursor int
}

// Generate sample downloads
func generateRandomDownloads(n int) []download.Download {
	downloads := make([]download.Download, n)
	for i := 0; i < n; i++ {
		downloads[i] = download.Download{
			Filename:   fmt.Sprintf("File %d", i),
			Status:     download.Status(download.InProgress),
			Progress:   int64(i * 10),
			TotalBytes: int64(100),
			StartTime:  time.Now(),
		}
	}
	return downloads
}

// Styles
var (
	tableSize   = 10
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#18FFFF")). // Pink headers
			Background(lipgloss.Color("")).        // Dark background
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder())

	rowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("")).
			Padding(0, 1)

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")). // White text
				Background(lipgloss.Color("")).        // Blue background for selected row
				Bold(true).
				Padding(0, 1)
)

// NewDownloadsTab initializes the tab
func NewDownloadsTab() DownloadsTab {
	d := generateRandomDownloads(20)

	return DownloadsTab{downloads: d, cursor: -1, pageStart: 0}
}

// calculateTimeLeft estimates time left for a download
func calculateTimeLeft(d download.Download) string {
	if d.Status == download.InProgress && d.Progress > 0 {
		elapsed := time.Since(d.StartTime).Seconds()
		total := float64(d.TotalBytes) / float64(d.Progress) * 100
		remaining := total - elapsed
		return fmt.Sprintf("%.0fs", remaining)
	}
	return "N/A"
}

// Column widths
var colWidths = []int{10, 6, 30, 10, 12}

func getActions(status download.Status) []string {
	switch status {
	case download.Completed:
		return []string{"Delete", "Back"}
	case download.Failed:
		return []string{"Retry", "Delete", "Back"}
	case download.Paused:
		return []string{"Cancel", "Resume", "Back"}
	case download.InProgress:
		return []string{"Pause", "Cancel", "Back"}
	default:
		return []string{}
	}
}

// RenderTable creates a properly aligned Lipgloss table
func (d DownloadsTab) RenderTable() string {
	columns := []string{"Filename", "Queue", "Progress", "Time Left", "Status"}

	// Render header
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
	// Render rows
	for i, entry := range d.downloads {
		if i < d.pageStart || i >= d.pageStart+tableSize {
			continue
		}
		styledRow := []string{
			selectedRowStyle.Copy().Width(colWidths[0]).Render(entry.Filename[max(len(entry.Filename)-20, 0):]),
			selectedRowStyle.Copy().Width(colWidths[1]).Render(fmt.Sprintf("%d", i+1)),
			selectedRowStyle.Copy().Width(colWidths[2]).Render(progressBar.ViewAs(float64(entry.Progress) / 100.0)),
			selectedRowStyle.Copy().Width(colWidths[3]).Render(calculateTimeLeft(entry)),
			selectedRowStyle.Copy().Width(colWidths[4]).Render(string(entry.Status)),
		}

		if i == d.cursor {
			table += "  " + strings.Join(styledRow, " ")
			if d.showOptions {
				actions := getActions(entry.Status)
				if len(actions) > 0 {
					var styledActions []string
					for j, action := range actions {
						if j == d.optionsCursor {
							styledActions = append(styledActions, selectedRowStyle.Render(action))
						} else {
							styledActions = append(styledActions, rowStyle.Render(action))
						}
					}
					// Join the styled actions with a separator and render them inside square brackets.
					actionStr := "[" + strings.Join(styledActions, " | ") + "]"
					table += "   " + actionStr
				}
			}
			table += "\n"
		} else {
			row := []string{
				rowStyle.Copy().Width(colWidths[0]).Render(entry.Filename),
				rowStyle.Copy().Width(colWidths[1]).Render(fmt.Sprintf("%d", i+1)),
				rowStyle.Copy().Width(colWidths[2]).Render(progressBar.ViewAs(float64(entry.Progress) / 100.0)),
				rowStyle.Copy().Width(colWidths[3]).Render(calculateTimeLeft(entry)),
				rowStyle.Copy().Width(colWidths[4]).Render(string(entry.Status)),
			}
			table += " " + strings.Join(row, " ") + "\n"
		}
	}

	return table
}

// Update handles user input
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
		case "left":
			if d.showOptions {
				if d.optionsCursor > 0 {
					d.optionsCursor--
				} else {
					d.optionsCursor = len(getActions(d.downloads[d.cursor].Status)) - 1
				}
			} else {
				d.cursor = -1
				d.isActive = false
			}
		case "right":
			if d.showOptions {
				if d.optionsCursor < len(getActions(d.downloads[d.cursor].Status))-1 {
					d.optionsCursor++
				} else {
					d.optionsCursor = 0
				}
			}
		case "enter":
			if d.showOptions {
				d.optionsCursor = 0
				//todo call the correct function
			}
			d.showOptions = !d.showOptions
		}
	}
	if d.cursor < d.pageStart {
		d.pageStart = d.cursor
	} else if d.pageStart+tableSize <= d.cursor {
		d.pageStart = d.cursor - tableSize + 1
	}
	return d, nil
}

// View renders the table
func (d DownloadsTab) View() string {
	return d.RenderTable()
}

// Tab interface methods
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
	return "Use '↑ / ↓' to navigate, " + "Press 'Enter' to show options"
}
