package database

import (
	"IDM/internal/download"
	"IDM/internal/queue"
)

type DataBase struct {
	QueuesList   []*queue.Queue
	DownloadList []*download.Download
}
