package internal

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
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
type ProgressMsg struct {
	Download *Download
}

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
	_, params, err := mime.ParseMediaType(contentDisp)
	if err != nil {
		return "", false
	}

	//fmt.Println("DEBUGGING PRINT !!! MediaType is: ", mediaType)

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
	SaveQueuesToFile()

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

	encoding := resp.Header.Get("Content-Encoding")
	if encoding == "gzip" {
		fmt.Println("Warning: Server is compressing the response, decoding needed!")
		// Decompress manually using gzip.Reader
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && start != 0 {
		fmt.Println("Warning: Server does not support partial content properly.")
	}

	partFileName := fmt.Sprintf("%s.part%d", download.FileName, partNum)
	fullPath := filepath.Join(download.Directory, partFileName)
	//file, err := os.Create(fullPath)
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("Error creating part file:", err)
		return
	}
	defer file.Close()

	buf := make([]byte, 1024)
	totalBytesRead := int64(0)

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

		// Control overlapping among workers
		remaining := end - start + 1 - totalBytesRead
		if int64(n) > remaining {
			n = int(remaining)
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
		percentage := float64(download.DownloadedBytes) / float64(download.FileSize) * 100
		download.Progress = int64(percentage)
		download.Manager.Mutex.Unlock()

		cmd := func() tea.Msg {
			return ProgressMsg{Download: download}
		}
		tea.Println(cmd)

		download.ShowProgress()
		download.changeDownloadStatus()
		SaveQueuesToFile()

		totalBytesRead += int64(n)
		if totalBytesRead >= (end - start + 1) {
			break
		}
	}
}
func (download *Download) StartDownload() error {

	if !download.CheckRangeSupport() {
		fmt.Println("Server does NOT support partial downloads. Switching to single-threaded mode...")
		download.Manager.Workers = 1
		download.Manager.ChunkSize = download.FileSize
	} else {
		fmt.Println("Server supports partial downloads. Using multi-threaded mode.")
	}

	if err := os.MkdirAll(download.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	err := download.GetFileSizeAndName()
	if err != nil {
		return err
	}

	chunkSize := download.FileSize / int64(download.Manager.Workers)
	remainingBytes := download.FileSize % int64(download.Manager.Workers)

	var wg sync.WaitGroup
	for i := 0; i < download.Manager.Workers; i++ {
		//start := int64(i) * download.Manager.ChunkSize
		start := int64(i) * chunkSize
		//end := start + download.Manager.ChunkSize - 1
		end := start + chunkSize - 1

		//if end >= download.FileSize {
		//	end = download.FileSize - 1
		//}
		if i == download.Manager.Workers-1 {
			end += remainingBytes
		}

		wg.Add(1)
		go download.downloadChunk(start, end, i, &wg)
	}

	wg.Wait()
	return mergeFiles(download)
}

func mergeFiles(download *Download) error {
	outputPath := filepath.Join(download.Directory, download.FileName)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create merged file: %v", err)
	}
	defer outputFile.Close()

	fmt.Println("merging downloaded parts...")

	for i := 0; i < download.Manager.Workers; i++ {
		partPath := filepath.Join(download.Directory, fmt.Sprintf("%s.part%d", download.FileName, i))

		if _, err := os.Stat(partPath); os.IsNotExist(err) {
			fmt.Printf("Warning: Part %d is missing, skipping...\n", i)
			continue
		}

		partFile, err := os.Open(partPath)
		if err != nil {
			return fmt.Errorf("Error opening part %d: %v\n", i, err)
		}

		_, err = io.Copy(outputFile, partFile)
		if err != nil {
			partFile.Close()
			fmt.Printf("Error copying part %d: %v\n", i, err)
			return err
		}
		partFile.Close()

		if err := os.Remove(partPath); err != nil {
			fmt.Printf("Warning: failed to remove part %d: %v\n", i, err)
		}
	}

	//fmt.Println("Download complete.")
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
	} else if download.Progress > 0 && download.Progress < download.FileSize {
		download.Status = InProgress
	} else if download.Progress >= download.FileSize {
		download.Status = Completed
	} else {
		download.Status = Failed
	}

	//	fmt.Printf("Download status updated: %s -> %s\n", download.FileName, download.Status)
}

func (dm *DownloadManager) deleteFromQueue(download *Download, queue *Queue) {
	//TODO
}

func (download *Download) Retry() {
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

func (download *Download) PauseDownload() {
	download.Manager.Mutex.Lock()
	defer download.Manager.Mutex.Unlock()
	download.Paused = true
	download.Status = Paused
	fmt.Printf("Paused download: %s\n", download.FileName)
}

// ResumeDownload resumes a paused download.
func (download *Download) ResumeDownload() {
	download.Manager.Mutex.Lock()
	defer download.Manager.Mutex.Unlock()
	download.Paused = false
	download.Status = InProgress
	// For simplicity, re-start the download. In a real app, resume logic would be more complex.
	go download.StartDownload()
	fmt.Printf("Resumed download: %s\n", download.FileName)
}

func (download *Download) CheckRangeSupport() bool {
	req, err := http.NewRequest("HEAD", download.URL, nil)
	if err != nil {
		fmt.Println("Error creating HEAD request:", err)
		return false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending HEAD request:", err)
		return false
	}
	defer resp.Body.Close()

	acceptRanges := resp.Header.Get("Accept-Ranges")
	return resp.StatusCode == http.StatusPartialContent && acceptRanges == "bytes"
}

func (download *Download) ShowProgress() {
	download.Manager.Mutex.Lock()
	defer download.Manager.Mutex.Unlock()

	download.Progress = int64(float64(download.DownloadedBytes) / float64(download.FileSize) * 100)
	//barLength := 30
	//filled := int(percentage / 100 * float64(barLength))
	//bar := strings.Repeat("█", filled) + strings.Repeat("-", barLength-filled)
	//
	//fmt.Printf("\rDownloading: [%s] %.2f%%", bar, percentage)
}
