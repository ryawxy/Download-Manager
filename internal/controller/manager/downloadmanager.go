package manager

import (
	model2 "IDM/internal/model"
	"IDM/internal/queue"
)

type DownloadManager struct {
}

func (d *DownloadManager) startDownload() {
	//TODO
}
func (d *DownloadManager) changeDownloadStatus(file *model2.File) {
	//TODO
}
func (d *DownloadManager) deleteFromQueue(file *model2.File, queue queue.Queue) {
	//TODO
}
func (d *DownloadManager) retry(file *model2.File) {
	//TODO
}
