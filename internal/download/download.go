package download

import "time"

type Status string

const (
	Paused     Status = "paused"
	Failed     Status = "failed"
	InProgress Status = "inProgress"
	Completed  Status = "completed"
)

type Download struct {
	Id         string //TODO:Random id generator?
	URL        string
	Filename   string
	Status     Status
	Progress   int64
	TotalBytes int64
	StartTime  time.Time
}
