package internal

import "time"

type Status string

const (
	Paused     Status = "paused"
	Failed     Status = "failed"
	InProgress Status = "inProgress"
	Completed  Status = "completed"
)

type Download struct {
	URL        string `json:"URL"`
	FileName   string `json:"FileName"`
	FileSize   int64  `json:"FileSize"`
	Status     Status `json:"status"`
	Progress   int64  `json:"progress"`
	TotalBytes int64  `json:"total_bytes"`
	StartTime  time.Time
}
