package manager

import (
	database2 "IDM/internal/database"
	"IDM/internal/download"
	"IDM/internal/queue"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type DownloadManager struct {
	URL       string `json:"url"`
	FileName  string `json:"fileName"`
	FileSize  int64  `json:"fileSize"`  // whole file
	ChunkSize int64  `json:"chunkSize"` // size that each worker process
	Workers   int
	Cancel    context.CancelFunc
	Ctx       context.Context
	Mutex     sync.Mutex
	Paused    bool `json:"paused"` // our goroutines must check this field...
}

var database database2.DataBase

/*
	NewDownloadManager is just a simple constructor, dont worry :)

better to implement at future i guess, we can create multiple DM
*/
func NewDownloadManager(url string, fileName string, workers int) *DownloadManager {
	ctx, cancel := context.WithCancel(context.Background())
	d := &DownloadManager{
		URL:      url,
		FileName: fileName,
		Workers:  workers,
		Cancel:   cancel,
		Ctx:      ctx,
		Paused:   false,
	}
	//err := SaveData(d)
	//if err != nil {
	//	return nil
	//}
	return d
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

func getFileNameFromURL(rawURL string) string {

	parsedURL, err := url.Parse(rawURL)
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
func (dm *DownloadManager) GetFileSizeAndName() error {
	headResp, err := http.Head(dm.URL)
	if err != nil {
		return err
	}
	defer headResp.Body.Close()

	if headResp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get file info: server returned %d - %s", headResp.StatusCode, headResp.Status)
	}

	dm.FileSize = headResp.ContentLength // converting response to int!
	dm.ChunkSize = dm.FileSize / int64(dm.Workers)

	if filename, ok := getFileNameFromHeader(headResp); ok {
		dm.FileName = filename
	} else {
		dm.FileName = getFileNameFromURL(dm.URL)
	}

	return nil
}

func (dm *DownloadManager) downloadChunk(start int64, end int64, partNum int, wg *sync.WaitGroup) {
	defer wg.Done()

	req, err := http.NewRequestWithContext(dm.Ctx, "GET", dm.URL, nil)
	if err != nil {
		fmt.Println("Error while creating request:", err)
		return
	}

	// setting the process range for each goroutine
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error during download:", err)
		return
	}
	// response body must be closed at the end
	defer resp.Body.Close()

	file, err := os.Create(fmt.Sprintf("%s.part%d", dm.FileName, partNum))
	if err != nil {
		fmt.Println("Error creating file part:", err)
		return
	}
	// dont forget to close files at the end!
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	buf := make([]byte, 1024)

	for {
		// When we pause download, we should stop goroutines sequentially by lock/unlock
		dm.Mutex.Lock()
		for dm.Paused {
			dm.Mutex.Unlock()
			//fmt.Printf("Goroutine %d paused...\n", partNum)
			select {
			case <-dm.Ctx.Done():
				fmt.Println("Download cancelled.")
				return
			default:
				time.Sleep(500 * time.Millisecond)
			}
			dm.Mutex.Lock()
		}
		dm.Mutex.Unlock()

		/* try to get at most len(buf) bytes data from server
		n is the actual number of bytes read (could be lower than len(buf))
		*/
		n, err := resp.Body.Read(buf)
		if err != nil {
			// EOF stands for End Of File (complete)
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading data:", err)
			return
		}

		// because n could be lower than len(buf), we just write the first n bytes of the buf slice
		_, err = file.Write(buf[:n])
		if err != nil {
			fmt.Println("Error writing file part:", err)
			return
		}
	}

	fmt.Printf("Downloaded [%d] [%d] bytes by goroutine %d\n", start, end, partNum)
}

func (dm *DownloadManager) StartDownload(download download.Download) error {
	fmt.Println("Download started...")

	// Just for error handling at first, and filling dm.FileSize at the end
	err := dm.GetFileSizeAndName()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	// اینجا میخوایم بگیم هر وورکر از کجا تا کجا بخونه، انگلیسی نتانم ۴ صبح :))
	for i := 0; i < dm.Workers; i++ {
		start := int64(i) * dm.ChunkSize
		end := start + dm.ChunkSize - 1
		// handling last part
		if end >= dm.FileSize {
			end = dm.FileSize - 1
		}

		wg.Add(1)
		go dm.downloadChunk(start, end, i, &wg)
	}

	// باید تردا وایسن تا کار همشون تموم شه، بعد ریترن کنیم
	wg.Wait()

	return dm.mergeFiles()
}

func (dm *DownloadManager) mergeFiles() error {
	fmt.Println("Merging downloaded chunks...")

	outputFile, err := os.Create(dm.FileName)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for i := 0; i < dm.Workers; i++ {
		partFileName := fmt.Sprintf("%s.part%d", dm.FileName, i)
		partFile, err := os.Open(partFileName)
		if err != nil {
			return err
		}

		_, err = io.Copy(outputFile, partFile)
		if err != nil {
			return err
		}

		// this will help us to debug pause/resume features! (uncomment at the end)
		//os.Remove(partFileName)
	}

	fmt.Println("Download complete.")
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
