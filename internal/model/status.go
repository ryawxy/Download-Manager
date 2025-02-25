package model

type Status int

const (
	paused Status = iota
	failed
	inProgress
	completed
)
