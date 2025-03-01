package internal

import (
	database2 "IDM/internal/database"
	"IDM/internal/download"
	"encoding/json"
	"os"
	"sync"
)

type Storage struct {
}

var database = database2.DataBase{}
var mu sync.Mutex

func SaveData(download download.Download) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := json.MarshalIndent(download, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile("download_list.json", data, 0644)
}
