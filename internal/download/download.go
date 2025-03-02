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
	URL        string `json:"URL"`
	Filename   string `json:"Filename"`
	Status     Status `json:"status"`
	Progress   int64  `json:"progress"`
	TotalBytes int64  `json:"total_bytes"`
	StartTime  time.Time
}
