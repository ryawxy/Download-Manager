package internal

import "time"

type Status string

const (
	Paused     Status = "paused"
	Failed     Status = "failed"
	InProgress Status = "inProgress"
	Completed  Status = "completed"
	Pending    Status = "pending"
	Cancelled  Status = "cancelled"
)

var DownloadsList []*Download

type Download struct {
	URL                string `json:"URL"`
	FileName           string `json:"FileName"`
	FileSize           int64  `json:"FileSize"`
	Status             Status `json:"status"`
	Progress           int64  `json:"progress"`
	StartTime          time.Time
	Manager            *DownloadManager `json:"-"`
	Paused             bool
	DownloadedBytes    int64  `json:"downloaded_bytes"`
	Directory          string `json:"directory"`
	QueueName          string `json:"queue_name"`
	LastUpdateTime     time.Time
	LastBytes          int64
	Speed              float64
	lastSpeedCalcTime  time.Time
	lastSpeedCalcBytes int64
	RetryCount         int    `json:"retry_count"`
	SelectedName       string `json:"selected_name"`
}
