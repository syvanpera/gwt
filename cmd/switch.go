package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/syvanpera/gwt/internal/gitutil"
	"github.com/syvanpera/gwt/internal/ops"
	"github.com/syvanpera/gwt/internal/ui"
)

func init() {
	var printPath bool

	cmd := &cobra.Command{
		Use:   "switch [worktree]",
		Short: "Emit shell command to switch to a worktree",
		Long:  "Prints a cd command for the selected worktree. Use with: eval \"$(gwt switch <worktree>)\"",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			root, err := gitutil.ResolveGWTRepoRoot(".")
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			worktrees, err := ops.ListWorktrees(root)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			seen := map[string]struct{}{}
			suggestions := make([]string, 0, len(worktrees)*2)
			for _, wt := range worktrees {
				if wt.Bare {
					continue
				}
				branch := strings.TrimPrefix(wt.Branch, "refs/heads/")
				name := filepath.Base(wt.Path)

				if strings.HasPrefix(branch, toComplete) {
					if _, ok := seen[branch]; !ok {
						suggestions = append(suggestions, branch)
						seen[branch] = struct{}{}
					}
				}
				if strings.HasPrefix(name, toComplete) {
					if _, ok := seen[name]; !ok {
						suggestions = append(suggestions, name)
						seen[name] = struct{}{}
					}
				}
			}
			return suggestions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := gitutil.ResolveGWTRepoRoot(".")
			if err != nil {
				return err
			}
			worktrees, err := ops.ListWorktrees(root)
			if err != nil {
				return err
			}

			selector := ""
			if len(args) == 1 {
				selector = args[0]
			} else {
				selector, err = promptSwitchSelector(worktrees, root)
				if err != nil {
					return err
				}
			}

			selectedPath, err := resolveSwitchPath(worktrees, root, selector)
			if err != nil {
				return err
			}

			if printPath {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), selectedPath)
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "cd %s\n", shellQuote(selectedPath))
			return nil
		},
	}

	cmd.Flags().BoolVar(&printPath, "print-path", false, "print absolute path only (for use with cd \"$(...)\")")
	rootCmd.AddCommand(cmd)
}

func resolveSwitchPath(worktrees []ops.Worktree, root, selector string) (string, error) {
	absSelector := selector
	if !filepath.IsAbs(absSelector) {
		absSelector = filepath.Clean(filepath.Join(root, selector))
	}

	for _, wt := range worktrees {
		if wt.Bare {
			continue
		}
		branch := strings.TrimPrefix(wt.Branch, "refs/heads/")
		if selector == wt.Path || absSelector == wt.Path || selector == filepath.Base(wt.Path) || selector == branch {
			return wt.Path, nil
		}
	}
	return "", fmt.Errorf("worktree not found: %s", selector)
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func promptSwitchSelector(worktrees []ops.Worktree, root string) (string, error) {
	if !isPromptTerminal() {
		return "", fmt.Errorf("worktree name is required in non-interactive mode")
	}

	items := make([]ops.Worktree, 0, len(worktrees))
	for _, wt := range worktrees {
		if wt.Bare {
			continue
		}
		items = append(items, wt)
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no worktrees available")
	}
	if len(items) == 1 {
		return items[0].Path, nil
	}

	options := make([]ui.WorktreeOption, 0, len(items))
	for _, wt := range items {
		branch := strings.TrimPrefix(wt.Branch, "refs/heads/")
		path := wt.Path
		if rel, err := filepath.Rel(root, wt.Path); err == nil {
			path = filepath.ToSlash(rel)
		}
		options = append(options, ui.WorktreeOption{
			Branch: branch,
			Path:   path,
		})
	}

	selectedRelPath, err := ui.PickWorktree(options)
	if err != nil {
		if err == context.Canceled {
			return "", fmt.Errorf("selection cancelled")
		}
		return "", err
	}
	return filepath.Clean(filepath.Join(root, selectedRelPath)), nil
}

func isPromptTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
