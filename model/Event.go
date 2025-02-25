package model

import "time"

type Event struct {
	status    Status
	Timestamp time.Time // May useful for logging?
}

func NewEvent() *Event {
	// TODO
	return nil
}

func (e Event) ToJSON() string {
	// TODO
	return ""
}

/*
TODO:
I think we need to add some methods here, which create new events
(like download started, paused, ...)

*/
