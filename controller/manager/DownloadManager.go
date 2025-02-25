package manager

import "IDM/model"

type DownloadManager struct {
}

func (d *DownloadManager) startDownload() {
	//TODO
}
func (d *DownloadManager) changeDownloadStatus(file *model.File) {
	//TODO
}
func (d *DownloadManager) deleteFromQueue(file *model.File, queue model.Queue) {
	//TODO
}
func (d *DownloadManager) retry(file *model.File) {
	//TODO
}
