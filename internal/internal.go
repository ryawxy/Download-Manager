package internal

import "fmt"

func Start() {
	err := LoadQueuesFromFile()
	if err != nil {
		fmt.Println("Error loading queues:", err)
	}
}

func GracefulShutdown() {
	err := SaveQueuesToFile()
	if err != nil {
		fmt.Println("Error saving queues:", err)
	}
}
