package main

import (
	"IDM/internal/controller/manager"
	"IDM/internal/download"
	"IDM/internal/tui"
	"bufio"
	"fmt"
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

	// TODO: program should extract the file name & format itself :(
	fileName := "download_output.mp3"
	workers := 4
	dm := manager.NewDownloadManager(url, fileName, workers)

	go func() {
		if err := dm.StartDownload(download.Download{}); err != nil {
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
func main() {
	tui.Start()
}
