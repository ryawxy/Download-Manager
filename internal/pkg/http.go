package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// TODO: user should input URL and fileName in the terminal (cli)
	url := "https://ensani.ir/file/download/article/20160216093124-10017-18.pdf"
	fileName := "download_test.pdf"

	// TODO: user should select desired path itself.
	err := downloadFile(url, fileName)
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println("Download success.")
	}

}

func downloadFile(url string, fileName string) error {

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	// closing file at the end
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

// TODO: need to implement retry, pause, cancel and resume functions
