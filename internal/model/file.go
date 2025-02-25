package model

type File struct {
	Status       Status
	Url          string
	ProgressRate float32
	queue        Queue
}
