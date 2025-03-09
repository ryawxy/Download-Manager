package main

import (
	"IDM/internal"
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter download URL: ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	// TODO: program should extract the file name & format itself :(
	fileName := "download_output.pdf"
	workers := 4
	dm := internal.NewDownloadManager(url, fileName, workers)

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
	//model := tui.NewMainStage()
	//p := tea.NewProgram(model)
	//if err := p.Start(); err != nil {
	//	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	//	os.Exit(1)
}
