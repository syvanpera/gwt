package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/syvanpera/gwt/internal/ops"
	"github.com/syvanpera/gwt/internal/ui"
)

func init() {
	var base string
	var noUpstream bool

	cmd := &cobra.Command{
		Use:     "checkout <branch> [directory]",
		Aliases: []string{"co"},
		Short:   "Create or open a branch as a worktree",
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := ops.CheckoutOptions{Branch: args[0], Base: base, NoUpstream: noUpstream}
			if len(args) > 1 {
				opts.Directory = args[1]
			}
			return ui.RunOperation("gwt checkout", "Create a new worktree for a branch", func(_ context.Context, em *ops.Emitter) error {
				return ops.Checkout(em, opts)
			})
		},
	}

	cmd.Flags().StringVarP(&base, "base", "B", "origin/main", "base branch for creating new branch")
	cmd.Flags().BoolVarP(&noUpstream, "no-upstream", "N", false, "skip remote upstream setup")
	rootCmd.AddCommand(cmd)
}
