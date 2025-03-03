package tui

import (
	"IDM/internal/download"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//todo show only 10 rows of download
//todo implement (retry, pause, resume, cancel) buttons

type DownloadsTab struct {
	downloads []download.Download
	cursor    int
	isActive  bool
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
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF79C6")). // Pink headers
			Background(lipgloss.Color("")).        // Dark background
			Bold(true).
			Padding(0, 1).
			Border(lipgloss.RoundedBorder())

	rowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")). // Light blue text
			Background(lipgloss.Color("")).        // Darker background
			Padding(0, 1)

	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8BE9FD")). // White text
				Background(lipgloss.Color("")).        // Blue background for selected row
				Bold(true).
				Padding(0, 1)
)

// Column widths
var colWidths = []int{12, 6, 10, 10, 12}

// NewDownloadsTab initializes the tab
func NewDownloadsTab() DownloadsTab {
	d := generateRandomDownloads(10)
	return DownloadsTab{downloads: d, cursor: -1}
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

// RenderTable creates a properly aligned Lipgloss table
func (d DownloadsTab) RenderTable() string {
	columns := []string{"Filename", "Queue", "Progress", "Time Left", "Status"}

	// Render header
	var headerRow []string
	for i, col := range columns {
		headerRow = append(headerRow, lipgloss.NewStyle().Width(colWidths[i]).Render(col))
	}
	table := headerStyle.Render(strings.Join(headerRow, " ")) + "\n"

	// Render rows
	for i, entry := range d.downloads {
		row := []string{
			lipgloss.NewStyle().Width(colWidths[0]).Render(entry.Filename),
			lipgloss.NewStyle().Width(colWidths[1]).Render(fmt.Sprintf("%d", i+1)),
			lipgloss.NewStyle().Width(colWidths[2]).Render(fmt.Sprintf("%d%%", entry.Progress)),
			lipgloss.NewStyle().Width(colWidths[3]).Render(calculateTimeLeft(entry)),
			lipgloss.NewStyle().Width(colWidths[4]).Render(string(entry.Status)),
		}

		// Apply selection style
		if i == d.cursor {
			table += selectedRowStyle.Render("  "+strings.Join(row, " ")) + "\n"
		} else {
			table += rowStyle.Render(" "+strings.Join(row, " ")) + "\n"
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
			d.isActive = false
			d.cursor = -1
		case "ctrl+c":
			return d, tea.Quit
		}
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
