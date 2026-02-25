package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/syvanpera/gwt/internal/ops"
)

type opFunc func(context.Context, *ops.Emitter) error

type eventMsg ops.Event

type doneMsg struct{}

var eventDelay time.Duration

func SetEventDelay(delay time.Duration) {
	eventDelay = delay
}

type model struct {
	title       string
	description string
	palette     palette
	spinner     spinner.Model
	status      ops.Status
	history     []historyEntry
	errorDetail string
	width       int
	height      int
	events      chan ops.Event
	started     time.Time
	err         error
}

type historyEntry struct {
	message string
	state   ops.State
}

type displayedError struct {
	err error
}

func (e displayedError) Error() string {
	return e.err.Error()
}

func (e displayedError) Unwrap() error {
	return e.err
}

func IsDisplayedError(err error) bool {
	var de displayedError
	return errors.As(err, &de)
}

func RunOperation(title, description string, fn opFunc) error {
	if !isInteractiveTerminal() {
		events := make(chan ops.Event, 128)
		em := ops.NewEmitter(events, eventDelay)
		go func() {
			for range events {
			}
		}()
		err := fn(context.Background(), em)
		close(events)
		return err
	}

	events := make(chan ops.Event, 32)
	em := ops.NewEmitter(events, eventDelay)
	ctx := context.Background()

	m := model{
		title:       title,
		description: description,
		palette:     newPalette(),
		spinner:     spinner.New(spinner.WithSpinner(spinner.Dot)),
		status: ops.Status{
			Label:           title,
			State:           ops.StateRunning,
			IsIndeterminate: true,
			StartedAt:       time.Now(),
		},
		events:  events,
		started: time.Now(),
	}

	go func() {
		err := fn(ctx, em)
		if err != nil {
			em.Failed(err)
		}
		close(events)
	}()

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return err
	}
	fm := result.(model)
	if fm.err != nil {
		return displayedError{err: fm.err}
	}
	return fm.err
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, waitForEvent(m.events))
}

func waitForEvent(ch <-chan ops.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return doneMsg{}
		}
		return eventMsg(ev)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.err = context.Canceled
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case eventMsg:
		ev := ops.Event(msg)
		m.applyEvent(ev)
		return m, tea.Batch(waitForEvent(m.events), m.spinner.Tick)
	case doneMsg:
		if m.status.State == ops.StateRunning {
			m.status.State = ops.StateSuccess
			m.status.EndedAt = time.Now()
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) applyEvent(ev ops.Event) {
	detail := ev.Message
	if strings.TrimSpace(detail) == "" {
		detail = string(ev.Type)
	}

	switch ev.Type {
	case ops.EventStarted:
		m.finalizeLastRunning(ops.StateSuccess)
		m.appendHistory(detail, ops.StateRunning)
		m.errorDetail = ""
		m.status.State = ops.StateRunning
		m.status.Message = ev.Message
		m.status.IsIndeterminate = true
		m.status.StartedAt = ev.Time
	case ops.EventProgress:
		lastMsg := m.lastHistoryMessage()
		if lastMsg != "" && lastMsg != detail {
			m.finalizeLastRunning(ops.StateSuccess)
		}
		m.appendHistory(detail, ops.StateRunning)
		m.errorDetail = ""
		m.status.State = ops.StateRunning
		m.status.Message = ev.Message
		m.status.IsIndeterminate = ev.Indeterminate
		if !ev.Indeterminate {
			m.status.ProgressPercent = clamp(ev.Progress)
		}
	case ops.EventCompleted:
		m.finalizeLastRunning(ops.StateSuccess)
		m.appendHistory(detail, ops.StateSuccess)
		m.errorDetail = ""
		m.status.State = ops.StateSuccess
		m.status.Message = ev.Message
		m.status.IsIndeterminate = false
		m.status.ProgressPercent = 1
		m.status.EndedAt = ev.Time
	case ops.EventFailed:
		m.finalizeLastRunning(ops.StateFailed)
		m.status.State = ops.StateFailed
		m.status.Message = "Operation failed"
		m.errorDetail = ev.Message
		m.status.IsIndeterminate = false
		m.status.EndedAt = ev.Time
		m.err = ev.Err
	}
}

func (m model) View() string {
	statusLine := lipgloss.JoinHorizontal(lipgloss.Left, m.palette.statusBadge(m.status.State), "  ", m.status.Message)

	indicator := m.spinner.View() + "  " + m.palette.running.Render("Working...")
	if m.status.State == ops.StateSuccess {
		indicator = m.palette.success.Render("✓ Done")
	}
	if m.status.State == ops.StateFailed {
		indicator = m.palette.failed.Render("✗ Failed")
	}

	elapsed := time.Since(m.started).Round(time.Second)
	elapsedLine := m.palette.muted.Render("Elapsed: " + elapsed.String())
	errorLine := ""
	if m.status.State == ops.StateFailed && strings.TrimSpace(m.errorDetail) != "" {
		errorLine = "\n" + m.palette.failed.Render("Error: "+m.errorDetail)
	}

	history := ""
	innerWidth := max(20, min(70, m.width-12))
	for _, h := range m.history {
		history += "\n" + m.renderHistoryLine(h, innerWidth)
	}
	panel := m.palette.panel.Width(max(40, min(90, m.width-4))).Render(statusLine + errorLine + "\n\n" + indicator + "\n" + elapsedLine + "\n" + history)
	return lipgloss.JoinVertical(lipgloss.Left, panel, "", m.palette.muted.Render("Press q to quit"))
}

func (m *model) appendHistory(message string, state ops.State) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	if len(m.history) > 0 {
		last := m.history[len(m.history)-1]
		if last.message == message && last.state == state {
			return
		}
	}
	m.history = append(m.history, historyEntry{message: message, state: state})
	if len(m.history) > 8 {
		m.history = m.history[len(m.history)-8:]
	}
}

func (m *model) finalizeLastRunning(finalState ops.State) {
	if len(m.history) == 0 {
		return
	}
	last := m.history[len(m.history)-1]
	if last.state != ops.StateRunning {
		return
	}
	last.state = finalState
	m.history[len(m.history)-1] = last
}

func (m model) lastHistoryMessage() string {
	if len(m.history) == 0 {
		return ""
	}
	return m.history[len(m.history)-1].message
}

func (m model) renderHistoryLine(entry historyEntry, width int) string {
	icon := "•"
	stateLabel := "running"
	stateStyle := m.palette.running
	switch entry.state {
	case ops.StateSuccess:
		icon = "✓"
		stateLabel = "done"
		stateStyle = m.palette.success
	case ops.StateFailed:
		icon = "✗"
		stateLabel = "failed"
		stateStyle = m.palette.failed
	}
	left := fmt.Sprintf("%s %s", icon, entry.message)
	total := len(left) + len(stateLabel) + 2
	if total < width {
		left += strings.Repeat(".", width-total)
	}
	return stateStyle.Render(left + "  " + stateLabel)
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isInteractiveTerminal() bool {
	in, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	if (in.Mode()&os.ModeCharDevice) == 0 || (fi.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return false
	}
	_ = tty.Close()
	return true
}
