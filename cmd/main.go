package main

import (
	"IDM/internal"
	"IDM/internal/tui"
	"fmt"
)

func main() {

}

func init() {
	err := internal.LoadQueuesFromFile()
	if err != nil {
		fmt.Println("Error loading queues:", err)
	}
	tui.Start()
}
