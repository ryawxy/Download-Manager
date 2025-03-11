package internal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Queue struct {
	Id                     string       `json:"id"`
	Downloads              []*Download  `json:"downloads"`
	Directory              string       `json:"directory"`
	NumberOfFilesLimit     int          `json:"number_of_files_limit"`
	BandwidthLimit         int          `json:"bandwidth_limit"`
	NumberOfTriesLimit     int          `json:"number_of_tries_limit"`
	StartTime              time.Time    `json:"start_time"`
	EndTime                time.Time    `json:"end_time"`
	MaxConcurrentDownloads int          `json:"max_concurrent_downloads"`
	TokenBucket            *TokenBucket `json:"-"`
	mutex                  sync.Mutex   `json:"-"`
	CancelFunc             func()       `json:"-"`
}

func NewQueue(id, directory string, numberOfFilesLimit int, bandwidthLimit int, maxConcurrent int, startTime, endTime time.Time) *Queue {
	rate := time.Second / time.Duration(bandwidthLimit)

	q := &Queue{
		Id:                     id,
		Downloads:              make([]*Download, 0),
		NumberOfFilesLimit:     numberOfFilesLimit,
		Directory:              directory,
		BandwidthLimit:         bandwidthLimit,
		MaxConcurrentDownloads: maxConcurrent,
		StartTime:              startTime,
		EndTime:                endTime,
		TokenBucket:            NewTokenBucket(bandwidthLimit, rate),
	}
	QueuesList[id] = q
	_ = SaveQueuesToFile()
	return q
}

func (q *Queue) StopDownloads() {
	if q.CancelFunc != nil {
		q.CancelFunc()
		fmt.Println("Downloads in queue", q.Id, "stopped due to end time.")
	}
}

func (q *Queue) EditQueue(maxConcurrent int, startTime, endTime time.Time, bandwidthLimit int) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.MaxConcurrentDownloads = maxConcurrent
	q.StartTime = startTime
	q.EndTime = endTime
	q.BandwidthLimit = bandwidthLimit

	rate := time.Second / time.Duration(bandwidthLimit)
	q.TokenBucket = NewTokenBucket(bandwidthLimit, rate)

	_ = SaveQueuesToFile()
	fmt.Println("Queue", q.Id, "updated successfully.")
	return nil
}

func (q *Queue) AddDownload(d *Download) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for _, queue := range QueuesList {
		if queue.Id == q.Id {
			d.Status = Pending
			q.Downloads = append(q.Downloads, d)
			_ = SaveQueuesToFile()
		}
	}

	return nil
}

func (q *Queue) RemoveDownload(name string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for i, d := range q.Downloads {
		if d.FileName == name {
			q.Downloads = append(q.Downloads[:i], q.Downloads[i+1:]...)
			_ = SaveQueuesToFile()
			return nil
		}
	}
	return errors.New("download not found")
}
func StartQueueDownloads(q *Queue) {
	fmt.Println("Starting downloads in queue:", q.Id)

	sem := make(chan struct{}, q.MaxConcurrentDownloads)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	q.CancelFunc = cancel

	time.AfterFunc(time.Until(q.EndTime), func() {
		q.StopDownloads()
	})

	for _, d := range q.Downloads {
		wg.Add(1)
		sem <- struct{}{}

		go func(d *Download) {
			defer wg.Done()
			dm := NewDownloadManager(d.URL, d.FileName, 4)
			dm.Ctx = ctx

			err := dm.GetFileSizeAndName()
			if err != nil {
				fmt.Println("Error getting file info for", d.URL)
				<-sem
				return
			}

			fmt.Printf("Downloading file: %s\n", dm.FileName)
			if err := dm.StartDownload(*d); err != nil {
				fmt.Println("Error:", err)
			}

			<-sem
		}(d)
	}

	wg.Wait()
	fmt.Println("All downloads completed in queue:", q.Id)
}
func ScheduleQueueDownloads(q *Queue) {
	now := time.Now()
	delay := q.StartTime.Sub(now)
	if delay <= 0 {
		StartQueueDownloads(q)
	} else {
		fmt.Println("Queue", q.Id, "will start in", delay)
		time.AfterFunc(delay, func() {
			StartQueueDownloads(q)
		})
	}
}
func ListQueues() {
	fmt.Println("\n----- Available QueuesList -----")
	if len(QueuesList) == 0 {
		fmt.Println("No queues available.")
		return
	}

	for name, q := range QueuesList {
		fmt.Printf("Queue Name: %s\n", name)
		fmt.Printf("  - Number of Downloads: %d\n", len(q.Downloads))
		fmt.Printf("  - Max Concurrent Downloads: %d\n", q.MaxConcurrentDownloads)
		fmt.Printf("  - Bandwidth Limit: %d bytes/sec\n", q.BandwidthLimit)
		fmt.Printf("  - File Limit: %d\n", q.NumberOfFilesLimit)
		fmt.Printf("  - Start Time: %s\n", q.StartTime.Format("15:04:05"))
		fmt.Printf("  - End Time: %s\n", q.EndTime.Format("15:04:05"))
		fmt.Println("-------------------------------")
	}
}
func DeleteQueue(queueName string) {
	if _, exists := QueuesList[queueName]; exists {
		delete(QueuesList, queueName)
		_ = SaveQueuesToFile()
		fmt.Println("Queue", queueName, "deleted successfully.")
	} else {
		fmt.Println("Queue not found!")
	}
}
