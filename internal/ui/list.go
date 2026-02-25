package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/syvanpera/gwt/internal/ops"
)

func RenderWorktreeList(worktrees []ops.Worktree, root string) string {
	p := newPalette()
	if len(worktrees) == 0 {
		return p.panel.Render(p.muted.Render("No worktrees found"))
	}

	title := p.header.Render("WORKTREES") + " " + p.muted.Render(fmt.Sprintf("(%d)", len(worktrees)))
	headers := p.muted.Render(fmt.Sprintf("  %-16s %-12s %s", "BRANCH", "HEAD", "PATH"))

	rows := make([]string, 0, len(worktrees))
	for _, wt := range worktrees {
		branch := strings.TrimPrefix(wt.Branch, "refs/heads/")
		head := shortHead(wt.Head)
		path := displayPath(root, wt.Path)

		icon := "✓"
		rowStyle := p.success
		if wt.Bare {
			branch = "(bare)"
			head = "-"
			icon = "•"
			rowStyle = p.muted
		}
		row := fmt.Sprintf("%s %-16s %-12s %s", icon, branch, head, path)
		rows = append(rows, rowStyle.Render(row))
	}

	content := title + "\n" + headers + "\n\n" + strings.Join(rows, "\n")
	return p.panel.Width(listPanelWidth()).Render(content)
}

func shortHead(head string) string {
	head = strings.TrimSpace(head)
	if head == "" {
		return "-"
	}
	if len(head) > 12 {
		return head[:12]
	}
	return head
}

func displayPath(root, path string) string {
	if root == "" {
		return filepath.Clean(path)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.Clean(path)
	}
	if rel == "." {
		return "."
	}
	return filepath.ToSlash(rel)
}

func listPanelWidth() int {
	width, err := strconv.Atoi(strings.TrimSpace(os.Getenv("COLUMNS")))
	if err != nil || width <= 0 {
		return 90
	}
	return listMax(40, listMin(90, width-4))
}

func listMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func listMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}
