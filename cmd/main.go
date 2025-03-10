package main

import (
	"IDM/internal"
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command (create <queue_name> <max_concurrent_downloads> <start_time> <end_time> / edit <queue_name> <max_concurrent_downloads> <start_time> <end_time> <bandwidth_limit> / delete <queue_name> / list / done): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "done" {
			break
		}

		parts := strings.SplitN(input, " ", 6)
		if len(parts) == 1 && parts[0] == "list" {
			internal.ListQueues()
		} else if len(parts) == 5 && parts[0] == "create" {
			queueName := parts[1]
			var maxConcurrent int
			fmt.Sscanf(parts[2], "%d", &maxConcurrent)

			startTime, err1 := time.Parse("15:04", parts[3])
			endTime, err2 := time.Parse("15:04", parts[4])
			if err1 != nil || err2 != nil {
				fmt.Println("Invalid time format! Use HH:MM (24-hour format)")
				continue
			}

			now := time.Now()
			startTime = time.Date(now.Year(), now.Month(), now.Day(), startTime.Hour(), startTime.Minute(), 0, 0, now.Location())
			endTime = time.Date(now.Year(), now.Month(), now.Day(), endTime.Hour(), endTime.Minute(), 0, 0, now.Location())

			if _, exists := internal.QueuesList[queueName]; exists {
				fmt.Println("Queue already exists!")
			} else {
				internal.QueuesList[queueName] = internal.NewQueue(queueName, "./downloads", 10, 1024*1024, maxConcurrent, startTime, endTime)
				fmt.Println("Queue created:", queueName, "with max concurrent downloads:", maxConcurrent, "Start:", startTime.Format("15:04"), "End:", endTime.Format("15:04"))
			}
		} else if len(parts) == 6 && parts[0] == "edit" {
			queueName := parts[1]
			var maxConcurrent int
			fmt.Sscanf(parts[2], "%d", &maxConcurrent)

			startTime, err1 := time.Parse("15:04", parts[3])
			endTime, err2 := time.Parse("15:04", parts[4])

			var bandwidthLimit int
			bandwidthLimit, err3 := fmt.Sscanf(parts[5], "%d", &bandwidthLimit)

			if err1 != nil || err2 != nil || err3 != nil {
				fmt.Println("Invalid time or bandwidth format!")
				continue
			}

			if queue, exists := internal.QueuesList[queueName]; exists {
				queue.EditQueue(maxConcurrent, startTime, endTime, bandwidthLimit)
			} else {
				fmt.Println("Queue not found!")
			}
		} else if len(parts) == 2 && parts[0] == "delete" {
			internal.DeleteQueue(parts[1])
		} else {
			fmt.Println("Invalid command.")
		}
	}

	for {
		fmt.Print("Enter download URL (or 'done' to finish): ")
		url, _ := reader.ReadString('\n')
		url = strings.TrimSpace(url)

		if url == "done" {
			break
		}

		download := &internal.Download{URL: url, FileName: ""}
		internal.DownloadList = append(internal.DownloadList, download)
		fmt.Println("Added download:", url)
	}
	fmt.Print("Enter queue name to assign downloads: ")
	queueName, _ := reader.ReadString('\n')
	queueName = strings.TrimSpace(queueName)

	queue, exists := internal.QueuesList[queueName]
	if !exists {
		fmt.Println("Queue not found!")
		return
	}

	for _, d := range internal.DownloadList {
		queue.AddDownload(d)
	}
	fmt.Println("Downloads assigned to queue:", queueName)
	for {
		fmt.Print("Enter command (start <queue_name> / list / exit): ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		parts := strings.SplitN(command, " ", 2)
		if len(parts) == 2 && parts[0] == "start" {
			qName := parts[1]
			if q, ok := internal.QueuesList[qName]; ok {
				internal.ScheduleQueueDownloads(q)
			} else {
				fmt.Println("Queue not found!")
			}
		} else if command == "list" {
			internal.ListQueues()
		} else if command == "exit" {
			fmt.Println("Exiting...")
			break
		} else {
			fmt.Println("Invalid command.")
		}
	}
}
