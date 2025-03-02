package manager

import (
	"encoding/json"
	"os"
	"sync"
)

type Storage struct {
}

var mu sync.Mutex

func SaveData(d *DownloadManager) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(&d, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile("download_list.json", data, 0644)
}
