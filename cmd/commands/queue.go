package commands

import (
	"IDM/internal"
	database2 "IDM/internal/database"
	"fmt"
	"time"
)

var database database2.DataBase

func CreateQueue(id, folder string, maxConcurrent, bandwidth int) *internal.Queue {
	startTime := time.Now()
	endTime := startTime.Add(24 * time.Hour)

	newQueue := internal.NewQueue(id, folder, maxConcurrent, bandwidth, startTime, endTime)
	return newQueue
}
func ShowQueues() []*internal.Queue {

	return database.QueuesList
}

func DeleteQueue(queues []*internal.Queue, id string) ([]*internal.Queue, error) {
	for i, q := range queues {
		if q.Id == id {
			queues = append(queues[:i], queues[i+1:]...)
			return queues, nil
		}
	}
	return queues, fmt.Errorf("queue with ID '%s' not found", id)
}

func ChangeQueueSettings(id string, numberOfFilesLimit int, directory string, bandwidth int, startTime, endTime time.Time) {
	queues := database.QueuesList

	var queueToUpdate *internal.Queue
	for _, q := range queues {
		if q.Id == id {
			queueToUpdate = q
			break
		}
	}

	if queueToUpdate == nil {
		return
	}

	err := queueToUpdate.UpdateSettings(numberOfFilesLimit, directory, bandwidth, startTime, endTime)
	if err != nil {
		return
	}

}
