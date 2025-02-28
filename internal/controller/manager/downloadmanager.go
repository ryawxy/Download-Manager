package manager

import (
	"IDM/internal/download"
	"IDM/internal/queue"
	"context"
	"fmt"
	"net/http"
	"sync"
)

type DownloadManager struct {
	URL       string
	FileName  string
	FileSize  int64 // whole file
	ChunkSize int64 // size that each worker process
	Workers   int
	Cancel    context.CancelFunc
	Ctx       context.Context
	Mutex     sync.Mutex
}

/*
http.Head() sends a HEAD request to server,
and returns headResponse only (not file content)
*/
func (dm *DownloadManager) getFileSize() error {
	headResp, err := http.Head(dm.URL)
	if err != nil {
		return err
	}
	if headResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get file info: server returned %d - %s", headResp.StatusCode, headResp.Status)
	}

	dm.FileSize = headResp.ContentLength // converting response to int!
	return nil
}

func (dm *DownloadManager) startDownload(download download.Download) error {
	fmt.Println("Download started...")

	// Just for error handling at first, and filling dm.FileSize at the end
	err := dm.getFileSize()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	chunkSize := dm.FileSize / int64(dm.Workers)

	// اینجا میخوایم بگیم هر وورکر از کجا تا کجا بخونه، انگلیسی نتانم ۴ صبح :))
	for i := 0; i < dm.Workers; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1

		// handling last part
		if end >= dm.FileSize {
			end = dm.FileSize - 1
		}

		wg.Add(1)
		/* TODO: need to implement a better download function here, followed by go syntax
		smth like "go downloadChunk(start, end, ...)"
		*/
	}

	// باید تردا وایسن تا کار همشون تموم شه، بعد ریترن کنیم
	wg.Wait()

	// TODO: need to implement another method for merging the chunks that workers proceed above,
	return nil
}

func (dm *DownloadManager) CancelDownload() {
	dm.Cancel()
	fmt.Println("Download cancelled.")
}

func (dm *DownloadManager) changeDownloadStatus(download download.Download) {
	//TODO
}
func (dm *DownloadManager) deleteFromQueue(download download.Download, queue *queue.Queue) {
	//TODO
}
func (dm *DownloadManager) retry(download download.Download) {
	//TODO
}
