package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const saveFilePath = "queues.json"

var storageMutex sync.Mutex

func SaveQueuesToFile() error {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	data, err := json.MarshalIndent(QueuesList, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(saveFilePath, data, 0644)
	if err != nil {
		return err
	}

	//fmt.Println("Queues saved successfully.")
	return nil
}
func LoadQueuesFromFile() error {
	q := &Queue{
		Id:             "Default",
		BandwidthLimit: 1000000, // 1,000,000 bytes/s ~ 1,000 KB/s ~ 1 MB/s
	}
	if q.BandwidthLimit != 0 {
		rate := time.Second / time.Duration(q.BandwidthLimit)
		q.TokenBucket = NewTokenBucket(q.BandwidthLimit, rate)
	}
	QueuesList["Default"] = q

	storageMutex.Lock()
	defer storageMutex.Unlock()

	data, err := os.ReadFile(saveFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No existing queue file found. Starting fresh.")

		}
	}

	err = json.Unmarshal(data, &QueuesList)
	if err != nil {
		return err
	}

	for _, q := range QueuesList {
		q.StartTime = q.StartTime.UTC()
		q.EndTime = q.EndTime.UTC()
		if q.BandwidthLimit != 0 {
			rate := time.Second / time.Duration(q.BandwidthLimit)
			q.TokenBucket = NewTokenBucket(q.BandwidthLimit, rate)
		} else {
			q.TokenBucket = nil
		}

		for _, d := range q.Downloads {
			DownloadsList = append(DownloadsList, d)
		}
	}

	ResumeInProgressDownloads()
	//fmt.Println("Queues loaded successfully.")
	return nil
}

func ResumeInProgressDownloads() {
	for _, download := range DownloadsList {
		if download.Status == InProgress {
			fmt.Printf("Resuming in-progress download: %s\n", download.FileName)

			queue := QueuesList[download.QueueName]
			if queue == nil {
				queue = QueuesList["Default"]
			}

			if download.Manager == nil {
				download.NewDownloadManager(WORKERS, QueuesList[download.QueueName].TokenBucket)
			}

			// reset progress if restart from scratch
			// download.DownloadedBytes = 0
			// download.Progress = 0

			// clean up partial files from previous run
			for i := 0; i < download.Manager.Workers; i++ {
				partPath := filepath.Join(download.Directory, fmt.Sprintf("%s.part%d", download.FileName, i))
				os.Remove(partPath)
			}

			go func(d *Download) {
				err := d.StartDownload()
				if err != nil {
					fmt.Printf("Failed to resume download %s: %v\n", d.FileName, err)
					d.Status = Failed
				}
				SaveQueuesToFile()
			}(download)
		}
	}
}

func (q *Queue) MarshalJSON() ([]byte, error) {
	type Alias Queue
	return json.Marshal(&struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		*Alias
	}{
		StartTime: q.StartTime.Format(time.RFC3339),
		EndTime:   q.EndTime.Format(time.RFC3339),
		Alias:     (*Alias)(q),
	})
}

func (q *Queue) UnmarshalJSON(data []byte) error {
	type Alias Queue
	aux := &struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		*Alias
	}{
		Alias: (*Alias)(q),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	startTime, err := time.Parse(time.RFC3339, aux.StartTime)
	if err != nil {
		return err
	}
	q.StartTime = startTime

	endTime, err := time.Parse(time.RFC3339, aux.EndTime)
	if err != nil {
		return err
	}
	q.EndTime = endTime

	if q.BandwidthLimit != 0 {
		q.TokenBucket = NewTokenBucket(q.BandwidthLimit, time.Second/time.Duration(q.BandwidthLimit))
	} else {
		q.TokenBucket = nil
	}

	return nil
}
