package database

import (
	"IDM/internal"
)

type DataBase struct {
	QueuesList   []*internal.Queue
	DownloadList []*internal.Download
}
