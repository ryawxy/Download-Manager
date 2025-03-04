package queue

import (
	"IDM/internal/download"
	"errors"
	"sync"
	"time"
)

type Queue struct {
	Id                 string //TODO:random String generator
	Downloads          []*download.Download
	Directory          string
	NumberOfFilesLimit int
	BandwidthLimit     int
	NumberOfTriesLimit int
	StartTime          time.Time
	EndTime            time.Time

	mutex sync.Mutex
}

func NewQueue(id, directory string, numberOfFilesLimit, bandwidthLimit int, startTime, endTime time.Time) *Queue {
	return &Queue{
		Id:                 id,
		Downloads:          make([]*download.Download, 0),
		NumberOfFilesLimit: numberOfFilesLimit,
		Directory:          directory,
		BandwidthLimit:     bandwidthLimit,
		StartTime:          startTime,
		EndTime:            endTime,
	}
}
func (q *Queue) AddDownload(d *download.Download) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Downloads = append(q.Downloads, d)
	return nil
}
func (q *Queue) RemoveDownload(name string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for i, d := range q.Downloads {
		if d.FileName == name {
			q.Downloads = append(q.Downloads[:i], q.Downloads[i+1:]...)
			return nil
		}
	}
	return errors.New("download not found")
}
func (q *Queue) UpdateSettings(numberOfFilesLimit int, directory string, bandwidth int, startTime, endTime time.Time) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.NumberOfTriesLimit = numberOfFilesLimit
	q.Directory = directory
	q.BandwidthLimit = bandwidth
	q.StartTime = startTime
	q.EndTime = endTime

	return nil
}
