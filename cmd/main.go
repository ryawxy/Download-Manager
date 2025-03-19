package main

import (
	"IDM/internal"
	"IDM/internal/tui"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func init() {
	err := internal.LoadQueuesFromFile()
	if err != nil {
		fmt.Println("Error loading queues:", err)
	}
	//	tui.Start()
}

func main() {
	tui.Start()
	//cliMain()
}
func cliMain() {
	err := internal.LoadQueuesFromFile()
	if err != nil {
		fmt.Println("Error loading queues:", err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command (create <queue_name> <max_concurrent_downloads> <start_time> <end_time> <directory> <bandwidthlimit>  \n/ edit <queue_name> <max_concurrent_downloads> <start_time> <end_time> <bandwidth_limit> / delete <queue_name> / list / done): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "done" {
			break
		}

		parts := strings.SplitN(input, " ", 7)
		if len(parts) == 1 && parts[0] == "list" {
			internal.ListQueues()
		} else if len(parts) == 7 && parts[0] == "create" {
			fmt.Println("###################")
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
				i, err := strconv.Atoi(parts[6])
				if err != nil {
					fmt.Println("Error converting string to integer:", err)
					return
				}
				internal.QueuesList[queueName] = internal.NewQueue(queueName, parts[5], 10, i, maxConcurrent, startTime, endTime)
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
				queue.EditQueue(queue.Directory, queue.NumberOfTriesLimit, maxConcurrent, startTime, endTime, bandwidthLimit)
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
		internal.DownloadsList = append(internal.DownloadsList, download)
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

	for _, d := range internal.DownloadsList {
		d.Directory = queue.Directory
		queue.AddDownload(d)
	}
	fmt.Println("Downloads assigned to queue:", queueName)
	for {
		fmt.Print("Enter command (start <queue_name> / pause <queue_name> / resume <queue_name> / retry <queue_name> <filename> / list / exit): ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		parts := strings.SplitN(command, " ", 3)
		if len(parts) == 2 {
			qName := parts[1]
			if q, ok := internal.QueuesList[qName]; ok {
				if parts[0] == "pause" {
					q.PauseQueue()
				} else if parts[0] == "resume" {
					q.ResumeQueue()
				} else if parts[0] == "start" {
					internal.ScheduleQueueDownloads(q)
				} else if parts[0] == "retry" {
					if len(parts) != 3 {
						fmt.Println("Usage: retry <queue_name> <filename>")
						continue
					}
					filename := parts[2]
					for _, d := range q.Downloads {
						if d.FileName == filename {
							if err := d.Retry(q); err != nil {
								fmt.Println("Retry error:", err)
							}
							break
						}
					}
				}
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
