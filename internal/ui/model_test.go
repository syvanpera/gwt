package ui

import (
	"errors"
	"testing"
	"time"

	"github.com/syvanpera/gwt/internal/ops"
)

func TestApplyEventTransitions(t *testing.T) {
	m := model{status: ops.Status{State: ops.StateRunning}, started: time.Now(), palette: newPalette()}

	m.applyEvent(ops.Event{Type: ops.EventProgress, Message: "step", Progress: 0.4, Time: time.Now()})
	if m.status.ProgressPercent != 0.4 || m.status.State != ops.StateRunning {
		t.Fatalf("unexpected running status: %+v", m.status)
	}

	m.applyEvent(ops.Event{Type: ops.EventCompleted, Message: "done", Time: time.Now()})
	if m.status.State != ops.StateSuccess {
		t.Fatalf("expected success, got %s", m.status.State)
	}

	err := errors.New("boom")
	m.applyEvent(ops.Event{Type: ops.EventFailed, Message: err.Error(), Err: err, Time: time.Now()})
	if m.status.State != ops.StateFailed {
		t.Fatalf("expected failed, got %s", m.status.State)
	}
	if m.err == nil {
		t.Fatalf("expected model error to be set")
	}
}
