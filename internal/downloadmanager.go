package internal

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type DownloadManager struct {
	ChunkSize   int64 `json:"chunkSize"` // size that each worker process
	Workers     int
	Cancel      context.CancelFunc
	Ctx         context.Context
	Mutex       sync.Mutex
	TokenBucket *TokenBucket
}

/*
	NewDownloadManager is just a simple constructor, don't worry :)

better to implement at future I guess, we can create multiple DM
*/
func (download *Download) NewDownloadManager(workers int, tb *TokenBucket) *DownloadManager {
	ctx, cancel := context.WithCancel(context.Background())
	download.Manager = &DownloadManager{
		Workers:     workers,
		Cancel:      cancel,
		Ctx:         ctx,
		TokenBucket: tb,
	}
	return download.Manager
}

func getFileNameFromHeader(resp *http.Response) (string, bool) {
	contentDisp := resp.Header.Get("Content-Disposition")
	if contentDisp == "" {
		return "", false // it means server didn't send any contentDisp
	}

	// it returns the media type automatically!
	mediaType, params, err := mime.ParseMediaType(contentDisp)
	if err != nil {
		return "", false
	}

	fmt.Println("DEBUGGING PRINT !!! MediaType is: ", mediaType)

	filename, ok := params["filename"]
	return filename, ok
}

func (download *Download) getFileNameFromURL() string {

	parsedURL, err := url.Parse(download.URL)
	if err != nil {
		fmt.Println("Invalid URL:", err)
		return "downloaded_file" // we're selecting a default name here // TODO random or smth else?
	}

	segments := strings.Split(parsedURL.Path, "/")
	filename := segments[len(segments)-1]

	if filename == "" || strings.Contains(filename, ".") == false {
		return "downloaded_file" // same as above
	}

	return filename
}

/*
http.Head() sends a HEAD request to server,
and returns headResponse only (not file content)
*/
func (download *Download) GetFileSizeAndName() error {
	headResp, err := http.Head(download.URL)
	if err != nil {
		return err
	}
	defer headResp.Body.Close()

	if headResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get file info: server returned %d - %s", headResp.StatusCode, headResp.Status)
	}

	download.FileSize = headResp.ContentLength // converting response to int!
	download.Manager.ChunkSize = download.FileSize / int64(download.Manager.Workers)

	if filename, ok := getFileNameFromHeader(headResp); ok {
		download.FileName = filename
	} else {
		download.FileName = download.getFileNameFromURL()
	}

	return nil
}

func (download *Download) downloadChunk(start int64, end int64, partNum int, wg *sync.WaitGroup) {
	defer wg.Done()

	req, err := http.NewRequestWithContext(download.Manager.Ctx, "GET", download.URL, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error during download:", err)
		return
	}
	defer resp.Body.Close()

	partFileName := fmt.Sprintf("%s.part%d", download.FileName, partNum)
	fullPath := filepath.Join(download.Directory, partFileName)
	file, err := os.Create(fullPath)
	if err != nil {
		fmt.Println("Error creating part file:", err)
		return
	}
	defer file.Close()

	buf := make([]byte, 1024)
	for {
		download.Manager.Mutex.Lock()
		for download.Paused {
			download.Manager.Mutex.Unlock()
			select {
			case <-download.Manager.Ctx.Done():
				fmt.Println("Download cancelled")
				return
			default:
				time.Sleep(500 * time.Millisecond)
			}
			download.Manager.Mutex.Lock()
		}
		download.Manager.Mutex.Unlock()

		n, err := resp.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading data:", err)
			return
		}

		// Apply rate limiting
		if download.Manager.TokenBucket != nil {
			download.Manager.TokenBucket.WaitAndTake(n)
		}

		_, err = file.Write(buf[:n])
		if err != nil {
			fmt.Println("Error writing to part file:", err)
			return
		}

		download.Manager.Mutex.Lock()
		download.DownloadedBytes += int64(n)
		download.Manager.Mutex.Unlock()

		download.ShowProgress()
	}
}
func (download *Download) StartDownload() error {
	if err := os.MkdirAll(download.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	err := download.GetFileSizeAndName()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	fmt.Println(download.Manager.Workers, "*****************************")
	for i := 0; i < download.Manager.Workers; i++ {
		start := int64(i) * download.Manager.ChunkSize
		end := start + download.Manager.ChunkSize - 1
		if end >= download.FileSize {
			end = download.FileSize - 1
		}

		wg.Add(1)
		go download.downloadChunk(start, end, i, &wg)
		fmt.Println(i)
	}

	wg.Wait()
	return mergeFiles(download)
}

func mergeFiles(download *Download) error {
	outputPath := filepath.Join(download.Directory, download.FileName)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for i := 0; i < download.Manager.Workers; i++ {
		partPath := filepath.Join(download.Directory, fmt.Sprintf("%s.part%d", download.FileName, i))
		partFile, err := os.Open(partPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(outputFile, partFile)
		if err != nil {
			return err
		}
		partFile.Close()
		os.Remove(partPath)
	}

	fmt.Println("Download complete.")
	return nil
}
func (download *Download) CancelDownload() {
	download.Manager.Cancel()
	fmt.Println("Download cancelled.")
}

func (download *Download) changeDownloadStatus() {
	download.Manager.Mutex.Lock()
	defer download.Manager.Mutex.Unlock()

	if download.Progress == 0 {
		download.Status = Pending
	} else if download.Progress > 0 && download.Progress < download.TotalBytes {
		download.Status = InProgress
	} else if download.Progress >= download.TotalBytes {
		download.Status = Completed
	} else {
		download.Status = Failed
	}

	fmt.Printf("Download status updated: %s -> %s\n", download.FileName, download.Status)
}

func (dm *DownloadManager) deleteFromQueue(download *Download, queue *Queue) {
	//TODO
}

func (download *Download) retry() {
	if download.Status != Failed {
		fmt.Println("Retry not allowed, download isn't in failed status")
		return
	}

	fmt.Printf("Retrying download: %s\n", download.FileName)
	download.Status = InProgress

	err := download.StartDownload()
	if err != nil {
		fmt.Println("Retry failed:", err)
		download.Status = Failed
		return
	}

	download.Status = Completed
	fmt.Println("Retry successful:", download.FileName)
}

func (download *Download) ShowProgress() {
	download.Manager.Mutex.Lock()
	defer download.Manager.Mutex.Unlock()

	percentage := float64(download.DownloadedBytes) / float64(download.FileSize) * 100
	barLength := 30
	filled := int(percentage / 100 * float64(barLength))
	bar := strings.Repeat("█", filled) + strings.Repeat("-", barLength-filled)

	fmt.Printf("\rDownloading: [%s] %.2f%%", bar, percentage)
}
