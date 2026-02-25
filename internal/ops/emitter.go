package ops

import "time"

type Emitter struct {
	ch    chan Event
	delay time.Duration
}

func NewEmitter(ch chan Event, delay time.Duration) *Emitter {
	return &Emitter{ch: ch, delay: delay}
}

func (e *Emitter) Started(label, message string) {
	e.ch <- Event{Type: EventStarted, OperationLabel: label, Message: message, Time: time.Now(), Indeterminate: true}
	e.pause()
}

func (e *Emitter) Progress(message string, progress float64) {
	e.ch <- Event{Type: EventProgress, Message: message, Progress: progress, Time: time.Now(), Indeterminate: false}
	e.pause()
}

func (e *Emitter) ProgressIndeterminate(message string) {
	e.ch <- Event{Type: EventProgress, Message: message, Time: time.Now(), Indeterminate: true}
	e.pause()
}

func (e *Emitter) Completed(message string) {
	e.ch <- Event{Type: EventCompleted, Message: message, Progress: 1.0, Time: time.Now(), Indeterminate: false}
	e.pause()
}

func (e *Emitter) Failed(err error) {
	e.ch <- Event{Type: EventFailed, Message: err.Error(), Err: err, Time: time.Now(), Indeterminate: false}
	e.pause()
}

func (e *Emitter) pause() {
	if e.delay > 0 {
		time.Sleep(e.delay)
	}
}
