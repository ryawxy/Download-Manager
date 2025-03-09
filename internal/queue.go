package internal

import (
	"errors"
	"sync"
	"time"
)

type Queue struct {
	Id                 string //TODO:random String generator
	Downloads          []*Download
	Directory          string
	NumberOfFilesLimit int
	BandwidthLimit     int64
	NumberOfTriesLimit int
	StartTime          time.Time
	EndTime            time.Time
	TokenBucket        *TokenBucket

	mutex sync.Mutex
}

func NewQueue(id, directory string, numberOfFilesLimit int, bandwidthLimit int64, startTime, endTime time.Time) *Queue {
	rate := time.Second / time.Duration(bandwidthLimit)

	return &Queue{
		Id:                 id,
		Downloads:          make([]*Download, 0),
		NumberOfFilesLimit: numberOfFilesLimit,
		Directory:          directory,
		BandwidthLimit:     bandwidthLimit,
		StartTime:          startTime,
		EndTime:            endTime,
		TokenBucket:        NewTokenBucket(bandwidthLimit, rate),
	}
}
func (q *Queue) AddDownload(d *Download) error {
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
func (q *Queue) UpdateSettings(numberOfFilesLimit int, directory string, bandwidth int64, startTime, endTime time.Time) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.NumberOfTriesLimit = numberOfFilesLimit
	q.Directory = directory
	q.BandwidthLimit = bandwidth
	q.StartTime = startTime
	q.EndTime = endTime

	return nil
}
