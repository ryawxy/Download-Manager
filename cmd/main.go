package main

import (
	"IDM/internal"
	"IDM/internal/tui"
	"bufio"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
	"strings"
)

func theirMain() {

	//var d download.Download = download.Download{
	//	URL:      "asdfffgg",
	//	Filename: "asdfffgg11",
	//}
	//err := manager.SaveData(d)
	//if err != nil {
	//	return
	//}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter download URL: ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	workers := 4
	dm := internal.NewDownloadManager(url, "", workers)

	err := dm.GetFileSizeAndName()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	fmt.Printf("Downloading file: %s\n", dm.FileName)

	go func() {
		if err := dm.StartDownload(internal.Download{}); err != nil {
			fmt.Println("Error:", err)
		}
	}()

	for {
		fmt.Print("Enter command (pause/resume/cancel): ")
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)

		switch cmd {
		case "pause":
			dm.Mutex.Lock()
			dm.Paused = true
			dm.Mutex.Unlock()
			fmt.Println("Download paused.")
		case "resume":
			dm.Mutex.Lock()
			dm.Paused = false
			dm.Mutex.Unlock()
			fmt.Println("Download resumed.")
		case "cancel":
			dm.CancelDownload()
			return
		default:
			fmt.Println("Unknown command.")
		}
	}
}

func myMain() {
	model := tui.NewMainStage()

	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	//myMain()
	theirMain()
}
