package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type Queue struct {
	Id                     string       `json:"id"`
	Downloads              []*Download  `json:"downloads"`
	Directory              string       `json:"directory"`
	DirectorySet           bool         `json:"directory_set"`
	BandwidthLimit         int          `json:"bandwidth_limit"`
	BandwidthSet           bool         `json:"bandwidth_set"`
	NumberOfTriesLimit     int          `json:"number_of_tries_limit"`
	RetriesSet             bool         `json:"retries_set"`
	StartTime              time.Time    `json:"start_time"`
	StartTimeSet           bool         `json:"start_time_set"`
	EndTime                time.Time    `json:"end_time"`
	EndTimeSet             bool         `json:"end_time_set"`
	MaxConcurrentDownloads int          `json:"max_concurrent_downloads"`
	MaxConcurrentSet       bool         `json:"max_concurrent_set"`
	TokenBucket            *TokenBucket `json:"-"`
	mutex                  sync.Mutex   `json:"-"`
	CancelFunc             func()       `json:"-"`
	Paused                 bool         `json:"paused"`
	HasStarted             bool         `json:"has_started"`
}

var QueuesList = make(map[string]*Queue)

func NewQueue(id, directory string, retriesLimit int, bandwidthLimit int,
	maxConcurrent int, startTime, endTime time.Time) *Queue {

	// invalid or duplicate name - only field which required here
	if id == "" || QueuesList[id] != nil {
		return nil
	}

	q := &Queue{
		Id:         id,
		Downloads:  make([]*Download, 0),
		HasStarted: false,
		//Directory:              directory,
		//NumberOfTriesLimit:     retriesLimit,
		//BandwidthLimit:         bandwidthLimit,
		//MaxConcurrentDownloads: maxConcurrent,
		//StartTime:              startTime,
		//EndTime:                endTime,
		//TokenBucket:            NewTokenBucket(bandwidthLimit, rate),
	}

	// Set fields only if provided
	if directory != "" {
		q.Directory = directory
		q.DirectorySet = true
	}
	if retriesLimit != 0 { // 0 means unset
		q.NumberOfTriesLimit = retriesLimit
		q.RetriesSet = true
	}
	if bandwidthLimit != 0 { // 0 means unset, no TokenBucket
		q.BandwidthLimit = bandwidthLimit
		q.BandwidthSet = true
		rate := time.Second / time.Duration(bandwidthLimit)
		q.TokenBucket = NewTokenBucket(bandwidthLimit, rate)
	}
	if maxConcurrent != 0 { // 0 means unset
		q.MaxConcurrentDownloads = maxConcurrent
		q.MaxConcurrentSet = true
	}
	if !startTime.IsZero() {
		q.StartTime = startTime
		q.StartTimeSet = true
	}
	if !endTime.IsZero() {
		q.EndTime = endTime
		q.EndTimeSet = true
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
func (q *Queue) EditQueue(directory string, retriesLimit, maxConcurrent int, startTime, endTime time.Time, bandwidthLimit int) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.Directory = directory
	q.NumberOfTriesLimit = retriesLimit
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
	q.Downloads = append(q.Downloads, d)
	SaveQueuesToFile()

	//for _, queue := range QueuesList {
	//	if queue.Id == q.Id {
	//		d.Status = Pending
	//		q.Downloads = append(q.Downloads, d)
	//		_ = SaveQueuesToFile()
	//	}
	//}

	return nil
}

func (q *Queue) RemoveDownload(name string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for i, d := range q.Downloads {
		if d.FileName == name {
			q.Downloads = append(q.Downloads[:i], q.Downloads[i+1:]...)
			_ = SaveQueuesToFile()

			for j, globalDL := range DownloadsList {
				if globalDL.FileName == name {
					DownloadsList = append(
						DownloadsList[:j],
						DownloadsList[j+1:]...,
					)
					break
				}
			}
			return nil
		}
	}
	return errors.New("download not found")
}

func StartQueueDownloads(q *Queue) {
	q.mutex.Lock()
	q.HasStarted = true
	if q.Paused {
		fmt.Printf("Queue %s is paused. Downloads won't start\n", q.Id)
		q.mutex.Unlock()
		return
	}
	q.mutex.Unlock()

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

		go func(d *Download, dir string, tb *TokenBucket) {
			defer wg.Done()

			q.mutex.Lock()
			if q.Paused {
				fmt.Println("Queue is paused, stopping download:", d.FileName)
				q.mutex.Unlock()
				<-sem
				return
			}
			q.mutex.Unlock()

			d.NewDownloadManager(workers, tb)
			d.Manager.Ctx = ctx

			d.Directory = dir

			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Println("Error creating directory:", err)
				<-sem
				return
			}

			err := d.GetFileSizeAndName()
			if err != nil {
				fmt.Printf("Error getting file info for %s: %v\n", d.URL, err)
				<-sem
				return
			}

			fmt.Printf("Downloading file: %s\n", d.FileName)
			if err := d.StartDownload(); err != nil {
				fmt.Println("Error:", err)
			}

			<-sem
		}(d, q.Directory, q.TokenBucket)
	}

	wg.Wait()
	fmt.Println("All downloads completed in queue:", q.Id)
}

func (q *Queue) PauseQueue() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.Paused {
		fmt.Println("Queue is already paused")
		return
	}

	q.Paused = true
	if q.CancelFunc != nil {
		q.CancelFunc() // it will cancel all ongoing downloads
	}

	fmt.Printf("Queue %s paused successfully\n", q.Id)
}

func (q *Queue) ResumeQueue() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !q.Paused {
		fmt.Println("Queue is already running")
		return
	}

	q.Paused = false
	fmt.Printf("Queue %s resumed successfully\n", q.Id)
	go StartQueueDownloads(q)
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
		fmt.Printf("  - Start Time: %s\n", q.StartTime.Format("15:04:05"))
		fmt.Printf("  - End Time: %s\n", q.EndTime.Format("15:04:05"))
		fmt.Println("-------------------------------")
	}
}
func DeleteQueue(queueName string) []*Queue {
	newDownloads := make([]*Download, 0)
	for _, dl := range DownloadsList {
		if dl.QueueName != queueName {
			newDownloads = append(newDownloads, dl)
		}
	}
	DownloadsList = newDownloads

	delete(QueuesList, queueName)
	_ = SaveQueuesToFile()
	fmt.Println("Queue", queueName, "deleted successfully.")

	var updatedQueues []*Queue
	for _, q := range QueuesList {
		updatedQueues = append(updatedQueues, q)
	}
	return updatedQueues
}
