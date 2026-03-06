package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorktreeOption struct {
	Branch string
	Path   string
}

type worktreeItem struct {
	branch string
	path   string
}

func (i worktreeItem) Title() string       { return i.branch }
func (i worktreeItem) Description() string { return i.path }
func (i worktreeItem) FilterValue() string { return i.branch + " " + i.path }

type pickerModel struct {
	list      list.Model
	itemCount int
	selected  string
	cancelled bool
}

type compactDelegate struct {
	normal   lipgloss.Style
	selected lipgloss.Style
}

func (d compactDelegate) Height() int  { return 1 }
func (d compactDelegate) Spacing() int { return 0 }
func (d compactDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}
func (d compactDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	wt, ok := item.(worktreeItem)
	if !ok {
		return
	}
	isSelected := index == m.Index()
	prefix := "  "
	style := d.normal
	if isSelected {
		prefix = "› "
		style = d.selected
	}
	_, _ = fmt.Fprint(w, style.Render(prefix+wt.branch))
}

func PickWorktree(options []WorktreeOption) (string, error) {
	if len(options) == 0 {
		return "", errors.New("no worktrees available")
	}
	items := make([]list.Item, 0, len(options))
	for _, opt := range options {
		items = append(items, worktreeItem{branch: opt.Branch, path: opt.Path})
	}

	delegate := compactDelegate{
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true),
	}

	initialHeight := max(1, len(items))
	l := list.New(items, delegate, 30, initialHeight)
	l.Title = ""
	l.Styles.Title = lipgloss.NewStyle()
	l.SetShowStatusBar(false)
	l.SetShowTitle(false)
	l.SetShowPagination(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	m := pickerModel{
		list:      l,
		itemCount: len(items),
	}
	p := tea.NewProgram(m, tea.WithOutput(os.Stderr))
	res, err := p.Run()
	if err != nil {
		return "", err
	}
	pm := res.(pickerModel)
	if pm.cancelled {
		return "", context.Canceled
	}
	if pm.selected == "" {
		return "", errors.New("selection cancelled")
	}
	return pm.selected, nil
}

func (m pickerModel) Init() tea.Cmd {
	return nil
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		width := max(18, min(44, msg.Width-2))
		maxHeight := max(1, msg.Height-2)
		height := min(maxHeight, max(1, m.itemCount))
		m.list.SetSize(width, height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(worktreeItem); ok {
				m.selected = item.path
			}
			return m, tea.Quit
		case "esc", "q", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m pickerModel) View() string {
	prompt := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Choose worktree")
	return prompt + "\n" + m.list.View()
}
