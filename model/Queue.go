package model

type Queue struct {
	Files              []File
	Directory          string
	NumberOfFilesLimit int
	BandWidth          int
	NumberOfTriesLimit int
	TimeInterval       [2]int
}
