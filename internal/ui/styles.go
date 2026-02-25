package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/syvanpera/gwt/internal/ops"
)

type palette struct {
	panel      lipgloss.Style
	header     lipgloss.Style
	muted      lipgloss.Style
	running    lipgloss.Style
	success    lipgloss.Style
	failed     lipgloss.Style
	statusBase lipgloss.Style
}

func newPalette() palette {
	return palette{
		panel:  lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1),
		header: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75")),
		muted:  lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		running: lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).Bold(true),
		success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).Bold(true),
		failed: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).Bold(true),
		statusBase: lipgloss.NewStyle().Padding(0, 1).Bold(true),
	}
}

func (p palette) statusBadge(state ops.State) string {
	var s lipgloss.Style
	var label string
	switch state {
	case ops.StateRunning:
		s = p.statusBase.Background(lipgloss.Color("39")).Foreground(lipgloss.Color("0"))
		label = " RUNNING "
	case ops.StateSuccess:
		s = p.statusBase.Background(lipgloss.Color("42")).Foreground(lipgloss.Color("0"))
		label = " OK "
	case ops.StateFailed:
		s = p.statusBase.Background(lipgloss.Color("196")).Foreground(lipgloss.Color("15"))
		label = " FAILED "
	default:
		s = p.statusBase.Background(lipgloss.Color("245")).Foreground(lipgloss.Color("0"))
		label = " IDLE "
	}
	return s.Render(label)
}
