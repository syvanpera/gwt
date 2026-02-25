package cmd

import (
	"github.com/spf13/cobra"
	"github.com/syvanpera/gwt/internal/gitutil"
	"github.com/syvanpera/gwt/internal/ops"
	"github.com/syvanpera/gwt/internal/ui"
)

func init() {
	var showAll bool

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List worktrees",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := gitutil.ResolveGWTRepoRoot(".")
			if err != nil {
				return err
			}
			worktrees, err := ops.ListWorktrees(root)
			if err != nil {
				return err
			}

			if !showAll {
				filtered := make([]ops.Worktree, 0, len(worktrees))
				for _, wt := range worktrees {
					if wt.Bare {
						continue
					}
					filtered = append(filtered, wt)
				}
				worktrees = filtered
			}

			cmd.Println(ui.RenderWorktreeList(worktrees, root))
			return nil
		},
	}
	cmd.Flags().BoolVar(&showAll, "all", false, "include bare repository entry")
	rootCmd.AddCommand(cmd)
}
