package manager

import (
	"IDM/internal/download"
	"IDM/internal/queue"
)

type DownloadManager struct {
}

func (d *DownloadManager) startDownload(download download.Download) {
	//TODO
}
func (d *DownloadManager) changeDownloadStatus(download download.Download) {
	//TODO
}
func (d *DownloadManager) deleteFromQueue(download download.Download, queue *queue.Queue) {
	//TODO
}
func (d *DownloadManager) retry(download download.Download) {
	//TODO
}
