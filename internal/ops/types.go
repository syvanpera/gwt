package ops

import "time"

type State string

const (
	StateRunning State = "running"
	StateSuccess State = "success"
	StateFailed  State = "failed"
)

type Status struct {
	ID              string
	Label           string
	State           State
	Message         string
	ProgressPercent float64
	IsIndeterminate bool
	StartedAt       time.Time
	EndedAt         time.Time
}

type EventType string

const (
	EventStarted   EventType = "started"
	EventProgress  EventType = "progress"
	EventCompleted EventType = "completed"
	EventFailed    EventType = "failed"
)

type Event struct {
	Type           EventType
	Message        string
	Progress       float64
	Indeterminate  bool
	Time           time.Time
	Err            error
	OperationLabel string
}
